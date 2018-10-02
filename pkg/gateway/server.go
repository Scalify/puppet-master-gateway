package gateway

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
)

// Server is an http handler serving the puppet-master api
type Server struct {
	logger                *logrus.Entry
	db                    db
	queue                 queue
	srv                   *http.Server
	apiToken              string
	enableAPI, enableJobs bool
}

// NewServer creates a new server
func NewServer(db db, queue queue, logger *logrus.Entry, apiToken string, enableAPI, enableJobs bool) (*Server, error) {
	return &Server{
		logger:     logger,
		queue:      queue,
		db:         db,
		apiToken:   apiToken,
		enableAPI:  enableAPI,
		enableJobs: enableJobs,
	}, nil
}

// Start opens the http port and handles the requests
func (s *Server) Start(ctx context.Context, listenPort int) error {
	if s.enableJobs {
		go s.consumeJobResults(ctx)
		go s.produceJobs(ctx)
	}

	if s.enableAPI {
		if err := s.setupAPI(ctx, listenPort); err != nil {
			return err
		}

		return s.srv.ListenAndServe()
	}

	return nil
}

func (s *Server) setupAPI(ctx context.Context, listenPort int) error {
	r := mux.NewRouter()
	authHandler := newAuthHandler(s.logger, s.apiToken)

	jobs := r.PathPrefix("/jobs").Subrouter()
	jobs.Use(authHandler.Middleware)
	jobs.HandleFunc("", s.GetJobs).Methods(http.MethodGet)
	jobs.HandleFunc("", s.CreateJob).Methods(http.MethodPost)
	jobs.HandleFunc("/{id}", s.GetJob).Methods(http.MethodGet)
	jobs.HandleFunc("/{id}", s.DeleteJob).Methods(http.MethodDelete)

	r.Methods(http.MethodGet).Path("/healthz").HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		if _, err := fmt.Fprint(rw, "ok"); err != nil {
			s.logger.Errorf("Failed to send ok: %v", err)
		}
	})

	s.srv = &http.Server{
		Addr:         fmt.Sprintf("0.0.0.0:%d", listenPort),
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      r,
	}

	if err := s.ensureQueues(); err != nil {
		return err
	}

	go func() {
		<-ctx.Done()

		ctxCancel, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		if err := s.srv.Shutdown(ctxCancel); err != nil {
			s.logger.Errorf("Failed to shutdown server: %v", err)
		}
	}()

	return nil
}

// Shutdown closes the http server
func (s *Server) Shutdown(ctx context.Context) error {
	return s.srv.Shutdown(ctx)
}
