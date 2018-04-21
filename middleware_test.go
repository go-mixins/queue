package queue_test

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/go-mixins/queue"
)

type testHCase struct {
	h     interface{}
	val   string
	err   error
	panic bool
}

var errTestH = errors.New("test error")
var errCaughtPanic = errors.New("panic caught")

type testArg struct {
	X float64
	S string
}

type customTestError int

func (c customTestError) Error() string {
	return fmt.Sprintf("%d", c)
}

func testFloatH(x float64) error {
	if x != 12345.6 {
		return errTestH
	}
	return nil
}

func testStringH(s string) error {
	if s != "hello" {
		return customTestError(3)
	}
	return nil
}

func testSliceH(t []int) error {
	if len(t) != 3 {
		return customTestError(1)
	}
	return nil
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

func testStructH(arg testArg) error {
	switch {
	case arg.X != 12345.6:
		return customTestError(1)
	case arg.S != "hello":
		return customTestError(2)
	}
	return nil
}

var testHCases = []testHCase{
	{h: testFloatH, val: "12345.6", err: nil},
	{h: testFloatH, val: "11", err: errTestH},
	{h: testStructH, val: `{"X": 12345.6, "S": "hello"}`, err: nil},
	{h: testStructH, val: `{"S": "hello"}`, err: customTestError(1)},
	{h: testStructH, val: `{"X": 12345.6}`, err: customTestError(2)},
	{h: testStructPtrH, val: `{"X": 12345.6, "S": "hello"}`, err: nil},
	{h: testStructPtrH, val: `{"S": "hello"}`, err: customTestError(1)},
	{h: testStructPtrH, val: `{"X": 12345.6}`, err: customTestError(2)},
	{h: testStringH, val: `"hello"`, err: nil},
	{h: testStringH, val: `""`, err: customTestError(3)},
	{h: testStringH, val: ``, err: errCaughtPanic, panic: true},
	{h: testSliceH, val: `[1,2,3]`, err: nil},
	{h: testSliceH, val: `[1,2]`, err: customTestError(1)},
}

func TestH(t *testing.T) {
	for _, tc := range testHCases {
		handler := queue.Recover(func(p interface{}) error {
			if !tc.panic {
				t.Errorf("panic: %v", p)
			}
			return errCaughtPanic
		})(queue.H(tc.h))
		err := handler([]byte(tc.val))
		if !reflect.DeepEqual(err, tc.err) {
			t.Errorf("invalid error: %#v", err)
		}
	}
}
