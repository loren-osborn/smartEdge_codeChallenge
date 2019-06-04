package main

import (
	"github.com/smartedge/codechallenge"
	"os"
	"testing"
)

// TestEntryPoint verifies that the injected dependencies are properly bound
// to the main code.
func TestDependencies(t *testing.T) {
	origRealEntryPoint := RealEntryPoint
	depObjs := make([]*codechallenge.Dependencies, 0, 1)
	RealEntryPoint = func(d *codechallenge.Dependencies) int {
		// each call makes depObjs 1 item longer
		depObjs = append(depObjs, d)
		return 0
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
		})
}
