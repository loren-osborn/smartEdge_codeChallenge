// Package codechallenge is a stub to call into the main body of the program.
// Having this in a separate main package allows us to test proper use of public
// and private methods, variables and types. It also allows us to properly
// document the public methods with godoc.
package main

import (
	"github.com/smartedge/codechallenge"
	"github.com/smartedge/codechallenge/deps"
)

// RealEntryPoint is how main() is loosely bound to codechallenge.RealMain()
var RealEntryPoint func(*deps.Dependencies) = codechallenge.RealMain

// main() calls RealEntryPoint, which defaults to codechallenge.RealMain() in
// production. At testing time, the test harness replaces RealEntryPoint with a
// stub, so both the production Dependencies structure, and production
// RealMain() can be validated independently.
func main() {
	RealEntryPoint(deps.Defaults)
}
