package stackdriver

// errorData contains the error payload.
//
// https://cloud.google.com/error-reporting/docs/formatting-error-messages
type errorData struct {
	ServiceContext serviceContext `json:"serviceContext"`

	// Message contains the stack trace that was reported or logged by the
	// service.
	Message string `json:"message"`

	Context context `json:"context"`
}

// serviceContext for which an error was reported.
//
// https://cloud.google.com/error-reporting/docs/formatting-error-messages#FIELDS.service_context
type serviceContext struct {
	Service string `json:"service"`
	Version string `json:"version,omitempty"`
}

// context contains data about the context in which the error occurred.
//
// https://cloud.google.com/error-reporting/docs/formatting-error-messages#FIELDS.context
type context struct {
	ReportLocation reportLocation `json:"reportLocation"`
}

// reportLocation contains the location in the source code where the decision
// was made to report the error, usually the place where it was logged. For a
// logged exception this would be the source line where the exception is
// logged, usually close to the place where it was caught.
//
// https://cloud.google.com/error-reporting/docs/formatting-error-messages#ErrorContext.FIELDS.report_location
type reportLocation struct {
	FilePath     string `json:"filePath,omitempty"`
	LineNumber   int    `json:"lineNumber,omitempty"`
	FunctionName string `json:"functionName,omitempty"`
}
