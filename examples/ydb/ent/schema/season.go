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

// Season holds the schema definition for the Season entity.
type Season struct {
	ent.Schema
}

// Annotations of the Season.
func (Season) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "seasons"},
	}
}

// Fields of the Season.
func (Season) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.Int64("series_id"),
		field.String("title").
			NotEmpty(),
		field.Time("first_aired"),
		field.Time("last_aired"),
	}
}

// Edges of the Season.
func (Season) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("series", Series.Type).
			Ref("seasons").
			Unique().
			Required().
			Field("series_id"),
		edge.To("episodes", Episode.Type),
	}
}
