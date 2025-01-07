package luatemplate

import "github.com/vela-public/onekit/lua"

func V2I(val lua.LValue) (interface{}, bool) {
	switch val.Type() {
	case lua.LTInt:
		return int(val.(lua.LInt)), true
	case lua.LTNumber:
		return float64(val.(lua.LNumber)), true
	case lua.LTBool:
		return bool(val.(lua.LBool)), true
	case lua.LTString:
		return val.String(), true
	default:
		return nil, false
	}
}

func Tab2ParamTab(L *lua.LState, tab *lua.LTable) []*Param {
	arr := tab.Array()
	if len(arr) == 0 {
		L.RaiseError("must param table got nil")
		return nil
	}

	pArr := make([]*Param, len(arr))
	for i, item := range arr {
		v, ok := item.(*Param)
		if ok {
			pArr[i] = v
			continue
		}
		L.RaiseError("param #%d must param.object got %s ", i, item.Type().String())
	}

	return pArr
}

func SetKV(kv []ParamKV, key string, val interface{}) []ParamKV {
	n := len(kv)
	if n == 0 {
		kv = append(kv, ParamKV{key, val})
		return kv
	}

	for i := 0; i < n; i++ {
		if kv[i].key == key {
			kv[i].val = val
			return kv
		}
	}

	kv = append(kv, ParamKV{key, val})
	return kv
}

func NewDefault[T any](n int) []T {
	arr := make([]T, n)
	return arr
}

func Pickup[T any](data []ParamKV, key string) (val T, ok bool) {
	n := len(data)
	if n == 0 {
		ok = false
		return
	}

	for i := 0; i < n; i++ {
		elem := data[i]
		if elem.key == key {
			val, ok = elem.val.(T)
			return
		}
	}

	ok = false
	return
}
