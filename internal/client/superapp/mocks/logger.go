package mocks

import "onmi/pkg/logger"

type MockLogger struct{}

func NewMockLogger() *MockLogger {
	return &MockLogger{}
}

func (l *MockLogger) WithFields(fields logger.Fields) logger.Logger {
	return l
}

func (l *MockLogger) WithErr(err error) logger.Logger {
	return l
}

func (l *MockLogger) WithTag(tag string) logger.Logger {
	return l
}

func (l *MockLogger) Warn(msg string, fields logger.Fields) {}

func (l *MockLogger) Warnf(format string, args ...interface{}) {}

func (l *MockLogger) Info(msg string, fields logger.Fields) {}

func (l *MockLogger) Infof(format string, args ...interface{}) {}

func (l *MockLogger) Error(msg string, fields logger.Fields) {}

func (l *MockLogger) Errorf(format string, args ...interface{}) {}

func (l *MockLogger) Debug(msg string, fields logger.Fields) {}

func (l *MockLogger) Debugf(format string, args ...interface{}) {}

func (l *MockLogger) Fatal(msg string, fields logger.Fields) {}

func (l *MockLogger) Fatalf(format string, args ...interface{}) {}

func (l *MockLogger) Panic(msg string, fields logger.Fields) {}

func (l *MockLogger) Panicf(format string, args ...interface{}) {}

func (l *MockLogger) SetLevel(lvl string) error {
	return nil
}
