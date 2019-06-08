package codechallenge_test

import (
	"github.com/smartedge/codechallenge"
	"github.com/smartedge/codechallenge/testtools/mocks"
	"testing"
)

// TestCallingMainWithMocks verifies that calling RealMain with mocked
// dependencies works as intended.
func TestCallingMainWithMocks(t *testing.T) {
	mockDepsBundle := mocks.NewDefaultMockDeps("Four score and seven years ago...", []string{"myProg"}, "/home/foobar", nil)
	err := mockDepsBundle.InvokeCallInMockedEnv(func() error {
		codechallenge.RealMain(mockDepsBundle.Deps)
		return nil
	})
	if err != nil {
		t.Errorf("Unexpected error calling mockDepsBundle.InvokeCallInMockedEnv(): %s", err.Error())
	}
	if exitStatus := mockDepsBundle.GetExitStatus(); exitStatus != 0 {
		t.Errorf("RealMain() should have a normal exit status of 0. Got %#v instead.", exitStatus)
	}
}
