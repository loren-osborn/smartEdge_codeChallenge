package main

import (
	// "github.com/smartedge/codechallenge"
	// "errors"
	// "reflect"
	"testing"
)

// TestProperlyCallMainCode verifies that the injected dependancies are
// properly passed to main code.
func TestProperlyCallMainCode(_ *testing.T) {
	realEntryPoint = getMockMainEntryPoint()
	// total := Sum(5, 5)
	// if total != 10 {
	// 	t.Errorf("Sum was incorrect, got: %d, want: %d.", total, 10)
	// }
}

func getMockMainEntryPoint() func() {
	return func() {}
}
