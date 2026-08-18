package gormstarter

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type querySQLRecorder struct {
	sql string
}

func (r *querySQLRecorder) LogMode(logger.LogLevel) logger.Interface { return r }

func (r *querySQLRecorder) Info(context.Context, string, ...any) {}

func (r *querySQLRecorder) Warn(context.Context, string, ...any) {}

func (r *querySQLRecorder) Error(context.Context, string, ...any) {}

func (r *querySQLRecorder) Trace(_ context.Context, _ time.Time, fc func() (string, int64), _ error) {
	r.sql, _ = fc()
}

func TestSelectOneByWhereDoesNotAppendLimit(t *testing.T) {
	recorder := &querySQLRecorder{}
	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN:                       "gorm:gorm@tcp(localhost:9910)/gorm?charset=utf8&parseTime=True&loc=Local",
		SkipInitializeWithVersion: true,
	}), &gorm.Config{
		DryRun:               true,
		DisableAutomaticPing: true,
		Logger:               recorder,
	})
	if err != nil {
		t.Fatalf("创建 DryRun 数据库失败：%v", err)
	}

	mapper := BaseMapper[validationModel]{model: validationModel{}, tx: db}
	var result validationModel
	_, err = mapper.SelectOneByWhere(WhereQuery{
		RawWhereSQL: "id = ? FOR UPDATE",
		Args:        []any{1},
	}, &result)
	if err != nil && !errors.Is(err, gorm.ErrDryRunModeUnsupported) {
		t.Fatalf("SelectOneByWhere 查询失败：%v", err)
	}

	upperSQL := strings.ToUpper(recorder.sql)
	if !strings.Contains(upperSQL, "FOR UPDATE") {
		t.Fatalf("生成的 SQL 未保留自定义尾部语句：%s", recorder.sql)
	}
	if strings.Contains(upperSQL, "LIMIT 1") {
		t.Fatalf("生成的 SQL 不应自动追加 LIMIT 1：%s", recorder.sql)
	}
}
