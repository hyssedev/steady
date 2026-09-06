package worker

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/hyssedev/steady/internal/config"
	"github.com/hyssedev/steady/internal/monitor"
)

type WorkerPool struct {
	ctx     context.Context
	cfg     config.Config
	channel chan *monitor.Monitor

	client  *http.Client
	workers []Worker
}

func NewWorkerPool(ctx context.Context, cfg config.Config, channel chan *monitor.Monitor) WorkerPool {
	return WorkerPool{
		ctx:     ctx,
		cfg:     cfg,
		channel: channel,

		client: newClient(cfg),
	}
}

func (wp *WorkerPool) Work() {
	for i := 1; i <= 3; i++ {
		worker := NewWorker(wp.ctx, i, wp.client, wp.channel)
		wp.workers = append(wp.workers, worker)

		go worker.work()
	}
}

type Worker struct {
	ctx     context.Context
	id      int
	client  *http.Client
	channel chan *monitor.Monitor
}

func NewWorker(ctx context.Context, id int, client *http.Client, channel chan *monitor.Monitor) Worker {
	return Worker{
		ctx:     ctx,
		id:      id,
		client:  client,
		channel: channel,
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

			resp, err := w.client.Do(req)
			if err != nil {
				fmt.Printf("Check %v failed, err: %v\n", job.Name, err)
				continue
			}

			_, _ = io.Copy(io.Discard, resp.Body)
			resp.Body.Close()

			if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusBadRequest {
				fmt.Printf("Check %v failed, status code: %v\n", job.Name, resp.StatusCode)
			}

			fmt.Printf("Check %v successful\n", job.Name)

		case <-w.ctx.Done():
			// TODO: clean-up
			return
		}
	}
}

func newClient(cfg config.Config) *http.Client {
	return &http.Client{
		// Transport: nil,
		// CheckRedirect: func(req *http.Request, via []*http.Request) error {
		// 	panic("TODO")
		// },
		// Jar:     nil,
		Timeout: cfg.Timeout,
	}
}
