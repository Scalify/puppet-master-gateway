package gateway

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aklinkert/go-logging"
)

func addAPITokenHeader(r *http.Request, apiToken string) {
	r.Header.Add("Authorization", fmt.Sprintf("Bearer %s", apiToken))
}

type testHandler struct {
	request bool
}

func (t *testHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	t.request = true
}

func TestAuthHandlerNoAuth(t *testing.T) {
	rw := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	h := &testHandler{}
	l := logging.NewTestLogger(t)
	b := newAuthHandler(l, "asdf")

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

func TestAuthHandlerWrongAuth(t *testing.T) {
	rw := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	h := &testHandler{}
	l := logging.NewTestLogger(t)
	b := newAuthHandler(l, "asdf")

	addAPITokenHeader(req, "qwertz")
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

func TestAuthHandler(t *testing.T) {
	rw := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	h := &testHandler{}
	l := logging.NewTestLogger(t)
	b := newAuthHandler(l, "test")

	addAPITokenHeader(req, "test")
	b.Middleware(h).ServeHTTP(rw, req)

	if !h.request {
		t.Errorf("Handler got no request")
	}

	if rw.Code != 200 {
		t.Errorf("Unexpected status code %q", rw.Result().Status)
	}
}
