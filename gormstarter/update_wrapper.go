package gormstarter

// UpdateWrapper is a mutable update builder; execution requires predicates and assignments.
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

// Set appends an assignment retaining zero/nil values; the last assignment for a resolved field wins.
func (w *UpdateWrapper[T]) Set(field Field[T], value any) *UpdateWrapper[T] {
	w.assignments = append(w.assignments, updateAssignment[T]{field: field, value: value})
	return w
}

// Where appends predicates combined with AND; empty predicates fail at execution.
func (w *UpdateWrapper[T]) Where(predicates ...Predicate[T]) *UpdateWrapper[T] {
	w.query.Where(predicates...)
	return w
}

// Eq builds a equality predicate with bound values; field metadata is validated at execution.
func (w *UpdateWrapper[T]) Eq(field Field[T], value any) *UpdateWrapper[T] {
	w.query.Eq(field, value)
	return w
}

// Ne builds a inequality predicate with bound values; field metadata is validated at execution.
func (w *UpdateWrapper[T]) Ne(field Field[T], value any) *UpdateWrapper[T] {
	w.query.Ne(field, value)
	return w
}

// Gt builds a greater-than predicate with bound values; field metadata is validated at execution.
func (w *UpdateWrapper[T]) Gt(field Field[T], value any) *UpdateWrapper[T] {
	w.query.Gt(field, value)
	return w
}

// Ge builds a greater-than-or-equal predicate with bound values; field metadata is validated at execution.
func (w *UpdateWrapper[T]) Ge(field Field[T], value any) *UpdateWrapper[T] {
	w.query.Ge(field, value)
	return w
}

// Lt builds a less-than predicate with bound values; field metadata is validated at execution.
func (w *UpdateWrapper[T]) Lt(field Field[T], value any) *UpdateWrapper[T] {
	w.query.Lt(field, value)
	return w
}

// Le builds a less-than-or-equal predicate with bound values; field metadata is validated at execution.
func (w *UpdateWrapper[T]) Le(field Field[T], value any) *UpdateWrapper[T] {
	w.query.Le(field, value)
	return w
}

// In builds set membership with bound values; empty sets follow GORM semantics.
func (w *UpdateWrapper[T]) In(field Field[T], values ...any) *UpdateWrapper[T] {
	w.query.In(field, values...)
	return w
}

// NotIn builds negated set membership; empty sets follow GORM semantics.
func (w *UpdateWrapper[T]) NotIn(field Field[T], values ...any) *UpdateWrapper[T] {
	w.query.NotIn(field, values...)
	return w
}

// Between builds the closed interval [start, end] with bound values; endpoint ordering is not validated.
func (w *UpdateWrapper[T]) Between(field Field[T], start, end any) *UpdateWrapper[T] {
	w.query.Between(field, start, end)
	return w
}

// IsNull builds an SQL NULL check for the field.
func (w *UpdateWrapper[T]) IsNull(field Field[T]) *UpdateWrapper[T] {
	w.query.IsNull(field)
	return w
}

// IsNotNull builds an SQL NOT NULL check for the field.
func (w *UpdateWrapper[T]) IsNotNull(field Field[T]) *UpdateWrapper[T] {
	w.query.IsNotNull(field)
	return w
}

// Like uses the supplied LIKE pattern without adding or escaping wildcards.
func (w *UpdateWrapper[T]) Like(field Field[T], value string) *UpdateWrapper[T] {
	w.query.Like(field, value)
	return w
}

// NotLike uses the supplied NOT LIKE pattern without adding or escaping wildcards.
func (w *UpdateWrapper[T]) NotLike(field Field[T], value string) *UpdateWrapper[T] {
	w.query.NotLike(field, value)
	return w
}

// HasPrefix adds a trailing LIKE wildcard; existing wildcard characters are not escaped.
func (w *UpdateWrapper[T]) HasPrefix(field Field[T], value string) *UpdateWrapper[T] {
	w.query.HasPrefix(field, value)
	return w
}

// HasSuffix adds a leading LIKE wildcard; existing wildcard characters are not escaped.
func (w *UpdateWrapper[T]) HasSuffix(field Field[T], value string) *UpdateWrapper[T] {
	w.query.HasSuffix(field, value)
	return w
}

// Contains adds surrounding LIKE wildcards; existing wildcard characters are not escaped.
func (w *UpdateWrapper[T]) Contains(field Field[T], value string) *UpdateWrapper[T] {
	w.query.Contains(field, value)
	return w
}
