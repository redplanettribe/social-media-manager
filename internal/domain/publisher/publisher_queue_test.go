package publisher

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/pedrodcsjostrom/opencm/internal/domain/platform"
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
			pf:= platform.NewMockPublisherFactory(t)
			pq := NewPublisherQueue(tt.cfg, pf)
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
	pf:= platform.NewMockPublisherFactory(t)
	pq := NewPublisherQueue(cfg, pf)
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

	mockPublisher := platform.NewMockPublisher(t)
	mockPublisher.On("Publish", mock.Anything, mock.Anything).Return(nil)

	mockPublisherFactory := platform.NewMockPublisherFactory(t)
	mockPublisherFactory.On("Create", mock.Anything, mock.Anything).Return(mockPublisher)

	pq := &publisherQueue{
		publishCh:        make(chan *post.QPost, cfg.PublishBuffer),
		failedCh:          make(chan *post.QPost, cfg.RetryBuffer),
		publisherFactory: mockPublisherFactory,
		cfg:              cfg,
		wg:               &sync.WaitGroup{},
	}

	testPost := &post.QPost{
		ID: "test-1",
		Platform: "test-platform",
		ApiKey:   "test-key",
	}

	pq.Start(ctx)
	pq.Enqueue(ctx, testPost)

	// Wait for processing
	time.Sleep(50 * time.Millisecond)

	mockPublisher.AssertExpectations(t)
	mockPublisherFactory.AssertExpectations(t)
	pq.Stop()
}

func TestPublisherQueue_runPublishWorker(t *testing.T) {
	tests := []struct {
		name            string
		cfg             *config.PublisherConfig
		posts           []*post.QPost
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
			posts: []*post.QPost{
				{ID: "1", Platform: "x", ApiKey: "key1"},
			},
			publishError:    nil,
			expectedRetries: 0,
		},
		{
			name: "Failed publish - should retry",
			cfg: &config.PublisherConfig{
				WorkerNum:     1,
				RetryNum:      1,
				PublishBuffer: 2,
				RetryBuffer:   2,
			},
			posts: []*post.QPost{
				{ID: "2", Platform: "x", ApiKey: "key2"},
			},
			publishError:    fmt.Errorf("publish error"),
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
			posts: []*post.QPost{
				{ID: "3", Platform: "x", ApiKey: "key3"},
				{ID: "4", Platform: "x", ApiKey: "key4"},
			},
			publishError:    fmt.Errorf("publish error"),
			expectedRetries: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			mockPublisher := platform.NewMockPublisher(t)
			mockPublisher.On("Publish", mock.Anything, mock.Anything).Return(tt.publishError)

			publisherFactory := platform.NewMockPublisherFactory(t)
			publisherFactory.On("Create", mock.Anything, mock.Anything).Return(mockPublisher)

			pq := &publisherQueue{
				publishCh:        make(chan *post.QPost, tt.cfg.PublishBuffer),
				failedCh:          make(chan *post.QPost, tt.cfg.RetryBuffer),
				publisherFactory: publisherFactory,
				cfg:              tt.cfg,
				wg:               &sync.WaitGroup{},
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
				"Unexpected number of retries")
			mockPublisher.AssertNumberOfCalls(t, "Publish", len(tt.posts))
		})
	}
}

func TestPublisherQueue_runRetryWorker(t *testing.T) {
	tests := []struct {
		name                 string
		cfg                  *config.PublisherConfig
		posts                []*post.QPost
		publishError         error
		expectedPublishCalls int
	}{
		{
			name: "Successful retry",
			cfg: &config.PublisherConfig{
				WorkerNum:     1,
				RetryNum:      1,
				PublishBuffer: 1,
				RetryBuffer:   2,
			},
			posts: []*post.QPost{
				{ID: "retry-1", Platform: "x", ApiKey: "key1"},
			},
			publishError:         nil,
			expectedPublishCalls: 1,
		},
		{
			name: "Failed retry - should continue without readding to channel",
			cfg: &config.PublisherConfig{
				WorkerNum:     1,
				RetryNum:      1,
				PublishBuffer: 1,
				RetryBuffer:   2,
			},
			posts: []*post.QPost{
				{ID: "retry-2", Platform: "x", ApiKey: "key2"},
			},
			publishError:         fmt.Errorf("permanent error"),
			expectedPublishCalls: 1,
		},
		{
			name: "Multiple posts - mixed success and failure",
			cfg: &config.PublisherConfig{
				WorkerNum:     1,
				RetryNum:      1,
				PublishBuffer: 1,
				RetryBuffer:   3,
			},
			posts: []*post.QPost{
				{ID: "retry-3", Platform: "x", ApiKey: "key3"},
				{ID: "retry-4", Platform: "x", ApiKey: "key4"},
				{ID: "retry-5", Platform: "x", ApiKey: "key5"},
			},
			publishError:         fmt.Errorf("permanent error"),
			expectedPublishCalls: 3, 
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			mockPublisher := platform.NewMockPublisher(t)
			mockPublisher.On("Publish", mock.Anything, mock.Anything).Return(tt.publishError)

			publisherFactory := platform.NewMockPublisherFactory(t)
			publisherFactory.On("Create", mock.Anything, mock.Anything).Return(mockPublisher)

			pq := &publisherQueue{
				publishCh:        make(chan *post.QPost, tt.cfg.PublishBuffer),
				failedCh:          make(chan *post.QPost, tt.cfg.RetryBuffer),
				publisherFactory: publisherFactory,
				cfg:              tt.cfg,
				wg:               &sync.WaitGroup{},
			}

			// Start the retry worker
			go pq.runFailedHandlerWorker(ctx)

			// Send posts to retry channel
			for _, p := range tt.posts {
				pq.failedCh <- p
			}

			// Wait for processing
			time.Sleep(50 * time.Millisecond)

			// Close and wait
			close(pq.failedCh)
			pq.wg.Wait()

			// Verify the publish calls
			mockPublisher.AssertNumberOfCalls(t, "Publish", tt.expectedPublishCalls)

			// Verify no posts remain in the retry channel
			assert.Equal(t, 0, len(pq.failedCh),
				"Unexpected number of posts remaining in retry channel")

			// Additional verification that channel is empty
			select {
			case _, ok := <-pq.failedCh:
				assert.False(t, ok, "Retry channel should be closed")
			default:
				// Channel is empty, which is expected
			}
		})
	}
}
