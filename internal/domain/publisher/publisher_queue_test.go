package publisher

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/pedrodcsjostrom/opencm/internal/domain/post"
	"github.com/pedrodcsjostrom/opencm/internal/infrastructure/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestPublisherQueue_Start(t *testing.T) {
	tests := []struct {
		name                           string
		cfg                            *config.PublisherConfig
		expectedNumberOfRunningWorkers int
	}{
		{
			name: "Start with multiple workers",
			cfg: &config.PublisherConfig{
				WorkerNum:     2,
				RetryNum:      1,
				PublishBuffer: 1,
				RetryBuffer:   1,
			},
			expectedNumberOfRunningWorkers: 3,
		},
		{
			name: "Start with single worker",
			cfg: &config.PublisherConfig{
				WorkerNum:     1,
				RetryNum:      0,
				PublishBuffer: 1,
				RetryBuffer:   1,
			},
			expectedNumberOfRunningWorkers: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			ps := NewMockService(t)
			pf := NewMockPublisherFactory(t)
			pq := NewPublisherQueue(tt.cfg, pf, ps)
			pq.Start(ctx)

			// Wait for workers to start
			time.Sleep(50 * time.Millisecond)

			assert.Equal(t, tt.expectedNumberOfRunningWorkers, pq.CountRunning(), "Unexpected number of running workers")
			pq.Stop()
		})
	}
}

func TestPublisherQueue_Stop(t *testing.T) {
	cfg := &config.PublisherConfig{
		WorkerNum:     2,
		RetryNum:      1,
		PublishBuffer: 1,
		RetryBuffer:   1,
	}

	ctx := context.Background()
	ps := NewMockService(t)
	pf := NewMockPublisherFactory(t)
	pq := NewPublisherQueue(cfg, pf, ps)
	pq.Start(ctx)

	// Wait for workers to start
	time.Sleep(50 * time.Millisecond)

	initialCount := pq.CountRunning()
	pq.Stop()

	assert.Equal(t, 0, pq.CountRunning(),
		"Expected all workers to stop, but some are still running")
	assert.Greater(t, initialCount, 0,
		"Expected initial worker count to be greater than 0")
}

func TestPublisherQueue_Enqueue(t *testing.T) {
	cfg := &config.PublisherConfig{
		WorkerNum:     1,
		RetryNum:      1,
		PublishBuffer: 2,
		RetryBuffer:   2,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mockService := NewMockService(t)
	mockService.On("PublishPostToSocialNetwork", 
		mock.Anything, 
		mock.Anything, 
		mock.Anything,
		mock.Anything,
	).Return(nil)

	pq := &publisherQueue{
		publishCh:        make(chan *post.PublishPost, cfg.PublishBuffer),
		failedCh:        make(chan *post.PublishPost, cfg.RetryBuffer),
		publisherFactory: NewMockPublisherFactory(t),
		service:         mockService,
		cfg:            cfg,
		wg:             &sync.WaitGroup{},
	}

	testPost := &post.PublishPost{
		ID:        "test-1",
		ProjectID: "project-1",
		Platform:  "test-platform",
		Secrets:   "test-key",
	}

	pq.Start(ctx)
	pq.Enqueue(ctx, testPost)

	// Wait for processing
	time.Sleep(50 * time.Millisecond)

	mockService.AssertExpectations(t)
	pq.Stop()
}

func TestPublisherQueue_runPublishWorker(t *testing.T) {
	tests := []struct {
		name            string
		cfg             *config.PublisherConfig
		posts           []*post.PublishPost
		publishError    error
		expectedRetries int
	}{
		{
			name: "Successful publish - no retries",
			cfg: &config.PublisherConfig{
				WorkerNum:     1,
				RetryNum:      1,
				PublishBuffer: 2,
				RetryBuffer:   2,
			},
			posts: []*post.PublishPost{
				{ID: "1", ProjectID: "proj-1", Platform: "x", Secrets: "key1"},
			},
			publishError:    nil,
			expectedRetries: 0,
		},
		{
			name: "Failed publish - should move to failed channel",
			cfg: &config.PublisherConfig{
				WorkerNum:     1,
				RetryNum:      1,
				PublishBuffer: 2,
				RetryBuffer:   2,
			},
			posts: []*post.PublishPost{
				{ID: "2", ProjectID: "proj-2", Platform: "x", Secrets: "key2"},
			},
			publishError:    assert.AnError,
			expectedRetries: 1,
		},
		{
			name: "Multiple posts - mixed success",
			cfg: &config.PublisherConfig{
				WorkerNum:     1,
				RetryNum:      1,
				PublishBuffer: 3,
				RetryBuffer:   3,
			},
			posts: []*post.PublishPost{
				{ID: "3", ProjectID: "proj-3", Platform: "x", Secrets: "key3"},
				{ID: "4", ProjectID: "proj-4", Platform: "x", Secrets: "key4"},
			},
			publishError:    assert.AnError,
			expectedRetries: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			mockService := NewMockService(t)
			mockService.On("PublishPostToSocialNetwork", 
				mock.Anything, 
				mock.Anything, 
				mock.Anything,
				mock.Anything,
			).Return(tt.publishError)

			pq := &publisherQueue{
				publishCh:        make(chan *post.PublishPost, tt.cfg.PublishBuffer),
				failedCh:        make(chan *post.PublishPost, tt.cfg.RetryBuffer),
				publisherFactory: NewMockPublisherFactory(t),
				service:         mockService,
				cfg:            tt.cfg,
				wg:             &sync.WaitGroup{},
			}

			// Start the worker
			go pq.runPublishWorker(ctx)

			// Send posts to the channel
			for _, p := range tt.posts {
				pq.publishCh <- p
			}

			// Wait for processing
			time.Sleep(50 * time.Millisecond)

			// Close and wait
			close(pq.publishCh)
			pq.wg.Wait()

			assert.Equal(t, tt.expectedRetries, len(pq.failedCh),
				"Unexpected number of posts in failed channel")
			mockService.AssertNumberOfCalls(t, "PublishPostToSocialNetwork", len(tt.posts))
		})
	}
}

func TestPublisherQueue_runFailedHandlerWorker(t *testing.T) {
	tests := []struct {
		name         string
		cfg          *config.PublisherConfig
		posts        []*post.PublishPost
		expectedLogs int
	}{
		{
			name: "Single failed post",
			cfg: &config.PublisherConfig{
				WorkerNum:     1,
				RetryNum:      1,
				PublishBuffer: 1,
				RetryBuffer:   2,
			},
			posts: []*post.PublishPost{
				{ID: "failed-1", ProjectID: "proj-1", Platform: "x", Secrets: "key1"},
			},
			expectedLogs: 1,
		},
		{
			name: "Multiple failed posts",
			cfg: &config.PublisherConfig{
				WorkerNum:     1,
				RetryNum:      1,
				PublishBuffer: 1,
				RetryBuffer:   3,
			},
			posts: []*post.PublishPost{
				{ID: "failed-2", ProjectID: "proj-2", Platform: "x", Secrets: "key2"},
				{ID: "failed-3", ProjectID: "proj-3", Platform: "x", Secrets: "key3"},
				{ID: "failed-4", ProjectID: "proj-4", Platform: "x", Secrets: "key4"},
			},
			expectedLogs: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			pq := &publisherQueue{
				publishCh:        make(chan *post.PublishPost, tt.cfg.PublishBuffer),
				failedCh:        make(chan *post.PublishPost, tt.cfg.RetryBuffer),
				publisherFactory: NewMockPublisherFactory(t),
				service:         NewMockService(t),
				cfg:            tt.cfg,
				wg:             &sync.WaitGroup{},
			}

			// Start the failed handler worker
			go pq.runFailedHandlerWorker(ctx)

			// Send posts to failed channel
			for _, p := range tt.posts {
				pq.failedCh <- p
			}

			// Wait for processing
			time.Sleep(50 * time.Millisecond)

			// Close and wait
			close(pq.failedCh)
			pq.wg.Wait()

			// Verify no posts remain in the failed channel
			assert.Equal(t, 0, len(pq.failedCh),
				"Unexpected number of posts remaining in failed channel")

			// Additional verification that channel is empty and closed
			select {
			case _, ok := <-pq.failedCh:
				assert.False(t, ok, "Failed channel should be closed")
			default:
				// Channel is empty, which is expected
			}
		})
	}
}