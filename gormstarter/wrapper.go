package gormstarter

import (
	"fmt"
	"reflect"
	"slices"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Column 表示模型中一个具有确定值类型的持久化字段。
// 字段选择器只在构造时执行，用于记录 Go 字段路径；数据库列名在查询执行时由 GORM Schema 解析。
type Column[T Model, V any] struct {
	fieldPath []int
}

// ColumnRef 是不携带字段值类型的列引用，兼容需要抹除值类型的投影和排序场景。
type ColumnRef[T Model] struct {
	fieldPath []int
}

// Field 是属于指定模型的字段引用，可统一用于条件、投影和排序；不同模型的字段不能混用。
type Field[T Model] interface {
	wrapperFieldPath() []int
}

func (c Column[T, V]) wrapperFieldPath() []int {
	return c.fieldPath
}

func (c ColumnRef[T]) wrapperFieldPath() []int {
	return c.fieldPath
}

// NewColumn 使用字段选择器创建列描述；selector 必须直接返回模型中唯一字段的地址。
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

// Ref 将 Column 转换为不携带值类型的列引用；新代码可直接把 Column 传给投影和排序 API。
func (c Column[T, V]) Ref() ColumnRef[T] {
	return ColumnRef[T]{fieldPath: slices.Clone(c.fieldPath)}
}

type predicateBuilder[T Model] func(*gorm.DB) (clause.Expression, error)

// Predicate 表示属于指定模型的可组合条件，主要用于 And、Or、Not 和 Where 等复杂查询。
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

// Eq 创建字段等于指定值的条件。
func (c Column[T, V]) Eq(value V) Predicate[T] {
	return comparisonPredicate(c, func(column clause.Column) clause.Expression { return clause.Eq{Column: column, Value: value} })
}

// Ne 创建字段不等于指定值的条件。
func (c Column[T, V]) Ne(value V) Predicate[T] {
	return comparisonPredicate(c, func(column clause.Column) clause.Expression { return clause.Neq{Column: column, Value: value} })
}

// Gt 创建字段大于指定值的条件，适用于支持有序比较的数据库字段。
func (c Column[T, V]) Gt(value V) Predicate[T] {
	return comparisonPredicate(c, func(column clause.Column) clause.Expression { return clause.Gt{Column: column, Value: value} })
}

// Ge 创建字段大于或等于指定值的条件，适用于支持有序比较的数据库字段。
func (c Column[T, V]) Ge(value V) Predicate[T] {
	return comparisonPredicate(c, func(column clause.Column) clause.Expression { return clause.Gte{Column: column, Value: value} })
}

// Lt 创建字段小于指定值的条件，适用于支持有序比较的数据库字段。
func (c Column[T, V]) Lt(value V) Predicate[T] {
	return comparisonPredicate(c, func(column clause.Column) clause.Expression { return clause.Lt{Column: column, Value: value} })
}

// Le 创建字段小于或等于指定值的条件，适用于支持有序比较的数据库字段。
func (c Column[T, V]) Le(value V) Predicate[T] {
	return comparisonPredicate(c, func(column clause.Column) clause.Expression { return clause.Lte{Column: column, Value: value} })
}

// In 创建字段包含于给定值集合的条件；空集合的数据库语义由 GORM 处理。
func (c Column[T, V]) In(values ...V) Predicate[T] {
	return comparisonPredicate(c, func(column clause.Column) clause.Expression {
		items := make([]any, len(values))
		for index := range values {
			items[index] = values[index]
		}
		return clause.IN{Column: column, Values: items}
	})
}

// NotIn 创建字段不包含于给定值集合的条件；空集合的数据库语义由 GORM 处理。
func (c Column[T, V]) NotIn(values ...V) Predicate[T] {
	predicate := c.In(values...)
	return Not(predicate)
}

// Between 创建字段位于闭区间 [start, end] 的条件。
func (c Column[T, V]) Between(start, end V) Predicate[T] {
	return comparisonPredicate(c, func(column clause.Column) clause.Expression {
		return clause.Expr{SQL: "? BETWEEN ? AND ?", Vars: []any{column, start, end}}
	})
}

// IsNull 创建字段值为 NULL 的条件。
func (c Column[T, V]) IsNull() Predicate[T] {
	return comparisonPredicate(c, func(column clause.Column) clause.Expression { return clause.Eq{Column: column, Value: nil} })
}

// IsNotNull 创建字段值不为 NULL 的条件。
func (c Column[T, V]) IsNotNull() Predicate[T] {
	return comparisonPredicate(c, func(column clause.Column) clause.Expression { return clause.Neq{Column: column, Value: nil} })
}

// Like 创建 SQL LIKE 条件，value 按原值作为匹配表达式传递，不自动补充或转义通配符。
func (c Column[T, V]) Like(value string) Predicate[T] {
	return comparisonPredicate(c, func(column clause.Column) clause.Expression { return clause.Like{Column: column, Value: value} })
}

// NotLike 创建 SQL NOT LIKE 条件，value 按原值作为匹配表达式传递，不自动补充或转义通配符。
func (c Column[T, V]) NotLike(value string) Predicate[T] {
	return Not(c.Like(value))
}

// HasPrefix 创建以指定文本开头的 LIKE 条件，不会自动转义文本中的通配符。
func (c Column[T, V]) HasPrefix(value string) Predicate[T] {
	return c.Like(value + "%")
}

// HasSuffix 创建以指定文本结尾的 LIKE 条件，不会自动转义文本中的通配符。
func (c Column[T, V]) HasSuffix(value string) Predicate[T] {
	return c.Like("%" + value)
}

// Contains 创建包含指定文本的 LIKE 条件，不会自动转义文本中的通配符。
func (c Column[T, V]) Contains(value string) Predicate[T] {
	return c.Like("%" + value + "%")
}

// And 将多个 Predicate 组合为一个 AND 条件，适用于显式分组的复杂查询。
func And[T Model](predicates ...Predicate[T]) Predicate[T] {
	return logicalPredicate(clause.And, predicates...)
}

// Or 将多个 Predicate 组合为一个 OR 条件，适用于显式分组的复杂查询。
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

// Not 对一个 Predicate 取反，空 Predicate 会在执行时返回 ErrInvalidPredicate。
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

// QueryWrapper 定义可组合的单表查询，承载条件、投影、排序及查询范围。
type QueryWrapper[T Model] struct {
	predicates []Predicate[T]
	columns    []Field[T]
	orders     []wrapperOrder[T]
	limit      *int
	offset     *int
}

// NewQueryWrapper 创建空查询 Wrapper；通常优先使用 mapper.Wrapper() 自动推导模型类型。
func NewQueryWrapper[T Model]() *QueryWrapper[T] {
	return &QueryWrapper[T]{}
}

// Where 追加 Predicate 条件；多次调用以及同次传入的多个条件均按 AND 连接。
func (q *QueryWrapper[T]) Where(predicates ...Predicate[T]) *QueryWrapper[T] {
	q.predicates = append(q.predicates, predicates...)
	return q
}

// Select 指定需要查询的字段；未调用时查询模型的全部字段。
func (q *QueryWrapper[T]) Select(columns ...Field[T]) *QueryWrapper[T] {
	q.columns = append(q.columns, columns...)
	return q
}

// OrderByAsc 按传入顺序追加升序字段，可多次调用形成多字段排序。
func (q *QueryWrapper[T]) OrderByAsc(columns ...Field[T]) *QueryWrapper[T] {
	for _, column := range columns {
		q.orders = append(q.orders, wrapperOrder[T]{column: column})
	}
	return q
}

// OrderByDesc 按传入顺序追加降序字段，可多次调用形成多字段排序。
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

// Eq 追加等值条件。
func (q *QueryWrapper[T]) Eq(field Field[T], value any) *QueryWrapper[T] {
	return q.Where(fieldComparisonPredicate[T](field, func(column clause.Column) clause.Expression { return clause.Eq{Column: column, Value: value} }))
}

// Ne 追加不等值条件。
func (q *QueryWrapper[T]) Ne(field Field[T], value any) *QueryWrapper[T] {
	return q.Where(fieldComparisonPredicate[T](field, func(column clause.Column) clause.Expression { return clause.Neq{Column: column, Value: value} }))
}

// Gt 追加大于条件，适用于支持有序比较的数据库字段。
func (q *QueryWrapper[T]) Gt(field Field[T], value any) *QueryWrapper[T] {
	return q.Where(fieldComparisonPredicate[T](field, func(column clause.Column) clause.Expression { return clause.Gt{Column: column, Value: value} }))
}

// Ge 追加大于或等于条件，适用于支持有序比较的数据库字段。
func (q *QueryWrapper[T]) Ge(field Field[T], value any) *QueryWrapper[T] {
	return q.Where(fieldComparisonPredicate[T](field, func(column clause.Column) clause.Expression { return clause.Gte{Column: column, Value: value} }))
}

// Lt 追加小于条件，适用于支持有序比较的数据库字段。
func (q *QueryWrapper[T]) Lt(field Field[T], value any) *QueryWrapper[T] {
	return q.Where(fieldComparisonPredicate[T](field, func(column clause.Column) clause.Expression { return clause.Lt{Column: column, Value: value} }))
}

// Le 追加小于或等于条件，适用于支持有序比较的数据库字段。
func (q *QueryWrapper[T]) Le(field Field[T], value any) *QueryWrapper[T] {
	return q.Where(fieldComparisonPredicate[T](field, func(column clause.Column) clause.Expression { return clause.Lte{Column: column, Value: value} }))
}

// In 追加集合包含条件；空集合的数据库语义由 GORM 处理。
func (q *QueryWrapper[T]) In(field Field[T], values ...any) *QueryWrapper[T] {
	return q.Where(fieldComparisonPredicate[T](field, func(column clause.Column) clause.Expression { return clause.IN{Column: column, Values: values} }))
}

// NotIn 追加集合排除条件；空集合的数据库语义由 GORM 处理。
func (q *QueryWrapper[T]) NotIn(field Field[T], values ...any) *QueryWrapper[T] {
	predicate := fieldComparisonPredicate[T](field, func(column clause.Column) clause.Expression { return clause.IN{Column: column, Values: values} })
	return q.Where(Not(predicate))
}

// Between 追加闭区间 [start, end] 条件。
func (q *QueryWrapper[T]) Between(field Field[T], start, end any) *QueryWrapper[T] {
	return q.Where(fieldComparisonPredicate[T](field, func(column clause.Column) clause.Expression {
		return clause.Expr{SQL: "? BETWEEN ? AND ?", Vars: []any{column, start, end}}
	}))
}

// IsNull 追加字段值为 NULL 的条件。
func (q *QueryWrapper[T]) IsNull(field Field[T]) *QueryWrapper[T] {
	return q.Where(fieldComparisonPredicate[T](field, func(column clause.Column) clause.Expression { return clause.Eq{Column: column, Value: nil} }))
}

// IsNotNull 追加字段值不为 NULL 的条件。
func (q *QueryWrapper[T]) IsNotNull(field Field[T]) *QueryWrapper[T] {
	return q.Where(fieldComparisonPredicate[T](field, func(column clause.Column) clause.Expression { return clause.Neq{Column: column, Value: nil} }))
}

// Like 追加 SQL LIKE 条件，value 按原值作为匹配表达式传递，不自动补充或转义通配符。
func (q *QueryWrapper[T]) Like(field Field[T], value string) *QueryWrapper[T] {
	return q.Where(fieldComparisonPredicate[T](field, func(column clause.Column) clause.Expression { return clause.Like{Column: column, Value: value} }))
}

// NotLike 追加 SQL NOT LIKE 条件，value 按原值作为匹配表达式传递，不自动补充或转义通配符。
func (q *QueryWrapper[T]) NotLike(field Field[T], value string) *QueryWrapper[T] {
	predicate := fieldComparisonPredicate[T](field, func(column clause.Column) clause.Expression { return clause.Like{Column: column, Value: value} })
	return q.Where(Not(predicate))
}

// HasPrefix 追加以指定文本开头的 LIKE 条件，不会自动转义文本中的通配符。
func (q *QueryWrapper[T]) HasPrefix(field Field[T], value string) *QueryWrapper[T] {
	return q.Like(field, value+"%")
}

// HasSuffix 追加以指定文本结尾的 LIKE 条件，不会自动转义文本中的通配符。
func (q *QueryWrapper[T]) HasSuffix(field Field[T], value string) *QueryWrapper[T] {
	return q.Like(field, "%"+value)
}

// Contains 追加包含指定文本的 LIKE 条件，不会自动转义文本中的通配符。
func (q *QueryWrapper[T]) Contains(field Field[T], value string) *QueryWrapper[T] {
	return q.Like(field, "%"+value+"%")
}

// Limit 限制最多返回的记录数；负数会在查询执行时返回 ErrInvalidQueryRange。
func (q *QueryWrapper[T]) Limit(limit int) *QueryWrapper[T] {
	q.limit = &limit
	return q
}

// Offset 设置跳过的记录数；负数会在查询执行时返回 ErrInvalidQueryRange。
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
