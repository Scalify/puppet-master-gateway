package gateway

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/Scalify/puppet-master-gateway/pkg/api"
	"github.com/Scalify/puppet-master-gateway/pkg/database"
	"github.com/gorilla/mux"
	"github.com/satori/go.uuid"
)

func (s *Server) getQueryParamAsInt(req *http.Request, param string, defaultValue int) (int, error) {
	str := req.URL.Query().Get(param)
	if str == "" {
		return defaultValue, nil
	}

	return strconv.Atoi(str)
}

func (s *Server) getPageAndPerPage(req *http.Request) (int, int, error) {
	page, err := s.getQueryParamAsInt(req, "page", 1)
	if err != nil {
		return 0, 0, err
	}

	perPage, err := s.getQueryParamAsInt(req, "per_page", 10)
	if err != nil {
		return 0, 0, err
	}

	return page, perPage, nil
}

// CreateJob stores a job in the database and starts a job worker for it
func (s *Server) CreateJob(rw http.ResponseWriter, req *http.Request) {
	rw.Header().Add(api.ContentTypeHeader, api.ContentTypeJSON)

	job := api.NewJob()
	if err := json.NewDecoder(req.Body).Decode(job); err != nil {
		s.logger.Errorf("Failed to decode json body: %v", err)
		rw.WriteHeader(http.StatusBadRequest)
		if _, errw := fmt.Fprintf(rw, jsonErrFailedToDecodeBody, err); errw != nil {
			s.logger.Error(errw)
		}
		return
	}

	job.Status = api.JobStatusCreated
	job.CreatedAt = api.JSONTime{Time: time.Now()}
	if job.UUID == "" {
		job.UUID = uuid.NewV4().String()
	}

	logger := s.logger.WithField(api.LogFieldJobID, job.UUID)

	if err := s.db.Save(job); err != nil {
		logger.Errorf("Failed to save job: %v", err)
		rw.WriteHeader(http.StatusInternalServerError)
		if _, errw := fmt.Fprintf(rw, jsonErrFailedToSaveJob, err); errw != nil {
			logger.Error(errw)
		}
		return
	}

	job.Rev = ""
	jobResponse := &api.JobResponse{Data: job}
	if err := json.NewEncoder(rw).Encode(jobResponse); err != nil {
		logger.Errorf("Failed to encode job: %v", err)
	}
}

// GetJobs returns a paginated list of jobs
func (s *Server) GetJobs(rw http.ResponseWriter, req *http.Request) {
	rw.Header().Add(api.ContentTypeHeader, api.ContentTypeJSON)

	page, perPage, err := s.getPageAndPerPage(req)
	if err != nil {
		s.logger.Errorf("Failed to get request params page and per_page from request: %v", err)
		rw.WriteHeader(http.StatusInternalServerError)
		if _, errw := fmt.Fprintf(rw, jsonErrFailedToFetchJobs, err); errw != nil {
			s.logger.Error(errw)
		}
		return
	}

	var jobs []*api.Job
	status := req.URL.Query().Get("status")
	if status != "" {
		jobs, err = s.db.GetListByStatus(status, page, perPage)
	} else {
		jobs, err = s.db.GetList(page, perPage)
	}

	if err != nil {
		s.logger.Errorf("Failed to load jobs: %v", err)
		rw.WriteHeader(http.StatusInternalServerError)
		if _, errw := fmt.Fprintf(rw, jsonErrFailedToFetchJobs, err); errw != nil {
			s.logger.Error(errw)
		}
		return
	}

	for i := range jobs {
		jobs[i].Rev = ""
	}

	jobsResponse := &api.JobsResponse{Data: jobs}
	if err := json.NewEncoder(rw).Encode(jobsResponse); err != nil {
		s.logger.Errorf("Failed to encode job: %v", err)
		if _, errw := fmt.Fprintf(rw, jsonErrFailedToFetchJobs, err); errw != nil {
			s.logger.Error(errw)
		}
	}

	s.logger.Debugf("Loaded job from database and sent to client")
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
			if _, errw := fmt.Fprintf(rw, jsonErrJobNotFound, jobID, err); errw != nil {
				s.logger.Error(errw)
			}
			return
		}

		logger.Errorf("Failed to load job: %v", err)
		rw.WriteHeader(http.StatusInternalServerError)
		if _, errw := fmt.Fprintf(rw, jsonErrFailedToFetchJob, err); errw != nil {
			s.logger.Error(errw)
		}
		return
	}

	job.Rev = ""
	jobResponse := &api.JobResponse{Data: job}
	if err := json.NewEncoder(rw).Encode(jobResponse); err != nil {
		logger.Errorf("Failed to encode job: %v", err)
		if _, errw := fmt.Fprintf(rw, jsonErrFailedToFetchJobs, err); errw != nil {
			s.logger.Error(errw)
		}
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
			if _, errw := fmt.Fprintf(rw, jsonErrJobNotFound, jobID, err); errw != nil {
				s.logger.Error(errw)
			}
			return
		}

		logger.Errorf("Failed to load job: %v", err)
		rw.WriteHeader(http.StatusInternalServerError)
		if _, errw := fmt.Fprintf(rw, jsonErrFailedToFetchJob, err); errw != nil {
			s.logger.Error(errw)
		}
		return
	}

	if err := s.db.Delete(job); err != nil {
		logger.Errorf("Failed to delete job: %v", err)
		rw.WriteHeader(http.StatusInternalServerError)
		if _, errw := fmt.Fprintf(rw, jsonErrFailedToDeleteJob, err); errw != nil {
			s.logger.Error(errw)
		}
		return
	}

	rw.WriteHeader(http.StatusNoContent)
}
