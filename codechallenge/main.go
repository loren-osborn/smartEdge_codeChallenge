// Package codechallenge is a stub to call into the main body of the program.
// Having this in a seperate main package allows us to test proper use of public
// and private methods, variables and types.
package main

import (
	"github.com/smartedge/codechallenge"
)

var realEntryPoint func()

func main() {
	codechallenge.RealMain()
}

// RealMain
