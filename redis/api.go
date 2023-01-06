package redis

import (
	"reflect"
	"strings"

	"github.com/go-redis/redis/v8"
	lua "github.com/yuin/gopher-lua"
)

/*
	redis.connect {
		addrs = "host1:port1,host2:port2",
		db = 0,
		username = "username",
		password = "password",
		readOnly = false,
	}

returns (redis.UniversalClient, err)
*/
func Open(L *lua.LState) int {
	tbl := L.CheckTable(1)
	options := redis.UniversalOptions{
		Addrs:        strings.Split(tbl.RawGetString("addrs").String(), ","),
		DB:           int(lua.LVAsNumber(tbl.RawGetString("db"))),
		Username:     tbl.RawGetString("username").String(),
		Password:     tbl.RawGetString("password").String(),
		ReadOnly:     lua.LVAsBool(tbl.RawGetString("readOnly")),
		MaxRetries:   3,
		PoolSize:     10000,
		MinIdleConns: 0,
	}
	cli := redis.NewUniversalClient(&options)
	err := cli.Ping(L.Context()).Err()
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	ud := L.NewUserData()
	ud.Value = cli
	L.Push(ud)
	return 1
}

func Call(L *lua.LState) int {
	ud := L.CheckUserData(1)
	cli := ud.Value.(*redis.Client)
	var args []interface{}
	for i := 2; i <= L.GetTop(); i++ {
		arg := L.Get(i)
		switch arg.Type() {
		case lua.LTNumber:
			args = append(args, lua.LVAsNumber(arg))
		case lua.LTString:
			args = append(args, arg.String())
		case lua.LTBool:
			args = append(args, lua.LVAsBool(arg))
		case lua.LTNil:
			args = append(args, nil)
		case lua.LTTable:
			args = append(args, arg.(*lua.LTable))
		default:
			L.Push(lua.LNil)
			L.Push(lua.LString("unsupported type: " + arg.Type().String()))
			return 2
		}
	}
	cmd := cli.Do(L.Context(), args...)
	if cmd.Err() != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(cmd.Err().Error()))
		return 2
	}

	val := cmd.Val()
	switch val := val.(type) {
	case string:
		L.Push(lua.LString(val))
	case int64:
		L.Push(lua.LNumber(val))
	case float64:
		L.Push(lua.LNumber(val))
	case bool:
		L.Push(lua.LBool(val))
	case []byte:
		L.Push(lua.LString(string(val)))
	case []interface{}:
		L.Push(createAnySlice(L, val))
	case *redis.StringCmd:
		L.Push(lua.LString(val.Val()))
	case *redis.IntCmd:
		L.Push(lua.LNumber(val.Val()))
	case *redis.FloatCmd:
		L.Push(lua.LNumber(val.Val()))
	case *redis.BoolCmd:
		L.Push(lua.LBool(val.Val()))
	case *redis.StatusCmd:
		L.Push(lua.LString(val.Val()))
	case *redis.SliceCmd:
		typedVal := val.Val()
		tbl := L.CreateTable(len(typedVal), 0)
		for i, v := range typedVal {
			tbl.RawSetInt(i+1, lua.LString(v.(string)))
		}
		L.Push(tbl)
	case *redis.StringSliceCmd:
		typedVal := val.Val()
		tbl := L.CreateTable(len(typedVal), 0)
		for i, v := range typedVal {
			tbl.RawSetInt(i+1, lua.LString(v))
		}
		L.Push(tbl)
	case *redis.IntSliceCmd:
		typedVal := val.Val()
		tbl := L.CreateTable(len(typedVal), 0)
		for i, v := range typedVal {
			tbl.RawSetInt(i+1, lua.LNumber(v))
		}
		L.Push(tbl)
	case *redis.StringStringMapCmd:
		typedVal := val.Val()
		tbl := L.CreateTable(0, len(typedVal))
		for k, v := range typedVal {
			tbl.RawSetString(k, lua.LString(v))
		}
		L.Push(tbl)
	case *redis.StringIntMapCmd:
		typedVal := val.Val()
		tbl := L.CreateTable(0, len(typedVal))
		for k, v := range typedVal {
			tbl.RawSetString(k, lua.LNumber(v))
		}
		L.Push(tbl)
	case *redis.StringStructMapCmd:
		typedVal := val.Val()
		tbl := L.CreateTable(len(typedVal), 0)
		i := 1
		for key := range typedVal {
			tbl.RawSetInt(i, lua.LString(key))
			i++
		}
		L.Push(tbl)
	default:
		L.Push(lua.LNil)
		L.Push(lua.LString("unsupported type: " + reflect.TypeOf(val).String()))
		return 2
	}
	return 1
}

func Close(L *lua.LState) int {
	ud := L.CheckUserData(1)
	cli := ud.Value.(*redis.Client)
	cli.Close()
	return 0
}

func createAnySlice(L *lua.LState, val []interface{}) *lua.LTable {
	tbl := L.CreateTable(len(val), 0)
	for i, v := range val {
		switch v := v.(type) {
		case string:
			tbl.RawSetInt(i+1, lua.LString(v))
		case int64:
			tbl.RawSetInt(i+1, lua.LNumber(v))
		case float64:
			tbl.RawSetInt(i+1, lua.LNumber(v))
		case bool:
			tbl.RawSetInt(i+1, lua.LBool(v))
		case []byte:
			tbl.RawSetInt(i+1, lua.LString(string(v)))
		case []interface{}:
			tbl.RawSetInt(i+1, createAnySlice(L, v))
		case *redis.StringCmd:
			tbl.RawSetInt(i+1, lua.LString(v.Val()))
		case *redis.IntCmd:
			tbl.RawSetInt(i+1, lua.LNumber(v.Val()))
		case *redis.FloatCmd:
			tbl.RawSetInt(i+1, lua.LNumber(v.Val()))
		case *redis.BoolCmd:
			tbl.RawSetInt(i+1, lua.LBool(v.Val()))
		case *redis.StatusCmd:
			tbl.RawSetInt(i+1, lua.LString(v.Val()))
		case *redis.SliceCmd:
			typedVal := v.Val()
			tbl2 := L.CreateTable(len(typedVal), 0)
			for i, v := range typedVal {
				tbl2.RawSetInt(i+1, lua.LString(v.(string)))
			}
			tbl.RawSetInt(i+1, tbl2)
		case *redis.StringSliceCmd:
			typedVal := v.Val()
			tbl2 := L.CreateTable(len(typedVal), 0)
			for i, v := range typedVal {
				tbl2.RawSetInt(i+1, lua.LString(v))
			}
			tbl.RawSetInt(i+1, tbl2)
		case *redis.IntSliceCmd:
			typedVal := v.Val()
			tbl2 := L.CreateTable(len(typedVal), 0)
			for i, v := range typedVal {
				tbl2.RawSetInt(i+1, lua.LNumber(v))
			}
			tbl.RawSetInt(i+1, tbl2)
		case *redis.StringStringMapCmd:
			typedVal := v.Val()
			tbl2 := L.CreateTable(0, len(typedVal))
			for k, v := range typedVal {
				tbl2.RawSetString(k, lua.LString(v))
			}
			tbl.RawSetInt(i+1, tbl2)
		case *redis.StringIntMapCmd:
			typedVal := v.Val()
			tbl2 := L.CreateTable(0, len(typedVal))
			for k, v := range typedVal {
				tbl2.RawSetString(k, lua.LNumber(v))
			}
			tbl.RawSetInt(i+1, tbl2)
		case *redis.StringStructMapCmd:
			typedVal := v.Val()
			tbl2 := L.CreateTable(len(typedVal), 0)
			i := 1
			for key := range typedVal {
				tbl2.RawSetInt(i, lua.LString(key))
				i++
			}
			tbl.RawSetInt(i+1, tbl2)
		default:
			tbl.RawSetInt(i+1, lua.LNil)
		}
	}
	return tbl
}
