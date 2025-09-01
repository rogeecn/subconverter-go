package logger

import (
	"os"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type Logger struct {
	*logrus.Logger
}

type Config struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
	Output string `yaml:"output"`
}

func New(config Config) *Logger {
	log := logrus.New()
	
	// 设置日志级别
	level, err := logrus.ParseLevel(config.Level)
	if err != nil {
		level = logrus.InfoLevel
	}
	log.SetLevel(level)
	
	// 设置日志格式
	switch config.Format {
	case "json":
		log.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
		})
	default:
		log.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
		})
	}
	
	// 设置日志输出
	switch config.Output {
	case "stderr":
		log.SetOutput(os.Stderr)
	default:
		log.SetOutput(os.Stdout)
	}
	
	return &Logger{log}
}

func (l *Logger) WithError(err error) *logrus.Entry {
	return l.WithField("error", errors.WithStack(err))
}

func (l *Logger) WithField(key string, value interface{}) *logrus.Entry {
	return l.Logger.WithField(key, value)
}

func (l *Logger) WithFields(fields logrus.Fields) *logrus.Entry {
	return l.Logger.WithFields(fields)
}