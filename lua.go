package luatemplate

// index 是一个Lua函数，用于获取LState中key对应的值
// L: Lua状态机
// key: 要获取值的key
// 返回值：返回对应key的值，如果key不存在则返回lua.LNil
