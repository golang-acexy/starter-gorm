package gormstarter

// PageWrapper 定义可组合的单表分页查询，分页范围由页码和每页记录数唯一确定。
// PageWrapper 不提供 Limit 和 Offset，避免与页码分页产生冲突。
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

// Number 返回从 1 开始的请求页码。
func (p *PageWrapper[T]) Number() int {
	if p == nil {
		return 0
	}
	return p.number
}

// Size 返回每页请求的记录数。
func (p *PageWrapper[T]) Size() int {
	if p == nil {
		return 0
	}
	return p.size
}

// Where 追加 Predicate 条件；多次调用以及同次传入的多个条件均按 AND 连接。
func (p *PageWrapper[T]) Where(predicates ...Predicate[T]) *PageWrapper[T] {
	p.query.Where(predicates...)
	return p
}

// Select 指定需要查询的字段；未调用时查询模型的全部字段。
func (p *PageWrapper[T]) Select(columns ...Field[T]) *PageWrapper[T] {
	p.query.Select(columns...)
	return p
}

// OrderByAsc 按传入顺序追加升序字段，可多次调用形成多字段排序。
func (p *PageWrapper[T]) OrderByAsc(columns ...Field[T]) *PageWrapper[T] {
	p.query.OrderByAsc(columns...)
	return p
}

// OrderByDesc 按传入顺序追加降序字段，可多次调用形成多字段排序。
func (p *PageWrapper[T]) OrderByDesc(columns ...Field[T]) *PageWrapper[T] {
	p.query.OrderByDesc(columns...)
	return p
}

// Eq 追加等值条件。
func (p *PageWrapper[T]) Eq(field Field[T], value any) *PageWrapper[T] {
	p.query.Eq(field, value)
	return p
}

// Ne 追加不等值条件。
func (p *PageWrapper[T]) Ne(field Field[T], value any) *PageWrapper[T] {
	p.query.Ne(field, value)
	return p
}

// Gt 追加大于条件，适用于支持有序比较的数据库字段。
func (p *PageWrapper[T]) Gt(field Field[T], value any) *PageWrapper[T] {
	p.query.Gt(field, value)
	return p
}

// Ge 追加大于或等于条件，适用于支持有序比较的数据库字段。
func (p *PageWrapper[T]) Ge(field Field[T], value any) *PageWrapper[T] {
	p.query.Ge(field, value)
	return p
}

// Lt 追加小于条件，适用于支持有序比较的数据库字段。
func (p *PageWrapper[T]) Lt(field Field[T], value any) *PageWrapper[T] {
	p.query.Lt(field, value)
	return p
}

// Le 追加小于或等于条件，适用于支持有序比较的数据库字段。
func (p *PageWrapper[T]) Le(field Field[T], value any) *PageWrapper[T] {
	p.query.Le(field, value)
	return p
}

// In 追加集合包含条件；空集合的数据库语义由 GORM 处理。
func (p *PageWrapper[T]) In(field Field[T], values ...any) *PageWrapper[T] {
	p.query.In(field, values...)
	return p
}

// NotIn 追加集合排除条件；空集合的数据库语义由 GORM 处理。
func (p *PageWrapper[T]) NotIn(field Field[T], values ...any) *PageWrapper[T] {
	p.query.NotIn(field, values...)
	return p
}

// Between 追加闭区间 [start, end] 条件。
func (p *PageWrapper[T]) Between(field Field[T], start, end any) *PageWrapper[T] {
	p.query.Between(field, start, end)
	return p
}

// IsNull 追加字段值为 NULL 的条件。
func (p *PageWrapper[T]) IsNull(field Field[T]) *PageWrapper[T] {
	p.query.IsNull(field)
	return p
}

// IsNotNull 追加字段值不为 NULL 的条件。
func (p *PageWrapper[T]) IsNotNull(field Field[T]) *PageWrapper[T] {
	p.query.IsNotNull(field)
	return p
}

// Like 追加 SQL LIKE 条件，value 按原值作为匹配表达式传递，不自动补充或转义通配符。
func (p *PageWrapper[T]) Like(field Field[T], value string) *PageWrapper[T] {
	p.query.Like(field, value)
	return p
}

// NotLike 追加 SQL NOT LIKE 条件，value 按原值作为匹配表达式传递，不自动补充或转义通配符。
func (p *PageWrapper[T]) NotLike(field Field[T], value string) *PageWrapper[T] {
	p.query.NotLike(field, value)
	return p
}

// HasPrefix 追加以指定文本开头的 LIKE 条件，不会自动转义文本中的通配符。
func (p *PageWrapper[T]) HasPrefix(field Field[T], value string) *PageWrapper[T] {
	p.query.HasPrefix(field, value)
	return p
}

// HasSuffix 追加以指定文本结尾的 LIKE 条件，不会自动转义文本中的通配符。
func (p *PageWrapper[T]) HasSuffix(field Field[T], value string) *PageWrapper[T] {
	p.query.HasSuffix(field, value)
	return p
}

// Contains 追加包含指定文本的 LIKE 条件，不会自动转义文本中的通配符。
func (p *PageWrapper[T]) Contains(field Field[T], value string) *PageWrapper[T] {
	p.query.Contains(field, value)
	return p
}
