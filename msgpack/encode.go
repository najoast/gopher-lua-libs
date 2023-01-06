package msgpack

import (
	"errors"
	"reflect"

	"github.com/vmihailenco/msgpack/v5"
	lua "github.com/yuin/gopher-lua"
)

const msgpackTableIsObject = "__msgpackTableIsObject"

var (
	errNested      = errors.New("cannot encode recursively nested tables to msgpack")
	errSparseArray = errors.New("cannot encode sparse array")
	errInvalidKeys = errors.New("cannot encode mixed or invalid key types")
)

type invalidTypeError lua.LValueType

func (i invalidTypeError) Error() string {
	return `cannot encode ` + lua.LValueType(i).String() + ` to msgpack`
}

type msgpackValue struct {
	lua.LValue
	visited map[*lua.LTable]bool
}

var _ msgpack.CustomEncoder = (*msgpackValue)(nil)

func marshalEmptyTable(table *lua.LTable) []byte {
	if mt, ok := table.Metatable.(*lua.LTable); ok {
		if lua.LVAsBool(mt.RawGetString(msgpackTableIsObject)) {
			return []byte("{}")
		}
	}
	return []byte("[]")
}

func (v *msgpackValue) EncodeMsgpack(enc *msgpack.Encoder) error {
	var err error
	switch converted := v.LValue.(type) {
	case lua.LBool:
		err = enc.EncodeBool(bool(converted))
	case lua.LNumber:
		err = enc.EncodeFloat64(float64(converted))
	case *lua.LNilType:
		err = enc.EncodeNil()
	case lua.LString:
		err = enc.EncodeString(string(converted))
	case *lua.LTable:
		if v.visited[converted] {
			return errNested
		}
		v.visited[converted] = true

		key, value := converted.Next(lua.LNil)

		switch key.Type() {
		case lua.LTNil: // empty table
			err = enc.EncodeBytes(marshalEmptyTable(converted))
		case lua.LTNumber:
			arr := make([]msgpackValue, 0, converted.Len())
			expectedKey := lua.LNumber(1)
			for key != lua.LNil {
				if key.Type() != lua.LTNumber {
					return errInvalidKeys
				}
				if expectedKey != key {
					return errSparseArray
				}
				arr = append(arr, msgpackValue{value, v.visited})
				expectedKey++
				key, value = converted.Next(key)
			}
			err = enc.EncodeValue(reflect.ValueOf(arr))
		case lua.LTString:
			obj := make(map[string]interface{}, converted.Len())
			for key != lua.LNil {
				if key.Type() != lua.LTString {
					return errInvalidKeys
				}
				obj[key.String()] = &msgpackValue{value, v.visited}
				key, value = converted.Next(key)
			}
			err = enc.EncodeMapSorted(obj)
		default:
			err = errInvalidKeys
		}
	default:
		err = invalidTypeError(v.LValue.Type())
	}
	return err
}

// ValueEncode returns the msgpack encoding of value.
func ValueEncode(value lua.LValue) ([]byte, error) {
	return msgpack.Marshal(&msgpackValue{
		LValue:  value,
		visited: make(map[*lua.LTable]bool),
	})
}
