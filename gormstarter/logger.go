package gormstarter

import (
	"context"
	"time"

	log "github.com/acexy/golang-toolkit/logger"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm/logger"
)

type logrusLogger struct {
	level log.Level
}

func effectiveSQLLoggerLevel(level log.Level) log.Level {
	if level < log.InfoLevel {
		return log.DebugLevel
	}
	return level
}

// LogMode returns this adapter without changing its configured SQL trace level.
func (l *logrusLogger) LogMode(level logger.LogLevel) logger.Interface {
	return l
}

// Trace logs SQL and affected rows at the configured level; errors always log at ERROR.
func (l *logrusLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if err != nil {
		elapsed := time.Since(begin)
		sql, rows := fc()
		log.Logrus().WithContext(ctx).Errorln(sql, "rows:", rows, "elapsed:", elapsed, err)
		return
	}
	if log.IsLevelEnabled(l.level) {
		elapsed := time.Since(begin)
		sql, rows := fc()
		log.Logrus().WithContext(ctx).Logln(logrus.Level(l.level), sql, "rows:", rows, "elapsed:", elapsed)
	}
}

// Info forwards formatted GORM informational events to the application logger.
func (l *logrusLogger) Info(ctx context.Context, msg string, data ...any) {
	log.Logrus().WithContext(ctx).Infof(msg, data...)
}

// Warn forwards formatted GORM warnings to the application logger.
func (l *logrusLogger) Warn(ctx context.Context, msg string, data ...any) {
	log.Logrus().WithContext(ctx).Warnf(msg, data...)
}

// Error forwards formatted GORM errors to the application logger.
func (l *logrusLogger) Error(ctx context.Context, msg string, data ...any) {
	log.Logrus().WithContext(ctx).Errorf(msg, data...)
}
