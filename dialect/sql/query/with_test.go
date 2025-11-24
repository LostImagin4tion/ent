package query

import (
	"strconv"
	"testing"

	"entgo.io/ent/dialect"
	"github.com/stretchr/testify/require"
)

func TestWithBuilder(t *testing.T) {
	tests := []struct {
		input     Querier
		wantQuery string
		wantArgs  []any
	}{
		{
			input: func() Querier {
				s1 := Select().
					From(Table("users")).
					Where(Not(And(EQ("name", "foo"), EQ("age", "bar"))))
				return Queries{With("users_view").As(s1), Select("name").From(Table("users_view"))}
			}(),
			wantQuery: "WITH `users_view` AS (SELECT * FROM `users` WHERE NOT (`name` = ? AND `age` = ?)) SELECT `name` FROM `users_view`",
			wantArgs:  []any{"foo", "bar"},
		},
		{
			input: func() Querier {
				d := Dialect(dialect.Postgres)
				s1 := d.Select().
					From(Table("users")).
					Where(Not(And(EQ("name", "foo"), EQ("age", "bar"))))
				return Queries{d.With("users_view").As(s1), d.Select("name").From(Table("users_view"))}
			}(),
			wantQuery: `WITH "users_view" AS (SELECT * FROM "users" WHERE NOT ("name" = $1 AND "age" = $2)) SELECT "name" FROM "users_view"`,
			wantArgs:  []any{"foo", "bar"},
		},
		{
			input:     Queries{With("users_view").As(Select().From(Table("users"))), Select().From(Table("users_view"))},
			wantQuery: "WITH `users_view` AS (SELECT * FROM `users`) SELECT * FROM `users_view`",
		},
		{
			input: func() Querier {
				base := Select("*").From(Table("groups"))
				return Queries{With("groups").As(base.Clone().Where(EQ("name", "bar"))), base.Select("age")}
			}(),
			wantQuery: "WITH `groups` AS (SELECT * FROM `groups` WHERE `name` = ?) SELECT `age` FROM `groups`",
			wantArgs:  []any{"bar"},
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
