package stackdriver

import (
	"errors"
	"runtime"
	"strings"
	"sync"

	"github.com/Sirupsen/logrus"
)

var (
	stackSkipsCallers = make([]uintptr, 0, 20)
	stackSkips        = map[logrus.Level]int{}
	stackSkipsMu      = sync.RWMutex{}
)

var ErrSkipNotFound = errors.New("could not find skips for log level")

// This implementation is copied from logrus-gce by znly.
// See https://github.com/znly/logrus-gce for more information.
func getSkipLevel(level logrus.Level) (int, error) {
	stackSkipsMu.RLock()
	if skip, ok := stackSkips[level]; ok {
		defer stackSkipsMu.RUnlock()
		return skip, nil

	}
	stackSkipsMu.RUnlock()

	stackSkipsMu.Lock()
	defer stackSkipsMu.Unlock()
	if skip, ok := stackSkips[level]; ok {
		return skip, nil

	}

	// detect until we escape logrus back to the client package skip out of
	// runtime and stackdriver package, hence 3
	stackSkipsCallers := make([]uintptr, 20)
	runtime.Callers(3, stackSkipsCallers)
	for i, pc := range stackSkipsCallers {
		f := runtime.FuncForPC(pc)
		if strings.Contains(f.Name(), "github.com/Sirupsen/logrus") {
			continue
		}
		stackSkips[level] = i + 1
		return i + 1, nil
	}
	return 0, ErrSkipNotFound
}
