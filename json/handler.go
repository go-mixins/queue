package json

import (
	"encoding/json"
	"reflect"

	queue "github.com/go-mixins/queue"
)

// Model constructs unmarshaling middleware for JSON wire encoding
func Model(model interface{}) queue.Middleware {
	typeModel := reflect.TypeOf(model)
	if typeModel.Kind() == reflect.Ptr {
		typeModel = typeModel.Elem()
	}
	return queue.Model(typeModel, json.Unmarshal)
}

// H constructs Handler for JSON wire encoding
func H(h interface{}) queue.Handler {
	return queue.H(h, json.Unmarshal)
}
