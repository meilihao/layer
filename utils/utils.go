package utils

import (
	"reflect"
	"strconv"
	"time"
)

func IsInts(k reflect.Kind) bool {
	switch k {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return true
	}
	return false
}

func IsUints(k reflect.Kind) bool {
	switch k {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return true
	}
	return false
}

func IsIntegers(t reflect.Type) bool {
	k := t.Kind()
	return IsInts(k) || IsUints(k)
}

func IsTimes(t reflect.Type) bool {
	k := t.Kind()
	return k == reflect.Int64 || k == reflect.Int || t == reflect.TypeOf(time.Time{})
}

func IsStruct(t reflect.Type) bool {
	return t.Kind() == reflect.Struct
}

func IsMapOrSlice(k reflect.Kind) bool {
	return k == reflect.Map || k == reflect.Slice
}

func IsStructs(t reflect.Type) bool {
	if IsMapOrSlice(t.Kind()) {
		t = t.Elem()
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
		return IsStruct(t)
	}
	return false
}

func PtrValue(i interface{}) (v reflect.Value, isPointer bool) {
	if i == nil {
		panic("nil value")
	}

	v = reflect.ValueOf(i)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			panic("nil pointer")
		} else {
			v = v.Elem()
			isPointer = true
		}
	}

	return
}

func IsZero(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Ptr:
		return v.IsNil()
	}
	return v.IsValid() && v.Type() == reflect.TypeOf(time.Time{}) && v.Interface().(time.Time).IsZero()
}

func ToString(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case int:
		return strconv.FormatInt(int64(v), 10)
	case int8:
		return strconv.FormatInt(int64(v), 10)
	case int16:
		return strconv.FormatInt(int64(v), 10)
	case int32:
		return strconv.FormatInt(int64(v), 10)
	case int64:
		return strconv.FormatInt(v, 10)
	case uint:
		return strconv.FormatUint(uint64(v), 10)
	case uint8:
		return strconv.FormatUint(uint64(v), 10)
	case uint16:
		return strconv.FormatUint(uint64(v), 10)
	case uint32:
		return strconv.FormatUint(uint64(v), 10)
	case uint64:
		return strconv.FormatUint(v, 10)
	}
	return ""
}
