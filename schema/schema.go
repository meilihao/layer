package schema

import (
	"errors"
	"fmt"
	"go/ast"
	"reflect"
	"sync"
)

var (
	// ErrUnsupportedType unsupported data type
	ErrUnsupportedType          = errors.New("unsupported type")
	ErrUnsupportedTag           = errors.New("unsupported tag")
	ErrTypeMismatchTag          = errors.New("type mismatch tag")
	ErrTagConflict              = errors.New("tag conflict")
	ErrTagConflictOthers        = errors.New("tag conflict others")
	ErrDuplicateColumn          = errors.New("duplicate column")
	ErrDuplicateAuotIncrColumn  = errors.New("duplicate autoincr column")
	ErrDuplicateVersionColumn   = errors.New("duplicate version column")
	ErrDuplicateDeletedAtColumn = errors.New("duplicate deletedat column")
	ErrDuplicateUpdatedAtColumn = errors.New("duplicate updatedat column")
	ErrParseLoop                = errors.New("parse loop")
	ErrAutoIncrWithPK           = errors.New("autoincr need with pk")
	ErrNoPK                     = errors.New("no primary key")
)

// Cache cache parsed schema
// 无法使用以TypePath()为key的map[string]*Schema的原因: struct作为字段时, 无论匿名与否, t.PkgPath()为空
type Cache struct {
	sync.RWMutex
	Store map[reflect.Type]*Schema
}

// NewCache generate Cache
func NewCache() *Cache {
	return &Cache{
		Store: make(map[reflect.Type]*Schema, 5),
	}
}

// Schema strcut's model
type Schema struct {
	Name             string
	RawName          string // raw dbnanme without mapping
	DBName           string
	ModelType        reflect.Type
	Columns          []*Column
	ColumnsByRawName map[string]*Column
	PrimaryColumns   []*Column
	AutoincrColumn   *Column
	Version          *Column
	DeletedAt        *Column
	UpdatedAt        *Column
}

// TypePath get pkgpath.typename
func TypePath(path, name string) string {
	return fmt.Sprintf("%s.%s", path, name)
}

// Tabler get tablename
type Tabler interface {
	TableName() string
}

var (
	SchemaCache = NewCache()
)

func Parse(dest interface{}, namer NameMapper) (*Schema, error) {
	if dest == nil {
		return nil, fmt.Errorf("%w: %+v", ErrUnsupportedType, dest)
	}

	t := reflect.TypeOf(dest)
	k := t.Kind()
	f1 := func() {
		t = t.Elem()
		k = t.Kind()
	}
	f2 := func() {
		if k == reflect.Ptr {
			f1()
		}
	}
	switch f2(); k {
	case reflect.Map:
		f1()
		f2()
	case reflect.Slice:
		f1()
		f2()
	}

	if k != reflect.Struct {
		if t.PkgPath() == "" {
			return nil, fmt.Errorf("%w: %+v", ErrUnsupportedType, dest)
		}
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedType, TypePath(t.PkgPath(), t.Name()))
	}

	SchemaCache.RLock()
	s, ok := SchemaCache.Store[t]
	SchemaCache.RUnlock()
	if !ok {
		SchemaCache.Lock()
		defer SchemaCache.Unlock()
		s, ok = SchemaCache.Store[t]
		if ok {
			return s, nil
		}

		return parse(t, namer, 1)
	}
	return s, nil
}

func parse(t reflect.Type, namer NameMapper, level int) (*Schema, error) {
	s := &Schema{
		Name:             t.Name(),
		RawName:          t.Name(),
		ModelType:        t,
		Columns:          []*Column{},
		ColumnsByRawName: map[string]*Column{},
		PrimaryColumns:   []*Column{},
	}

	// (*Schema).parse()前放入cache避免递归解析时因无法找到而出现死循环
	SchemaCache.Store[s.ModelType] = s

	if err := s.parse(namer, level); err != nil {
		delete(SchemaCache.Store, s.ModelType)
		return nil, err
	}

	return s, nil
}

// Parse parse model
func (schema *Schema) parse(namer NameMapper, level int) error {
	if level > 2 {
		return fmt.Errorf("%w: %+v", ErrParseLoop, schema.ModelType)
	}

	modelValue := reflect.New(schema.ModelType)
	if tabler, ok := modelValue.Interface().(Tabler); ok {
		schema.RawName = tabler.TableName()
	}
	schema.DBName = namer.EntityMap(schema.RawName)

	for i := 0; i < schema.ModelType.NumField(); i++ {
		if fieldStruct := schema.ModelType.Field(i); ast.IsExported(fieldStruct.Name) {
			field, err := schema.ParseField(fieldStruct, namer, level)
			if err != nil {
				return err
			}

			if field.EmbeddedSchema != nil {
				for _, ec := range field.EmbeddedSchema.Columns {
					for k, v := range field.TagSettings {
						ec.Field.TagSettings[k] = v
					}

					if err = schema.AddDBColumn(field.EmbeddedPrefix, field, ec.Field, namer, true); err != nil {
						return err
					}
				}
			} else if field.Relationship != nil {
				switch field.Relationship.Type {
				case RelOne2One, RelMany2One:
					for _, rc := range field.Relationship.Schema.PrimaryColumns {
						if err = schema.AddDBColumn(field.RawName, field, rc.Field, namer, false); err != nil {
							return err
						}
					}
				case RelOne2Many:
				case RelMany2Many:
				}
			} else {
				if err = schema.AddDBColumn("", nil, field, namer, true); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (schema *Schema) AddDBColumn(prefix string, parent, field *Field, namer NameMapper, canPK bool) error {
	if field.RawName != "" {
		c := &Column{
			Schema:  schema,
			RawName: field.RawName,
			Field:   field,
			Parent:  parent,
		}

		if prefix != "" {
			c.RawName = prefix + field.RawName
		}
		c.DBName = namer.EntityMap(c.RawName)

		if _, ok := schema.ColumnsByRawName[c.DBName]; ok {
			return fmt.Errorf("%w: %s", ErrDuplicateColumn, c.RawName)
		}

		schema.Columns = append(schema.Columns, c)
		schema.ColumnsByRawName[c.RawName] = c

		if field.PrimaryKey && canPK {
			c.IsPK = true
			schema.PrimaryColumns = append(schema.PrimaryColumns, c)

			if field.AutoIncr {
				if schema.AutoincrColumn == nil {
					schema.AutoincrColumn = c
				} else {
					return fmt.Errorf("%w : %s", ErrDuplicateAuotIncrColumn, c.RawName)
				}
			}
		}
		if field.Version {
			if schema.Version == nil {
				schema.Version = c
			} else {
				return fmt.Errorf("%w : %s", ErrDuplicateVersionColumn, c.RawName)
			}
		}
		if field.AutoDeletedAt {
			if schema.DeletedAt == nil {
				schema.DeletedAt = c
			} else {
				return fmt.Errorf("%w : %s", ErrDuplicateDeletedAtColumn, c.RawName)
			}
		}
		if field.AutoUpdatedAt {
			if schema.UpdatedAt == nil {
				schema.UpdatedAt = c
			} else {
				return fmt.Errorf("%w : %s", ErrDuplicateUpdatedAtColumn, c.RawName)
			}
		}
	}

	//field.setupValuerAndSetter()

	return nil
}
