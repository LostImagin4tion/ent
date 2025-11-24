package query

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"entgo.io/ent/dialect"
)

// Selector is a builder for the `SELECT` statement.
type Selector struct {
	Builder
	// ctx stores contextual data typically from
	// generated code such as alternate table schemas.
	ctx       context.Context
	as        string
	selection []selection
	from      []TableView
	joins     []join
	collected [][]*Predicate
	where     *Predicate
	or        bool
	not       bool
	order     []any
	group     []string
	having    *Predicate
	limit     *int
	offset    *int
	distinct  bool
	setOps    []setOp
	prefix    Queries
	lock      *LockOptions
}

// join table option.
type join struct {
	on    *Predicate
	kind  string
	table TableView
}

// clone a joiner.
func (j join) clone() join {
	if sel, ok := j.table.(*Selector); ok {
		j.table = sel.Clone()
	}
	j.on = j.on.clone()
	return j
}

// New returns a new Selector with the same dialect and context.
func (s *Selector) New() *Selector {
	c := Dialect(s.dialect).Select()
	if s.ctx != nil {
		c = c.WithContext(s.ctx)
	}
	return c
}

// WithContext sets the context into the *Selector.
func (s *Selector) WithContext(ctx context.Context) *Selector {
	if ctx == nil {
		panic("nil context")
	}
	s.ctx = ctx
	return s
}

// Context returns the Selector context or Background
// if nil.
func (s *Selector) Context() context.Context {
	if s.ctx != nil {
		return s.ctx
	}
	return context.Background()
}

// Select returns a new selector for the `SELECT` statement.
//
//	t1 := Table("users").As("u")
//	t2 := Select().From(Table("groups")).Where(EQ("user_id", 10)).As("g")
//	return Select(t1.C("id"), t2.C("name")).
//			From(t1).
//			Join(t2).
//			On(t1.C("id"), t2.C("user_id"))
func Select(columns ...string) *Selector {
	return (&Selector{}).Select(columns...)
}

// SelectExpr is like Select, but supports passing arbitrary
// expressions for SELECT clause.
func SelectExpr(exprs ...Querier) *Selector {
	return (&Selector{}).SelectExpr(exprs...)
}

// selection represents a column or an expression selection.
type selection struct {
	x  Querier
	c  string
	as string
}

// Select changes the columns selection of the SELECT statement.
// Empty selection means all columns *.
func (s *Selector) Select(columns ...string) *Selector {
	s.selection = make([]selection, len(columns))
	for i := range columns {
		s.selection[i] = selection{c: columns[i]}
	}
	return s
}

// SelectDistinct selects distinct columns.
func (s *Selector) SelectDistinct(columns ...string) *Selector {
	return s.Select(columns...).Distinct()
}

// AppendSelect appends additional columns to the SELECT statement.
func (s *Selector) AppendSelect(columns ...string) *Selector {
	for i := range columns {
		s.selection = append(s.selection, selection{c: columns[i]})
	}
	return s
}

// AppendSelectAs appends additional column to the SELECT statement with the given alias.
func (s *Selector) AppendSelectAs(column, as string) *Selector {
	s.selection = append(s.selection, selection{c: column, as: as})
	return s
}

// SelectExpr changes the columns selection of the SELECT statement
// with custom list of expressions.
func (s *Selector) SelectExpr(exprs ...Querier) *Selector {
	s.selection = make([]selection, len(exprs))
	for i := range exprs {
		s.selection[i] = selection{x: exprs[i]}
	}
	return s
}

// AppendSelectExpr appends additional expressions to the SELECT statement.
func (s *Selector) AppendSelectExpr(exprs ...Querier) *Selector {
	for i := range exprs {
		s.selection = append(s.selection, selection{x: exprs[i]})
	}
	return s
}

// AppendSelectExprAs appends additional expressions to the SELECT statement with the given name.
func (s *Selector) AppendSelectExprAs(expr Querier, as string) *Selector {
	x := expr
	if _, ok := expr.(*raw); !ok {
		x = ExprFunc(func(b *Builder) {
			b.S("(").Join(expr).S(")")
		})
	}
	s.selection = append(s.selection, selection{
		x:  x,
		as: as,
	})
	return s
}

// FindSelection returns all occurrences in the selection that match the given column name.
// For example, for column "a" the following match: a, "a", "t"."a", "t"."b" AS "a".
func (s *Selector) FindSelection(name string) (matches []string) {
	matchC := func(qualified string) bool {
		switch ident, pg := s.isIdent(qualified), s.postgres(); {
		case !ident:
			if i := strings.IndexRune(qualified, '.'); i > 0 {
				return qualified[i+1:] == name
			}
		case ident && pg:
			if i := strings.Index(qualified, `"."`); i > 0 {
				return s.unquote(qualified[i+2:]) == name
			}
		case ident:
			if i := strings.Index(qualified, "`.`"); i > 0 {
				return s.unquote(qualified[i+2:]) == name
			}
		}
		return false
	}
	for _, c := range s.selection {
		switch {
		// Match aliases.
		case c.as != "":
			if ident := s.isIdent(c.as); !ident && c.as == name || ident && s.unquote(c.as) == name {
				matches = append(matches, c.as)
			}
		// Match qualified columns.
		case c.c != "" && s.isQualified(c.c) && matchC(c.c):
			matches = append(matches, c.c)
		// Match unqualified columns.
		case c.c != "" && (c.c == name || s.isIdent(c.c) && s.unquote(c.c) == name):
			matches = append(matches, c.c)
		}
	}
	return matches
}

// SelectedColumns returns the selected columns in the Selector.
func (s *Selector) SelectedColumns() []string {
	columns := make([]string, 0, len(s.selection))
	for i := range s.selection {
		if c := s.selection[i].c; c != "" {
			columns = append(columns, c)
		}
	}
	return columns
}

// UnqualifiedColumns returns an unqualified version of the
// selected columns in the Selector. e.g. "t1"."c" => "c".
func (s *Selector) UnqualifiedColumns() []string {
	columns := make([]string, 0, len(s.selection))
	for i := range s.selection {
		c := s.selection[i].c
		if c == "" {
			continue
		}
		if s.isIdent(c) {
			parts := strings.FieldsFunc(c, func(r rune) bool {
				return r == '`' || r == '"'
			})
			if n := len(parts); n > 0 && parts[n-1] != "" {
				c = parts[n-1]
			}
		}
		columns = append(columns, c)
	}
	return columns
}

// From sets the source of `FROM` clause.
func (s *Selector) From(t TableView) *Selector {
	s.from = nil
	return s.AppendFrom(t)
}

// AppendFrom appends a new TableView to the `FROM` clause.
func (s *Selector) AppendFrom(t TableView) *Selector {
	s.from = append(s.from, t)
	if st, ok := t.(state); ok {
		st.SetDialect(s.dialect)
	}
	return s
}

// FromExpr sets the expression of `FROM` clause.
func (s *Selector) FromExpr(x Querier) *Selector {
	s.from = nil
	return s.AppendFromExpr(x)
}

// AppendFromExpr appends an expression (Queries) to the `FROM` clause.
func (s *Selector) AppendFromExpr(x Querier) *Selector {
	s.from = append(s.from, &queryView{Querier: x})
	if st, ok := x.(state); ok {
		st.SetDialect(s.dialect)
	}
	return s
}

// Distinct adds the DISTINCT keyword to the `SELECT` statement.
func (s *Selector) Distinct() *Selector {
	s.distinct = true
	return s
}

// SetDistinct sets explicitly if the returned rows are distinct or indistinct.
func (s *Selector) SetDistinct(v bool) *Selector {
	s.distinct = v
	return s
}

// Limit adds the `LIMIT` clause to the `SELECT` statement.
func (s *Selector) Limit(limit int) *Selector {
	s.limit = &limit
	return s
}

// Offset adds the `OFFSET` clause to the `SELECT` statement.
func (s *Selector) Offset(offset int) *Selector {
	s.offset = &offset
	return s
}

// CollectPredicates indicates the appended predicated should be collected
// and not appended to the `WHERE` clause.
func (s *Selector) CollectPredicates() *Selector {
	s.collected = append(s.collected, []*Predicate{})
	return s
}

// CollectedPredicates returns the collected predicates.
func (s *Selector) CollectedPredicates() []*Predicate {
	if len(s.collected) == 0 {
		return nil
	}
	return s.collected[len(s.collected)-1]
}

// UncollectedPredicates stop collecting predicates.
func (s *Selector) UncollectedPredicates() *Selector {
	if len(s.collected) > 0 {
		s.collected = s.collected[:len(s.collected)-1]
	}
	return s
}

// Where sets or appends the given predicate to the statement.
func (s *Selector) Where(p *Predicate) *Selector {
	if len(s.collected) > 0 {
		s.collected[len(s.collected)-1] = append(s.collected[len(s.collected)-1], p)
		return s
	}
	if s.not {
		p = Not(p)
		s.not = false
	}
	switch {
	case s.where == nil:
		s.where = p
	case s.where != nil && s.or:
		s.where = Or(s.where, p)
		s.or = false
	default:
		s.where = And(s.where, p)
	}
	return s
}

// P returns the predicate of a selector.
func (s *Selector) P() *Predicate {
	return s.where
}

// SetP sets explicitly the predicate function for the selector and clear its previous state.
func (s *Selector) SetP(p *Predicate) *Selector {
	s.where = p
	s.or = false
	s.not = false
	return s
}

// FromSelect copies the predicate from a selector.
func (s *Selector) FromSelect(s2 *Selector) *Selector {
	s.where = s2.where
	return s
}

// Not sets the next coming predicate with not.
func (s *Selector) Not() *Selector {
	s.not = true
	return s
}

// Or sets the next coming predicate with OR operator (disjunction).
func (s *Selector) Or() *Selector {
	s.or = true
	return s
}

// Table returns the selected table.
func (s *Selector) Table() *SelectTable {
	if len(s.from) == 0 {
		return nil
	}
	return selectTable(s.from[0])
}

// selectTable returns a *SelectTable from the given TableView.
func selectTable(t TableView) *SelectTable {
	if t == nil {
		return nil
	}
	switch view := t.(type) {
	case *SelectTable:
		return view
	case *Selector:
		if len(view.from) == 0 {
			return nil
		}
		return selectTable(view.from[0])
	case *queryView, *WithBuilder:
		return nil
	default:
		panic(fmt.Sprintf("unexpected TableView %T", t))
	}
}

// TableName returns the name of the selected table or alias of selector.
func (s *Selector) TableName() string {
	switch view := s.from[0].(type) {
	case *SelectTable:
		return view.name
	case *Selector:
		return view.as
	default:
		panic(fmt.Sprintf("unhandled TableView type %T", s.from))
	}
}

// HasJoins reports if the selector has any JOINs.
func (s *Selector) HasJoins() bool {
	return len(s.joins) > 0
}

// JoinedTable returns the first joined table with the given name.
func (s *Selector) JoinedTable(name string) (*SelectTable, bool) {
	for _, j := range s.joins {
		if t := selectTable(j.table); t != nil && t.name == name {
			return t, true
		}
	}
	return nil, false
}

// JoinedTableView returns the first joined TableView with the given name or alias.
func (s *Selector) JoinedTableView(name string) (TableView, bool) {
	for _, j := range s.joins {
		switch t := j.table.(type) {
		case *SelectTable:
			if t.name == name || t.as == name {
				return t, true
			}
		case *Selector:
			if t.as == name {
				return t, true
			}
			for _, t2 := range t.from {
				if t3 := selectTable(t2); t3 != nil && (t3.name == name || t3.as == name) {
					return t3, true
				}
			}
		}
	}
	return nil, false
}

// Join appends a `JOIN` clause to the statement.
func (s *Selector) Join(t TableView) *Selector {
	return s.join("JOIN", t)
}

// LeftJoin appends a `LEFT JOIN` clause to the statement.
func (s *Selector) LeftJoin(t TableView) *Selector {
	return s.join("LEFT JOIN", t)
}

// RightJoin appends a `RIGHT JOIN` clause to the statement.
func (s *Selector) RightJoin(t TableView) *Selector {
	return s.join("RIGHT JOIN", t)
}

// FullJoin appends a `FULL JOIN` clause to the statement.
func (s *Selector) FullJoin(t TableView) *Selector {
	return s.join("FULL JOIN", t)
}

// join adds a join table to the selector with the given kind.
func (s *Selector) join(kind string, t TableView) *Selector {
	s.joins = append(s.joins, join{
		kind:  kind,
		table: t,
	})
	switch view := t.(type) {
	case *SelectTable:
		if view.as == "" {
			view.as = "t" + strconv.Itoa(len(s.joins))
		}
	case *Selector:
		if view.as == "" {
			view.as = "t" + strconv.Itoa(len(s.joins))
		}
	}
	if st, ok := t.(state); ok {
		st.SetDialect(s.dialect)
	}
	return s
}

type (
	// setOp represents a set/compound operation.
	setOp struct {
		Type      setOpType // Set operation type.
		All       bool      // Quantifier was set to ALL (defaults to DISTINCT).
		TableView           // Query or table to operate on.
	}
	// setOpType is a set operation type.
	setOpType string
)

const (
	setOpTypeUnion     setOpType = "UNION"
	setOpTypeExcept    setOpType = "EXCEPT"
	setOpTypeIntersect setOpType = "INTERSECT"
)

// Union appends the UNION (DISTINCT) clause to the query.
func (s *Selector) Union(t TableView) *Selector {
	if s1, ok := t.(*Selector); ok && s == s1 {
		s.AddError(errors.New("self UNION is not supported. Create a clone or a new selector instead"))
		return s
	}
	s.setOps = append(s.setOps, setOp{
		Type:      setOpTypeUnion,
		TableView: t,
	})
	return s
}

// UnionAll appends the UNION ALL clause to the query.
func (s *Selector) UnionAll(t TableView) *Selector {
	s.setOps = append(s.setOps, setOp{
		Type:      setOpTypeUnion,
		All:       true,
		TableView: t,
	})
	return s
}

// UnionDistinct appends the UNION DISTINCT clause to the query.
// Deprecated: use Union instead as by default, duplicate rows
// are eliminated unless ALL is specified.
func (s *Selector) UnionDistinct(t TableView) *Selector {
	return s.Union(t)
}

// Except appends the EXCEPT clause to the query.
func (s *Selector) Except(t TableView) *Selector {
	s.setOps = append(s.setOps, setOp{
		Type:      setOpTypeExcept,
		TableView: t,
	})
	return s
}

// ExceptAll appends the EXCEPT ALL clause to the query.
func (s *Selector) ExceptAll(t TableView) *Selector {
	if s.sqlite() {
		s.AddError(errors.New("EXCEPT ALL is not supported by SQLite"))
	} else {
		s.setOps = append(s.setOps, setOp{
			Type:      setOpTypeExcept,
			All:       true,
			TableView: t,
		})
	}
	return s
}

// Intersect appends the INTERSECT clause to the query.
func (s *Selector) Intersect(t TableView) *Selector {
	s.setOps = append(s.setOps, setOp{
		Type:      setOpTypeIntersect,
		TableView: t,
	})
	return s
}

// IntersectAll appends the INTERSECT ALL clause to the query.
func (s *Selector) IntersectAll(t TableView) *Selector {
	if s.sqlite() {
		s.AddError(errors.New("INTERSECT ALL is not supported by SQLite"))
	} else {
		s.setOps = append(s.setOps, setOp{
			Type:      setOpTypeIntersect,
			All:       true,
			TableView: t,
		})
	}
	return s
}

// Prefix prefixes the query with list of queries.
func (s *Selector) Prefix(queries ...Querier) *Selector {
	s.prefix = append(s.prefix, queries...)
	return s
}

// C returns a formatted string for a selected column from this statement.
func (s *Selector) C(column string) string {
	// Skip formatting qualified columns.
	if s.isQualified(column) {
		return column
	}
	if s.as != "" {
		b := &Builder{dialect: s.dialect}
		b.Ident(s.as)
		b.WriteByte('.')
		b.Ident(column)
		return b.String()
	}
	return s.Table().C(column)
}

// Columns returns a list of formatted strings for a selected columns from this statement.
func (s *Selector) Columns(columns ...string) []string {
	names := make([]string, 0, len(columns))
	for _, c := range columns {
		names = append(names, s.C(c))
	}
	return names
}

// OnP sets or appends the given predicate for the `ON` clause of the statement.
func (s *Selector) OnP(p *Predicate) *Selector {
	if len(s.joins) > 0 {
		join := &s.joins[len(s.joins)-1]
		switch {
		case join.on == nil:
			join.on = p
		default:
			join.on = And(join.on, p)
		}
	}
	return s
}

// On sets the `ON` clause for the `JOIN` operation.
func (s *Selector) On(c1, c2 string) *Selector {
	s.OnP(P(func(builder *Builder) {
		builder.Ident(c1).WriteOp(OpEQ).Ident(c2)
	}))
	return s
}

// As give this selection an alias.
func (s *Selector) As(alias string) *Selector {
	s.as = alias
	return s
}

// Count sets the Select statement to be a `SELECT COUNT(*)`.
func (s *Selector) Count(columns ...string) *Selector {
	column := "*"
	if len(columns) > 0 {
		b := &Builder{}
		b.IdentComma(columns...)
		column = b.String()
	}
	s.Select(Count(column))
	return s
}

// LockAction tells the transaction what to do in case of
// requesting a row that is locked by other transaction.
type LockAction string

const (
	// NoWait means never wait and returns an error.
	NoWait LockAction = "NOWAIT"
	// SkipLocked means never wait and skip.
	SkipLocked LockAction = "SKIP LOCKED"
)

// LockStrength defines the strength of the lock (see the list below).
type LockStrength string

// A list of all locking clauses.
const (
	LockShare       LockStrength = "SHARE"
	LockUpdate      LockStrength = "UPDATE"
	LockNoKeyUpdate LockStrength = "NO KEY UPDATE"
	LockKeyShare    LockStrength = "KEY SHARE"
)

type (
	// LockOptions defines a SELECT statement
	// lock for protecting concurrent updates.
	LockOptions struct {
		// Strength of the lock.
		Strength LockStrength
		// Action of the lock.
		Action LockAction
		// Tables are an option tables.
		Tables []string
		// custom clause for locking.
		clause string
	}
	// LockOption allows configuring the LockOptions using functional options.
	LockOption func(*LockOptions)
)

// WithLockAction sets the Action of the lock.
func WithLockAction(action LockAction) LockOption {
	return func(c *LockOptions) {
		c.Action = action
	}
}

// WithLockTables sets the Tables of the lock.
func WithLockTables(tables ...string) LockOption {
	return func(c *LockOptions) {
		c.Tables = tables
	}
}

// WithLockClause allows providing a custom clause for
// locking the statement. For example, in MySQL <= 8.22:
//
//	Select().
//	From(Table("users")).
//	ForShare(
//		WithLockClause("LOCK IN SHARE MODE"),
//	)
func WithLockClause(clause string) LockOption {
	return func(c *LockOptions) {
		c.clause = clause
	}
}

// For sets the lock configuration for suffixing the `SELECT`
// statement with the `FOR [SHARE | UPDATE] ...` clause.
func (s *Selector) For(l LockStrength, opts ...LockOption) *Selector {
	if s.Dialect() == dialect.SQLite {
		s.AddError(errors.New("sql: SELECT .. FOR UPDATE/SHARE not supported in SQLite"))
	}
	s.lock = &LockOptions{Strength: l}
	for _, opt := range opts {
		opt(s.lock)
	}
	return s
}

// ForShare sets the lock configuration for suffixing the
// `SELECT` statement with the `FOR SHARE` clause.
func (s *Selector) ForShare(opts ...LockOption) *Selector {
	return s.For(LockShare, opts...)
}

// ForUpdate sets the lock configuration for suffixing the
// `SELECT` statement with the `FOR UPDATE` clause.
func (s *Selector) ForUpdate(opts ...LockOption) *Selector {
	return s.For(LockUpdate, opts...)
}

// Clone returns a duplicate of the selector, including all associated steps. It can be
// used to prepare common SELECT statements and use them differently after the clone is made.
func (s *Selector) Clone() *Selector {
	if s == nil {
		return nil
	}
	joins := make([]join, len(s.joins))
	for i := range s.joins {
		joins[i] = s.joins[i].clone()
	}
	return &Selector{
		Builder:   s.Builder.clone(),
		ctx:       s.ctx,
		as:        s.as,
		or:        s.or,
		not:       s.not,
		from:      s.from,
		limit:     s.limit,
		offset:    s.offset,
		distinct:  s.distinct,
		where:     s.where.clone(),
		having:    s.having.clone(),
		joins:     append([]join{}, joins...),
		group:     append([]string{}, s.group...),
		order:     append([]any{}, s.order...),
		selection: append([]selection{}, s.selection...),
	}
}

// Asc adds the ASC suffix for the given column.
func Asc(column string) string {
	b := &Builder{}
	b.Ident(column).WriteString(" ASC")
	return b.String()
}

// Desc adds the DESC suffix for the given column.
func Desc(column string) string {
	b := &Builder{}
	b.Ident(column).WriteString(" DESC")
	return b.String()
}

// DescExpr returns a new expression where the DESC suffix is added.
func DescExpr(x Querier) Querier {
	return ExprFunc(func(b *Builder) {
		b.Join(x)
		b.WriteString(" DESC")
	})
}

// OrderBy appends the `ORDER BY` clause to the `SELECT` statement.
func (s *Selector) OrderBy(columns ...string) *Selector {
	for i := range columns {
		s.order = append(s.order, columns[i])
	}
	return s
}

// OrderColumns returns the ordered columns in the Selector.
// Note, this function skips columns selected with expressions.
func (s *Selector) OrderColumns() []string {
	columns := make([]string, 0, len(s.order))
	for i := range s.order {
		if c, ok := s.order[i].(string); ok {
			columns = append(columns, c)
		}
	}
	return columns
}

// OrderExpr appends the `ORDER BY` clause to the `SELECT`
// statement with custom list of expressions.
func (s *Selector) OrderExpr(exprs ...Querier) *Selector {
	for i := range exprs {
		s.order = append(s.order, exprs[i])
	}
	return s
}

// OrderExprFunc appends the `ORDER BY` expression that evaluates
// the given function.
func (s *Selector) OrderExprFunc(f func(*Builder)) *Selector {
	return s.OrderExpr(
		Dialect(s.Dialect()).Expr(f),
	)
}

// ClearOrder clears the ORDER BY clause to be empty.
func (s *Selector) ClearOrder() *Selector {
	s.order = nil
	return s
}

// GroupBy appends the `GROUP BY` clause to the `SELECT` statement.
func (s *Selector) GroupBy(columns ...string) *Selector {
	s.group = append(s.group, columns...)
	return s
}

// Having appends a predicate for the `HAVING` clause.
func (s *Selector) Having(p *Predicate) *Selector {
	s.having = p
	return s
}

// Query returns query representation of a `SELECT` statement.
func (s *Selector) Query() (string, []any) {
	b := s.Builder.clone()
	s.joinPrefix(&b)
	b.WriteString("SELECT ")
	if s.distinct {
		b.WriteString("DISTINCT ")
	}
	if len(s.selection) > 0 {
		s.joinSelect(&b)
	} else {
		b.WriteString("*")
	}
	if len(s.from) > 0 {
		b.WriteString(" FROM ")
	}
	for i, from := range s.from {
		if i > 0 {
			b.Comma()
		}
		switch t := from.(type) {
		case *SelectTable:
			t.SetDialect(s.dialect)
			b.WriteString(t.ref())
		case *Selector:
			t.SetDialect(s.dialect)
			b.Wrap(func(b *Builder) {
				b.Join(t)
			})
			if t.as != "" {
				b.WriteString(" AS ")
				b.Ident(t.as)
			}
		case *WithBuilder:
			t.SetDialect(s.dialect)
			b.Ident(t.Name())
		case *queryView:
			b.Join(t.Querier)
		}
	}
	for _, join := range s.joins {
		b.WriteString(" " + join.kind + " ")
		switch view := join.table.(type) {
		case *SelectTable:
			view.SetDialect(s.dialect)
			b.WriteString(view.ref())
		case *Selector:
			view.SetDialect(s.dialect)
			b.Wrap(func(b *Builder) {
				b.Join(view)
			})
			b.WriteString(" AS ")
			b.Ident(view.as)
		case *WithBuilder:
			view.SetDialect(s.dialect)
			b.Ident(view.Name())
		}
		if join.on != nil {
			b.WriteString(" ON ")
			b.Join(join.on)
		}
	}
	if s.where != nil {
		b.WriteString(" WHERE ")
		b.Join(s.where)
	}
	if len(s.group) > 0 {
		b.WriteString(" GROUP BY ")
		b.IdentComma(s.group...)
	}
	if s.having != nil {
		b.WriteString(" HAVING ")
		b.Join(s.having)
	}
	if len(s.setOps) > 0 {
		s.joinSetOps(&b)
	}
	joinOrder(s.order, &b)
	if s.limit != nil {
		b.WriteString(" LIMIT ")
		b.WriteString(strconv.Itoa(*s.limit))
	}
	if s.offset != nil {
		b.WriteString(" OFFSET ")
		b.WriteString(strconv.Itoa(*s.offset))
	}
	s.joinLock(&b)
	s.total = b.total
	s.AddError(b.Err())
	return b.String(), b.args
}

func (s *Selector) joinPrefix(b *Builder) {
	if len(s.prefix) > 0 {
		b.join(s.prefix, " ")
		b.Pad()
	}
}

func (s *Selector) joinLock(b *Builder) {
	if s.lock == nil {
		return
	}
	b.Pad()
	if s.lock.clause != "" {
		b.WriteString(s.lock.clause)
		return
	}
	b.WriteString("FOR ").WriteString(string(s.lock.Strength))
	if len(s.lock.Tables) > 0 {
		b.WriteString(" OF ").IdentComma(s.lock.Tables...)
	}
	if s.lock.Action != "" {
		b.Pad().WriteString(string(s.lock.Action))
	}
}

func (s *Selector) joinSetOps(b *Builder) {
	for _, op := range s.setOps {
		b.WriteString(" " + string(op.Type) + " ")
		if op.All {
			b.WriteString("ALL ")
		}
		switch view := op.TableView.(type) {
		case *SelectTable:
			view.SetDialect(s.dialect)
			b.WriteString(view.ref())
		case *Selector:
			view.SetDialect(s.dialect)
			b.Join(view)
			if view.as != "" {
				b.WriteString(" AS ")
				b.Ident(view.as)
			}
		}
	}
}

func joinOrder(order []any, b *Builder) {
	if len(order) == 0 {
		return
	}
	b.WriteString(" ORDER BY ")
	for i := range order {
		if i > 0 {
			b.Comma()
		}
		switch r := order[i].(type) {
		case string:
			b.Ident(r)
		case Querier:
			b.Join(r)
		}
	}
}

func joinReturning(columns []string, b *Builder) {
	if len(columns) == 0 || (!b.postgres() && !b.sqlite()) {
		return
	}
	b.WriteString(" RETURNING ")
	b.IdentComma(columns...)
}

func (s *Selector) joinSelect(b *Builder) {
	for i, sc := range s.selection {
		if i > 0 {
			b.Comma()
		}
		switch {
		case sc.c != "":
			b.Ident(sc.c)
		case sc.x != nil:
			b.Join(sc.x)
		}
		if sc.as != "" {
			b.WriteString(" AS ")
			b.Ident(sc.as)
		}
	}
}

// implement the table view interface.
func (*Selector) view() {}
