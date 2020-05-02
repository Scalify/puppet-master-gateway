package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/streadway/amqp"

	"github.com/scalify/puppet-master-gateway/pkg/api"
	"github.com/scalify/puppet-master-gateway/pkg/database"
)

func (s *Server) ensureQueues() error {
	var err error
	var queues = []string{api.QueueNameJobs, api.QueueNameJobResults}

	for _, queueName := range queues {
		_, err = s.queue.QueueDeclare(queueName, true, false, false, false, nil)
		if err != nil {
			return fmt.Errorf("unable to create queue %s: %v", queueName, err)
		}

		s.logger.Debugf("Checked queue %s for existence.", queueName)
	}

	return nil
}

func (s *Server) publishNewJob(job *api.Job) error {
	b, err := json.Marshal(job)
	if err != nil {
		return err
	}

	return s.queue.Publish("", api.QueueNameJobs, false, false, amqp.Publishing{
		ContentType: api.ContentTypeJSON,
		Body:        b,
	})
}

func (s *Server) consumeJobResults(ctx context.Context) {
	if err := s.queue.Qos(1, 0, false); err != nil {
		s.logger.Fatalf("Failed to set queue QOS: %v", err)
	}

	consumer, err := s.queue.Consume(api.QueueNameJobResults, "coordinator", false, false, false, false, nil)
	if err != nil {
		s.logger.Fatalf("Failed to create queue consumer: %v", err)
		return
	}

	var msg amqp.Delivery

	for {
		select {
		case <-ctx.Done():
			return
		case msg = <-consumer:
			if string(msg.Body) == "" {
				continue
			}

			s.handleJobResult(msg)
		}
	}
}

func (s *Server) nack(msg amqp.Delivery, requeue bool) {
	if err := msg.Nack(false, requeue); err != nil {
		s.logger.Errorf("Failed to nack message (requeue=%v): %v", err, requeue)
	}
}

func (s *Server) ack(msg amqp.Delivery) {
	if err := msg.Ack(false); err != nil {
		s.logger.Errorf("Failed to aack message: %v", err)
	}
}

func (s *Server) handleJobResult(msg amqp.Delivery) {
	s.logger.Debugf("Consuming message from queue: %v", string(msg.Body))

	var result api.JobResult
	if err := json.Unmarshal(msg.Body, &result); err != nil {
		s.logger.Errorf("Failed to unmarshal json body: %v", err)
		return
	}

	if result.UUID == "" {
		s.logger.Errorf("Failed to process job result: object has no UUID")
		s.nack(msg, false)
		return
	}

	l := s.loggerForJob(result.UUID)
	l.Debugf("Loading job from db")

	job, err := s.db.Get(result.UUID)
	if err != nil {
		if err == database.ErrNotFound {
			l.Errorf("Job %q does not exist in DB, skipping.", result.UUID)
			s.ack(msg)
			return
		}

		l.Errorf("Failed to load job from db: %v", err)
		s.nack(msg, true)
		return
	}

	if job.Status == api.JobStatusDone {
		l.Error("Consumed job result was already persisted - at least the job has the status == done.")
		s.nack(msg, false)
		return
	}

	job.Status = api.JobStatusDone
	job.Logs = result.Logs
	job.Error = result.Error
	job.Results = result.Results
	job.StartedAt = result.StartedAt
	job.FinishedAt = result.FinishedAt
	job.Duration = result.Duration

	if err := s.db.Save(job); err != nil {
		l.Errorf("Failed to save job back to db: %v", err)
		s.nack(msg, true)
		return
	}

	if err := msg.Ack(false); err != nil {
		l.Errorf("Failed to ack message: %v", err)
		return
	}

	l.Debugf("Done processing job result")
}

func (s *Server) produceJobs(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}

		jobs, err := s.db.GetListByStatus(api.JobStatusCreated, 1, 100)
		if err != nil {
			s.logger.Errorf("Failed to get created jobs: %v", err)
		}

		s.logger.Debugf("Got %d created jobs from db.", len(jobs))

		for _, job := range jobs {
			l := s.loggerForJob(job.UUID)
			if err := s.publishNewJob(job); err != nil {
				if err == amqp.ErrClosed {
					l.Fatalf("amqp connection is closed, aborting.")
				}

				l.Errorf("Failed to queue job: %v", err)
				continue
			}

			job.Status = api.JobStatusQueued
			if err := s.db.Save(job); err != nil {
				l.Errorf("Failed to save updated job: %v", err)
			}

			l.Debugf("Queued job.")
		}
	}
}
