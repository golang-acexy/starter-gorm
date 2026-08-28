package gormstarter

import (
	"testing"

	"github.com/acexy/golang-toolkit/logger"
	"gorm.io/gorm"
)

func TestRawGormDBUsesPublishedRuntimeSnapshot(t *testing.T) {
	previousRuntime := gormRuntimeState.Swap(nil)
	previousState := gormState
	t.Cleanup(func() {
		gormRuntimeState.Store(previousRuntime)
		gormState = previousState
	})

	mysqlDB := new(gorm.DB)
	postgresDB := new(gorm.DB)
	gormRuntimeState.Store(&gormRuntime{
		databases: map[DBType]*gorm.DB{
			DBTypeMySQL:    mysqlDB,
			DBTypePostgres: postgresDB,
		},
		defaultDBType: DBTypeMySQL,
	})

	if db := RawGormDB(); db != mysqlDB {
		t.Fatal("默认数据库与已发布运行时快照不一致")
	}
	if db := RawPostgresGormDB(); db != postgresDB {
		t.Fatal("PostgreSQL 数据库与已发布运行时快照不一致")
	}

	gormRuntimeState.Store(nil)
	if db := RawGormDB(); db != nil {
		t.Fatal("运行时快照撤销后仍返回数据库实例")
	}
}

func TestGetConfigCachesIndependentDatabaseConfig(t *testing.T) {
	mysqlConfig := &MySQLConfig{DatabaseConfig: DatabaseConfig{Database: "original"}}
	starter := &GormStarter{Config: GormConfig{MySQL: mysqlConfig}}

	config := starter.getConfig()
	mysqlConfig.Database = "changed"

	if config.MySQL.Database != "original" {
		t.Fatalf("缓存配置被外部修改：%s", config.MySQL.Database)
	}
	if starter.getConfig() != config {
		t.Fatal("配置未按 starter 实例缓存")
	}
}

func TestEffectiveSQLLoggerLevel(t *testing.T) {
	if level := effectiveSQLLoggerLevel(0); level != logger.DebugLevel {
		t.Fatalf("默认 SQL 日志级别 = %d，期望 %d", level, logger.DebugLevel)
	}
}
