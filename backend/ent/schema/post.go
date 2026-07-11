package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	entschema "entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type Post struct{ ent.Schema }

func (Post) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").SchemaType(map[string]string{dialect.Postgres: "uuid"}).Immutable(),
		field.String("author_id").SchemaType(map[string]string{dialect.Postgres: "uuid"}).Immutable(),
		field.String("body").SchemaType(map[string]string{dialect.Postgres: "varchar(8800)"}).MaxLen(8800),
		field.Time("created_at").Immutable(),
		field.Time("updated_at"),
	}
}

func (Post) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("author", Profile.Type).Ref("posts").Field("author_id").Unique().Required().Immutable(),
		edge.To("comments", Comment.Type).StorageKey(edge.Symbol("comments_post_id_fkey")).Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("likes", PostLike.Type).StorageKey(edge.Symbol("post_likes_post_id_fkey")).Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("saves", SavedPost.Type).StorageKey(edge.Symbol("saved_posts_post_id_fkey")).Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("media", PostMedia.Type).StorageKey(edge.Symbol("post_media_post_id_fkey")).Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}

func (Post) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("created_at", "id").StorageKey("posts_feed_idx").Annotations(entsql.DescColumns("created_at", "id")),
		index.Fields("author_id").StorageKey("posts_author_id_idx"),
	}
}

func (Post) Annotations() []entschema.Annotation {
	return []entschema.Annotation{&entsql.Annotation{Checks: map[string]string{
		"posts_body_nonempty": "length(btrim(body)) > 0",
		"posts_timestamps":    "updated_at >= created_at",
	}}}
}
