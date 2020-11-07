package schema

var (
	RelationshipTags = []string{TagOne2One, TagOne2Many, TagMany2One, TagMany2Many}
)

// RelationshipType relationship type
type RelationshipType string

const (
	RelOne2One   RelationshipType = RelationshipType(TagOne2One)
	RelOne2Many  RelationshipType = RelationshipType(TagOne2Many)
	RelMany2One  RelationshipType = RelationshipType(TagMany2One)
	RelMany2Many RelationshipType = RelationshipType(TagMany2Many)
)

type Relationship struct {
	Type   RelationshipType
	Schema *Schema
}

func hasRelationshipTag(tags map[string]string) string {
	for _, v := range RelationshipTags {
		if _, ok := tags[v]; ok {
			return v
		}
	}

	return ""
}
