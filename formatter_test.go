package stackdriver

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestFormatter(t *testing.T) {
	skipTimestamp = true

	for _, tt := range formatterTests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer

			logger := logrus.New()
			logger.Out = &out
			logger.Formatter = NewFormatter(
				WithService("test"),
				WithVersion("0.1"),
			)

			tt.run(logger)
			got, err := json.Marshal(tt.out)
			if err != nil {
				t.Error(err)
			}
			assert.JSONEq(t, out.String(), string(got))
		})
	}
}

var formatterTests = []struct {
	run  func(*logrus.Logger)
	out  map[string]interface{}
	name string
}{
	{
		name: "With Field",
		run: func(logger *logrus.Logger) {
			logger.WithField("foo", "bar").Info("my log entry")
		},
		out: map[string]interface{}{
			"severity": "INFO",
			"message":  "my log entry",
			"context": map[string]interface{}{
				"data": map[string]interface{}{
					"foo": "bar",
				},
			},
		},
	},
	{
		name: "WithField and WithError",
		run: func(logger *logrus.Logger) {
			logger.
				WithField("foo", "bar").
				WithError(errors.New("test error")).
				Info("my log entry")
		},
		out: map[string]interface{}{
			"severity": "INFO",
			"message":  "my log entry",
			"context": map[string]interface{}{
				"data": map[string]interface{}{
					"foo":   "bar",
					"error": "test error",
				},
			},
		},
	},
	{
		name: "WithField and Error",
		run: func(logger *logrus.Logger) {
			logger.WithField("foo", "bar").Error("my log entry")
		},
		out: map[string]interface{}{
			"severity": "ERROR",
			"message":  "my log entry",
			"serviceContext": map[string]interface{}{
				"service": "test",
				"version": "0.1",
			},
			"context": map[string]interface{}{
				"data": map[string]interface{}{
					"foo": "bar",
				},
				"reportLocation": map[string]interface{}{
					"file":     "github.com/icco/logrus-stackdriver-formatter/formatter.go",
					"line":     263.0,
					"function": "(*Formatter).Format",
				},
			},
			"sourceLocation": map[string]interface{}{
				"file":     "github.com/icco/logrus-stackdriver-formatter/formatter.go",
				"line":     263.0,
				"function": "(*Formatter).Format",
			},
		},
	},
	{
		name: "WithField, WithError and Error",
		run: func(logger *logrus.Logger) {
			logger.
				WithField("foo", "bar").
				WithError(errors.New("test error")).
				Error("my log entry")
		},
		out: map[string]interface{}{
			"severity": "ERROR",
			"message":  "my log entry: test error",
			"serviceContext": map[string]interface{}{
				"service": "test",
				"version": "0.1",
			},
			"context": map[string]interface{}{
				"data": map[string]interface{}{
					"foo": "bar",
				},
				"reportLocation": map[string]interface{}{
					"file":     "github.com/icco/logrus-stackdriver-formatter/formatter.go",
					"line":     263.0,
					"function": "(*Formatter).Format",
				},
			},
			"sourceLocation": map[string]interface{}{
				"file":     "github.com/icco/logrus-stackdriver-formatter/formatter.go",
				"line":     263.0,
				"function": "(*Formatter).Format",
			},
		},
	},
	{
		name: "WithField, HTTPRequest and Error",
		run: func(logger *logrus.Logger) {
			logger.
				WithFields(logrus.Fields{
					"foo": "bar",
					"httpRequest": map[string]interface{}{
						"requestMethod": "GET",
					},
				}).
				Error("my log entry")
		},
		out: map[string]interface{}{
			"severity": "ERROR",
			"message":  "my log entry",
			"serviceContext": map[string]interface{}{
				"service": "test",
				"version": "0.1",
			},
			"context": map[string]interface{}{
				"data": map[string]interface{}{
					"foo": "bar",
					"httpRequest": map[string]interface{}{
						"requestMethod": "GET",
					},
				},
				"reportLocation": map[string]interface{}{
					"file":     "github.com/icco/logrus-stackdriver-formatter/formatter.go",
					"line":     263.0,
					"function": "(*Formatter).Format",
				},
			},
			"sourceLocation": map[string]interface{}{
				"file":     "github.com/icco/logrus-stackdriver-formatter/formatter.go",
				"line":     263.0,
				"function": "(*Formatter).Format",
			},
		},
	},
}
