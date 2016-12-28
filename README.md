# logrus-stackdriver-formatter

[logrus](https://github.com/sirupsen/logrus) formatter for Stackdriver.

In addition to supporting level-based logging to Stackdriver, for Error, Fatal and Panic levels it will append error context for [Error Reporting](https://cloud.google.com/error-reporting/).

## Installation

```
go get -u github.com/TV4/logrus-stackdriver-formatter
```

## Usage

```go
package main

import (
    "github.com/Sirupsen/logrus"
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
