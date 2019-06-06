package mocks

import (
	"bytes"
	"github.com/smartedge/codechallenge"
)

// MockDepsBundle is a bundle of dependencies along with a mock environment it
// talks to.
type MockDepsBundle struct {
	Deps        *codechallenge.Dependencies
	OutBuf      *bytes.Buffer
	ErrBuf      *bytes.Buffer
	exitHarness *OsExitHarness
}

// GetExitStatus returns the value that was passed to mock of os.Exit() or 0 if none.
func (mdb *MockDepsBundle) GetExitStatus() int {
	return mdb.exitHarness.GetExitStatus()
}

// InvokeCallInMockedEnv run passed function, responding liked mocked
// environment.
func (mdb *MockDepsBundle) InvokeCallInMockedEnv(wrapped func()) {
	mdb.exitHarness.InvokeCallThatMightExit(wrapped)
}

// NewDefaultMockDeps generates a mock environment, along with a
// *codechallenge.Dependencies that operates in this mock environment.
func NewDefaultMockDeps(stdinContent string, _ []string, _ string, _ *map[string]*string) *MockDepsBundle {
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
	}
}
