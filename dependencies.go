package codechallenge

import (
	"io"
)

// OsDependencies contains all external dependencies from the os package.
type OsDependencies struct {
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
	Exit   func(int)
}

// Dependencies contains all external dependencies injected by main()
// into RealMain()
type Dependencies struct {
	Os OsDependencies
}
