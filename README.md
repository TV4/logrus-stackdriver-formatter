# logrus-stackdriver-formatter

[![Build Status](https://travis-ci.org/TV4/logrus-stackdriver-formatter.svg?branch=master)](https://travis-ci.org/TV4/logrus-stackdriver-formatter)
[![Go Report Card](https://goreportcard.com/badge/github.com/TV4/logrus-stackdriver-formatter)](https://goreportcard.com/report/github.com/TV4/logrus-stackdriver-formatter)
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat)](https://godoc.org/github.com/TV4/logrus-stackdriver-formatter)
[![License MIT](https://img.shields.io/badge/license-MIT-lightgrey.svg?style=flat)](https://github.com/TV4/logrus-stackdriver-formatter#license)

[logrus](https://github.com/sirupsen/logrus) formatter for Stackdriver.

In addition to supporting level-based logging to Stackdriver, for Error, Fatal and Panic levels it will append error context for [Error Reporting](https://cloud.google.com/error-reporting/).

## Installation

```shell
go get -u github.com/TV4/logrus-stackdriver-formatter
```

## Usage

```go
package main

import (
    "github.com/sirupsen/logrus"
    stackdriver "github.com/TV4/logrus-stackdriver-formatter"
)

var log = logrus.New()

func init() {
    log.Formatter = stackdriver.NewFormatter(
        stackdriver.WithService("your-service"), 
        stackdriver.WithVersion("v0.1.0"),
    )
    log.Level = logrus.DebugLevel

    log.Info("ready to log!")
}
```

Here's a sample entry (prettified) from the example:

```json
{
  "serviceContext": {
    "service": "test-service",
    "version": "v0.1.0"
  },
  "message": "unable to parse integer: strconv.ParseInt: parsing \"text\": invalid syntax",
  "severity": "ERROR",
  "context": {
    "reportLocation": {
      "filePath": "github.com/TV4/logrus-stackdriver-formatter/example_test.go",
      "lineNumber": 21,
      "functionName": "ExampleLogError"
    }
  }
}
```

## HTTP request context

If you'd like to add additional context like the `httpRequest`, here's a convenience function for creating a HTTP logger:

```go
func httpLogger(logger *logrus.Logger, r *http.Request) *logrus.Entry {
    return logger.WithFields(logrus.Fields{
        "httpRequest": map[string]interface{}{
            "method":    r.Method,
            "url":       r.URL.String(),
            "userAgent": r.Header.Get("User-Agent"),
            "referrer":  r.Header.Get("Referer"),
        },
    })
}
```

Then, in your HTTP handler, create a new context logger and all your log entries will have the HTTP request context appended to them:

```go
func handler(w http.ResponseWriter, r *http.Request) {
    httplog := httpLogger(log, r)
    // ...
    httplog.Infof("Logging with HTTP request context")
}
```
