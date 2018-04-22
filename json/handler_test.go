package json_test

import (
	"fmt"
	"testing"

	. "github.com/go-mixins/queue/json"
)

type testArg struct {
	X float64
	S string
}

type customTestError int

func (c customTestError) Error() string {
	return fmt.Sprintf("%d", c)
}

func testStructPtrH(arg *testArg) error {
	switch {
	case arg.X != 12345.6:
		return customTestError(1)
	case arg.S != "hello":
		return customTestError(2)
	}
	return nil
}

func BenchmarkF(b *testing.B) {
	h := F(testStructPtrH)
	for i := 0; i < b.N; i++ {
		if err := h([]byte(`{"X": 12345.6, "S": "hello"}`)); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkType(b *testing.B) {
	h := Type(testArg{})(func(val interface{}) error { return testStructPtrH(val.(*testArg)) })
	for i := 0; i < b.N; i++ {
		if err := h([]byte(`{"X": 12345.6, "S": "hello"}`)); err != nil {
			b.Fatal(err)
		}
	}
}
