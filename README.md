# starter-gorm

`starter-gorm` brings GORM into the golang-acexy starter lifecycle. It manages MySQL and PostgreSQL connections and provides generic Mapper APIs plus composable, model-aware query and update wrappers.

## Highlights

- One starter can manage MySQL, PostgreSQL, or both.
- Models can select their database through `DBType()`.
- `BaseMapper[T]` covers common query, insert, update, delete, count, and pagination operations.
- `QueryWrapper` builds complex single-table queries without hard-coded column names.
- `UpdateWrapper` combines reusable conditions with explicit `Set` assignments, including zero and `nil` values.
- Empty-condition updates and deletes are rejected before reaching GORM.
- Transactions preserve the concrete Mapper type.
- GORM naming strategies, `column` tags, and embedded fields are resolved at execution time.

`cloud-database/rds` builds business-oriented Repository APIs on top of this module; database lifecycle remains owned by `starter-gorm`.

## Requirements

- Go `1.26.7`

## Installation

```bash
go get github.com/golang-acexy/starter-gorm
```

## Quick Start

Register `GormStarter` with the parent loader:

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

Use `LazyConfig` to resolve configuration immediately before startup. It takes precedence over `Config` when both are present. Stop connections through the parent loader so they participate in the shared shutdown lifecycle.

## Configuration

Common `DatabaseConfig` fields:

| Field | Purpose |
| --- | --- |
| `Username`, `Password` | Database credentials. |
| `Host`, `Port`, `Database` | Connection target. |
| `TimeUTC` | Use UTC for GORM-generated timestamps. |
| `DryRun` | Generate SQL without executing it. |

MySQL adds `Charset` and `URLParams`; PostgreSQL adds `Timezone` and `SSLMode`. `SQLLoggerLevel` controls SQL logging, and `InitFunc` runs after startup with a snapshot of the initialized database map.

Configure both `MySQL` and `Postgres` in the same `GormConfig` when needed. Database routing follows these rules:

- A model implementing `ModelWithDBType` uses its declared database.
- With one configured database, ordinary models use that database.
- With both configured, ordinary models default to MySQL.
- Selecting an uninitialized database returns `ErrDatabaseNotRegistered`.

```go
func (Employee) DBType() gormstarter.DBType {
	return gormstarter.DBTypePostgres
}
```

## Model and Mapper

A model provides its table name. Embedding `BaseMapper[T]` exposes the complete Mapper API without a constructor:

```go
type Teacher struct {
	ID        uint64                `gorm:"primaryKey"`
	CreatedAt gormstarter.Timestamp `gorm:"column:created_at"`
	Name      string
	Sex       uint
	Age       uint
}

func (Teacher) TableName() string { return "teacher" }

type TeacherMapper struct {
	gormstarter.BaseMapper[Teacher]
}
```

The focused interfaces `QueryMapper`, `InsertMapper`, `UpdateMapper`, and `DeleteMapper` are aggregated by `Mapper[T]`. `RawMapper` provides access to the current GORM session.

Common operations:

```go
mapper := TeacherMapper{}
teacher := Teacher{Name: "Alice", Age: 30}

_, err := mapper.Insert(&teacher)

var selected Teacher
_, err = mapper.SelectByID(teacher.ID, &selected)

_, err = mapper.UpdateByIDWithMap(
	map[string]any{"name": "Alice Smith", "sex": 0},
	teacher.ID,
)

_, err = mapper.DeleteByID(teacher.ID)
```

Condition suffixes describe how a query is built:

- `ByCond`: model condition; zero-value fields follow GORM's condition behavior.
- `ByMap`: map condition; zero values can be expressed explicitly.
- `ByWhere`: raw SQL condition with arguments.
- `ByGorm`: callback for custom GORM construction.
- `WithMap`: map values used for inserts or updates.

Count operations use the same condition query structures; projection and ordering do not participate in the count:

```go
total, err := mapper.CountByCond(
	gormstarter.NewCondQuery(Teacher{Sex: 1}),
)
```

Set `QueryOptions.Limit` to a positive value for top-N list queries. Zero leaves the result unrestricted; negative values return `ErrInvalidQueryRange`. Count, single-result, and pagination queries do not apply this limit.

Query structures also provide immutable chain methods for concise construction:

```go
query := gormstarter.NewCondQuery(Teacher{Sex: 1}).
	OrderBy("created_at desc").
	Select("id", "name").
	WithLimit(20)

page := gormstarter.NewPageQuery(Teacher{Sex: 1}, 1, 20).
	OrderBy("created_at desc").
	Select("id", "name")
```

`Insert` includes zero-value fields. Use `InsertWithoutZeroFields` to omit them, or map and explicit-column APIs when zero values must be controlled precisely.

## Type-safe Wrappers

Columns are declared from Go field selectors. Their database names are resolved from the active GORM schema, so Wrapper code stays aligned with model tags and naming strategies.

```go
type TeacherColumns struct {
	ID        gormstarter.Column[Teacher, uint64]
	Name      gormstarter.Column[Teacher, string]
	Age       gormstarter.Column[Teacher, uint]
	CreatedAt gormstarter.Column[Teacher, gormstarter.Timestamp]
}

var teacherColumns = TeacherColumns{
	ID:        gormstarter.NewColumn(func(v *Teacher) *uint64 { return &v.ID }),
	Name:      gormstarter.NewColumn(func(v *Teacher) *string { return &v.Name }),
	Age:       gormstarter.NewColumn(func(v *Teacher) *uint { return &v.Age }),
	CreatedAt: gormstarter.NewColumn(func(v *Teacher) *gormstarter.Timestamp { return &v.CreatedAt }),
}

func (TeacherMapper) Columns() *TeacherColumns { return &teacherColumns }
```

### QueryWrapper

```go
mapper := TeacherMapper{}
c := mapper.Columns()

query := mapper.Wrapper().
	Ge(c.Age, 18).
	Where(gormstarter.Or(
		c.Name.HasPrefix("Ace"),
		c.Name.Contains("admin"),
	)).
	Select(c.ID, c.Name).
	OrderByDesc(c.CreatedAt).
	Limit(20)

var teachers []*Teacher
_, err := mapper.SelectByWrapper(query, &teachers)
```

Wrapper query APIs include `SelectOneByWrapper`, `SelectByWrapper`, and `CountByWrapper`. Count queries ignore projection, ordering, limit, and offset.

### UpdateWrapper

```go
update := mapper.UpdateWrapper().
	Eq(c.ID, 1).
	Set(c.Name, "updated").
	Set(c.Age, 0)

rows, err := mapper.UpdateByWrapper(update)
```

An update requires at least one condition and one assignment. Zero and `nil` assignments are retained; repeated assignments to the same field use the last value.

## Pagination

Page numbers start at `1`; page number and page size must both be positive.

```go
pageQuery := mapper.PageWrapper(1, 20).
	Eq(c.Age, 18).
	Select(c.ID, c.Name).
	OrderByDesc(c.CreatedAt)

var records []*Teacher
total, err := mapper.SelectPageByWrapper(pageQuery, &records)
```

`PageWrapper` is an independent pagination type. It reuses query conditions, projection, and ordering, but does not expose `Limit` or `Offset`; ordinary `QueryWrapper` values cannot be passed to `SelectPageByWrapper`.

Struct, map, and raw-Where pagination variants are available through `PageQuery`, `MapPageQuery`, and `WherePageQuery`; `SelectPageByGorm` accepts callbacks for fully custom queries. `TimeRanges` adds left-closed, right-open time filters to both count and page queries; either bound may be omitted.

## Transactions

Expose a small model-specific helper to preserve the concrete Mapper type:

```go
func (m TeacherMapper) WithTxMapper(tx *gorm.DB) TeacherMapper {
	return TeacherMapper{BaseMapper: m.BaseMapper.GetBaseMapperWithTx(tx)}
}
```

Use the transaction Mapper for every operation in the transaction. The code that starts the transaction owns commit and rollback, and a completed transaction Mapper must not be reused.

```go
db := gormstarter.RawMysqlGormDB()
tx := db.Begin()
defer tx.Rollback()

txMapper := mapper.WithTxMapper(tx)
if _, err := txMapper.Insert(&Teacher{Name: "Alice"}); err != nil {
	return err
}
return tx.Commit().Error
```

## Raw GORM Access

Use typed accessors when the database type matters:

```go
mysqlDB := gormstarter.RawMysqlGormDB()
postgresDB := gormstarter.RawPostgresGormDB()
db := gormstarter.RawGormDB(gormstarter.DBTypePostgres)
```

Raw accessors return `nil` when the requested database is unavailable. Mapper operations return package errors instead. `TableGormDB` applies the model table, while `CurrentGormDB` returns the bound transaction or the model-selected database.

## Safety and Errors

- Empty-condition updates and deletes return `ErrEmptyCondition`.
- Empty Wrapper assignments return `ErrNoFieldToUpdate`.
- Invalid pagination or Wrapper ranges return `ErrInvalidPage` or `ErrInvalidQueryRange`.
- Nil entities return `ErrNilEntity`; empty ID batches return `ErrEmptyIDs`.
- An unavailable starter or database returns `ErrGormStarterNotStarted` or `ErrDatabaseNotRegistered`.
- Raw Where updates and deletes reject empty expressions with `ErrEmptyWhereSQL`.

GORM's missing-Where protection remains the final safeguard after package-level validation.

## Design Notes

- Runtime database instances are published as one immutable atomic snapshot and managed by one starter lifecycle; raw accessors do not hold lifecycle locks.
- `BaseMapper` is a value type; transaction helpers return a new Mapper.
- MySQL and PostgreSQL start and stop together when both are configured.
- `Timestamp` supports SQL scanning, driver values, and shared JSON timestamp conversion.
- A successfully stopped GORM starter is not restartable through the parent lifecycle.

## Testing

Run deterministic unit tests with `go test ./...`. Tests that connect to MySQL or PostgreSQL use the `integration` build tag and read connection settings from `STARTER_GORM_MYSQL_*` or `STARTER_GORM_POSTGRES_*` environment variables; run them explicitly with `go test -tags=integration ./test/...` after preparing the databases.
