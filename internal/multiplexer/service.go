package multiplexer

import (
	"context"

	"github.com/bynov/multiplexer/internal/pool"
)

type Service struct {
	maxWorkers int
}

func NewService(maxWorkers int) Service {
	return Service{
		maxWorkers: maxWorkers,
	}
}

func (s Service) GetAllContent(ctx context.Context, urls []string) (content []string, err error) {
	return pool.New(s.maxWorkers).Do(ctx, urls)
}
