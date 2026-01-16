// Copyright 2019-present Facebook Inc. All rights reserved.
// This source code is licensed under the Apache 2.0 license found
// in the LICENSE file in the root directory of this source tree.

package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Episode holds the schema definition for the Episode entity.
type Episode struct {
	ent.Schema
}

// Annotations of the Episode.
func (Episode) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "episodes"},
	}
}

// Fields of the Episode.
func (Episode) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.Int64("season_id"),
		field.String("title").
			NotEmpty(),
		field.Time("air_date"),
	}
}

// Edges of the Episode.
func (Episode) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("season", Season.Type).
			Ref("episodes").
			Unique().
			Required().
			Field("season_id"),
	}
}
