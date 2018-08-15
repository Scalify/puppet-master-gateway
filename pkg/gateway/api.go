package gateway

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Scalify/puppet-master-gateway/pkg/api"
	"github.com/Scalify/puppet-master-gateway/pkg/database"
	"github.com/gorilla/mux"
	"github.com/satori/go.uuid"
)

// CreateJob stores a job in the database and starts a job worker for it
func (s *Server) CreateJob(rw http.ResponseWriter, req *http.Request) {
	rw.Header().Add(api.ContentTypeHeader, api.ContentTypeJSON)

	job := api.NewJob()
	if err := json.NewDecoder(req.Body).Decode(job); err != nil {
		s.logger.Errorf("Failed to decode json body: %v", err)
		rw.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(rw, jsonErrFailedToDecodeBody, err)
		return
	}

	job.Status = api.JobStatusCreated
	job.CreatedAt = time.Now()
	if job.UUID == "" {
		job.UUID = uuid.NewV4().String()
	}

	logger := s.logger.WithField(api.LogFieldJobID, job.UUID)

	if err := s.db.Save(job); err != nil {
		logger.Errorf("Failed to save job: %v", err)
		rw.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(rw, jsonErrFailedToSaveJob, err)
		return
	}

	logger.Debugf("Wrote job to database")
	rw.WriteHeader(http.StatusNoContent)
	job.Rev = ""

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

// DeleteJob deletes a job from the database
func (s *Server) DeleteJob(rw http.ResponseWriter, req *http.Request) {
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

	if err := s.db.Delete(job); err != nil {
		logger.Errorf("Failed to delete job: %v", err)
		rw.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(rw, jsonErrFailedToDeleteJob, err)
		return
	}

	rw.WriteHeader(http.StatusNoContent)
}
