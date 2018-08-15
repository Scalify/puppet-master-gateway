package gateway

import (
	"net/http"
	"net/http/httptest"
	"testing"

	internalTesting "github.com/Scalify/puppet-master-gateway/pkg/internal/testing"
)

type testHandler struct {
	request bool
}

func (t *testHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	t.request = true
}

func TestBasicAuthNoAuth(t *testing.T) {
	rw := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	h := &testHandler{}
	_, l := internalTesting.NewTestLogger()
	b := newBasicAuth(l, "asdf", "asdfgh")

	b.Middleware(h).ServeHTTP(rw, req)

	if h.request {
		t.Errorf("Handler got request while not having authenticated")
	}

	if rw.Body.String() != "Unauthorized." {
		t.Errorf("Unexpected body %q", rw.Body.String())
	}

	if rw.Code != 401 {
		t.Errorf("Unexpected status code %q", rw.Code)
	}
}

func TestBasicAuthWrongAuth(t *testing.T) {
	rw := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	h := &testHandler{}
	_, l := internalTesting.NewTestLogger()
	b := newBasicAuth(l, "asdf", "asdfgh")

	req.SetBasicAuth("test", "test")
	b.Middleware(h).ServeHTTP(rw, req)

	if h.request {
		t.Errorf("Handler got request while not having authenticated")
	}

	if rw.Body.String() != "Unauthorized." {
		t.Errorf("Unexpected body %q", rw.Body.String())
	}

	if rw.Code != 401 {
		t.Errorf("Unexpected status code %q", rw.Code)
	}
}

func TestBasicAuth(t *testing.T) {
	rw := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	h := &testHandler{}
	_, l := internalTesting.NewTestLogger()
	b := newBasicAuth(l, "test", "test")

	req.SetBasicAuth("test", "test")
	b.Middleware(h).ServeHTTP(rw, req)

	if !h.request {
		t.Errorf("Handler got no request")
	}

	if rw.Code != 200 {
		t.Errorf("Unexpected status code %q", rw.Result().Status)
	}
}
