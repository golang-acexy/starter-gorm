package gormstarter

import (
	"fmt"
	"reflect"
	"slices"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Column binds a model field to its Go value type; database names are resolved against GORM schema at execution.
type Column[T Model, V any] struct {
	fieldPath []int
}

// ColumnRef is a model-bound field reference without a Go value-type parameter.
type ColumnRef[T Model] struct {
	fieldPath []int
}

// Field is a model-bound reference for predicates, projection, and ordering.
type Field[T Model] interface {
	wrapperFieldPath() []int
}

func (c Column[T, V]) wrapperFieldPath() []int {
	return c.fieldPath
}

func (c ColumnRef[T]) wrapperFieldPath() []int {
	return c.fieldPath
}

// NewColumn resolves a unique field address selected on a new struct model.
// A nil selector or nil/external/ambiguous result panics with a package error; persistence is checked at execution.
func NewColumn[T Model, V any](selector func(*T) *V) Column[T, V] {
	return Column[T, V]{fieldPath: mustResolveFieldPath(selector)}
}

func mustResolveFieldPath[T Model, V any](selector func(*T) *V) []int {
	if selector == nil {
		panic(ErrNilColumnSelector)
	}
	model := new(T)
	initializeEmbeddedPointers(reflect.ValueOf(model).Elem())
	selected := selector(model)
	if selected == nil {
		panic(ErrInvalidColumnSelector)
	}
	target := reflect.ValueOf(selected).Pointer()
	targetType := reflect.TypeOf(selected)
	paths := make([][]int, 0, 1)
	collectMatchingFieldPaths(reflect.ValueOf(model).Elem(), nil, target, targetType, &paths)
	if len(paths) != 1 {
		panic(ErrInvalidColumnSelector)
	}
	return paths[0]
}

func initializeEmbeddedPointers(value reflect.Value) {
	for index := 0; index < value.NumField(); index++ {
		field := value.Field(index)
		structField := value.Type().Field(index)
		if !structField.Anonymous {
			continue
		}
		if field.Kind() == reflect.Ptr && field.Type().Elem().Kind() == reflect.Struct && field.CanSet() {
			field.Set(reflect.New(field.Type().Elem()))
			initializeEmbeddedPointers(field.Elem())
		} else if field.Kind() == reflect.Struct {
			initializeEmbeddedPointers(field)
		}
	}
}

func collectMatchingFieldPaths(value reflect.Value, prefix []int, target uintptr, targetType reflect.Type, paths *[][]int) {
	typ := value.Type()
	for index := 0; index < value.NumField(); index++ {
		field := value.Field(index)
		structField := typ.Field(index)
		path := append(slices.Clone(prefix), index)
		if field.CanAddr() && field.Addr().Type() == targetType && field.Addr().Pointer() == target {
			*paths = append(*paths, path)
		}
		if structField.Anonymous && field.Kind() == reflect.Struct {
			collectMatchingFieldPaths(field, path, target, targetType, paths)
		} else if structField.Anonymous && field.Kind() == reflect.Ptr && !field.IsNil() && field.Elem().Kind() == reflect.Struct {
			collectMatchingFieldPaths(field.Elem(), path, target, targetType, paths)
		}
	}
}

func resolveColumnName[T Model](db *gorm.DB, fieldPath []int) (string, error) {
	statement := &gorm.Statement{DB: db}
	if err := statement.Parse(new(T)); err != nil {
		return "", err
	}
	for _, field := range statement.Schema.Fields {
		if slices.Equal(field.StructField.Index, fieldPath) {
			if field.DBName == "" {
				return "", ErrNonPersistentColumn
			}
			return field.DBName, nil
		}
	}
	return "", ErrNonPersistentColumn
}

// Ref returns a copied field path without its value-type parameter; Column itself can also be used as Field.
func (c Column[T, V]) Ref() ColumnRef[T] {
	return ColumnRef[T]{fieldPath: slices.Clone(c.fieldPath)}
}

type predicateBuilder[T Model] func(*gorm.DB) (clause.Expression, error)

// Predicate builds a model-bound SQL expression; its zero value is invalid at execution.
type Predicate[T Model] struct {
	build predicateBuilder[T]
}

func comparisonPredicate[T Model, V any](column Column[T, V], build func(clause.Column) clause.Expression) Predicate[T] {
	return Predicate[T]{build: func(db *gorm.DB) (clause.Expression, error) {
		name, err := resolveColumnName[T](db, column.fieldPath)
		if err != nil {
			return nil, err
		}
		return build(clause.Column{Name: name}), nil
	}}
}

// Eq builds a equality predicate with bound values; field metadata is validated at execution.
func (c Column[T, V]) Eq(value V) Predicate[T] {
	return comparisonPredicate(c, func(column clause.Column) clause.Expression { return clause.Eq{Column: column, Value: value} })
}

// Ne builds a inequality predicate with bound values; field metadata is validated at execution.
func (c Column[T, V]) Ne(value V) Predicate[T] {
	return comparisonPredicate(c, func(column clause.Column) clause.Expression { return clause.Neq{Column: column, Value: value} })
}

// Gt builds a greater-than predicate with bound values; field metadata is validated at execution.
func (c Column[T, V]) Gt(value V) Predicate[T] {
	return comparisonPredicate(c, func(column clause.Column) clause.Expression { return clause.Gt{Column: column, Value: value} })
}

// Ge builds a greater-than-or-equal predicate with bound values; field metadata is validated at execution.
func (c Column[T, V]) Ge(value V) Predicate[T] {
	return comparisonPredicate(c, func(column clause.Column) clause.Expression { return clause.Gte{Column: column, Value: value} })
}

// Lt builds a less-than predicate with bound values; field metadata is validated at execution.
func (c Column[T, V]) Lt(value V) Predicate[T] {
	return comparisonPredicate(c, func(column clause.Column) clause.Expression { return clause.Lt{Column: column, Value: value} })
}

// Le builds a less-than-or-equal predicate with bound values; field metadata is validated at execution.
func (c Column[T, V]) Le(value V) Predicate[T] {
	return comparisonPredicate(c, func(column clause.Column) clause.Expression { return clause.Lte{Column: column, Value: value} })
}

// In builds set membership with bound values; empty sets follow GORM semantics.
func (c Column[T, V]) In(values ...V) Predicate[T] {
	return comparisonPredicate(c, func(column clause.Column) clause.Expression {
		items := make([]any, len(values))
		for index := range values {
			items[index] = values[index]
		}
		return clause.IN{Column: column, Values: items}
	})
}

// NotIn builds negated set membership; empty sets follow GORM semantics.
func (c Column[T, V]) NotIn(values ...V) Predicate[T] {
	predicate := c.In(values...)
	return Not(predicate)
}

// Between builds the closed interval [start, end] with bound values; endpoint ordering is not validated.
func (c Column[T, V]) Between(start, end V) Predicate[T] {
	return comparisonPredicate(c, func(column clause.Column) clause.Expression {
		return clause.Expr{SQL: "? BETWEEN ? AND ?", Vars: []any{column, start, end}}
	})
}

// IsNull builds an SQL NULL check for the field.
func (c Column[T, V]) IsNull() Predicate[T] {
	return comparisonPredicate(c, func(column clause.Column) clause.Expression { return clause.Eq{Column: column, Value: nil} })
}

// IsNotNull builds an SQL NOT NULL check for the field.
func (c Column[T, V]) IsNotNull() Predicate[T] {
	return comparisonPredicate(c, func(column clause.Column) clause.Expression { return clause.Neq{Column: column, Value: nil} })
}

// Like uses the supplied LIKE pattern without adding or escaping wildcards.
func (c Column[T, V]) Like(value string) Predicate[T] {
	return comparisonPredicate(c, func(column clause.Column) clause.Expression { return clause.Like{Column: column, Value: value} })
}

// NotLike uses the supplied NOT LIKE pattern without adding or escaping wildcards.
func (c Column[T, V]) NotLike(value string) Predicate[T] {
	return Not(c.Like(value))
}

// HasPrefix adds a trailing LIKE wildcard; existing wildcard characters are not escaped.
func (c Column[T, V]) HasPrefix(value string) Predicate[T] {
	return c.Like(value + "%")
}

// HasSuffix adds a leading LIKE wildcard; existing wildcard characters are not escaped.
func (c Column[T, V]) HasSuffix(value string) Predicate[T] {
	return c.Like("%" + value)
}

// Contains adds surrounding LIKE wildcards; existing wildcard characters are not escaped.
func (c Column[T, V]) Contains(value string) Predicate[T] {
	return c.Like("%" + value + "%")
}

// And groups predicates using AND; a zero predicate fails at execution and empty groups follow GORM semantics.
func And[T Model](predicates ...Predicate[T]) Predicate[T] {
	return logicalPredicate(clause.And, predicates...)
}

// Or groups predicates using OR; a zero predicate fails at execution and empty groups follow GORM semantics.
func Or[T Model](predicates ...Predicate[T]) Predicate[T] {
	return logicalPredicate(clause.Or, predicates...)
}

func logicalPredicate[T Model](combine func(...clause.Expression) clause.Expression, predicates ...Predicate[T]) Predicate[T] {
	return Predicate[T]{build: func(db *gorm.DB) (clause.Expression, error) {
		expressions := make([]clause.Expression, 0, len(predicates))
		for _, predicate := range predicates {
			if predicate.build == nil {
				return nil, ErrInvalidPredicate
			}
			expression, err := predicate.build(db)
			if err != nil {
				return nil, err
			}
			expressions = append(expressions, expression)
		}
		return combine(expressions...), nil
	}}
}

// Not negates a predicate; a zero predicate fails at execution.
func Not[T Model](predicate Predicate[T]) Predicate[T] {
	return Predicate[T]{build: func(db *gorm.DB) (clause.Expression, error) {
		if predicate.build == nil {
			return nil, ErrInvalidPredicate
		}
		expression, err := predicate.build(db)
		if err != nil {
			return nil, err
		}
		return clause.Not(expression), nil
	}}
}

type wrapperOrder[T Model] struct {
	column Field[T]
	desc   bool
}

// QueryWrapper is a mutable single-table builder; use its constructor and do not mutate it concurrently.
type QueryWrapper[T Model] struct {
	predicates []Predicate[T]
	columns    []Field[T]
	orders     []wrapperOrder[T]
	limit      *int
	offset     *int
}

// NewQueryWrapper creates an empty mutable builder; an empty query is permitted for reads.
func NewQueryWrapper[T Model]() *QueryWrapper[T] {
	return &QueryWrapper[T]{}
}

// Where appends predicates combined with AND; empty predicates fail at execution.
func (q *QueryWrapper[T]) Where(predicates ...Predicate[T]) *QueryWrapper[T] {
	q.predicates = append(q.predicates, predicates...)
	return q
}

// Select appends projection fields; no selected fields uses model columns. Persistence is checked at execution.
func (q *QueryWrapper[T]) Select(columns ...Field[T]) *QueryWrapper[T] {
	q.columns = append(q.columns, columns...)
	return q
}

// OrderByAsc appends ascending fields in precedence order; nil/nonpersistent fields fail at execution.
func (q *QueryWrapper[T]) OrderByAsc(columns ...Field[T]) *QueryWrapper[T] {
	for _, column := range columns {
		q.orders = append(q.orders, wrapperOrder[T]{column: column})
	}
	return q
}

// OrderByDesc appends descending fields in precedence order; nil/nonpersistent fields fail at execution.
func (q *QueryWrapper[T]) OrderByDesc(columns ...Field[T]) *QueryWrapper[T] {
	for _, column := range columns {
		q.orders = append(q.orders, wrapperOrder[T]{column: column, desc: true})
	}
	return q
}

func fieldComparisonPredicate[T Model](field Field[T], build func(clause.Column) clause.Expression) Predicate[T] {
	return Predicate[T]{build: func(db *gorm.DB) (clause.Expression, error) {
		if field == nil || len(field.wrapperFieldPath()) == 0 {
			return nil, ErrInvalidColumnSelector
		}
		name, err := resolveColumnName[T](db, field.wrapperFieldPath())
		if err != nil {
			return nil, err
		}
		return build(clause.Column{Name: name}), nil
	}}
}

// Eq builds a equality predicate with bound values; field metadata is validated at execution.
func (q *QueryWrapper[T]) Eq(field Field[T], value any) *QueryWrapper[T] {
	return q.Where(fieldComparisonPredicate[T](field, func(column clause.Column) clause.Expression { return clause.Eq{Column: column, Value: value} }))
}

// Ne builds a inequality predicate with bound values; field metadata is validated at execution.
func (q *QueryWrapper[T]) Ne(field Field[T], value any) *QueryWrapper[T] {
	return q.Where(fieldComparisonPredicate[T](field, func(column clause.Column) clause.Expression { return clause.Neq{Column: column, Value: value} }))
}

// Gt builds a greater-than predicate with bound values; field metadata is validated at execution.
func (q *QueryWrapper[T]) Gt(field Field[T], value any) *QueryWrapper[T] {
	return q.Where(fieldComparisonPredicate[T](field, func(column clause.Column) clause.Expression { return clause.Gt{Column: column, Value: value} }))
}

// Ge builds a greater-than-or-equal predicate with bound values; field metadata is validated at execution.
func (q *QueryWrapper[T]) Ge(field Field[T], value any) *QueryWrapper[T] {
	return q.Where(fieldComparisonPredicate[T](field, func(column clause.Column) clause.Expression { return clause.Gte{Column: column, Value: value} }))
}

// Lt builds a less-than predicate with bound values; field metadata is validated at execution.
func (q *QueryWrapper[T]) Lt(field Field[T], value any) *QueryWrapper[T] {
	return q.Where(fieldComparisonPredicate[T](field, func(column clause.Column) clause.Expression { return clause.Lt{Column: column, Value: value} }))
}

// Le builds a less-than-or-equal predicate with bound values; field metadata is validated at execution.
func (q *QueryWrapper[T]) Le(field Field[T], value any) *QueryWrapper[T] {
	return q.Where(fieldComparisonPredicate[T](field, func(column clause.Column) clause.Expression { return clause.Lte{Column: column, Value: value} }))
}

// In builds set membership with bound values; empty sets follow GORM semantics.
func (q *QueryWrapper[T]) In(field Field[T], values ...any) *QueryWrapper[T] {
	return q.Where(fieldComparisonPredicate[T](field, func(column clause.Column) clause.Expression { return clause.IN{Column: column, Values: values} }))
}

// NotIn builds negated set membership; empty sets follow GORM semantics.
func (q *QueryWrapper[T]) NotIn(field Field[T], values ...any) *QueryWrapper[T] {
	predicate := fieldComparisonPredicate[T](field, func(column clause.Column) clause.Expression { return clause.IN{Column: column, Values: values} })
	return q.Where(Not(predicate))
}

// Between builds the closed interval [start, end] with bound values; endpoint ordering is not validated.
func (q *QueryWrapper[T]) Between(field Field[T], start, end any) *QueryWrapper[T] {
	return q.Where(fieldComparisonPredicate[T](field, func(column clause.Column) clause.Expression {
		return clause.Expr{SQL: "? BETWEEN ? AND ?", Vars: []any{column, start, end}}
	}))
}

// IsNull builds an SQL NULL check for the field.
func (q *QueryWrapper[T]) IsNull(field Field[T]) *QueryWrapper[T] {
	return q.Where(fieldComparisonPredicate[T](field, func(column clause.Column) clause.Expression { return clause.Eq{Column: column, Value: nil} }))
}

// IsNotNull builds an SQL NOT NULL check for the field.
func (q *QueryWrapper[T]) IsNotNull(field Field[T]) *QueryWrapper[T] {
	return q.Where(fieldComparisonPredicate[T](field, func(column clause.Column) clause.Expression { return clause.Neq{Column: column, Value: nil} }))
}

// Like uses the supplied LIKE pattern without adding or escaping wildcards.
func (q *QueryWrapper[T]) Like(field Field[T], value string) *QueryWrapper[T] {
	return q.Where(fieldComparisonPredicate[T](field, func(column clause.Column) clause.Expression { return clause.Like{Column: column, Value: value} }))
}

// NotLike uses the supplied NOT LIKE pattern without adding or escaping wildcards.
func (q *QueryWrapper[T]) NotLike(field Field[T], value string) *QueryWrapper[T] {
	predicate := fieldComparisonPredicate[T](field, func(column clause.Column) clause.Expression { return clause.Like{Column: column, Value: value} })
	return q.Where(Not(predicate))
}

// HasPrefix adds a trailing LIKE wildcard; existing wildcard characters are not escaped.
func (q *QueryWrapper[T]) HasPrefix(field Field[T], value string) *QueryWrapper[T] {
	return q.Like(field, value+"%")
}

// HasSuffix adds a leading LIKE wildcard; existing wildcard characters are not escaped.
func (q *QueryWrapper[T]) HasSuffix(field Field[T], value string) *QueryWrapper[T] {
	return q.Like(field, "%"+value)
}

// Contains adds surrounding LIKE wildcards; existing wildcard characters are not escaped.
func (q *QueryWrapper[T]) Contains(field Field[T], value string) *QueryWrapper[T] {
	return q.Like(field, "%"+value+"%")
}

// Limit sets the maximum rows; zero requests LIMIT 0 and negatives fail when list-range options are applied.
func (q *QueryWrapper[T]) Limit(limit int) *QueryWrapper[T] {
	q.limit = &limit
	return q
}

// Offset sets skipped rows; negatives fail when list-range options are applied.
func (q *QueryWrapper[T]) Offset(offset int) *QueryWrapper[T] {
	q.offset = &offset
	return q
}

func applyWrapper[T Model](db *gorm.DB, query *QueryWrapper[T], includeSelect, includeOrder, includeLimit bool) (*gorm.DB, error) {
	if query == nil {
		return nil, ErrNilQueryWrapper
	}
	for _, predicate := range query.predicates {
		if predicate.build == nil {
			return nil, ErrInvalidPredicate
		}
		expression, err := predicate.build(db)
		if err != nil {
			return nil, err
		}
		db = db.Where(expression)
	}
	if includeSelect && len(query.columns) > 0 {
		columns := make([]clause.Column, 0, len(query.columns))
		for _, column := range query.columns {
			if column == nil {
				return nil, ErrInvalidColumnSelector
			}
			name, err := resolveColumnName[T](db, column.wrapperFieldPath())
			if err != nil {
				return nil, err
			}
			columns = append(columns, clause.Column{Name: name})
		}
		db = db.Clauses(clause.Select{Columns: columns})
	}
	if includeOrder {
		for _, order := range query.orders {
			if order.column == nil {
				return nil, ErrInvalidColumnSelector
			}
			name, err := resolveColumnName[T](db, order.column.wrapperFieldPath())
			if err != nil {
				return nil, err
			}
			db = db.Order(clause.OrderByColumn{Column: clause.Column{Name: name}, Desc: order.desc})
		}
	}
	if includeLimit {
		if query.limit != nil {
			if *query.limit < 0 {
				return nil, fmt.Errorf("%w: limit", ErrInvalidQueryRange)
			}
			db = db.Limit(*query.limit)
		}
		if query.offset != nil {
			if *query.offset < 0 {
				return nil, fmt.Errorf("%w: offset", ErrInvalidQueryRange)
			}
			db = db.Offset(*query.offset)
		}
	}
	return db, nil
}
