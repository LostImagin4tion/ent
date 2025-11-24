package query

import (
	"strconv"
	"strings"
	"testing"

	"entgo.io/ent/dialect"
	"github.com/stretchr/testify/require"
)

func TestSelector(t *testing.T) {
	tests := []struct {
		input     Querier
		wantQuery string
		wantArgs  []any
	}{
		{
			input: Select().
				From(Table("users")).
				Where(EQ("name", "Alex")),
			wantQuery: "SELECT * FROM `users` WHERE `name` = ?",
			wantArgs:  []any{"Alex"},
		},
		{
			input: Dialect(dialect.Postgres).
				Select().
				From(Table("users")),
			wantQuery: `SELECT * FROM "users"`,
		},
		{
			input: Dialect(dialect.Postgres).
				Select().
				From(Table("users")).
				Where(EQ("name", "Ariel")),
			wantQuery: `SELECT * FROM "users" WHERE "name" = $1`,
			wantArgs:  []any{"Ariel"},
		},
		{
			input: Select().
				From(Table("users")).
				Where(Or(EQ("name", "BAR"), EQ("name", "BAZ"))),
			wantQuery: "SELECT * FROM `users` WHERE `name` = ? OR `name` = ?",
			wantArgs:  []any{"BAR", "BAZ"},
		},
		{
			input: func() Querier {
				t1, t2 := Table("users"), Table("pets")
				return Dialect(dialect.Postgres).
					Select().
					From(t1).
					Where(GT(t1.C("age"), 30)).
					Where(
						And(
							Exists(Select().From(t2).Where(ColumnsEQ(t2.C("owner_id"), t1.C("id")))),
							NotExists(Select().From(t2).Where(ColumnsEQ(t2.C("owner_id"), t1.C("id")))),
						),
					)
			}(),
			wantQuery: `SELECT * FROM "users" WHERE "users"."age" > $1 AND (EXISTS (SELECT * FROM "pets" WHERE "pets"."owner_id" = "users"."id") AND NOT EXISTS (SELECT * FROM "pets" WHERE "pets"."owner_id" = "users"."id"))`,
			wantArgs:  []any{30},
		},
		{
			input:     Select().From(Table("users")),
			wantQuery: "SELECT * FROM `users`",
		},
		{
			input:     Dialect(dialect.Postgres).Select().From(Table("users")),
			wantQuery: `SELECT * FROM "users"`,
		},
		{
			input:     Select().From(Table("users").Unquote()),
			wantQuery: "SELECT * FROM users",
		},
		{
			input:     Dialect(dialect.Postgres).Select().From(Table("users").Unquote()),
			wantQuery: "SELECT * FROM users",
		},
		{
			input:     Select().From(Table("users").As("u")),
			wantQuery: "SELECT * FROM `users` AS `u`",
		},
		{
			input:     Dialect(dialect.Postgres).Select().From(Table("users").As("u")),
			wantQuery: `SELECT * FROM "users" AS "u"`,
		},
		{
			input: func() Querier {
				t1 := Table("users").As("u")
				t2 := Table("groups").As("g")
				return Dialect(dialect.Postgres).Select(t1.C("id"), t2.C("name")).From(t1).Join(t2)
			}(),
			wantQuery: `SELECT "u"."id", "g"."name" FROM "users" AS "u" JOIN "groups" AS "g"`,
		},
		{
			input: func() Querier {
				t1 := Table("users").As("u")
				return Select(t1.Columns("name", "age")...).From(t1)
			}(),
			wantQuery: "SELECT `u`.`name`, `u`.`age` FROM `users` AS `u`",
		},
		{
			input: func() Querier {
				t1 := Table("users").As("u")
				return Dialect(dialect.Postgres).
					Select(t1.Columns("name", "age")...).From(t1)
			}(),
			wantQuery: `SELECT "u"."name", "u"."age" FROM "users" AS "u"`,
		},
		{
			input: func() Querier {
				t1 := Dialect(dialect.Postgres).
					Table("users").As("u")
				return Dialect(dialect.Postgres).
					Select(t1.Columns("name", "age")...).From(t1)
			}(),
			wantQuery: `SELECT "u"."name", "u"."age" FROM "users" AS "u"`,
		},
		{
			input: func() Querier {
				selector := Select().From(Table("users")).As("t")
				return selector.Select(selector.C("name"))
			}(),
			wantQuery: "SELECT `t`.`name` FROM `users`",
		},
		{
			input: func() Querier {
				selector := Dialect(dialect.Postgres).
					Select().From(Table("users")).As("t")
				return selector.Select(selector.C("name"))
			}(),
			wantQuery: `SELECT "t"."name" FROM "users"`,
		},
		{
			input: Select().
				From(Table("users")).
				Where(Not(And(EQ("name", "foo"), EQ("age", "bar")))),
			wantQuery: "SELECT * FROM `users` WHERE NOT (`name` = ? AND `age` = ?)",
			wantArgs:  []any{"foo", "bar"},
		},
		{
			input: Dialect(dialect.Postgres).
				Select().
				From(Table("users")).
				Where(Not(And(EQ("name", "foo"), EQ("age", "bar")))),
			wantQuery: `SELECT * FROM "users" WHERE NOT ("name" = $1 AND "age" = $2)`,
			wantArgs:  []any{"foo", "bar"},
		},
		{
			input: Select().
				From(Table("users")).
				Where(Or(EqualFold("name", "BAR"), EqualFold("name", "BAZ"))),
			wantQuery: "SELECT * FROM `users` WHERE LOWER(`name`) = ? OR LOWER(`name`) = ?",
			wantArgs:  []any{"bar", "baz"},
		},
		{
			input: Dialect(dialect.Postgres).
				Select().
				From(Table("users")).
				Where(Or(EqualFold("name", "BAR"), EqualFold("name", "BAZ"))),
			wantQuery: `SELECT * FROM "users" WHERE "name" ILIKE $1 OR "name" ILIKE $2`,
			wantArgs:  []any{"bar", "baz"},
		},
		{
			input: Dialect(dialect.Postgres).
				Select().
				From(Table("users")).
				Where(Or(EqualFold("name", "BAR%"), EqualFold("name", "%BAZ"))),
			wantQuery: `SELECT * FROM "users" WHERE "name" ILIKE $1 OR "name" ILIKE $2`,
			wantArgs:  []any{"bar\\%", "\\%baz"},
		},
		{
			input: Dialect(dialect.Postgres).
				Select().
				From(Table("users")).
				Where(Or(EqualFold("name", "BAR\\"), EqualFold("name", "\\BAZ"))),
			wantQuery: `SELECT * FROM "users" WHERE "name" ILIKE $1 OR "name" ILIKE $2`,
			wantArgs:  []any{"bar\\\\", "\\\\baz"},
		},
		{
			input: Dialect(dialect.MySQL).
				Select().
				From(Table("users")).
				Where(Or(EqualFold("name", "BAR"), EqualFold("name", "BAZ"))),
			wantQuery: "SELECT * FROM `users` WHERE `name` COLLATE utf8mb4_general_ci = ? OR `name` COLLATE utf8mb4_general_ci = ?",
			wantArgs:  []any{"bar", "baz"},
		},
		{
			input: Dialect(dialect.SQLite).
				Select().
				From(Table("users")).
				Where(And(ContainsFold("name", "Ariel"), ContainsFold("nick", "Bar"))),
			wantQuery: "SELECT * FROM `users` WHERE LOWER(`name`) LIKE ? AND LOWER(`nick`) LIKE ?",
			wantArgs:  []any{"%ariel%", "%bar%"},
		},
		{
			input: Dialect(dialect.Postgres).
				Select().
				From(Table("users")).
				Where(And(ContainsFold("name", "Ariel"), ContainsFold("nick", "Bar"))),
			wantQuery: `SELECT * FROM "users" WHERE "name" ILIKE $1 AND "nick" ILIKE $2`,
			wantArgs:  []any{"%ariel%", "%bar%"},
		},
		{
			input: Dialect(dialect.MySQL).
				Select().
				From(Table("users")).
				Where(And(ContainsFold("name", "Ariel"), ContainsFold("nick", "Bar"))),
			wantQuery: "SELECT * FROM `users` WHERE `name` COLLATE utf8mb4_general_ci LIKE ? AND `nick` COLLATE utf8mb4_general_ci LIKE ?",
			wantArgs:  []any{"%ariel%", "%bar%"},
		},
		{
			input: func() Querier {
				s1 := Select().From(Table("users")).Where(Not(And(EQ("name", "foo"), EQ("age", "bar")))).As("users_view")
				return Select("name").From(s1)
			}(),
			wantQuery: "SELECT `name` FROM (SELECT * FROM `users` WHERE NOT (`name` = ? AND `age` = ?)) AS `users_view`",
			wantArgs:  []any{"foo", "bar"},
		},
		{
			input: func() Querier {
				d := Dialect(dialect.Postgres)
				s1 := d.Select().From(Table("users")).Where(Not(And(EQ("name", "foo"), EQ("age", "bar")))).As("users_view")
				return d.Select("name").From(s1)
			}(),
			wantQuery: `SELECT "name" FROM (SELECT * FROM "users" WHERE NOT ("name" = $1 AND "age" = $2)) AS "users_view"`,
			wantArgs:  []any{"foo", "bar"},
		},
		{
			input: func() Querier {
				t1 := Table("users")
				return Select().
					From(t1).
					Where(In(t1.C("id"), Select("owner_id").From(Table("pets")).Where(EQ("name", "pedro"))))
			}(),
			wantQuery: "SELECT * FROM `users` WHERE `users`.`id` IN (SELECT `owner_id` FROM `pets` WHERE `name` = ?)",
			wantArgs:  []any{"pedro"},
		},
		{
			input: func() Querier {
				t1 := Table("users")
				return Dialect(dialect.Postgres).
					Select().
					From(t1).
					Where(In(t1.C("id"), Select("owner_id").From(Table("pets")).Where(EQ("name", "pedro"))))
			}(),
			wantQuery: `SELECT * FROM "users" WHERE "users"."id" IN (SELECT "owner_id" FROM "pets" WHERE "name" = $1)`,
			wantArgs:  []any{"pedro"},
		},
		{
			input: func() Querier {
				t1 := Table("users")
				return Select().
					From(t1).
					Where(Not(In(t1.C("id"), Select("owner_id").From(Table("pets")).Where(EQ("name", "pedro")))))
			}(),
			wantQuery: "SELECT * FROM `users` WHERE NOT (`users`.`id` IN (SELECT `owner_id` FROM `pets` WHERE `name` = ?))",
			wantArgs:  []any{"pedro"},
		},
		{
			input: func() Querier {
				t1 := Table("users")
				return Dialect(dialect.Postgres).
					Select().
					From(t1).
					Where(Not(In(t1.C("id"), Select("owner_id").From(Table("pets")).Where(EQ("name", "pedro")))))
			}(),
			wantQuery: `SELECT * FROM "users" WHERE NOT ("users"."id" IN (SELECT "owner_id" FROM "pets" WHERE "name" = $1))`,
			wantArgs:  []any{"pedro"},
		},
		{
			input: Select("*").
				From(Table("users")).
				Limit(1),
			wantQuery: "SELECT * FROM `users` LIMIT 1",
		},
		{
			input: Dialect(dialect.Postgres).
				Select("*").
				From(Table("users")).
				Limit(1),
			wantQuery: `SELECT * FROM "users" LIMIT 1`,
		},
		{
			input:     Select("age").Distinct().From(Table("users")),
			wantQuery: "SELECT DISTINCT `age` FROM `users`",
		},
		{
			input: Dialect(dialect.Postgres).
				Select("age").
				Distinct().
				From(Table("users")),
			wantQuery: `SELECT DISTINCT "age" FROM "users"`,
		},
		{
			input:     Select("age", "name").From(Table("users")).Distinct().OrderBy("name"),
			wantQuery: "SELECT DISTINCT `age`, `name` FROM `users` ORDER BY `name`",
		},
		{
			input: Dialect(dialect.Postgres).
				Select("age", "name").
				From(Table("users")).
				Distinct().
				OrderBy("name"),
			wantQuery: `SELECT DISTINCT "age", "name" FROM "users" ORDER BY "name"`,
		},
		{
			input:     Select("age").From(Table("users")).Where(EQ("name", "foo")).Or().Where(EQ("name", "bar")),
			wantQuery: "SELECT `age` FROM `users` WHERE `name` = ? OR `name` = ?",
			wantArgs:  []any{"foo", "bar"},
		},
		{
			input: Dialect(dialect.Postgres).
				Select("age").
				From(Table("users")).
				Where(EQ("name", "foo")).Or().Where(EQ("name", "bar")),
			wantQuery: `SELECT "age" FROM "users" WHERE "name" = $1 OR "name" = $2`,
			wantArgs:  []any{"foo", "bar"},
		},
		{
			input:     SelectExpr(Raw("1")),
			wantQuery: "SELECT 1",
		},
		{
			input:     Select("*").From(SelectExpr(Raw("1")).As("s")),
			wantQuery: "SELECT * FROM (SELECT 1) AS `s`",
		},
		{
			input: func() Querier {
				t1 := Table("users")
				return Dialect(dialect.Postgres).
					Select().
					From(t1).
					Where(CompositeGT(t1.Columns("id", "name"), 1, "Ariel"))
			}(),
			wantQuery: `SELECT * FROM "users" WHERE ("users"."id", "users"."name") > ($1, $2)`,
			wantArgs:  []any{1, "Ariel"},
		},
		{
			input: func() Querier {
				t1 := Table("users")
				return Dialect(dialect.Postgres).
					Select().
					From(t1).
					Where(And(EQ("name", "Ariel"), CompositeGT(t1.Columns("id", "name"), 1, "Ariel")))
			}(),
			wantQuery: `SELECT * FROM "users" WHERE "name" = $1 AND ("users"."id", "users"."name") > ($2, $3)`,
			wantArgs:  []any{"Ariel", 1, "Ariel"},
		},
		{
			input: func() Querier {
				t1 := Table("users")
				return Dialect(dialect.Postgres).
					Select().
					From(t1).
					Where(And(EQ("name", "Ariel"), Or(EQ("surname", "Doe"), CompositeGT(t1.Columns("id", "name"), 1, "Ariel"))))
			}(),
			wantQuery: `SELECT * FROM "users" WHERE "name" = $1 AND ("surname" = $2 OR ("users"."id", "users"."name") > ($3, $4))`,
			wantArgs:  []any{"Ariel", "Doe", 1, "Ariel"},
		},
		{
			input: func() Querier {
				t1 := Table("users")
				return Dialect(dialect.Postgres).
					Select().
					From(Table("users")).
					Where(And(EQ("name", "Ariel"), CompositeLT(t1.Columns("id", "name"), 1, "Ariel")))
			}(),
			wantQuery: `SELECT * FROM "users" WHERE "name" = $1 AND ("users"."id", "users"."name") < ($2, $3)`,
			wantArgs:  []any{"Ariel", 1, "Ariel"},
		},
		{
			input: Select().
				From(Table("pragma_table_info('t1')").Unquote()).
				OrderBy("pk"),
			wantQuery: "SELECT * FROM pragma_table_info('t1') ORDER BY `pk`",
		},
		{
			input: Dialect(dialect.Postgres).
				Select("*").
				From(Table("users")).
				Where(Or(
					And(EQ("id", 1), InInts("group_id", 2, 3)),
					And(EQ("id", 2), InValues("group_id", 4, 5)),
				)).
				Where(And(
					Or(EQ("a", "a"), And(EQ("b", "b"), EQ("c", "c"))),
					Not(Or(IsNull("d"), NotNull("e"))),
				)).
				Or().
				Where(And(NEQ("f", "f"), NEQ("g", "g"))),
			wantQuery: strings.NewReplacer("\n", "", "\t", "").Replace(`
			SELECT * FROM "users"
 WHERE
	 (
		(("id" = $1 AND "group_id" IN ($2, $3)) OR ("id" = $4 AND "group_id" IN ($5, $6)))
		 AND
		 (("a" = $7 OR ("b" = $8 AND "c" = $9)) AND (NOT ("d" IS NULL OR "e" IS NOT NULL)))
	)
	 OR ("f" <> $10 AND "g" <> $11)`),
			wantArgs: []any{1, 2, 3, 2, 4, 5, "a", "b", "c", "f", "g"},
		},
		{
			input: Dialect(dialect.Postgres).
				Select("*").
				From(Table("test")).
				Where(P(func(b *Builder) {
					b.WriteString("nlevel(").Ident("path").WriteByte(')').WriteOp(OpGT).Arg(1)
				})),
			wantQuery: `SELECT * FROM "test" WHERE nlevel("path") > $1`,
			wantArgs:  []any{1},
		},
		{
			input: Dialect(dialect.Postgres).
				Select("*").
				From(Table("test")).
				Where(P(func(b *Builder) {
					b.WriteString("nlevel(").Ident("path").WriteByte(')').WriteOp(OpGT).Arg(1)
				})),
			wantQuery: `SELECT * FROM "test" WHERE nlevel("path") > $1`,
			wantArgs:  []any{1},
		},
		{
			input:     Select("id").From(Table("users")).Where(ExprP("DATE(last_login_at) >= ?", "2022-05-03")),
			wantQuery: "SELECT `id` FROM `users` WHERE DATE(last_login_at) >= ?",
			wantArgs:  []any{"2022-05-03"},
		},
		{
			input: Select("id").
				From(Table("users")).
				Where(P(func(b *Builder) {
					b.WriteString("DATE(").Ident("last_login_at").WriteString(") >= ").Arg("2022-05-03")
				})),
			wantQuery: "SELECT `id` FROM `users` WHERE DATE(`last_login_at`) >= ?",
			wantArgs:  []any{"2022-05-03"},
		},
		{
			input:     Select("id").From(Table("events")).Where(ExprP("DATE_ADD(date, INTERVAL duration MINUTE) BETWEEN ? AND ?", "2022-05-03", "2022-05-04")),
			wantQuery: "SELECT `id` FROM `events` WHERE DATE_ADD(date, INTERVAL duration MINUTE) BETWEEN ? AND ?",
			wantArgs:  []any{"2022-05-03", "2022-05-04"},
		},
		{
			input: Select("id").
				From(Table("events")).
				Where(P(func(b *Builder) {
					b.WriteString("DATE_ADD(date, INTERVAL duration MINUTE) BETWEEN ").Arg("2022-05-03").WriteString(" AND ").Arg("2022-05-04")
				})),
			wantQuery: "SELECT `id` FROM `events` WHERE DATE_ADD(date, INTERVAL duration MINUTE) BETWEEN ? AND ?",
			wantArgs:  []any{"2022-05-03", "2022-05-04"},
		},
		{
			input: Dialect(dialect.Postgres).
				Select("*").
				From(Table("users")).
				Where(ExprP("name = $1", "pedro")).
				Where(P(func(b *Builder) {
					b.Join(Expr("name = $2", "pedro"))
				})).
				Where(EQ("name", "pedro")).
				Where(
					And(
						In(
							"id",
							Select("owner_id").
								From(Table("pets")).
								Where(EQ("name", "luna")),
						),
						EQ("active", true),
					),
				),
			wantQuery: `SELECT * FROM "users" WHERE ((name = $1 AND name = $2) AND "name" = $3) AND ("id" IN (SELECT "owner_id" FROM "pets" WHERE "name" = $4) AND "active")`,
			wantArgs:  []any{"pedro", "pedro", "pedro", "luna"},
		},
		{
			input: func() Querier {
				t1 := Table("users")
				return Dialect(dialect.Postgres).
					Select().
					From(t1).
					Where(ColumnsEQ(t1.C("id1"), t1.C("id2"))).
					Where(ColumnsNEQ(t1.C("id1"), t1.C("id2"))).
					Where(ColumnsGT(t1.C("id1"), t1.C("id2"))).
					Where(ColumnsGTE(t1.C("id1"), t1.C("id2"))).
					Where(ColumnsLT(t1.C("id1"), t1.C("id2"))).
					Where(ColumnsLTE(t1.C("id1"), t1.C("id2")))
			}(),
			wantQuery: strings.ReplaceAll(`
SELECT * FROM "users" 
WHERE (((("users"."id1" = "users"."id2" AND "users"."id1" <> "users"."id2") 
AND "users"."id1" > "users"."id2") AND "users"."id1" >= "users"."id2") 
AND "users"."id1" < "users"."id2") AND "users"."id1" <= "users"."id2"`, "\n", ""),
		},
		{
			input: Select("name").
				From(Select("name", "age").From(Table("users"))),
			wantQuery: "SELECT `name` FROM (SELECT `name`, `age` FROM `users`)",
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

func TestSelector_Join(t *testing.T) {
	tests := []struct {
		input     Querier
		wantQuery string
		wantArgs  []any
	}{
		{
			input: func() Querier {
				t1 := Table("users").As("u")
				t2 := Table("groups").As("g")
				return Select(t1.C("id"), t2.C("name")).From(t1).Join(t2)
			}(),
			wantQuery: "SELECT `u`.`id`, `g`.`name` FROM `users` AS `u` JOIN `groups` AS `g`",
		},
		{
			input: func() Querier {
				t1 := Table("users").As("u")
				t2 := Table("groups").As("g")
				return Select(t1.C("id"), t2.C("name")).
					From(t1).
					Join(t2).
					On(t1.C("id"), t2.C("user_id"))
			}(),
			wantQuery: "SELECT `u`.`id`, `g`.`name` FROM `users` AS `u` JOIN `groups` AS `g` ON `u`.`id` = `g`.`user_id`",
		},
		{
			input: func() Querier {
				t1 := Table("users").As("u")
				t2 := Table("groups").As("g")
				return Dialect(dialect.Postgres).
					Select(t1.C("id"), t2.C("name")).
					From(t1).
					Join(t2).
					On(t1.C("id"), t2.C("user_id"))
			}(),
			wantQuery: `SELECT "u"."id", "g"."name" FROM "users" AS "u" JOIN "groups" AS "g" ON "u"."id" = "g"."user_id"`,
		},
		{
			input: func() Querier {
				t1 := Table("users").As("u")
				t2 := Table("groups").As("g")
				return Select(t1.C("id"), t2.C("name")).
					From(t1).
					Join(t2).
					On(t1.C("id"), t2.C("user_id")).
					Where(And(EQ(t1.C("name"), "bar"), NotNull(t2.C("name"))))
			}(),
			wantQuery: "SELECT `u`.`id`, `g`.`name` FROM `users` AS `u` JOIN `groups` AS `g` ON `u`.`id` = `g`.`user_id` WHERE `u`.`name` = ? AND `g`.`name` IS NOT NULL",
			wantArgs:  []any{"bar"},
		},
		{
			input: func() Querier {
				t1 := Table("users").As("u")
				t2 := Table("groups").As("g")
				return Dialect(dialect.Postgres).
					Select(t1.C("id"), t2.C("name")).
					From(t1).
					Join(t2).
					On(t1.C("id"), t2.C("user_id")).
					Where(And(EQ(t1.C("name"), "bar"), NotNull(t2.C("name"))))
			}(),
			wantQuery: `SELECT "u"."id", "g"."name" FROM "users" AS "u" JOIN "groups" AS "g" ON "u"."id" = "g"."user_id" WHERE "u"."name" = $1 AND "g"."name" IS NOT NULL`,
			wantArgs:  []any{"bar"},
		},
		{
			input: func() Querier {
				t1 := Table("users").As("u")
				t2 := Select().From(Table("groups")).Where(EQ("user_id", 10)).As("g")
				return Select(t1.C("id"), t2.C("name")).
					From(t1).
					Join(t2).
					On(t1.C("id"), t2.C("user_id"))
			}(),
			wantQuery: "SELECT `u`.`id`, `g`.`name` FROM `users` AS `u` JOIN (SELECT * FROM `groups` WHERE `user_id` = ?) AS `g` ON `u`.`id` = `g`.`user_id`",
			wantArgs:  []any{10},
		},
		{
			input: func() Querier {
				d := Dialect(dialect.Postgres)
				t1 := d.Table("users").As("u")
				t2 := d.Select().From(Table("groups")).Where(EQ("user_id", 10)).As("g")
				return d.Select(t1.C("id"), t2.C("name")).
					From(t1).
					Join(t2).
					On(t1.C("id"), t2.C("user_id"))
			}(),
			wantQuery: `SELECT "u"."id", "g"."name" FROM "users" AS "u" JOIN (SELECT * FROM "groups" WHERE "user_id" = $1) AS "g" ON "u"."id" = "g"."user_id"`,
			wantArgs:  []any{10},
		},
		{
			input: func() Querier {
				t1 := Table("users")
				t2 := Table("groups")
				t3 := Table("user_groups")
				return Select(t1.C("*")).From(t1).
					Join(t3).On(t1.C("id"), t3.C("user_id")).
					Join(t2).On(t2.C("id"), t3.C("group_id"))
			}(),
			wantQuery: "SELECT `users`.* FROM `users` JOIN `user_groups` AS `t1` ON `users`.`id` = `t1`.`user_id` JOIN `groups` AS `t2` ON `t2`.`id` = `t1`.`group_id`",
		},
		{
			input: func() Querier {
				d := Dialect(dialect.Postgres)
				t1 := d.Table("users")
				t2 := d.Table("groups")
				t3 := d.Table("user_groups")
				return d.Select(t1.C("*")).From(t1).
					Join(t3).On(t1.C("id"), t3.C("user_id")).
					Join(t2).On(t2.C("id"), t3.C("group_id"))
			}(),
			wantQuery: `SELECT "users".* FROM "users" JOIN "user_groups" AS "t1" ON "users"."id" = "t1"."user_id" JOIN "groups" AS "t2" ON "t2"."id" = "t1"."group_id"`,
		},
		{
			input: func() Querier {
				builder := Dialect(dialect.Postgres)
				t1 := builder.Table("groups")
				t2 := builder.Table("users")
				t3 := builder.Table("user_groups")
				t4 := builder.Select(t3.C("id")).
					From(t3).
					Join(t2).
					On(t3.C("id"), t2.C("id2")).
					Where(EQ(t2.C("id"), "baz"))
				return builder.Select().
					From(t1).
					Join(t4).
					On(t1.C("id"), t4.C("id")).Limit(1)
			}(),
			wantQuery: `SELECT * FROM "groups" JOIN (SELECT "user_groups"."id" FROM "user_groups" JOIN "users" AS "t1" ON "user_groups"."id" = "t1"."id2" WHERE "t1"."id" = $1) AS "t1" ON "groups"."id" = "t1"."id" LIMIT 1`,
			wantArgs:  []any{"baz"},
		},
		{
			input: func() Querier {
				t1, t2 := Table("users").Schema("s1"), Table("pets").Schema("s2")
				return Select("*").
					From(t1).Join(t2).
					OnP(P(func(b *Builder) {
						b.Ident(t1.C("id")).WriteOp(OpEQ).Ident(t2.C("owner_id"))
					})).
					Where(EQ(t2.C("name"), "pedro"))
			}(),
			wantQuery: "SELECT * FROM `s1`.`users` JOIN `s2`.`pets` AS `t1` ON `s1`.`users`.`id` = `t1`.`owner_id` WHERE `t1`.`name` = ?",
			wantArgs:  []any{"pedro"},
		},
		{
			input: func() Querier {
				t1, t2 := Table("users").Schema("s1"), Table("pets").Schema("s2")
				sel := Select("*").
					From(t1).Join(t2).
					OnP(P(func(b *Builder) {
						b.Ident(t1.C("id")).WriteOp(OpEQ).Ident(t2.C("owner_id"))
					})).
					Where(EQ(t2.C("name"), "pedro"))
				sel.SetDialect(dialect.SQLite)
				return sel
			}(),
			wantQuery: "SELECT * FROM `users` JOIN `pets` AS `t1` ON `users`.`id` = `t1`.`owner_id` WHERE `t1`.`name` = ?",
			wantArgs:  []any{"pedro"},
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

func TestSelector_Aggregations(t *testing.T) {
	tests := []struct {
		input     Querier
		wantQuery string
		wantArgs  []any
	}{
		{
			input:     Select().Count().From(Table("users")),
			wantQuery: "SELECT COUNT(*) FROM `users`",
		},
		{
			input: Dialect(dialect.Postgres).
				Select().Count().From(Table("users")),
			wantQuery: `SELECT COUNT(*) FROM "users"`,
		},
		{
			input:     Select().Count(Distinct("id")).From(Table("users")),
			wantQuery: "SELECT COUNT(DISTINCT `id`) FROM `users`",
		},
		{
			input: Dialect(dialect.Postgres).
				Select().Count(Distinct("id")).From(Table("users")),
			wantQuery: `SELECT COUNT(DISTINCT "id") FROM "users"`,
		},
		{
			input: func() Querier {
				t1 := Table("users")
				t2 := Select().From(Table("groups"))
				t3 := Select().Count().From(t1).Join(t1).On(t2.C("id"), t1.C("blocked_id"))
				return t3.Count(Distinct(t3.Columns("id", "name")...))
			}(),
			wantQuery: "SELECT COUNT(DISTINCT `t1`.`id`, `t1`.`name`) FROM `users` AS `t1` JOIN `users` AS `t1` ON `groups`.`id` = `t1`.`blocked_id`",
		},
		{
			input: func() Querier {
				d := Dialect(dialect.Postgres)
				t1 := d.Table("users")
				t2 := d.Select().From(Table("groups"))
				t3 := d.Select().Count().From(t1).Join(t1).On(t2.C("id"), t1.C("blocked_id"))
				return t3.Count(Distinct(t3.Columns("id", "name")...))
			}(),
			wantQuery: `SELECT COUNT(DISTINCT "t1"."id", "t1"."name") FROM "users" AS "t1" JOIN "users" AS "t1" ON "groups"."id" = "t1"."blocked_id"`,
		},
		{
			input:     Select(Sum("age"), Min("age")).From(Table("users")),
			wantQuery: "SELECT SUM(`age`), MIN(`age`) FROM `users`",
		},
		{
			input: Dialect(dialect.Postgres).
				Select(Sum("age"), Min("age")).
				From(Table("users")),
			wantQuery: `SELECT SUM("age"), MIN("age") FROM "users"`,
		},
		{
			input: func() Querier {
				t1 := Table("users").As("u")
				return Select(As(Max(t1.C("age")), "max_age")).From(t1)
			}(),
			wantQuery: "SELECT MAX(`u`.`age`) AS `max_age` FROM `users` AS `u`",
		},
		{
			input: func() Querier {
				t1 := Table("users").As("u")
				return Dialect(dialect.Postgres).
					Select(As(Max(t1.C("age")), "max_age")).
					From(t1)
			}(),
			wantQuery: `SELECT MAX("u"."age") AS "max_age" FROM "users" AS "u"`,
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

func TestSelector_OrderByExpr(t *testing.T) {
	query, args := Select("*").
		From(Table("users")).
		Where(GT("age", 28)).
		OrderBy("name").
		OrderExpr(Expr("CASE WHEN id=? THEN id WHEN id=? THEN name END DESC", 1, 2)).
		Query()
	require.Equal(t, "SELECT * FROM `users` WHERE `age` > ? ORDER BY `name`, CASE WHEN id=? THEN id WHEN id=? THEN name END DESC", query)
	require.Equal(t, []any{28, 1, 2}, args)

	query, args = Dialect(dialect.Postgres).
		Select("*").
		From(Table("users")).
		Where(GT("age", 28)).
		OrderBy("name").
		OrderExpr(ExprFunc(func(b *Builder) {
			b.WriteString("CASE")
			b.WriteString(" WHEN ").Ident("id").WriteOp(OpEQ).Arg(1).WriteString(" THEN ").Ident("id")
			b.WriteString(" WHEN ").Ident("id").WriteOp(OpEQ).Arg(2).WriteString(" THEN ").Ident("name")
			b.WriteString(" END DESC")
		})).
		Query()
	require.Equal(t, `SELECT * FROM "users" WHERE "age" > $1 ORDER BY "name", CASE WHEN "id" = $2 THEN "id" WHEN "id" = $3 THEN "name" END DESC`, query)
	require.Equal(t, []any{28, 1, 2}, args)
}

func TestSelector_GroupBy(t *testing.T) {
	tests := []struct {
		input     Querier
		wantQuery string
		wantArgs  []any
	}{
		{
			input: Select("name", Count("*")).
				From(Table("users")).
				GroupBy("name"),
			wantQuery: "SELECT `name`, COUNT(*) FROM `users` GROUP BY `name`",
		},
		{
			input: Dialect(dialect.Postgres).
				Select("name", Count("*")).
				From(Table("users")).
				GroupBy("name"),
			wantQuery: `SELECT "name", COUNT(*) FROM "users" GROUP BY "name"`,
		},
		{
			input: Select("name", Count("*")).
				From(Table("users")).
				GroupBy("name").
				OrderBy("name"),
			wantQuery: "SELECT `name`, COUNT(*) FROM `users` GROUP BY `name` ORDER BY `name`",
		},
		{
			input: Dialect(dialect.Postgres).
				Select("name", Count("*")).
				From(Table("users")).
				GroupBy("name").
				OrderBy("name"),
			wantQuery: `SELECT "name", COUNT(*) FROM "users" GROUP BY "name" ORDER BY "name"`,
		},
		{
			input: Select("name", "age", Count("*")).
				From(Table("users")).
				GroupBy("name", "age").
				OrderBy(Desc("name"), "age"),
			wantQuery: "SELECT `name`, `age`, COUNT(*) FROM `users` GROUP BY `name`, `age` ORDER BY `name` DESC, `age`",
		},
		{
			input: Dialect(dialect.Postgres).
				Select("name", "age", Count("*")).
				From(Table("users")).
				GroupBy("name", "age").
				OrderBy(Desc("name"), "age"),
			wantQuery: `SELECT "name", "age", COUNT(*) FROM "users" GROUP BY "name", "age" ORDER BY "name" DESC, "age"`,
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

func TestSelector_JoinGroupBy(t *testing.T) {
	tests := []struct {
		input     Querier
		wantQuery string
		wantArgs  []any
	}{
		{
			input: func() Querier {
				t1 := Table("users").As("u")
				t2 := Table("user_groups").As("ug")
				return Select(t1.C("id"), As(Count("`*`"), "group_count")).
					From(t1).
					LeftJoin(t2).
					On(t1.C("id"), t2.C("user_id")).
					GroupBy(t1.C("id"))
			}(),
			wantQuery: "SELECT `u`.`id`, COUNT(`*`) AS `group_count` FROM `users` AS `u` LEFT JOIN `user_groups` AS `ug` ON `u`.`id` = `ug`.`user_id` GROUP BY `u`.`id`",
		},
		{
			input: func() Querier {
				t1 := Table("users").As("u")
				t2 := Table("user_groups").As("ug")
				return Select(t1.C("id"), As(Count("`*`"), "group_count")).
					From(t1).
					LeftJoin(t2).
					OnP(P(func(b *Builder) {
						b.Ident(t1.C("id")).WriteOp(OpEQ).Ident(t2.C("user_id"))
					})).
					GroupBy(t1.C("id")).Clone()
			}(),
			wantQuery: "SELECT `u`.`id`, COUNT(`*`) AS `group_count` FROM `users` AS `u` LEFT JOIN `user_groups` AS `ug` ON `u`.`id` = `ug`.`user_id` GROUP BY `u`.`id`",
		},
		{
			input: func() Querier {
				t1 := Table("groups").As("g")
				t2 := Table("user_groups").As("ug")
				return Select(t1.C("id"), As(Count("`*`"), "user_count")).
					From(t1).
					RightJoin(t2).
					On(t1.C("id"), t2.C("group_id")).
					GroupBy(t1.C("id"))
			}(),
			wantQuery: "SELECT `g`.`id`, COUNT(`*`) AS `user_count` FROM `groups` AS `g` RIGHT JOIN `user_groups` AS `ug` ON `g`.`id` = `ug`.`group_id` GROUP BY `g`.`id`",
		},
		{
			input: func() Querier {
				t1 := Table("groups").As("g")
				t2 := Table("user_groups").As("ug")
				return Select(t1.C("id"), As(Count("`*`"), "user_count")).
					From(t1).
					FullJoin(t2).
					On(t1.C("id"), t2.C("group_id")).
					GroupBy(t1.C("id"))
			}(),
			wantQuery: "SELECT `g`.`id`, COUNT(`*`) AS `user_count` FROM `groups` AS `g` FULL JOIN `user_groups` AS `ug` ON `g`.`id` = `ug`.`group_id` GROUP BY `g`.`id`",
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

func TestSelector_ClearOrder(t *testing.T) {
	query, args := Select("*").
		From(Table("users")).
		OrderBy("name").
		ClearOrder().
		OrderBy("id").
		Query()
	require.Equal(t, "SELECT * FROM `users` ORDER BY `id`", query)
	require.Empty(t, args)
}

func TestSelector_SelectExpr(t *testing.T) {
	query, args := SelectExpr(
		Expr("?", "a"),
		ExprFunc(func(b *Builder) {
			b.Ident("first_name").WriteOp(OpAdd).Ident("last_name")
		}),
		ExprFunc(func(b *Builder) {
			b.WriteString("COALESCE(").Ident("age").Comma().Arg(0).WriteByte(')')
		}),
		Expr("?", "b"),
	).From(Table("users")).Query()
	require.Equal(t, "SELECT ?, `first_name` + `last_name`, COALESCE(`age`, ?), ? FROM `users`", query)
	require.Equal(t, []any{"a", 0, "b"}, args)

	query, args = Dialect(dialect.Postgres).
		Select("name").
		AppendSelectExpr(
			Expr("age + $1", 1),
			ExprFunc(func(b *Builder) {
				b.Wrap(func(b *Builder) {
					b.WriteString("similarity(").Ident("name").Comma().Arg("A").WriteByte(')')
					b.WriteOp(OpAdd)
					b.WriteString("similarity(").Ident("desc").Comma().Arg("D").WriteByte(')')
				})
				b.WriteString(" AS s")
			}),
			Expr("rank + $4", 10),
		).
		From(Table("users")).
		Query()
	require.Equal(t, `SELECT "name", age + $1, (similarity("name", $2) + similarity("desc", $3)) AS s, rank + $4 FROM "users"`, query)
	require.Equal(t, []any{1, "A", "D", 10}, args)
}

func TestSelector_Union(t *testing.T) {
	query, args := Dialect(dialect.Postgres).
		Select("*").
		From(Table("users")).
		Where(EQ("active", true)).
		Union(
			Select("*").
				From(Table("old_users1")).
				Where(
					And(
						EQ("is_active", true),
						GT("age", 20),
					),
				),
		).
		UnionAll(
			Select("*").
				From(Table("old_users2")).
				Where(
					And(
						EQ("is_active", "true"),
						LT("age", 18),
					),
				),
		).
		Query()
	require.Equal(t, `SELECT * FROM "users" WHERE "active" UNION SELECT * FROM "old_users1" WHERE "is_active" AND "age" > $1 UNION ALL SELECT * FROM "old_users2" WHERE "is_active" = $2 AND "age" < $3`, query)
	require.Equal(t, []any{20, "true", 18}, args)
}

func TestSelector_Except(t *testing.T) {
	query, args := Dialect(dialect.Postgres).
		Select("*").
		From(Table("users")).
		Where(EQ("active", true)).
		Except(
			Select("*").
				From(Table("old_users1")).
				Where(
					And(
						EQ("is_active", true),
						GT("age", 20),
					),
				),
		).
		ExceptAll(
			Select("*").
				From(Table("old_users2")).
				Where(
					And(
						EQ("is_active", "true"),
						LT("age", 18),
					),
				),
		).
		Query()
	require.Equal(t, `SELECT * FROM "users" WHERE "active" EXCEPT SELECT * FROM "old_users1" WHERE "is_active" AND "age" > $1 EXCEPT ALL SELECT * FROM "old_users2" WHERE "is_active" = $2 AND "age" < $3`, query)
	require.Equal(t, []any{20, "true", 18}, args)
}

func TestSelector_Intersect(t *testing.T) {
	query, args := Dialect(dialect.Postgres).
		Select("*").
		From(Table("users")).
		Where(EQ("active", true)).
		Intersect(
			Select("*").
				From(Table("old_users1")).
				Where(
					And(
						EQ("is_active", true),
						GT("age", 20),
					),
				),
		).
		IntersectAll(
			Select("*").
				From(Table("old_users2")).
				Where(
					And(
						EQ("is_active", "true"),
						LT("age", 18),
					),
				),
		).
		Query()
	require.Equal(t, `SELECT * FROM "users" WHERE "active" INTERSECT SELECT * FROM "old_users1" WHERE "is_active" AND "age" > $1 INTERSECT ALL SELECT * FROM "old_users2" WHERE "is_active" = $2 AND "age" < $3`, query)
	require.Equal(t, []any{20, "true", 18}, args)
}

func TestSelector_SetOperatorWithRecursive(t *testing.T) {
	t1, t2, t3 := Table("files"), Table("files"), Table("path")
	n := Queries{
		WithRecursive("path", "id", "name", "parent_id").
			As(Select(t1.Columns("id", "name", "parent_id")...).
				From(t1).
				Where(
					And(
						IsNull(t1.C("parent_id")),
						EQ(t1.C("deleted"), false),
					),
				).
				UnionAll(
					Select(t2.Columns("id", "name", "parent_id")...).
						From(t2).
						Join(t3).
						On(t2.C("parent_id"), t3.C("id")).
						Where(
							EQ(t2.C("deleted"), false),
						),
				),
			),
		Select(t3.Columns("id", "name", "parent_id")...).
			From(t3),
	}
	query, args := n.Query()
	require.Equal(t, "WITH RECURSIVE `path`(`id`, `name`, `parent_id`) AS (SELECT `files`.`id`, `files`.`name`, `files`.`parent_id` FROM `files` WHERE `files`.`parent_id` IS NULL AND NOT `files`.`deleted` UNION ALL SELECT `files`.`id`, `files`.`name`, `files`.`parent_id` FROM `files` JOIN `path` AS `t1` ON `files`.`parent_id` = `t1`.`id` WHERE NOT `files`.`deleted`) SELECT `t1`.`id`, `t1`.`name`, `t1`.`parent_id` FROM `path` AS `t1`", query)
	require.Nil(t, args)
}

func TestSelectWithLock(t *testing.T) {
	query, args := Dialect(dialect.MySQL).
		Select().
		From(Table("users")).
		Where(EQ("id", 1)).
		ForUpdate().
		Query()
	require.Equal(t, "SELECT * FROM `users` WHERE `id` = ? FOR UPDATE", query)
	require.Equal(t, 1, args[0])

	query, args = Dialect(dialect.Postgres).
		Select().
		From(Table("users")).
		Where(EQ("id", 1)).
		ForUpdate(WithLockAction(NoWait)).
		Query()
	require.Equal(t, `SELECT * FROM "users" WHERE "id" = $1 FOR UPDATE NOWAIT`, query)
	require.Equal(t, 1, args[0])

	users, pets := Table("users"), Table("pets")
	query, args = Dialect(dialect.Postgres).
		Select().
		From(pets).
		Join(users).
		On(pets.C("owner_id"), users.C("id")).
		Where(EQ("id", 20)).
		ForUpdate(
			WithLockAction(SkipLocked),
			WithLockTables("pets"),
		).
		Query()
	require.Equal(t, `SELECT * FROM "pets" JOIN "users" AS "t1" ON "pets"."owner_id" = "t1"."id" WHERE "id" = $1 FOR UPDATE OF "pets" SKIP LOCKED`, query)
	require.Equal(t, 20, args[0])

	query, args = Dialect(dialect.MySQL).
		Select().
		From(Table("users")).
		Where(EQ("id", 20)).
		ForShare(WithLockClause("LOCK IN SHARE MODE")).
		Query()
	require.Equal(t, "SELECT * FROM `users` WHERE `id` = ? LOCK IN SHARE MODE", query)
	require.Equal(t, 20, args[0])

	s := Dialect(dialect.SQLite).
		Select().
		From(Table("users")).
		Where(EQ("id", 1)).
		ForUpdate()
	s.Query()
	require.EqualError(t, s.Err(), "sql: SELECT .. FOR UPDATE/SHARE not supported in SQLite")
}

func TestSelector_UnionOrderBy(t *testing.T) {
	table := Table("users")
	query, _ := Dialect(dialect.Postgres).
		Select("*").
		From(table).
		Where(EQ("active", true)).
		Union(Select("*").From(Table("old_users1"))).
		OrderBy(table.C("whatever")).
		Query()
	require.Equal(t, `SELECT * FROM "users" WHERE "active" UNION SELECT * FROM "old_users1" ORDER BY "users"."whatever"`, query)
}

func TestSelector_UnqualifiedColumns(t *testing.T) {
	t1, t2 := Table("t1"), Table("t2")
	s := Select(t1.C("a"), t2.C("b"))
	require.Equal(t, []string{"`t1`.`a`", "`t2`.`b`"}, s.SelectedColumns())
	require.Equal(t, []string{"a", "b"}, s.UnqualifiedColumns())

	d := Dialect(dialect.Postgres)
	t1, t2 = d.Table("t1"), d.Table("t2")
	s = d.Select(t1.C("a"), t2.C("b"))
	require.Equal(t, []string{`"t1"."a"`, `"t2"."b"`}, s.SelectedColumns())
	require.Equal(t, []string{"a", "b"}, s.UnqualifiedColumns())
}

func TestMultipleFrom(t *testing.T) {
	query, args := Dialect(dialect.Postgres).
		Select("items.*", As("ts_rank_cd(search, search_query)", "rank")).
		From(Table("items")).
		AppendFrom(Table("to_tsquery('neutrino|(dark & matter)')").As("search_query")).
		Where(P(func(b *Builder) {
			b.WriteString("search @@ search_query")
		})).
		OrderBy(Desc("rank")).
		Query()
	require.Empty(t, args)
	require.Equal(t, `SELECT items.*, ts_rank_cd(search, search_query) AS "rank" FROM "items", to_tsquery('neutrino|(dark & matter)') AS "search_query" WHERE search @@ search_query ORDER BY "rank" DESC`, query)

	query, args = Dialect(dialect.Postgres).
		Select("items.*", As("ts_rank_cd(search, search_query)", "rank")).
		From(Table("items")).
		AppendFromExpr(Expr("to_tsquery($1) AS search_query", "neutrino|(dark & matter)")).
		Where(P(func(b *Builder) {
			b.WriteString("search @@ search_query")
		})).
		Query()
	require.Equal(t, []any{"neutrino|(dark & matter)"}, args)
	require.Equal(t, `SELECT items.*, ts_rank_cd(search, search_query) AS "rank" FROM "items", to_tsquery($1) AS search_query WHERE search @@ search_query`, query)

	query, args = Dialect(dialect.Postgres).
		Select("items.*", As("ts_rank_cd(search, search_query)", "rank")).
		From(Table("items")).
		Where(EQ("value", 10)).
		AppendFromExpr(ExprFunc(func(b *Builder) {
			b.WriteString("to_tsquery(").Arg("neutrino|(dark & matter)").WriteString(") AS search_query")
		})).
		Where(P(func(b *Builder) {
			b.WriteString("search @@ search_query")
		})).
		Query()
	require.Equal(t, []any{"neutrino|(dark & matter)", 10}, args)
	require.Equal(t, `SELECT items.*, ts_rank_cd(search, search_query) AS "rank" FROM "items", to_tsquery($1) AS search_query WHERE "value" = $2 AND search @@ search_query`, query)
}

func TestSelector_HasJoins(t *testing.T) {
	s := Select("*").From(Table("t1"))
	require.False(t, s.HasJoins())
	s.Join(Table("t2"))
	require.True(t, s.HasJoins())
}

func TestSelector_JoinedTable(t *testing.T) {
	s := Select("*").From(Table("t1"))
	t2, ok := s.JoinedTable("t2")
	require.False(t, ok)
	require.Nil(t, t2)
	s.Join(Table("t2").As("t2"))
	t2, ok = s.JoinedTable("t2")
	require.True(t, ok)
	require.Equal(t, "`t2`.`c`", t2.C("c"))
	s.LeftJoin(Select().From(Table("t3").As("t3")).Where(EQ("id", 1)))
	t3, ok := s.JoinedTable("t3")
	require.True(t, ok)
	require.Equal(t, "`t3`.`c`", t3.C("c"))
}

func TestSelector_JoinedTableView(t *testing.T) {
	s := Select("*").From(Table("t1"))
	t2, ok := s.JoinedTableView("t2")
	require.False(t, ok)
	require.Nil(t, t2)
	s.Join(Table("users").As("t2"))
	t2, ok = s.JoinedTableView("t2")
	require.True(t, ok)
	require.Equal(t, "`t2`.`c`", t2.C("c"))
	s.LeftJoin(Select().From(Table("pets").As("t3")).Where(EQ("id", 1)).As("t4"))
	t3, ok := s.JoinedTableView("t3")
	require.True(t, ok)
	require.Equal(t, "`t3`.`c`", t3.C("c"))
	t4, ok := s.JoinedTableView("t4")
	require.True(t, ok)
	require.Equal(t, "`t4`.`c`", t4.C("c"))
}

func TestSelector_Columns(t *testing.T) {
	t.Run("MySQL", func(t *testing.T) {
		s := Select("*").From(Table("users"))
		require.Equal(t, []string{"`users`.`c`"}, s.Columns("c"))
		// Already quoted.
		require.Equal(t, []string{"`users`.`c`"}, s.Columns("`c`"))
		t2 := Table("t2").As("t2")
		s.Join(t2)
		// Already quoted.
		require.Equal(t, []string{"`t2`.`c1`"}, s.Columns(t2.C("c1")))
		require.Equal(t, []string{"t2.c1"}, s.Columns("t2.c1"))
	})
	t.Run("Postgres", func(t *testing.T) {
		b := Dialect(dialect.Postgres)
		s := b.Select("*").From(Table("users"))
		require.Equal(t, []string{`"users"."c"`}, s.Columns("c"))
		// Already quoted.
		require.Equal(t, []string{`"users"."c"`}, s.Columns(`"c"`))
		t2 := b.Table("t2").As("t2")
		s.Join(t2)
		// Already quoted.
		require.Equal(t, []string{`"t2"."c1"`}, s.Columns(t2.C("c1")))
		require.Equal(t, []string{"t2.c1"}, s.Columns("t2.c1"))
	})
}

func TestSelector_SelectedColumn(t *testing.T) {
	t.Run("MySQL", func(t *testing.T) {
		s := Select("*").From(Table("t1"))
		require.Empty(t, s.FindSelection("c"))
		s.Select("c")
		require.Equal(t, []string{"c"}, s.FindSelection("c"))
		s.Select(s.C("c"))
		require.Equal(t, []string{"`t1`.`c`"}, s.FindSelection("c"))
		s.AppendSelectAs(s.C("d"), "e")
		require.Equal(t, []string{"e"}, s.FindSelection("e"))
		require.Empty(t, s.FindSelection("d"))
		t2 := Table("t2").As("t2")
		s.Join(t2)
		s.Select(t2.C("e"), "t2.e", s.C("e"), "t1.e", "e")
		require.Equal(t, []string{"`t2`.`e`", "t2.e", "`t1`.`e`", "t1.e", "e"}, s.FindSelection("e"))
		s.AppendSelectExprAs(ExprFunc(func(b *Builder) {
			b.S("COUNT(").Ident("post_id").S(")")
		}), "post_count")
		require.Equal(t, []string{"post_count"}, s.FindSelection("post_count"))
	})
	t.Run("Postgres", func(t *testing.T) {
		b := Dialect(dialect.Postgres)
		s := b.Select("*").From(Table("t1"))
		require.Empty(t, s.FindSelection("c"))
		s.Select("c")
		require.Equal(t, []string{"c"}, s.FindSelection("c"))
		s.Select(s.C("c"))
		require.Equal(t, []string{`"t1"."c"`}, s.FindSelection("c"))
		s.AppendSelectAs(s.C("d"), "e")
		require.Equal(t, []string{"e"}, s.FindSelection("e"))
		require.Empty(t, s.FindSelection("d"))
		t2 := b.Table("t2").As("t2")
		s.Join(t2)
		s.Select(t2.C("e"), "t2.e", s.C("e"), "t1.e", "e")
		require.Equal(t, []string{`"t2"."e"`, "t2.e", `"t1"."e"`, "t1.e", "e"}, s.FindSelection("e"))
	})
}

func TestFormattedColumnFromSubQuery(t *testing.T) {
	q := Select("*").From(Select("*").AppendSelectExprAs(P(func(b *Builder) {
		b.SetDialect(dialect.Postgres)
		b.WriteString("calculate_score")
		b.Wrap(func(bb *Builder) {
			bb.WriteString(Table("table_name").C("field_name")).Comma().Args("test")
		})
	}), "score").From(Table("table_name").As("table_name_alias")))
	require.Equal(t, "`table_name_alias`.`score`", q.C("score"))
}
