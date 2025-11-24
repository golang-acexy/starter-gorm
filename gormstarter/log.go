package gormstarter

import (
	"context"
	"time"

	log "github.com/acexy/golang-toolkit/logger"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm/logger"
)

type logrusLogger struct {
}

func (l *logrusLogger) LogMode(level logger.LogLevel) logger.Interface {
	return l
}

// Trace gorm打印sql的专用日志级别
func (l *logrusLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if err != nil {
		elapsed := time.Since(begin)
		sql, rows := fc()
		log.Logrus().WithContext(ctx).Errorln(sql, "rows:", rows, "elapsed:", elapsed, err)
		return
	}
	if log.IsLevelEnabled(sqlLoggerLevel) {
		elapsed := time.Since(begin)
		sql, rows := fc()
		log.Logrus().WithContext(ctx).Logln(logrus.Level(sqlLoggerLevel), sql, "rows:", rows, "elapsed:", elapsed)
	}
}

func (l *logrusLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	log.Logrus().WithContext(ctx).Infof(msg, data...)
}

func (l *logrusLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	log.Logrus().WithContext(ctx).Warnf(msg, data...)
}

func (l *logrusLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	log.Logrus().WithContext(ctx).Errorf(msg, data...)
}
