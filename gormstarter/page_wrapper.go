package gormstarter

// PageWrapper is a mutable page builder with no Limit/Offset setters; positive page inputs are validated at execution.
type PageWrapper[T Model] struct {
	query  *QueryWrapper[T]
	number int
	size   int
}

func newPageWrapper[T Model](number, size int) *PageWrapper[T] {
	return &PageWrapper[T]{
		query:  NewQueryWrapper[T](),
		number: number,
		size:   size,
	}
}

// Number returns the one-based page input, or zero for a nil page wrapper.
func (p *PageWrapper[T]) Number() int {
	if p == nil {
		return 0
	}
	return p.number
}

// Size returns the page-size input, or zero for a nil page wrapper.
func (p *PageWrapper[T]) Size() int {
	if p == nil {
		return 0
	}
	return p.size
}

// Where appends predicates combined with AND; empty predicates fail at execution.
func (p *PageWrapper[T]) Where(predicates ...Predicate[T]) *PageWrapper[T] {
	p.query.Where(predicates...)
	return p
}

// Select appends projection fields; no selected fields uses model columns. Persistence is checked at execution.
func (p *PageWrapper[T]) Select(columns ...Field[T]) *PageWrapper[T] {
	p.query.Select(columns...)
	return p
}

// OrderByAsc appends ascending fields in precedence order; nil/nonpersistent fields fail at execution.
func (p *PageWrapper[T]) OrderByAsc(columns ...Field[T]) *PageWrapper[T] {
	p.query.OrderByAsc(columns...)
	return p
}

// OrderByDesc appends descending fields in precedence order; nil/nonpersistent fields fail at execution.
func (p *PageWrapper[T]) OrderByDesc(columns ...Field[T]) *PageWrapper[T] {
	p.query.OrderByDesc(columns...)
	return p
}

// Eq builds a equality predicate with bound values; field metadata is validated at execution.
func (p *PageWrapper[T]) Eq(field Field[T], value any) *PageWrapper[T] {
	p.query.Eq(field, value)
	return p
}

// Ne builds a inequality predicate with bound values; field metadata is validated at execution.
func (p *PageWrapper[T]) Ne(field Field[T], value any) *PageWrapper[T] {
	p.query.Ne(field, value)
	return p
}

// Gt builds a greater-than predicate with bound values; field metadata is validated at execution.
func (p *PageWrapper[T]) Gt(field Field[T], value any) *PageWrapper[T] {
	p.query.Gt(field, value)
	return p
}

// Ge builds a greater-than-or-equal predicate with bound values; field metadata is validated at execution.
func (p *PageWrapper[T]) Ge(field Field[T], value any) *PageWrapper[T] {
	p.query.Ge(field, value)
	return p
}

// Lt builds a less-than predicate with bound values; field metadata is validated at execution.
func (p *PageWrapper[T]) Lt(field Field[T], value any) *PageWrapper[T] {
	p.query.Lt(field, value)
	return p
}

// Le builds a less-than-or-equal predicate with bound values; field metadata is validated at execution.
func (p *PageWrapper[T]) Le(field Field[T], value any) *PageWrapper[T] {
	p.query.Le(field, value)
	return p
}

// In builds set membership with bound values; empty sets follow GORM semantics.
func (p *PageWrapper[T]) In(field Field[T], values ...any) *PageWrapper[T] {
	p.query.In(field, values...)
	return p
}

// NotIn builds negated set membership; empty sets follow GORM semantics.
func (p *PageWrapper[T]) NotIn(field Field[T], values ...any) *PageWrapper[T] {
	p.query.NotIn(field, values...)
	return p
}

// Between builds the closed interval [start, end] with bound values; endpoint ordering is not validated.
func (p *PageWrapper[T]) Between(field Field[T], start, end any) *PageWrapper[T] {
	p.query.Between(field, start, end)
	return p
}

// IsNull builds an SQL NULL check for the field.
func (p *PageWrapper[T]) IsNull(field Field[T]) *PageWrapper[T] {
	p.query.IsNull(field)
	return p
}

// IsNotNull builds an SQL NOT NULL check for the field.
func (p *PageWrapper[T]) IsNotNull(field Field[T]) *PageWrapper[T] {
	p.query.IsNotNull(field)
	return p
}

// Like uses the supplied LIKE pattern without adding or escaping wildcards.
func (p *PageWrapper[T]) Like(field Field[T], value string) *PageWrapper[T] {
	p.query.Like(field, value)
	return p
}

// NotLike uses the supplied NOT LIKE pattern without adding or escaping wildcards.
func (p *PageWrapper[T]) NotLike(field Field[T], value string) *PageWrapper[T] {
	p.query.NotLike(field, value)
	return p
}

// HasPrefix adds a trailing LIKE wildcard; existing wildcard characters are not escaped.
func (p *PageWrapper[T]) HasPrefix(field Field[T], value string) *PageWrapper[T] {
	p.query.HasPrefix(field, value)
	return p
}

// HasSuffix adds a leading LIKE wildcard; existing wildcard characters are not escaped.
func (p *PageWrapper[T]) HasSuffix(field Field[T], value string) *PageWrapper[T] {
	p.query.HasSuffix(field, value)
	return p
}

// Contains adds surrounding LIKE wildcards; existing wildcard characters are not escaped.
func (p *PageWrapper[T]) Contains(field Field[T], value string) *PageWrapper[T] {
	p.query.Contains(field, value)
	return p
}
