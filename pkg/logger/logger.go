package logger

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	DebugLevel = "debug"
)

type Logger interface {
	Warnf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Debugf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
	Panicf(format string, args ...interface{})
}

type Log struct {
	zap *zap.Logger
}

func New(lvl string) (*Log, error) {
	zapConfig := zap.NewProductionConfig()
	zapConfig.Encoding = "json"
	zapConfig.OutputPaths = []string{"stdout"}
	zapConfig.ErrorOutputPaths = []string{"stdout"}
	zapConfig.EncoderConfig = zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		CallerKey:      "caller",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	if err := zapConfig.Level.UnmarshalText([]byte(lvl)); err != nil {
		return nil, fmt.Errorf("unmarshal level: %w", err)
	}

	l, err := zapConfig.Build()
	if err != nil {
		return nil, fmt.Errorf("build zap logger: %w", err)
	}

	log := &Log{
		zap: l.WithOptions(zap.AddCallerSkip(1)),
	}

	return log, nil
}

func (l *Log) Warnf(format string, args ...interface{}) {
	if l == nil {
		return
	}
	l.zap.Warn(fmt.Sprintf(format, args...))
}

func (l *Log) Infof(format string, args ...interface{}) {
	if l == nil {
		return
	}
	l.zap.Info(fmt.Sprintf(format, args...))
}

func (l *Log) Errorf(format string, args ...interface{}) {
	if l == nil {
		return
	}
	l.zap.Error(fmt.Sprintf(format, args...))
}

func (l *Log) Debugf(format string, args ...interface{}) {
	if l == nil {
		return
	}
	l.zap.Debug(fmt.Sprintf(format, args...))
}

func (l *Log) Fatalf(format string, args ...interface{}) {
	if l == nil {
		return
	}
	l.zap.Fatal(fmt.Sprintf(format, args...))
}

func (l *Log) Panicf(format string, args ...interface{}) {
	if l == nil {
		return
	}
	l.zap.Panic(fmt.Sprintf(format, args...))
}
