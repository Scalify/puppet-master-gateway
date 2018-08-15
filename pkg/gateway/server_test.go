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

func TestServerStart(t *testing.T) {
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
	req.SetBasicAuth("test", "test")
	rw := httptest.NewRecorder()
	s.srv.Handler.ServeHTTP(rw, req)

	if rw.Code != 204 {
		t.Errorf("Unexpected healthz response: %v", rw.Code)
	}
}

func TestServerShutdown(t *testing.T) {
	q := internalTesting.NewTestQueue()
	db := internalTesting.NewTestDB()
	_, l := internalTesting.NewTestLogger()

	s, err := NewServer(db, q, l, "test", "test")
	if err != nil {
		t.Log(err)
		t.Fatal(err)
	}

	ctx := context.Background()
	go s.Start(ctx, 0)
	time.Sleep(100 * time.Millisecond)

	if err := s.Shutdown(ctx); err != nil {
		t.Log(err)
		t.Fatal(err)
	}
}