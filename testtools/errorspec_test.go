package testtools_test

import (
	"errors"
	"fmt"
	"github.com/smartedge/codechallenge/testtools"
	"testing"
)

// TestErrorSpecString tests *ErrorSpec's String() method
func TestErrorSpecString(t *testing.T) {
	for _, specVal := range []struct {
		spec *testtools.ErrorSpec
		desc string
	}{
		// No error
		{
			spec: nil,
			desc: "No error expected",
		},
		// Specific error type 1
		{
			spec: &testtools.ErrorSpec{
				Type:    "*myPkg.myErrType",
				Message: "foo",
			},
			desc: "Expected a *myPkg.myErrType error with message \"foo\"",
		},
		// Specific error type 2
		{
			spec: &testtools.ErrorSpec{
				Type:    "[]yourPkg.yourErrorSliceType",
				Message: "bar",
			},
			desc: "Expected a []yourPkg.yourErrorSliceType error with message \"bar\"",
		},
		// untyped error
		{
			spec: &testtools.ErrorSpec{
				Type:    "",
				Message: "fizzbuzz",
			},
			desc: "Expected an error with message \"fizzbuzz\"",
		},
	} {
		t.Run(
			fmt.Sprintf("ErrorSpec for %#v", specVal.desc),
			func(tt *testing.T) {
				result := specVal.spec.String()
				if result != specVal.desc {
					tt.Errorf("*ErrorSpec\n\t%#v reported itself as\n\t%#v", specVal.desc, result)
				}
			})
	}
}

// TestGetErrorSpecFrom tests NewErrorSpecFrom() function to generate *ErrorSpec
// that matches argument.
func TestNewErrorSpecFrom(t *testing.T) {
	for _, tc := range []struct {
		input error
	}{
		// No error
		{input: nil},
		// An errorString error
		{input: errors.New("foo")},
		// Another errorString error
		{input: errors.New("bar")},
		// A fooError
		{input: &fooError{}},
	} {
		t.Run(
			fmt.Sprintf("testtools.NewErrorSpecFrom(%#v)", tc.input),
			func(tt *testing.T) {
				var result *testtools.ErrorSpec = testtools.NewErrorSpecFrom(tc.input)
				if matchErr := result.EnsureMatches(tc.input); matchErr != nil {
					tt.Errorf(
						"When running:\n"+
							"\tresult := *testtools.NewErrorSpecFrom(%#v)\n"+
							"\terr := result.EnsureMatches(%#v)\n"+
							"Should make err == nil. Instead got:\n"+
							"\tresult: %#v\n"+
							"\terr: %#v\n"+
							"\terr.Error(): %#v",
						tc.input,
						tc.input,
						result,
						matchErr,
						matchErr.Error())
				}
				if (result != nil) && (result.Type == "") {
					tt.Errorf(
						"When running:\n"+
							"\t*testtools.NewErrorSpecFrom(%#v)\n"+
							"Should not have an empty Type, but it does:\n"+
							"\tresult: %#v",
						tc.input,
						result)
				}
			})
	}
}

// TestErrorSpecEnsureMatches tests *ErrorSpec's EnsureMatches() method
func TestErrorSpecEnsureMatches(t *testing.T) {
	// generateEnsureMatchesSubtest returns a subtest 2-tuple (name, subtest function)
	// verifying the correct result from *ErrorSpec.EnsureMatches()
	generateEnsureMatchesSubtest := func(shouldMatch bool, compareSpec *testtools.ErrorSpec, actualErr error) (string, func(*testing.T)) {
		// Here we create the subtest name:
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
		// create the subtest function:
		subtestFunc := func(tt *testing.T) {
			matchErr := compareSpec.EnsureMatches(actualErr)
			if (matchErr == nil) != shouldMatch {
				// This is a logic error, producing a false positive or negative
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
				// This is only validating the message is what we expect
				expectedErrVal := "didn't see any error"
				if actualErr != nil {
					expectedErrVal = fmt.Sprintf("saw a %T with message %#v instead", actualErr, actualErr.Error())
				}
				expectedSpec := compareSpec.String()
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
	// 4 usefull error types: nested, so 16 test cases
	errorValSpecs := []struct {
		errVal error
		spec   *testtools.ErrorSpec
	}{
		// No error
		{
			errVal: nil,
			spec:   nil,
		},
		// Generic error
		{
			errVal: errors.New("This is a generic error"),
			spec: &testtools.ErrorSpec{
				Type:    "*errors.errorString",
				Message: "This is a generic error",
			},
		},
		// A custom error type, with a constant output
		{
			errVal: &fooError{},
			spec: &testtools.ErrorSpec{
				Type:    "*testtools_test.fooError",
				Message: "Bar",
			},
		},
		// A different error type, with the same output
		{
			errVal: errors.New("Bar"),
			spec: &testtools.ErrorSpec{
				Type:    "*errors.errorString",
				Message: "Bar",
			},
		},
	}
	for specIdx, specVal := range errorValSpecs {
		for compErrIdx, compErrVal := range errorValSpecs {
			// Ensure that ErrorSpec's that don't assert a type match correctly
			// (ony perform subtest when indexs match *OR* Messages DO NOT match)
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
			// Ensure that ErrorSpec's that do assert a type, match correctly:
			t.Run(generateEnsureMatchesSubtest(specIdx == compErrIdx, specVal.spec, compErrVal.errVal))
		}
	}
}

type fooError struct{}

func (fe *fooError) Error() string {
	return "Bar"
}
