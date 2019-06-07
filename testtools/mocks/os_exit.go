package mocks

// OsExitHarness provides an easy mechanism for simulating an os.Exit in a mock.
type OsExitHarness struct {
	osExitCode int
}

// NewOsExitMockHarness generates new structure to simulate os.Exit() with
func NewOsExitMockHarness() *OsExitHarness {
	return &OsExitHarness{}
}

// GetMock returns the mocked implementation of os.Exit().
func (emh *OsExitHarness) GetMock() func(int) {
	return func(status int) {
		emh.osExitCode = status
		panic(emh)
	}
}

// InvokeCallThatMightExit run passed function, catching any calls to mocked os.Exit().
func (emh *OsExitHarness) InvokeCallThatMightExit(wrapped func() error) error {
	defer func() {
		if r := recover(); r != nil {
			if v, ok := r.(*OsExitHarness); !ok || (v != emh) {
				// not our panic... re-panic it.
				panic(r)
			}
			// Exit was called, and now caught.
			// Continue with not further warnings.
		}
	}()
	return wrapped()
}

// GetExitStatus returns the value that was passed to mock of os.Exit() or 0 if none.
func (emh *OsExitHarness) GetExitStatus() int {
	return emh.osExitCode
}
