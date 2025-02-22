package zlib

import (
	lua "github.com/yuin/gopher-lua"
)

// Preload adds crypto to the given Lua state's package.preload table. After it
// has been preloaded, it can be loaded using require:
//
//	local crypto = require("crypto")
func Preload(L *lua.LState) {
	L.PreloadModule("zlib", Loader)
}

// Loader is the module loader function.
func Loader(L *lua.LState) int {
	t := L.NewTable()
	L.SetFuncs(t, api)
	L.Push(t)
	return 1
}

var api = map[string]lua.LGFunction{
	"compress":   Compress,
	"decompress": Decompress,
}
