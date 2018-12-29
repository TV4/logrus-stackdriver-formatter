package stackdriver

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
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

type ServiceContext struct {
	Service string `json:"service,omitempty"`
	Version string `json:"version,omitempty"`
}

type ReportLocation struct {
	FilePath     string `json:"file,omitempty"`
	LineNumber   int    `json:"line,omitempty"`
	FunctionName string `json:"function,omitempty"`
}

type Context struct {
	Data           map[string]interface{} `json:"data,omitempty"`
	ReportLocation *ReportLocation        `json:"reportLocation,omitempty"`
	HTTPRequest    map[string]interface{} `json:"httpRequest,omitempty"`
}

type HttpRequest struct {
	RequestMethod                  string `json:"requestMethod,omitempty"`
	RequestUrl                     string `json:"requestUrl,omitempty"`
	RequestSize                    string `json:"requestSize,omitempty"`
	Status                         string `json:"status,omitempty"`
	ResponseSize                   string `json:"responseSize,omitempty"`
	UserAgent                      string `json:"userAgent,omitempty"`
	RemoteIp                       string `json:"remoteIp,omitempty"`
	ServerIp                       string `json:"serverIp,omitempty"`
	Referer                        string `json:"referer,omitempty"`
	Latency                        string `json:"latency,omitempty"`
	CacheLookup                    bool   `json:"cacheLookup,omitempty"`
	CacheHit                       bool   `json:"cacheHit,omitempty"`
	CacheValidatedWithOriginServer bool   `json:"cacheValidatedWithOriginServer,omitempty"`
	CacheFillBytes                 string `json:"cacheFillBytes,omitempty"`
	Protocol                       string `json:"protocol,omitempty"`
}

type Entry struct {
	LogName        string          `json:"logName,omitempty"`
	Timestamp      string          `json:"timestamp,omitempty"`
	Severity       severity        `json:"severity,omitempty"`
	HTTPRequest    *HttpRequest    `json:"httpRequest,omitempty"`
	Trace          string          `json:"trace,omitempty"`
	ServiceContext *ServiceContext `json:"serviceContext,omitempty"`
	Message        string          `json:"message,omitempty"`
	Context        *Context        `json:"context,omitempty"`
	SourceLocation *ReportLocation `json:"sourceLocation,omitempty"`
}

// Formatter implements Stackdriver formatting for logrus.
type Formatter struct {
	Service   string
	Version   string
	ProjectID string
	StackSkip []string
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

func WithProjectID(i string) Option {
	return func(f *Formatter) {
		f.ProjectID = i
	}
}

// WithStackSkip lets you configure which packages should be skipped for locating the error.
func WithStackSkip(v string) Option {
	return func(f *Formatter) {
		f.StackSkip = append(f.StackSkip, v)
	}
}

// NewFormatter returns a new Formatter.
func NewFormatter(options ...Option) *Formatter {
	fmtr := Formatter{
		StackSkip: []string{
			"github.com/sirupsen/logrus",
		},
	}
	for _, option := range options {
		option(&fmtr)
	}
	return &fmtr
}

func (f *Formatter) errorOrigin() (stack.Call, error) {
	skip := func(pkg string) bool {
		for _, skip := range f.StackSkip {
			if pkg == skip {
				return true
			}
		}
		return false
	}

	// We start at 2 to skip this call and our caller's call.
	for i := 2; ; i++ {
		c := stack.Caller(i)
		// ErrNoFunc indicates we're over traversing the stack.
		if _, err := c.MarshalText(); err != nil {
			return stack.Call{}, nil
		}
		pkg := fmt.Sprintf("%+k", c)
		// Remove vendoring from package path.
		parts := strings.SplitN(pkg, "/vendor/", 2)
		pkg = parts[len(parts)-1]
		if !skip(pkg) {
			return c, nil
		}
	}
}

// taken from https://github.com/sirupsen/logrus/blob/master/json_formatter.go#L51
func replaceErrors(source logrus.Fields) logrus.Fields {
	data := make(logrus.Fields, len(source))
	for k, v := range source {
		switch v := v.(type) {
		case error:
			// Otherwise errors are ignored by `encoding/json`
			// https://github.com/sirupsen/logrus/issues/137
			data[k] = v.Error()
		default:
			data[k] = v
		}
	}
	return data
}

// Format formats a logrus entry according to the Stackdriver specifications.
func (f *Formatter) Format(e *logrus.Entry) ([]byte, error) {
	severity := levelsToSeverity[e.Level]

	ee := Entry{
		Message:  e.Message,
		Severity: severity,
		Context: &Context{
			Data: replaceErrors(e.Data),
		},
	}

	if val, ok := e.Data["trace"]; ok {
		ee.Trace = val.(string)
	}

	if val, ok := e.Data["httpRequest"]; ok {
		ee.HTTPRequest = val.(*HttpRequest)
	}

	if val, ok := e.Data["logID"]; ok {
		ee.LogName = "projects/" + f.ProjectID + "/logs/" + val.(string)
	}

	if !skipTimestamp {
		ee.Timestamp = time.Now().UTC().Format(time.RFC3339Nano)
	}

	switch severity {
	case severityError, severityCritical, severityAlert:
		ee.ServiceContext = &ServiceContext{
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
		if c, err := f.errorOrigin(); err == nil {
			lineNumber, _ := strconv.ParseInt(fmt.Sprintf("%d", c), 10, 64)

			ee.Context.ReportLocation = &ReportLocation{
				FilePath:     fmt.Sprintf("%+s", c),
				LineNumber:   int(lineNumber),
				FunctionName: fmt.Sprintf("%n", c),
			}

			ee.SourceLocation = &ReportLocation{
				FilePath:     fmt.Sprintf("%+s", c),
				LineNumber:   int(lineNumber),
				FunctionName: fmt.Sprintf("%n", c),
			}
		}
	}

	b, err := json.Marshal(ee)
	if err != nil {
		return nil, err
	}

	return append(b, '\n'), nil
}
