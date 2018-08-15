package testing

import (
	"bytes"
	"strings"
	"testing"

	"github.com/Sirupsen/logrus"
)

func NewTestLogger() (*bytes.Buffer, *logrus.Entry) {
	b := bytes.NewBufferString("")
	l := logrus.New()
	l.SetOutput(b)
	return b, l.WithFields(logrus.Fields{})
}

func CheckLogger(t *testing.T, buf *bytes.Buffer) {
	logs := buf.String()

	if strings.Contains(logs, "level=error") || strings.Contains(logs, "level=fatal") {
		t.Fatal(logs)
	}
}
