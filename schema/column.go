package schema

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"reflect"
	"time"

	"github.com/meilihao/layer/utils"
)

// 引入column原因, 避免多个struct 引用同一struct时, 修改了其FieldName
type Column struct {
	Schema  *Schema
	RawName string
	DBName  string
	Field   *Field
	Parent  *Field // for relationship or embedded
	IsPK    bool
}

func (c *Column) Err(s string) error {
	return fmt.Errorf("table %s column:%s: %s", c.Schema.RawName, c.RawName, s)
}
func (c *Column) ErrSet() error {
	return c.Err("can not set")
}

func (c *Column) IsVersion() bool {
	return c.Field.Version
}

func (c *Column) IsAutoCreatedAt() bool {
	return c.Field.AutoCreatedAt
}
func (c *Column) IsAutoUpdatedAt() bool {
	return c.Field.AutoUpdatedAt
}
func (c *Column) IsAutoDeletedAt() bool {
	return c.Field.AutoDeletedAt
}

func (c *Column) ConvertInteger(i int64) interface{} {
	f := c.Field
	t := f.IndirectFieldType
	if k := t.Kind(); utils.IsInts(k) {
		p := reflect.New(t)
		q := p.Elem()
		if !q.OverflowInt(i) {
			q.SetInt(i)
			if f.IsPointer {
				return p.Interface()
			}
			return q.Interface()
		}
	} else if utils.IsUints(k) {
		p := reflect.New(t)
		q := p.Elem()
		if u := uint64(i); !q.OverflowUint(u) {
			q.SetUint(u)
			if f.IsPointer {
				return p.Interface()
			}
			return q.Interface()
		}
	}
	return nil
}

func LimitTime(t time.Time, level TimeType) time.Time {
	var tmp time.Time

	switch level {
	case UnixNanosecond:
		tmp = t
	case UnixMillisecond:
		tmp = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond()/1e6, t.Location())
	default:
		tmp = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), 0, t.Location())
	}

	return tmp
}

func (c *Column) ConvertTime(t time.Time, tz *time.Location, level TimeType) interface{} {
	if tz != nil {
		t = t.In(tz)
	}

	f := c.Field
	if x := f.IndirectFieldType; x == TypeTime {
		t = LimitTime(t, level)
		if f.IsPointer {
			return &t
		} else {
			return t
		}
	} else {
		var tmp int64
		switch level {
		case UnixNanosecond:
			tmp = t.UnixNano()
		case UnixMillisecond:
			tmp = t.UnixNano() / 1e6
		default:
			tmp = t.Unix()
		}

		switch x.Kind() {
		case reflect.Int:
			if i := int(tmp); f.IsPointer {
				return &i
			} else {
				return i
			}
		case reflect.Uint:
			if i := uint(tmp); f.IsPointer {
				return &i
			} else {
				return i
			}
		case reflect.Int64:
			if i := int64(tmp); f.IsPointer {
				return &i
			} else {
				return i
			}
		case reflect.Uint64:
			if i := uint64(tmp); f.IsPointer {
				return &i
			} else {
				return i
			}
		}
	}

	return nil
}

func (c *Column) SetTime(v reflect.Value, t time.Time, tz *time.Location, level TimeType) bool {
	if tz != nil {
		t = t.In(tz)
	}

	if v, ok := c.FieldValue(v); ok {
		var i interface{} = c.ConvertTime(t, tz, level)
		f := c.Field

		if i != nil {
			if f.IsPointer {
				if v.IsNil() {
					if v.CanSet() {
						v.Set(reflect.ValueOf(i))
						return true
					} else {
						return false
					}
				}
			}
			if v.CanSet() {
				v.Set(reflect.ValueOf(i))
				return true
			}
		}
	}
	return false
}

func (c *Column) FieldValue(v reflect.Value) (reflect.Value, bool) {
	if c.Parent == nil {
		return v.Field(c.Field.StructField.Index[0]), true
	}

	f := c.Parent
	v = v.Field(f.StructField.Index[0])
	if f.IsPointer {
		if v.IsNil() {
			if v.CanSet() {
				v.Set(reflect.New(f.IndirectFieldType))
			} else {
				return v, false
			}
		}

		v = v.Elem()
	}

	return v.Field(c.Field.StructField.Index[0]), true
}

func (c *Column) ErrGet() error {
	return c.Err("can not get")
}

func (c *Column) Get(v reflect.Value) (interface{}, error) {
	if v, ok := c.FieldValue(v); ok {
		return c.convert(v)
	}

	return nil, c.ErrGet()
}
func (c *Column) convert(v reflect.Value) (interface{}, error) {
	if f := c.Field; !f.IsValuer {
		if f.IsJSON || f.IsXML {
			var b []byte
			var err error
			if v.CanInterface() {
				if f.IsJSON {
					b, err = json.Marshal(v.Interface())
				} else if f.IsXML {
					b, err = xml.Marshal(v.Interface())
				}

				return b, err
			} else {
				return nil, c.ErrGet()
			}
		}
	}
	if v.CanInterface() {
		return v.Interface(), nil
	}
	return nil, c.ErrGet()
}

func (c *Column) SetInteger(v reflect.Value, i int64) bool {
	if v, ok := c.FieldValue(v); ok {
		if f := c.Field; f.IsPointer {
			if v.IsNil() {
				if v.CanSet() {
					v.Set(reflect.New(f.IndirectFieldType))
				} else {
					return false
				}
			}
			v = v.Elem()
		}
		if v.CanSet() {
			if k := v.Kind(); utils.IsInts(k) {
				if !v.OverflowInt(i) {
					v.SetInt(i)
					return true
				}
			} else if utils.IsUints(k) {
				if u := uint64(i); !v.OverflowUint(u) {
					v.SetUint(u)
					return true
				}
			}
		}
	}
	return false
}

func (c *Column) IsZero(v reflect.Value) bool {
	if v, ok := c.FieldValue(v); ok {
		return utils.IsZero(v)
	}
	return false
}

func (c *Column) GetInteger(v reflect.Value) (int64, bool) {
	if v, ok := c.FieldValue(v); ok {
		f := c.Field
		if f.IsPointer {
			if v.IsNil() {
				return 0, false
			}
			v = v.Elem()
		}
		if k := v.Kind(); utils.IsInts(k) {
			return v.Int(), true
		} else if utils.IsUints(k) {
			return int64(v.Uint()), true
		}
	}
	return 0, false
}

func (c *Column) Scan(v reflect.Value) (interface{}, func() error, bool) {
	if v, ok := c.FieldValue(v); ok {
		if f := c.Field; !f.IsScanner {
			if f.IsJSON || f.IsXML {
				var b []byte
				return &b, func() error {
					if f.IsJSON {
						return json.Unmarshal(b, v.Addr().Interface())
					} else if f.IsXML {
						return xml.Unmarshal(b, v.Addr().Interface())
					}
					panic(false)
				}, true
			}
		}
		if v.CanAddr() {
			return v.Addr().Interface(), nil, true
		}
	}
	return nil, nil, false
}
