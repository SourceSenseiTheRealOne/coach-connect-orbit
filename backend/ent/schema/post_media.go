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

type PostMedia struct{ ent.Schema }

func (PostMedia) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").SchemaType(map[string]string{dialect.Postgres: "uuid"}).Immutable(),
		field.String("post_id").SchemaType(map[string]string{dialect.Postgres: "uuid"}).Immutable(),
		field.String("object_key").SchemaType(map[string]string{dialect.Postgres: "varchar(512)"}).Immutable().MaxLen(512),
		field.String("public_url").SchemaType(map[string]string{dialect.Postgres: "varchar(2048)"}).MaxLen(2048),
		field.String("alt_text").SchemaType(map[string]string{dialect.Postgres: "varchar(240)"}).MaxLen(240),
		field.String("mime_type").SchemaType(map[string]string{dialect.Postgres: "varchar(64)"}).MaxLen(64),
		field.Int("width").SchemaType(map[string]string{dialect.Postgres: "integer"}).Positive(),
		field.Int("height").SchemaType(map[string]string{dialect.Postgres: "integer"}).Positive(),
		field.Time("created_at").Immutable(),
	}
}

func (PostMedia) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("post", Post.Type).Ref("media").Field("post_id").Unique().Required().Immutable(),
	}
}

func (PostMedia) Indexes() []ent.Index {
	return []ent.Index{index.Fields("post_id").StorageKey("post_media_post_id_idx").Unique()}
}

func (PostMedia) Annotations() []entschema.Annotation {
	return []entschema.Annotation{&entsql.Annotation{Checks: map[string]string{
		"post_media_width_positive":  "width > 0",
		"post_media_height_positive": "height > 0",
	}}}
}
