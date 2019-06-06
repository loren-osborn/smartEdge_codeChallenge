package codechallenge_test

import (
	"bytes"
	"github.com/smartedge/codechallenge"
	"github.com/smartedge/codechallenge/testtools/mocks"
	"testing"
)

// TestCallingMainWithMocks verifies that calling RealMain with mocked
// dependencies works as intended.
func TestCallingMainWithMocks(t *testing.T) {
	osExitHarness := mocks.NewOsExitMockHarness()
	codechallenge.RealMain(&codechallenge.Dependencies{
		Os: codechallenge.OsDependencies{
			Stdin:  &bytes.Buffer{},
			Stdout: &bytes.Buffer{},
			Stderr: &bytes.Buffer{},
			Exit:   osExitHarness.GetMock(),
		},
	})
	if exitStatus := osExitHarness.GetExitStatus(); exitStatus != 0 {
		t.Errorf("RealMain() should have a normal exit status of 0. Got %#v instead.", exitStatus)
	}
}
