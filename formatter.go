package stackdriver

import (
	"encoding/json"
	"fmt"
	"runtime"

	"github.com/Sirupsen/logrus"
)

type severity string

const (
	severityDefault  severity = "DEFAULT"
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

type logEntry struct {
	Severity      severity    `json:"severity"`
	TextPayload   string      `json:"textPayload,omitempty"`
	StructPayload interface{} `json:"jsonPayload,omitempty"`
}

type errorData struct {
	ServiceContext serviceContext `json:"serviceContext"`
	Message        string         `json:"message"`
	Context        context        `json:"context"`
}

type serviceContext struct {
	Service string `json:"service"`
	Version string `json:"version,omitempty"`
}

type context struct {
	ReportLocation reportLocation `json:"reportLocation"`
}

type reportLocation struct {
	FilePath     string `json:"filePath"`
	LineNumber   int    `json:"lineNumber"`
	FunctionName string `json:"functionName"`
}

func getSkipLevel() int {
	return 4
}

type Formatter struct {
	Service string
	Version string
}

func NewFormatter(service, version string) *Formatter {
	return &Formatter{
		Service: service,
		Version: version,
	}
}

func (f *Formatter) Format(e *logrus.Entry) ([]byte, error) {
	entry := logEntry{
		Severity: levelsToSeverity[e.Level],
	}

	payload := make(map[string]interface{})
	for k, v := range e.Data {
		payload[k] = v
	}

	switch entry.Severity {
	case severityError:
		fallthrough
	case severityCritical:
		fallthrough
	case severityAlert:
		payload["serviceContext"] = serviceContext{
			Service: f.Service,
			Version: f.Version,
		}

		if err, ok := payload["error"]; ok {
			payload["message"] = fmt.Sprintf("%s: %s", e.Message, err)
			delete(payload, "error")
		} else {
			payload["message"] = e.Message
		}

		var ctx context
		skip := getSkipLevel()
		if pc, file, line, ok := runtime.Caller(skip); ok {
			fn := runtime.FuncForPC(pc)
			ctx.ReportLocation = reportLocation{
				FilePath:     file,
				LineNumber:   line,
				FunctionName: fn.Name(),
			}
		}
		payload["context"] = ctx

		entry.StructPayload = payload
	default:
		if len(payload) == 0 {
			entry.TextPayload = e.Message
		} else {
			payload["message"] = e.Message
			entry.StructPayload = payload
		}
	}

	b, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return nil, err
	}

	return append(b, byte('\n')), nil
}
