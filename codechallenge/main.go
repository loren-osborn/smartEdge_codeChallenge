// Package codechallenge is a stub to call into the main body of the program.
// Having this in a seperate main package allows us to test proper use of public
// and private methods, variables and types. It also allows us to properly
// document the public methods with godoc.
package main

import (
	"github.com/smartedge/codechallenge"
	"os"
)

// RealEntryPoint is how main() is loosely bound to codechallenge.RealMain()
var RealEntryPoint func(*codechallenge.Dependencies) = codechallenge.RealMain

// main() calls RealEntryPoint (which defaults to codechallenge.RealMain())
// which is RealMain() is called in production, but at testing time, the
// test harness replaces RealEntryPoint with a stub, so both the production
// Dependencies structure, and production RealMain() can be validated
// independantly
func main() {
	RealEntryPoint(&codechallenge.Dependencies{
		Os: codechallenge.OsDependencies{
			Args:      os.Args,
			Stdin:     os.Stdin,
			Stdout:    os.Stdout,
			Stderr:    os.Stderr,
			Exit:      os.Exit,
			Getenv:    os.Getenv,
			Setenv:    os.Setenv,
			MkdirAll:  os.MkdirAll,
			RemoveAll: os.RemoveAll,
		},
	})
}
