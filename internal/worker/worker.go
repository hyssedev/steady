package worker

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

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
			req := newRequest(job)
			resp, err := w.client.Do(req)
			if err != nil {
				fmt.Printf("Check %v failed\n", job.Name)
				continue
			}
			defer resp.Body.Close()

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

func newRequest(monitor *monitor.Monitor) *http.Request {
	url, _ := url.Parse(monitor.URL)

	return &http.Request{
		Method: http.MethodGet,
		URL:    url,
		// Header:     http.Header{},
		// Body:       nil,
		// GetBody: func() (io.ReadCloser, error) {
		// 	panic("TODO")
		// },
		// ContentLength:    0,
		// TransferEncoding: []string{},
		// Close:            false,
		// Host:             "",
		// Form:             url.Values{},
		// PostForm:         url.Values{},
		// MultipartForm:    &multipart.Form{},
		// Trailer:          http.Header{},
		// RemoteAddr:       "",
		// RequestURI:       monitor.URL,
		// TLS:              &tls.ConnectionState{},
		// Cancel:           make(<-chan struct{}),
		// Response:         &http.Response{},
		// Pattern:          "",
	}
}
