package query

// Func represents an SQL function.
type Func struct {
	Builder
	fns []func(*Builder)
}

// Lower wraps the given column with the LOWER function.
//
//	P().EQ(sql.Lower("name"), "a8m")
func Lower(ident string) string {
	f := &Func{}
	f.Lower(ident)
	return f.String()
}

// Lower wraps the given ident with the LOWER function.
func (f *Func) Lower(ident string) {
	f.byName("LOWER", ident)
}

// Count wraps the ident with the COUNT aggregation function.
func Count(ident string) string {
	f := &Func{}
	f.Count(ident)
	return f.String()
}

// Count wraps the ident with the COUNT aggregation function.
func (f *Func) Count(ident string) {
	f.byName("COUNT", ident)
}

// Max wraps the ident with the MAX aggregation function.
func Max(ident string) string {
	f := &Func{}
	f.Max(ident)
	return f.String()
}

// Max wraps the ident with the MAX aggregation function.
func (f *Func) Max(ident string) {
	f.byName("MAX", ident)
}

// Min wraps the ident with the MIN aggregation function.
func Min(ident string) string {
	f := &Func{}
	f.Min(ident)
	return f.String()
}

// Min wraps the ident with the MIN aggregation function.
func (f *Func) Min(ident string) {
	f.byName("MIN", ident)
}

// Sum wraps the ident with the SUM aggregation function.
func Sum(ident string) string {
	f := &Func{}
	f.Sum(ident)
	return f.String()
}

// Sum wraps the ident with the SUM aggregation function.
func (f *Func) Sum(ident string) {
	f.byName("SUM", ident)
}

// Avg wraps the ident with the AVG aggregation function.
func Avg(ident string) string {
	f := &Func{}
	f.Avg(ident)
	return f.String()
}

// Avg wraps the ident with the AVG aggregation function.
func (f *Func) Avg(ident string) {
	f.byName("AVG", ident)
}

// byName wraps an identifier with a function name.
func (f *Func) byName(fn, ident string) {
	f.Append(func(b *Builder) {
		f.WriteString(fn)
		f.Wrap(func(b *Builder) {
			b.Ident(ident)
		})
	})
}

// Append appends a new function to the function callbacks.
// The callback list are executed on call to String.
func (f *Func) Append(fn func(*Builder)) *Func {
	f.fns = append(f.fns, fn)
	return f
}

// String implements the fmt.Stringer.
func (f *Func) String() string {
	for _, fn := range f.fns {
		fn(&f.Builder)
	}
	return f.Builder.String()
}
