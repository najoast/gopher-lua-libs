package redis

import (
	lua "github.com/yuin/gopher-lua"
)

// Preload adds redis to the given Lua state's package.preload table. After it
// has been preloaded, it can be loaded using require:
//
//	local redis = require("redis")
func Preload(L *lua.LState) {
	L.PreloadModule("redis", Loader)
}

// Loader is the module loader function.
func Loader(L *lua.LState) int {
	t := L.NewTable()
	L.SetFuncs(t, api)
	L.Push(t)
	return 1
}

var api = map[string]lua.LGFunction{
	"open":  Open,
	"call":  Call,
	"close": Close,
}
