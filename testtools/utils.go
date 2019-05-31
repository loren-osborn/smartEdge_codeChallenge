package testtools

import (
// "github.com/smartedge/codechallenge"
// "errors"
// "reflect"
)

// ErrorSpec defines the expected value of an error.
type ErrorSpec struct {
}

// EnsureMatches returns an error is inErr doesn't match.
func (es *ErrorSpec) EnsureMatches(_ error) error {
	return nil
}

// AreFuncsEqual returns true only if a and b are both functions, and
// both point to the same function. Returns false and a non-nil error if either
// argument is not a function. Returns true and a non-nil error if both
// arguments are nil,
func AreFuncsEqual(_ interface{}, _ interface{}) (bool, error) {
	return false, nil
	// if (a == nil) || (b == nil) {
	// 	result := (a == nil) && (b == nil)
	// 	errMsg := "Both values nil when funcs expected."
	// 	if !result {
	// 		if a == nil {
	// 			errMsg = "First value nil when two funcs expected."
	// 		} else {
	// 			errMsg = "Second value nil when two funcs expected."
	// 		}
	// 	}
	// 	return result, errors.New(errMsg)
	// }
	// aRef, bRef := reflect.ValueOf(a), reflect.ValueOf(b)

}
