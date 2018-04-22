package queue

import (
	"reflect"
)

type unmarshalFunc func(data []byte, dest interface{}) error

// Model constructs middleware to unmarshal message bytes into Go object
func Model(typeModel reflect.Type, unmarshal unmarshalFunc) Middleware {
	return func(next Handler) Handler {
		return func(val interface{}) error {
			if data, ok := val.([]byte); ok {
				val = reflect.New(typeModel).Interface()
				if err := unmarshal(data, val); err != nil {
					panic(err)
				}
			}
			return next(val)
		}
	}
}

var errorInterface = reflect.TypeOf((*error)(nil)).Elem()

// H constructs Handler from a single-argument func and unmarshaler
func H(h interface{}, unmarshal unmarshalFunc) Handler {
	fv := reflect.ValueOf(h)
	t := fv.Type()
	switch {
	case t.Kind() != reflect.Func:
		panic("handler must be func")
	case t.NumIn() != 1:
		panic("handler must have one argument")
	case t.NumOut() != 1:
		panic("handler must have one return value")
	case !t.Out(0).Implements(errorInterface):
		panic("handler must return error")
	}
	return Model(t.In(0), unmarshal)(func(val interface{}) error {
		res := fv.Call([]reflect.Value{reflect.ValueOf(val).Elem()})
		if res[0].IsNil() {
			return nil
		}
		return res[0].Interface().(error)
	})
}