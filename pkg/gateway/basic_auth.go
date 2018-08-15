package gateway

import (
	"crypto/subtle"
	"net/http"

	"github.com/Sirupsen/logrus"
)

type basicAuthMiddleware struct {
	logger             *logrus.Entry
	username, password string
}

func newBasicAuth(logger *logrus.Entry, username, password string) *basicAuthMiddleware {
	return &basicAuthMiddleware{
		logger:   logger,
		username: username,
		password: password,
	}
}

func (m basicAuthMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		user, pass, ok := r.BasicAuth()

		if !ok || subtle.ConstantTimeCompare([]byte(user), []byte(m.username)) != 1 || subtle.ConstantTimeCompare([]byte(pass), []byte(m.password)) != 1 {
			w.WriteHeader(http.StatusUnauthorized)
			w.Header().Set("WWW-Authenticate", `Basic realm="Basic authentication required"`)
			if _, err := w.Write([]byte("Unauthorized.")); err != nil {
				m.logger.Errorf("Unabling writing unauthorized: %v", err)
			}

			return
		}

		next.ServeHTTP(w, r)
	})
}
