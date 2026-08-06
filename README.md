# starter-gorm

`starter-gorm` is the relational database starter for the golang-acexy starter/cloud ecosystem. It integrates GORM with the shared starter lifecycle, provides managed MySQL and PostgreSQL connections, routes models to the correct database type, and exposes a generic `BaseMapper` for common SQL operations.

## Ecosystem Role

Use this starter for direct relational persistence and SQL-oriented Mapper APIs. `cloud-database/rds` builds business-oriented Repository APIs on top of these mappers without taking ownership of database lifecycle.

## Requirements

Current module Go version: `1.25.8`.

## Installation

```bash
go get github.com/golang-acexy/starter-gorm
```

## Starter Usage

Register one `GormStarter` with the parent loader. A single starter instance can manage MySQL, PostgreSQL, or both database types.

```go
starter := &gormstarter.GormStarter{
	Config: gormstarter.GormConfig{
		SQLLoggerLevel: logger.InfoLevel,
		MySQL: &gormstarter.MySQLConfig{
			DatabaseConfig: gormstarter.DatabaseConfig{
				Username: "YOUR_USERNAME",
				Password: "YOUR_PASSWORD",
				Host:     "127.0.0.1",
				Port:     3306,
				Database: "app",
			},
			Charset:   "utf8mb4",
			URLParams: "timeout=5s&readTimeout=10s",
		},
	},
}

loader := parent.InitStarterLoader([]parent.Starter{starter})
if err := loader.Start(); err != nil {
	panic(err)
}
```

Use `LazyConfig` when configuration must be resolved immediately before startup. When both `Config` and `LazyConfig` are provided, `LazyConfig` takes precedence.

```go
starter := &gormstarter.GormStarter{
	LazyConfig: func() gormstarter.GormConfig {
		return loadDatabaseConfig()
	},
}
```

Stop the starter through the parent loader so database connections participate in the shared shutdown lifecycle.

```go
_, err := loader.StopAllByRegisteredOrder(10 * time.Second)
if err != nil {
	panic(err)
}
```

## Database Configuration

Common database options are defined by `DatabaseConfig`:

| Field | Description |
| --- | --- |
| `Username` | Database username. |
| `Password` | Database password. |
| `Host` | Database host. |
| `Port` | Database port. |
| `Database` | Database name. |
| `TimeUTC` | Uses UTC for GORM-generated timestamps. |
| `DryRun` | Generates SQL without executing it. |

MySQL-specific options:

| Field | Description |
| --- | --- |
| `Charset` | Connection charset; defaults to `utf8mb4`. |
| `URLParams` | Additional URL query parameters. `parseTime` and `charset` remain starter-managed. |

PostgreSQL-specific options:

| Field | Description |
| --- | --- |
| `Timezone` | Connection timezone; defaults to `UTC`. |
| `SSLMode` | PostgreSQL SSL mode; defaults to `disable`. |

`SQLLoggerLevel` configures the shared SQL logger. `InitFunc` runs after startup and receives a snapshot of the initialized database map.

## MySQL and PostgreSQL Together

Configure both database types on the same starter:

```go
starter := &gormstarter.GormStarter{
	Config: gormstarter.GormConfig{
		MySQL: &gormstarter.MySQLConfig{
			DatabaseConfig: gormstarter.DatabaseConfig{
				Username: "YOUR_MYSQL_USERNAME",
				Password: "YOUR_MYSQL_PASSWORD",
				Host:     "127.0.0.1",
				Port:     3306,
				Database: "app",
			},
		},
		Postgres: &gormstarter.PostgresConfig{
			DatabaseConfig: gormstarter.DatabaseConfig{
				Username: "YOUR_POSTGRES_USERNAME",
				Password: "YOUR_POSTGRES_PASSWORD",
				Host:     "127.0.0.1",
				Port:     5432,
				Database: "analytics",
			},
			Timezone: "UTC",
			SSLMode:  "require",
		},
	},
}
```

Database selection follows these rules:

- With one configured database type, models without `DBType()` use that database.
- With both database types configured, models without `DBType()` use MySQL.
- A model implementing `ModelWithDBType` is always routed to its declared database type.
- Declaring a database type that was not initialized returns `ErrDatabaseNotRegistered`.

```go
type Employee struct {
	ID   uint64
	Name string
}

func (Employee) TableName() string {
	return "employee"
}

func (Employee) DBType() gormstarter.DBType {
	return gormstarter.DBTypePostgres
}
```

## Model and Mapper

A model only needs to provide its table name. Embed `BaseMapper[T]` in a model-specific mapper to obtain the complete mapper API.

```go
type Teacher struct {
	ID        uint64                `gorm:"primaryKey"`
	CreatedAt gormstarter.Timestamp `gorm:"column:create_time"`
	UpdatedAt gormstarter.Timestamp `gorm:"column:update_time"`
	Name      string
	Sex       uint
	Age       uint
}

func (Teacher) TableName() string {
	return "teacher"
}

type TeacherMapper struct {
	gormstarter.BaseMapper[Teacher]
}
```

No constructor or interface assertion is required. Methods embedded from `BaseMapper[Teacher]` are promoted automatically to `TeacherMapper`.

The mapper API is split into focused interfaces and aggregated by `Mapper[T]`:

- `RawMapper` provides access to the current GORM session.
- `QueryMapper[T]` provides select, count, and pagination operations.
- `InsertMapper[T]` provides insert operations.
- `UpdateMapper[T]` provides update operations.
- `DeleteMapper[T]` provides delete operations.
- `Mapper[T]` combines all capabilities above.

## Common Mapper Operations

All mapper methods return the affected row count and an error unless documented otherwise.

```go
mapper := TeacherMapper{}

teacher := Teacher{Name: "Alice", Age: 30}
if _, err := mapper.Insert(&teacher); err != nil {
	return err
}

var selected Teacher
if _, err := mapper.SelectByID(teacher.ID, &selected); err != nil {
	return err
}

if _, err := mapper.UpdateByIDWithMap(
	map[string]any{"name": "Alice Smith", "sex": 0},
	teacher.ID,
); err != nil {
	return err
}

if _, err := mapper.DeleteByID(teacher.ID); err != nil {
	return err
}
```

Method suffixes describe the condition source:

- `ByCond` uses a model value as the condition and ignores its zero-value fields.
- `ByMap` uses `map[string]any` as the condition and can express zero values explicitly.
- `ByWhere` accepts a raw SQL `WHERE` expression and its arguments.
- `ByGorm` accepts callbacks for custom GORM query construction.
- `WithMap` means the map contains values to insert or update rather than query conditions.

Examples:

```go
var teachers []*Teacher

_, err := mapper.SelectByCond(
	Teacher{Sex: 1},
	"id desc",
	&teachers,
)

_, err = mapper.SelectByMap(
	map[string]any{"sex": 0},
	"id desc",
	&teachers,
)

_, err = mapper.SelectByWhere(
	"age >= ?",
	"id desc",
	&teachers,
	18,
)
```

Empty maps are rejected for condition-based update and delete operations to prevent accidental full-table changes.

## Zero-Value Writes

`Insert` includes zero-value fields by default. Pass columns to omit when needed:

```go
_, err := mapper.Insert(&teacher, "created_at")
```

Use `InsertWithoutZeroFields` to include only non-zero fields, or `InsertWithMap` to control values explicitly:

```go
_, err := mapper.InsertWithoutZeroFields(&teacher)

_, err = mapper.InsertWithMap(map[string]any{
	"name": "Alice",
	"sex":  0,
})
```

Updates support the same explicit zero-value control:

```go
_, err := mapper.UpdateByIDWithoutZeroFields(&teacher, "sex")

_, err = mapper.UpdateByCondWithZeroFields(
	&Teacher{Sex: 0},
	Teacher{Name: "Alice"},
	"sex",
)
```

## Pagination

Pagination starts at page number `1`. Both page number and page size must be greater than zero.

```go
var teachers []*Teacher

total, err := mapper.SelectPageByMap(
	map[string]any{"sex": 1},
	gormstarter.PageQuery{
		PageNumber:     1,
		PageSize:       20,
		OrderBySQL:     "id desc",
		SpecifyColumns: []string{"id", "name", "age"},
	},
	&teachers,
)
```

`total` is the number of matching rows before pagination. Invalid pagination returns `ErrInvalidPage`.

Typed conditions are input values, while query destinations remain pointers: use `T` for a condition, `*T` for one result, and `*[]*T` for multiple results.

`PageQuery.TimeRanges` adds left-closed, right-open time filters to both the count and page queries:

```go
start := time.Date(2026, time.August, 1, 0, 0, 0, 0, time.UTC)
end := start.Add(24 * time.Hour)

total, err := mapper.SelectPageByCond(
	Teacher{Sex: 1},
	gormstarter.PageQuery{
		PageNumber: 1,
		PageSize:   20,
		OrderBySQL: "id desc",
		TimeRanges: []gormstarter.TimeRange{
			{Field: "created_at", StartTime: &start, EndTime: &end},
		},
	},
	&teachers,
)
```

`StartTime` or `EndTime` may be nil for an open bound. `Field` is a database column identifier and must come from trusted application configuration rather than unchecked client input.

## Transactions

Transactions use native GORM transaction lifecycle. A model-specific mapper should expose a small wrapper that preserves its concrete type:

```go
func (m TeacherMapper) WithTxMapper(tx *gorm.DB) TeacherMapper {
	return TeacherMapper{
		BaseMapper: m.BaseMapper.GetBaseMapperWithTx(tx),
	}
}
```

Use the transaction mapper for every operation that belongs to the transaction:

```go
db := gormstarter.RawMysqlGormDB()
if db == nil {
	return gormstarter.ErrGormStarterNotStarted
}

tx := db.Begin()
if tx.Error != nil {
	return tx.Error
}
defer tx.Rollback()

txMapper := mapper.WithTxMapper(tx)
if _, err := txMapper.Insert(&Teacher{Name: "Alice"}); err != nil {
	return err
}

return tx.Commit().Error
```

Use `NewBaseMapperWithTx` when a new transaction should be created from the database selected for the mapper model:

```go
txMapper := mapper.NewBaseMapperWithTx()
tx := txMapper.CurrentGormDB()
defer tx.Rollback()
```

The caller owns commit and rollback. Do not reuse a transaction mapper after its transaction has completed.

## Raw GORM Access

Use the typed accessors when database type matters:

```go
mysqlDB := gormstarter.RawMysqlGormDB()
postgresDB := gormstarter.RawPostgresGormDB()
```

`RawGormDB()` returns the default database. Pass a type explicitly to select one:

```go
db := gormstarter.RawGormDB(gormstarter.DBTypePostgres)
```

Raw accessors return `nil` when the requested database is not initialized. Mapper operations return package errors instead of exposing a nil database.

For mapper-scoped raw access:

```go
tableDB := mapper.TableGormDB()
currentDB := mapper.CurrentGormDB()
```

`TableGormDB` applies the model table name. `CurrentGormDB` returns the bound transaction when present, otherwise the model-selected database. The starter lifecycle guarantees these mapper accessors are used only after database startup.

## Common Errors

- `ErrNoDatabaseConfigured`: neither MySQL nor PostgreSQL was configured.
- `ErrGormStarterAlreadyStarted`: the package-level starter instance is already running.
- `ErrGormStarterNotStarted`: no default database is available.
- `ErrDatabaseNotRegistered`: a model requested an uninitialized database type.
- `ErrGormStopTimeout`: database shutdown did not complete before the timeout.
- `ErrInvalidPage`: page number or page size is invalid.
- `ErrNoFieldToSave`: an insert contains no writable fields.
- `ErrNoFieldToUpdate`: an update contains no fields.
- `ErrEmptyCondition`: a protected update or delete has an empty condition.

## Design Notes

- Runtime database instances are package-global and managed by one starter lifecycle.
- Start the GORM starter once and stop it through the parent loader.
- MySQL and PostgreSQL are initialized and stopped together when both are configured.
- The map passed to `InitFunc` and returned by `Start` is a copy; modifying it does not replace the starter registry.
- `BaseMapper` is a value type. Transaction helpers return a new mapper rather than modifying the original mapper.
- Condition structs ignore zero values according to GORM behavior. Use map-based or explicit-column APIs when zero is meaningful.
- `Timestamp` supports SQL scanning, driver values, and JSON timestamp conversion through the shared toolkit.
- The standard GORM starter does not allow parent-managed restart after successful shutdown.
