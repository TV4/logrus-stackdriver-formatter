package stackdriver

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/TV4/logrus-stackdriver-formatter/test"
	"github.com/kr/pretty"
	"github.com/sirupsen/logrus"
)

func TestStackSkip(t *testing.T) {
	var out bytes.Buffer

	logger := logrus.New()
	logger.Out = &out
	logger.Formatter = NewFormatter(
		WithService("test"),
		WithVersion("0.1"),
		WithStackSkip("github.com/TV4/logrus-stackdriver-formatter/test"),
	)

	mylog := test.LogWrapper{
		Logger: logger,
	}

	mylog.Error("my log entry")

	var got map[string]interface{}
	json.Unmarshal(out.Bytes(), &got)

	want := map[string]interface{}{
		"severity": "ERROR",
		"message":  "my log entry",
		"serviceContext": map[string]interface{}{
			"service": "test",
			"version": "0.1",
		},
		"context": map[string]interface{}{
			"reportLocation": map[string]interface{}{
				"filePath":     "github.com/TV4/logrus-stackdriver-formatter/stackskip_test.go",
				"lineNumber":   29.0,
				"functionName": "TestStackSkip",
			},
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("unexpected output = %# v; want = %# v", pretty.Formatter(got), pretty.Formatter(want))
	}
}
