package mocks

type MockLogger struct{}

func NewMockLogger() *MockLogger {
	return &MockLogger{}
}

func (l *MockLogger) Warnf(format string, args ...interface{}) {}

func (l *MockLogger) Infof(format string, args ...interface{}) {}

func (l *MockLogger) Errorf(format string, args ...interface{}) {}

func (l *MockLogger) Debugf(format string, args ...interface{}) {}

func (l *MockLogger) Fatalf(format string, args ...interface{}) {}

func (l *MockLogger) Panicf(format string, args ...interface{}) {}
