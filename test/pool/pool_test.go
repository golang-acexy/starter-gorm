//go:build integration

package pool_test

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/golang-acexy/starter-gorm/gormstarter"
)

func databaseConfig(t *testing.T, prefix string, port uint, pool gormstarter.PoolConfig) gormstarter.DatabaseConfig {
	t.Helper()
	for _, suffix := range []string{"HOST", "DATABASE", "USERNAME"} {
		if os.Getenv(prefix+suffix) == "" {
			t.Fatalf("configure %s%s before running pool integration tests", prefix, suffix)
		}
	}
	if value := os.Getenv(prefix + "PORT"); value != "" {
		parsed, err := strconv.ParseUint(value, 10, 16)
		if err != nil || parsed == 0 {
			t.Fatalf("invalid %sPORT", prefix)
		}
		port = uint(parsed)
	}
	return gormstarter.DatabaseConfig{
		Host: os.Getenv(prefix + "HOST"), Port: port,
		Database: os.Getenv(prefix + "DATABASE"),
		Username: os.Getenv(prefix + "USERNAME"), Password: os.Getenv(prefix + "PASSWORD"),
		Pool: pool,
	}
}

func configFor(t *testing.T, dialect gormstarter.DBType, pool gormstarter.PoolConfig) gormstarter.GormConfig {
	t.Helper()
	if dialect == gormstarter.DBTypeMySQL {
		return gormstarter.GormConfig{MySQL: &gormstarter.MySQLConfig{
			DatabaseConfig: databaseConfig(t, "STARTER_GORM_MYSQL_", 13306, pool),
			Charset:        "utf8mb4", URLParams: "timeout=3s&readTimeout=5s&writeTimeout=5s",
		}}
	}
	return gormstarter.GormConfig{Postgres: &gormstarter.PostgresConfig{
		DatabaseConfig: databaseConfig(t, "STARTER_GORM_POSTGRES_", 5432, pool),
	}}
}

func start(t *testing.T, config gormstarter.GormConfig) {
	t.Helper()
	starter := &gormstarter.GormStarter{Config: config}
	if _, err := starter.Start(); err != nil {
		t.Fatal("database startup failed; check the integration connection settings and database availability")
	}
	t.Cleanup(func() {
		graceful, stopped, err := starter.Stop(5 * time.Second)
		if err != nil || !graceful || !stopped {
			t.Errorf("starter cleanup failed: graceful=%v stopped=%v", graceful, stopped)
		}
	})
}

func sqlPool(t *testing.T, dialect gormstarter.DBType) *sql.DB {
	t.Helper()
	db := gormstarter.RawGormDB(dialect)
	if db == nil {
		t.Fatal("database was not published")
	}
	pool, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	return pool
}

func acquire(t *testing.T, db *sql.DB) *sql.Conn {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	conn, err := db.Conn(ctx)
	if err != nil {
		t.Fatal("could not acquire database connection")
	}
	t.Cleanup(func() { _ = conn.Close() })
	return conn
}

func selectOne(t *testing.T, conn *sql.Conn) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var value int
	if err := conn.QueryRowContext(ctx, "SELECT 1").Scan(&value); err != nil || value != 1 {
		t.Fatal("SELECT 1 failed")
	}
}

func eventually(t *testing.T, condition func() bool) {
	t.Helper()
	timer := time.NewTimer(5 * time.Second)
	defer timer.Stop()
	ticker := time.NewTicker(20 * time.Millisecond)
	defer ticker.Stop()
	for !condition() {
		select {
		case <-timer.C:
			t.Fatal("pool did not reach the expected state within five seconds")
		case <-ticker.C:
		}
	}
}

func TestPoolConfiguration(t *testing.T) {
	// Starter state is process-wide; these cases must run sequentially.
	for _, dialect := range []gormstarter.DBType{gormstarter.DBTypeMySQL, gormstarter.DBTypePostgres} {
		t.Run(string(dialect), func(t *testing.T) {
			t.Run("limits_and_wait", func(t *testing.T) {
				start(t, configFor(t, dialect, gormstarter.PoolConfig{MaxOpenConns: new(2), MaxIdleConns: new(1)}))
				db := sqlPool(t, dialect)
				first, second := acquire(t, db), acquire(t, db)
				selectOne(t, first)
				selectOne(t, second)
				before := db.Stats()
				if before.MaxOpenConnections != 2 || before.InUse != 2 || before.OpenConnections != 2 {
					t.Fatalf("unexpected pool stats: %+v", before)
				}
				ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
				defer cancel()
				third, err := db.Conn(ctx)
				if third != nil {
					_ = third.Close()
				}
				if !errors.Is(err, context.DeadlineExceeded) {
					t.Fatalf("expected capacity wait deadline, got %v", err)
				}
				after := db.Stats()
				if after.WaitCount <= before.WaitCount || after.WaitDuration <= before.WaitDuration || after.OpenConnections != 2 {
					t.Fatalf("pool capacity was not enforced: %+v", after)
				}
				if err := first.Close(); err != nil {
					t.Fatal(err)
				}
				replacement := acquire(t, db)
				selectOne(t, replacement)
				_ = replacement.Close()
				_ = second.Close()
				stats := db.Stats()
				if stats.InUse != 0 || stats.Idle != 1 || stats.MaxIdleClosed == 0 {
					t.Fatalf("idle limit was not enforced: %+v", stats)
				}
			})
			for _, tc := range []struct {
				name                     string
				pool                     gormstarter.PoolConfig
				count, wantMax, wantIdle int
			}{
				{"defaults", gormstarter.PoolConfig{}, 3, 0, 2},
				{"explicit_zero", gormstarter.PoolConfig{MaxOpenConns: new(0), MaxIdleConns: new(0), ConnMaxIdleTime: new(time.Duration(0)), ConnMaxLifetime: new(time.Duration(0))}, 3, 0, 0},
				{"idle_capped_by_open", gormstarter.PoolConfig{MaxOpenConns: new(1), MaxIdleConns: new(3)}, 1, 1, 1},
			} {
				t.Run(tc.name, func(t *testing.T) {
					start(t, configFor(t, dialect, tc.pool))
					db := sqlPool(t, dialect)
					conns := make([]*sql.Conn, tc.count)
					for i := range conns {
						conns[i] = acquire(t, db)
						selectOne(t, conns[i])
					}
					for _, conn := range conns {
						if err := conn.Close(); err != nil {
							t.Fatal(err)
						}
					}
					stats := db.Stats()
					if stats.MaxOpenConnections != tc.wantMax || stats.Idle != tc.wantIdle || stats.InUse != 0 {
						t.Fatalf("unexpected pool stats: %+v", stats)
					}
				})
			}
			t.Run("idle_expiration", func(t *testing.T) {
				start(t, configFor(t, dialect, gormstarter.PoolConfig{MaxOpenConns: new(1), MaxIdleConns: new(1), ConnMaxIdleTime: new(100 * time.Millisecond)}))
				db := sqlPool(t, dialect)
				conn := acquire(t, db)
				selectOne(t, conn)
				before := db.Stats().MaxIdleTimeClosed
				if err := conn.Close(); err != nil {
					t.Fatal(err)
				}
				eventually(t, func() bool {
					stats := db.Stats()
					return stats.MaxIdleTimeClosed > before && stats.OpenConnections == 0
				})
				selectOne(t, acquire(t, db))
			})
			t.Run("lifetime_preserves_in_use_connection", func(t *testing.T) {
				const lifetime = 100 * time.Millisecond
				start(t, configFor(t, dialect, gormstarter.PoolConfig{MaxOpenConns: new(1), MaxIdleConns: new(1), ConnMaxLifetime: new(lifetime)}))
				db := sqlPool(t, dialect)
				conn := acquire(t, db)
				before := db.Stats().MaxLifetimeClosed
				// Keep the real connection checked out past its configured age.
				timer := time.NewTimer(2 * lifetime)
				defer timer.Stop()
				<-timer.C
				selectOne(t, conn)
				if stats := db.Stats(); stats.InUse != 1 || stats.MaxLifetimeClosed != before {
					t.Fatalf("in-use connection expired prematurely: %+v", stats)
				}
				if err := conn.Close(); err != nil {
					t.Fatal(err)
				}
				eventually(t, func() bool { return db.Stats().MaxLifetimeClosed > before })
				selectOne(t, acquire(t, db))
			})
			if dialect == gormstarter.DBTypeMySQL {
				t.Run("url_params_coexist", func(t *testing.T) {
					start(t, configFor(t, dialect, gormstarter.PoolConfig{MaxOpenConns: new(2), MaxIdleConns: new(1)}))
					db := sqlPool(t, dialect)
					conn := acquire(t, db)
					ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					defer cancel()
					var charset string
					if err := conn.QueryRowContext(ctx, "SELECT @@character_set_connection").Scan(&charset); err != nil {
						t.Fatal("could not read connection charset")
					}
					if charset != "utf8mb4" || db.Stats().MaxOpenConnections != 2 {
						t.Fatalf("unexpected charset %q or pool limit", charset)
					}
				})
			}
		})
	}
}

func TestIndependentDatabasePools(t *testing.T) {
	mysql := configFor(t, gormstarter.DBTypeMySQL, gormstarter.PoolConfig{MaxOpenConns: new(1), MaxIdleConns: new(0)})
	postgres := configFor(t, gormstarter.DBTypePostgres, gormstarter.PoolConfig{MaxOpenConns: new(2), MaxIdleConns: new(1)})
	start(t, gormstarter.GormConfig{MySQL: mysql.MySQL, Postgres: postgres.Postgres})
	mysqlDB, postgresDB := sqlPool(t, gormstarter.DBTypeMySQL), sqlPool(t, gormstarter.DBTypePostgres)
	if mysqlDB == postgresDB {
		t.Fatal("dialects share a pool")
	}
	mysqlConn := acquire(t, mysqlDB)
	first, second := acquire(t, postgresDB), acquire(t, postgresDB)
	selectOne(t, first)
	selectOne(t, second)
	if mysqlDB.Stats().MaxOpenConnections != 1 || postgresDB.Stats().MaxOpenConnections != 2 {
		t.Fatal("dialect limits are not independent")
	}
	_ = mysqlConn.Close()
	_ = first.Close()
	_ = second.Close()
	if mysqlDB.Stats().Idle != 0 || postgresDB.Stats().Idle != 1 {
		t.Fatal("dialect idle limits are not independent")
	}
}
