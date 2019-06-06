package mocks_test

import (
	"github.com/smartedge/codechallenge/testtools/mocks"
	"testing"
)

// TestOsExitMock tests the mock for os.Exit()
func TestOsExitMock(t *testing.T) {
	osExitHarness := mocks.NewOsExitMockHarness()
	var exitStatus int
	if exitStatus = osExitHarness.GetExitStatus(); exitStatus != 0 {
		t.Errorf("osExitHarness.GetExitStatus() should default to 0. Got %#v instead.", exitStatus)
	}
	var mockedOsExit func(int) = osExitHarness.GetMock()
	var codeBeforeExitRan bool
	var codeAfterExitRan bool
	osExitHarness.InvokeCallThatMightExit(func() {
		codeBeforeExitRan = true
		nestedExitHarness := mocks.NewOsExitMockHarness()
		// this check tests that the nested harness has no effect on the osExitHarness
		nestedExitHarness.InvokeCallThatMightExit(func() {
			mockedOsExit(5)
		})
		codeAfterExitRan = true
	})
	if !codeBeforeExitRan {
		t.Error("Code passed to osExitHarness.InvokeCallThatMightExit() should run.")
	}
	if codeAfterExitRan {
		t.Error("Code after mockedOsExit() shouldn't execute.")
	}
	if exitStatus = osExitHarness.GetExitStatus(); exitStatus != 5 {
		t.Errorf("osExitHarness.GetExitStatus() should be set to 5. Got %#v instead.", exitStatus)
	}
}
