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

	"github.com/cicconee/weather-app/internal/admin"
	"github.com/cicconee/weather-app/internal/alert"
	"github.com/cicconee/weather-app/internal/forecast"
	"github.com/cicconee/weather-app/internal/state"
	"github.com/go-chi/chi/v5"
)

type Server struct {
	Router    *chi.Mux
	Addr      string
	Interval  time.Duration
	Logger    *log.Logger
	States    *state.Service
	Alerts    *alert.Service
	Forecasts *forecast.Service
	Admins    *admin.Service

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
	s.handler = NewHandler(s.Logger)
	s.handler.states = s.States
	s.handler.alerts = s.Alerts
	s.handler.forecasts = s.Forecasts
	s.handler.admins = s.Admins
	s.setRoutes()

	s.shutdownCh = make(chan os.Signal, 1)
	signal.Notify(s.shutdownCh, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	workerKillCh := make(chan struct{}, 1)
	s.workerKillCh = workerKillCh
	s.worker = &worker{
		alerts: s.Alerts,
		d:      s.interval(),
		killCh: workerKillCh,
	}

	s.wg = &sync.WaitGroup{}
}

func (s *Server) setRoutes() {
	s.Router.Get("/", s.handler.HelloWorld())
	s.Router.Get("/alerts", s.handler.HandleGetAlerts())
	s.Router.Get("/forecasts", s.handler.HandleGetForecast())

	// Set the admin routes.
	adminValidater := AdminValidater{
		admins: s.Admins,
		logger: s.Logger,
	}

	s.Router.Post("/admins/login", s.handler.HandlePostLogin())
	s.Router.Post("/admins/signup", s.handler.HandlePostSignup())
	s.Router.Post("/admins/states", adminValidater.Validate(s.handler.HandleCreateState()))
	s.Router.Post("/admins/states/sync", adminValidater.Validate(s.handler.HandleSyncState()))
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

	if s.States == nil {
		return errors.New("states is nil")
	}

	if s.Alerts == nil {
		return errors.New("alerts is nil")
	}

	if s.Forecasts == nil {
		return errors.New("forecasts is nil")
	}

	if s.Admins == nil {
		return errors.New("admins is nil")
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
