package logger

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	InfoLevel  = "info"
	WarnLevel  = "warn"
	ErrorLevel = "error"
	DebugLevel = "debug"
	FatalLevel = "fatal"
	PanicLevel = "panic"
)

type Fields map[string]interface{}

type Logger interface {
	WithFields(fields Fields) Logger
	WithErr(err error) Logger
	WithTag(tag string) Logger
	Warn(msg string, fields Fields)
	Warnf(format string, args ...interface{})
	Info(msg string, fields Fields)
	Infof(format string, args ...interface{})
	Error(msg string, fields Fields)
	Errorf(format string, args ...interface{})
	Debug(msg string, fields Fields)
	Debugf(format string, args ...interface{})
	Fatal(msg string, fields Fields)
	Fatalf(format string, args ...interface{})
	Panic(msg string, fields Fields)
	Panicf(format string, args ...interface{})
	SetLevel(lvl string) error
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

func (l *Log) WithFields(fields Fields) Logger {
	if l == nil {
		return l
	}

	return &Log{zap: l.zap.With(l.withFields(fields)...)}
}

func (l *Log) WithErr(err error) Logger {
	if l == nil {
		return l
	}
	return &Log{
		zap: l.zap.With(zap.Error(err)),
	}
}

func (l *Log) WithTag(tag string) Logger {
	if l == nil {
		return l
	}
	return &Log{zap: l.zap.With(zap.Field{
		Key:    "_tag",
		Type:   zapcore.StringType,
		String: tag,
	})}
}

func (l *Log) Warn(msg string, fields Fields) {
	if l == nil {
		return
	}
	l.zap.Warn(msg, l.withFields(fields)...)
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

func (l *Log) Info(msg string, fields Fields) {
	if l == nil {
		return
	}
	l.zap.Info(msg, l.withFields(fields)...)
}

func (l *Log) Error(msg string, fields Fields) {
	if l == nil {
		return
	}
	l.zap.Error(msg, l.withFields(fields)...)
}

func (l *Log) Errorf(format string, args ...interface{}) {
	if l == nil {
		return
	}
	l.zap.Error(fmt.Sprintf(format, args...))
}

func (l *Log) Debug(msg string, fields Fields) {
	if l == nil {
		return
	}
	l.zap.Debug(msg, l.withFields(fields)...)
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

func (l *Log) Fatal(msg string, fields Fields) {
	if l == nil {
		return
	}
	l.zap.Fatal(msg, l.withFields(fields)...)
}

func (l *Log) Panicf(format string, args ...interface{}) {
	if l == nil {
		return
	}
	l.zap.Panic(fmt.Sprintf(format, args...))
}

func (l *Log) Panic(msg string, fields Fields) {
	if l == nil {
		return
	}
	l.zap.Panic(msg, l.withFields(fields)...)
}

func (l *Log) SetLevel(lvl string) error {
	var level zapcore.Level
	err := level.Set(lvl)
	if err != nil {
		return fmt.Errorf("level '%s' set error: %v", lvl, err)
	}
	l.zap.Core().Enabled(level)
	return nil
}

func (l *Log) withFields(fields Fields) []zap.Field {
	fs := make([]zap.Field, 0, len(fields))
	for k, v := range fields {
		fs = append(fs, zap.Any(k, v))
	}
	return fs
}
