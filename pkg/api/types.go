package api

import (
	"time"
)

// A Job is executed by the executor and stored in the database and holds all information
// required to let the puppets dance in the browser
type Job struct {
	ID         string                 `json:"id"`
	Rev        string                 `json:"_rev,omitempty"`
	Code       string                 `json:"code"`
	Status     string                 `json:"status"`
	Vars       map[string]string      `json:"vars"`
	Modules    map[string]string      `json:"modules"`
	Error      string                 `json:"error"`
	Logs       []Log                  `json:"logs"`
	Results    map[string]interface{} `json:"results"`
	CreatedAt  time.Time              `json:"created_at"`
	FinishedAt *time.Time             `json:"finished_at"`
	Duration   int                    `json:"duration"`
}

// NewJob creates a new Job instance
func NewJob() *Job {
	return &Job{
		Vars:    make(map[string]string),
		Modules: make(map[string]string),
		Results: make(map[string]interface{}),
		Logs:    make([]Log, 0),
	}
}

// A JobResult is emitted after a worker did the job and synced to database
type JobResult struct {
	JobID      string                 `json:"job_id"`
	Error      string                 `json:"error"`
	Logs       []Log                  `json:"logs"`
	Results    map[string]interface{} `json:"results"`
	FinishedAt *time.Time             `json:"finished_at"`
	Duration   int                    `json:"duration"`
}

// A Log represents a log line
type Log struct {
	Time    time.Time `json:"time"`
	Level   string    `json:"level"`
	Message string    `json:"message"`
}
