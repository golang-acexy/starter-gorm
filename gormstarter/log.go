package gormstarter

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm/logger"
)

type logrusLogger struct {
	log *logrus.Logger
}

func (l *logrusLogger) LogMode(level logger.LogLevel) logger.Interface {
	return l
}

func (l *logrusLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if err != nil {
		elapsed := time.Since(begin)
		sql, rows := fc()
		fields := logrus.Fields{
			"rows":    rows,
			"elapsed": elapsed,
		}
		l.log.WithContext(ctx).WithFields(fields).Error(sql, err)
		return
	}
	lv := logrus.Level(sqlLoggerLevel)
	if l.log.IsLevelEnabled(lv) {
		elapsed := time.Since(begin)
		sql, rows := fc()
		fields := logrus.Fields{
			"rows":    rows,
			"elapsed": elapsed,
		}
		l.log.WithContext(ctx).WithFields(fields).Log(lv, sql)
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
