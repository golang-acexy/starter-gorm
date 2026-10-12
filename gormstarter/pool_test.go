package gormstarter

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"testing"
	"testing/synctest"

	"gorm.io/gorm"
	"time"
)

type poolTestConnector struct{}
type poolTestDriver struct{}
type poolTestConn struct{}

func (poolTestConnector) Connect(context.Context) (driver.Conn, error) { return poolTestConn{}, nil }
func (poolTestConnector) Driver() driver.Driver                        { return poolTestDriver{} }
func (poolTestDriver) Open(string) (driver.Conn, error)                { return poolTestConn{}, nil }
func (poolTestConn) Prepare(string) (driver.Stmt, error)               { return nil, errors.New("unsupported") }
func (poolTestConn) Close() error                                      { return nil }
func (poolTestConn) Begin() (driver.Tx, error)                         { return nil, errors.New("unsupported") }

func TestPoolConfigBehavior(t *testing.T) {
	for _, tc := range []struct {
		name              string
		config            PoolConfig
		wantMax, wantIdle int
	}{
		{"defaults", PoolConfig{}, 0, 2},
		{"open cap", PoolConfig{MaxOpenConns: new(1)}, 1, 1},
		{"idle cap", PoolConfig{MaxOpenConns: new(2), MaxIdleConns: new(5)}, 2, 2},
		{"no idle", PoolConfig{MaxIdleConns: new(0)}, 0, 0},
		{"unlimited", PoolConfig{MaxOpenConns: new(0)}, 0, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db := sql.OpenDB(poolTestConnector{})
			defer db.Close()
			applyPoolConfig(db, tc.config)
			count := 3
			if tc.wantMax > 0 {
				count = tc.wantMax
			}
			conns := make([]*sql.Conn, count)
			for i := range conns {
				conn, err := db.Conn(context.Background())
				if err != nil {
					t.Fatal(err)
				}
				conns[i] = conn
			}
			for _, conn := range conns {
				if err := conn.Close(); err != nil {
					t.Fatal(err)
				}
			}
			stats := db.Stats()
			if stats.MaxOpenConnections != tc.wantMax || stats.Idle != tc.wantIdle {
				t.Fatalf("unexpected pool stats: %+v", stats)
			}
		})
	}
}

func TestPoolConfigLifetimeAndIdleTime(t *testing.T) {
	for _, idle := range []bool{false, true} {
		t.Run(map[bool]string{false: "lifetime", true: "idle time"}[idle], func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				db := sql.OpenDB(poolTestConnector{})
				defer db.Close()
				config := PoolConfig{}
				if idle {
					config.ConnMaxIdleTime = new(time.Second)
				} else {
					config.ConnMaxLifetime = new(time.Second)
				}
				applyPoolConfig(db, config)
				if err := db.Ping(); err != nil {
					t.Fatal(err)
				}
				// Omitted durations must survive unrelated updates. Virtual time runs the SQL cleaner deterministically.
				applyPoolConfig(db, PoolConfig{MaxOpenConns: new(3)})
				time.Sleep(2 * time.Second)
				synctest.Wait()
				stats := db.Stats()
				if (idle && stats.MaxIdleTimeClosed == 0) || (!idle && stats.MaxLifetimeClosed == 0) {
					t.Fatalf("expiration not applied: %+v", stats)
				}
				zero := time.Duration(0)
				applyPoolConfig(db, PoolConfig{ConnMaxIdleTime: &zero, ConnMaxLifetime: &zero})
				if err := db.Ping(); err != nil {
					t.Fatal(err)
				}
				before := db.Stats()
				time.Sleep(2 * time.Second)
				synctest.Wait()
				if err := db.Ping(); err != nil {
					t.Fatal(err)
				}
				after := db.Stats()
				if before.MaxLifetimeClosed != after.MaxLifetimeClosed || before.MaxIdleTimeClosed != after.MaxIdleTimeClosed {
					t.Fatal("explicit zero did not disable expiration")
				}
			})
		})
	}
}

func TestPoolConfigSnapshot(t *testing.T) {
	open, idle := 8, 3
	idleTime, lifetime := time.Minute, time.Hour
	pool := PoolConfig{&open, &idle, &idleTime, &lifetime}
	starter := &GormStarter{Config: GormConfig{
		MySQL:    &MySQLConfig{DatabaseConfig: DatabaseConfig{Pool: pool}},
		Postgres: &PostgresConfig{DatabaseConfig: DatabaseConfig{Pool: pool}},
	}}
	config := starter.getConfig()
	open, idle, idleTime, lifetime = 0, 0, 0, 0
	for _, snapshot := range []PoolConfig{config.MySQL.Pool, config.Postgres.Pool} {
		if *snapshot.MaxOpenConns != 8 || *snapshot.MaxIdleConns != 3 || *snapshot.ConnMaxIdleTime != time.Minute || *snapshot.ConnMaxLifetime != time.Hour {
			t.Fatal("external mutation changed cached pool configuration")
		}
	}
	*config.MySQL.Pool.MaxOpenConns = 2
	if *config.Postgres.Pool.MaxOpenConns != 8 {
		t.Fatal("dialects share pool configuration pointers")
	}
}

func TestStartRejectsInvalidPoolsBeforeOpeningDatabases(t *testing.T) {
	for _, pool := range []PoolConfig{
		{MaxOpenConns: new(-1)}, {MaxIdleConns: new(-1)},
		{ConnMaxIdleTime: new(-time.Second)}, {ConnMaxLifetime: new(-time.Second)},
	} {
		for _, mysql := range []bool{false, true} {
			config := GormConfig{MySQL: &MySQLConfig{}}
			if mysql {
				config.MySQL.Pool = pool
			} else {
				config.Postgres = &PostgresConfig{DatabaseConfig: DatabaseConfig{Pool: pool}}
			}
			starter := &GormStarter{Config: config}
			if _, err := starter.Start(); !errors.Is(err, ErrInvalidPoolConfig) {
				t.Fatalf("expected pool error, got %v", err)
			}
		}
	}
}

func TestOpenDatabaseAppliesPoolConfig(t *testing.T) {
	for _, dialect := range []DBType{DBTypeMySQL, DBTypePostgres} {
		t.Run(string(dialect), func(t *testing.T) {
			sqlDB := sql.OpenDB(poolTestConnector{})
			defer sqlDB.Close()
			starter := &GormStarter{}
			config := DatabaseConfig{Pool: PoolConfig{MaxOpenConns: new(7), MaxIdleConns: new(0)}}
			_, err := starter.openDatabase(dialect, config, func(raw *gorm.Config) (*gorm.DB, error) {
				if raw.DisableAutomaticPing {
					t.Fatal("automatic Ping behavior changed")
				}
				return &gorm.DB{Config: &gorm.Config{ConnPool: sqlDB}}, nil
			})
			if err != nil {
				t.Fatal(err)
			}
			stats := sqlDB.Stats()
			if stats.MaxOpenConnections != 7 || stats.Idle != 0 || stats.MaxIdleClosed == 0 {
				t.Fatalf("pool settings were not applied before explicit Ping: %+v", stats)
			}
		})
	}
}
