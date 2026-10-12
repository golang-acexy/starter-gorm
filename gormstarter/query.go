package gormstarter

import "slices"

// NewCondQuery stores struct predicates without validation; reference-bearing fields remain borrowed.
func NewCondQuery[T Model](condition T) CondQuery[T] {
	return CondQuery[T]{Condition: condition}
}

// NewMapQuery borrows the predicate map; keep it stable during execution.
func NewMapQuery(condition map[string]any) MapQuery {
	return MapQuery{Condition: condition}
}

// NewWhereQuery stores trusted SQL and copies the args slice; referenced argument values remain borrowed.
func NewWhereQuery(rawWhereSQL string, args ...any) WhereQuery {
	return WhereQuery{RawWhereSQL: rawWhereSQL, Args: slices.Clone(args)}
}

// NewPageQuery stores struct predicates and page inputs; validation is deferred until execution.
func NewPageQuery[T Model](condition T, number, size int) PageQuery[T] {
	return PageQuery[T]{Condition: condition, PageOptions: PageOptions{Number: number, Size: size}}
}

// NewMapPageQuery borrows map predicates and stores page inputs without validating them.
func NewMapPageQuery(condition map[string]any, number, size int) MapPageQuery {
	return MapPageQuery{Condition: condition, PageOptions: PageOptions{Number: number, Size: size}}
}

// NewWherePageQuery stores trusted SQL/page inputs and copies args without deep-copying argument values.
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

// OrderBy returns a query copy replacing the trusted ordering SQL; do not pass unvalidated external expressions.
func (q CondQuery[T]) OrderBy(orderBySQL string) CondQuery[T] {
	q.QueryOptions = withQueryOrder(q.QueryOptions, orderBySQL)
	return q
}

// Select returns a query copy replacing projection fields with a copied slice; referenced SQL expressions must be trusted.
func (q CondQuery[T]) Select(columns ...string) CondQuery[T] {
	q.QueryOptions = withQuerySelect(q.QueryOptions, columns)
	return q
}

// WithTimeRanges returns a query copy replacing copied range values; endpoint pointers remain borrowed.
func (q CondQuery[T]) WithTimeRanges(ranges ...TimeRange) CondQuery[T] {
	q.QueryOptions = withQueryTimeRanges(q.QueryOptions, ranges)
	return q
}

// WithLimit returns a query copy with a list-only limit; zero is unrestricted and negatives fail at execution.
func (q CondQuery[T]) WithLimit(limit int) CondQuery[T] {
	q.QueryOptions = withQueryLimit(q.QueryOptions, limit)
	return q
}

// OrderBy returns a query copy replacing the trusted ordering SQL; do not pass unvalidated external expressions.
func (q MapQuery) OrderBy(orderBySQL string) MapQuery {
	q.QueryOptions = withQueryOrder(q.QueryOptions, orderBySQL)
	return q
}

// Select returns a query copy replacing projection fields with a copied slice; referenced SQL expressions must be trusted.
func (q MapQuery) Select(columns ...string) MapQuery {
	q.QueryOptions = withQuerySelect(q.QueryOptions, columns)
	return q
}

// WithTimeRanges returns a query copy replacing copied range values; endpoint pointers remain borrowed.
func (q MapQuery) WithTimeRanges(ranges ...TimeRange) MapQuery {
	q.QueryOptions = withQueryTimeRanges(q.QueryOptions, ranges)
	return q
}

// WithLimit returns a query copy with a list-only limit; zero is unrestricted and negatives fail at execution.
func (q MapQuery) WithLimit(limit int) MapQuery {
	q.QueryOptions = withQueryLimit(q.QueryOptions, limit)
	return q
}

// OrderBy returns a query copy replacing the trusted ordering SQL; do not pass unvalidated external expressions.
func (q WhereQuery) OrderBy(orderBySQL string) WhereQuery {
	q.QueryOptions = withQueryOrder(q.QueryOptions, orderBySQL)
	return q
}

// Select returns a query copy replacing projection fields with a copied slice; referenced SQL expressions must be trusted.
func (q WhereQuery) Select(columns ...string) WhereQuery {
	q.QueryOptions = withQuerySelect(q.QueryOptions, columns)
	return q
}

// WithTimeRanges returns a query copy replacing copied range values; endpoint pointers remain borrowed.
func (q WhereQuery) WithTimeRanges(ranges ...TimeRange) WhereQuery {
	q.QueryOptions = withQueryTimeRanges(q.QueryOptions, ranges)
	return q
}

// WithLimit returns a query copy with a list-only limit; zero is unrestricted and negatives fail at execution.
func (q WhereQuery) WithLimit(limit int) WhereQuery {
	q.QueryOptions = withQueryLimit(q.QueryOptions, limit)
	return q
}

// OrderBy returns a query copy replacing the trusted ordering SQL; do not pass unvalidated external expressions.
func (q PageQuery[T]) OrderBy(orderBySQL string) PageQuery[T] {
	q.PageOptions = withPageOrder(q.PageOptions, orderBySQL)
	return q
}

// Select returns a query copy replacing projection fields with a copied slice; referenced SQL expressions must be trusted.
func (q PageQuery[T]) Select(columns ...string) PageQuery[T] {
	q.PageOptions = withPageSelect(q.PageOptions, columns)
	return q
}

// WithTimeRanges returns a query copy replacing copied range values; endpoint pointers remain borrowed.
func (q PageQuery[T]) WithTimeRanges(ranges ...TimeRange) PageQuery[T] {
	q.PageOptions = withPageTimeRanges(q.PageOptions, ranges)
	return q
}

// OrderBy returns a query copy replacing the trusted ordering SQL; do not pass unvalidated external expressions.
func (q MapPageQuery) OrderBy(orderBySQL string) MapPageQuery {
	q.PageOptions = withPageOrder(q.PageOptions, orderBySQL)
	return q
}

// Select returns a query copy replacing projection fields with a copied slice; referenced SQL expressions must be trusted.
func (q MapPageQuery) Select(columns ...string) MapPageQuery {
	q.PageOptions = withPageSelect(q.PageOptions, columns)
	return q
}

// WithTimeRanges returns a query copy replacing copied range values; endpoint pointers remain borrowed.
func (q MapPageQuery) WithTimeRanges(ranges ...TimeRange) MapPageQuery {
	q.PageOptions = withPageTimeRanges(q.PageOptions, ranges)
	return q
}

// OrderBy returns a query copy replacing the trusted ordering SQL; do not pass unvalidated external expressions.
func (q WherePageQuery) OrderBy(orderBySQL string) WherePageQuery {
	q.PageOptions = withPageOrder(q.PageOptions, orderBySQL)
	return q
}

// Select returns a query copy replacing projection fields with a copied slice; referenced SQL expressions must be trusted.
func (q WherePageQuery) Select(columns ...string) WherePageQuery {
	q.PageOptions = withPageSelect(q.PageOptions, columns)
	return q
}

// WithTimeRanges returns a query copy replacing copied range values; endpoint pointers remain borrowed.
func (q WherePageQuery) WithTimeRanges(ranges ...TimeRange) WherePageQuery {
	q.PageOptions = withPageTimeRanges(q.PageOptions, ranges)
	return q
}
