package gateway

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/scalify/puppet-master-gateway/pkg/api"
)

var testTime = api.JSONTime{Time: time.Now()}

func newTestJob(t *testing.T, uuid string) (*api.Job, []byte) {
	job := api.NewJob()
	job.UUID = uuid
	job.Code = "test-" + uuid
	job.CreatedAt = testTime
	job.Modules = map[string]string{
		"1234": "1234567890",
	}

	b, err := json.Marshal(job)
	if err != nil {
		t.Log(err)
		t.Fatal(err)
	}

	return job, b
}

func newTestJobResult(t *testing.T, uuid string) (*api.JobResult, []byte) {
	res := api.NewJobResult()
	res.UUID = uuid
	res.Error = "test-" + uuid
	res.StartedAt = &testTime
	res.FinishedAt = &testTime
	res.Results = map[string]interface{}{
		"1234": "1234567890",
	}

	b, err := json.Marshal(res)
	if err != nil {
		t.Log(err)
		t.Fatal(err)
	}

	return res, b
}
