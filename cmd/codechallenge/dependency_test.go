package main

import (
	"github.com/smartedge/codechallenge/deps"
	"github.com/smartedge/codechallenge/testtools"
	"os"
	"testing"
)

// TestEntryPoint verifies that the injected dependencies are properly bound
// to the main code.
func TestDependencies(t *testing.T) {
	origRealEntryPoint := RealEntryPoint
	depObjs := make([]*deps.Dependencies, 0, 1)
	RealEntryPoint = func(d *deps.Dependencies) {
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
			if eq, actualErr := testtools.AreFuncsEqual(os.Getenv, depObjs[0].Os.Getenv); actualErr != nil {
				tt.Error(actualErr.Error())
			} else if !eq {
				tt.Error("d.Os.Getenv should evaluate to os.Getenv")
			}
			if eq, actualErr := testtools.AreFuncsEqual(os.Setenv, depObjs[0].Os.Setenv); actualErr != nil {
				tt.Error(actualErr.Error())
			} else if !eq {
				tt.Error("d.Os.Setenv should evaluate to os.Setenv")
			}
			if eq, actualErr := testtools.AreFuncsEqual(os.MkdirAll, depObjs[0].Os.MkdirAll); actualErr != nil {
				tt.Error(actualErr.Error())
			} else if !eq {
				tt.Error("d.Os.MkdirAll should evaluate to os.MkdirAll")
			}
			if eq, actualErr := testtools.AreFuncsEqual(os.MkdirAll, depObjs[0].Os.MkdirAll); actualErr != nil {
				tt.Error(actualErr.Error())
			} else if !eq {
				tt.Error("d.Os.MkdirAll should evaluate to os.MkdirAll")
			}
			if eq, actualErr := testtools.AreFuncsEqual(os.RemoveAll, depObjs[0].Os.RemoveAll); actualErr != nil {
				tt.Error(actualErr.Error())
			} else if !eq {
				tt.Error("d.Os.RemoveAll should evaluate to os.RemoveAll")
			}
		})
}
