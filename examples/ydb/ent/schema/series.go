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

// Series holds the schema definition for the Series entity.
type Series struct {
	ent.Schema
}

// Annotations of the Series.
func (Series) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "series"},
	}
}

// Fields of the Series.
func (Series) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.String("title").
			NotEmpty(),
		field.Text("info").
			Optional(),
		field.Time("release_date"),
	}
}

// Edges of the Series.
func (Series) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("seasons", Season.Type),
	}
}
