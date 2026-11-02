//go:build integration

package pool_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
	"weak"

	applog "github.com/acexy/golang-toolkit/logger"
	"github.com/golang-acexy/starter-gorm/gormstarter"
	"gorm.io/gorm"
)

type pressureRow struct {
	ID      int64
	Payload string
}

// The derived table exercises real drivers without reading or modifying application tables.
var pressureTable = func() string {
	parts := make([]string, 32)
	for i := range parts {
		parts[i] = fmt.Sprintf("SELECT %d AS id, REPEAT('x', 1024) AS payload", i+1)
	}
	return "(" + strings.Join(parts, " UNION ALL ") + ") AS pressure_rows"
}()

func (pressureRow) TableName() string { return pressureTable }

type invalidScanRow struct{ ID int64 }

func (invalidScanRow) TableName() string { return "(SELECT 'not-a-number' AS id) AS pressure_invalid" }

var pressureID = gormstarter.NewColumn(func(r *pressureRow) *int64 { return &r.ID })

func quietPressureLogs(t *testing.T) {
	t.Helper()
	logger := applog.Logrus()
	previous := logger.Out
	logger.SetOutput(io.Discard)
	t.Cleanup(func() { logger.SetOutput(previous) })
}

func pressureStart(t *testing.T, dialect gormstarter.DBType, capacity int) (*gormstarter.GormStarter, *gorm.DB, *sql.DB) {
	t.Helper()
	config := configFor(t, dialect, gormstarter.PoolConfig{MaxOpenConns: &capacity, MaxIdleConns: &capacity})
	starter := &gormstarter.GormStarter{Config: config}
	if _, err := starter.Start(); err != nil {
		t.Fatal("pressure database startup failed")
	}
	db := gormstarter.RawGormDB(dialect)
	pool, err := db.DB()
	if err != nil {
		t.Fatal("could not access pressure pool")
	}
	t.Cleanup(func() {
		if gormstarter.RawGormDB(dialect) == nil {
			return
		}
		graceful, stopped, err := starter.Stop(5 * time.Second)
		if err != nil || !graceful || !stopped {
			t.Error("pressure pool shutdown failed")
		}
		if stats := pool.Stats(); stats.OpenConnections != 0 || stats.InUse != 0 {
			t.Errorf("pool remains open: %+v", stats)
		}
	})
	return starter, db, pool
}

func pressureOperation(ctx context.Context, db *gorm.DB, mode string, iteration int) ([]*pressureRow, error) {
	var base gormstarter.BaseMapper[pressureRow]
	mapper := base.GetBaseMapperWithTx(db.WithContext(ctx))
	var rows []*pressureRow
	var err error
	switch mode {
	case "gorm":
		err = db.WithContext(ctx).Table(pressureTable).Where("id >= ?", 1).Order("id ASC").Limit(16).Scan(&rows).Error
	case "where":
		_, err = mapper.SelectByWhere(gormstarter.NewWhereQuery("id >= ?", 1).OrderBy("id ASC").WithLimit(16), &rows)
	case "wrapper":
		_, err = mapper.SelectByWrapper(mapper.Wrapper().Where(pressureID.Ge(1)).OrderByAsc(pressureID).Limit(16), &rows)
	case "page":
		var total int64
		total, err = mapper.SelectPageByWrapper(mapper.PageWrapper(2, 8).Where(pressureID.Ge(1)).OrderByAsc(pressureID), &rows)
		if err == nil && total != 32 {
			return nil, fmt.Errorf("unexpected page total: %d", total)
		}
	case "transaction":
		transactionMapper := base.NewBaseMapperWithTx(&sql.TxOptions{ReadOnly: true})
		tx := transactionMapper.CurrentGormDB()
		if tx.Error != nil {
			return nil, tx.Error
		}
		defer tx.Rollback()
		_, err = transactionMapper.GetBaseMapperWithTx(tx.WithContext(ctx)).SelectByWhere(gormstarter.NewWhereQuery("id >= ?", 1).OrderBy("id ASC").WithLimit(16), &rows)
		if err == nil {
			if iteration%2 == 0 {
				err = tx.Commit().Error
			} else {
				err = tx.Rollback().Error
			}
		}
	}
	if err != nil {
		return nil, err
	}
	count, firstID := 16, int64(1)
	if mode == "page" {
		count, firstID = 8, 9
	}
	if len(rows) != count {
		return nil, fmt.Errorf("unexpected row count: %d", len(rows))
	}
	for i, row := range rows {
		if row == nil || row.ID != firstID+int64(i) || len(row.Payload) != 1024 {
			return nil, errors.New("result validation failed")
		}
	}
	return rows, nil
}

func pressureBatch(t *testing.T, db *gorm.DB, mode string, count int) []weak.Pointer[pressureRow] {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	var wg sync.WaitGroup
	var mu sync.Mutex
	refs := make([]weak.Pointer[pressureRow], 0, count*16)
	for worker := 0; worker < 16; worker++ {
		wg.Add(1)
		go func(worker int) {
			defer wg.Done()
			local := make([]weak.Pointer[pressureRow], 0, (count/16+1)*16)
			for i := worker; i < count; i += 16 {
				rows, err := pressureOperation(ctx, db, mode, i)
				if err != nil {
					t.Errorf("%s operation failed (%T)", mode, err)
					return
				}
				for _, row := range rows {
					local = append(local, weak.Make(row))
				}
			}
			mu.Lock()
			refs = append(refs, local...)
			mu.Unlock()
		}(worker)
	}
	wg.Wait()
	// Check before canceling the batch context so cancellation cannot hide unclosed rows.
	pool, err := db.DB()
	if err != nil {
		t.Error("could not inspect pool after batch")
	} else if pool.Stats().InUse != 0 {
		t.Error("batch retained connections before context cancellation")
	}
	return refs
}

func TestDatabaseMemoryPressure(t *testing.T) {
	if os.Getenv("GORM_MEMORY_STRESS") != "1" {
		t.Skip("set GORM_MEMORY_STRESS=1 for real database pressure measurements")
	}
	quietPressureLogs(t)
	for _, dialect := range []gormstarter.DBType{gormstarter.DBTypeMySQL, gormstarter.DBTypePostgres} {
		t.Run(string(dialect), func(t *testing.T) {
			_, db, pool := pressureStart(t, dialect, 8)
			var version string
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			err := db.WithContext(ctx).Raw("SELECT VERSION()").Scan(&version).Error
			cancel()
			if err != nil {
				t.Fatal("could not read database version")
			}
			t.Logf("database_version=%s", version)
			for _, mode := range []string{"gorm", "where", "wrapper", "page", "transaction"} {
				t.Run(mode, func(t *testing.T) {
					pressureBatch(t, db, mode, 100)
					if t.Failed() {
						return
					}
					runtime.GC()
					runtime.GC()
					var baseline runtime.MemStats
					runtime.ReadMemStats(&baseline)
					baseGoroutines, baseStats := runtime.NumGoroutine(), pool.Stats()
					for batch := 1; batch <= 5; batch++ {
						started := time.Now()
						refs := pressureBatch(t, db, mode, 500)
						elapsed := time.Since(started)
						if t.Failed() {
							return
						}
						expected := 500 * 16
						if mode == "page" {
							expected = 500 * 8
						}
						if len(refs) != expected {
							t.Fatalf("unexpected observed results: %d", len(refs))
						}
						runtime.GC()
						runtime.GC()
						live := 0
						for _, ref := range refs {
							if ref.Value() != nil {
								live++
							}
						}
						stats := pool.Stats()
						var memory runtime.MemStats
						runtime.ReadMemStats(&memory)
						t.Logf("batch=%d ops=%d ops_s=%.0f live_results=%d heap=%d heap_delta=%d total_alloc_delta=%d goroutines=%d base_goroutines=%d in_use=%d idle=%d open=%d waits=%d wait_ms=%d", batch, batch*500, 500/elapsed.Seconds(), live, memory.HeapAlloc, int64(memory.HeapAlloc)-int64(baseline.HeapAlloc), memory.TotalAlloc-baseline.TotalAlloc, runtime.NumGoroutine(), baseGoroutines, stats.InUse, stats.Idle, stats.OpenConnections, stats.WaitCount-baseStats.WaitCount, (stats.WaitDuration - baseStats.WaitDuration).Milliseconds())
						if live != 0 || stats.InUse != 0 || stats.OpenConnections > 8 {
							t.Fatal("result or connection retention detected")
						}
					}
				})
			}
		})
	}
}

func TestDatabaseFailureAndTransactionRelease(t *testing.T) {
	quietPressureLogs(t)
	for _, dialect := range []gormstarter.DBType{gormstarter.DBTypeMySQL, gormstarter.DBTypePostgres} {
		t.Run(string(dialect), func(t *testing.T) {
			_, db, pool := pressureStart(t, dialect, 1)
			for i := 0; i < 50; i++ {
				var mapper gormstarter.BaseMapper[invalidScanRow]
				var result []*invalidScanRow
				if _, err := mapper.SelectByWhere(gormstarter.NewWhereQuery("1=1"), &result); err == nil {
					t.Fatal("expected scan conversion error")
				}
				if pool.Stats().InUse != 0 {
					t.Fatal("failed scan retained connection")
				}
			}
			var mapper gormstarter.BaseMapper[pressureRow]
			transaction := mapper.NewBaseMapperWithTx(&sql.TxOptions{ReadOnly: true}).CurrentGormDB()
			if transaction.Error != nil {
				t.Fatal("could not begin read-only transaction")
			}
			defer transaction.Rollback()
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
			conn, err := pool.Conn(ctx)
			cancel()
			if conn != nil {
				conn.Close()
			}
			if !errors.Is(err, context.DeadlineExceeded) || pool.Stats().InUse != 1 {
				t.Fatal("transaction did not retain its owned connection")
			}
			if transaction.Rollback().Error != nil || pool.Stats().InUse != 0 {
				t.Fatal("rollback did not release connection")
			}
			sleepSQL := "SELECT SLEEP(0.2)"
			if dialect == gormstarter.DBTypePostgres {
				sleepSQL = "SELECT pg_sleep(0.2)"
			}
			for i := 0; i < 5; i++ {
				ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
				var result any
				err := db.WithContext(ctx).Raw(sleepSQL).Scan(&result).Error
				contextErr := ctx.Err()
				cancel()
				if err == nil || contextErr == nil {
					t.Fatal("query cancellation did not occur")
				}
				eventually(t, func() bool { return pool.Stats().InUse == 0 })
				ctx, cancel = context.WithTimeout(context.Background(), time.Second)
				err = pool.PingContext(ctx)
				cancel()
				if err != nil {
					t.Fatal("pool did not recover after cancellation")
				}
			}
		})
	}
}

type databaseCycleRefs struct {
	orm  weak.Pointer[gorm.DB]
	pool weak.Pointer[sql.DB]
}

func databaseCycle(t *testing.T, dialect gormstarter.DBType) databaseCycleRefs {
	t.Helper()
	starter := &gormstarter.GormStarter{Config: configFor(t, dialect, gormstarter.PoolConfig{MaxOpenConns: new(2), MaxIdleConns: new(2)})}
	if _, err := starter.Start(); err != nil {
		t.Fatal("lifecycle startup failed")
	}
	defer func() {
		if gormstarter.RawGormDB(dialect) != nil {
			starter.Stop(time.Second)
		}
	}()
	db := gormstarter.RawGormDB(dialect)
	pool, err := db.DB()
	if err != nil {
		t.Fatal("lifecycle pool unavailable")
	}
	refs := databaseCycleRefs{weak.Make(db), weak.Make(pool)}
	graceful, stopped, err := starter.Stop(time.Second)
	if err != nil || !graceful || !stopped {
		t.Fatal("lifecycle shutdown failed")
	}
	if stats := pool.Stats(); stats.OpenConnections != 0 || stats.InUse != 0 {
		t.Fatal("shutdown did not close connections")
	}
	return refs
}

func TestDatabaseLifecycleReleasesPools(t *testing.T) {
	quietPressureLogs(t)
	for _, dialect := range []gormstarter.DBType{gormstarter.DBTypeMySQL, gormstarter.DBTypePostgres} {
		t.Run(string(dialect), func(t *testing.T) {
			refs := make([]databaseCycleRefs, 0, 20)
			for i := 0; i < 20; i++ {
				refs = append(refs, databaseCycle(t, dialect))
			}
			eventually(t, func() bool {
				runtime.GC()
				runtime.GC()
				for _, ref := range refs {
					if ref.orm.Value() != nil || ref.pool.Value() != nil {
						return false
					}
				}
				return true
			})
		})
	}
}

func TestDatabaseStopWithBorrowedTransaction(t *testing.T) {
	quietPressureLogs(t)
	for _, dialect := range []gormstarter.DBType{gormstarter.DBTypeMySQL, gormstarter.DBTypePostgres} {
		t.Run(string(dialect), func(t *testing.T) {
			starter, _, pool := pressureStart(t, dialect, 1)
			var mapper gormstarter.BaseMapper[pressureRow]
			tx := mapper.NewBaseMapperWithTx(&sql.TxOptions{ReadOnly: true}).CurrentGormDB()
			if tx.Error != nil {
				t.Fatal("could not begin held transaction")
			}
			defer tx.Rollback()
			graceful, stopped, err := starter.Stop(time.Millisecond)
			if !errors.Is(err, gormstarter.ErrGormStopTimeout) || graceful || !stopped || gormstarter.RawGormDB(dialect) != nil {
				t.Fatal("unexpected shutdown state with a borrowed connection")
			}
			if pool.Stats().InUse != 1 {
				t.Fatal("borrowed transaction unexpectedly released")
			}
			if tx.Rollback().Error != nil {
				t.Fatal("held transaction rollback failed")
			}
			eventually(t, func() bool { stats := pool.Stats(); return stats.InUse == 0 && stats.OpenConnections == 0 })
		})
	}
}
