package pool

import (
	"context"
	"io/ioutil"
	"net/http"
	"sort"
	"sync"
	"time"
)

type Pool struct {
	maxWorkers int

	wg sync.WaitGroup

	cl      *http.Client
	errChan chan error
}

type orderedURL struct {
	url   string
	index uint8
}

type orderedContent struct {
	content string
	index   uint8
}

func New(maxWorkers int) *Pool {
	return &Pool{
		cl: &http.Client{
			Timeout: time.Second,
		},
		maxWorkers: maxWorkers,

		errChan: make(chan error, 20),
	}
}

func (p *Pool) Do(ctx context.Context, urls []string) ([]string, error) {
	var work = make(chan orderedURL, len(urls))
	var result = make(chan orderedContent, len(urls))

	var done = make(chan struct{}, 1)

	p.wg.Add(p.maxWorkers)

	// Create a separate ctx for workers
	workerCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Spawn workers
	for i := 0; i < p.maxWorkers; i++ {
		go p.worker(workerCtx, work, result)
	}

	go func() {
		p.wg.Wait()
		close(done)
	}()

	// Add work
	for k, url := range urls {
		work <- orderedURL{url: url, index: uint8(k)}
	}

	close(work)

	// Wait for: some errors || ctx cancellation || all ok
	select {
	case <-ctx.Done():
		return nil, nil
	case err := <-p.errChan:
		return nil, err
	case <-done:
	}

	close(result)

	// Restore original order
	var ordResult = make([]orderedContent, 0, len(urls))
	for r := range result {
		ordResult = append(ordResult, r)
	}

	sort.Slice(ordResult, func(i, j int) bool {
		return ordResult[i].index < ordResult[j].index
	})

	var out = make([]string, 0, len(urls))
	for _, v := range ordResult {
		out = append(out, v.content)
	}

	return out, nil
}

func (p *Pool) worker(ctx context.Context, urls <-chan orderedURL, results chan<- orderedContent) {
	defer p.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case url, ok := <-urls:
			if !ok {
				return
			}

			req, err := http.NewRequestWithContext(ctx, http.MethodGet, url.url, nil)
			if err != nil {
				p.errChan <- err
				return
			}

			resp, err := p.cl.Do(req)
			if err != nil {
				p.errChan <- err
				return
			}

			b, err := ioutil.ReadAll(resp.Body)
			_ = resp.Body.Close()
			if err != nil {
				p.errChan <- err
				return
			}

			results <- orderedContent{content: string(b), index: url.index}
		}
	}

}
