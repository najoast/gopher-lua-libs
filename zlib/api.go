package zlib

import (
	"bytes"
	"compress/zlib"

	lua "github.com/yuin/gopher-lua"
)

func Compress(L *lua.LState) int {
	str := L.CheckString(1)
	level := zlib.DefaultCompression
	if L.GetTop() > 1 {
		level = L.CheckInt(2)
	}

	var b bytes.Buffer
	w, err := zlib.NewWriterLevel(&b, level)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	if _, err = w.Write([]byte(str)); err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}
	w.Close()
	L.Push(lua.LString(b.String()))
	return 1
}

func Decompress(L *lua.LState) int {
	str := L.CheckString(1)
	r, err := zlib.NewReader(bytes.NewBufferString(str))
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	var b bytes.Buffer
	if _, err = b.ReadFrom(r); err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}
	r.Close()
	L.Push(lua.LString(b.String()))
	return 1
}
