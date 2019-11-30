package gobox

import (
	"fmt"
	"reflect"
	"time"
)

func Mask(src interface{}) interface{} {
	if src == nil {
		return nil
	}
	original := reflect.ValueOf(src)
	cpy := reflect.New(original.Type()).Elem()
	maskRecursive(original, cpy)
	return cpy.Interface()
}

func maskRecursive(src, dst reflect.Value) {
	switch src.Kind() {
	case reflect.Ptr:
		originalValue := src.Elem()
		if !originalValue.IsValid() {
			return
		}
		dst.Set(reflect.New(originalValue.Type()))
		maskRecursive(originalValue, dst.Elem())

	case reflect.Interface:
		if src.IsNil() {
			return
		}
		originalValue := src.Elem()
		copyValue := reflect.New(originalValue.Type()).Elem()
		maskRecursive(originalValue, copyValue)
		dst.Set(copyValue)

	case reflect.Struct:
		t, ok := src.Interface().(time.Time)
		if ok {
			dst.Set(reflect.ValueOf(t))
			return
		}
		for i := 0; i < src.NumField(); i++ {
			if src.Type().Field(i).PkgPath != "" {
				continue
			}
			if src.Type().Field(i).Tag.Get("mask") != "" {
				if src.Field(i).Kind() == reflect.String {
					res := fmt.Sprintf("%s+%s", src.Field(i), "mask")
					dst.Field(i).SetString(res)
					continue
				}
				if src.Field(i).Kind() == reflect.Ptr {
					if reflect.Indirect(src.Field(i)).Kind() == reflect.String {
						res := fmt.Sprintf("%s+%s", src.Field(i).Elem(), "mask")
						dst.Field(i).Set(reflect.ValueOf(&res))
						continue
					}
				}
			}
			maskRecursive(src.Field(i), dst.Field(i))
		}

	case reflect.Slice:
		if src.IsNil() {
			return
		}
		dst.Set(reflect.MakeSlice(src.Type(), src.Len(), src.Cap()))
		for i := 0; i < src.Len(); i++ {
			maskRecursive(src.Index(i), dst.Index(i))
		}

	case reflect.Map:
		if src.IsNil() {
			return
		}
		dst.Set(reflect.MakeMap(src.Type()))
		for _, key := range src.MapKeys() {
			originalValue := src.MapIndex(key)
			copyValue := reflect.New(originalValue.Type()).Elem()
			maskRecursive(originalValue, copyValue)
			copyKey := Copy(key.Interface())
			dst.SetMapIndex(reflect.ValueOf(copyKey), copyValue)
		}
	default:
		dst.Set(src)
	}
}
