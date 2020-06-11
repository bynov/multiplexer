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
	"golang.org/x/net/netutil"

	"github.com/bynov/multiplexer/internal/multiplexer"
	"github.com/bynov/multiplexer/internal/transport"
)

const (
	connectionLimit = 100
	port            = 8080
)

func main() {
	svc := multiplexer.NewService(4)

	// Setup listener with connection limit
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Error().Err(err).Send()
		os.Exit(1)
	}
	defer func() { _ = l.Close() }()

	l = netutil.LimitListener(l, connectionLimit)

	// Init router
	r := chi.NewRouter()
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
