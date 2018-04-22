package json

import (
	"encoding/json"
	"reflect"

	queue "github.com/go-mixins/queue"
)

// Type constructs unmarshaling middleware for JSON wire encoding
func Type(model interface{}) queue.Middleware {
	typeModel := reflect.TypeOf(model)
	if typeModel.Kind() == reflect.Ptr {
		typeModel = typeModel.Elem()
	}
	return queue.TypeH(typeModel, json.Unmarshal)
}

// F constructs Handler for JSON wire encoding from provided func
func F(f interface{}) queue.Handler {
	return queue.FuncH(f, json.Unmarshal)
}
