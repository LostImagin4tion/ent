package query

import (
	"strconv"
	"testing"

	"entgo.io/ent/dialect"
	"github.com/stretchr/testify/require"
)

func TestInsert(t *testing.T) {
	tests := []struct {
		input     Querier
		wantQuery string
		wantArgs  []any
	}{
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
	}
	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			query, args := tt.input.Query()
			require.Equal(t, tt.wantQuery, query)
			require.Equal(t, tt.wantArgs, args)
		})
	}
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
