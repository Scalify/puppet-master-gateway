package api

import "encoding/json"

// A Job is executed by the executor and stored in the database and holds all information
// required to let the puppets dance in the browser
type Job struct {
	ID      string            `json:"id"`
	Rev     string            `json:"_rev,omitempty"`
	Code    string            `json:"code"`
	Status  string            `json:"status"`
	Vars    map[string]string `json:"vars"`
	Modules map[string]string `json:"modules"`
	Error   string            `json:"error,omitempty"`
	Logs    json.RawMessage   `json:"logs,omitempty"`
}

// A JobResult is emitted after a worker did the job and synced to database
type JobResult struct {
	JobID string          `json:"job_id"`
	Error string          `json:"error"`
	Logs  json.RawMessage `json:"logs"`
}
