package query

import (
	"testing"

	"entgo.io/ent/dialect"
	"github.com/stretchr/testify/require"
)

func TestReusePredicates(t *testing.T) {
	tests := []struct {
		p         *Predicate
		wantQuery string
		wantArgs  []any
	}{
		{
			p:         EQ("active", false),
			wantQuery: `SELECT * FROM "users" WHERE NOT "active"`,
		},
		{
			p: Or(
				EQ("a", "a"),
				EQ("b", "b"),
			),
			wantQuery: `SELECT * FROM "users" WHERE "a" = $1 OR "b" = $2`,
			wantArgs:  []any{"a", "b"},
		},
		{
			p: Or(
				EQ("a", "a"),
				In("b"),
			),
			wantQuery: `SELECT * FROM "users" WHERE "a" = $1 OR FALSE`,
			wantArgs:  []any{"a"},
		},
		{
			p: And(
				EQ("active", true),
				HasPrefix("name", "foo"),
				HasSuffix("name", "bar"),
				Or(
					In("id", Select("oid").From(Table("audit"))),
					In("id", Select("oid").From(Table("history"))),
				),
			),
			wantQuery: `SELECT * FROM "users" WHERE "active" AND "name" LIKE $1 AND "name" LIKE $2 AND ("id" IN (SELECT "oid" FROM "audit") OR "id" IN (SELECT "oid" FROM "history"))`,
			wantArgs:  []any{"foo%", "%bar"},
		},
		{
			p: func() *Predicate {
				t1 := Table("groups")
				pivot := Table("user_groups")
				matches := Select(pivot.C("user_id")).
					From(pivot).
					Join(t1).
					On(pivot.C("group_id"), t1.C("id")).
					Where(EQ(t1.C("name"), "ent"))
				return And(
					GT("balance", 0),
					In("id", matches),
					GT("balance", 100),
				)
			}(),
			wantQuery: `SELECT * FROM "users" WHERE "balance" > $1 AND "id" IN (SELECT "user_groups"."user_id" FROM "user_groups" JOIN "groups" AS "t1" ON "user_groups"."group_id" = "t1"."id" WHERE "t1"."name" = $2) AND "balance" > $3`,
			wantArgs:  []any{0, "ent", 100},
		},
	}
	for _, tt := range tests {
		query, args := Dialect(dialect.Postgres).Select().From(Table("users")).Where(tt.p).Query()
		require.Equal(t, tt.wantQuery, query)
		require.Equal(t, tt.wantArgs, args)
		query, args = Dialect(dialect.Postgres).Select().From(Table("users")).Where(tt.p).Query()
		require.Equal(t, tt.wantQuery, query)
		require.Equal(t, tt.wantArgs, args)
	}
}

func TestBoolPredicates(t *testing.T) {
	t1, t2 := Table("users"), Table("posts")
	query, args := Select().
		From(t1).
		Join(t2).
		On(t1.C("id"), t2.C("author_id")).
		Where(
			And(
				EQ(t1.C("active"), true),
				NEQ(t2.C("deleted"), true),
			),
		).
		Query()
	require.Nil(t, args)
	require.Equal(t, "SELECT * FROM `users` JOIN `posts` AS `t1` ON `users`.`id` = `t1`.`author_id` WHERE `users`.`active` AND NOT `t1`.`deleted`", query)
}

func TestColumnsHasPrefix(t *testing.T) {
	t.Run("MySQL", func(t *testing.T) {
		query, args := Dialect(dialect.MySQL).
			Select("*").From(Table("t1")).Where(ColumnsHasPrefix("a", "b")).Query()
		require.Equal(t, "SELECT * FROM `t1` WHERE `a` LIKE CONCAT(REPLACE(REPLACE(`b`, '_', '\\_'), '%', '\\%'), '%')", query)
		require.Empty(t, args)
	})
	t.Run("Postgres", func(t *testing.T) {
		query, args := Dialect(dialect.Postgres).
			Select("*").From(Table("t1")).Where(ColumnsHasPrefix("a", "b")).Query()
		require.Equal(t, `SELECT * FROM "t1" WHERE "a" LIKE (REPLACE(REPLACE("b", '_', '\_'), '%', '\%') || '%')`, query)
		require.Empty(t, args)
	})
	t.Run("SQLite", func(t *testing.T) {
		query, args := Dialect(dialect.SQLite).
			Select("*").From(Table("t1")).Where(ColumnsHasPrefix("a", "b")).Query()
		require.Equal(t, "SELECT * FROM `t1` WHERE `a` LIKE (REPLACE(REPLACE(`b`, '_', '\\_'), '%', '\\%') || '%') ESCAPE ?", query)
		require.Equal(t, []any{`\`}, args)
	})
}
