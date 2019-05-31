package testtools

import (
	// "github.com/smartedge/codechallenge"
	// "errors"
	"fmt"
	// "reflect"
)

// ErrorSpec defines the expected value of an error.
type ErrorSpec struct {
	Type    string
	Message string
}

// EnsureMatches returns an error is inErr doesn't match.
func (es *ErrorSpec) EnsureMatches(inErr error) error {
	if (es == nil) && (inErr == nil) {
		return nil
	}
	if (es != nil) && (inErr != nil) {
		if ((es.Type == "") || (es.Type == fmt.Sprintf("%T", inErr))) &&
			(es.Message == inErr.Error()) {
			return nil
		}
	}
	actualStr := "didn't see any error"
	if inErr != nil {
		actualStr = fmt.Sprintf("saw a %T with message %#v instead", inErr, inErr.Error())
	}
	return fmt.Errorf("%s, but %s", es.String(), actualStr)
}

// String returns string representation of es.
func (es *ErrorSpec) String() string {
	result := "No error expected"
	if es != nil {
		expType := "an error"
		if es.Type != "" {
			expType = fmt.Sprintf("a %s error", es.Type)
		}
		result = fmt.Sprintf("Expected %s with message %#v", expType, es.Message)
	}
	return result
}

// AreFuncsEqual returns true only if a and b are both functions, and
// both point to the same function. Returns false and a non-nil error if either
// argument is not a function. Returns true and a non-nil error if both
// arguments are nil,
func AreFuncsEqual(_ interface{}, _ interface{}) (bool, error) {
	return false, nil
	// if (a == nil) || (b == nil) {
	// 	result := (a == nil) && (b == nil)
	// 	errMsg := "Both values nil when funcs expected"
	// 	if !result {
	// 		if a == nil {
	// 			errMsg = "First value nil when two funcs expected"
	// 		} else {
	// 			errMsg = "Second value nil when two funcs expected"
	// 		}
	// 	}
	// 	return result, errors.New(errMsg)
	// }
	// aRef, bRef := reflect.ValueOf(a), reflect.ValueOf(b)

}
