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

type Comment struct{ ent.Schema }

func (Comment) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").SchemaType(map[string]string{dialect.Postgres: "uuid"}).Immutable(),
		field.String("post_id").SchemaType(map[string]string{dialect.Postgres: "uuid"}).Immutable(),
		field.String("author_id").SchemaType(map[string]string{dialect.Postgres: "uuid"}).Immutable(),
		field.String("body").SchemaType(map[string]string{dialect.Postgres: "varchar(8800)"}).MaxLen(8800),
		field.Time("created_at").Immutable(),
		field.Time("updated_at"),
	}
}

func (Comment) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("post", Post.Type).Ref("comments").Field("post_id").Unique().Required().Immutable(),
		edge.From("author", Profile.Type).Ref("comments").Field("author_id").Unique().Required().Immutable(),
	}
}

func (Comment) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("post_id", "created_at", "id").StorageKey("comments_post_feed_idx").Annotations(entsql.DescColumns("created_at", "id")),
		index.Fields("author_id").StorageKey("comments_author_id_idx"),
	}
}

func (Comment) Annotations() []entschema.Annotation {
	return []entschema.Annotation{&entsql.Annotation{Checks: map[string]string{
		"comments_body_nonempty": "length(btrim(body)) > 0",
		"comments_timestamps":    "updated_at >= created_at",
	}}}
}
