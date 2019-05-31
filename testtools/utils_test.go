package testtools_test

import (
	"github.com/smartedge/codechallenge/testtools"
	// "errors"
	// "reflect"
	"testing"
)

// TestErrorSpec tests *ErrorSpec type
func TestErrorSpec(_ *testing.T) {
	var nilErrSpec *testtools.ErrorSpec
	var nilError error
	var _ = nilErrSpec.EnsureMatches(nilError)

	// eq, err := AreFuncsEqual(TestAreFuncsEqual, TestErrorSpec)
	// if eq {
	// 	t.Error("AreFuncsEqual() reports TestAreFuncsEqual() and TestErrorSpec() are the same function")
	// }
	// if err != nil {
	// 	t.Errorf("AreFuncsEqual() should only return an error if passed a non-func: TestAreFuncsEqual() and TestErrorSpec() were passed, and we got error: %T: %#v", err, err.Error())
	// }
	// total := Sum(5, 5)
	// if total != 10 {
	// 	t.Errorf("Sum was incorrect, got: %d, want: %d.", total, 10)
	// }
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
