package jmerge

import (
	"reflect"
)

const MaxDepth = 32

func Merge(dst, src interface{}) interface{} {
	return merge(dst, src, 0)
}

func merge(dst, src interface{}, depth int) interface{} {
	if depth > MaxDepth {
		panic("too deep!")
	}

	cRecurse := func(dst, src interface{}, depth int) interface{} {
		dv := reflect.ValueOf(dst)
		ds := reflect.ValueOf(src)
		// if this is a map or a slice on both the src and dest we need to call ourselves recursively
		if dv.Kind() == ds.Kind() && (dv.Kind() == reflect.Map || dv.Kind() == reflect.Slice) {
			return merge(dst, src, depth+1)
		}
		return src
	}

	dv := reflect.ValueOf(dst)
	sv := reflect.ValueOf(src)

	if dv.Kind() == reflect.Map && dv.Kind() == sv.Kind() {
		srcMap := toMap(sv)
		destMap := toMap(dv)
		for key, srcVal := range srcMap {
			destMap[key] = cRecurse(destMap[key], srcVal, depth)
		}
		return destMap
	}

	if dv.Kind() == reflect.Slice && dv.Kind() == sv.Kind() {
		srcSlice := toSlice(sv)
		destSlice := toSlice(dv)
		for i := 0; i < len(srcSlice); i++ {
			if len(destSlice) >= i+1 {
				destSlice[i] = cRecurse(destSlice[i], srcSlice[i], depth)
			} else {
				destSlice = append(destSlice, srcSlice[i])
			}
		}
		return destSlice
	}

	return dst
}

func toMap(value reflect.Value) map[string]interface{} {
	m := map[string]interface{}{}
	for _, k := range value.MapKeys() {
		m[k.String()] = value.MapIndex(k).Interface()
	}
	return m
}

func toSlice(value reflect.Value) []interface{} {
	s := make([]interface{}, value.Len())
	for i := 0; i < value.Len(); i++ {
		s[i] = value.Index(i).Addr().Interface()
	}
	return s
}
