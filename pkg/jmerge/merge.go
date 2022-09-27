package jmerge

import (
	"reflect"
)

const MaxDepth = 32

func Merge(dst, src map[string]interface{}) map[string]interface{} {
	return merge(dst, src, 0).(map[string]interface{})
}

func cRecursion(dst, src interface{}, depth int) interface{} {
	dv := reflect.ValueOf(dst)
	ds := reflect.ValueOf(src)
	// if this is a map or a slice on both the src and dest we need to call ourselves recursively
	if dv.Kind() == ds.Kind() && (dv.Kind() == reflect.Map || dv.Kind() == reflect.Slice) {
		return merge(dst, src, depth+1)
	}
	return src
}

func merge(dst, src interface{}, depth int) interface{} {
	if depth > MaxDepth {
		panic("too deep!")
	}

	dv := reflect.ValueOf(dst)
	sv := reflect.ValueOf(src)

	if dv.Kind() == reflect.Map && dv.Kind() == sv.Kind() {
		srcMap := toMap(sv)
		destMap := toMap(dv)
		for key, srcVal := range srcMap {
			destMap[key] = cRecursion(destMap[key], srcVal, depth)
		}
		return destMap
	}

	if dv.Kind() == reflect.Slice && dv.Kind() == sv.Kind() {
		srcSlice := toSlice(sv)
		destSlice := toSlice(dv)
		if len(srcSlice) == 0 {
			return dst
		}

		for i := 0; i < len(srcSlice); i++ {
			srcName := getName(srcSlice[i])
			// this is mapping based on a name property in the map if it exists
			if len(srcName) > 0 {
				if di := findDestMapIndexInSliceByName(srcName, destSlice); di != -1 {
					destSlice[di] = cRecursion(destSlice[di], srcSlice[i], depth)
				} else if len(destSlice) >= i+1 {
					if len(getName(destSlice[i])) > 0 {
						destSlice = append(destSlice, srcSlice[i])
					} else {
						destSlice[i] = cRecursion(destSlice[i], srcSlice[i], depth)
					}
				} else {
					destSlice = append(destSlice, srcSlice[i])
				}
				// if the name property doesn't exist we just append the src to the array
			} else {
				if len(destSlice) >= i+1 {
					destSlice[i] = cRecursion(destSlice[i], srcSlice[i], depth)
				} else {
					destSlice = append(destSlice, srcSlice[i])
				}
			}
		}
		return destSlice
	}
	return dst
}

func findDestMapIndexInSliceByName(name string, destSlice []interface{}) int {
	for i := 0; i < len(destSlice); i++ {
		if getName(destSlice[i]) == name {
			return i
		}
	}
	return -1
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
		s[i] = value.Index(i).Interface()
	}
	return s
}

func getName(value interface{}) string {
	rv := reflect.ValueOf(value)
	if rv.Kind() == reflect.Map {
		if val, ok := toMap(rv)["name"]; ok {
			return val.(string)
		} else {
			return ""
		}
	}
	return ""
}
