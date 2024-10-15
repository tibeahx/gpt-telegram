package logger

import (
	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

func init() {
	log = logrus.New()
}

func GetLogger() *logrus.Logger {
	return log
}
