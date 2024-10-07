package logger

import (
	"remote-runner/config"

	"github.com/sirupsen/logrus"
)

func MakeLogarusWithFields(subComponent string, additionalFields ...logrus.Fields) *logrus.Entry {
	fields := logrus.Fields{
		"sub_component": subComponent,
		"instance_id":   config.GetConfigs().GetInstanceID(),
	}

	if len(additionalFields) > 0 && additionalFields[0] != nil {
		for key, value := range additionalFields[0] {
			fields[key] = value
		}
	}

	return logrus.WithFields(fields)
}
func Info(message string, subComponent string, additionalFields ...logrus.Fields) {
	MakeLogarusWithFields(subComponent, additionalFields...).Info(message + " ")
}
func Error(message string, subComponent string, err error, additionalFields ...logrus.Fields) {
	if err != nil && additionalFields != nil {
		additionalFields = append(additionalFields, logrus.Fields{"error": err})
	} else {
		additionalFields = []logrus.Fields{{"error": err}}
	}
	MakeLogarusWithFields(subComponent, additionalFields...).Error(message+" ", err)
}
func Debug(message string, subComponent string, additionalFields ...logrus.Fields) {
	MakeLogarusWithFields(subComponent, additionalFields...).Debug(message + " ")
}
func Warn(message string, subComponent string, additionalFields ...logrus.Fields) {
	MakeLogarusWithFields(subComponent, additionalFields...).Warn(message + " ")
}

func SetupLogging() {
	// Set up the logger
	logFormatter := config.GetConfigs().GetConfigsWithDefault("log_formatter", "text")
	switch logFormatter {
	case "json":
		logrus.SetFormatter(&logrus.JSONFormatter{})
	case "text":
		logrus.SetFormatter(&logrus.TextFormatter{})
	default:
		logrus.SetFormatter(&logrus.TextFormatter{})
	}

	logLevel := config.GetConfigs().GetConfigsWithDefault("log_level", "info")
	switch logLevel {
	case "info":
		logrus.SetLevel(logrus.InfoLevel)
	case "debug":
		logrus.SetLevel(logrus.DebugLevel)
	case "warn":
		logrus.SetLevel(logrus.WarnLevel)
	case "error":
		logrus.SetLevel(logrus.ErrorLevel)
	default:
		logrus.SetLevel(logrus.InfoLevel)
	}
}
