package schema

import (
	. "github.com/tarqeem/ims/translate"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"

	. "github.com/tarqeem/template/utl/ent"
)

// Project holds the schema definition for the Project entity.
type Project struct {
	ent.Schema
}

// Fields of the Project.
func (Project) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").Unique(),
		field.String("owner"),
		field.String("location"),
		field.Enum("type").Values(ValidProjectTypes...),
		field.Enum("project_nature").Values("Greenfield", "Brownfield"),
		StringOneOf("delivery_strategies", ValidProjectDeliveryStrategies),
		StringOneOf("state", ValidProjectStates),
		StringOneOf("contracting_strategies", ValidProjectContractingStrategies),
		field.Int("dollar_value").NonNegative(),
		field.String("execution_location"),
		field.Int("tlsp").NonNegative().Default(0),
		field.Int("jvp").NonNegative().Default(0),
		field.Int("ish").NonNegative().Default(0),
	}
}

// Edges of the Project.
func (Project) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("leader", User.Type).
			Ref("leader_of_project"),
		edge.From("coordinator", User.Type).
			Ref("coordinator_of_project"),
		edge.To("members", User.Type),
		edge.To("issues", Issue.Type),
	}
}
