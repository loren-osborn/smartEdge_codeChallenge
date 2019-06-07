package codechallenge

import (
	"io"
	"os"
)

// OsDependencies contains all external dependencies from the os package.
type OsDependencies struct {
	Stdin     io.Reader
	Stdout    io.Writer
	Stderr    io.Writer
	Exit      func(int)
	Getenv    func(string) string
	Setenv    func(string, string) error
	MkdirAll  func(string, os.FileMode) error
	RemoveAll func(string) error
}

// Dependencies contains all external dependencies injected by main()
// into RealMain()
type Dependencies struct {
	Os OsDependencies
}
