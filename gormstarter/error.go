package gormstarter

import "errors"

var (
	ErrGormStarterAlreadyStarted = errors.New("gorm starter already started")
	ErrGormStarterNotStarted     = errors.New("gorm starter not started")
	ErrNoDatabaseConfigured      = errors.New("no database configured")
	ErrDatabaseNotRegistered     = errors.New("database type not registered")
	ErrGormStopTimeout           = errors.New("waiting for gorm starter shutdown timeout")
	ErrInvalidPage               = errors.New("page number and page size must be greater than zero")
	ErrNoFieldToSave             = errors.New("no field to save")
	ErrNoFieldToUpdate           = errors.New("no field to update")
	ErrEmptyCondition            = errors.New("condition must not be empty")
	ErrNilEntity                 = errors.New("entity must not be nil")
	ErrEmptyID                   = errors.New("id must not be empty")
	ErrEmptyIDs                  = errors.New("ids must not be empty")
	ErrEmptyWhereSQL             = errors.New("where SQL must not be empty")
	ErrNilColumnSelector         = errors.New("column selector must not be nil")
	ErrInvalidColumnSelector     = errors.New("column selector must return exactly one model field")
	ErrNonPersistentColumn       = errors.New("column selector references a non-persistent model field")
	ErrInvalidPredicate          = errors.New("predicate must not be empty")
	ErrNilQueryWrapper           = errors.New("query wrapper must not be nil")
	ErrInvalidQueryRange         = errors.New("query limit and offset must not be negative")
	ErrNilUpdateWrapper          = errors.New("update wrapper must not be nil")
)
