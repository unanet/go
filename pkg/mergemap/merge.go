package mergemap

import "reflect"

const MaxDepth = 32

func Merge(dst, src interface{}) interface{} {
	return merge(dst, src, 0)
}

func mergeMap(dst, src map[string]interface{}, depth int) map[string]interface{} {
	if depth > MaxDepth {
		panic("too deep!")
	}

	for key, srcVal := range src {
		if dstVal, ok := dst[key]; ok {
			srcMap, srcMapOk := toMap(srcVal)
			dstMap, dstMapOk := mapify(dstVal)
			if srcMapOk && dstMapOk {
				srcVal = mergeMap(dstMap, srcMap, depth+1)
			}
		}
		dst[key] = srcVal
	}
	return dst
}

func merge(dst, src interface{}, depth int) interface{} {
	if depth > MaxDepth {
		panic("too deep!")
	}

	dv := reflect.ValueOf(dst)
	sv := reflect.ValueOf(src)

	if dv.Kind() == reflect.Map {
		srcMap := toMap(sv)
		destMap := toMap(dv)
		for key, srcVal := range srcMap {
			dv := reflect.ValueOf(destMap[key])
			if dv.Kind() != reflect.Map || dv.Kind() != reflect.Slice {
				destMap[key] = srcVal
			} else {
				srcVal = merge(destMap[key], srcVal, depth+1)
			}

			destMap[key] = srcVal
		}

	}

	if dv.Kind() == reflect.Slice {
		srcSlice := toSlice(sv)
		destSlice := toSlice(dv)
		for index, srcVal := range srcSlice {
			dv := reflect.ValueOf(destSlice[index])
			if dv.Kind() != reflect.Map || dv.Kind() != reflect.Slice {
				destSlice[index] = srcVal
			} else {
				return merge(destSlice[index], srcVal, depth+1)
			}
		}
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
		s[i] = value.Index(i)
	}
	return s
}
