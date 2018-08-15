package testing

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/Sirupsen/logrus"
)

// NewTestLogger returns a logger with buffer setup, so that nothing really get's logged.
func NewTestLogger() (*bytes.Buffer, *logrus.Entry) {
	b := bytes.NewBufferString("")
	l := logrus.New()
	l.SetOutput(b)
	return b, l.WithFields(logrus.Fields{})
}

// CheckLogger checks the output for errors and fatals
func CheckLogger(t *testing.T, buf fmt.Stringer) {
	logs := buf.String()

	if strings.Contains(logs, "level=error") || strings.Contains(logs, "level=fatal") {
		t.Fatal(logs)
	}
}
