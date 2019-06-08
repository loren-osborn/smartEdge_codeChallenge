package testtools

import (
	"errors"
	"fmt"
	"reflect"
)

// AreFuncsEqual returns true only if a and b are both functions, and
// both point to the same function. Returns false and a non-nil error if either
// argument is not a function. Returns true and a non-nil error if both
// arguments are nil,
func AreFuncsEqual(a interface{}, b interface{}) (bool, error) {
	checkTwoVals := func(matcher func(int) bool, matchDesc string, expectedDesc string) error {
		matches := 0
		for i := 0; i < 2; i++ {
			if matcher(i) {
				matches++
			}
		}
		if matches > 0 {
			errMsg := fmt.Sprintf("Both values %s when %s expected", matchDesc, expectedDesc)
			if matches == 1 {
				which := "Second"
				if matcher(0) {
					which = "First"
				}
				errMsg = fmt.Sprintf("%s value %s when two %s expected", which, matchDesc, expectedDesc)
			}
			return errors.New(errMsg)
		}
		return nil
	}

	valueInfos := [2]struct {
		val             interface{}
		isNil           bool
		valueReflection reflect.Value
		typeReflection  reflect.Type
	}{}
	valueInfos[0].val = a
	valueInfos[1].val = b
	for i := 0; i < 2; i++ {
		valueInfos[i].isNil = valueInfos[i].val == nil
		if !valueInfos[i].isNil {
			valueInfos[i].valueReflection = reflect.ValueOf(valueInfos[i].val)
			valueInfos[i].isNil = valueInfos[i].valueReflection.IsNil()
			valueInfos[i].typeReflection = valueInfos[i].valueReflection.Type()
		}
	}
	if err := checkTwoVals(func(i int) bool { return valueInfos[i].isNil }, "nil", "funcs"); err != nil {
		result := valueInfos[0].isNil && valueInfos[1].isNil && (valueInfos[0].typeReflection == valueInfos[1].typeReflection)
		return result, err
	}
	if err := checkTwoVals(func(i int) bool { return valueInfos[i].typeReflection.Kind() != reflect.Func }, "not a func", "funcs"); err != nil {
		return false, err
	}
	return (valueInfos[0].valueReflection.Pointer() == valueInfos[1].valueReflection.Pointer()), nil
}

// AreStringSlicesEqual determines if two string slices are equal. Equality
// distinguishes nil-ness, but not capacity
func AreStringSlicesEqual(a []string, b []string) bool {
	if (a == nil) || (b == nil) {
		return (a == nil) && (b == nil)
	}
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if b[i] != v {
			return false
		}
	}
	return true
}

// CloneStringSlice create a non-shared copy of inSlice
func CloneStringSlice(inSlice []string) []string {
	if inSlice == nil {
		return nil
	}
	result := make([]string, len(inSlice), cap(inSlice))
	for i, v := range inSlice {
		result[i] = v
	}
	return result
}

// WrapFuncCallWithCounter wraps the provided function, adding a returned
// pointer to a call counter.
func WrapFuncCallWithCounter(f func()) (func(), *int) {
	counter := 0
	wrapped := func() {
		f()
		counter++
	}
	return wrapped, &counter
}

// LoopReader an io.Reader that endlessly reads the same series of bytes
type LoopReader struct {
	Data string // If empty, will be automatically promoted to "\x00"
	Pos  int
}

// Read provides an endless stream of predictable (looping) bytes.
func (lr *LoopReader) Read(p []byte) (n int, err error) {
	// initialization
	if lr.Data == "" {
		lr.Data = "\x00"
	}
	if (lr.Pos < 0) || (lr.Pos >= len(lr.Data)) {
		lr.Pos = 0
	}
	for i := range p {
		p[i] = []byte(lr.Data)[(lr.Pos+i)%len(lr.Data)]
	}
	lr.Pos = (lr.Pos + len(p)) % len(lr.Data)
	return len(p), nil
}
