package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Scalify/puppet-master-gateway/pkg/api"
	internalTesting "github.com/Scalify/puppet-master-gateway/pkg/internal/testing"
)

func TestServerStartUnauthorizedJobs(t *testing.T) {
	q := internalTesting.NewTestQueue()
	db := internalTesting.NewTestDB()
	_, l := internalTesting.NewTestLogger()

	s, err := NewServer(db, q, l, "test")
	if err != nil {
		t.Log(err)
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(10*time.Second))
	defer cancel()
	go s.Start(ctx, 0)
	time.Sleep(10 * time.Millisecond)

	req := httptest.NewRequest(http.MethodPost, "/jobs", bytes.NewBuffer([]byte("{}")))
	rw := httptest.NewRecorder()
	s.srv.Handler.ServeHTTP(rw, req)

	t.Logf("Jobs response: %q", rw.Body.String())

	if rw.Body.String() != "Unauthorized." {
		t.Errorf("Unexpected healthz response: %q", rw.Body.String())
	}
}

func TestServerStartUnauthorizedHealthz(t *testing.T) {
	q := internalTesting.NewTestQueue()
	db := internalTesting.NewTestDB()
	_, l := internalTesting.NewTestLogger()

	s, err := NewServer(db, q, l, "test")
	if err != nil {
		t.Log(err)
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(10*time.Second))
	defer cancel()
	go s.Start(ctx, 0)
	time.Sleep(10 * time.Millisecond)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rw := httptest.NewRecorder()
	s.srv.Handler.ServeHTTP(rw, req)

	t.Logf("Healthz response: %q", rw.Body.String())

	if rw.Body.String() != "ok" {
		t.Errorf("Unexpected healthz response: %q", rw.Body.String())
	}
}

func TestServerCreateJob(t *testing.T) {
	q := internalTesting.NewTestQueue()
	db := internalTesting.NewTestDB()
	_, l := internalTesting.NewTestLogger()

	s, err := NewServer(db, q, l, "test")
	if err != nil {
		t.Log(err)
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(10*time.Second))
	defer cancel()
	go s.Start(ctx, 0)
	time.Sleep(100 * time.Millisecond)

	job, b := newTestJob(t, "asdf-1234-asdf-1234")
	req := httptest.NewRequest(http.MethodPost, "/jobs", bytes.NewReader(b))
	addApiTokenHeader(req, "test")
	rw := httptest.NewRecorder()
	s.srv.Handler.ServeHTTP(rw, req)

	if rw.Code != 200 {
		t.Errorf("Unexpected http response: %v", rw.Result().Status)
	}

	t.Logf("CreateJob response: %q", rw.Body.String())

	if rw.Body.String() == "" {
		t.Errorf("Unexpected CreateJob response: %q", rw.Body.String())
	}

	t.Logf("length Jobs: %v", len(db.SavedJobs))
	if len(db.SavedJobs) != 1 {
		t.Fatalf("Unexpected count of sent Jobs: %d", len(db.SavedJobs))
	}

	t.Logf("message in db: %+v", db.SavedJobs[0])
	if db.SavedJobs[0].UUID != job.UUID {
		t.Fatalf("Jobs are not equal: %+v , %+v", job.UUID, db.SavedJobs[0].UUID)
	}

	responseJob := &api.JobResponse{}
	if err := json.Unmarshal(rw.Body.Bytes(), responseJob); err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}

	if responseJob.Data.UUID != job.UUID {
		t.Fatalf("Jobs are not equal: %+v , %+v", job.UUID, responseJob.Data.UUID)
	}
}

func TestServerGetJob(t *testing.T) {
	q := internalTesting.NewTestQueue()
	db := internalTesting.NewTestDB()
	_, l := internalTesting.NewTestLogger()

	s, err := NewServer(db, q, l, "test")
	if err != nil {
		t.Log(err)
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(10*time.Second))
	defer cancel()
	go s.Start(ctx, 0)
	time.Sleep(100 * time.Millisecond)

	job, _ := newTestJob(t, "asdf-1234-asdf-1234")
	db.Jobs = append(db.Jobs, job)

	req := httptest.NewRequest(http.MethodGet, "/jobs/asdf-1234-asdf-1234", nil)
	addApiTokenHeader(req, "test")
	rw := httptest.NewRecorder()
	s.srv.Handler.ServeHTTP(rw, req)

	if rw.Code != 200 {
		t.Errorf("Unexpected http response: %v", rw.Result().Status)
	}

	t.Logf("GetJob response: %q", rw.Body.String())

	if rw.Body.String() == "" {
		t.Errorf("Unexpected GetJob response: %q", rw.Body.String())
	}

	responseJob := &api.JobResponse{}
	if err := json.Unmarshal(rw.Body.Bytes(), responseJob); err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}

	if responseJob.Data.UUID != job.UUID {
		t.Fatalf("Jobs are not equal: %+v , %+v", job.UUID, responseJob.Data.UUID)
	}
}

func TestServerGetJobs(t *testing.T) {
	q := internalTesting.NewTestQueue()
	db := internalTesting.NewTestDB()
	_, l := internalTesting.NewTestLogger()

	s, err := NewServer(db, q, l, "test")
	if err != nil {
		t.Log(err)
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(10*time.Second))
	defer cancel()
	go s.Start(ctx, 0)
	time.Sleep(100 * time.Millisecond)

	job1, _ := newTestJob(t, "asdf-1234-asdf-1234")
	job2, _ := newTestJob(t, "asdf-5678-asdf-5678")
	db.Jobs = append(db.Jobs, job1, job2)

	req := httptest.NewRequest(http.MethodGet, "/jobs", nil)
	addApiTokenHeader(req, "test")
	rw := httptest.NewRecorder()
	s.srv.Handler.ServeHTTP(rw, req)

	if rw.Code != 200 {
		t.Errorf("Unexpected http response: %v", rw.Result().Status)
	}

	t.Logf("GetJob response: %q", rw.Body.String())

	if rw.Body.String() == "" {
		t.Errorf("Unexpected GetJob response: %q", rw.Body.String())
	}

	responseJobs := &api.JobsResponse{}
	if err := json.Unmarshal(rw.Body.Bytes(), responseJobs); err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}

	t.Logf("length retrieved Jobs: %v", len(responseJobs.Data))
	if len(responseJobs.Data) != 2 {
		t.Fatalf("Unexpected count of retrieved Jobs: %d", len(responseJobs.Data))
	}

	if responseJobs.Data[0].UUID != job1.UUID {
		t.Fatalf("Jobs are not equal: %+v , %+v", job1.UUID, responseJobs.Data[0].UUID)
	}

	if responseJobs.Data[1].UUID != job2.UUID {
		t.Fatalf("Jobs are not equal: %+v , %+v", job2.UUID, responseJobs.Data[1].UUID)
	}
}

func TestServerDeleteJob(t *testing.T) {
	q := internalTesting.NewTestQueue()
	db := internalTesting.NewTestDB()
	_, l := internalTesting.NewTestLogger()

	s, err := NewServer(db, q, l, "test")
	if err != nil {
		t.Log(err)
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(10*time.Second))
	defer cancel()
	go s.Start(ctx, 0)
	time.Sleep(100 * time.Millisecond)

	job, _ := newTestJob(t, "asdf-1234-asdf-1234")
	db.Jobs = append(db.Jobs, job)

	req := httptest.NewRequest(http.MethodDelete, "/jobs/asdf-1234-asdf-1234", nil)
	addApiTokenHeader(req, "test")
	rw := httptest.NewRecorder()
	s.srv.Handler.ServeHTTP(rw, req)

	if rw.Code != 204 {
		t.Errorf("Unexpected http response: %v", rw.Result().Status)
	}

	t.Logf("DeleteJob response: %q", rw.Body.String())

	if rw.Body.String() != "" {
		t.Errorf("Unexpected DeleteJob response: %q", rw.Body.String())
	}

	t.Logf("length deleted Jobs: %v", len(db.DeletedJobs))
	if len(db.DeletedJobs) != 1 {
		t.Fatalf("Unexpected count of sent Jobs: %d", len(db.DeletedJobs))
	}

	t.Logf("message deleted in db: %+v", db.DeletedJobs[0])
	if db.DeletedJobs[0].UUID != job.UUID {
		t.Fatalf("Jobs are not equal: %+v , %+v", job.UUID, db.DeletedJobs[0].UUID)
	}
}
