package query

// TableView is a view that returns a table view. Can be a Table, Selector or a View (WITH statement).
type TableView interface {
	view()
	// C returns a formatted string prefixed
	// with the table view qualifier.
	C(string) string
}
