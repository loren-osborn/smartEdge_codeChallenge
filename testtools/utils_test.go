package testtools_test

import (
	"errors"
	"fmt"
	"github.com/smartedge/codechallenge/testtools"
	// "reflect"
	"testing"
)

// TestErrorSpec tests *ErrorSpec type
func TestErrorSpec(t *testing.T) {
	errorValSpecs := []struct {
		errVal error
		spec   *testtools.ErrorSpec
	}{
		{
			errVal: nil,
			spec:   nil,
		},
		{
			errVal: errors.New("This is a generic error"),
			spec: &testtools.ErrorSpec{
				Type:    "*errors.errorString",
				Message: "This is a generic error",
			},
		},
		{
			errVal: errors.New("Bar"),
			spec: &testtools.ErrorSpec{
				Type:    "*errors.errorString",
				Message: "Bar",
			},
		},
		{
			errVal: &fooError{},
			spec: &testtools.ErrorSpec{
				Type:    "*testtools_test.fooError",
				Message: "Bar",
			},
		},
	}
	getSpecString := func(spec *testtools.ErrorSpec) string {
		result := "No error expected"
		if spec != nil {
			expType := "an error"
			if spec.Type != "" {
				expType = fmt.Sprintf("a %s error", spec.Type)
			}
			result = fmt.Sprintf("Expected %s with message %#v", expType, spec.Message)
		}
		return result
	}
	generateStringSubtest := func(spec *testtools.ErrorSpec) (string, func(*testing.T)) {
		expectedStr := getSpecString(spec)
		subtestName := fmt.Sprintf("Testing String() method of %s", expectedStr)
		subtestFunc := func(tt *testing.T) {
			result := spec.String()
			if result != expectedStr {
				tt.Errorf("*ErrorSpec\n\t%#v reported itself as\n\t%#v", expectedStr, result)
			}
		}
		return subtestName, subtestFunc
	}
	generateEnsureMatchesSubtest := func(shouldMatch bool, compareSpec *testtools.ErrorSpec, actualErr error) (string, func(*testing.T)) {
		whatWeAreComparing := "two nils"
		if (compareSpec != nil) || (actualErr != nil) {
			if (compareSpec != nil) != (actualErr != nil) {
				// one nil, but which one?
				if compareSpec == nil {
					whatWeAreComparing = "a nil ErrorSpec to a non-nil error"
				} else if compareSpec.Type == "" {
					whatWeAreComparing = "an untyped ErrorSpec to a nil error"
				} else {
					whatWeAreComparing = "a typed ErrorSpec to a nil error"
				}
			} else {
				if compareSpec.Type == "" {
					whatWeAreComparing = fmt.Sprintf(
						"an untyped ErrorSpec for %#v to a %T: %#v error",
						compareSpec.Message,
						actualErr,
						actualErr.Error(),
					)
				} else {
					whatWeAreComparing = fmt.Sprintf(
						"an typed ErrorSpec for %s: %#v to a %T: %#v error",
						compareSpec.Type,
						compareSpec.Message,
						actualErr,
						actualErr.Error(),
					)
				}
			}
		}
		subtestName := fmt.Sprintf("Testing comparison of %s", whatWeAreComparing)
		subtestFunc := func(tt *testing.T) {
			matchErr := compareSpec.EnsureMatches(actualErr)
			if (matchErr == nil) != shouldMatch {
				whatHappened := "reported match"
				result := "no mismatch"
				if matchErr != nil {
					whatHappened = "reported mismatch"
					result = fmt.Sprintf("%T: %#v", matchErr, matchErr.Error())
				}
				tt.Errorf(
					"Incorrectly %s when comparing %#v to %T: %#v. As a result it returned %s",
					whatHappened,
					compareSpec,
					actualErr,
					actualErr.Error(),
					result,
				)
			} else if !shouldMatch {
				expectedErrVal := "didn't see any error"
				if actualErr != nil {
					expectedErrVal = fmt.Sprintf("saw a %T with message %#v instead", actualErr, actualErr.Error())
				}
				expectedSpec := getSpecString(compareSpec)
				expectedMessage := fmt.Sprintf("%s, but %s", expectedSpec, expectedErrVal)
				if matchErr.Error() != expectedMessage {
					tt.Errorf(
						"Incorrectly reporting mismatch with\n\t%#v when\n\t%#v was expected",
						matchErr.Error(),
						expectedMessage,
					)
				}
			}
		}
		return subtestName, subtestFunc
	}
	for specIdx, specVal := range errorValSpecs {
		t.Run(generateStringSubtest(specVal.spec))
		for compErrIdx, compErrVal := range errorValSpecs {
			// Ensure that ErrorSpec's that don't assert a type match correctly:
			if (specVal.errVal != nil) &&
				((specIdx == compErrIdx) ||
					(compErrVal.errVal == nil) ||
					(specVal.spec.Message != compErrVal.spec.Message)) {
				typelessSpec := &testtools.ErrorSpec{
					Type:    "",
					Message: specVal.spec.Message,
				}
				t.Run(generateEnsureMatchesSubtest(specIdx == compErrIdx, typelessSpec, compErrVal.errVal))
			}
			// Ensure that ErrorSpec's that do assert a type match correctly:
			t.Run(generateEnsureMatchesSubtest(specIdx == compErrIdx, specVal.spec, compErrVal.errVal))
		}
	}
}

type fooError struct{}

func (fe *fooError) Error() string {
	return "Bar"
}

// TestAreFuncsEqual tests AreFuncsEqual()
func TestAreFuncsEqual(t *testing.T) {
	eq, err := testtools.AreFuncsEqual(TestAreFuncsEqual, TestErrorSpec)
	if eq {
		t.Error("AreFuncsEqual() reports TestAreFuncsEqual() and TestErrorSpec() are the same function")
	}
	if err != nil {
		t.Errorf("AreFuncsEqual() should only return an error if passed a non-func: TestAreFuncsEqual() and TestErrorSpec() were passed, and we got error: %T: %#v", err, err.Error())
	}
	// total := Sum(5, 5)
	// if total != 10 {
	// 	t.Errorf("Sum was incorrect, got: %d, want: %d.", total, 10)
	// }
}
