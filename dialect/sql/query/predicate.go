package query

import (
	"database/sql/driver"
	"fmt"
	"strings"

	"entgo.io/ent/dialect"
)

// Predicate is a where predicate.
type Predicate struct {
	Builder
	depth int
	fns   []func(*Builder)
}

// P creates a new predicate.
//
//	P().EQ("name", "a8m").And().EQ("age", 30)
func P(fns ...func(*Builder)) *Predicate {
	return &Predicate{fns: fns}
}

// ExprP creates a new predicate from the given expression.
//
//	ExprP("A = ? AND B > ?", args...)
func ExprP(exr string, args ...any) *Predicate {
	return P(func(b *Builder) {
		b.Join(Expr(exr, args...))
	})
}

// Or combines all given predicates with OR between them.
//
//	Or(EQ("name", "foo"), EQ("name", "bar"))
func Or(preds ...*Predicate) *Predicate {
	p := P()
	return p.Append(func(b *Builder) {
		p.mayWrap(preds, b, "OR")
	})
}

// False appends the FALSE keyword to the predicate.
//
//	Delete().From("users").Where(False())
func False() *Predicate {
	return P().False()
}

// False appends FALSE to the predicate.
func (p *Predicate) False() *Predicate {
	return p.Append(func(b *Builder) {
		b.WriteString("FALSE")
	})
}

// Not wraps the given predicate with the not predicate.
//
//	Not(Or(EQ("name", "foo"), EQ("name", "bar")))
func Not(pred *Predicate) *Predicate {
	return P().Not().Append(func(b *Builder) {
		b.Wrap(func(b *Builder) {
			b.Join(pred)
		})
	})
}

// Not appends NOT to the predicate.
func (p *Predicate) Not() *Predicate {
	return p.Append(func(b *Builder) {
		b.WriteString("NOT ")
	})
}

// ColumnsOp returns a new predicate between 2 columns.
func ColumnsOp(col1, col2 string, op Op) *Predicate {
	return P().ColumnsOp(col1, col2, op)
}

// ColumnsOp appends the given predicate between 2 columns.
func (p *Predicate) ColumnsOp(col1, col2 string, op Op) *Predicate {
	return p.Append(func(b *Builder) {
		b.Ident(col1)
		b.WriteOp(op)
		b.Ident(col2)
	})
}

// And combines all given predicates with AND between them.
func And(preds ...*Predicate) *Predicate {
	p := P()
	return p.Append(func(b *Builder) {
		p.mayWrap(preds, b, "AND")
	})
}

// IsTrue appends a predicate that checks if the column value is truthy.
func IsTrue(col string) *Predicate {
	return P().IsTrue(col)
}

// IsTrue appends a predicate that checks if the column value is truthy.
func (p *Predicate) IsTrue(col string) *Predicate {
	return p.Append(func(b *Builder) {
		b.Ident(col)
	})
}

// IsFalse appends a predicate that checks if the column value is falsey.
func IsFalse(col string) *Predicate {
	return P().IsFalse(col)
}

// IsFalse appends a predicate that checks if the column value is falsey.
func (p *Predicate) IsFalse(col string) *Predicate {
	return p.Append(func(b *Builder) {
		b.WriteString("NOT ").Ident(col)
	})
}

// EQ returns a "=" predicate.
func EQ(col string, value any) *Predicate {
	return P().EQ(col, value)
}

// EQ appends a "=" predicate.
func (p *Predicate) EQ(col string, arg any) *Predicate {
	// A small optimization to avoid passing
	// arguments when it can be avoided.
	switch arg := arg.(type) {
	case bool:
		if arg {
			return IsTrue(col)
		}
		return IsFalse(col)
	default:
		return p.Append(func(b *Builder) {
			b.Ident(col)
			b.WriteOp(OpEQ)
			p.arg(b, arg)
		})
	}
}

// ColumnsEQ appends a "=" predicate between 2 columns.
func ColumnsEQ(col1, col2 string) *Predicate {
	return P().ColumnsEQ(col1, col2)
}

// ColumnsEQ appends a "=" predicate between 2 columns.
func (p *Predicate) ColumnsEQ(col1, col2 string) *Predicate {
	return p.ColumnsOp(col1, col2, OpEQ)
}

// NEQ returns a "<>" predicate.
func NEQ(col string, value any) *Predicate {
	return P().NEQ(col, value)
}

// NEQ appends a "<>" predicate.
func (p *Predicate) NEQ(col string, arg any) *Predicate {
	// A small optimization to avoid passing
	// arguments when it can be avoided.
	switch arg := arg.(type) {
	case bool:
		if arg {
			return IsFalse(col)
		}
		return IsTrue(col)
	default:
		return p.Append(func(b *Builder) {
			b.Ident(col)
			b.WriteOp(OpNEQ)
			p.arg(b, arg)
		})
	}
}

// ColumnsNEQ appends a "<>" predicate between 2 columns.
func ColumnsNEQ(col1, col2 string) *Predicate {
	return P().ColumnsNEQ(col1, col2)
}

// ColumnsNEQ appends a "<>" predicate between 2 columns.
func (p *Predicate) ColumnsNEQ(col1, col2 string) *Predicate {
	return p.ColumnsOp(col1, col2, OpNEQ)
}

// LT returns a "<" predicate.
func LT(col string, value any) *Predicate {
	return P().LT(col, value)
}

// LT appends a "<" predicate.
func (p *Predicate) LT(col string, arg any) *Predicate {
	return p.Append(func(b *Builder) {
		b.Ident(col)
		p.WriteOp(OpLT)
		p.arg(b, arg)
	})
}

// ColumnsLT appends a "<" predicate between 2 columns.
func ColumnsLT(col1, col2 string) *Predicate {
	return P().ColumnsLT(col1, col2)
}

// ColumnsLT appends a "<" predicate between 2 columns.
func (p *Predicate) ColumnsLT(col1, col2 string) *Predicate {
	return p.ColumnsOp(col1, col2, OpLT)
}

// LTE returns a "<=" predicate.
func LTE(col string, value any) *Predicate {
	return P().LTE(col, value)
}

// LTE appends a "<=" predicate.
func (p *Predicate) LTE(col string, arg any) *Predicate {
	return p.Append(func(b *Builder) {
		b.Ident(col)
		p.WriteOp(OpLTE)
		p.arg(b, arg)
	})
}

// ColumnsLTE appends a "<=" predicate between 2 columns.
func ColumnsLTE(col1, col2 string) *Predicate {
	return P().ColumnsLTE(col1, col2)
}

// ColumnsLTE appends a "<=" predicate between 2 columns.
func (p *Predicate) ColumnsLTE(col1, col2 string) *Predicate {
	return p.ColumnsOp(col1, col2, OpLTE)
}

// GT returns a ">" predicate.
func GT(col string, value any) *Predicate {
	return P().GT(col, value)
}

// GT appends a ">" predicate.
func (p *Predicate) GT(col string, arg any) *Predicate {
	return p.Append(func(b *Builder) {
		b.Ident(col)
		p.WriteOp(OpGT)
		p.arg(b, arg)
	})
}

// ColumnsGT appends a ">" predicate between 2 columns.
func ColumnsGT(col1, col2 string) *Predicate {
	return P().ColumnsGT(col1, col2)
}

// ColumnsGT appends a ">" predicate between 2 columns.
func (p *Predicate) ColumnsGT(col1, col2 string) *Predicate {
	return p.ColumnsOp(col1, col2, OpGT)
}

// GTE returns a ">=" predicate.
func GTE(col string, value any) *Predicate {
	return P().GTE(col, value)
}

// GTE appends a ">=" predicate.
func (p *Predicate) GTE(col string, arg any) *Predicate {
	return p.Append(func(b *Builder) {
		b.Ident(col)
		p.WriteOp(OpGTE)
		p.arg(b, arg)
	})
}

// ColumnsGTE appends a ">=" predicate between 2 columns.
func ColumnsGTE(col1, col2 string) *Predicate {
	return P().ColumnsGTE(col1, col2)
}

// ColumnsGTE appends a ">=" predicate between 2 columns.
func (p *Predicate) ColumnsGTE(col1, col2 string) *Predicate {
	return p.ColumnsOp(col1, col2, OpGTE)
}

// NotNull returns the `IS NOT NULL` predicate.
func NotNull(col string) *Predicate {
	return P().NotNull(col)
}

// NotNull appends the `IS NOT NULL` predicate.
func (p *Predicate) NotNull(col string) *Predicate {
	return p.Append(func(b *Builder) {
		b.Ident(col).WriteString(" IS NOT NULL")
	})
}

// IsNull returns the `IS NULL` predicate.
func IsNull(col string) *Predicate {
	return P().IsNull(col)
}

// IsNull appends the `IS NULL` predicate.
func (p *Predicate) IsNull(col string) *Predicate {
	return p.Append(func(b *Builder) {
		b.Ident(col).WriteString(" IS NULL")
	})
}

// In returns the `IN` predicate.
func In(col string, args ...any) *Predicate {
	return P().In(col, args...)
}

// In appends the `IN` predicate.
func (p *Predicate) In(col string, args ...any) *Predicate {
	// If no arguments were provided, append the FALSE constant, since
	// we cannot apply "IN ()". This will make this predicate falsy.
	if len(args) == 0 {
		return p.False()
	}
	return p.Append(func(b *Builder) {
		b.Ident(col).WriteOp(OpIn)
		b.Wrap(func(b *Builder) {
			if s, ok := args[0].(*Selector); ok {
				b.Join(s)
			} else {
				b.Args(args...)
			}
		})
	})
}

// InInts returns the `IN` predicate for ints.
func InInts(col string, args ...int) *Predicate {
	return P().InInts(col, args...)
}

// InValues adds the `IN` predicate for slice of driver.Value.
func InValues(col string, args ...driver.Value) *Predicate {
	return P().InValues(col, args...)
}

// InInts adds the `IN` predicate for ints.
func (p *Predicate) InInts(col string, args ...int) *Predicate {
	iface := make([]any, len(args))
	for i := range args {
		iface[i] = args[i]
	}
	return p.In(col, iface...)
}

// InValues adds the `IN` predicate for slice of driver.Value.
func (p *Predicate) InValues(col string, args ...driver.Value) *Predicate {
	iface := make([]any, len(args))
	for i := range args {
		iface[i] = args[i]
	}
	return p.In(col, iface...)
}

// NotIn returns the `Not IN` predicate.
func NotIn(col string, args ...any) *Predicate {
	return P().NotIn(col, args...)
}

// NotIn appends the `Not IN` predicate.
func (p *Predicate) NotIn(col string, args ...any) *Predicate {
	// If no arguments were provided, append the NOT FALSE constant, since
	// we cannot apply "NOT IN ()". This will make this predicate truthy.
	if len(args) == 0 {
		return Not(p.False())
	}
	return p.Append(func(b *Builder) {
		b.Ident(col).WriteOp(OpNotIn)
		b.Wrap(func(b *Builder) {
			if s, ok := args[0].(*Selector); ok {
				b.Join(s)
			} else {
				b.Args(args...)
			}
		})
	})
}

// Exists returns the `Exists` predicate.
func Exists(query Querier) *Predicate {
	return P().Exists(query)
}

// Exists appends the `EXISTS` predicate with the given query.
func (p *Predicate) Exists(query Querier) *Predicate {
	return p.Append(func(b *Builder) {
		b.WriteString("EXISTS ")
		b.Wrap(func(b *Builder) {
			b.Join(query)
		})
	})
}

// NotExists returns the `NotExists` predicate.
func NotExists(query Querier) *Predicate {
	return P().NotExists(query)
}

// NotExists appends the `NOT EXISTS` predicate with the given query.
func (p *Predicate) NotExists(query Querier) *Predicate {
	return p.Append(func(b *Builder) {
		b.WriteString("NOT EXISTS ")
		b.Wrap(func(b *Builder) {
			b.Join(query)
		})
	})
}

// Like returns the `LIKE` predicate.
func Like(col, pattern string) *Predicate {
	return P().Like(col, pattern)
}

// Like appends the `LIKE` predicate.
func (p *Predicate) Like(col, pattern string) *Predicate {
	return p.Append(func(b *Builder) {
		b.Ident(col).WriteOp(OpLike)
		b.Arg(pattern)
	})
}

// escape escapes w with the default escape character ('/'),
// to be used by the pattern matching functions below.
// The second return value indicates if w was escaped or not.
func escape(w string) (string, bool) {
	var n int
	for i := range w {
		if c := w[i]; c == '%' || c == '_' || c == '\\' {
			n++
		}
	}
	// No characters to escape.
	if n == 0 {
		return w, false
	}
	var b strings.Builder
	b.Grow(len(w) + n)
	for _, c := range w {
		if c == '%' || c == '_' || c == '\\' {
			b.WriteByte('\\')
		}
		b.WriteRune(c)
	}
	return b.String(), true
}

func (p *Predicate) escapedLike(col, left, right, word string) *Predicate {
	return p.Append(func(b *Builder) {
		w, escaped := escape(word)
		b.Ident(col).WriteOp(OpLike)
		b.Arg(left + w + right)
		if p.dialect == dialect.SQLite && escaped {
			p.WriteString(" ESCAPE ").Arg("\\")
		}
	})
}

// ContainsFold is a helper predicate that applies the LIKE predicate with case-folding.
func (p *Predicate) escapedLikeFold(col, left, substr, right string) *Predicate {
	return p.Append(func(b *Builder) {
		w, escaped := escape(substr)
		switch b.dialect {
		case dialect.MySQL:
			// We assume the CHARACTER SET is configured to utf8mb4,
			// because this how it is defined in dialect/sql/schema.
			b.Ident(col).WriteString(" COLLATE utf8mb4_general_ci LIKE ")
			b.Arg(left + strings.ToLower(w) + right)
		case dialect.Postgres:
			b.Ident(col).WriteString(" ILIKE ")
			b.Arg(left + strings.ToLower(w) + right)
		default: // SQLite.
			var f Func
			f.SetDialect(b.dialect)
			f.Lower(col)
			b.WriteString(f.String()).WriteString(" LIKE ")
			b.Arg(left + strings.ToLower(w) + right)
			if escaped {
				p.WriteString(" ESCAPE ").Arg("\\")
			}
		}
	})
}

// HasPrefix is a helper predicate that checks prefix using the LIKE predicate.
func HasPrefix(col, prefix string) *Predicate {
	return P().HasPrefix(col, prefix)
}

// HasPrefix is a helper predicate that checks prefix using the LIKE predicate.
func (p *Predicate) HasPrefix(col, prefix string) *Predicate {
	return p.escapedLike(col, "", "%", prefix)
}

// HasPrefixFold is a helper predicate that checks prefix using the ILIKE predicate.
func HasPrefixFold(col, prefix string) *Predicate {
	return P().HasPrefixFold(col, prefix)
}

// HasPrefixFold is a helper predicate that checks prefix using the ILIKE predicate.
func (p *Predicate) HasPrefixFold(col, prefix string) *Predicate {
	return p.escapedLikeFold(col, "", prefix, "%")
}

// ColumnsHasPrefix appends a new predicate that checks if the given column begins with the other column (prefix).
func ColumnsHasPrefix(col, prefixC string) *Predicate {
	return P().ColumnsHasPrefix(col, prefixC)
}

// ColumnsHasPrefix appends a new predicate that checks if the given column begins with the other column (prefix).
func (p *Predicate) ColumnsHasPrefix(col, prefixC string) *Predicate {
	return p.Append(func(b *Builder) {
		switch p.dialect {
		case dialect.MySQL:
			b.Ident(col)
			b.WriteOp(OpLike)
			b.S("CONCAT(REPLACE(REPLACE(").Ident(prefixC).S(", '_', '\\_'), '%', '\\%'), '%')")
		case dialect.Postgres, dialect.SQLite:
			b.Ident(col)
			b.WriteOp(OpLike)
			b.S("(REPLACE(REPLACE(").Ident(prefixC).S(", '_', '\\_'), '%', '\\%') || '%')")
			if p.dialect == dialect.SQLite {
				p.WriteString(" ESCAPE ").Arg("\\")
			}
		default:
			b.AddError(fmt.Errorf("ColumnsHasPrefix: unsupported dialect: %q", p.dialect))
		}
	})
}

// HasSuffix is a helper predicate that checks suffix using the LIKE predicate.
func HasSuffix(col, suffix string) *Predicate { return P().HasSuffix(col, suffix) }

// HasSuffix is a helper predicate that checks suffix using the LIKE predicate.
func (p *Predicate) HasSuffix(col, suffix string) *Predicate {
	return p.escapedLike(col, "%", "", suffix)
}

// HasSuffixFold is a helper predicate that checks suffix using the ILIKE predicate.
func HasSuffixFold(col, suffix string) *Predicate { return P().HasSuffixFold(col, suffix) }

// HasSuffixFold is a helper predicate that checks suffix using the ILIKE predicate.
func (p *Predicate) HasSuffixFold(col, suffix string) *Predicate {
	return p.escapedLikeFold(col, "%", suffix, "")
}

// EqualFold is a helper predicate that applies the "=" predicate with case-folding.
func EqualFold(col, sub string) *Predicate { return P().EqualFold(col, sub) }

// EqualFold is a helper predicate that applies the "=" predicate with case-folding.
func (p *Predicate) EqualFold(col, sub string) *Predicate {
	return p.Append(func(b *Builder) {
		f := &Func{}
		f.SetDialect(b.dialect)
		switch b.dialect {
		case dialect.MySQL:
			// We assume the CHARACTER SET is configured to utf8mb4,
			// because this how it is defined in dialect/sql/schema.
			b.Ident(col).WriteString(" COLLATE utf8mb4_general_ci = ")
			b.Arg(strings.ToLower(sub))
		case dialect.Postgres:
			b.Ident(col).WriteString(" ILIKE ")
			w, _ := escape(sub)
			b.Arg(strings.ToLower(w))
		default: // SQLite.
			f.Lower(col)
			b.WriteString(f.String())
			b.WriteOp(OpEQ)
			b.Arg(strings.ToLower(sub))
		}
	})
}

// Contains is a helper predicate that checks substring using the LIKE predicate.
func Contains(col, sub string) *Predicate { return P().Contains(col, sub) }

// Contains is a helper predicate that checks substring using the LIKE predicate.
func (p *Predicate) Contains(col, substr string) *Predicate {
	return p.escapedLike(col, "%", "%", substr)
}

// ContainsFold is a helper predicate that checks substring using the LIKE predicate with case-folding.
func ContainsFold(col, sub string) *Predicate { return P().ContainsFold(col, sub) }

// ContainsFold is a helper predicate that applies the LIKE predicate with case-folding.
func (p *Predicate) ContainsFold(col, substr string) *Predicate {
	return p.escapedLikeFold(col, "%", substr, "%")
}

// CompositeGT returns a composite ">" predicate
func CompositeGT(columns []string, args ...any) *Predicate {
	return P().CompositeGT(columns, args...)
}

// CompositeLT returns a composite "<" predicate
func CompositeLT(columns []string, args ...any) *Predicate {
	return P().CompositeLT(columns, args...)
}

func (p *Predicate) compositeP(operator string, columns []string, args ...any) *Predicate {
	return p.Append(func(b *Builder) {
		b.Wrap(func(nb *Builder) {
			nb.IdentComma(columns...)
		})
		b.WriteString(operator)
		b.WriteString("(")
		b.Args(args...)
		b.WriteString(")")
	})
}

// CompositeGT returns a composite ">" predicate.
func (p *Predicate) CompositeGT(columns []string, args ...any) *Predicate {
	const operator = " > "
	return p.compositeP(operator, columns, args...)
}

// CompositeLT appends a composite "<" predicate.
func (p *Predicate) CompositeLT(columns []string, args ...any) *Predicate {
	const operator = " < "
	return p.compositeP(operator, columns, args...)
}

// Append appends a new function to the predicate callbacks.
// The callback list are executed on call to Query.
func (p *Predicate) Append(f func(*Builder)) *Predicate {
	p.fns = append(p.fns, f)
	return p
}

// Query returns query representation of a predicate.
func (p *Predicate) Query() (string, []any) {
	if p.Len() > 0 || len(p.args) > 0 {
		p.Reset()
		p.args = nil
	}
	for _, f := range p.fns {
		f(&p.Builder)
	}
	return p.String(), p.args
}

// arg calls Builder.Arg, but wraps `a` with parens in case of a Selector.
func (*Predicate) arg(b *Builder, a any) {
	switch a.(type) {
	case *Selector:
		b.Wrap(func(b *Builder) {
			b.Arg(a)
		})
	default:
		b.Arg(a)
	}
}

// clone returns a shallow clone of p.
func (p *Predicate) clone() *Predicate {
	if p == nil {
		return p
	}
	return &Predicate{fns: append([]func(*Builder){}, p.fns...)}
}

func (p *Predicate) mayWrap(preds []*Predicate, b *Builder, op string) {
	switch n := len(preds); {
	case n == 1:
		b.Join(preds[0])
		return
	case n > 1 && p.depth != 0:
		b.WriteByte('(')
		defer b.WriteByte(')')
	}
	for i := range preds {
		preds[i].depth = p.depth + 1
		if i > 0 {
			b.WriteByte(' ')
			b.WriteString(op)
			b.WriteByte(' ')
		}
		if len(preds[i].fns) > 1 {
			b.Wrap(func(b *Builder) {
				b.Join(preds[i])
			})
		} else {
			b.Join(preds[i])
		}
	}
}
