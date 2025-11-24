// Copyright 2019-present Facebook Inc. All rights reserved.
// This source code is licensed under the Apache 2.0 license found
// in the LICENSE file in the root directory of this source tree.

package query

import (
	"context"
	"database/sql/driver"
	"fmt"
	"testing"

	"entgo.io/ent/dialect"
	"github.com/stretchr/testify/require"
)

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
