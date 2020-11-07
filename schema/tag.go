package schema

import (
	"reflect"

	"github.com/meilihao/layer/utils"
)

const (
	TagName      = "name"
	TagPK        = "pk"
	TagAutoIncr  = "autoincr"
	TagDefault   = "default"
	TagSize      = "size"
	TagPrecision = "precision"
	TagNotNull   = "notnull"
	TagUnique    = "unique"
	TagIndex     = "index"
	TagComment   = "comment"
	TagCreatedAt = "created_at"
	TagUpdatedAt = "updated_at"
	TagDeletedAt = "deleted_at"
	TagType      = "type"
	TagVersion   = "version"
	TagEmbedded  = "embedded"
	TagOne2One   = "one2one"
	TagOne2Many  = "one2many"
	TagMany2One  = "many2one" // Foreign Key
	TagMany2Many = "many2many"
	TagJSON      = "json"
	TagXML       = "xml"
)

const (
	ConflictGroupIndex        = "index"
	ConflictGroupAuto         = "auto"
	ConflictGroupTime         = "time"
	ConflictGroupOthers       = "others"
	ConflictGroupRelationship = "relationship"
	ConflictGroupEncoding     = "encoding"
)

// TagChecker check tag
type TagChecker struct {
	Name          string
	ConflictGroup []string
	CheckFn       func(reflect.Type) bool
}

var TagCheckers = map[string]*TagChecker{
	TagName: &TagChecker{
		Name: TagName,
	},
	TagPK: &TagChecker{
		Name:          TagPK,
		ConflictGroup: []string{ConflictGroupIndex},
	},
	TagAutoIncr: &TagChecker{
		Name:          TagAutoIncr,
		ConflictGroup: []string{ConflictGroupAuto},
		CheckFn:       utils.IsIntegers,
	},
	TagDefault: &TagChecker{
		Name: TagDefault,
	},
	TagSize: &TagChecker{
		Name: TagSize,
	},
	TagPrecision: &TagChecker{
		Name: TagPrecision,
	},
	TagNotNull: &TagChecker{
		Name: TagNotNull,
	},
	TagUnique: &TagChecker{
		Name:          TagUnique,
		ConflictGroup: []string{ConflictGroupIndex},
	},
	TagIndex: &TagChecker{
		Name:          TagIndex,
		ConflictGroup: []string{ConflictGroupIndex},
	},
	TagComment: &TagChecker{
		Name: TagComment,
	},
	TagCreatedAt: &TagChecker{
		Name:          TagCreatedAt,
		ConflictGroup: []string{ConflictGroupTime, ConflictGroupAuto},
		CheckFn:       utils.IsTimes,
	},
	TagUpdatedAt: &TagChecker{
		Name:          TagUpdatedAt,
		ConflictGroup: []string{ConflictGroupTime, ConflictGroupAuto},
		CheckFn:       utils.IsTimes,
	},
	TagDeletedAt: &TagChecker{
		Name:          TagDeletedAt,
		ConflictGroup: []string{ConflictGroupTime, ConflictGroupAuto},
		CheckFn:       utils.IsTimes,
	},
	TagType: &TagChecker{
		Name: TagType,
	},
	TagVersion: &TagChecker{
		Name:          TagVersion,
		ConflictGroup: []string{ConflictGroupAuto},
		CheckFn:       utils.IsIntegers,
	},
	TagEmbedded: &TagChecker{
		Name:          TagEmbedded,
		ConflictGroup: []string{ConflictGroupOthers},
	},
	TagOne2One: &TagChecker{
		Name:          TagOne2One,
		ConflictGroup: []string{ConflictGroupRelationship},
		CheckFn:       utils.IsStruct,
	},
	TagOne2Many: &TagChecker{
		Name:          TagOne2Many,
		ConflictGroup: []string{ConflictGroupRelationship},
		CheckFn:       utils.IsStructs,
	},
	TagMany2One: &TagChecker{
		Name:          TagMany2One,
		ConflictGroup: []string{ConflictGroupRelationship},
		CheckFn:       utils.IsStruct,
	},
	TagMany2Many: &TagChecker{
		Name:          TagMany2Many,
		ConflictGroup: []string{ConflictGroupRelationship},
		CheckFn:       utils.IsStructs,
	},
	TagJSON: &TagChecker{
		Name:          TagJSON,
		ConflictGroup: []string{ConflictGroupEncoding},
	},
	TagXML: &TagChecker{
		Name:          TagXML,
		ConflictGroup: []string{ConflictGroupEncoding},
	},
}
