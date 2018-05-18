package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/satori/go.uuid"
	"gitlab.com/scalifyme/puppet-master/puppet-master/pkg/api"
	"gitlab.com/scalifyme/puppet-master/puppet-master/pkg/database"
)

// Server is an http handler serving the puppet-master api
type Server struct {
	logger *logrus.Entry
	db     db
	srv    *http.Server
}

// NewServer creates a new server
func NewServer(db db, logger *logrus.Entry) (*Server, error) {
	return &Server{
		logger: logger,
		db:     db,
	}, nil
}

// Start opens the http port and handles the requests
func (s *Server) Start(ctx context.Context, listenPort int) error {
	r := mux.NewRouter()
	jobs := r.PathPrefix("/jobs").Subrouter()
	jobs.Methods(http.MethodGet).Path("/{id}").HandlerFunc(s.GetJob)
	jobs.Methods(http.MethodPost).Path("/").HandlerFunc(s.CreateJob)

	r.Methods(http.MethodGet).Path("/_healthz").HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(rw, "ok")
	})

	s.srv = &http.Server{
		Addr:         fmt.Sprintf("0.0.0.0:%d", listenPort),
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      r,
	}

	if err := s.srv.ListenAndServe(); err != nil {
		return err
	}

	<-ctx.Done()

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	return s.srv.Shutdown(ctx)
}

// Shutdown closes the http server
func (s *Server) Shutdown(ctx context.Context) error {
	return s.srv.Shutdown(ctx)
}

// CreateJob stores a job in the database and starts a job worker for it
func (s *Server) CreateJob(rw http.ResponseWriter, req *http.Request) {
	rw.Header().Add(api.ContentTypeHeader, api.ContentTypeJSON)

	job := &api.Job{}
	if err := json.NewDecoder(req.Body).Decode(job); err != nil {
		s.logger.Errorf("Failed to decode json body: %v", err)
		rw.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(rw, jsonErrFailedToDecodeBody, err)
		return
	}

	job.Status = api.JobStatusNew
	job.ID = uuid.NewV4().String()
	logger := s.logger.WithField(api.LogFieldJobID, job.ID)

	if err := s.db.Save(job); err != nil {
		logger.Errorf("Failed to save job: %v", err)
		rw.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(rw, jsonErrFailedToSaveJob, err)
		return
	}

	logger.Debugf("Wrote job to database")

	if err := json.NewEncoder(rw).Encode(job); err != nil {
		logger.Errorf("Failed to encode job: %v", err)
	}
}

// GetJob reads the job from the database and returns it
func (s *Server) GetJob(rw http.ResponseWriter, req *http.Request) {
	rw.Header().Add(api.ContentTypeHeader, api.ContentTypeJSON)

	vars := mux.Vars(req)
	jobID := vars["id"]
	logger := s.logger.WithField(api.LogFieldJobID, jobID)

	job, err := s.db.Get(jobID)
	if err != nil {
		if err == database.ErrNotFound {
			logger.Debugf("Failed to find job in database")
			rw.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(rw, jsonErrJobNotFound, jobID, err)
			return
		}

		logger.Errorf("Failed to load job: %v", err)
		rw.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(rw, jsonErrFailedToFetchJob, err)
		return
	}

	job.Rev = ""

	if err := json.NewEncoder(rw).Encode(job); err != nil {
		logger.Errorf("Failed to encode job: %v", err)
	}

	logger.Debugf("Loaded job from database and sent to client")
}
