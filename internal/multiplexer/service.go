package multiplexer

import (
	"context"

	"github.com/bynov/multiplexer/internal/pool"
)

type Service interface {
	GetAllContent(ctx context.Context, urls []string) (content []string, err error)
}

type service struct {
	maxWorkers int
}

func NewService(maxWorkers int) Service {
	return service{
		maxWorkers: maxWorkers,
	}
}

func (s service) GetAllContent(ctx context.Context, urls []string) (content []string, err error) {
	return pool.New(s.maxWorkers).Do(ctx, urls)
}
