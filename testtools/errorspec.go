package testtools

import (
	"fmt"
)

// ErrorSpec defines the expected value of an error.
type ErrorSpec struct {
	Type    string
	Message string
}

// NewErrorSpecFrom returns a *ErrorSpec that matches inErr.
func NewErrorSpecFrom(inErr error) *ErrorSpec {
	if inErr == nil {
		return nil
	}
	return &ErrorSpec{
		Type:    fmt.Sprintf("%T", inErr),
		Message: inErr.Error(),
	}
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
