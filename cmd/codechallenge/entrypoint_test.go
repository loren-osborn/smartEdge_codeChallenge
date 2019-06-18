package main

import (
	"github.com/smartedge/codechallenge"
	"github.com/smartedge/codechallenge/deps"
	"github.com/smartedge/codechallenge/testtools"
	"testing"
)

// TestEntryPoint verifies that the injected dependencies are properly bound
// to the main code.
func TestEntryPoint(t *testing.T) {
	t.Run(
		"RealEntryPoint properly intialized to codechallenge.RealMain()",
		func(tt *testing.T) {
			if matches, err := testtools.AreFuncsEqual(
				RealEntryPoint, codechallenge.RealMain); err != nil {
				tt.Error(err.Error())
			} else if !matches {
				tt.Error("RealEntryPoint should default to codechallenge.RealMain()")
			}
		})
	t.Run(
		"main() calls RealEntryPoint",
		func(tt *testing.T) {
			origRealEntryPoint := RealEntryPoint
			depObjs := make([]*deps.Dependencies, 0, 1)
			RealEntryPoint = func(d *deps.Dependencies) {
				// each call makes depObjs 1 item longer
				depObjs = append(depObjs, d)
			}
			if len(depObjs) != 0 {
				tt.Errorf("depObjs (%#v) was supposed to be empty, and wasn't", depObjs)
			}
			main()
			if len(depObjs) != 1 {
				tt.Errorf(
					"depObjs (%#v) was supposed to contain 1 item, but instead contained %d",
					depObjs, len(depObjs))
			}
			RealEntryPoint = origRealEntryPoint
		})
}
