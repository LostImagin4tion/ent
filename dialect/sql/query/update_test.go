package query

import (
	"strconv"
	"testing"

	"entgo.io/ent/dialect"
	"github.com/stretchr/testify/require"
)

func TestUpdateBuilder(t *testing.T) {
	tests := []struct {
		input     Querier
		wantQuery string
		wantArgs  []any
	}{
		{
			input:     Update("users").Set("name", "foo"),
			wantQuery: "UPDATE `users` SET `name` = ?",
			wantArgs:  []any{"foo"},
		},
		{
			input:     Update("users").Set("name", "foo").Schema("mydb"),
			wantQuery: "UPDATE `mydb`.`users` SET `name` = ?",
			wantArgs:  []any{"foo"},
		},
		{
			input:     Dialect(dialect.Postgres).Update("users").Set("name", "foo"),
			wantQuery: `UPDATE "users" SET "name" = $1`,
			wantArgs:  []any{"foo"},
		},
		{
			input:     Dialect(dialect.Postgres).Update("users").Set("name", "foo").Returning("*"),
			wantQuery: `UPDATE "users" SET "name" = $1 RETURNING *`,
			wantArgs:  []any{"foo"},
		},
		{
			input:     Dialect(dialect.Postgres).Update("users").Set("name", "foo").Returning("id", "name"),
			wantQuery: `UPDATE "users" SET "name" = $1 RETURNING "id", "name"`,
			wantArgs:  []any{"foo"},
		},
		{
			input:     Dialect(dialect.Postgres).Update("users").Set("name", "foo").Schema("mydb"),
			wantQuery: `UPDATE "mydb"."users" SET "name" = $1`,
			wantArgs:  []any{"foo"},
		},
		{
			input:     Dialect(dialect.SQLite).Update("users").Set("name", "foo").Schema("mydb"),
			wantQuery: "UPDATE `users` SET `name` = ?",
			wantArgs:  []any{"foo"},
		},
		{
			input:     Update("users").Set("name", "foo").Set("age", 10),
			wantQuery: "UPDATE `users` SET `name` = ?, `age` = ?",
			wantArgs:  []any{"foo", 10},
		},
		{
			input:     Dialect(dialect.SQLite).Update("users").Set("name", "foo").Returning("id", "name").OrderBy("name").Limit(10),
			wantQuery: "UPDATE `users` SET `name` = ? RETURNING `id`, `name` ORDER BY `name` LIMIT 10",
			wantArgs:  []any{"foo"},
		},
		{
			input:     Dialect(dialect.Postgres).Update("users").Set("name", "foo").Set("age", 10),
			wantQuery: `UPDATE "users" SET "name" = $1, "age" = $2`,
			wantArgs:  []any{"foo", 10},
		},
		{
			input: Dialect(dialect.Postgres).Update("users").
				Set("active", false).
				Where(P(func(b *Builder) {
					b.Ident("name").WriteString(" SIMILAR TO ").Arg("(b|c)%")
				})),
			wantQuery: `UPDATE "users" SET "active" = $1 WHERE "name" SIMILAR TO $2`,
			wantArgs:  []any{false, "(b|c)%"},
		},
		{
			input:     Update("users").Set("name", "foo").Where(EQ("name", "bar")),
			wantQuery: "UPDATE `users` SET `name` = ? WHERE `name` = ?",
			wantArgs:  []any{"foo", "bar"},
		},
		{
			input:     Update("users").Set("name", "foo").Where(EQ("name", Expr("?", "bar"))),
			wantQuery: "UPDATE `users` SET `name` = ? WHERE `name` = ?",
			wantArgs:  []any{"foo", "bar"},
		},
		{
			input:     Dialect(dialect.Postgres).Update("users").Set("name", "foo").Where(EQ("name", "bar")),
			wantQuery: `UPDATE "users" SET "name" = $1 WHERE "name" = $2`,
			wantArgs:  []any{"foo", "bar"},
		},
		{
			input: func() Querier {
				p1, p2 := EQ("name", "bar"), Or(EQ("age", 10), EQ("age", 20))
				return Dialect(dialect.Postgres).
					Update("users").
					Set("name", "foo").
					Where(p1).
					Where(p2).
					Where(p1).
					Where(p2)
			}(),
			wantQuery: `UPDATE "users" SET "name" = $1 WHERE (("name" = $2 AND ("age" = $3 OR "age" = $4)) AND "name" = $5) AND ("age" = $6 OR "age" = $7)`,
			wantArgs:  []any{"foo", "bar", 10, 20, "bar", 10, 20},
		},
		{
			input:     Update("users").Set("name", "foo").SetNull("spouse_id"),
			wantQuery: "UPDATE `users` SET `spouse_id` = NULL, `name` = ?",
			wantArgs:  []any{"foo"},
		},
		{
			input:     Dialect(dialect.Postgres).Update("users").Set("name", "foo").SetNull("spouse_id"),
			wantQuery: `UPDATE "users" SET "spouse_id" = NULL, "name" = $1`,
			wantArgs:  []any{"foo"},
		},
		{
			input: Update("users").Set("name", "foo").
				Where(EQ("name", "bar")).
				Where(EQ("age", 20)),
			wantQuery: "UPDATE `users` SET `name` = ? WHERE `name` = ? AND `age` = ?",
			wantArgs:  []any{"foo", "bar", 20},
		},
		{
			input: Dialect(dialect.Postgres).
				Update("users").
				Set("name", "foo").
				Where(EQ("name", "bar")).
				Where(EQ("age", 20)),
			wantQuery: `UPDATE "users" SET "name" = $1 WHERE "name" = $2 AND "age" = $3`,
			wantArgs:  []any{"foo", "bar", 20},
		},
		{
			input: Update("users").
				Set("name", "foo").
				Set("age", 10).
				Where(Or(EQ("name", "bar"), EQ("name", "baz"))),
			wantQuery: "UPDATE `users` SET `name` = ?, `age` = ? WHERE `name` = ? OR `name` = ?",
			wantArgs:  []any{"foo", 10, "bar", "baz"},
		},
		{
			input: Dialect(dialect.Postgres).
				Update("users").
				Set("name", "foo").
				Set("age", 10).
				Where(Or(EQ("name", "bar"), EQ("name", "baz"))),
			wantQuery: `UPDATE "users" SET "name" = $1, "age" = $2 WHERE "name" = $3 OR "name" = $4`,
			wantArgs:  []any{"foo", 10, "bar", "baz"},
		},
		{
			input: Update("users").
				Set("name", "foo").
				Set("age", 10).
				Where(P().EQ("name", "foo")),
			wantQuery: "UPDATE `users` SET `name` = ?, `age` = ? WHERE `name` = ?",
			wantArgs:  []any{"foo", 10, "foo"},
		},
		{
			input: Dialect(dialect.Postgres).
				Update("users").
				Add("rank", 10).
				Where(
					Or(
						EQ("rank", Select("rank").From(Table("ranks")).Where(EQ("name", "foo"))),
						GT("score", Select("score").From(Table("scores")).Where(GT("count", 0))),
					),
				),
			wantQuery: `UPDATE "users" SET "rank" = COALESCE("users"."rank", 0) + $1 WHERE "rank" = (SELECT "rank" FROM "ranks" WHERE "name" = $2) OR "score" > (SELECT "score" FROM "scores" WHERE "count" > $3)`,
			wantArgs:  []any{10, "foo", 0},
		},
		{
			input: Update("users").
				Add("rank", 10).
				Where(
					Or(
						EQ("rank", Select("rank").From(Table("ranks")).Where(EQ("name", "foo"))),
						GT("score", Select("score").From(Table("scores")).Where(GT("count", 0))),
					),
				),
			wantQuery: "UPDATE `users` SET `rank` = COALESCE(`users`.`rank`, 0) + ? WHERE `rank` = (SELECT `rank` FROM `ranks` WHERE `name` = ?) OR `score` > (SELECT `score` FROM `scores` WHERE `count` > ?)",
			wantArgs:  []any{10, "foo", 0},
		},
		{
			input: Dialect(dialect.Postgres).
				Update("users").
				Set("name", "foo").
				Set("age", 10).
				Where(P().EQ("name", "foo")),
			wantQuery: `UPDATE "users" SET "name" = $1, "age" = $2 WHERE "name" = $3`,
			wantArgs:  []any{"foo", 10, "foo"},
		},
		{
			input: Update("users").
				Set("name", "foo").
				Where(And(In("name", "bar", "baz"), NotIn("age", 1, 2))),
			wantQuery: "UPDATE `users` SET `name` = ? WHERE `name` IN (?, ?) AND `age` NOT IN (?, ?)",
			wantArgs:  []any{"foo", "bar", "baz", 1, 2},
		},
		{
			input: Dialect(dialect.Postgres).
				Update("users").
				Set("name", "foo").
				Where(And(In("name", "bar", "baz"), NotIn("age", 1, 2))),
			wantQuery: `UPDATE "users" SET "name" = $1 WHERE "name" IN ($2, $3) AND "age" NOT IN ($4, $5)`,
			wantArgs:  []any{"foo", "bar", "baz", 1, 2},
		},
		{
			input: Update("users").
				Set("name", "foo").
				Where(And(HasPrefix("nickname", "a8m"), Contains("lastname", "mash"))),
			wantQuery: "UPDATE `users` SET `name` = ? WHERE `nickname` LIKE ? AND `lastname` LIKE ?",
			wantArgs:  []any{"foo", "a8m%", "%mash%"},
		},
		{
			input: Dialect(dialect.Postgres).
				Update("users").
				Set("name", "foo").
				Where(And(HasPrefix("nickname", "a8m"), Contains("lastname", "mash"))),
			wantQuery: `UPDATE "users" SET "name" = $1 WHERE "nickname" LIKE $2 AND "lastname" LIKE $3`,
			wantArgs:  []any{"foo", "a8m%", "%mash%"},
		},
		{
			input: Update("users").
				Add("age", 1).
				Where(HasPrefix("nickname", "a8m")),
			wantQuery: "UPDATE `users` SET `age` = COALESCE(`users`.`age`, 0) + ? WHERE `nickname` LIKE ?",
			wantArgs:  []any{1, "a8m%"},
		},
		{
			input: Update("users").
				Set("age", 1).
				Add("age", 2).
				Where(HasPrefix("nickname", "a8m")),
			wantQuery: "UPDATE `users` SET `age` = ?, `age` = COALESCE(`users`.`age`, 0) + ? WHERE `nickname` LIKE ?",
			wantArgs:  []any{1, 2, "a8m%"},
		},
		{
			input: Update("users").
				Add("age", 2).
				Set("age", 1).
				Where(HasPrefix("nickname", "a8m")),
			wantQuery: "UPDATE `users` SET `age` = ? WHERE `nickname` LIKE ?",
			wantArgs:  []any{1, "a8m%"},
		},
		{
			input: Dialect(dialect.Postgres).
				Update("users").
				Add("age", 1).
				Where(HasPrefix("nickname", "a8m")),
			wantQuery: `UPDATE "users" SET "age" = COALESCE("users"."age", 0) + $1 WHERE "nickname" LIKE $2`,
			wantArgs:  []any{1, "a8m%"},
		},
		{
			input: Update("users").
				Set("name", "foo").
				Where(And(HasPrefixFold("nickname", "a8m"), Contains("lastname", "mash"))),
			wantQuery: "UPDATE `users` SET `name` = ? WHERE LOWER(`nickname`) LIKE ? AND `lastname` LIKE ?",
			wantArgs:  []any{"foo", "a8m%", "%mash%"},
		},
		{
			input: Dialect(dialect.Postgres).
				Update("users").
				Set("name", "foo").
				Where(And(HasPrefixFold("nickname", "a8m"), Contains("lastname", "mash"))),
			wantQuery: `UPDATE "users" SET "name" = $1 WHERE "nickname" ILIKE $2 AND "lastname" LIKE $3`,
			wantArgs:  []any{"foo", "a8m%", "%mash%"},
		},
		{
			input: Update("users").
				Add("age", 1).
				Where(HasPrefixFold("nickname", "a8m")),
			wantQuery: "UPDATE `users` SET `age` = COALESCE(`users`.`age`, 0) + ? WHERE LOWER(`nickname`) LIKE ?",
			wantArgs:  []any{1, "a8m%"},
		},
		{
			input: Update("users").
				Set("age", 1).
				Add("age", 2).
				Where(HasPrefixFold("nickname", "a8m")),
			wantQuery: "UPDATE `users` SET `age` = ?, `age` = COALESCE(`users`.`age`, 0) + ? WHERE LOWER(`nickname`) LIKE ?",
			wantArgs:  []any{1, 2, "a8m%"},
		},
		{
			input: Update("users").
				Add("age", 2).
				Set("age", 1).
				Where(HasPrefixFold("nickname", "a8m")),
			wantQuery: "UPDATE `users` SET `age` = ? WHERE LOWER(`nickname`) LIKE ?",
			wantArgs:  []any{1, "a8m%"},
		},
		{
			input: Dialect(dialect.Postgres).
				Update("users").
				Add("age", 1).
				Where(HasPrefixFold("nickname", "a8m")),
			wantQuery: `UPDATE "users" SET "age" = COALESCE("users"."age", 0) + $1 WHERE "nickname" ILIKE $2`,
			wantArgs:  []any{1, "a8m%"},
		},
		{
			input: Dialect(dialect.Postgres).
				Update("users").
				Set("name", "foo").
				Where(And(HasSuffixFold("nickname", "a8m"), Contains("lastname", "mash"))),
			wantQuery: `UPDATE "users" SET "name" = $1 WHERE "nickname" ILIKE $2 AND "lastname" LIKE $3`,
			wantArgs:  []any{"foo", "%a8m", "%mash%"},
		},
		{
			input: Update("users").
				Add("age", 1).
				Where(HasSuffixFold("nickname", "a8m")),
			wantQuery: "UPDATE `users` SET `age` = COALESCE(`users`.`age`, 0) + ? WHERE LOWER(`nickname`) LIKE ?",
			wantArgs:  []any{1, "%a8m"},
		},
		{
			input: Update("users").
				Set("age", 1).
				Add("age", 2).
				Where(HasSuffixFold("nickname", "a8m")),
			wantQuery: "UPDATE `users` SET `age` = ?, `age` = COALESCE(`users`.`age`, 0) + ? WHERE LOWER(`nickname`) LIKE ?",
			wantArgs:  []any{1, 2, "%a8m"},
		},
		{
			input: Update("users").
				Add("age", 2).
				Set("age", 1).
				Where(HasSuffixFold("nickname", "a8m")),
			wantQuery: "UPDATE `users` SET `age` = ? WHERE LOWER(`nickname`) LIKE ?",
			wantArgs:  []any{1, "%a8m"},
		},
		{
			input: Dialect(dialect.Postgres).
				Update("users").
				Add("age", 1).
				Where(HasSuffixFold("nickname", "a8m")),
			wantQuery: `UPDATE "users" SET "age" = COALESCE("users"."age", 0) + $1 WHERE "nickname" ILIKE $2`,
			wantArgs:  []any{1, "%a8m"},
		},
		{
			input: Update("users").
				Add("age", 1).
				Set("nickname", "a8m").
				Add("version", 10).
				Set("name", "mashraki"),
			wantQuery: "UPDATE `users` SET `age` = COALESCE(`users`.`age`, 0) + ?, `nickname` = ?, `version` = COALESCE(`users`.`version`, 0) + ?, `name` = ?",
			wantArgs:  []any{1, "a8m", 10, "mashraki"},
		},
		{
			input: Dialect(dialect.Postgres).
				Update("users").
				Add("age", 1).
				Set("nickname", "a8m").
				Add("version", 10).
				Set("name", "mashraki"),
			wantQuery: `UPDATE "users" SET "age" = COALESCE("users"."age", 0) + $1, "nickname" = $2, "version" = COALESCE("users"."version", 0) + $3, "name" = $4`,
			wantArgs:  []any{1, "a8m", 10, "mashraki"},
		},
		{
			input: Dialect(dialect.Postgres).
				Update("users").
				Add("age", 1).
				Set("nickname", "a8m").
				Add("version", 10).
				Set("name", "mashraki").
				Set("first", "ariel").
				Add("score", 1e5).
				Where(Or(EQ("age", 1), EQ("age", 2))),
			wantQuery: `UPDATE "users" SET "age" = COALESCE("users"."age", 0) + $1, "nickname" = $2, "version" = COALESCE("users"."version", 0) + $3, "name" = $4, "first" = $5, "score" = COALESCE("users"."score", 0) + $6 WHERE "age" = $7 OR "age" = $8`,
			wantArgs:  []any{1, "a8m", 10, "mashraki", "ariel", 1e5, 1, 2},
		},
		{
			input: Update("users").
				Set("name", "foo").
				Set("age", 10).
				Where(And(EQ("name", "foo"), EQ("age", 20))),
			wantQuery: "UPDATE `users` SET `name` = ?, `age` = ? WHERE `name` = ? AND `age` = ?",
			wantArgs:  []any{"foo", 10, "foo", 20},
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

func TestUpdateBuilder_SetExpr(t *testing.T) {
	d := Dialect(dialect.Postgres)
	excluded := d.Table("excluded")
	query, args := d.Update("users").
		Set("name", "Ariel").
		Set("active", Expr("NOT(active)")).
		Set("age", Expr(excluded.C("age"))).
		Set("x", ExprFunc(func(b *Builder) {
			b.WriteString(excluded.C("x")).WriteString(" || ' (formerly ' || ").Ident("x").WriteString(" || ')'")
		})).
		Set("y", ExprFunc(func(b *Builder) {
			b.Arg("~").WriteOp(OpAdd).WriteString(excluded.C("y")).WriteOp(OpAdd).Arg("~")
		})).
		Query()
	require.Equal(t, `UPDATE "users" SET "name" = $1, "active" = NOT(active), "age" = "excluded"."age", "x" = "excluded"."x" || ' (formerly ' || "x" || ')', "y" = $2 + "excluded"."y" + $3`, query)
	require.Equal(t, []any{"Ariel", "~", "~"}, args)
}

func TestUpdateBuilder_OrderBy(t *testing.T) {
	u := Dialect(dialect.MySQL).Update("users").Set("id", Expr("`id` + 1")).OrderBy("id")
	require.NoError(t, u.Err())
	query, args := u.Query()
	require.Nil(t, args)
	require.Equal(t, "UPDATE `users` SET `id` = `id` + 1 ORDER BY `id`", query)

	u = Dialect(dialect.Postgres).Update("users").Set("id", Expr("id + 1")).OrderBy("id")
	require.Error(t, u.Err())
}

func TestUpdateBuilder_WithPrefix(t *testing.T) {
	u := Dialect(dialect.MySQL).
		Update("users").
		Prefix(ExprFunc(func(b *Builder) {
			b.WriteString("SET @i = ").Arg(1).WriteByte(';')
		})).
		Set("id", Expr("(@i:=@i+1)")).
		OrderBy("id")
	require.NoError(t, u.Err())
	query, args := u.Query()
	require.Equal(t, []any{1}, args)
	require.Equal(t, "SET @i = ?; UPDATE `users` SET `id` = (@i:=@i+1) ORDER BY `id`", query)

	u = Dialect(dialect.MySQL).
		Update("users").
		Prefix(Expr("SET @i = 1;")).
		Set("id", Expr("(@i:=@i+1)")).
		OrderBy("id")
	require.NoError(t, u.Err())
	query, args = u.Query()
	require.Empty(t, args)
	require.Equal(t, "SET @i = 1; UPDATE `users` SET `id` = (@i:=@i+1) ORDER BY `id`", query)
}
