package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type PostLike struct{ ent.Schema }

func (PostLike) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").SchemaType(map[string]string{dialect.Postgres: "uuid"}).Immutable(),
		field.String("post_id").SchemaType(map[string]string{dialect.Postgres: "uuid"}).Immutable(),
		field.String("profile_id").SchemaType(map[string]string{dialect.Postgres: "uuid"}).Immutable(),
		field.Time("created_at").Immutable(),
	}
}

func (PostLike) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("post", Post.Type).Ref("likes").Field("post_id").Unique().Required().Immutable(),
		edge.From("profile", Profile.Type).Ref("likes").Field("profile_id").Unique().Required().Immutable(),
	}
}

func (PostLike) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("post_id", "profile_id").StorageKey("post_likes_post_profile_unique_idx").Unique(),
		index.Fields("profile_id", "created_at").StorageKey("post_likes_profile_idx").Annotations(entsql.DescColumns("created_at")),
	}
}
