package gateway

import "encoding/json"

// A Job is executed by the executor and stored in the database and holds all information
// required to let the puppets dance in the browser
type Job struct {
	ID        string            `json:"id"`
	Rev       string            `json:"-"`
	Code      string            `json:"code"`
	Variables map[string]string `json:"variables"`
	ExitCode  int               `json:"exit_code"`
	Logs      json.RawMessage   `json:"logs"`
}
