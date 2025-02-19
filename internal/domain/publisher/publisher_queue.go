package publisher

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/redplanettribe/social-media-manager/internal/domain/post"
	"github.com/redplanettribe/social-media-manager/internal/infrastructure/config"
)

type PublisherQueue interface {
	Start(ctx context.Context)
	Stop()
	Enqueue(ctx context.Context, p *post.PublishPost)
	CountRunning() int
}

// PublisherQueue manages the channels and workers for publishing
type publisherQueue struct {
	publishCh        chan *post.PublishPost
	failedCh         chan *post.PublishPost
	publisherFactory PublisherFactory
	cfg              *config.PublisherConfig
	wg               *sync.WaitGroup
	running          int32
	service          Service
}

// NewPublisherQueue initializes the queue with desired worker counts
func NewPublisherQueue(cfg *config.PublisherConfig, pf PublisherFactory, svc Service) PublisherQueue {
	return &publisherQueue{
		publishCh:        make(chan *post.PublishPost, cfg.PublishBuffer),
		failedCh:         make(chan *post.PublishPost, cfg.RetryBuffer),
		publisherFactory: pf,
		cfg:              cfg,
		wg:               &sync.WaitGroup{},
		running:          0,
		service:          svc,
	}
}

// Start spins up workers for publishing and retry handling
func (pq *publisherQueue) Start(ctx context.Context) {
	for i := 0; i < pq.cfg.WorkerNum; i++ {
		go pq.runPublishWorker(ctx)
	}
	for i := 0; i < pq.cfg.RetryNum; i++ {
		go pq.runFailedHandlerWorker(ctx)
	}
}

// Stop signals workers to finish
func (pq *publisherQueue) Stop() {
	close(pq.publishCh)
	close(pq.failedCh)
	pq.wg.Wait()
}

// Enqueue adds a post to the publishCh
func (pq *publisherQueue) Enqueue(ctx context.Context, p *post.PublishPost) {
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
				pq.failedCh <- p
			}
		}
	}
}

// runFailedHandlerWorker tries to re-publish failed posts
func (pq *publisherQueue) runFailedHandlerWorker(ctx context.Context) {
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
		case p, ok := <-pq.failedCh:
			if !ok {
				return
			}
			// we could do any number of things here, like exponential backoff, logging, changing status in db, etc.
			// at the moment let's only logg it
			fmt.Println("Processing failed post :", p.ID)
		}
	}
}

// CountRunning returns how many goroutines are active
func (pq *publisherQueue) CountRunning() int {
	return pq.getRunning()
}

// publishPost sends a post to the correct publisher
func (pq *publisherQueue) publishPost(ctx context.Context, p *post.PublishPost) error {
	return pq.service.PublishPostToSocialNetwork(ctx, p.ProjectID, p.ID, p.Platform)
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
