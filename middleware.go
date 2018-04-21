package queue

import (
	"encoding/json"
	"reflect"
)

// Model constructs middleware to unmarshal message bytes into Go object
func Model(typeModel reflect.Type, unmarshal func(data []byte, dest interface{}) error) Middleware {
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

// JSON constructs unmarshaling middleware for JSON wire encoding
func JSON(model interface{}) Middleware {
	typeModel := reflect.TypeOf(model)
	if typeModel.Kind() == reflect.Ptr {
		typeModel = typeModel.Elem()
	}
	return Model(typeModel, json.Unmarshal)
}

// Recover prevents handler from propagating panic by calling `catch` on it
func Recover(catch Handler) Middleware {
	return func(next Handler) Handler {
		return func(val interface{}) (err error) {
			defer func() {
				if p := recover(); p != nil {
					err = catch(p)
				}
			}()
			return next(val)
		}
	}
}

var errorInterface = reflect.TypeOf((*error)(nil)).Elem()

// H constructs Handler from a single-argument func
func H(h interface{}) Handler {
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
	return Model(t.In(0), json.Unmarshal)(func(val interface{}) error {
		res := fv.Call([]reflect.Value{reflect.ValueOf(val).Elem()})
		if res[0].IsNil() {
			return nil
		}
		return res[0].Interface().(error)
	})
}
