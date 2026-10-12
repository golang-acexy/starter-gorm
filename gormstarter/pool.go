package gormstarter

import (
	"database/sql"
	"fmt"
	"time"
)

// PoolConfig configures an independent SQL pool for either dialect.
// Nil fields preserve defaults; negative values are rejected before opening databases.
type PoolConfig struct {
	// MaxOpenConns limits open connections, including idle ones; zero means unlimited.
	MaxOpenConns *int
	// MaxIdleConns limits retained idle connections; zero retains none.
	// A positive MaxOpenConns caps the effective idle limit.
	MaxIdleConns *int
	// ConnMaxIdleTime limits connection idle time; zero disables the limit.
	ConnMaxIdleTime *time.Duration
	// ConnMaxLifetime limits connection reuse age; zero disables the limit.
	// Expiration does not interrupt an operation already using the connection.
	ConnMaxLifetime *time.Duration
}

func clonePoolValue[T any](value *T) *T {
	if value == nil {
		return nil
	}
	copy := *value
	return &copy
}

func clonePoolConfig(config PoolConfig) PoolConfig {
	return PoolConfig{
		MaxOpenConns:    clonePoolValue(config.MaxOpenConns),
		MaxIdleConns:    clonePoolValue(config.MaxIdleConns),
		ConnMaxIdleTime: clonePoolValue(config.ConnMaxIdleTime),
		ConnMaxLifetime: clonePoolValue(config.ConnMaxLifetime),
	}
}

func validatePoolConfig(config PoolConfig) error {
	switch {
	case config.MaxOpenConns != nil && *config.MaxOpenConns < 0:
		return fmt.Errorf("%w: MaxOpenConns must not be negative", ErrInvalidPoolConfig)
	case config.MaxIdleConns != nil && *config.MaxIdleConns < 0:
		return fmt.Errorf("%w: MaxIdleConns must not be negative", ErrInvalidPoolConfig)
	case config.ConnMaxIdleTime != nil && *config.ConnMaxIdleTime < 0:
		return fmt.Errorf("%w: ConnMaxIdleTime must not be negative", ErrInvalidPoolConfig)
	case config.ConnMaxLifetime != nil && *config.ConnMaxLifetime < 0:
		return fmt.Errorf("%w: ConnMaxLifetime must not be negative", ErrInvalidPoolConfig)
	}
	return nil
}

func applyPoolConfig(db *sql.DB, config PoolConfig) {
	if config.MaxOpenConns != nil {
		db.SetMaxOpenConns(*config.MaxOpenConns)
	}
	if config.MaxIdleConns != nil {
		db.SetMaxIdleConns(*config.MaxIdleConns)
	}
	if config.ConnMaxIdleTime != nil {
		db.SetConnMaxIdleTime(*config.ConnMaxIdleTime)
	}
	if config.ConnMaxLifetime != nil {
		db.SetConnMaxLifetime(*config.ConnMaxLifetime)
	}
}
