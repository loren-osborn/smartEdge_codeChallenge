package mocks

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"github.com/smartedge/codechallenge"
	"os"
)

// MockDepsBundle is a bundle of dependencies along with a mock environment it
// talks to.
type MockDepsBundle struct {
	Deps        *codechallenge.Dependencies
	OutBuf      *bytes.Buffer
	ErrBuf      *bytes.Buffer
	exitHarness *OsExitHarness
	argList     []string
}

// GetExitStatus returns the value that was passed to mock of os.Exit() or 0 if none.
func (mdb *MockDepsBundle) GetExitStatus() int {
	return mdb.exitHarness.GetExitStatus()
}

// InvokeCallInMockedEnv run passed function, responding liked mocked
// environment.
func (mdb *MockDepsBundle) InvokeCallInMockedEnv(wrapped func()) {
	// Save command line argument state:
	realOsArgsList := os.Args
	realFlagCommandLineUsage := flag.CommandLine.Usage
	realFlagCommandLine := flag.CommandLine
	realFlagErrHelp := flag.ErrHelp
	realFlagUsage := flag.Usage

	// Restore command line argument state: (before return)
	defer func() {
		flag.Usage = realFlagUsage
		flag.ErrHelp = realFlagErrHelp
		flag.CommandLine = realFlagCommandLine
		flag.CommandLine.Usage = realFlagCommandLineUsage
		os.Args = realOsArgsList
	}()

	// Reset command line argument to initial state: (with fake mdb.argList)
	os.Args = mdb.argList
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	flag.CommandLine.Usage = func() {
		flag.Usage()
	}
	flag.ErrHelp = errors.New("flag: help requested")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}

	// Run the code requested:
	mdb.exitHarness.InvokeCallThatMightExit(wrapped)
}

// NewDefaultMockDeps generates a mock environment, along with a
// *codechallenge.Dependencies that operates in this mock environment.
func NewDefaultMockDeps(stdinContent string, cmdLnArgs []string, _ string, _ *map[string]*string) *MockDepsBundle {
	fakeStdout := &bytes.Buffer{}
	fakeStderr := &bytes.Buffer{}
	osExitHarness := NewOsExitMockHarness()
	return &MockDepsBundle{
		Deps: &codechallenge.Dependencies{
			Os: codechallenge.OsDependencies{
				Stdin:  bytes.NewBufferString(stdinContent),
				Stdout: fakeStdout,
				Stderr: fakeStderr,
				Exit:   osExitHarness.GetMock(),
			},
		},
		OutBuf:      fakeStdout,
		ErrBuf:      fakeStderr,
		exitHarness: osExitHarness,
		argList:     cmdLnArgs,
	}
}
