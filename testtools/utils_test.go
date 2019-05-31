package testtools_test

import (
	"errors"
	"fmt"
	"github.com/smartedge/codechallenge/testtools"
	"strings"
	// "reflect"
	"testing"
)

// TestErrorSpec tests *ErrorSpec type's properties and String() and
// EnsureMatches() methods
func TestErrorSpec(t *testing.T) {
	// getSpecString(), generateStringSubtest() and
	// generateEnsureMatchesSubtest() are functions internal to this test.

	// getSpecString returns a string representation of a *ErrorSpec.
	// This is practically identical to the code in *ErrorSpec.String() that
	// generateStringSubtest() tests, so that may be slightly dubious
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

	// generateStringSubtest returns a subtest 2-tuple (name, subtest function)
	// verifying the correct result from *ErrorSpec.String(). Verifies it
	// matches output from getSpecString()
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

	// generateStringSubtest returns a subtest 2-tuple (name, subtest function)
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
		t.Run(generateStringSubtest(specVal.spec))
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

// TestAreFuncsEqual tests AreFuncsEqual()
func TestAreFuncsEqual(t *testing.T) {
	// genErrSpecGetter returns a closure that injects substitute values for
	// error message substrings `{{funcA}}` and `{{funcB}}` in the given order
	// into the provided ErrorSpec
	genErrSpecGetter := func(spec *testtools.ErrorSpec) func(string, string, string, string) *testtools.ErrorSpec {
		return func(aLabel, aName, bLabel, bName string) *testtools.ErrorSpec {
			if spec == nil {
				return nil
			}
			r := strings.NewReplacer(
				"{{funcA}}", aName,
				"{{funcB}}", bName,
				fmt.Sprintf("{{func%sOrdinal}}", aLabel), "first",
				fmt.Sprintf("{{Func%sOrdinal}}", aLabel), "First",
				fmt.Sprintf("{{func%sOrdinal}}", bLabel), "second",
				fmt.Sprintf("{{Func%sOrdinal}}", bLabel), "Second")
			return &testtools.ErrorSpec{
				Type:    spec.Type,
				Message: r.Replace(spec.Message),
			}
		}
	}

	type testCaseSpec struct {
		funcA         interface{}
		funcAName     string
		funcB         interface{}
		funcBName     string
		ExpectedMatch bool
		ExpectedErr   *testtools.ErrorSpec
	}
	var nilFunc func()
	var otherNilFunc func(int) string
	var nilInt *int
	for _, symetricalTc := range []testCaseSpec{
		// Different functions
		{
			funcA:         TestAreFuncsEqual,
			funcAName:     "TestAreFuncsEqual",
			funcB:         TestErrorSpec,
			funcBName:     "TestErrorSpec",
			ExpectedMatch: false,
			ExpectedErr:   nil,
		},
		// untyped nils
		{
			funcA:         nil,
			funcAName:     "(untyped) nil",
			funcB:         nil,
			funcBName:     "(untyped) nil",
			ExpectedMatch: true,
			ExpectedErr: &testtools.ErrorSpec{
				Type:    "*errors.errorString",
				Message: "Both values nil when funcs expected",
			},
		},
		// Nil function and untyped nil
		{
			funcA:         nilFunc,
			funcAName:     "(func()) nil",
			funcB:         nil,
			funcBName:     "(untyped) nil",
			ExpectedMatch: false,
			ExpectedErr: &testtools.ErrorSpec{
				Type:    "*errors.errorString",
				Message: "Both values nil when funcs expected",
			},
		},
		// One untyped nil and one valid func
		{
			funcA:         nil,
			funcAName:     "(untyped) nil",
			funcB:         TestErrorSpec,
			funcBName:     "TestErrorSpec",
			ExpectedMatch: false,
			ExpectedErr: &testtools.ErrorSpec{
				Type:    "*errors.errorString",
				Message: "{{FuncAOrdinal}} value nil when two funcs expected",
			},
		},
		// two func() typed nils
		{
			funcA:         nilFunc,
			funcAName:     "(func()) nil",
			funcB:         nilFunc,
			funcBName:     "(func()) nil",
			ExpectedMatch: true,
			ExpectedErr: &testtools.ErrorSpec{
				Type:    "*errors.errorString",
				Message: "Both values nil when funcs expected",
			},
		},
		// two mismatched func typed nils
		{
			funcA:         nilFunc,
			funcAName:     "(func()) nil",
			funcB:         otherNilFunc,
			funcBName:     "(func(int) string) nil",
			ExpectedMatch: false,
			ExpectedErr: &testtools.ErrorSpec{
				Type:    "*errors.errorString",
				Message: "Both values nil when funcs expected",
			},
		},
		// One pointer typed nil
		{
			funcA:         nilInt,
			funcAName:     "(*int) nil",
			funcB:         nilFunc,
			funcBName:     "(func()) nil",
			ExpectedMatch: false,
			ExpectedErr: &testtools.ErrorSpec{
				Type:    "*errors.errorString",
				Message: "Both values nil when funcs expected",
			},
		},
		// One array slice and one valid func
		{
			funcA:         []int{1, 2, 3},
			funcAName:     "[]int{1,2,3}",
			funcB:         TestErrorSpec,
			funcBName:     "TestErrorSpec",
			ExpectedMatch: false,
			ExpectedErr: &testtools.ErrorSpec{
				Type:    "*errors.errorString",
				Message: "{{FuncAOrdinal}} value not a func when two funcs expected",
			},
		},
		// Happy Path
		{
			funcA:         TestErrorSpec,
			funcAName:     "TestErrorSpec",
			funcB:         TestErrorSpec,
			funcBName:     "TestErrorSpec",
			ExpectedMatch: true,
			ExpectedErr:   nil,
		},
	} {
		getErrSpec := genErrSpecGetter(symetricalTc.ExpectedErr)
		for _, testCase := range []testCaseSpec{
			// Forwards
			{
				funcA:         symetricalTc.funcA,
				funcAName:     symetricalTc.funcAName,
				funcB:         symetricalTc.funcB,
				funcBName:     symetricalTc.funcBName,
				ExpectedMatch: symetricalTc.ExpectedMatch,
				ExpectedErr:   getErrSpec("A", symetricalTc.funcAName, "B", symetricalTc.funcBName),
			},
			// Backwards
			{
				funcA:         symetricalTc.funcB,
				funcAName:     symetricalTc.funcBName,
				funcB:         symetricalTc.funcA,
				funcBName:     symetricalTc.funcAName,
				ExpectedMatch: symetricalTc.ExpectedMatch,
				ExpectedErr:   getErrSpec("B", symetricalTc.funcBName, "A", symetricalTc.funcAName),
			},
		} {
			eq, actualErr := testtools.AreFuncsEqual(testCase.funcA, testCase.funcB)
			if eq != testCase.ExpectedMatch {
				actualEquality := "are not"
				expectedEquality := "should"
				if eq {
					actualEquality = "are"
					expectedEquality = "should not"
				}
				t.Errorf(
					"AreFuncsEqual() reports %s() and %s() %s the same function when they %s be",
					testCase.funcAName,
					testCase.funcBName,
					actualEquality,
					expectedEquality)
			}
			if err := testCase.ExpectedErr.EnsureMatches(actualErr); err != nil {
				t.Error(err.Error())
			}
		}
	}
}
