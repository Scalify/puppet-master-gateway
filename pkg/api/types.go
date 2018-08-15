package api

import (
	"reflect"
	"time"
)

// A Job is executed by the executor and stored in the database and holds all information
// required to let the puppets dance in the browser
type Job struct {
	UUID       string                 `json:"uuid"`
	Rev        string                 `json:"_rev,omitempty"`
	Code       string                 `json:"code"`
	Status     string                 `json:"status"`
	Vars       map[string]string      `json:"vars"`
	Modules    map[string]string      `json:"modules"`
	Error      string                 `json:"error"`
	Logs       []Log                  `json:"logs"`
	Results    map[string]interface{} `json:"results"`
	CreatedAt  time.Time              `json:"created_at"`
	StartedAt  *time.Time             `json:"started_at"`
	FinishedAt *time.Time             `json:"finished_at"`
	Duration   int                    `json:"duration"`
}

// NewJob creates a new Job instance
func NewJob() *Job {
	return &Job{
		Vars:    make(map[string]string),
		Modules: make(map[string]string),
	}
}

// Equal returns true when both given Jobs are equal
// nolint: gocyclo
func (j *Job) Equal(j2 *Job) bool {
	return j.UUID == j2.UUID &&
		j.Code == j2.Code &&
		j.Status == j2.Status &&
		reflect.DeepEqual(j.Modules, j2.Modules) &&
		reflect.DeepEqual(j.Vars, j2.Vars) &&
		datesAreEqual(&j.CreatedAt, &j2.CreatedAt) &&
		datesAreEqual(j.StartedAt, j2.StartedAt) &&
		datesAreEqual(j.FinishedAt, j2.FinishedAt) &&
		j.Error == j2.Error &&
		reflect.DeepEqual(j.Results, j2.Results) &&
		reflect.DeepEqual(j.Logs, j2.Logs)
}

// A JobResult is emitted after a worker did the job and synced to database
type JobResult struct {
	UUID       string                 `json:"uuid"`
	Error      string                 `json:"error"`
	Logs       []Log                  `json:"logs"`
	Results    map[string]interface{} `json:"results"`
	StartedAt  *time.Time             `json:"started_at"`
	FinishedAt *time.Time             `json:"finished_at"`
	Duration   int                    `json:"duration"`
}

// NewJobResult creates a new JobResult instance
func NewJobResult() *JobResult {
	return &JobResult{
		Results: make(map[string]interface{}),
		Logs:    make([]Log, 0),
	}
}

// Equal returns true when both given JobResults are equal
func (j *JobResult) Equal(j2 *JobResult) bool {
	return j.UUID == j2.UUID &&
		datesAreEqual(j.StartedAt, j2.StartedAt) &&
		datesAreEqual(j.FinishedAt, j2.FinishedAt) &&
		j.Error == j2.Error &&
		reflect.DeepEqual(j.Results, j2.Results) &&
		reflect.DeepEqual(j.Logs, j2.Logs)
}

func datesAreEqual(t1 *time.Time, t2 *time.Time) bool {
	if (t1 == nil && t2 != nil) || (t1 != nil && t2 == nil) {
		return false
	}

	if t1 == nil && t2 == nil {
		return true
	}

	return (*t1).Equal(*t2)
}

// A Log represents a log line
type Log struct {
	Time    time.Time `json:"time"`
	Level   string    `json:"level"`
	Message string    `json:"message"`
}

// JobResponse is the wrapper around a job when returned through API
type JobResponse struct {
	Data *Job `json:"data"`
}
