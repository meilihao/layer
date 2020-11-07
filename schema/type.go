package schema

import (
	"database/sql"
	"database/sql/driver"
	"reflect"
	"time"
)

var (
	TypeTime      = reflect.TypeOf(time.Time{})
	TypeEmpty     = reflect.TypeOf(struct{}{})
	TypeScanner   = reflect.TypeOf((*sql.Scanner)(nil)).Elem()
	TypeValuer    = reflect.TypeOf((*driver.Valuer)(nil)).Elem()
	TypeString    = reflect.TypeOf("")
	TypeInterface = reflect.TypeOf(([]interface{})(nil)).Elem()
)

var (
	ZeroEmpty = reflect.Zero(TypeEmpty)
)

type DataType string

const (
	Bool   DataType = "bool"
	Int    DataType = "int"
	Uint   DataType = "uint"
	Float  DataType = "float"
	String DataType = "string"
	Time   DataType = "time"
	Bytes  DataType = "bytes"
)

type TimeType int64

const (
	UnixSecond      TimeType = 1
	UnixMillisecond TimeType = 2
	UnixNanosecond  TimeType = 3
)
