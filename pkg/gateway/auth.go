package gateway

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/Sirupsen/logrus"
)

const (
	autHeaderSchema string = "Bearer "
)

type authHandlerMiddleware struct {
	logger   *logrus.Entry
	apiToken string
}

func newAuthHandler(logger *logrus.Entry, apiToken string) *authHandlerMiddleware {
	return &authHandlerMiddleware{
		logger:   logger,
		apiToken: apiToken,
	}
}

func (m authHandlerMiddleware) getToken(r *http.Request) string {
	h := r.Header.Get("Authorization")

	if strings.HasPrefix(h, autHeaderSchema) {
		return h[len(autHeaderSchema):]
	}

	return h
}

func (m authHandlerMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := m.getToken(r)

		if h != m.apiToken {
			w.WriteHeader(http.StatusUnauthorized)
			if _, err := w.Write([]byte("Unauthorized.")); err != nil {
				m.logger.Errorf("failed to write unauthorized: %v", err)
			}

			return
		}

		next.ServeHTTP(w, r)
	})
}

func addApiTokenHeader(r *http.Request, apiToken string) {
	r.Header.Add("Authorization", fmt.Sprintf("Bearer %s", apiToken))
}
