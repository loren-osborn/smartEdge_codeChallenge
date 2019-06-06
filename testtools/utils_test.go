package testtools_test

import (
	"errors"
	"fmt"
	"github.com/smartedge/codechallenge/testtools"
	"strings"
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

// TestAreFuncsEqual tests AreFuncsEqual() can compare function pointers
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
			funcB:         TestErrorSpecString,
			funcBName:     "TestErrorSpecString",
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
			funcB:         TestErrorSpecString,
			funcBName:     "TestErrorSpecString",
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
			funcB:         TestErrorSpecString,
			funcBName:     "TestErrorSpecString",
			ExpectedMatch: false,
			ExpectedErr: &testtools.ErrorSpec{
				Type:    "*errors.errorString",
				Message: "{{FuncAOrdinal}} value not a func when two funcs expected",
			},
		},
		// Happy Path
		{
			funcA:         TestErrorSpecString,
			funcAName:     "TestErrorSpecString",
			funcB:         TestErrorSpecString,
			funcBName:     "TestErrorSpecString",
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
				ExpectedErr: getErrSpec(
					"A", symetricalTc.funcAName, "B", symetricalTc.funcBName),
			},
			// Backwards
			{
				funcA:         symetricalTc.funcB,
				funcAName:     symetricalTc.funcBName,
				funcB:         symetricalTc.funcA,
				funcBName:     symetricalTc.funcAName,
				ExpectedMatch: symetricalTc.ExpectedMatch,
				ExpectedErr: getErrSpec(
					"B", symetricalTc.funcBName, "A", symetricalTc.funcAName),
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

// TestAreStringSlicesEqual tests AreStringSlicesEqual() properly compares slices of strings
func TestAreStringSlicesEqual(t *testing.T) {
	doReMi := []string{"do", "re", "mi"}
	doReMiDuplicate := []string{"do", "re", "mi"}    // same content as doReMi, elsewhere in memory
	faSoLaTi := []string{"fa", "so", "la", "ti"}     // totally distinct
	nums := []string{"1", "2", "3"}                  // same length (3) as doReMi
	emptySlice := []string{}                         // same length (3) as doReMi
	emptyElementSlice := []string{""}                // empty is always a good test case.
	var zeroSlice []string                           // a nil slice is also important to test
	doReMiWithSpareCapacity := make([]string, 3, 13) // to test slices with different internal configurations
	doReMiWithSpareCapacity[0] = "do"
	doReMiWithSpareCapacity[1] = "re"
	doReMiWithSpareCapacity[2] = "mi"
	if !testtools.AreStringSlicesEqual(doReMi, doReMi) {
		t.Errorf("A string slice must be identical to itself. AreStringSlicesEqual() says it isn't:\n\t%#v\n\t%#v", doReMi, doReMi)
	}
	if !testtools.AreStringSlicesEqual(doReMi, doReMiDuplicate) {
		t.Errorf("Two identical string slices must be equal. AreStringSlicesEqual() says they aren't:\n\t%#v\n\t%#v", doReMi, doReMiDuplicate)
	}
	if testtools.AreStringSlicesEqual(doReMi, faSoLaTi) {
		t.Errorf("Two distinct string slices must not be equal. AreStringSlicesEqual() says they are:\n\t%#v\n\t%#v", doReMi, faSoLaTi)
	}
	if testtools.AreStringSlicesEqual(doReMi, nums) {
		t.Errorf("Two distinct string slices with the same number of items must not be equal. AreStringSlicesEqual() says they are:\n\t%#v\n\t%#v", doReMi, nums)
	}
	if testtools.AreStringSlicesEqual(doReMi, emptySlice) {
		t.Errorf("An empty and a non-empty string slice must not be equal. AreStringSlicesEqual() says they are:\n\t%#v\n\t%#v", doReMi, emptySlice)
	}
	if testtools.AreStringSlicesEqual(doReMi, zeroSlice) {
		t.Errorf("A nil slice and a non-empty string slice must not be equal. AreStringSlicesEqual() says they are:\n\t%#v\n\t%#v", doReMi, zeroSlice)
	}
	if !testtools.AreStringSlicesEqual(emptySlice, []string{}) {
		t.Errorf("Two empty string slices must be equal. AreStringSlicesEqual() says they aren't\n\t%#v\n\t%#v", emptySlice, []string{})
	}
	if testtools.AreStringSlicesEqual(emptySlice, emptyElementSlice) {
		t.Errorf("An empty string slices must not be equal to a string slice with a single empty element. AreStringSlicesEqual() says they aren't\n\t%#v\n\t%#v", emptySlice, emptyElementSlice)
	}
	if !testtools.AreStringSlicesEqual(zeroSlice, nil) {
		t.Errorf("Two nil string slices must be equal. AreStringSlicesEqual() says they aren't\n\t%#v\n\t%#v", zeroSlice, []string(nil))
	}
	if !testtools.AreStringSlicesEqual(emptySlice, []string{}) {
		t.Errorf("Two empty string slices must be equal. AreStringSlicesEqual() says they aren't\n\t%#v\n\t%#v", emptySlice, []string{})
	}
	if testtools.AreStringSlicesEqual(zeroSlice, emptySlice) {
		t.Errorf("A nil slice and an empty slice must not be equal. AreStringSlicesEqual() says they aren\n\t%#v\n\t%#v", zeroSlice, emptySlice)
	}
	if !testtools.AreStringSlicesEqual(doReMi, doReMiWithSpareCapacity) {
		t.Errorf("Two identical string slices with different capacities must be equal. AreStringSlicesEqual() says they aren't:\n\t%#v\n\t%#v", doReMi, doReMiWithSpareCapacity)
	}
}

// TestCloneStringSlice tests CloneStringSlice() properly clones a slice of strings
func TestCloneStringSlice(t *testing.T) {
	for _, tc := range []struct {
		slice []string
		desc  string
	}{
		{
			slice: []string{"do", "re", "mi"},
			desc:  "a non-empty string slice",
		},
		{
			slice: []string{""},
			desc:  "a string slice with a single empty element",
		},
		{
			slice: []string{},
			desc:  "an empty string slice",
		},
		{
			slice: nil,
			desc:  "a nil string slice",
		},
	} {
		t.Run(
			fmt.Sprintf("Ensuring CloneStringSlice() can clone %s", tc.desc),
			func(tt *testing.T) {
				if !testtools.AreStringSlicesEqual(tc.slice, tc.slice) {
					tt.Errorf("%s isn't equal to itself.\n\t%#v", tc.desc, tc.slice)
				}
				clone := testtools.CloneStringSlice(tc.slice)
				if !testtools.AreStringSlicesEqual(tc.slice, clone) {
					tt.Errorf("%s isn't equal to itself.\n\t%#v\n\t%#v", tc.desc, tc.slice, clone)
				}
				if len(tc.slice) > 0 {
					naiveCopy := tc.slice
					tc.slice[0] = "MODIFIED"
					if !testtools.AreStringSlicesEqual(tc.slice, naiveCopy) {
						tt.Errorf("%s and naiveCopy should have changed.\n\t%#v\n\t%#v", tc.desc, tc.slice, naiveCopy)
					}
					if testtools.AreStringSlicesEqual(tc.slice, clone) {
						tt.Errorf("%s should have changed, but clone should not have.\n\t%#v\n\t%#v", tc.desc, tc.slice, clone)
					}
				}
			},
		)
	}
}

// TestWrapFuncCallWithCounter tests testtools.WrapFuncCallWithCounter()
// properly adds a call counter to the provided function
func TestWrapFuncCallWithCounter(t *testing.T) {
	aExpected := 0
	bExpected := 0
	aCounter := 0
	bCounter := 0
	aFunc := func() {
		aCounter++
	}
	bFunc := func() {
		bCounter++
	}
	aWrapped, aPtrCounter := testtools.WrapFuncCallWithCounter(aFunc)
	bWrapped, bPtrCounter := testtools.WrapFuncCallWithCounter(bFunc)
	if aCounter != 0 {
		t.Errorf("aFunc should have run zero times. Instead ran %d", aCounter)
	}
	if bCounter != 0 {
		t.Errorf("bFunc should have run zero times. Instead ran %d", bCounter)
	}
	if *aPtrCounter != 0 {
		t.Errorf("aPtrCounter reports aFunc ran %d times, but never ran.", *aPtrCounter)
	}
	if *bPtrCounter != 0 {
		t.Errorf("aPtrCounter reports aFunc ran %d times, but never ran.", *aPtrCounter)
	}
	for i := 0; i < 5; i++ {
		aWrapped()
		aExpected++
		if aCounter != aExpected {
			t.Errorf("aFunc should have run %d times. Instead ran %d", aExpected, aCounter)
		}
		if *aPtrCounter != aCounter {
			t.Errorf("aPtrCounter reports aFunc ran %d times, but really ran %d.", *aPtrCounter, aCounter)
		}
		for j := 0; j <= 5; j++ {
			bWrapped()
			bExpected++
			if bCounter != bExpected {
				t.Errorf("bFunc should have run %d times. Instead ran %d", bExpected, bCounter)
			}
			if *bPtrCounter != bCounter {
				t.Errorf("bPtrCounter reports bFunc ran %d times, but really ran %d.", *bPtrCounter, bCounter)
			}

		}
	}

}
