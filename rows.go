package layer

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/meilihao/layer/schema"
	"github.com/meilihao/layer/utils"
	"github.com/rs/zerolog/log"
)

type Rows struct {
	err  error
	rows *sql.Rows
	l    *Layer
}

func (r *Rows) Close() error {
	if r.err != nil {
		return r.err
	}
	return r.rows.Close()
}

func (r *Rows) Columns() ([]string, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.rows.Columns()
}

func (r *Rows) Err() error {
	if r.err != nil {
		return r.err
	}
	return r.rows.Err()
}

func (r *Rows) Next() bool {
	if r.err != nil {
		return false
	}
	return r.rows.Next()
}

func (r *Rows) scanValue(columns []string, v reflect.Value) error {
	var err error
	if err = r.rows.Scan(v.Addr().Interface()); err != nil {
		return err
	} else {
		err = r.rows.Close()
	}

	return err
}

func (r *Rows) scanStruct(s *schema.Schema, columns []string, v reflect.Value) error {
	m := make(map[string]*schema.Column, len(s.Columns))
	for _, v := range s.Columns {
		m[r.l.opts.nameMapper.EntityMap(v.RawName)] = v
	}

	vs := make([]interface{}, len(columns))
	f := make([]func() error, len(columns))

	var ok bool
	var c *schema.Column
	for i, name := range columns {
		if c = m[name]; c == nil {
			return errors.New("column not found: " + name)
		}

		vs[i], f[i], ok = c.Scan(v)
		if !ok {
			return c.ErrSet()
		}
	}
	if err := r.rows.Scan(vs...); err != nil {
		return err
	}
	for _, i := range f {
		if i != nil {
			if err := i(); err != nil {
				return err
			}
		}
	}
	return nil
}

// Scan *T
func (r *Rows) Scan(i interface{}) error {
	if r.err != nil {
		return r.err
	}

	// v := reflect.ValueOf(i)
	// if v.Kind() != reflect.Ptr {
	// 	return errors.New("not pointer")
	// } else if v.IsNil() {
	// 	return errors.New("not nil")
	// } else if v.Elem().Kind() == reflect.Ptr {
	// 	return errors.New("a pointer to a pointer is not allowed")
	// }
	v, _ := utils.PtrValue(i)

	columns, err := r.rows.Columns()
	if err != nil {
		return err
	}

	switch v.Kind() {
	case reflect.Struct:
		s, err := schema.Parse(i, r.l.opts.nameMapper)
		if err != nil {
			return err
		}

		return r.scanStruct(s, columns, v)
	default:
		if len(columns) != 1 {
			return fmt.Errorf("no struct only support one column with one value")
		}

		return r.scanValue(columns, v)
	}
}

// One Scan and Close, be careful with sql.RawBytes.
func (r *Rows) One(i interface{}) (ok bool, err error) {
	if r.err != nil {
		return false, r.err
	}

	defer func() {
		if err = r.rows.Close(); err != nil {
			log.Error().Err(err).Send()
		}
	}()

	if r.rows.Next() {
		err = r.Scan(i)
		ok = err == nil
	} else {
		err = r.rows.Err()
	}

	return
}

func (r *Rows) allStruct(columns []string, s *schema.Schema, v reflect.Value) (err error) {
	x := v.Kind() == reflect.Map
	var mk map[string]bool
	if x {
		mk = make(map[string]bool, len(s.PrimaryColumns))
	}

	m := make(map[string]*schema.Column, len(s.Columns))
	for _, v := range s.Columns {
		m[r.l.opts.nameMapper.EntityMap(v.RawName)] = v

		if x && v.IsPK {
			mk[r.l.opts.nameMapper.EntityMap(v.RawName)] = true
		}
	}

	a := make([]*schema.Column, 0, len(columns))
	var c *schema.Column
	for _, name := range columns {
		if c = m[name]; c == nil {
			return errors.New("column not found: " + name)
		}

		if x && c.IsPK {
			delete(mk, r.l.opts.nameMapper.EntityMap(c.RawName))
		}

		a = append(a, c)
	}

	if x && len(mk) > 0 {
		l := make([]string, 0, len(mk))
		for k := range mk {
			l = append(l, k)
		}

		sort.Strings(l)

		return fmt.Errorf("primary key column not exist: %v", l)
	}

	y := v.Type().Elem().Kind() == reflect.Ptr

	b := make([]interface{}, len(a))
	f := make([]func() error, len(a))
	for r.rows.Next() {
		p := reflect.New(s.ModelType)
		q := p.Elem()
		var ok bool
		for i, c := range a {
			b[i], f[i], ok = c.Scan(q)
			if !ok {
				return c.ErrSet()
			}
		}
		if err = r.rows.Scan(b...); err != nil {
			break
		}
		for _, i := range f {
			if i != nil {
				if err = i(); err != nil {
					break
				}
			}
		}

		if x {
			var ok bool
			var k reflect.Value
			if len(s.PrimaryColumns) == 1 {
				k, ok = s.PrimaryColumns[0].FieldValue(q)

				if !ok {
					err = c.ErrGet()
					break
				}
			} else { // multi key
				ks := make([]interface{}, 0, len(s.PrimaryColumns))
				fs := make([]string, 0, len(s.PrimaryColumns))
				for _, c := range s.PrimaryColumns {
					k, ok = c.FieldValue(q)

					if !ok {
						err = c.ErrGet()
						break
					}

					ks = append(ks, k)
					fs = append(fs, "%v")
				}

				k = reflect.ValueOf(fmt.Sprintf(strings.Join(fs, "."), ks...))
			}

			if v.MapIndex(k).IsValid() {
				err = fmt.Errorf("duplicate key: %v", interface{}(k))
				break
			}

			if y {
				v.SetMapIndex(k, p)
			} else {
				v.SetMapIndex(k, q)
			}
		} else if y {
			v.Set(reflect.Append(v, p))
		} else {
			v.Set(reflect.Append(v, q))
		}
	}
	if err == nil {
		err = r.Err()
	}
	if err == nil {
		err = r.Close()
	}
	return err
}

func (r *Rows) allValue(columns []string, v reflect.Value) (err error) {
	x := v.Type().Elem()
	y := v.Type().Elem().Kind() == reflect.Ptr
	if y {
		x = x.Elem()
	}

	for r.rows.Next() {
		p := reflect.New(x)
		q := p.Elem()

		if err = r.rows.Scan(p.Interface()); err != nil {
			break
		}

		if y {
			v.Set(reflect.Append(v, p))
		} else {
			v.Set(reflect.Append(v, q))
		}
	}
	if err == nil {
		err = r.Err()
	}
	if err == nil {
		err = r.Close()
	}
	return err
}

// All *[]T, map[PK]T
// Key PKs -> string
func (r *Rows) All(i interface{}) error {
	if r.err != nil {
		return r.err
	}

	var err error
	defer func() {
		if err = r.rows.Close(); err != nil {
			log.Error().Err(err).Send()
		}
	}()

	v, p := utils.PtrValue(i)
	columns, err := r.Columns()
	if err != nil {
		return err
	}

	switch v.Kind() {
	case reflect.Map:
		s, err := schema.Parse(i, r.l.opts.nameMapper)
		if err != nil {
			return err
		}

		if len(s.PrimaryColumns) == 0 {
			return schema.ErrNoPK
		}

		var kt reflect.Type
		if len(s.PrimaryColumns) == 1 {
			kt = s.PrimaryColumns[0].Field.FieldType
		} else {
			kt = schema.TypeString
		}

		if v.IsNil() {
			if p {
				v.Set(reflect.MakeMap(kt))
			} else {
				panic("nil map")
			}
		}

		return r.allStruct(columns, s, v)
	case reflect.Slice:
		if !p {
			panic("not pointer")
		}

		t := v.Type().Elem()
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}

		switch t.Kind() {
		case reflect.Struct:
			s, err := schema.Parse(i, r.l.opts.nameMapper)
			if err != nil {
				return err
			}

			return r.allStruct(columns, s, v)
		default:
			if len(columns) != 1 {
				return fmt.Errorf("no struct only support one column")
			}

			return r.allValue(columns, v)
		}
	}

	panic(schema.ErrUnsupportedType)
}
