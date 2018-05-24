package coordinator

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/streadway/amqp"
	"gitlab.com/scalifyme/puppet-master/puppet-master/pkg/api"
)

// Coordinator produces queue job messages and consumes the results, like coordinating the puppets
type Coordinator struct {
	logger *logrus.Entry
	db     db
	queue  queue
}

// New returns a new coordinator instance
func New(logger *logrus.Entry, db db, queue queue) (*Coordinator, error) {
	coordinator := &Coordinator{
		logger: logger,
		db:     db,
		queue:  queue,
	}

	if err := coordinator.ensureQueues(); err != nil {
		return nil, err
	}

	return coordinator, nil
}

func (c *Coordinator) ensureQueues() error {
	var err error
	var queues = []string{api.QueueNameJobs, api.QueueNameJobResults}

	for _, queueName := range queues {
		_, err = c.queue.QueueDeclare(queueName, true, false, false, false, nil)
		if err != nil {
			return fmt.Errorf("unable to create queue %s: %v", queueName, err)
		}

		c.logger.Debugf("Checked queue %s for existence.", queueName)
	}

	return nil
}

func (c *Coordinator) publishNewJob(job *api.Job) error {
	b, err := json.Marshal(job)
	if err != nil {
		return err
	}

	return c.queue.Publish("", api.QueueNameJobs, false, false, amqp.Publishing{
		ContentType: api.ContentTypeJSON,
		Body:        b,
	})
}

// Start starts the coordinator processes
func (c *Coordinator) Start(ctx context.Context) error {
	go c.consumeJobResults(ctx)
	go c.produceJobs(ctx)

	<-ctx.Done()

	return nil
}

func (c *Coordinator) consumeJobResults(ctx context.Context) {
	logger := c.logger.WithField(task, taskConsumeJobResults)

	if err := c.queue.Qos(1, 0, false); err != nil {
		logger.Fatalf("Failed to set queue QOS: %v", err)
	}

	consumer, err := c.queue.Consume(api.QueueNameJobResults, "coordinator", false, false, false, false, nil)
	if err != nil {
		logger.Fatalf("Failed to create queue consumer: %v", err)
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

			c.handleJobResult(logger, msg)
		}
	}
}

func (c *Coordinator) handleJobResult(logger *logrus.Entry, msg amqp.Delivery) {
	logger.Debugf("Consuming message from queue: %v", string(msg.Body))

	var result api.JobResult
	if err := json.Unmarshal(msg.Body, &result); err != nil {
		logger.Errorf("Failed to unmarshal json body: %v", err)
		return
	}

	if result.JobID == "" {
		logger.Errorf("Failed to process job result: object has no jobID")
		if err := msg.Ack(false); err != nil {
			logger.Errorf("Failed to ack message: %v", err)
		}
		return
	}

	l := logger.WithField(api.LogFieldJobID, result.JobID)
	l.Debugf("Loading job from db")

	job, err := c.db.Get(result.JobID)
	if err != nil {
		l.Errorf("Failed to load job from db: %v", err)
		return
	}

	if job.Status != api.JobStatusQueued {
		l.Errorf("Job jas unexpected status %s, expected %s", job.Status, api.JobStatusQueued)
		return
	}

	job.Status = api.JobStatusDone
	job.Logs = result.Logs
	job.Error = result.Error
	job.Results = result.Results
	job.FinishedAt = result.FinishedAt
	job.Duration = result.Duration

	if err := c.db.Save(job); err != nil {
		l.Errorf("Failed to save job back to db: %v", err)
	}

	if err := msg.Ack(false); err != nil {
		l.Errorf("Failed to ack message: %v", err)
		return
	}

	l.Debugf("Done processing job result")
}

func (c *Coordinator) produceJobs(ctx context.Context) {
	logger := c.logger.WithField(task, taskProduceJobs)
	ticker := time.NewTicker(10 * time.Second)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}

		jobs, err := c.db.GetByStatus(api.JobStatusNew, 10)
		if err != nil {
			logger.Fatalf("Failed to get jobs in status new: %v", err)
		}

		logger.Debugf("Got %d jobs in state new from db.", len(jobs))

		for _, job := range jobs {
			l := logger.WithField(api.LogFieldJobID, job.ID)
			if err := c.publishNewJob(job); err != nil {
				l.Errorf("Failed to queue job: %v", err)
				continue
			}

			job.Status = api.JobStatusQueued
			if err := c.db.Save(job); err != nil {
				l.Errorf("Failed to save job to db: %v", err)
			}

			l.Debugf("Queued job.")
		}
	}
}
