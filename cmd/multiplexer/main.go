package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi"
	"github.com/rs/zerolog/log"

	"github.com/bynov/multiplexer/internal/multiplexer"
	"github.com/bynov/multiplexer/internal/transport"
)

const (
	requestLimit       = 1
	port               = 8080
	outgoingLimitation = 4
)

func main() {
	svc := multiplexer.NewService(outgoingLimitation)

	// Setup listener with connection limit
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Error().Err(err).Send()
		os.Exit(1)
	}
	defer func() { _ = l.Close() }()

	// Init router
	r := chi.NewRouter()
	r.Use(NewLimitMiddleWare(requestLimit))
	r.Route("/v1", func(r chi.Router) {
		r.Post("/content", transport.NewGetContentEndpoint(svc))
	})

	s := http.Server{
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		Handler:      r,
	}

	idleConnsClosed := make(chan struct{})
	go func() {
		gracefulStop := make(chan os.Signal, 1)
		signal.Notify(gracefulStop, syscall.SIGINT, syscall.SIGTERM)
		<-gracefulStop

		if err := s.Shutdown(context.Background()); err != nil {
			log.Error().Str("errType", "shutdownHTTPServer").Err(err).Send()
		}
		close(idleConnsClosed)
	}()

	go func() { _ = s.Serve(l) }()

	log.Info().Str("HTTP server is started", s.Addr).Send()

	<-idleConnsClosed
	log.Info().Msg("Service gracefully stopped")
}

func NewLimitMiddleWare(limit int) func(http.Handler) http.Handler {
	if limit <= 0 {
		panic("Invalid limit")
	}

	sem := make(chan struct{}, limit)
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			select {
			case sem <- struct{}{}:
				next.ServeHTTP(w, r)
				<-sem
			default:
				w.WriteHeader(http.StatusTooManyRequests)
			}
		}

		return http.HandlerFunc(fn)
	}
}
