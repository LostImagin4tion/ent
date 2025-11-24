package query

import (
	"testing"

	"github.com/stretchr/testify/require"
)

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
