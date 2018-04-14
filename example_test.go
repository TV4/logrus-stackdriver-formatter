package stackdriver_test

import (
	"os"
	"strconv"

	stackdriver "github.com/TV4/logrus-stackdriver-formatter"
	"github.com/sirupsen/logrus"
)

func ExampleLogError() {
	logger := logrus.New()
	logger.Out = os.Stdout
	logger.Formatter = stackdriver.NewFormatter(
		stackdriver.WithService("test-service"),
		stackdriver.WithVersion("v0.1.0"),
	)

	logger.Info("application up and running")

	_, err := strconv.ParseInt("text", 10, 64)
	if err != nil {
		logger.WithError(err).Errorln("unable to parse integer")
	}

	// Output:
	// {"message":"application up and running","severity":"INFO","context":{}}
	// {"serviceContext":{"service":"test-service","version":"v0.1.0"},"message":"unable to parse integer: strconv.ParseInt: parsing \"text\": invalid syntax","severity":"ERROR","context":{"reportLocation":{"filePath":"github.com/TV4/logrus-stackdriver-formatter/example_test.go","lineNumber":23,"functionName":"ExampleLogError"}}}
}
