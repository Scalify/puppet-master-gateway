package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/streadway/amqp"
	"gitlab.com/scalifyme/puppet-master/gateway/pkg/gateway"
)

const (
	ContentTypeHeader = "Content-Type"
	ContentTypeJSON   = "application/json"

	jsonErrFailedToDecodeBody = "{\"error\":\"Failed to decode json body: \"%q\"\"}"
	jsonErrFailedToFetchJob   = "{\"error\":\"Failed to fetch job: \"%q\"\"}"
	jsonErrFailedToSaveJob    = "{\"error\":\"Failed to save job: \"%q\"\"}"
)

type db interface {
	Get(id string) (*gateway.Job, error)
	Save(job *gateway.Job) (error)
	Delete(job *gateway.Job) (error)
}

type queueConn interface {
	QueueDeclare(name string, durable, autoDelete, exclusive, noWait bool, args amqp.Table) (amqp.Queue, error)
	Consume(queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp.Table) (<-chan amqp.Delivery, error)
	Publish(exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error
}

type Server struct {
	logger *logrus.Entry
	db     db
	queue  queueConn
}

// NewServer creates a new server
func NewServer(db db, queue queueConn, logger *logrus.Entry) (*Server, error) {
	server := &Server{
		logger: logger,
		db:     db,
		queue:  queue,
	}

	if err := server.ensureQueues(); err != nil {
		return nil, err
	}

	return server, nil
}

func (s *Server) ensureQueues() error {
	var err error
	var queues = []string{"puppet-master-jobs", "puppet-master-job-results"}

	for _, queueName := range queues {
		_, err = s.queue.QueueDeclare(queueName, true, false, false, false, nil)
		if err != nil {
			return fmt.Errorf("unable to create queue %s: %v", queueName, err)
		}

		s.logger.Debugf("Checked queue %s for existence.", queueName)
	}

	return nil
}

func (s *Server) Serve(listenPort int) {
	r := mux.NewRouter()
	r.Methods(http.MethodPost).Path("/jobs").HandlerFunc(s.CreateJob)
	r.Methods(http.MethodGet).Path("/jobs/{id}").HandlerFunc(s.GetJob)
	r.Methods(http.MethodGet).Path("/_healthz").HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {})

	srv := &http.Server{
		Addr:         fmt.Sprintf("0.0.0.0:%d", listenPort),
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			s.logger.Fatal(err)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGKILL, syscall.SIGTERM)

	<-c

	ctx, cancel := context.WithTimeout(context.Background(), 10)
	defer cancel()
	srv.Shutdown(ctx)
	s.logger.Info("shutting down")
	os.Exit(0)
}

// CreateJob stores a job in the database and starts a job worker for it
func (s *Server) CreateJob(rw http.ResponseWriter, req *http.Request) {
	rw.Header().Add(ContentTypeHeader, ContentTypeJSON)

	var job *gateway.Job
	if err := json.NewDecoder(req.Body).Decode(job); err != nil {
		s.logger.Errorf("Failed to decode json body: %v", err)
		rw.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(rw, jsonErrFailedToDecodeBody, err)
		return
	}

	if err := s.db.Save(job); err != nil {
		s.logger.Errorf("Failed to save job: %v", err)
		rw.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(rw, jsonErrFailedToSaveJob, err)
		return
	}

	if err := json.NewEncoder(rw).Encode(job); err != nil {
		s.logger.Errorf("Failed to encode job: %v", err)
	}
}

// GetJob reads the job from the database and returns it
func (s *Server) GetJob(rw http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	jobID := vars["id"]
	var job *gateway.Job

	job, err := s.db.Get(jobID)
	if err != nil {
		s.logger.Errorf("Failed to load job: %v", err)
		rw.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(rw, jsonErrFailedToFetchJob, err)
		return
	}

	if err := json.NewEncoder(rw).Encode(job); err != nil {
		s.logger.Errorf("Failed to encode job: %v", err)
	}
}
