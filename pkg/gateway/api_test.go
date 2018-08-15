package gateway

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	internalTesting "github.com/Scalify/puppet-master-gateway/pkg/internal/testing"
)

func TestServerStartUnauthorizedJobs(t *testing.T) {
	q := internalTesting.NewTestQueue()
	db := internalTesting.NewTestDB()
	_, l := internalTesting.NewTestLogger()

	s, err := NewServer(db, q, l, "test", "test")
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

	s, err := NewServer(db, q, l, "test", "test")
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

	s, err := NewServer(db, q, l, "test", "test")
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
	req.SetBasicAuth("test", "test")
	rw := httptest.NewRecorder()
	s.srv.Handler.ServeHTTP(rw, req)

	if rw.Code != 204 {
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
}

func TestServerGetJob(t *testing.T) {
	q := internalTesting.NewTestQueue()
	db := internalTesting.NewTestDB()
	_, l := internalTesting.NewTestLogger()

	s, err := NewServer(db, q, l, "test", "test")
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
	req.SetBasicAuth("test", "test")
	rw := httptest.NewRecorder()
	s.srv.Handler.ServeHTTP(rw, req)

	if rw.Code != 200 {
		t.Errorf("Unexpected http response: %v", rw.Result().Status)
	}

	t.Logf("GetJob response: %q", rw.Body.String())

	if rw.Body.String() == "" {
		t.Errorf("Unexpected GetJob response: %q", rw.Body.String())
	}
}

func TestServerDeleteJob(t *testing.T) {
	q := internalTesting.NewTestQueue()
	db := internalTesting.NewTestDB()
	_, l := internalTesting.NewTestLogger()

	s, err := NewServer(db, q, l, "test", "test")
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
	req.SetBasicAuth("test", "test")
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
