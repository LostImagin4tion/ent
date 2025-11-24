package query

// Expr returns an SQL expression that implements the Querier interface.
func Expr(exr string, args ...any) Querier { return &expr{s: exr, args: args} }

type expr struct {
	s    string
	args []any
}

func (e *expr) Query() (string, []any) { return e.s, e.args }

// ExprFunc returns an expression function that implements the Querier interface.
//
//	Update("users").
//		Set("x", ExprFunc(func(b *Builder) {
//			// The sql.Builder config (argc and dialect)
//			// was set before the function was executed.
//			b.Ident("x").WriteOp(OpAdd).Arg(1)
//		}))
func ExprFunc(fn func(*Builder)) Querier {
	return &exprFunc{fn: fn}
}

type exprFunc struct {
	Builder
	fn func(*Builder)
}

func (e *exprFunc) Query() (string, []any) {
	b := e.Builder.clone()
	e.fn(&b)
	return b.Query()
}
