package stackdriver

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/go-stack/stack"
	"github.com/sirupsen/logrus"
)

var skipTimestamp bool

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

type serviceContext struct {
	Service string `json:"service,omitempty"`
	Version string `json:"version,omitempty"`
}

type reportLocation struct {
	FilePath     string `json:"filePath,omitempty"`
	LineNumber   int    `json:"lineNumber,omitempty"`
	FunctionName string `json:"functionName,omitempty"`
}

type context struct {
	Data           map[string]interface{} `json:"data,omitempty"`
	ReportLocation *reportLocation        `json:"reportLocation,omitempty"`
	HTTPRequest    map[string]interface{} `json:"httpRequest,omitempty"`
}

type entry struct {
	Timestamp      string          `json:"timestamp,omitempty"`
	ServiceContext *serviceContext `json:"serviceContext,omitempty"`
	Message        string          `json:"message,omitempty"`
	Severity       severity        `json:"severity,omitempty"`
	Context        *context        `json:"context,omitempty"`
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
	severity := levelsToSeverity[e.Level]

	ee := entry{

		Message:  e.Message,
		Severity: severity,
		Context: &context{
			Data: e.Data,
		},
	}

	if !skipTimestamp {
		ee.Timestamp = time.Now().UTC().Format(time.RFC3339)
	}

	switch severity {
	case severityError, severityCritical, severityAlert:
		ee.ServiceContext = &serviceContext{
			Service: f.Service,
			Version: f.Version,
		}

		// When using WithError(), the error is sent separately, but Error
		// Reporting expects it to be a part of the message so we append it
		// instead.
		if err, ok := ee.Context.Data["error"]; ok {
			ee.Message = fmt.Sprintf("%s: %s", e.Message, err)
			delete(ee.Context.Data, "error")
		} else {
			ee.Message = e.Message
		}

		// As a convenience, when using supplying the httpRequest field, it
		// gets special care.
		if reqData, ok := ee.Context.Data["httpRequest"]; ok {
			if req, ok := reqData.(map[string]interface{}); ok {
				ee.Context.HTTPRequest = req
				delete(ee.Context.Data, "httpRequest")
			}
		}

		// Extract report location from call stack.
		c := stack.Caller(4)

		lineNumber, _ := strconv.ParseInt(fmt.Sprintf("%d", c), 10, 64)

		ee.Context.ReportLocation = &reportLocation{
			FilePath:     fmt.Sprintf("%+s", c),
			LineNumber:   int(lineNumber),
			FunctionName: fmt.Sprintf("%n", c),
		}
	}

	b, err := json.Marshal(ee)
	if err != nil {
		return nil, err
	}

	return append(b, '\n'), nil
}
