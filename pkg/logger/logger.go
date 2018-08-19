package logger

import (
	log "github.com/sirupsen/logrus"
)

var l *logger

type JSON map[string]string

type Logger interface {
	Debug(string, JSON)
	Info(string, JSON)
	Warn(string, JSON)
	Error(string, JSON)
	Fatal(string, JSON)
	Panic(string, JSON)
}

type logger struct {
	logger *log.Logger
}

func (l *logger) Debug(message string, labeles JSON) {}

func (l *logger) Info(message string, labeles JSON) {
	l.logger.WithFields(convertJsonIntoFields(labeles)).Info(message)
}

func (l *logger) Warn(message string, labeles JSON) {}

func (l *logger) Error(message string, labeles JSON) {}

func (l *logger) Fatal(message string, labeles JSON) {}

func (l *logger) Panic(message string, labeles JSON) {}

func Get() Logger {
	return &logger{
		logger: log.New(),
	}
}

func convertJsonIntoFields(json JSON) log.Fields {
	fields := log.Fields{}
	for key, val := range json {
		fields[key] = val
	}
	return fields
}
