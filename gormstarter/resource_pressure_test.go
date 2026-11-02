package gormstarter

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"sync/atomic"
	"testing"

	applog "github.com/acexy/golang-toolkit/logger"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type failingPingConnector struct {
	pings, closes atomic.Int64
	failOn        int64
}
type failingPingConn struct{ owner *failingPingConnector }
type unusedPressureDriver struct{}

var errPressurePing = errors.New("pressure ping failure")

func (c *failingPingConnector) Connect(context.Context) (driver.Conn, error) {
	return &failingPingConn{c}, nil
}
func (*failingPingConnector) Driver() driver.Driver           { return unusedPressureDriver{} }
func (unusedPressureDriver) Open(string) (driver.Conn, error) { return nil, driver.ErrSkip }
func (*failingPingConn) Prepare(string) (driver.Stmt, error)  { return nil, driver.ErrSkip }
func (*failingPingConn) Begin() (driver.Tx, error)            { return nil, driver.ErrSkip }
func (c *failingPingConn) Close() error                       { c.owner.closes.Add(1); return nil }
func (c *failingPingConn) Ping(context.Context) error {
	if c.owner.pings.Add(1) == c.owner.failOn {
		return errPressurePing
	}
	return nil
}

func TestFailedStartupPingClosesPool(t *testing.T) {
	log := applog.Logrus()
	previous := log.Out
	log.SetOutput(io.Discard)
	defer log.SetOutput(previous)
	for _, failOn := range []int64{1, 2} {
		connector := &failingPingConnector{failOn: failOn}
		pool := sql.OpenDB(connector)
		starter := &GormStarter{}
		_, err := starter.openDatabase(DBTypeMySQL, DatabaseConfig{}, func(config *gorm.Config) (*gorm.DB, error) {
			return gorm.Open(mysql.New(mysql.Config{Conn: pool, SkipInitializeWithVersion: true}), config)
		})
		stats := pool.Stats()
		closedPing := pool.Ping()
		pool.Close()
		if !errors.Is(err, errPressurePing) || connector.pings.Load() != failOn || connector.closes.Load() != 1 || stats.OpenConnections != 0 || closedPing == nil {
			t.Fatalf("failed ping cleanup: phase=%d pings=%d closes=%d open=%d error=%v", failOn, connector.pings.Load(), connector.closes.Load(), stats.OpenConnections, err)
		}
	}
}

type pressureQueryModel struct {
	ID   int64
	Name string
}

func (pressureQueryModel) TableName() string { return "pressure_query" }

func BenchmarkQueryConstruction(b *testing.B) {
	pool := sql.OpenDB(poolTestConnector{})
	b.Cleanup(func() { pool.Close() })
	db, err := gorm.Open(mysql.New(mysql.Config{Conn: pool, SkipInitializeWithVersion: true}), &gorm.Config{DisableAutomaticPing: true, DryRun: true, Logger: logger.Discard})
	if err != nil {
		b.Fatal(err)
	}
	id := NewColumn(func(m *pressureQueryModel) *int64 { return &m.ID })
	name := NewColumn(func(m *pressureQueryModel) *string { return &m.Name })
	for _, mode := range []string{"gorm", "wrapper"} {
		b.Run(mode, func(b *testing.B) {
			build := func() *gorm.DB {
				tx := db.Table("pressure_query")
				if mode == "gorm" {
					return tx.Where("id >= ?", 1).Where("name LIKE ?", "a%").Select("id", "name").Order("id ASC").Limit(16)
				}
				query := NewQueryWrapper[pressureQueryModel]().Where(id.Ge(1), name.Like("a%")).Select(id, name).OrderByAsc(id).Limit(16)
				tx, err := applyWrapper(tx, query, true, true, true)
				if err != nil {
					b.Fatal(err)
				}
				return tx
			}
			var warmup []pressureQueryModel
			if err := build().Find(&warmup).Error; err != nil {
				b.Fatal(err)
			}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				var result []pressureQueryModel
				tx := build().Find(&result)
				if tx.Error != nil || tx.Statement.SQL.Len() == 0 {
					b.Fatal("SQL construction failed")
				}
			}
		})
	}
}
