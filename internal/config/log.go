package config

import (
	"strings"

	"github.com/sirupsen/logrus"
)

func SetupLogger(level, format string) *logrus.Logger {
	logger := logrus.New()

	lvl, err := logrus.ParseLevel(strings.ToLower(level))
	if err != nil {
		logger.Warnf("invalid log level '%s', defaulting to 'info'", level)
		lvl = logrus.InfoLevel
	}
	logger.SetLevel(lvl)

	switch strings.ToLower(format) {
	case "json":
		logger.SetFormatter(&logrus.JSONFormatter{})
	case "text":
		fallthrough
	default:
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
		})
	}

	return logger
}
