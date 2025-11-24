// Copyright 2019-present Facebook Inc. All rights reserved.
// This source code is licensed under the Apache 2.0 license found
// in the LICENSE file in the root directory of this source tree.

// Package sql provides wrappers around the standard database/sql package
// to allow the generated code to interact with a statically-typed API.
//
// Users that are interacting with this package should be aware that the
// following builders don't check the given SQL syntax nor validate or escape
// user-inputs. ~All validations are expected to be happened in the generated
// ent package.
package query

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"entgo.io/ent/dialect"
)

// Builder is the base query builder for the sql dsl.
type Builder struct {
	sb        *strings.Builder // underlying builder.
	dialect   string           // configured dialect.
	args      []any            // query parameters.
	total     int              // total number of parameters in query tree.
	errs      []error          // errors that added during the query construction.
	qualifier string           // qualifier to prefix identifiers (e.g. table name).
}

// Quote quotes the given identifier with the characters based
// on the configured dialect. It defaults to "`".
func (b *Builder) Quote(ident string) string {
	quote := "`"
	switch {
	case b.postgres():
		// If it was quoted with the wrong
		// identifier character.
		if strings.Contains(ident, "`") {
			return strings.ReplaceAll(ident, "`", `"`)
		}
		quote = `"`
	// An identifier for unknown dialect.
	case b.dialect == "" && strings.ContainsAny(ident, "`\""):
		return ident
	}
	return quote + ident + quote
}

// Ident appends the given string as an identifier.
func (b *Builder) Ident(s string) *Builder {
	switch {
	case len(s) == 0:
	case !strings.HasSuffix(s, "*") && !b.isIdent(s) && !IsFunc(s) && !isModifier(s) && !isAlias(s):
		if b.qualifier != "" {
			b.WriteString(b.Quote(b.qualifier)).WriteByte('.')
		}
		b.WriteString(b.Quote(s))
	case (IsFunc(s) || isModifier(s) || isAlias(s)) && b.postgres():
		// Modifiers and aggregation functions that
		// were called without dialect information.
		b.WriteString(strings.ReplaceAll(s, "`", `"`))
	default:
		b.WriteString(s)
	}
	return b
}

// IdentComma calls Ident on all arguments and adds a comma between them.
func (b *Builder) IdentComma(s ...string) *Builder {
	for i := range s {
		if i > 0 {
			b.Comma()
		}
		b.Ident(s[i])
	}
	return b
}

// String returns the accumulated string.
func (b *Builder) String() string {
	if b.sb == nil {
		return ""
	}
	return b.sb.String()
}

// WriteByte wraps the Buffer.WriteByte to make it chainable with other methods.
func (b *Builder) WriteByte(c byte) *Builder {
	if b.sb == nil {
		b.sb = &strings.Builder{}
	}
	b.sb.WriteByte(c)
	return b
}

// WriteString wraps the Buffer.WriteString to make it chainable with other methods.
func (b *Builder) WriteString(s string) *Builder {
	if b.sb == nil {
		b.sb = &strings.Builder{}
	}
	b.sb.WriteString(s)
	return b
}

// S is a short version of WriteString.
func (b *Builder) S(s string) *Builder {
	return b.WriteString(s)
}

// Len returns the number of accumulated bytes.
func (b *Builder) Len() int {
	if b.sb == nil {
		return 0
	}
	return b.sb.Len()
}

// Reset resets the Builder to be empty.
func (b *Builder) Reset() *Builder {
	if b.sb != nil {
		b.sb.Reset()
	}
	return b
}

// AddError appends an error to the builder errors.
func (b *Builder) AddError(err error) *Builder {
	// allowed nil error make build process easier
	if err != nil {
		b.errs = append(b.errs, err)
	}
	return b
}

func (b *Builder) writeSchema(schema string) {
	if schema != "" && b.dialect != dialect.SQLite {
		b.Ident(schema).WriteByte('.')
	}
}

// Err returns a concatenated error of all errors encountered during
// the query-building, or were added manually by calling AddError.
func (b *Builder) Err() error {
	if len(b.errs) == 0 {
		return nil
	}
	br := strings.Builder{}
	for i := range b.errs {
		if i > 0 {
			br.WriteString("; ")
		}
		br.WriteString(b.errs[i].Error())
	}
	return errors.New(br.String())
}

// An Op represents an operator.
type Op int

// Predicate and arithmetic operators.
const (
	OpEQ      Op = iota // =
	OpNEQ               // <>
	OpGT                // >
	OpGTE               // >=
	OpLT                // <
	OpLTE               // <=
	OpIn                // IN
	OpNotIn             // NOT IN
	OpLike              // LIKE
	OpIsNull            // IS NULL
	OpNotNull           // IS NOT NULL
	OpAdd               // +
	OpSub               // -
	OpMul               // *
	OpDiv               // / (Quotient)
	OpMod               // % (Reminder)
)

var ops = [...]string{
	OpEQ:      "=",
	OpNEQ:     "<>",
	OpGT:      ">",
	OpGTE:     ">=",
	OpLT:      "<",
	OpLTE:     "<=",
	OpIn:      "IN",
	OpNotIn:   "NOT IN",
	OpLike:    "LIKE",
	OpIsNull:  "IS NULL",
	OpNotNull: "IS NOT NULL",
	OpAdd:     "+",
	OpSub:     "-",
	OpMul:     "*",
	OpDiv:     "/",
	OpMod:     "%",
}

// WriteOp writes an operator to the builder.
func (b *Builder) WriteOp(op Op) *Builder {
	switch {
	case op >= OpEQ && op <= OpLike || op >= OpAdd && op <= OpMod:
		b.Pad().WriteString(ops[op]).Pad()
	case op == OpIsNull || op == OpNotNull:
		b.Pad().WriteString(ops[op])
	default:
		panic(fmt.Sprintf("invalid op %d", op))
	}
	return b
}

type (
	// StmtInfo holds an information regarding
	// the statement
	StmtInfo struct {
		// The Dialect of the SQL driver.
		Dialect string
	}
	// ParamFormatter wraps the FormatPram function.
	ParamFormatter interface {
		// The FormatParam function lets users define
		// custom placeholder formatting for their types.
		// For example, formatting the default placeholder
		// from '?' to 'ST_GeomFromWKB(?)' for MySQL dialect.
		FormatParam(placeholder string, info *StmtInfo) string
	}
)

// Arg appends an input argument to the builder.
func (b *Builder) Arg(a any) *Builder {
	switch v := a.(type) {
	case nil:
		b.WriteString("NULL")
		return b
	case *raw:
		b.WriteString(v.s)
		return b
	case Querier:
		b.Join(v)
		return b
	}
	// Default placeholder param (MySQL and SQLite).
	format := "?"
	if b.postgres() {
		// Postgres' arguments are referenced using the syntax $n.
		// $1 refers to the 1st argument, $2 to the 2nd, and so on.
		format = "$" + strconv.Itoa(b.total+1)
	}
	if f, ok := a.(ParamFormatter); ok {
		format = f.FormatParam(format, &StmtInfo{
			Dialect: b.dialect,
		})
	}
	return b.Argf(format, a)
}

// Args appends a list of arguments to the builder.
func (b *Builder) Args(a ...any) *Builder {
	for i := range a {
		if i > 0 {
			b.Comma()
		}
		b.Arg(a[i])
	}
	return b
}

// Argf appends an input argument to the builder
// with the given format. For example:
//
//	FormatArg("JSON(?)", b).
//	FormatArg("ST_GeomFromText(?)", geom)
func (b *Builder) Argf(format string, a any) *Builder {
	switch a := a.(type) {
	case nil:
		b.WriteString("NULL")
		return b
	case *raw:
		b.WriteString(a.s)
		return b
	case Querier:
		b.Join(a)
		return b
	}
	b.total++
	b.args = append(b.args, a)
	b.WriteString(format)
	return b
}

// Comma adds a comma to the query.
func (b *Builder) Comma() *Builder {
	return b.WriteString(", ")
}

// Pad adds a space to the query.
func (b *Builder) Pad() *Builder {
	return b.WriteByte(' ')
}

// Join joins a list of Queries to the builder.
func (b *Builder) Join(qs ...Querier) *Builder {
	return b.join(qs, "")
}

// JoinComma joins a list of Queries and adds comma between them.
func (b *Builder) JoinComma(qs ...Querier) *Builder {
	return b.join(qs, ", ")
}

// join a list of Queries to the builder with a given separator.
func (b *Builder) join(qs []Querier, sep string) *Builder {
	for i, q := range qs {
		if i > 0 {
			b.WriteString(sep)
		}
		st, ok := q.(state)
		if ok {
			st.SetDialect(b.dialect)
			st.SetTotal(b.total)
		}
		query, args := q.Query()
		b.WriteString(query)
		b.args = append(b.args, args...)
		b.total += len(args)
		if qe, ok := q.(querierErr); ok {
			if err := qe.Err(); err != nil {
				b.AddError(err)
			}
		}
	}
	return b
}

// Wrap gets a callback, and wraps its result with parentheses.
func (b *Builder) Wrap(f func(*Builder)) *Builder {
	nb := &Builder{dialect: b.dialect, total: b.total, sb: &strings.Builder{}}
	nb.WriteByte('(')
	f(nb)
	nb.WriteByte(')')
	b.WriteString(nb.String())
	b.args = append(b.args, nb.args...)
	b.total = nb.total
	return b
}

// Nested gets a callback, and wraps its result with parentheses.
//
// Deprecated: Use Builder.Wrap instead.
func (b *Builder) Nested(f func(*Builder)) *Builder {
	return b.Wrap(f)
}

// SetDialect sets the builder dialect. It's used for garnering dialect specific queries.
func (b *Builder) SetDialect(dialect string) {
	b.dialect = dialect
}

// Dialect returns the dialect of the builder.
func (b Builder) Dialect() string {
	return b.dialect
}

// Total returns the total number of arguments so far.
func (b Builder) Total() int {
	return b.total
}

// SetTotal sets the value of the total arguments.
// Used to pass this information between sub queries/expressions.
func (b *Builder) SetTotal(total int) {
	b.total = total
}

// Query implements the Querier interface.
func (b Builder) Query() (string, []any) {
	return b.String(), b.args
}

// clone returns a shallow clone of a builder.
func (b Builder) clone() Builder {
	c := Builder{dialect: b.dialect, total: b.total, sb: &strings.Builder{}}
	if len(b.args) > 0 {
		c.args = append(c.args, b.args...)
	}
	if b.sb != nil {
		c.sb.WriteString(b.sb.String())
	}
	return c
}

// postgres reports if the builder dialect is PostgreSQL.
func (b Builder) postgres() bool {
	return b.Dialect() == dialect.Postgres
}

// sqlite reports if the builder dialect is SQLite.
func (b Builder) sqlite() bool {
	return b.Dialect() == dialect.SQLite
}

// fromIdent sets the builder dialect from the identifier format.
func (b *Builder) fromIdent(ident string) {
	if strings.Contains(ident, `"`) {
		b.SetDialect(dialect.Postgres)
	}
	// otherwise, use the default.
}

// isIdent reports if the given string is a dialect identifier.
func (b *Builder) isIdent(s string) bool {
	switch {
	case b.postgres():
		return strings.Contains(s, `"`)
	default:
		return strings.Contains(s, "`")
	}
}

// unquote database identifiers.
func (b *Builder) unquote(s string) string {
	switch pg := b.postgres(); {
	case len(s) < 2:
	case !pg && s[0] == '`' && s[len(s)-1] == '`', pg && s[0] == '"' && s[len(s)-1] == '"':
		if u, err := strconv.Unquote(s); err == nil {
			return u
		}
	}
	return s
}

// isQualified reports if the given string is a qualified identifier.
func (b *Builder) isQualified(s string) bool {
	ident, pg := b.isIdent(s), b.postgres()
	return !ident && len(s) > 2 && strings.ContainsRune(s[1:len(s)-1], '.') || // <qualifier>.<column>
		ident && pg && strings.Contains(s, `"."`) || // "qualifier"."column"
		ident && !pg && strings.Contains(s, "`.`") // `qualifier`.`column`
}

// state wraps all methods for setting and getting
// update state between all queries in the query tree.
type state interface {
	Dialect() string
	SetDialect(string)
	Total() int
	SetTotal(int)
}
