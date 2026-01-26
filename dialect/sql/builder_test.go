// Copyright 2019-present Facebook Inc. All rights reserved.
// This source code is licensed under the Apache 2.0 license found
// in the LICENSE file in the root directory of this source tree.

package sql

import (
	"context"
	"database/sql/driver"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"entgo.io/ent/dialect"
	"github.com/stretchr/testify/require"
)

func TestBuilder(t *testing.T) {
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
		{
			input:     Insert("users").Columns("age").Values(1),
			wantQuery: "INSERT INTO `users` (`age`) VALUES (?)",
			wantArgs:  []any{1},
		},
		{
			input:     Insert("users").Columns("age").Values(1).Schema("mydb"),
			wantQuery: "INSERT INTO `mydb`.`users` (`age`) VALUES (?)",
			wantArgs:  []any{1},
		},
		{
			input:     Dialect(dialect.Postgres).Insert("users").Columns("age").Values(1),
			wantQuery: `INSERT INTO "users" ("age") VALUES ($1)`,
			wantArgs:  []any{1},
		},
		{
			input:     Dialect(dialect.Postgres).Insert("users").Columns("age").Values(1).Schema("mydb"),
			wantQuery: `INSERT INTO "mydb"."users" ("age") VALUES ($1)`,
			wantArgs:  []any{1},
		},
		{
			input:     Dialect(dialect.SQLite).Insert("users").Columns("age").Values(1).Schema("mydb"),
			wantQuery: "INSERT INTO `users` (`age`) VALUES (?)",
			wantArgs:  []any{1},
		},
		{
			input:     Dialect(dialect.YDB).Insert("users").Columns("age").Values(1),
			wantQuery: "INSERT INTO `users` (`age`) VALUES ($p0)",
			wantArgs:  []any{driver.NamedValue{Name: "p0", Value: 1}},
		},
		{
			input:     Dialect(dialect.Postgres).Insert("users").Columns("age").Values(1).Returning("id"),
			wantQuery: `INSERT INTO "users" ("age") VALUES ($1) RETURNING "id"`,
			wantArgs:  []any{1},
		},
		{
			input:     Dialect(dialect.Postgres).Insert("users").Columns("age").Values(1).Returning("id").Returning("name"),
			wantQuery: `INSERT INTO "users" ("age") VALUES ($1) RETURNING "name"`,
			wantArgs:  []any{1},
		},
		{
			input:     Insert("users").Columns("name", "age").Values("a8m", 10),
			wantQuery: "INSERT INTO `users` (`name`, `age`) VALUES (?, ?)",
			wantArgs:  []any{"a8m", 10},
		},
		{
			input:     Dialect(dialect.Postgres).Insert("users").Columns("name", "age").Values("a8m", 10),
			wantQuery: `INSERT INTO "users" ("name", "age") VALUES ($1, $2)`,
			wantArgs:  []any{"a8m", 10},
		},
		{
			input:     Insert("users").Columns("name", "age").Values("a8m", 10).Values("foo", 20),
			wantQuery: "INSERT INTO `users` (`name`, `age`) VALUES (?, ?), (?, ?)",
			wantArgs:  []any{"a8m", 10, "foo", 20},
		},
		{
			input:     Dialect(dialect.Postgres).Insert("users").Columns("name", "age").Values("a8m", 10).Values("foo", 20),
			wantQuery: `INSERT INTO "users" ("name", "age") VALUES ($1, $2), ($3, $4)`,
			wantArgs:  []any{"a8m", 10, "foo", 20},
		},
		{
			input: Dialect(dialect.Postgres).Insert("users").
				Columns("name", "age").
				Values("a8m", 10).
				Values("foo", 20).
				Values("bar", 30),
			wantQuery: `INSERT INTO "users" ("name", "age") VALUES ($1, $2), ($3, $4), ($5, $6)`,
			wantArgs:  []any{"a8m", 10, "foo", 20, "bar", 30},
		},
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
			input:     Dialect(dialect.YDB).Update("users").Set("name", "foo"),
			wantQuery: "UPDATE `users` SET `name` = $p0",
			wantArgs:  []any{driver.NamedValue{Name: "p0", Value: "foo"}},
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
			input: Dialect(dialect.YDB).
				Select().
				From(Table("users")).
				Where(EQ("name", "Alex")),
			wantQuery: "SELECT * FROM `users` WHERE `name` = $p0",
			wantArgs:  []any{driver.NamedValue{Name: "p0", Value: "Alex"}},
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
			input: Update("users").
				Set("name", "foo").
				Set("age", 10).
				Where(And(EQ("name", "foo"), EQ("age", 20))),
			wantQuery: "UPDATE `users` SET `name` = ?, `age` = ? WHERE `name` = ? AND `age` = ?",
			wantArgs:  []any{"foo", 10, "foo", 20},
		},
		{
			input: Delete("users").
				Where(NotNull("parent_id")),
			wantQuery: "DELETE FROM `users` WHERE `parent_id` IS NOT NULL",
		},
		{
			input: Delete("users").
				Where(NotNull("parent_id")).
				Schema("mydb"),
			wantQuery: "DELETE FROM `mydb`.`users` WHERE `parent_id` IS NOT NULL",
		},
		{
			input: Dialect(dialect.SQLite).
				Delete("users").
				Where(NotNull("parent_id")).
				Schema("mydb"),
			wantQuery: "DELETE FROM `users` WHERE `parent_id` IS NOT NULL",
		},
		{
			input: Dialect(dialect.YDB).
				Delete("users").
				Where(NotNull("parent_id")),
			wantQuery: "DELETE FROM `users` WHERE `parent_id` IS NOT NULL",
		},
		{
			input: Dialect(dialect.Postgres).
				Delete("users").
				Where(IsNull("parent_id")),
			wantQuery: `DELETE FROM "users" WHERE "parent_id" IS NULL`,
		},
		{
			input: Dialect(dialect.Postgres).
				Delete("users").
				Where(IsNull("parent_id")).
				Schema("mydb"),
			wantQuery: `DELETE FROM "mydb"."users" WHERE "parent_id" IS NULL`,
		},
		{
			input: Delete("users").
				Where(And(IsNull("parent_id"), NotIn("name", "foo", "bar"))),
			wantQuery: "DELETE FROM `users` WHERE `parent_id` IS NULL AND `name` NOT IN (?, ?)",
			wantArgs:  []any{"foo", "bar"},
		},
		{
			input: Dialect(dialect.Postgres).
				Delete("users").
				Where(And(IsNull("parent_id"), NotIn("name", "foo", "bar"))),
			wantQuery: `DELETE FROM "users" WHERE "parent_id" IS NULL AND "name" NOT IN ($1, $2)`,
			wantArgs:  []any{"foo", "bar"},
		},
		{
			input: Delete("users").
				Where(And(IsNull("parent_id"), In("name"))),
			wantQuery: "DELETE FROM `users` WHERE `parent_id` IS NULL AND FALSE",
		},
		{
			input: Delete("users").
				Where(And(IsNull("parent_id"), NotIn("name"))),
			wantQuery: "DELETE FROM `users` WHERE `parent_id` IS NULL AND (NOT (FALSE))",
		},
		{
			input: Delete("users").
				Where(And(False(), False())),
			wantQuery: "DELETE FROM `users` WHERE FALSE AND FALSE",
		},
		{
			input: Dialect(dialect.Postgres).
				Delete("users").
				Where(And(False(), False())),
			wantQuery: `DELETE FROM "users" WHERE FALSE AND FALSE`,
		},
		{
			input: Delete("users").
				Where(Or(NotNull("parent_id"), EQ("parent_id", 10))),
			wantQuery: "DELETE FROM `users` WHERE `parent_id` IS NOT NULL OR `parent_id` = ?",
			wantArgs:  []any{10},
		},
		{
			input: Dialect(dialect.Postgres).
				Delete("users").
				Where(Or(NotNull("parent_id"), EQ("parent_id", 10))),
			wantQuery: `DELETE FROM "users" WHERE "parent_id" IS NOT NULL OR "parent_id" = $1`,
			wantArgs:  []any{10},
		},
		{
			input: Delete("users").
				Where(
					Or(
						And(EQ("name", "foo"), EQ("age", 10)),
						And(EQ("name", "bar"), EQ("age", 20)),
						And(
							EQ("name", "qux"),
							Or(EQ("age", 1), EQ("age", 2)),
						),
					),
				),
			wantQuery: "DELETE FROM `users` WHERE (`name` = ? AND `age` = ?) OR (`name` = ? AND `age` = ?) OR (`name` = ? AND (`age` = ? OR `age` = ?))",
			wantArgs:  []any{"foo", 10, "bar", 20, "qux", 1, 2},
		},
		{
			input: Dialect(dialect.Postgres).
				Delete("users").
				Where(
					Or(
						And(EQ("name", "foo"), EQ("age", 10)),
						And(EQ("name", "bar"), EQ("age", 20)),
						And(
							EQ("name", "qux"),
							Or(EQ("age", 1), EQ("age", 2)),
						),
					),
				),
			wantQuery: `DELETE FROM "users" WHERE ("name" = $1 AND "age" = $2) OR ("name" = $3 AND "age" = $4) OR ("name" = $5 AND ("age" = $6 OR "age" = $7))`,
			wantArgs:  []any{"foo", 10, "bar", 20, "qux", 1, 2},
		},
		{
			input: Delete("users").
				Where(
					Or(
						And(EQ("name", "foo"), EQ("age", 10)),
						And(EQ("name", "bar"), EQ("age", 20)),
					),
				).
				Where(EQ("role", "admin")),
			wantQuery: "DELETE FROM `users` WHERE ((`name` = ? AND `age` = ?) OR (`name` = ? AND `age` = ?)) AND `role` = ?",
			wantArgs:  []any{"foo", 10, "bar", 20, "admin"},
		},
		{
			input: Dialect(dialect.Postgres).
				Delete("users").
				Where(
					Or(
						And(EQ("name", "foo"), EQ("age", 10)),
						And(EQ("name", "bar"), EQ("age", 20)),
					),
				).
				Where(EQ("role", "admin")),
			wantQuery: `DELETE FROM "users" WHERE (("name" = $1 AND "age" = $2) OR ("name" = $3 AND "age" = $4)) AND "role" = $5`,
			wantArgs:  []any{"foo", 10, "bar", 20, "admin"},
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
				return Select(t1.C("id"), t2.C("name")).From(t1).Join(t2)
			}(),
			wantQuery: "SELECT `u`.`id`, `g`.`name` FROM `users` AS `u` JOIN `groups` AS `g`",
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
				d := Dialect(dialect.YDB)
				t1 := d.Table("users").As("u")
				t2 := d.Table("groups").As("g")
				return d.Select(t1.C("id"), t2.C("name")).
					From(t1).
					Join(t2).
					On(t1.C("id"), t2.C("user_id"))
			}(),
			wantQuery: "SELECT `u`.`id` AS `id`, `g`.`name` AS `name` FROM `users` AS `u` JOIN `groups` AS `g` ON `u`.`id` = `g`.`user_id`",
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
				selector := Select().Where(Or(EQ("name", "foo"), EQ("name", "bar")))
				return Delete("users").FromSelect(selector)
			}(),
			wantQuery: "DELETE FROM `users` WHERE `name` = ? OR `name` = ?",
			wantArgs:  []any{"foo", "bar"},
		},
		{
			input: func() Querier {
				d := Dialect(dialect.Postgres)
				selector := d.Select().Where(Or(EQ("name", "foo"), EQ("name", "bar")))
				return d.Delete("users").FromSelect(selector)
			}(),
			wantQuery: `DELETE FROM "users" WHERE "name" = $1 OR "name" = $2`,
			wantArgs:  []any{"foo", "bar"},
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
			input: func() Querier {
				selector := Select().From(Table("groups")).Where(EQ("name", "foo"))
				return Delete("users").FromSelect(selector)
			}(),
			wantQuery: "DELETE FROM `groups` WHERE `name` = ?",
			wantArgs:  []any{"foo"},
		},
		{
			input: func() Querier {
				d := Dialect(dialect.Postgres)
				selector := d.Select().From(Table("groups")).Where(EQ("name", "foo"))
				return d.Delete("users").FromSelect(selector)
			}(),
			wantQuery: `DELETE FROM "groups" WHERE "name" = $1`,
			wantArgs:  []any{"foo"},
		},
		{
			input: func() Querier {
				selector := Select()
				return Delete("users").FromSelect(selector)
			}(),
			wantQuery: "DELETE FROM `users`",
		},
		{
			input: func() Querier {
				d := Dialect(dialect.Postgres)
				selector := d.Select()
				return d.Delete("users").FromSelect(selector)
			}(),
			wantQuery: `DELETE FROM "users"`,
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
			input:     Select().Count().From(Table("users")),
			wantQuery: "SELECT COUNT(*) FROM `users`",
		},
		{
			input: Dialect(dialect.Postgres).
				Select().Count().From(Table("users")),
			wantQuery: `SELECT COUNT(*) FROM "users"`,
		},
		{
			input: Dialect(dialect.YDB).
				Select().Count().From(Table("users")),
			wantQuery: "SELECT COUNT(*) AS `count` FROM `users`",
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
			input: Dialect(dialect.YDB).
				Select("name", Count("*")).
				From(Table("users")).
				GroupBy("name"),
			wantQuery: "SELECT `name`, COUNT(*) AS `count` FROM `users` GROUP BY `name`",
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
			input: Dialect(dialect.YDB).
				Select("*").
				From(Table("users")).
				OrderBy("id").
				Limit(10),
			wantQuery: "SELECT * FROM `users` ORDER BY `id` LIMIT 10",
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
		{
			input: Dialect(dialect.YDB).
				Delete("users").
				Where(EQ("id", 1)),
			wantQuery: "DELETE FROM `users` WHERE `id` = $p0",
			wantArgs:  []any{driver.NamedValue{Name: "p0", Value: 1}},
		},
		{
			input: Dialect(dialect.YDB).
				Delete("users").
				Where(EQ("status", "inactive")).
				Where(LT("last_login", "2024-01-01")),
			wantQuery: "DELETE FROM `users` WHERE `status` = $p0 AND `last_login` < $p1",
			wantArgs: []any{
				driver.NamedValue{Name: "p0", Value: "inactive"},
				driver.NamedValue{Name: "p1", Value: "2024-01-01"},
			},
		},
		{
			input: Dialect(dialect.YDB).
				Delete("users").
				Where(And(
					EQ("status", "cancelled"),
					GT("age", 18),
				)),
			wantQuery: "DELETE FROM `users` WHERE `status` = $p0 AND `age` > $p1",
			wantArgs: []any{
				driver.NamedValue{Name: "p0", Value: "cancelled"},
				driver.NamedValue{Name: "p1", Value: 18},
			},
		},
		{
			input: Dialect(dialect.YDB).
				Delete("users").
				Where(Or(
					EQ("status", "cancelled"),
					EQ("status", "expired"),
				)),
			wantQuery: "DELETE FROM `users` WHERE `status` = $p0 OR `status` = $p1",
			wantArgs: []any{
				driver.NamedValue{Name: "p0", Value: "cancelled"},
				driver.NamedValue{Name: "p1", Value: "expired"},
			},
		},
		{
			input: func() Querier {
				d := Dialect(dialect.YDB)
				subquery := d.Select("id", "email").
					From(Table("users")).
					Where(EQ("status", "ToDelete"))
				return d.Delete("users").On(subquery)
			}(),
			wantQuery: "DELETE FROM `users` ON SELECT `id`, `email` FROM `users` WHERE `status` = $p0",
			wantArgs:  []any{driver.NamedValue{Name: "p0", Value: "ToDelete"}},
		},
		{
			input: func() Querier {
				d := Dialect(dialect.YDB)
				subquery := d.Select("*").
					From(Table("temp_delete_list"))
				return d.Delete("users").On(subquery)
			}(),
			wantQuery: "DELETE FROM `users` ON SELECT * FROM `temp_delete_list`",
		},
		{
			input: Dialect(dialect.YDB).
				Delete("orders").
				Where(EQ("status", "cancelled")).
				Returning("*"),
			wantQuery: "DELETE FROM `orders` WHERE `status` = $p0 RETURNING *",
			wantArgs:  []any{driver.NamedValue{Name: "p0", Value: "cancelled"}},
		},
		{
			input: Dialect(dialect.YDB).
				Delete("orders").
				Where(EQ("status", "cancelled")).
				Returning("order_id", "order_date"),
			wantQuery: "DELETE FROM `orders` WHERE `status` = $p0 RETURNING `order_id`, `order_date`",
			wantArgs:  []any{driver.NamedValue{Name: "p0", Value: "cancelled"}},
		},
		{
			input: func() Querier {
				d := Dialect(dialect.YDB)
				subquery := d.Select("id").
					From(Table("temp_ids"))
				return d.Delete("users").
					On(subquery).
					Returning("id", "name")
			}(),
			wantQuery: "DELETE FROM `users` ON SELECT `id` FROM `temp_ids` RETURNING `id`, `name`",
		},
		{
			input:     Dialect(dialect.YDB).Delete("users"),
			wantQuery: "DELETE FROM `users`",
		},
		{
			input: Dialect(dialect.YDB).
				Insert("users").
				Columns("name", "age").
				Values("a8m", 10),
			wantQuery: "INSERT INTO `users` (`name`, `age`) VALUES ($p0, $p1)",
			wantArgs: []any{
				driver.NamedValue{Name: "p0", Value: "a8m"},
				driver.NamedValue{Name: "p1", Value: 10},
			},
		},
		{
			input: Dialect(dialect.YDB).
				Update("users").
				Set("name", "foo").
				Where(EQ("id", 1)),
			wantQuery: "UPDATE `users` SET `name` = $p0 WHERE `id` = $p1",
			wantArgs: []any{
				driver.NamedValue{Name: "p0", Value: "foo"},
				driver.NamedValue{Name: "p1", Value: 1},
			},
		},
		{
			input: Dialect(dialect.YDB).
				Select("*").
				From(Table("users")).
				Where(EQ("name", "Alex")),
			wantQuery: "SELECT * FROM `users` WHERE `name` = $p0",
			wantArgs:  []any{driver.NamedValue{Name: "p0", Value: "Alex"}},
		},
		{
			input: Dialect(dialect.YDB).
				Insert("users").
				Columns("name", "age").
				Values("a8m", 10).
				Returning("id"),
			wantQuery: "INSERT INTO `users` (`name`, `age`) VALUES ($p0, $p1) RETURNING `id`",
			wantArgs: []any{
				driver.NamedValue{Name: "p0", Value: "a8m"},
				driver.NamedValue{Name: "p1", Value: 10},
			},
		},
		{
			input: Dialect(dialect.YDB).
				Insert("users").
				Columns("name", "age").
				Values("a8m", 10).
				Returning("*"),
			wantQuery: "INSERT INTO `users` (`name`, `age`) VALUES ($p0, $p1) RETURNING *",
			wantArgs: []any{
				driver.NamedValue{Name: "p0", Value: "a8m"},
				driver.NamedValue{Name: "p1", Value: 10},
			},
		},
		{
			input: Dialect(dialect.YDB).
				Upsert("users").
				Columns("id", "name", "age").
				Values(1, "a8m", 10),
			wantQuery: "UPSERT INTO `users` (`id`, `name`, `age`) VALUES ($p0, $p1, $p2)",
			wantArgs: []any{
				driver.NamedValue{Name: "p0", Value: 1},
				driver.NamedValue{Name: "p1", Value: "a8m"},
				driver.NamedValue{Name: "p2", Value: 10},
			},
		},
		{
			input: Dialect(dialect.YDB).
				Upsert("users").
				Columns("id", "name", "age").
				Values(1, "a8m", 10).
				Values(2, "foo", 20),
			wantQuery: "UPSERT INTO `users` (`id`, `name`, `age`) VALUES ($p0, $p1, $p2), ($p3, $p4, $p5)",
			wantArgs: []any{
				driver.NamedValue{Name: "p0", Value: 1},
				driver.NamedValue{Name: "p1", Value: "a8m"},
				driver.NamedValue{Name: "p2", Value: 10},
				driver.NamedValue{Name: "p3", Value: 2},
				driver.NamedValue{Name: "p4", Value: "foo"},
				driver.NamedValue{Name: "p5", Value: 20},
			},
		},
		{
			input: Dialect(dialect.YDB).
				Upsert("orders").
				Columns("order_id", "status", "amount").
				Values(1001, "shipped", 500).
				Returning("*"),
			wantQuery: "UPSERT INTO `orders` (`order_id`, `status`, `amount`) VALUES ($p0, $p1, $p2) RETURNING *",
			wantArgs: []any{
				driver.NamedValue{Name: "p0", Value: 1001},
				driver.NamedValue{Name: "p1", Value: "shipped"},
				driver.NamedValue{Name: "p2", Value: 500},
			},
		},
		{
			input: Dialect(dialect.YDB).
				Upsert("users").
				Columns("user_id", "name", "email").
				Values(42, "John Doe", "john@example.com").
				Returning("user_id", "email"),
			wantQuery: "UPSERT INTO `users` (`user_id`, `name`, `email`) VALUES ($p0, $p1, $p2) RETURNING `user_id`, `email`",
			wantArgs: []any{
				driver.NamedValue{Name: "p0", Value: 42},
				driver.NamedValue{Name: "p1", Value: "John Doe"},
				driver.NamedValue{Name: "p2", Value: "john@example.com"},
			},
		},
		{
			input: Dialect(dialect.YDB).
				Replace("users").
				Columns("id", "name", "age").
				Values(1, "a8m", 10),
			wantQuery: "REPLACE INTO `users` (`id`, `name`, `age`) VALUES ($p0, $p1, $p2)",
			wantArgs: []any{
				driver.NamedValue{Name: "p0", Value: 1},
				driver.NamedValue{Name: "p1", Value: "a8m"},
				driver.NamedValue{Name: "p2", Value: 10},
			},
		},
		{
			input: Dialect(dialect.YDB).
				Replace("users").
				Columns("id", "name", "age").
				Values(1, "a8m", 10).
				Values(2, "foo", 20),
			wantQuery: "REPLACE INTO `users` (`id`, `name`, `age`) VALUES ($p0, $p1, $p2), ($p3, $p4, $p5)",
			wantArgs: []any{
				driver.NamedValue{Name: "p0", Value: 1},
				driver.NamedValue{Name: "p1", Value: "a8m"},
				driver.NamedValue{Name: "p2", Value: 10},
				driver.NamedValue{Name: "p3", Value: 2},
				driver.NamedValue{Name: "p4", Value: "foo"},
				driver.NamedValue{Name: "p5", Value: 20},
			},
		},
		{
			input: Dialect(dialect.YDB).
				Replace("orders").
				Columns("order_id", "status", "amount").
				Values(1001, "shipped", 500).
				Returning("*"),
			wantQuery: "REPLACE INTO `orders` (`order_id`, `status`, `amount`) VALUES ($p0, $p1, $p2) RETURNING *",
			wantArgs: []any{
				driver.NamedValue{Name: "p0", Value: 1001},
				driver.NamedValue{Name: "p1", Value: "shipped"},
				driver.NamedValue{Name: "p2", Value: 500},
			},
		},
		{
			input: Dialect(dialect.YDB).
				Update("users").
				Set("name", "foo").
				Where(EQ("id", 1)).
				Returning("*"),
			wantQuery: "UPDATE `users` SET `name` = $p0 WHERE `id` = $p1 RETURNING *",
			wantArgs: []any{
				driver.NamedValue{Name: "p0", Value: "foo"},
				driver.NamedValue{Name: "p1", Value: 1},
			},
		},
		{
			input: Dialect(dialect.YDB).
				Update("orders").
				Set("status", "shipped").
				Where(LT("order_date", "2023-01-01")).
				Returning("order_id", "status"),
			wantQuery: "UPDATE `orders` SET `status` = $p0 WHERE `order_date` < $p1 RETURNING `order_id`, `status`",
			wantArgs: []any{
				driver.NamedValue{Name: "p0", Value: "shipped"},
				driver.NamedValue{Name: "p1", Value: "2023-01-01"},
			},
		},
		{
			input: func() Querier {
				d := Dialect(dialect.YDB)
				subquery := d.Select("key", "value").
					From(Table("temp_updates")).
					Where(EQ("status", "pending"))
				return d.Update("users").On(subquery)
			}(),
			wantQuery: "UPDATE `users` ON SELECT `key`, `value` FROM `temp_updates` WHERE `status` = $p0",
			wantArgs:  []any{driver.NamedValue{Name: "p0", Value: "pending"}},
		},
		{
			input: func() Querier {
				d := Dialect(dialect.YDB)
				subquery := d.Select("*").
					From(Table("staged_data"))
				return d.Update("users").
					On(subquery).
					Returning("id", "name")
			}(),
			wantQuery: "UPDATE `users` ON SELECT * FROM `staged_data` RETURNING `id`, `name`",
		},
		{
			input: func() Querier {
				d := Dialect(dialect.YDB)
				t1 := d.Table("orders")
				t2 := d.Table("users").View("idx_email").As("u")
				return d.Select(t1.C("id"), t2.C("name")).
					From(t1).
					Join(t2).
					On(t1.C("user_id"), t2.C("id"))
			}(),
			wantQuery: "SELECT `orders`.`id` AS `id`, `u`.`name` AS `name` FROM `orders` JOIN `users` VIEW `idx_email` AS `u` ON `orders`.`user_id` = `u`.`id`",
		},
		{
			input: func() Querier {
				d := Dialect(dialect.YDB)
				t1 := d.Table("a_table").As("a")
				t2 := d.Table("b_table").View("b_index_ref").As("b")
				return d.Select(t1.C("value"), t2.C("value")).
					From(t1).
					Join(t2).
					On(t1.C("ref"), t2.C("ref"))
			}(),
			wantQuery: "SELECT `a`.`value` AS `value`, `b`.`value` AS `value` FROM `a_table` AS `a` JOIN `b_table` VIEW `b_index_ref` AS `b` ON `a`.`ref` = `b`.`ref`",
		},
		{
			input: func() Querier {
				d := Dialect(dialect.YDB)
				t1 := d.Table("orders")
				t2 := d.Table("products").View("idx_category")
				return d.Select("*").
					From(t1).
					LeftJoin(t2).
					On(t1.C("product_id"), t2.C("id")).
					Where(EQ(t2.C("category"), "Electronics"))
			}(),
			wantQuery: "SELECT * FROM `orders` LEFT JOIN `products` VIEW `idx_category` AS `t1` ON `orders`.`product_id` = `t1`.`id` WHERE `t1`.`category` = $p0",
			wantArgs:  []any{driver.NamedValue{Name: "p0", Value: "Electronics"}},
		},
		{
			input: func() Querier {
				d := Dialect(dialect.YDB)
				t1 := d.Table("users")
				t2 := d.Table("blacklist")
				return d.Select(t1.C("id"), t1.C("name")).
					From(t1).
					LeftSemiJoin(t2).
					On(t1.C("id"), t2.C("user_id"))
			}(),
			wantQuery: "SELECT `users`.`id` AS `id`, `users`.`name` AS `name` FROM `users` LEFT SEMI JOIN `blacklist` AS `t1` ON `users`.`id` = `t1`.`user_id`",
		},
		{
			input: func() Querier {
				d := Dialect(dialect.YDB)
				t1 := d.Table("orders")
				t2 := d.Table("active_users").As("t1")
				return d.Select(t2.C("id"), t2.C("email")).
					From(t1).
					RightSemiJoin(t2).
					On(t1.C("user_id"), t2.C("id"))
			}(),
			wantQuery: "SELECT `t1`.`id` AS `id`, `t1`.`email` AS `email` FROM `orders` RIGHT SEMI JOIN `active_users` AS `t1` ON `orders`.`user_id` = `t1`.`id`",
		},
		{
			input: func() Querier {
				d := Dialect(dialect.YDB)
				t1 := d.Table("users")
				t2 := d.Table("deleted_users")
				return d.Select(t1.C("id"), t1.C("name")).
					From(t1).
					LeftOnlyJoin(t2).
					On(t1.C("id"), t2.C("id"))
			}(),
			wantQuery: "SELECT `users`.`id` AS `id`, `users`.`name` AS `name` FROM `users` LEFT ONLY JOIN `deleted_users` AS `t1` ON `users`.`id` = `t1`.`id`",
		},
		{
			input: func() Querier {
				d := Dialect(dialect.YDB)
				t1 := d.Table("archived")
				t2 := d.Table("users").As("t1")
				return d.Select(t2.C("id"), t2.C("status")).
					From(t1).
					RightOnlyJoin(t2).
					On(t1.C("user_id"), t2.C("id"))
			}(),
			wantQuery: "SELECT `t1`.`id` AS `id`, `t1`.`status` AS `status` FROM `archived` RIGHT ONLY JOIN `users` AS `t1` ON `archived`.`user_id` = `t1`.`id`",
		},
		{
			input: func() Querier {
				d := Dialect(dialect.YDB)
				t1 := d.Table("table_a").As("a")
				t2 := d.Table("table_b").As("b")
				return d.Select(t1.C("key"), t2.C("key")).
					From(t1).
					ExclusionJoin(t2).
					On(t1.C("key"), t2.C("key"))
			}(),
			wantQuery: "SELECT `a`.`key` AS `key`, `b`.`key` AS `key` FROM `table_a` AS `a` EXCLUSION JOIN `table_b` AS `b` ON `a`.`key` = `b`.`key`",
		},
		{
			input: func() Querier {
				t1 := Table("users").As("u")
				t2 := Table("groups").As("g")
				return Select(t1.C("id"), t2.C("name")).
					From(t1).
					CrossJoin(t2)
			}(),
			wantQuery: "SELECT `u`.`id`, `g`.`name` FROM `users` AS `u` CROSS JOIN `groups` AS `g`",
		},
		{
			input: func() Querier {
				d := Dialect(dialect.YDB)
				t1 := d.Table("a_table").As("a")
				t2 := d.Table("b_table").As("b")
				t3 := d.Table("c_table").As("c")
				return d.Select(t1.C("value"), t2.C("value"), t3.C("column2")).
					From(t1).
					CrossJoin(t2).
					LeftJoin(t3).
					OnP(And(EQ(t3.C("ref"), Expr(t1.C("key"))), EQ(t3.C("column1"), Expr(t2.C("value")))))
			}(),
			wantQuery: "SELECT `a`.`value` AS `value`, `b`.`value` AS `value`, `c`.`column2` AS `column2` FROM `a_table` AS `a` CROSS JOIN `b_table` AS `b` LEFT JOIN `c_table` AS `c` ON `c`.`ref` = `a`.`key` AND `c`.`column1` = `b`.`value`",
		},
		{
			input: Dialect(dialect.YDB).
				Select("value").
				From(Table("my_table")).
				Distinct(),
			wantQuery: "SELECT DISTINCT `value` FROM `my_table`",
		},
		{
			input: Dialect(dialect.YDB).
				Select("*").
				From(Table("users")).
				Distinct(),
			wantQuery: "SELECT DISTINCT * FROM `users`",
		},
		{
			input: Dialect(dialect.YDB).
				Select("status", "type").
				From(Table("orders")).
				Distinct().
				Where(GT("created_at", "2024-01-01")),
			wantQuery: "SELECT DISTINCT `status`, `type` FROM `orders` WHERE `created_at` > $p0",
			wantArgs:  []any{driver.NamedValue{Name: "p0", Value: "2024-01-01"}},
		},
		{
			input: Dialect(dialect.YDB).
				Select("category").
				From(Table("products")).
				Distinct().
				OrderBy("category").
				Limit(10),
			wantQuery: "SELECT DISTINCT `category` FROM `products` ORDER BY `category` LIMIT 10",
		},
		{
			input: Dialect(dialect.YDB).
				Select("name", "age").
				From(Table("users")).
				Distinct(),
			wantQuery: "SELECT DISTINCT `name`, `age` FROM `users`",
		},
		{
			input: Dialect(dialect.YDB).
				Select("email").
				From(Table("subscribers")).
				Where(EQ("active", true)).
				Distinct(),
			wantQuery: "SELECT DISTINCT `email` FROM `subscribers` WHERE `active`",
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

func TestBuilder_Err(t *testing.T) {
	b := Select("i-")
	require.NoError(t, b.Err())
	b.AddError(fmt.Errorf("invalid"))
	require.EqualError(t, b.Err(), "invalid")
	b.AddError(fmt.Errorf("unexpected"))
	require.EqualError(t, b.Err(), "invalid; unexpected")
	b.Where(P(func(builder *Builder) {
		builder.AddError(fmt.Errorf("inner"))
	}))
	_, _ = b.Query()
	require.EqualError(t, b.Err(), "invalid; unexpected; inner")
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

	query, args = Dialect(dialect.YDB).
		Select("key").
		From(Table("T1")).
		Union(
			Dialect(dialect.YDB).
				Select("key").
				From(Table("T2")),
		).
		Query()
	require.Equal(t, "SELECT `key` FROM `T1` UNION SELECT `key` FROM `T2`", query)
	require.Empty(t, args)

	query, args = Dialect(dialect.YDB).
		Select("id", "name").
		From(Table("users")).
		Where(EQ("status", "active")).
		Union(
			Dialect(dialect.YDB).
				Select("id", "name").
				From(Table("users")).
				Where(EQ("status", "inactive")),
		).
		Query()
	require.Equal(t, "SELECT `id`, `name` FROM `users` WHERE `status` = $p0 UNION SELECT `id`, `name` FROM `users` WHERE `status` = $p1", query)
	require.Equal(t, []any{
		driver.NamedValue{Name: "p0", Value: "active"},
		driver.NamedValue{Name: "p1", Value: "inactive"},
	}, args)

	query, args = Dialect(dialect.YDB).
		Select("x").
		UnionAll(Dialect(dialect.YDB).Select("y")).
		UnionAll(Dialect(dialect.YDB).Select("z")).
		Query()
	require.Equal(t, "SELECT `x` UNION ALL SELECT `y` UNION ALL SELECT `z`", query)
	require.Empty(t, args)

	query, args = Dialect(dialect.YDB).
		Select("id").
		From(Table("orders")).
		Where(EQ("status", "pending")).
		Union(
			Dialect(dialect.YDB).
				Select("id").
				From(Table("orders")).
				Where(EQ("status", "processing")),
		).
		OrderBy("id").
		Limit(100).
		Query()
	require.Equal(t, "SELECT `id` FROM `orders` WHERE `status` = $p0 UNION SELECT `id` FROM `orders` WHERE `status` = $p1 ORDER BY `id` LIMIT 100", query)
	require.Equal(t, []any{
		driver.NamedValue{Name: "p0", Value: "pending"},
		driver.NamedValue{Name: "p1", Value: "processing"},
	}, args)
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

func TestSelector_AssumeOrderBy_YDB(t *testing.T) {
	query, args := Dialect(dialect.YDB).
		Select("*").
		From(Table("my_table")).
		AssumeOrderBy("key").
		Query()
	require.Equal(t, "SELECT * FROM `my_table` ASSUME ORDER BY `key`", query)
	require.Empty(t, args)

	query, args = Dialect(dialect.YDB).
		Select("key", "subkey").
		From(Table("my_table")).
		AssumeOrderBy("key", Desc("subkey")).
		Query()
	require.Equal(t, "SELECT `key`, `subkey` FROM `my_table` ASSUME ORDER BY `key`, `subkey` DESC", query)
	require.Empty(t, args)

	query, args = Dialect(dialect.YDB).
		Select("*").
		From(Table("orders")).
		Where(EQ("status", "active")).
		AssumeOrderBy("created_at").
		Query()
	require.Equal(t, "SELECT * FROM `orders` WHERE `status` = $p0 ASSUME ORDER BY `created_at`", query)
	require.Equal(t, []any{driver.NamedValue{Name: "p0", Value: "active"}}, args)

	query, args = Dialect(dialect.YDB).
		Select("id", "name").
		From(Table("users")).
		AssumeOrderBy("id").
		Limit(10).
		Query()
	require.Equal(t, "SELECT `id`, `name` FROM `users` ASSUME ORDER BY `id` LIMIT 10", query)
	require.Empty(t, args)

	t.Run("OrderBy then AssumeOrderBy should error", func(t *testing.T) {
		selector := Dialect(dialect.YDB).
			Select("*").
			From(Table("users")).
			OrderBy("id").
			AssumeOrderBy("id")

		err := selector.Err()
		require.Error(t, err)
	})

	t.Run("AssumeOrderBy then OrderBy should error", func(t *testing.T) {
		selector := Dialect(dialect.YDB).
			Select("*").
			From(Table("users")).
			AssumeOrderBy("id").
			OrderBy("id")

		err := selector.Err()
		require.Error(t, err)
	})

	t.Run("Non-YDB dialect with AssumeOrderBy should error", func(t *testing.T) {
		selector := Dialect(dialect.Postgres).
			Select("*").
			From(Table("users")).
			AssumeOrderBy("name")

		err := selector.Err()
		require.Error(t, err)
	})
}

func TestSelector_VIEW_SecondaryIndex_YDB(t *testing.T) {
	t.Run("Simple SELECT with VIEW in FROM clause", func(t *testing.T) {
		d := Dialect(dialect.YDB)
		query, args := d.Select("series_id", "title", "info", "release_date", "views", "uploaded_user_id").
			From(d.Table("series").View("views_index")).
			Where(GTE("views", 1000)).
			Query()

		require.Equal(t, "SELECT `series_id`, `title`, `info`, `release_date`, `views`, `uploaded_user_id` FROM `series` VIEW `views_index` WHERE `views` >= $p0", query)
		require.Equal(t, []any{driver.NamedValue{Name: "p0", Value: 1000}}, args)
	})

	t.Run("JOIN with VIEW on both tables", func(t *testing.T) {
		d := Dialect(dialect.YDB)
		series := d.Table("series").View("users_index").As("t1")
		users := d.Table("users").View("name_index").As("t2")

		query, args := d.Select(series.C("series_id"), series.C("title")).
			From(series).
			Join(users).
			On(series.C("uploaded_user_id"), users.C("user_id")).
			Where(EQ(users.C("name"), "John Doe")).
			Query()

		require.Equal(t, "SELECT `t1`.`series_id` AS `series_id`, `t1`.`title` AS `title` FROM `series` VIEW `users_index` AS `t1` JOIN `users` VIEW `name_index` AS `t2` ON `t1`.`uploaded_user_id` = `t2`.`user_id` WHERE `t2`.`name` = $p0", query)
		require.Equal(t, []any{driver.NamedValue{Name: "p0", Value: "John Doe"}}, args)
	})

	t.Run("VIEW on non-YDB dialect should error", func(t *testing.T) {
		d := Dialect(dialect.Postgres)
		table := d.Table("users").View("idx_name")

		err := table.Err()
		require.Error(t, err)
	})
}

func TestBatchUpdate_YDB(t *testing.T) {
	t.Run("Basic BATCH UPDATE with SET and WHERE", func(t *testing.T) {
		d := Dialect(dialect.YDB)
		query, args := d.BatchUpdate("my_table").
			Set("Value1", "foo").
			Set("Value2", 0).
			Where(GT("Key1", 1)).
			Query()

		require.Equal(t, "BATCH UPDATE `my_table` SET `Value1` = $p0, `Value2` = $p1 WHERE `Key1` > $p2", query)
		require.Equal(t, []any{
			driver.NamedValue{Name: "p0", Value: "foo"},
			driver.NamedValue{Name: "p1", Value: 0},
			driver.NamedValue{Name: "p2", Value: 1},
		}, args)
	})

	t.Run("BATCH UPDATE on non-YDB dialect should error", func(t *testing.T) {
		builder := Dialect(dialect.MySQL).
			BatchUpdate("users").
			Set("status", "active").
			Where(GT("created_at", "2024-01-01"))

		query, args, err := builder.QueryErr()

		require.Empty(t, query)
		require.Empty(t, args)
		require.Error(t, err)
	})

	t.Run("BATCH UPDATE with RETURNING should error", func(t *testing.T) {
		builder := Dialect(dialect.YDB).
			BatchUpdate("users").
			Set("status", "active").
			Returning("id", "status")

		query, args, err := builder.QueryErr()

		require.Empty(t, query)
		require.Empty(t, args)
		require.Error(t, err)
	})

	t.Run("BATCH UPDATE with UPDATE ON pattern should error", func(t *testing.T) {
		d := Dialect(dialect.YDB)
		subquery := d.Select("id").From(Table("orders")).Where(EQ("status", "pending"))

		builder := d.BatchUpdate("users").
			Set("status", "active").
			On(subquery)

		query, args, err := builder.QueryErr()

		require.Empty(t, query)
		require.Empty(t, args)
		require.Error(t, err)
		require.Contains(t, err.Error(), "BATCH UPDATE: UPDATE ON pattern is not supported")
	})
}

func TestBatchDelete_YDB(t *testing.T) {
	t.Run("Basic BATCH DELETE", func(t *testing.T) {
		d := Dialect(dialect.YDB)
		query, args := d.BatchDelete("my_table").
			Where(And(GT("Key1", 1), GTE("Key2", "One"))).
			Query()

		require.Equal(t, "BATCH DELETE FROM `my_table` WHERE `Key1` > $p0 AND `Key2` >= $p1", query)
		require.Equal(t, []any{
			driver.NamedValue{Name: "p0", Value: 1},
			driver.NamedValue{Name: "p1", Value: "One"},
		}, args)
	})

	t.Run("BATCH DELETE on non-YDB dialect should error", func(t *testing.T) {
		builder := Dialect(dialect.MySQL).
			BatchDelete("users").
			Where(GT("id", 100))

		query, args, err := builder.QueryErr()

		require.Empty(t, query)
		require.Empty(t, args)
		require.Error(t, err)
	})

	t.Run("BATCH DELETE with RETURNING should error", func(t *testing.T) {
		builder := Dialect(dialect.YDB).
			BatchDelete("users").
			Where(GT("id", 100)).
			Returning("id")

		query, args, err := builder.QueryErr()

		require.Empty(t, query)
		require.Empty(t, args)
		require.Error(t, err)
	})

	t.Run("BATCH DELETE with DELETE ON pattern should error", func(t *testing.T) {
		d := Dialect(dialect.YDB)
		subquery := d.Select("id").From(Table("users")).Where(EQ("status", "deleted"))

		builder := d.BatchDelete("users").
			On(subquery)

		query, args, err := builder.QueryErr()

		require.Empty(t, query)
		require.Empty(t, args)
		require.Error(t, err)
		require.Contains(t, err.Error(), "BATCH DELETE: DELETE ON pattern is not supported")
	})
}

func TestCreateView_YDB(t *testing.T) {
	t.Run("Basic view with security_invoker", func(t *testing.T) {
		d := Dialect(dialect.YDB)
		query, args := d.CreateView("recent_series").
			As(
				d.Select("*").
					From(Table("db")).
					Where(GT("release_date", "2020-01-01")),
			).
			Query()

		require.Contains(t, query, "CREATE VIEW `recent_series` WITH (security_invoker = TRUE) AS SELECT * FROM `db`")
		require.Contains(t, query, "WHERE `release_date` > $p0")
		require.Equal(t, []any{driver.NamedValue{Name: "p0", Value: "2020-01-01"}}, args)
	})

	t.Run("Non-YDB dialect should not generate WITH clause", func(t *testing.T) {
		query, _ := Dialect(dialect.Postgres).
			CreateView("my_view").
			As(Select("*").From(Table("users"))).
			Query()

		// Postgres should not have the WITH clause
		require.NotContains(t, query, "WITH (security_invoker")
		require.Contains(t, query, `CREATE VIEW "my_view" AS SELECT * FROM "users"`)
	})
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

func TestBuilderContext(t *testing.T) {
	type key string
	want := "myval"
	ctx := context.WithValue(context.Background(), key("mykey"), want)
	sel := Dialect(dialect.Postgres).Select().WithContext(ctx)
	if got := sel.Context().Value(key("mykey")).(string); got != want {
		t.Fatalf("expected selector context key to be %q but got %q", want, got)
	}
	if got := sel.Clone().Context().Value(key("mykey")).(string); got != want {
		t.Fatalf("expected cloned selector context key to be %q but got %q", want, got)
	}
}

type point struct {
	xy []float64
	*testing.T
}

// FormatParam implements the sql.ParamFormatter interface.
func (p point) FormatParam(placeholder string, info *StmtInfo) string {
	require.Equal(p.T, dialect.MySQL, info.Dialect)
	return "ST_GeomFromWKB(" + placeholder + ")"
}

// Value implements the driver.Valuer interface.
func (p point) Value() (driver.Value, error) {
	return p.xy, nil
}

func TestParamFormatter(t *testing.T) {
	p := point{xy: []float64{1, 2}, T: t}
	query, args := Dialect(dialect.MySQL).
		Select().
		From(Table("users")).
		Where(EQ("point", p)).
		Query()
	require.Equal(t, "SELECT * FROM `users` WHERE `point` = ST_GeomFromWKB(?)", query)
	require.Equal(t, p, args[0])
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
	require.Contains(t, s.Err().Error(), "SELECT .. FOR UPDATE/SHARE is not supported")
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

func TestInsert_OnConflict(t *testing.T) {
	t.Run("Postgres", func(t *testing.T) { // And SQLite.
		query, args := Dialect(dialect.Postgres).
			Insert("users").
			Columns("id", "email", "creation_time").
			Values("1", "user@example.com", 1633279231).
			OnConflict(
				ConflictColumns("email"),
				ConflictWhere(EQ("name", "Ariel")),
				ResolveWithNewValues(),
				// Update all new values excepts id field.
				ResolveWith(func(u *UpdateSet) {
					u.SetIgnore("id")
					u.SetIgnore("creation_time")
					u.Add("version", 1)
				}),
				UpdateWhere(NEQ("updated_at", 0)),
			).
			Query()
		require.Equal(t, `INSERT INTO "users" ("id", "email", "creation_time") VALUES ($1, $2, $3) ON CONFLICT ("email") WHERE "name" = $4 DO UPDATE SET "id" = "users"."id", "email" = "excluded"."email", "creation_time" = "users"."creation_time", "version" = COALESCE("users"."version", 0) + $5 WHERE "users"."updated_at" <> $6`, query)
		require.Equal(t, []any{"1", "user@example.com", 1633279231, "Ariel", 1, 0}, args)

		query, args = Dialect(dialect.Postgres).
			Insert("users").
			Columns("id", "name").
			Values("1", "Mashraki").
			OnConflict(
				ConflictConstraint("users_pkey"),
				DoNothing(),
			).
			Query()
		require.Equal(t, `INSERT INTO "users" ("id", "name") VALUES ($1, $2) ON CONFLICT ON CONSTRAINT "users_pkey" DO NOTHING`, query)
		require.Equal(t, []any{"1", "Mashraki"}, args)

		query, args = Dialect(dialect.Postgres).
			Insert("users").
			Columns("id").
			Values(1).
			OnConflict(
				DoNothing(),
			).
			Query()
		require.Equal(t, `INSERT INTO "users" ("id") VALUES ($1) ON CONFLICT DO NOTHING`, query)
		require.Equal(t, []any{1}, args)

		query, args = Dialect(dialect.Postgres).
			Insert("users").
			Columns("id").
			Values(1).
			OnConflict(
				ConflictColumns("id"),
				ResolveWithIgnore(),
			).
			Query()
		require.Equal(t, `INSERT INTO "users" ("id") VALUES ($1) ON CONFLICT ("id") DO UPDATE SET "id" = "users"."id"`, query)
		require.Equal(t, []any{1}, args)

		query, args = Dialect(dialect.Postgres).
			Insert("users").
			Columns("id", "name").
			Values(1, "Mashraki").
			OnConflict(
				ConflictColumns("name"),
				ResolveWith(func(s *UpdateSet) {
					s.SetExcluded("name")
					s.SetNull("created_at")
				}),
			).
			Query()
		require.Equal(t, `INSERT INTO "users" ("id", "name") VALUES ($1, $2) ON CONFLICT ("name") DO UPDATE SET "created_at" = NULL, "name" = "excluded"."name"`, query)
		require.Equal(t, []any{1, "Mashraki"}, args)
	})

	t.Run("MySQL", func(t *testing.T) {
		query, args := Dialect(dialect.MySQL).
			Insert("users").
			Columns("id", "email").
			Values("1", "user@example.com").
			OnConflict(
				ResolveWithNewValues(),
			).
			Query()
		require.Equal(t, "INSERT INTO `users` (`id`, `email`) VALUES (?, ?) ON DUPLICATE KEY UPDATE `id` = VALUES(`id`), `email` = VALUES(`email`)", query)
		require.Equal(t, []any{"1", "user@example.com"}, args)

		query, args = Dialect(dialect.MySQL).
			Insert("users").
			Columns("id", "email").
			Values("1", "user@example.com").
			OnConflict(
				ResolveWithIgnore(),
			).
			Query()
		require.Equal(t, "INSERT INTO `users` (`id`, `email`) VALUES (?, ?) ON DUPLICATE KEY UPDATE `id` = `users`.`id`, `email` = `users`.`email`", query)
		require.Equal(t, []any{"1", "user@example.com"}, args)

		query, args = Dialect(dialect.MySQL).
			Insert("users").
			Columns("id", "name").
			Values("1", "Mashraki").
			OnConflict(
				ResolveWith(func(s *UpdateSet) {
					s.SetExcluded("name")
					s.SetNull("created_at")
					s.Add("version", 1)
				}),
			).
			Query()
		require.Equal(t, "INSERT INTO `users` (`id`, `name`) VALUES (?, ?) ON DUPLICATE KEY UPDATE `created_at` = NULL, `name` = VALUES(`name`), `version` = COALESCE(`users`.`version`, 0) + ?", query)
		require.Equal(t, []any{"1", "Mashraki", 1}, args)

		query, args = Dialect(dialect.MySQL).
			Insert("users").
			Columns("name", "rank").
			Values("Mashraki", nil).
			OnConflict(
				ResolveWithNewValues(),
				ResolveWith(func(s *UpdateSet) {
					s.Set("id", Expr("LAST_INSERT_ID(`id`)"))
				}),
			).
			Query()
		require.Equal(t, "INSERT INTO `users` (`name`, `rank`) VALUES (?, NULL) ON DUPLICATE KEY UPDATE `name` = VALUES(`name`), `rank` = VALUES(`rank`), `id` = LAST_INSERT_ID(`id`)", query)
		require.Equal(t, []any{"Mashraki"}, args)

		query, args = Dialect(dialect.MySQL).
			Insert("users").
			Columns("name", "rank").
			Values("Ariel", 10).
			Values("Mashraki", nil).
			OnConflict(
				ResolveWithNewValues(),
				ResolveWith(func(s *UpdateSet) {
					s.Set("id", Expr("LAST_INSERT_ID(`id`)"))
				}),
			).
			Query()
		require.Equal(t, "INSERT INTO `users` (`name`, `rank`) VALUES (?, ?), (?, NULL) ON DUPLICATE KEY UPDATE `name` = VALUES(`name`), `rank` = VALUES(`rank`), `id` = LAST_INSERT_ID(`id`)", query)
		require.Equal(t, []any{"Ariel", 10, "Mashraki"}, args)
	})
}

func TestEscapePatterns(t *testing.T) {
	q, args := Dialect(dialect.MySQL).
		Update("users").
		SetNull("name").
		Where(
			Or(
				HasPrefix("nickname", "%a8m%"),
				HasSuffix("nickname", "_alexsn_"),
				Contains("nickname", "\\pedro\\"),
				ContainsFold("nickname", "%AbcD%efg"),
			),
		).
		Query()
	require.Equal(t, "UPDATE `users` SET `name` = NULL WHERE `nickname` LIKE ? OR `nickname` LIKE ? OR `nickname` LIKE ? OR `nickname` COLLATE utf8mb4_general_ci LIKE ?", q)
	require.Equal(t, []any{"\\%a8m\\%%", "%\\_alexsn\\_", "%\\\\pedro\\\\%", "%\\%abcd\\%efg%"}, args)

	q, args = Dialect(dialect.SQLite).
		Update("users").
		SetNull("name").
		Where(
			Or(
				HasPrefix("nickname", "%a8m%"),
				HasSuffix("nickname", "_alexsn_"),
				Contains("nickname", "\\pedro\\"),
				ContainsFold("nickname", "%AbcD%efg"),
			),
		).
		Query()
	require.Equal(t, "UPDATE `users` SET `name` = NULL WHERE `nickname` LIKE ? ESCAPE ? OR `nickname` LIKE ? ESCAPE ? OR `nickname` LIKE ? ESCAPE ? OR LOWER(`nickname`) LIKE ? ESCAPE ?", q)
	require.Equal(t, []any{"\\%a8m\\%%", "\\", "%\\_alexsn\\_", "\\", "%\\\\pedro\\\\%", "\\", "%\\%abcd\\%efg%", "\\"}, args)

	q, args = Select("*").From(Table("dataset")).
		Where(Contains("title", "_第一")).Query()
	require.Equal(t, "SELECT * FROM `dataset` WHERE `title` LIKE ?", q)
	require.Equal(t, []any{"%\\_第一%"}, args)
}

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

func TestWindowFunction(t *testing.T) {
	posts := Table("posts")
	base := Select(posts.Columns("id", "content", "author_id")...).
		From(posts).
		Where(EQ("active", true))
	with := With("active_posts").
		As(base).
		With("selected_posts").
		As(
			Select().
				AppendSelect("*").
				AppendSelectExprAs(
					RowNumber().PartitionBy("author_id").OrderBy("id").OrderExpr(Expr("f(`s`)")),
					"row_number",
				).
				From(Table("active_posts")),
		)
	query, args := Select("*").From(Table("selected_posts")).Where(LTE("row_number", 2)).Prefix(with).Query()
	require.Equal(t, "WITH `active_posts` AS (SELECT `posts`.`id`, `posts`.`content`, `posts`.`author_id` FROM `posts` WHERE `active`), `selected_posts` AS (SELECT *, (ROW_NUMBER() OVER (PARTITION BY `author_id` ORDER BY `id`, f(`s`))) AS `row_number` FROM `active_posts`) SELECT * FROM `selected_posts` WHERE `row_number` <= ?", query)
	require.Equal(t, []any{2}, args)
}

func TestWindowFunction_Select(t *testing.T) {
	posts := Table("posts")
	q := Select().
		AppendSelect("*").
		AppendSelectExprAs(
			Window(func(b *Builder) {
				b.WriteString(Sum(posts.C("duration")))
			}).PartitionBy("author_id").OrderBy("id"), "duration").
		From(posts)

	query, args := q.Query()
	require.Equal(t, "SELECT *, (SUM(`posts`.`duration`) OVER (PARTITION BY `author_id` ORDER BY `id`)) AS `duration` FROM `posts`", query)
	require.Nil(t, args)
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
