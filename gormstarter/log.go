package gormstarter

import (
	"context"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm/logger"
	"time"
)

type logrusLogger struct {
	log *logrus.Logger
}

func (l *logrusLogger) LogMode(level logger.LogLevel) logger.Interface {
	return l
}

func (l *logrusLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	elapsed := time.Since(begin)
	sql, rows := fc()
	fields := logrus.Fields{
		"rows":    rows,
		"elapsed": elapsed,
	}
	if err != nil {
		l.log.WithContext(ctx).WithFields(fields).Error(sql, err)
	} else {
		l.log.WithContext(ctx).WithFields(fields).Debug(sql)
	}
}

func (l *logrusLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	l.log.WithContext(ctx).Infof(msg, data...)
}

func (l *logrusLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	l.log.WithContext(ctx).Warnf(msg, data...)
}

func (l *logrusLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	l.log.WithContext(ctx).Errorf(msg, data...)
}
