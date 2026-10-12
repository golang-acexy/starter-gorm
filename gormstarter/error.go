package gormstarter

import "errors"

var (
	// ErrInvalidPoolConfig reports a negative SQL pool setting.
	ErrInvalidPoolConfig = errors.New("invalid SQL pool configuration")
	// ErrGormStarterAlreadyStarted reports gorm starter already started.
	ErrGormStarterAlreadyStarted = errors.New("gorm starter already started")
	// ErrGormStarterNotStarted reports gorm starter not started.
	ErrGormStarterNotStarted = errors.New("gorm starter not started")
	// ErrNoDatabaseConfigured reports no database configured.
	ErrNoDatabaseConfigured = errors.New("no database configured")
	// ErrDatabaseNotRegistered reports database type not registered.
	ErrDatabaseNotRegistered = errors.New("database type not registered")
	// ErrGormStopTimeout reports waiting for gorm starter shutdown timeout.
	ErrGormStopTimeout = errors.New("waiting for gorm starter shutdown timeout")
	// ErrInvalidPage reports page number and page size must be greater than zero.
	ErrInvalidPage = errors.New("page number and page size must be greater than zero")
	// ErrNoFieldToSave reports no field to save.
	ErrNoFieldToSave = errors.New("no field to save")
	// ErrNoFieldToUpdate reports no field to update.
	ErrNoFieldToUpdate = errors.New("no field to update")
	// ErrEmptyCondition reports condition must not be empty.
	ErrEmptyCondition = errors.New("condition must not be empty")
	// ErrNilEntity reports entity must not be nil.
	ErrNilEntity = errors.New("entity must not be nil")
	// ErrEmptyID reports id must not be empty.
	ErrEmptyID = errors.New("id must not be empty")
	// ErrEmptyIDs reports ids must not be empty.
	ErrEmptyIDs = errors.New("ids must not be empty")
	// ErrEmptyWhereSQL reports where SQL must not be empty.
	ErrEmptyWhereSQL = errors.New("where SQL must not be empty")
	// ErrNilColumnSelector reports column selector must not be nil.
	ErrNilColumnSelector = errors.New("column selector must not be nil")
	// ErrInvalidColumnSelector reports column selector must return exactly one model field.
	ErrInvalidColumnSelector = errors.New("column selector must return exactly one model field")
	// ErrNonPersistentColumn reports column selector references a non-persistent model field.
	ErrNonPersistentColumn = errors.New("column selector references a non-persistent model field")
	// ErrInvalidPredicate reports predicate must not be empty.
	ErrInvalidPredicate = errors.New("predicate must not be empty")
	// ErrNilQueryWrapper reports query wrapper must not be nil.
	ErrNilQueryWrapper = errors.New("query wrapper must not be nil")
	// ErrInvalidQueryRange reports query limit and offset must not be negative.
	ErrInvalidQueryRange = errors.New("query limit and offset must not be negative")
	// ErrNilUpdateWrapper reports update wrapper must not be nil.
	ErrNilUpdateWrapper = errors.New("update wrapper must not be nil")
)
