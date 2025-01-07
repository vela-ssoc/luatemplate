package luatemplate

import "github.com/vela-public/onekit/lua"

// indexL 是一个Lua函数，用于获取LState中key对应的值
// L: Lua状态机
// key: 要获取值的key
// 返回值：返回对应key的值，如果key不存在则返回lua.LNil
func indexL(L *lua.LState, key string) lua.LValue {
	switch key {
	case "param":
		return lua.NewExport("lua.param.export", lua.WithIndex(ParamIndexL))
	case "template":
		return lua.NewFunction(NewTemplateL)
	}

	return lua.LNil
}

/*
	local t = template{
		version = "0.0.1",
		auth    = "vela",
	}

	local p = template.param

	t.param = {
		setup = p.lua(),
	}

*/

func GenLuaCodeL() lua.Export {
	return lua.NewExport("lua.gee.export", lua.WithIndex(indexL))
}
