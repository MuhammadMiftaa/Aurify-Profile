package log

import (
	"os"

	"refina-profile/config/env"

	"github.com/sirupsen/logrus"
)

var Log *logrus.Logger

func SetupLogger() {
	Log = logrus.New()

	Log.SetOutput(os.Stdout)

	if env.Cfg.Server.Mode == "production" {
		Log.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
		})
		Log.SetLevel(logrus.InfoLevel)
	} else {
		Log.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
			ForceColors:     true,
		})
		Log.SetLevel(logrus.DebugLevel)
	}
}

// Helper functions for structured logging
func Info(message string, fields map[string]any) {
	Log.WithFields(logrus.Fields(fields)).Info(message)
}

func Debug(message string, fields map[string]any) {
	Log.WithFields(logrus.Fields(fields)).Debug(message)
}

func Warn(message string, fields map[string]any) {
	Log.WithFields(logrus.Fields(fields)).Warn(message)
}

func Error(message string, fields map[string]any) {
	Log.WithFields(logrus.Fields(fields)).Error(message)
}

func Fatal(message string, fields map[string]any) {
	Log.WithFields(logrus.Fields(fields)).Fatal(message)
}
