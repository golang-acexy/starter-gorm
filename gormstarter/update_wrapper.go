package gormstarter

// UpdateWrapper 定义单表条件更新，复用 QueryWrapper 的条件能力并保存待更新的字段值。
type UpdateWrapper[T Model] struct {
	query       *QueryWrapper[T]
	assignments []updateAssignment[T]
}

type updateAssignment[T Model] struct {
	field Field[T]
	value any
}

func newUpdateWrapper[T Model]() *UpdateWrapper[T] {
	return &UpdateWrapper[T]{query: NewQueryWrapper[T]()}
}

// Set 追加一个字段赋值，零值和 nil 也会被保留；同一字段重复设置时以最后一次为准。
func (w *UpdateWrapper[T]) Set(field Field[T], value any) *UpdateWrapper[T] {
	w.assignments = append(w.assignments, updateAssignment[T]{field: field, value: value})
	return w
}

// Where 追加可复用的 Predicate 条件；多次调用以及同次传入的多个条件均按 AND 连接。
func (w *UpdateWrapper[T]) Where(predicates ...Predicate[T]) *UpdateWrapper[T] {
	w.query.Where(predicates...)
	return w
}

// Eq 追加等值更新条件。
func (w *UpdateWrapper[T]) Eq(field Field[T], value any) *UpdateWrapper[T] {
	w.query.Eq(field, value)
	return w
}

// Ne 追加不等值更新条件。
func (w *UpdateWrapper[T]) Ne(field Field[T], value any) *UpdateWrapper[T] {
	w.query.Ne(field, value)
	return w
}

// Gt 追加大于更新条件，适用于支持有序比较的数据库字段。
func (w *UpdateWrapper[T]) Gt(field Field[T], value any) *UpdateWrapper[T] {
	w.query.Gt(field, value)
	return w
}

// Ge 追加大于或等于更新条件，适用于支持有序比较的数据库字段。
func (w *UpdateWrapper[T]) Ge(field Field[T], value any) *UpdateWrapper[T] {
	w.query.Ge(field, value)
	return w
}

// Lt 追加小于更新条件，适用于支持有序比较的数据库字段。
func (w *UpdateWrapper[T]) Lt(field Field[T], value any) *UpdateWrapper[T] {
	w.query.Lt(field, value)
	return w
}

// Le 追加小于或等于更新条件，适用于支持有序比较的数据库字段。
func (w *UpdateWrapper[T]) Le(field Field[T], value any) *UpdateWrapper[T] {
	w.query.Le(field, value)
	return w
}

// In 追加集合包含更新条件；空集合的数据库语义由 GORM 处理。
func (w *UpdateWrapper[T]) In(field Field[T], values ...any) *UpdateWrapper[T] {
	w.query.In(field, values...)
	return w
}

// NotIn 追加集合排除更新条件；空集合的数据库语义由 GORM 处理。
func (w *UpdateWrapper[T]) NotIn(field Field[T], values ...any) *UpdateWrapper[T] {
	w.query.NotIn(field, values...)
	return w
}

// Between 追加闭区间 [start, end] 更新条件。
func (w *UpdateWrapper[T]) Between(field Field[T], start, end any) *UpdateWrapper[T] {
	w.query.Between(field, start, end)
	return w
}

// IsNull 追加字段值为 NULL 的更新条件。
func (w *UpdateWrapper[T]) IsNull(field Field[T]) *UpdateWrapper[T] {
	w.query.IsNull(field)
	return w
}

// IsNotNull 追加字段值不为 NULL 的更新条件。
func (w *UpdateWrapper[T]) IsNotNull(field Field[T]) *UpdateWrapper[T] {
	w.query.IsNotNull(field)
	return w
}

// Like 追加 SQL LIKE 更新条件，value 按原值作为匹配表达式传递，不自动补充或转义通配符。
func (w *UpdateWrapper[T]) Like(field Field[T], value string) *UpdateWrapper[T] {
	w.query.Like(field, value)
	return w
}

// NotLike 追加 SQL NOT LIKE 更新条件，value 按原值作为匹配表达式传递，不自动补充或转义通配符。
func (w *UpdateWrapper[T]) NotLike(field Field[T], value string) *UpdateWrapper[T] {
	w.query.NotLike(field, value)
	return w
}

// HasPrefix 追加以指定文本开头的 LIKE 更新条件，不会自动转义文本中的通配符。
func (w *UpdateWrapper[T]) HasPrefix(field Field[T], value string) *UpdateWrapper[T] {
	w.query.HasPrefix(field, value)
	return w
}

// HasSuffix 追加以指定文本结尾的 LIKE 更新条件，不会自动转义文本中的通配符。
func (w *UpdateWrapper[T]) HasSuffix(field Field[T], value string) *UpdateWrapper[T] {
	w.query.HasSuffix(field, value)
	return w
}

// Contains 追加包含指定文本的 LIKE 更新条件，不会自动转义文本中的通配符。
func (w *UpdateWrapper[T]) Contains(field Field[T], value string) *UpdateWrapper[T] {
	w.query.Contains(field, value)
	return w
}
