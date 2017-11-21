package stackdriver

// serviceContext for which an error was reported.
//
// https://cloud.google.com/error-reporting/docs/formatting-error-messages#FIELDS.service_context
type serviceContext struct {
	Service string `json:"service"`
	Version string `json:"version,omitempty"`
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
