package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
)

type Server struct {
	Router   *chi.Mux
	Addr     string
	Interval time.Duration
	Logger   *log.Logger

	handler      *Handler
	shutdownCh   chan os.Signal
	worker       *worker
	workerKillCh chan<- struct{}
	wg           *sync.WaitGroup
}

func (s *Server) addr() string {
	if s.Addr == "" {
		s.Addr = "8080"
	}

	return fmt.Sprintf(":%s", s.Addr)
}

func (s *Server) interval() time.Duration {
	if s.Interval == 0 {
		s.Interval = 5 * time.Second
	}

	return s.Interval
}

func (s *Server) init() {
	s.handler = &Handler{}
	s.setRoutes()

	s.shutdownCh = make(chan os.Signal, 1)
	signal.Notify(s.shutdownCh, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	workerKillCh := make(chan struct{}, 1)
	s.workerKillCh = workerKillCh
	s.worker = &worker{
		d:      s.interval(),
		killCh: workerKillCh,
	}

	s.wg = &sync.WaitGroup{}
}

func (s *Server) setRoutes() {
	s.Router.Get("/", s.handler.HelloWorld())
}

func (s *Server) run(runFn func()) {
	go func() {
		s.wg.Add(1)
		defer s.wg.Done()

		runFn()
	}()
}

func (s *Server) listenAndServe() error {
	httpServer := &http.Server{
		Addr:    s.addr(),
		Handler: s.Router,
	}

	startCh := make(chan error, 1)
	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			startCh <- fmt.Errorf("failed to start server: %w", err)
		}
	}()

	// Wait for either a shutdown signal or an error if the server
	// cannot start.
	select {
	case err := <-startCh:
		return err
	case <-s.shutdownCh:
		ctx, cancel := context.WithTimeout(context.Background(), 7*time.Second)
		defer func() {
			defer cancel()

			// Kill background worker.
			s.workerKillCh <- struct{}{}

			// Wait for all resources to stop.
			s.wg.Wait()
		}()

		// Gracefully shutdown the http server.
		if err := httpServer.Shutdown(ctx); err != nil {
			return fmt.Errorf("failed to shutdown server: %w", err)
		}
	}

	return nil
}

func (s *Server) validate() error {
	if s.Router == nil {
		return errors.New("router is nil")
	}

	if s.Logger == nil {
		return errors.New("logger is nil")
	}

	return nil
}

func (s *Server) Start() error {
	if err := s.validate(); err != nil {
		return err
	}

	s.init()
	s.run(func() {
		s.worker.start()
	})

	return s.listenAndServe()
}
