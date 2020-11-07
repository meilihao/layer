package schema

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/meilihao/layer/utils"
)

const (
	DefaultTagIdentifier = "layer"
)

// Field
type Field struct {
	Name                  string
	RawName               string
	FieldType             reflect.Type
	IndirectFieldType     reflect.Type
	StructField           reflect.StructField
	DataType              DataType
	IsScanner             bool
	IsValuer              bool
	IsPointer             bool
	IsJSON                bool
	IsXML                 bool
	PrimaryKey            bool
	AutoIncr              bool
	NotNull               bool
	Unique                bool
	Version               bool
	HasDefaultValue       bool
	DefaultValue          string
	DefaultValueInterface interface{}
	Size                  int
	Precision             int
	Comment               string
	Tag                   reflect.StructTag
	TagSettings           map[string]string
	AutoCreatedAt         bool
	AutoUpdatedAt         bool
	AutoDeletedAt         bool
	TimeLevel             TimeType
	Schema                *Schema
	EmbeddedSchema        *Schema
	EmbeddedPrefix        string
	Relationship          *Relationship
}

func (schema *Schema) ParseField(fieldStruct reflect.StructField, namer NameMapper, level int) (*Field, error) {
	var err error

	field := &Field{
		Name:              fieldStruct.Name,
		RawName:           fieldStruct.Name,
		FieldType:         fieldStruct.Type,
		IndirectFieldType: fieldStruct.Type,
		StructField:       fieldStruct,
		Tag:               fieldStruct.Tag,
		TagSettings:       ParseTag(fieldStruct.Tag.Get(DefaultTagIdentifier)),
		Schema:            schema,
	}

	if v, ok := field.TagSettings[TagName]; ok {
		if v == "-" {
			field.RawName = ""

			return field, nil
		}

		if v != "" {
			field.RawName = v
		}
	}

	if field.FieldType.Implements(TypeScanner) {
		field.IsScanner = true
	}
	if field.FieldType.Implements(TypeValuer) {
		field.IsValuer = true
	}
	if field.FieldType.Kind() == reflect.Ptr {
		field.IsPointer = true
		field.IndirectFieldType = field.IndirectFieldType.Elem()
	}

	cg := make(map[string]string, 3) // conflict group check
	var checker *TagChecker
	for tag := range field.TagSettings {
		if tag == TagName {
			continue
		}

		checker = TagCheckers[tag]
		if checker == nil {
			return nil, fmt.Errorf("%w: %s", ErrUnsupportedTag, tag)
		}

		if checker.CheckFn != nil && !checker.CheckFn(field.IndirectFieldType) {
			return nil, fmt.Errorf("%w: %s", ErrTypeMismatchTag, tag)
		}

		for _, g := range checker.ConflictGroup {
			if v, ok := cg[g]; ok {
				return nil, fmt.Errorf("%w: (%s:%s)", ErrTagConflict, v, tag)
			} else {
				cg[g] = tag
			}
		}

		if v, ok := cg[ConflictGroupOthers]; ok && len(cg) > 1 {
			return nil, fmt.Errorf("%w: %s", ErrTagConflictOthers, v)
		}

		if tag == TagJSON {
			field.IsJSON = true

			continue
		}

		if tag == TagXML {
			field.IsXML = true

			continue
		}
	}

	fieldValue := reflect.New(field.IndirectFieldType)
	if _, ok := field.TagSettings[TagPK]; ok {
		field.PrimaryKey = true
	}

	if _, ok := field.TagSettings[TagAutoIncr]; ok {
		field.AutoIncr = true

		if !field.PrimaryKey {
			return nil, fmt.Errorf("%w, %s", ErrAutoIncrWithPK, field.Name)
		}
	}

	if v, ok := field.TagSettings[TagDefault]; ok {
		field.HasDefaultValue = true
		field.DefaultValue = v
	}

	if v, ok := field.TagSettings[TagSize]; ok {
		if field.Size, err = strconv.Atoi(v); err != nil {
			field.Size = -1
		}
	}

	if v, ok := field.TagSettings[TagPrecision]; ok {
		field.Precision, _ = strconv.Atoi(v)
	}

	if _, ok := field.TagSettings[TagNotNull]; ok {
		field.NotNull = true
	}

	if _, ok := field.TagSettings[TagUnique]; ok {
		field.Unique = true
	}

	if v, ok := field.TagSettings[TagComment]; ok {
		field.Comment = v
	}

	// default value is function(`default=uuid_generate_v4()`) or null or blank (primary keys)
	skipParseDefaultValue := strings.Contains(field.DefaultValue, "(") &&
		strings.Contains(field.DefaultValue, ")") || strings.ToLower(field.DefaultValue) == "null" || field.DefaultValue == ""
	switch reflect.Indirect(fieldValue).Kind() {
	case reflect.Bool:
		field.DataType = Bool
		if field.HasDefaultValue && !skipParseDefaultValue {
			if field.DefaultValueInterface, err = strconv.ParseBool(field.DefaultValue); err != nil {
				return nil, fmt.Errorf("failed to parse %v as default value for bool, got error: %v", field.DefaultValue, err)
			}
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		field.DataType = Int
		if field.HasDefaultValue && !skipParseDefaultValue {
			if field.DefaultValueInterface, err = strconv.ParseInt(field.DefaultValue, 0, 64); err != nil {
				return nil, fmt.Errorf("failed to parse %v as default value for int, got error: %v", field.DefaultValue, err)
			}
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		field.DataType = Uint
		if field.HasDefaultValue && !skipParseDefaultValue {
			if field.DefaultValueInterface, err = strconv.ParseUint(field.DefaultValue, 0, 64); err != nil {
				return nil, fmt.Errorf("failed to parse %v as default value for uint, got error: %v", field.DefaultValue, err)
			}
		}
	case reflect.Float32, reflect.Float64:
		field.DataType = Float
		if field.HasDefaultValue && !skipParseDefaultValue {
			if field.DefaultValueInterface, err = strconv.ParseFloat(field.DefaultValue, 64); err != nil {
				return nil, fmt.Errorf("failed to parse %v as default value for float, got error: %v", field.DefaultValue, err)
			}
		}
	case reflect.String:
		field.DataType = String

		if field.HasDefaultValue && !skipParseDefaultValue {
			field.DefaultValue = strings.Trim(field.DefaultValue, "'")
			field.DefaultValue = strings.Trim(field.DefaultValue, "\"")
			field.DefaultValueInterface = field.DefaultValue
		}
	case reflect.Struct:
		if _, ok := fieldValue.Interface().(*time.Time); ok {
			field.DataType = Time
		} else if fieldValue.Type().ConvertibleTo(TypeTime) {
			field.DataType = Time
		} else if fieldValue.Type().ConvertibleTo(reflect.TypeOf(&time.Time{})) {
			field.DataType = Time
		}
	case reflect.Array, reflect.Slice:
		if reflect.Indirect(fieldValue).Type().Elem() == reflect.TypeOf(uint8(0)) {
			field.DataType = Bytes
		}
	}

	if v := field.TagSettings[TagType]; v != "" {
		switch DataType(v) {
		case Bool, Int, Uint, Float, String, Time, Bytes:
			field.DataType = DataType(v)
		default:
			field.DataType = DataType(v)
		}
	}

	if _, ok := field.TagSettings[TagVersion]; ok {
		field.Version = true
		field.DataType = DataType(field.IndirectFieldType.Name())
	}

	for _, tag := range []string{TagCreatedAt, TagUpdatedAt, TagDeletedAt} {
		if v, ok := field.TagSettings[tag]; ok {
			if v == "nano" {
				field.TimeLevel = UnixNanosecond
			} else if v == "milli" {
				field.TimeLevel = UnixMillisecond
			} else {
				field.TimeLevel = UnixSecond
			}

			switch tag {
			case TagCreatedAt:
				field.AutoCreatedAt = true
			case TagUpdatedAt:
				field.AutoUpdatedAt = true
			default:
				field.AutoDeletedAt = true
			}
		}

	}

	if field.Size == 0 {
		switch reflect.Indirect(fieldValue).Kind() {
		case reflect.Int, reflect.Int64, reflect.Uint, reflect.Uint64, reflect.Float64:
			field.Size = 64
		case reflect.Int8, reflect.Uint8:
			field.Size = 8
		case reflect.Int16, reflect.Uint16:
			field.Size = 16
		case reflect.Int32, reflect.Uint32, reflect.Float32:
			field.Size = 32
		}
	}

	if prefix, ok := field.TagSettings[TagEmbedded]; ok || (fieldStruct.Anonymous && !field.IsValuer) {
		if field.IndirectFieldType.Kind() != reflect.Struct {
			return nil, fmt.Errorf("invalid embedded struct for %v's field %v, should be struct, but got %v", field.Schema.Name, field.Name, field.FieldType)
		}

		if prefix != "" {
			field.EmbeddedPrefix = prefix
		}

		if s, ok := SchemaCache.Store[field.IndirectFieldType]; ok {
			field.EmbeddedSchema = s
		} else if field.EmbeddedSchema, err = parse(field.IndirectFieldType, namer, level+1); err != nil {
			return nil, err
		}
	}

	if relationship := hasRelationshipTag(field.TagSettings); relationship != "" {
		field.Relationship = &Relationship{
			Type: RelationshipType(relationship),
		}

		if s, ok := SchemaCache.Store[RelationshipFieldStruct(field.IndirectFieldType)]; ok {
			field.Relationship.Schema = s
		} else if field.Relationship.Schema, err = parse(RelationshipFieldStruct(field.IndirectFieldType), namer, level+1); err != nil {
			return nil, err
		}
	}

	return field, nil
}

// ParseTag parse struct tag
func ParseTag(raw string) map[string]string {
	settings := map[string]string{}

	raw = strings.TrimSpace(raw)
	if raw == "" {
		settings[TagName] = ""
		return settings
	}

	if !strings.Contains(raw, ";") {
		settings[TagName] = raw
		return settings
	}

	units := strings.Split(strings.TrimSpace(raw), ";")
	settings[TagName] = strings.TrimSpace(units[0])

	for i := 1; i < len(units); i++ {
		units[i] = strings.TrimSpace(units[i])
		if units[i] == "" {
			continue
		}

		subs := strings.Split(units[i], "=")
		subs[0] = strings.TrimSpace(subs[0])
		if len(subs) >= 2 {
			settings[subs[0]] = strings.TrimSpace(subs[1])
		} else {
			settings[subs[0]] = ""
		}
	}

	return settings
}

func RelationshipFieldStruct(t reflect.Type) reflect.Type {
	if utils.IsMapOrSlice(t.Kind()) {
		t = t.Elem()
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
	}
	return t
}
