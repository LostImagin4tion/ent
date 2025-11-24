package query

import (
	"strconv"
	"testing"

	"entgo.io/ent/dialect"
	"github.com/stretchr/testify/require"
)

func TestCreateViewBuilder(t *testing.T) {
	tests := []struct {
		input     Querier
		wantQuery string
		wantArgs  []any
	}{
		{
			input: CreateView("clean_users").
				Columns(
					Column("id").Type("int"),
					Column("name").Type("varchar(255)"),
				).
				As(Select("id", "name").From(Table("users"))),
			wantQuery: "CREATE VIEW `clean_users` (`id` int, `name` varchar(255)) AS SELECT `id`, `name` FROM `users`",
		},
		{
			input: Dialect(dialect.Postgres).
				CreateView("clean_users").
				Columns(
					Column("id").Type("int"),
					Column("name").Type("varchar(255)"),
				).
				As(Select("id", "name").From(Table("users"))),
			wantQuery: `CREATE VIEW "clean_users" ("id" int, "name" varchar(255)) AS SELECT "id", "name" FROM "users"`,
		},
		{
			input: CreateView("clean_users").
				Schema("schema").
				As(Select("id", "name").From(Table("users"))),
			wantQuery: "CREATE VIEW `schema`.`clean_users` AS SELECT `id`, `name` FROM `users`",
		},
	}
	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			query, args := tt.input.Query()
			require.Equal(t, tt.wantQuery, query)
			require.Equal(t, tt.wantArgs, args)
		})
	}
}
