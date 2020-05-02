package gateway

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/aklinkert/go-logging"

	internalTesting "github.com/scalify/puppet-master-gateway/pkg/internal/testing"
)

func TestServerStart(t *testing.T) {
	q := internalTesting.NewTestQueue()
	db := internalTesting.NewTestDB()
	l := logging.NewTestLogger(t)

	s, err := NewServer(db, q, l, "test", true, true)
	if err != nil {
		t.Log(err)
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(10*time.Second))
	defer cancel()
	go func() {
		if err := s.Start(ctx, 0); err != nil {
			l.Fatal(fmt.Errorf("failed to start server: %v", err))
		}
	}()

	time.Sleep(10 * time.Millisecond)
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rw := httptest.NewRecorder()
	s.srv.Handler.ServeHTTP(rw, req)

	if rw.Code != 200 {
		t.Errorf("Unexpected healthz response: %v", rw.Code)
	}
}

func TestServerShutdown(t *testing.T) {
	q := internalTesting.NewTestQueue()
	db := internalTesting.NewTestDB()
	l := logging.NewTestLogger(t)

	s, err := NewServer(db, q, l, "test", true, true)
	if err != nil {
		t.Log(err)
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Millisecond)
	defer cancel()
	go startServer(ctx, t, s)
	time.Sleep(100 * time.Millisecond)

	if err := s.Shutdown(ctx); err != nil {
		t.Log(err)
		t.Fatal(err)
	}
}
