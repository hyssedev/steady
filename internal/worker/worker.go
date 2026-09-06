package worker

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/hyssedev/steady/internal/database"
	"github.com/hyssedev/steady/internal/monitor"
)

type WorkerPool struct {
	ctx     context.Context
	timeout time.Duration
	channel chan monitor.Monitor
	db      database.Database

	client *http.Client
}

func NewWorkerPool(
	ctx context.Context,
	timeout time.Duration,
	channel chan monitor.Monitor,
	db database.Database,
) WorkerPool {
	return WorkerPool{
		ctx:     ctx,
		channel: channel,
		db:      db,

		client: newClient(timeout),
	}
}

func (wp WorkerPool) Work() {
	for i := 1; i <= 3; i++ {
		worker := NewWorker(wp.ctx, i, wp.client, wp.channel, wp.db)

		go worker.work()
	}
}

type Worker struct {
	ctx     context.Context
	id      int
	client  *http.Client
	channel chan monitor.Monitor

	db database.Database
}

func NewWorker(
	ctx context.Context,
	id int,
	client *http.Client,
	channel chan monitor.Monitor,
	db database.Database,
) Worker {
	return Worker{
		ctx:     ctx,
		id:      id,
		client:  client,
		channel: channel,
		db:      db,
	}
}

func (w Worker) work() {
	for {
		select {
		case job := <-w.channel:
			req, err := http.NewRequestWithContext(w.ctx, http.MethodGet, job.URL.String(), nil)
			if err != nil {
				fmt.Printf("Error creating request for job %v\n", job.Name)
				continue
			}

			startedAt := time.Now()

			resp, err := w.client.Do(req)
			latency := time.Since(startedAt)

			if err != nil {
				fmt.Printf("Check %v failed, err: %v\n", job.Name, err)

				if err := w.db.SaveCheck(w.ctx, job.ID, false, nil, latency.Milliseconds(), err); err != nil {
					fmt.Printf("Save check %v failed, err: %v\n", job.Name, err)
				}
				continue
			}

			_, _ = io.Copy(io.Discard, resp.Body)
			resp.Body.Close()

			if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusBadRequest {
				fmt.Printf("Check %v failed, status code: %v\n", job.Name, resp.StatusCode)

				if err := w.db.SaveCheck(
					w.ctx,
					job.ID,
					false,
					&resp.StatusCode,
					latency.Milliseconds(),
					err,
				); err != nil {
					fmt.Printf("Save check %v failed, err: %v\n", job.Name, err)
				}
				continue
			}

			fmt.Printf("Check %v successful\n", job.Name)
			if err := w.db.SaveCheck(w.ctx, job.ID, true, &resp.StatusCode, latency.Milliseconds(), err); err != nil {
				fmt.Printf("Save check %v failed, err: %v\n", job.Name, err)
			}
		case <-w.ctx.Done():
			return
		}
	}
}

func newClient(timeout time.Duration) *http.Client {
	return &http.Client{
		Timeout: timeout,
	}
}
