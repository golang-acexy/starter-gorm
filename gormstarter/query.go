package gormstarter

import "slices"

// NewCondQuery 创建实体条件查询。
func NewCondQuery[T Model](condition T) CondQuery[T] {
	return CondQuery[T]{Condition: condition}
}

// NewMapQuery 创建 Map 条件查询。
func NewMapQuery(condition map[string]any) MapQuery {
	return MapQuery{Condition: condition}
}

// NewWhereQuery 创建原始 SQL 条件查询，args 按占位符顺序传入。
func NewWhereQuery(rawWhereSQL string, args ...any) WhereQuery {
	return WhereQuery{RawWhereSQL: rawWhereSQL, Args: slices.Clone(args)}
}

// NewPageQuery 创建实体条件分页查询。
func NewPageQuery[T Model](condition T, number, size int) PageQuery[T] {
	return PageQuery[T]{Condition: condition, PageOptions: PageOptions{Number: number, Size: size}}
}

// NewMapPageQuery 创建 Map 条件分页查询。
func NewMapPageQuery(condition map[string]any, number, size int) MapPageQuery {
	return MapPageQuery{Condition: condition, PageOptions: PageOptions{Number: number, Size: size}}
}

// NewWherePageQuery 创建原始 SQL 条件分页查询，args 按占位符顺序传入。
func NewWherePageQuery(rawWhereSQL string, number, size int, args ...any) WherePageQuery {
	return WherePageQuery{RawWhereSQL: rawWhereSQL, Args: slices.Clone(args), PageOptions: PageOptions{Number: number, Size: size}}
}

func withQueryOrder(options QueryOptions, orderBySQL string) QueryOptions {
	options.OrderBySQL = orderBySQL
	return options
}

func withQuerySelect(options QueryOptions, columns []string) QueryOptions {
	options.SelectColumns = slices.Clone(columns)
	return options
}

func withQueryTimeRanges(options QueryOptions, ranges []TimeRange) QueryOptions {
	options.TimeRanges = slices.Clone(ranges)
	return options
}

func withQueryLimit(options QueryOptions, limit int) QueryOptions {
	options.Limit = limit
	return options
}

func withPageOrder(options PageOptions, orderBySQL string) PageOptions {
	options.OrderBySQL = orderBySQL
	return options
}

func withPageSelect(options PageOptions, columns []string) PageOptions {
	options.SelectColumns = slices.Clone(columns)
	return options
}

func withPageTimeRanges(options PageOptions, ranges []TimeRange) PageOptions {
	options.TimeRanges = slices.Clone(ranges)
	return options
}

// OrderBy 设置实体条件查询的排序 SQL，再次调用会覆盖原值。
func (q CondQuery[T]) OrderBy(orderBySQL string) CondQuery[T] {
	q.QueryOptions = withQueryOrder(q.QueryOptions, orderBySQL)
	return q
}

// Select 设置实体条件查询的投影字段，再次调用会覆盖原值。
func (q CondQuery[T]) Select(columns ...string) CondQuery[T] {
	q.QueryOptions = withQuerySelect(q.QueryOptions, columns)
	return q
}

// WithTimeRanges 设置实体条件查询的时间范围，再次调用会覆盖原值。
func (q CondQuery[T]) WithTimeRanges(ranges ...TimeRange) CondQuery[T] {
	q.QueryOptions = withQueryTimeRanges(q.QueryOptions, ranges)
	return q
}

// WithLimit 设置实体条件列表查询的最大返回数量。
func (q CondQuery[T]) WithLimit(limit int) CondQuery[T] {
	q.QueryOptions = withQueryLimit(q.QueryOptions, limit)
	return q
}

// OrderBy 设置 Map 条件查询的排序 SQL，再次调用会覆盖原值。
func (q MapQuery) OrderBy(orderBySQL string) MapQuery {
	q.QueryOptions = withQueryOrder(q.QueryOptions, orderBySQL)
	return q
}

// Select 设置 Map 条件查询的投影字段，再次调用会覆盖原值。
func (q MapQuery) Select(columns ...string) MapQuery {
	q.QueryOptions = withQuerySelect(q.QueryOptions, columns)
	return q
}

// WithTimeRanges 设置 Map 条件查询的时间范围，再次调用会覆盖原值。
func (q MapQuery) WithTimeRanges(ranges ...TimeRange) MapQuery {
	q.QueryOptions = withQueryTimeRanges(q.QueryOptions, ranges)
	return q
}

// WithLimit 设置 Map 条件列表查询的最大返回数量。
func (q MapQuery) WithLimit(limit int) MapQuery {
	q.QueryOptions = withQueryLimit(q.QueryOptions, limit)
	return q
}

// OrderBy 设置原始 SQL 条件查询的排序 SQL，再次调用会覆盖原值。
func (q WhereQuery) OrderBy(orderBySQL string) WhereQuery {
	q.QueryOptions = withQueryOrder(q.QueryOptions, orderBySQL)
	return q
}

// Select 设置原始 SQL 条件查询的投影字段，再次调用会覆盖原值。
func (q WhereQuery) Select(columns ...string) WhereQuery {
	q.QueryOptions = withQuerySelect(q.QueryOptions, columns)
	return q
}

// WithTimeRanges 设置原始 SQL 条件查询的时间范围，再次调用会覆盖原值。
func (q WhereQuery) WithTimeRanges(ranges ...TimeRange) WhereQuery {
	q.QueryOptions = withQueryTimeRanges(q.QueryOptions, ranges)
	return q
}

// WithLimit 设置原始 SQL 条件列表查询的最大返回数量。
func (q WhereQuery) WithLimit(limit int) WhereQuery {
	q.QueryOptions = withQueryLimit(q.QueryOptions, limit)
	return q
}

// OrderBy 设置实体条件分页查询的排序 SQL，再次调用会覆盖原值。
func (q PageQuery[T]) OrderBy(orderBySQL string) PageQuery[T] {
	q.PageOptions = withPageOrder(q.PageOptions, orderBySQL)
	return q
}

// Select 设置实体条件分页查询的投影字段，再次调用会覆盖原值。
func (q PageQuery[T]) Select(columns ...string) PageQuery[T] {
	q.PageOptions = withPageSelect(q.PageOptions, columns)
	return q
}

// WithTimeRanges 设置实体条件分页查询的时间范围，再次调用会覆盖原值。
func (q PageQuery[T]) WithTimeRanges(ranges ...TimeRange) PageQuery[T] {
	q.PageOptions = withPageTimeRanges(q.PageOptions, ranges)
	return q
}

// OrderBy 设置 Map 条件分页查询的排序 SQL，再次调用会覆盖原值。
func (q MapPageQuery) OrderBy(orderBySQL string) MapPageQuery {
	q.PageOptions = withPageOrder(q.PageOptions, orderBySQL)
	return q
}

// Select 设置 Map 条件分页查询的投影字段，再次调用会覆盖原值。
func (q MapPageQuery) Select(columns ...string) MapPageQuery {
	q.PageOptions = withPageSelect(q.PageOptions, columns)
	return q
}

// WithTimeRanges 设置 Map 条件分页查询的时间范围，再次调用会覆盖原值。
func (q MapPageQuery) WithTimeRanges(ranges ...TimeRange) MapPageQuery {
	q.PageOptions = withPageTimeRanges(q.PageOptions, ranges)
	return q
}

// OrderBy 设置原始 SQL 条件分页查询的排序 SQL，再次调用会覆盖原值。
func (q WherePageQuery) OrderBy(orderBySQL string) WherePageQuery {
	q.PageOptions = withPageOrder(q.PageOptions, orderBySQL)
	return q
}

// Select 设置原始 SQL 条件分页查询的投影字段，再次调用会覆盖原值。
func (q WherePageQuery) Select(columns ...string) WherePageQuery {
	q.PageOptions = withPageSelect(q.PageOptions, columns)
	return q
}

// WithTimeRanges 设置原始 SQL 条件分页查询的时间范围，再次调用会覆盖原值。
func (q WherePageQuery) WithTimeRanges(ranges ...TimeRange) WherePageQuery {
	q.PageOptions = withPageTimeRanges(q.PageOptions, ranges)
	return q
}
