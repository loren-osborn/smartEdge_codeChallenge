package main

import (
	"bytes"
	"github.com/smartedge/codechallenge"
	"github.com/smartedge/codechallenge/testtools"
	"github.com/smartedge/codechallenge/testtools/mocks"
	"os"
	"testing"
)

// TestEntryPoint verifies that the injected dependencies are properly bound
// to the main code.
func TestDependencies(t *testing.T) {
	origRealEntryPoint := RealEntryPoint
	depObjs := make([]*codechallenge.Dependencies, 0, 1)
	RealEntryPoint = func(d *codechallenge.Dependencies) {
		// each call makes depObjs 1 item longer
		depObjs = append(depObjs, d)
	}
	main()
	RealEntryPoint = origRealEntryPoint
	if len(depObjs) < 1 {
		t.Fatal("No dependency object found.\nCan not continue without it.")
	}
	t.Run(
		"Verify dependencies: os.Stdin, os.Stdout, os.Stderr",
		func(tt *testing.T) {
			if os.Stdin != depObjs[0].Os.Stdin {
				tt.Error("d.Os.Stdin should resolve to os.Stdin")
			}
			if os.Stdout != depObjs[0].Os.Stdout {
				tt.Error("d.Os.Stdout should resolve to os.Stdout")
			}
			if os.Stderr != depObjs[0].Os.Stderr {
				tt.Error("d.Os.Stderr should resolve to os.Stderr")
			}
			if eq, actualErr := testtools.AreFuncsEqual(os.Exit, depObjs[0].Os.Exit); actualErr != nil {
				tt.Error(actualErr.Error())
			} else if !eq {
				tt.Error("d.Os.Exit should evaluate to os.Exit")
			}
		})
}

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
