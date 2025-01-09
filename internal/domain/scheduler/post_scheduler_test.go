package scheduler

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/pedrodcsjostrom/opencm/internal/domain/post"
	"github.com/pedrodcsjostrom/opencm/internal/domain/project"
	pq "github.com/pedrodcsjostrom/opencm/internal/domain/publisher"
	"github.com/pedrodcsjostrom/opencm/internal/infrastructure/config"
)

func TestPostScheduler_Start(t *testing.T) {
	tests := []struct {
		name     string
		interval time.Duration
		setup    func(*post.MockService, *project.MockService, *pq.MockPublisherQueue)
	}{
		{
			name:     "starts scheduler with correct interval",
			interval: 100 * time.Millisecond,
			setup: func(mps *post.MockService, mpjs *project.MockService, mpq *pq.MockPublisherQueue) {
				mps.On("FindScheduledReadyPosts", mock.Anything, 0, 100).Return([]*post.PublishPost{}, nil)
				mpjs.On("FindActiveProjectsChunk", mock.Anything, 0, 100).Return([]*project.Project{}, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
			defer cancel()

			mockPostSvc := post.NewMockService(t)
			mockProjectSvc := project.NewMockService(t)
			mockPubQueue := pq.NewMockPublisherQueue(t)

			tt.setup(mockPostSvc, mockProjectSvc, mockPubQueue)

			cfg := &config.SchedulerConfig{
				Interval:      tt.interval,
				ChannelBuffer: 10,
			}

			scheduler := NewPostScheduler(mockPostSvc, mockProjectSvc, mockPubQueue, cfg)
			scheduler.Start(ctx)

			time.Sleep(150 * time.Millisecond)
			scheduler.Stop()

			// Verify that services were called at least once
			mockPostSvc.AssertExpectations(t)
			mockProjectSvc.AssertExpectations(t)
		})
	}
}

func TestPostScheduler_ScanAndEnqueue(t *testing.T) {
	tests := []struct {
		name            string
		scheduledPosts  []*post.PublishPost
		projects        []*project.Project
		projectPosts    map[string]*post.PublishPost // projectID -> post
		expectedErrors  bool
		expectedEnqueue int
	}{
		{
			name: "successful scan with both scheduled and project posts",
			scheduledPosts: []*post.PublishPost{
				{ID: "scheduled1", Platform: "platform1"},
				{ID: "scheduled2", Platform: "platform2"},
			},
			projects: []*project.Project{
				{ID: "proj1"},
				{ID: "proj2"},
			},
			projectPosts: map[string]*post.PublishPost{
				"proj1": {ID: "proj1-post", Platform: "platform1"},
				"proj2": {ID: "proj2-post", Platform: "platform2"},
			},
			expectedErrors:  false,
			expectedEnqueue: 4,
		},
		{
			name: "deduplication of posts",
			scheduledPosts: []*post.PublishPost{
				{ID: "post1", Platform: "platform1"},
				{ID: "post1", Platform: "platform1"}, // duplicate
			},
			projects: []*project.Project{
				{ID: "proj1"},
			},
			projectPosts: map[string]*post.PublishPost{
				"proj1": {ID: "post1", Platform: "platform1"}, // duplicate
			},
			expectedErrors:  false,
			expectedEnqueue: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			mockPostSvc := post.NewMockService(t)
			mockProjectSvc := project.NewMockService(t)
			mockPubQueue := pq.NewMockPublisherQueue(t)

			// Setup scheduled posts
			mockPostSvc.On("FindScheduledReadyPosts", mock.Anything, 0, 100).
				Return(tt.scheduledPosts, nil)
			mockPostSvc.On("FindScheduledReadyPosts", mock.Anything, 100, 100).
				Return([]*post.PublishPost{}, nil)

			// Setup projects
			mockProjectSvc.On("FindActiveProjectsChunk", mock.Anything, 0, 100).
				Return(tt.projects, nil)
			mockProjectSvc.On("FindActiveProjectsChunk", mock.Anything, 100, 100).
				Return([]*project.Project{}, nil)

			// Setup project posts
			for _, proj := range tt.projects {
				if post, exists := tt.projectPosts[proj.ID]; exists {
					mockProjectSvc.On("FindOneReadyPostInQueue", mock.Anything, proj.ID).
						Return(post.ID, nil)
					mockPostSvc.On("GetQueuePost", mock.Anything, post.ID).
						Return(post, nil)
				}
			}

			// Setup publisher queue
			mockPubQueue.On("Enqueue", mock.Anything, mock.Anything).Return()

			cfg := &config.SchedulerConfig{
				Interval:      time.Second,
				ChannelBuffer: 10,
			}

			scheduler := NewPostScheduler(mockPostSvc, mockProjectSvc, mockPubQueue, cfg)
			err := scheduler.scanAndEnqueue(ctx)

			if tt.expectedErrors {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Verify enqueue calls
			mockPubQueue.AssertNumberOfCalls(t, "Enqueue", tt.expectedEnqueue)
		})
	}
}

func TestPostScheduler_ScanScheduledPosts(t *testing.T) {
	tests := []struct {
		name          string
		chunks        [][]*post.PublishPost
		expectedError error
		expectedPosts int
		contextCancel bool
	}{
		{
			name: "successful scan multiple chunks",
			chunks: [][]*post.PublishPost{
				{{ID: "1"}, {ID: "2"}},
				{{ID: "3"}},
				{},
			},
			expectedError: nil,
			expectedPosts: 3,
			contextCancel: false,
		},
		{
			name: "error during scan",
			chunks: [][]*post.PublishPost{
				{{ID: "1"}},
			},
			expectedError: fmt.Errorf("scan error"),
			expectedPosts: 0,
			contextCancel: false,
		},
		{
			name: "context cancellation",
			chunks: [][]*post.PublishPost{
				{{ID: "1"}},
			},
			contextCancel: true,
			expectedError: context.Canceled,
			expectedPosts: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			mockPostSvc := post.NewMockService(t)
			mockProjectSvc := project.NewMockService(t)
			mockPubQueue := pq.NewMockPublisherQueue(t)

			// Setup chunks
			for i, chunk := range tt.chunks {
				if tt.contextCancel {
					return
				}
				if tt.expectedError != nil && i == len(tt.chunks)-1 {
					mockPostSvc.On("FindScheduledReadyPosts", mock.Anything, i*100, 100).
						Return(nil, tt.expectedError)
					break
				}
				mockPostSvc.On("FindScheduledReadyPosts", mock.Anything, i*100, 100).
					Return(chunk, nil)
			}

			cfg := &config.SchedulerConfig{
				Interval:      time.Second,
				ChannelBuffer: 10,
			}

			scheduler := NewPostScheduler(mockPostSvc, mockProjectSvc, mockPubQueue, cfg)

			// Create channel to collect posts
			posts := make(chan *post.PublishPost, 100)

			if tt.contextCancel {
				cancel()
			}

			err := scheduler.scanScheduledPosts(ctx, posts)

			if tt.expectedError != nil {
				assert.Error(t, err)
				if tt.contextCancel {
					assert.Equal(t, context.Canceled, err)
				}
			} else {
				assert.NoError(t, err)
			}

			close(posts)

			receivedPosts := 0
			for range posts {
				receivedPosts++
			}

			if !tt.contextCancel {
				assert.Equal(t, tt.expectedPosts, receivedPosts)
			}
		})
	}
}

func TestPostScheduler_ScanProjectQueues(t *testing.T) {
	tests := []struct {
		name          string
		projects      [][]*project.Project
		projectPosts  map[string]*post.PublishPost
		expectedError error
		expectedPosts int
	}{
		{
			name: "successful scan multiple projects",
			projects: [][]*project.Project{
				{{ID: "proj1"}, {ID: "proj2"}},
				{},
			},
			projectPosts: map[string]*post.PublishPost{
				"proj1": {ID: "post1"},
				"proj2": {ID: "post2"},
			},
			expectedError: nil,
			expectedPosts: 2,
		},
		{
			name: "error finding projects",
			projects: [][]*project.Project{
				{{ID: "proj1"}},
			},
			expectedError: fmt.Errorf("project scan error"),
			expectedPosts: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			mockPostSvc := post.NewMockService(t)
			mockProjectSvc := project.NewMockService(t)
			mockPubQueue := pq.NewMockPublisherQueue(t)

			// Setup project chunks
			for i, chunk := range tt.projects {
				if tt.expectedError != nil && i == len(tt.projects)-1 {
					mockProjectSvc.On("FindActiveProjectsChunk", mock.Anything, i*100, 100).
						Return(nil, tt.expectedError)
					break
				}
				mockProjectSvc.On("FindActiveProjectsChunk", mock.Anything, i*100, 100).
					Return(chunk, nil)
			}

			// Setup project posts
			for projectID, qPost := range tt.projectPosts {
				mockProjectSvc.On("FindOneReadyPostInQueue", mock.Anything, projectID).
					Return(qPost.ID, nil)
				mockPostSvc.On("GetQueuePost", mock.Anything, qPost.ID).
					Return(qPost, nil)
			}

			cfg := &config.SchedulerConfig{
				Interval:      time.Second,
				ChannelBuffer: 10,
			}

			scheduler := NewPostScheduler(mockPostSvc, mockProjectSvc, mockPubQueue, cfg)

			// Create channel to collect posts
			posts := make(chan *post.PublishPost, 100)

			err := scheduler.scanProjectQueues(ctx, posts)

			if tt.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			close(posts)
			receivedPosts := 0
			for range posts {
				receivedPosts++
			}

			assert.Equal(t, tt.expectedPosts, receivedPosts)
		})
	}
}
