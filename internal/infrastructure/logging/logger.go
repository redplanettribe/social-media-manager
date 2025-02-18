package logging

import (
	"os"

	"github.com/sirupsen/logrus"

	"github.com/redplanettribe/social-media-manager/internal/infrastructure/config"
)

func NewLogger(cfg *config.LoggerConfig) *logrus.Logger {
	logger := logrus.New()

	level, err := logrus.ParseLevel(cfg.Level)
	if err != nil {
		logger.SetLevel(logrus.DebugLevel)
	} else {
		logger.SetLevel(level)
	}

	logger.SetOutput(os.Stdout)
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	return logger
}
