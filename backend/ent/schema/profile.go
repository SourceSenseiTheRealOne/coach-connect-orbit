package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	entschema "entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type Profile struct{ ent.Schema }

func (Profile) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").SchemaType(map[string]string{dialect.Postgres: "uuid"}).Immutable(),
		field.String("clerk_subject").SchemaType(map[string]string{dialect.Postgres: "varchar(256)"}).Unique().Immutable().MaxLen(256),
		field.String("display_name").SchemaType(map[string]string{dialect.Postgres: "varchar(80)"}).MaxLen(80),
		field.String("avatar_url").SchemaType(map[string]string{dialect.Postgres: "varchar(2048)"}).Optional().Nillable().MaxLen(2048),
		field.Time("created_at").Immutable(),
		field.Time("updated_at"),
	}
}

func (Profile) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("posts", Post.Type).StorageKey(edge.Symbol("posts_author_id_fkey")).Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("comments", Comment.Type).StorageKey(edge.Symbol("comments_author_id_fkey")).Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("likes", PostLike.Type).StorageKey(edge.Symbol("post_likes_profile_id_fkey")).Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("saved_posts", SavedPost.Type).StorageKey(edge.Symbol("saved_posts_profile_id_fkey")).Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}

func (Profile) Annotations() []entschema.Annotation {
	return []entschema.Annotation{&entsql.Annotation{Checks: map[string]string{
		"profiles_timestamps": "updated_at >= created_at",
	}}}
}
