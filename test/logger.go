package test

import "github.com/sirupsen/logrus"

type LogWrapper struct {
	Logger *logrus.Logger
}

func (l *LogWrapper) Error(msg string) {
	l.Logger.Error(msg)
}
