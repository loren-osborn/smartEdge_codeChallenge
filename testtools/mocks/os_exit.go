package mocks

// OsExitHarness provides an easy mechanism for simulating an os.Exit in a mock.
type OsExitHarness interface {
	GetMock() func(int)
	InvokeCallThatMightExit(func())
	GetExitStatus() int
}

// osExitHarnessStruct is the implementation of OsExitHarness
type osExitHarnessStruct struct {
	osExitCode int
}

// GetMock returns the mocked implementation of os.Exit().
func (emh *osExitHarnessStruct) GetMock() func(int) {
	return func(status int) {
		emh.osExitCode = status
		panic(emh)
	}
}

// InvokeCallThatMightExit run passed function, catching any calls to mocked os.Exit().
func (emh *osExitHarnessStruct) InvokeCallThatMightExit(wrapped func()) {
	defer func() {
		if r := recover(); r != nil {
			if v, ok := r.(*osExitHarnessStruct); !ok || (v != emh) {
				// not our panic... re-panic it.
				panic(r)
			}
			// Exit was called, and now caught.
			// Continue with not further warnings.
		}
	}()
	wrapped()
}

// GetExitStatus returns the value that was passed to mock of os.Exit() or 0 if none.
func (emh *osExitHarnessStruct) GetExitStatus() int {
	return emh.osExitCode
}

// NewOsExitMockHarness generates new structure to simulate os.Exit() with
func NewOsExitMockHarness() OsExitHarness {
	return &osExitHarnessStruct{}
}
