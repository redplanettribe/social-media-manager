package publisher

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/pedrodcsjostrom/opencm/internal/domain/post"
	"github.com/pedrodcsjostrom/opencm/internal/infrastructure/config"
	"github.com/pedrodcsjostrom/opencm/internal/infrastructure/platforms"
)

//go:generate mockery --name=PublisherQueue --case=underscore --inpackage
type PublisherQueue interface {
	Start(ctx context.Context)
	Stop()
	Enqueue(ctx context.Context, p *post.QPost)
	CountRunning() int
}

// PublisherQueue manages the channels and workers for publishing
type publisherQueue struct {
	publishCh        chan *post.QPost
	retryCh          chan *post.QPost
	publisherFactory platforms.PublisherFactory
	cfg              *config.PublisherConfig
	wg               *sync.WaitGroup
	running          int32
}

// NewPublisherQueue initializes the queue with desired worker counts
func NewPublisherQueue(cfg *config.PublisherConfig) PublisherQueue {
	return &publisherQueue{
		publishCh:        make(chan *post.QPost, cfg.PublishBuffer),
		retryCh:          make(chan *post.QPost, cfg.RetryBuffer),
		publisherFactory: platforms.NewPublisherFactory(),
		cfg:              cfg,
		wg:               &sync.WaitGroup{},
		running:          0,
	}
}

// Start spins up workers for publishing and retry handling
func (pq *publisherQueue) Start(ctx context.Context) {
	for i := 0; i < pq.cfg.WorkerNum; i++ {
		go pq.runPublishWorker(ctx)
	}
	for i := 0; i < pq.cfg.RetryNum; i++ {
		go pq.runRetryWorker(ctx)
	}
}

// Stop signals workers to finish
func (pq *publisherQueue) Stop() {
	close(pq.publishCh)
	close(pq.retryCh)
	pq.wg.Wait()
}

// Enqueue adds a post to the publishCh
func (pq *publisherQueue) Enqueue(ctx context.Context, p *post.QPost) {
	pq.publishCh <- p
}

// runPublishWorker consumes publishCh, on error sends to retryCh
func (pq *publisherQueue) runPublishWorker(ctx context.Context) {
	pq.wg.Add(1)
    pq.incrementRunning()
    defer func() {
        pq.decrementRunning()
        pq.wg.Done()
    }()

    for {
        select {
        case <-ctx.Done():
            return
        case p, ok := <-pq.publishCh:
            if !ok {
                return
            }
            if err := pq.publishPost(ctx, p); err != nil {
                pq.retryCh <- p
            }
        }
    }
}

// runRetryWorker tries to re-publish failed posts
func (pq *publisherQueue) runRetryWorker(ctx context.Context) {
	pq.wg.Add(1)
    pq.incrementRunning()
    defer func() {
        pq.decrementRunning()
        pq.wg.Done()
    }()

    for {
        select {
        case <-ctx.Done():
            return
        case p, ok := <-pq.retryCh:
            if !ok {
                return
            }
            if err := pq.publishPost(ctx, p); err != nil {
                fmt.Printf("Post %s failed again: %v\n", p.ID, err)
                // handle permanent failure, logging, etc.
				// remove from retryCh to avoid infinite loop

                continue
            }
        }
    }
}

// CountRunning returns how many goroutines are active
func (pq *publisherQueue) CountRunning() int {
    return pq.getRunning()
}

// publishPost sends a post to the correct publisher
func (pq *publisherQueue) publishPost(ctx context.Context, p *post.QPost) error {
	pub := pq.publisherFactory.Create(p.Platform, p.ApiKey)
	if err := pub.Publish(ctx, p); err != nil {
		return err
	}
	return nil
}

func (pq *publisherQueue) incrementRunning() {
	atomic.AddInt32(&pq.running, 1)
}

func (pq *publisherQueue) decrementRunning() {
	atomic.AddInt32(&pq.running, -1)
}

func (pq *publisherQueue) getRunning() int {
	return int(atomic.LoadInt32(&pq.running))
}