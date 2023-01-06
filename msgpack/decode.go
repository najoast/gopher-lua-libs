package msgpack

import (
	"bytes"
	"reflect"
	"time"

	"github.com/vmihailenco/msgpack/v5"
	lua "github.com/yuin/gopher-lua"
)

// ValueDecode converts the msgpack encoded data to Lua values.

func ValueDecode(L *lua.LState, data []byte) (lua.LValue, error) {
	dec := msgpack.NewDecoder(bytes.NewReader(data))
	dec.SetMapDecoder(func(d *msgpack.Decoder) (interface{}, error) {
		return d.DecodeUntypedMap()
	})
	var value interface{}
	err := dec.Decode(&value)
	if err != nil {
		return nil, err

	}
	return decode(L, value), nil
}

func decode(L *lua.LState, value interface{}) lua.LValue {
	switch converted := value.(type) {
	case bool:
		return lua.LBool(converted)
	case int8:
		return lua.LNumber(converted)
	case int16:
		return lua.LNumber(converted)
	case int32:
		return lua.LNumber(converted)
	case int64:
		return lua.LNumber(converted)
	case uint8:
		return lua.LNumber(converted)
	case uint16:
		return lua.LNumber(converted)
	case uint32:
		return lua.LNumber(converted)
	case uint64:
		return lua.LNumber(converted)
	case int:
		return lua.LNumber(converted)
	case uint:
		return lua.LNumber(converted)
	case float32:
		return lua.LNumber(converted)
	case float64:
		return lua.LNumber(converted)
	case string:
		return lua.LString(converted)
	case []interface{}:
		arr := L.CreateTable(len(converted), 0)
		for _, item := range converted {
			arr.Append(decode(L, item))
		}
		return arr

	case map[string]interface{}:
		tbl := L.CreateTable(0, len(converted))
		L.SetMetatable(tbl, L.GetTypeMetatable(msgpackTableIsObject))
		for key, item := range converted {
			tbl.RawSetH(lua.LString(key), decode(L, item))
		}
		return tbl

	case map[interface{}]interface{}:
		tbl := L.CreateTable(0, len(converted))
		L.SetMetatable(tbl, L.GetTypeMetatable(msgpackTableIsObject))
		for key, item := range converted {
			tbl.RawSet(decode(L, key), decode(L, item))
		}
		return tbl

	case nil:
		return lua.LNil

	case time.Time:
		return lua.LString(converted.Format(time.RFC3339))
	default:
		panic(`cannot decode ` + reflect.TypeOf(value).String() + ` to Lua`)
	}
}
