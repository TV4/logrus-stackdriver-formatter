package stackdriver

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/go-stack/stack"
	"github.com/sirupsen/logrus"
)

type severity string

const (
	severityDebug    severity = "DEBUG"
	severityInfo     severity = "INFO"
	severityWarning  severity = "WARNING"
	severityError    severity = "ERROR"
	severityCritical severity = "CRITICAL"
	severityAlert    severity = "ALERT"
)

var levelsToSeverity = map[logrus.Level]severity{
	logrus.DebugLevel: severityDebug,
	logrus.InfoLevel:  severityInfo,
	logrus.WarnLevel:  severityWarning,
	logrus.ErrorLevel: severityError,
	logrus.FatalLevel: severityCritical,
	logrus.PanicLevel: severityAlert,
}

// Formatter implements Stackdriver formatting for logrus.
type Formatter struct {
	Service string
	Version string
}

// Option lets you configure the Formatter.
type Option func(*Formatter)

// WithService lets you configure the service name used for error reporting.
func WithService(n string) Option {
	return func(f *Formatter) {
		f.Service = n
	}
}

// WithVersion lets you configure the service version used for error reporting.
func WithVersion(v string) Option {
	return func(f *Formatter) {
		f.Version = v
	}
}

// NewFormatter returns a new Formatter.
func NewFormatter(options ...Option) *Formatter {
	var fmtr Formatter
	for _, option := range options {
		option(&fmtr)
	}
	return &fmtr
}

// Format formats a logrus entry according to the Stackdriver specifications.
func (f *Formatter) Format(e *logrus.Entry) ([]byte, error) {
	payload := make(map[string]interface{})

	severity := levelsToSeverity[e.Level]

	// Copy entry data to the error payload.
	for k, v := range e.Data {
		payload[k] = v
	}

	switch severity {
	case severityError:
		fallthrough
	case severityCritical:
		fallthrough
	case severityAlert:
		payload["serviceContext"] = serviceContext{
			Service: f.Service,
			Version: f.Version,
		}

		// When using WithError(), the error is sent separately, but Error
		// Reporting expects it to be a part of the message so we append it
		// instead.
		if err, ok := payload["error"]; ok {
			payload["message"] = fmt.Sprintf("%s: %s", e.Message, err)
			delete(payload, "error")
		} else {
			payload["message"] = e.Message
		}

		c := stack.Caller(4)

		var (
			filePath      = fmt.Sprintf("%+s", c)
			lineNumber, _ = strconv.ParseInt(fmt.Sprintf("%d", c), 10, 64)
			functionName  = fmt.Sprintf("%n", c)
		)

		// Make sure not to replace the context, in case user specified httpRequest.
		if _, ok := payload["context"]; !ok {
			payload["context"] = make(map[string]interface{})
		}

		ctx := payload["context"].(map[string]interface{})
		ctx["reportLocation"] = reportLocation{
			FilePath:     filePath,
			LineNumber:   int(lineNumber),
			FunctionName: functionName,
		}
		payload["context"] = ctx
	default:
		payload["message"] = e.Message
	}

	payload["severity"] = severity

	b, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return append(b, '\n'), nil
}
