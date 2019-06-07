package mocks_test

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"github.com/smartedge/codechallenge"
	"github.com/smartedge/codechallenge/testtools"
	"github.com/smartedge/codechallenge/testtools/mocks"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

type mockEnvTestCaseInfo struct {
	fakeInContent  string
	fakeOutContent string
	fakeErrContent string
	fakeArgList    []string
	homeDirPath    string
	fakeFileSystem *map[string]*string
	fakeExitStatus int
}

var mockEnvTestCases []mockEnvTestCaseInfo = []mockEnvTestCaseInfo{
	{
		fakeInContent:  "sample input",
		fakeOutContent: "sample output",
		fakeErrContent: "sample error",
		fakeArgList:    []string{"progName"},
		homeDirPath:    "/home/me",
		fakeFileSystem: &map[string]*string{},
		fakeExitStatus: 7,
	},
	{
		fakeInContent:  "other\n\ttest\tinput  \n",
		fakeOutContent: "I have a result!",
		fakeErrContent: "but I stubbed my toe",
		fakeArgList:    []string{"codechallenge", "-rsa"},
		homeDirPath:    "/not_root/",
		fakeFileSystem: &map[string]*string{},
		fakeExitStatus: 3,
	},
}

// TestNewDefaultMockDepsNotNil tests supplier of default mocks for
// CodeChallenge tests is populated.
func TestNewDefaultMockDepsNotNil(t *testing.T) {
	for _, tc := range mockEnvTestCases {
		mockDepsBundle := mocks.NewDefaultMockDeps(
			tc.fakeInContent, tc.fakeArgList, tc.homeDirPath, tc.fakeFileSystem)
		if mockDepsBundle == nil {
			t.Error("mockDepsBundle should never be nil.")
		}
		var deps *codechallenge.Dependencies = mockDepsBundle.Deps
		if deps == nil {
			t.Error("mockDepsBundle.Deps should never be nil.")
		}
		var nativeDeps *codechallenge.Dependencies = mockDepsBundle.NativeDeps
		if nativeDeps == nil {
			t.Error("mockDepsBundle.NativeDeps should never be nil.")
		}
	}
}

// TestNewDefaultMockDepsStdin tests mocks read from Stdin.
func TestNewDefaultMockDepsStdin(t *testing.T) {
	for _, tc := range mockEnvTestCases {
		mockDepsBundle := mocks.NewDefaultMockDeps(
			tc.fakeInContent, tc.fakeArgList, tc.homeDirPath, tc.fakeFileSystem)
		if actualInput, err := ioutil.ReadAll(mockDepsBundle.Deps.Os.Stdin); err != nil {
			t.Error(err.Error())
		} else if string(actualInput) != tc.fakeInContent {
			t.Errorf("mockDepsBundle.Deps.Os.Stdin was expected to supply\n\t%#v but instead yeilded\n\t%#v.", tc.fakeInContent, string(actualInput))
		}
	}
}

// TestNewDefaultMockDepsStdoutAndStderr tests mocks write to fake Stdout and Stderr.
func TestNewDefaultMockDepsStdoutAndStderr(t *testing.T) {
	for _, tc := range mockEnvTestCases {
		mockDepsBundle := mocks.NewDefaultMockDeps(
			tc.fakeInContent, tc.fakeArgList, tc.homeDirPath, tc.fakeFileSystem)
		for _, writerInfo := range []struct {
			content string
			name    string
			writer  io.Writer
			byteBuf *bytes.Buffer
		}{
			{
				content: tc.fakeOutContent,
				name:    "Stdout",
				writer:  mockDepsBundle.Deps.Os.Stdout,
				byteBuf: mockDepsBundle.OutBuf,
			},
			{
				content: tc.fakeErrContent,
				name:    "Stderr",
				writer:  mockDepsBundle.Deps.Os.Stderr,
				byteBuf: mockDepsBundle.ErrBuf,
			},
		} {
			if n, err := io.WriteString(writerInfo.writer, writerInfo.content); err != nil {
				t.Error(err.Error())
			} else {
				if n != len(writerInfo.content) {
					t.Errorf("io.WriteString reported %d bytes written to mockDepsBundle.Deps.Os.%s but %d expected.", n, writerInfo.name, len(writerInfo.content))
				}
				if writerInfo.byteBuf.String() != writerInfo.content {
					t.Errorf("mockDepsBundle.Deps.Os.%s was expected to receive\n\t%#v but instead saw\n\t%#v.", writerInfo.name, writerInfo.content, writerInfo.byteBuf.String())
				}
			}
		}
	}
}

// TestNewDefaultMockDepsExitStatus tests faked exit mechanism is operating.
func TestNewDefaultMockDepsExitStatus(t *testing.T) {
	for _, tc := range mockEnvTestCases {
		mockDepsBundle := mocks.NewDefaultMockDeps(
			tc.fakeInContent, tc.fakeArgList, tc.homeDirPath, tc.fakeFileSystem)

		if mockDepsBundle.GetExitStatus() != 0 {
			t.Errorf("mockDepsBundle.GetExitStatus() should default to 0. Got %#v instead.", mockDepsBundle.GetExitStatus())
		}
		var codeAfterExitExecuted bool
		err := mockDepsBundle.InvokeCallInMockedEnv(func() error {
			mockDepsBundle.Deps.Os.Exit(tc.fakeExitStatus)
			codeAfterExitExecuted = true
			return nil
		})
		if err != nil {
			t.Errorf("Unexpected error calling mockDepsBundle.InvokeCallInMockedEnv(): %s", err.Error())
		}
		if mockDepsBundle.GetExitStatus() != tc.fakeExitStatus {
			t.Errorf("mockDepsBundle.GetExitStatus() should report value of %d passed to mockDepsBundle.Deps.Os.Exit(). Got %#v instead.", tc.fakeExitStatus, mockDepsBundle.GetExitStatus())
		}
		if codeAfterExitExecuted {
			t.Error("Call to mockDepsBundle.Deps.Os.Exit() didn't interrupt flow of execution.")
		}
		dummyMockEnvExitError := errors.New("Dummy Mock environment Exit error.")
		err = mockDepsBundle.InvokeCallInMockedEnv(func() error {
			return dummyMockEnvExitError
		})
		if matchErr := testtools.NewErrorSpecFrom(dummyMockEnvExitError).EnsureMatches(err); matchErr != nil {
			t.Error(matchErr.Error())
		}
	}
}

// TestNewDefaultMockDepsArgsList tests that the run environment command line
// arguments are properly saved, mocked and restored.
func TestNewDefaultMockDepsArgsList(t *testing.T) {
	for _, tc := range mockEnvTestCases {
		realArgList := testtools.CloneStringSlice(os.Args)
		if testtools.AreStringSlicesEqual(os.Args, tc.fakeArgList) {
			t.Errorf("Not STRICTLY an error, but this means \"go test\" has the same args list as tc.fakeArgList which is VERY unlikely! Got tc.fakeArgList == %#v and os.Args == %#v.", tc.fakeArgList, os.Args)
		}
		realFlagCommandLineOutput := flag.CommandLine.Output()
		blackHoleWriter := &bytes.Buffer{}
		flag.CommandLine.SetOutput(blackHoleWriter)
		realFlagCommandLine := flag.CommandLine
		realFlagErrHelp := flag.ErrHelp
		realFlagUsage := flag.Usage
		realFlagCommandLineUsage := flag.CommandLine.Usage
		wrappedRealFlagUsage, realFlagUsageCallCount := testtools.WrapFuncCallWithCounter(flag.Usage)
		wrappedRealFlagCommandLineUsage, realFlagCommandLineUsageCallCount := testtools.WrapFuncCallWithCounter(flag.CommandLine.Usage)
		flag.Usage = wrappedRealFlagUsage
		flag.CommandLine.Usage = wrappedRealFlagCommandLineUsage
		mockDepsBundle := mocks.NewDefaultMockDeps(
			tc.fakeInContent, tc.fakeArgList, tc.homeDirPath, tc.fakeFileSystem)
		err := mockDepsBundle.InvokeCallInMockedEnv(func() error {
			if !testtools.AreStringSlicesEqual(os.Args, tc.fakeArgList) {
				t.Errorf("os.Args should be identical to tc.fakeArgList within the mocked environment.  Got tc.fakeArgList == %#v and os.Args == %#v.", tc.fakeArgList, os.Args)
			}
			if blackHoleWriter == flag.CommandLine.Output() {
				t.Error("flag.CommandLine.Output() mock not installed.")
			}
			actualUsageBuff := &bytes.Buffer{}
			flag.CommandLine.SetOutput(actualUsageBuff)
			unmockedFlagUsageCallDelta := *realFlagUsageCallCount
			flag.Usage()
			unmockedFlagUsageCallDelta = *realFlagUsageCallCount - unmockedFlagUsageCallDelta
			if unmockedFlagUsageCallDelta != 0 {
				t.Error("Mocked flag.Usage() not installed")
			}
			expectedUsageBuff := &bytes.Buffer{}
			flag.CommandLine.SetOutput(expectedUsageBuff)
			fmt.Fprintf(expectedUsageBuff, "Usage of %s:\n", tc.fakeArgList[0])
			flag.PrintDefaults()
			if actualUsageBuff.String() != expectedUsageBuff.String() {
				t.Errorf("Incorrect default Mocked flag.Usage() output:\nexpected:\n\t%#v\nactual:\n\t%#v", expectedUsageBuff.String(), actualUsageBuff.String())
			}
			flag.CommandLine.SetOutput(mockDepsBundle.Deps.Os.Stderr)
			unmockedFlagCommandLineUsageCallDelta := *realFlagCommandLineUsageCallCount
			flag.CommandLine.Usage()
			unmockedFlagCommandLineUsageCallDelta = *realFlagCommandLineUsageCallCount - unmockedFlagCommandLineUsageCallDelta
			if unmockedFlagCommandLineUsageCallDelta != 0 {
				t.Error("Mocked flag.CommandLine.Usage() not installed")
			}
			if realFlagCommandLine == flag.CommandLine {
				t.Error("Mock version of flag.CommandLine not installed")
			}
			if realFlagErrHelp == flag.ErrHelp {
				t.Error("Mock version of flag.ErrHelp not installed")
			}
			return nil
		})
		if err != nil {
			t.Errorf("Unexpected error calling mockDepsBundle.InvokeCallInMockedEnv(): %s", err.Error())
		}
		if !testtools.AreStringSlicesEqual(os.Args, realArgList) {
			t.Errorf("os.Args should be restored to realArgList after exiting the mocked environment.  Got realArgList == %#v and os.Args == %#v.", realArgList, os.Args)
		}
		restoredFlagUsage := flag.Usage
		flag.Usage = func() {}
		flag.CommandLine.Usage()
		flag.Usage = restoredFlagUsage
		if realFlagCommandLine != flag.CommandLine {
			t.Error("Real version of flag.CommandLine not restored")
		}
		if realFlagErrHelp != flag.ErrHelp {
			t.Error("Real version of flag.ErrHelp not restored")
		}
		flagCommandLineUsageCallDelta := *realFlagCommandLineUsageCallCount
		flag.CommandLine.Usage()
		flagCommandLineUsageCallDelta = *realFlagCommandLineUsageCallCount - flagCommandLineUsageCallDelta
		if flagCommandLineUsageCallDelta != 1 {
			t.Error("Real flag.CommandLine.Usage() not restored")
		}
		flagUsageCallDelta := *realFlagUsageCallCount
		flag.Usage()
		flagUsageCallDelta = *realFlagUsageCallCount - flagUsageCallDelta
		if flagUsageCallDelta != 1 {
			t.Error("Real flag.Usage() not restored")
		}

		flag.CommandLine.Usage = realFlagCommandLineUsage
		flag.Usage = realFlagUsage
		flag.CommandLine.SetOutput(realFlagCommandLineOutput)
	}
}

var anticipatedFakeRootPath string = filepath.Join(os.TempDir(), "tmpfs", fmt.Sprintf("sm_codechallenge_test_%d", os.Getpid()))

// TestNewDefaultMockDepsExitStatus tests mocking of the home dir and
// filesystem calls. Due to language constraints, it's turned out to be more
// practical to map filesystem calls into a temporary filesystem directory
// rather than simulating filesystem activity in memory.
func TestNewDefaultMockDepsFileSystem(t *testing.T) {
	realHomeDir := os.Getenv("HOME")
	for _, tc := range mockEnvTestCases {
		mockDepsBundle := mocks.NewDefaultMockDeps(
			tc.fakeInContent, tc.fakeArgList, tc.homeDirPath, tc.fakeFileSystem)
		if _, err := os.Stat(anticipatedFakeRootPath); !os.IsNotExist(err) {
			t.Errorf("Path %#v should not exist yet, but it does.", anticipatedFakeRootPath)
		}
		err := mockDepsBundle.InvokeCallInMockedEnv(func() error {
			if os.Getenv("HOME") != tc.homeDirPath {
				t.Errorf("Mocked HOME path should be %#v, but saw %#v instead", tc.homeDirPath, os.Getenv("HOME"))
			}
			anticipatedFakeHomePath := filepath.Join(anticipatedFakeRootPath, tc.homeDirPath)
			if _, err := os.Stat(anticipatedFakeHomePath); os.IsNotExist(err) {
				t.Errorf("Path %#v should exist inside mock but doesn't", anticipatedFakeHomePath)
			}
			return nil
		})
		if err != nil {
			t.Errorf("Unexpected error calling mockDepsBundle.InvokeCallInMockedEnv(): %s", err.Error())
		}
		if os.Getenv("HOME") != realHomeDir {
			t.Errorf("HOME path should be restored to %#v, but saw %#v instead", realHomeDir, os.Getenv("HOME"))
		}
		if _, err := os.Stat(anticipatedFakeRootPath); !os.IsNotExist(err) {
			t.Errorf("Path %#v should not exist anymore, but it still does.", anticipatedFakeRootPath)
		}
	}
}

// TestNewDefaultMockDepsRogueHomeDir tests rare cases where home directory
// contains illegal \0 characters.
func TestNewDefaultMockDepsRogueHomeDir(t *testing.T) {
	mockDepsBundle := mocks.NewDefaultMockDeps("data", []string{"progname"}, "illegal\x00name", &map[string]*string{})
	realHomeDir := os.Getenv("HOME")
	argCodeRan := false
	err := mockDepsBundle.InvokeCallInMockedEnv(func() error {
		argCodeRan = true
		return nil
	})
	expectedOsErr := &testtools.ErrorSpec{
		Type:    "*os.SyscallError",
		Message: "setenv: invalid argument",
	}
	if matchErr := expectedOsErr.EnsureMatches(err); matchErr != nil {
		t.Error(matchErr.Error())
	}
	if os.Getenv("HOME") != realHomeDir {
		t.Errorf("HOME path should be restored to %#v, but saw %#v instead", realHomeDir, os.Getenv("HOME"))
	}
	if argCodeRan {
		t.Error("closure code should not have run")
	}
	mockDepsBundle = mocks.NewDefaultMockDeps("data", []string{"progname"}, "/some/path", &map[string]*string{})
	dummySetenvError := errors.New("Dummy Setenv() error message")
	setenvCallLog := make([][2]string, 0, 5)
	mockDepsBundle.NativeDeps.Os.Setenv = func(name, val string) error {
		setenvCallLog = append(setenvCallLog, [2]string{name, val})
		if (name != "HOME") || (val != realHomeDir) {
			return os.Setenv(name, val)
		}
		return dummySetenvError
	}
	err = mockDepsBundle.InvokeCallInMockedEnv(func() error {
		argCodeRan = true
		return nil
	})
	expectedDummyErr := testtools.NewErrorSpecFrom(dummySetenvError)
	if matchErr := expectedDummyErr.EnsureMatches(err); matchErr != nil {
		t.Error(matchErr.Error())
	}
	if !argCodeRan {
		t.Error("Error should have occured after closure code ran")
	}
	if len(setenvCallLog) != 2 {
		t.Errorf("mockDepsBundle.NativeDeps.Os.Setenv should have been called twice, but call log shows:\n%#v", setenvCallLog)
	}
	dummyMockEnvExitError := errors.New("Dummy Mock environment Exit error.")
	err = mockDepsBundle.InvokeCallInMockedEnv(func() error {
		return dummyMockEnvExitError
	})
	if matchErr := testtools.NewErrorSpecFrom(dummyMockEnvExitError).EnsureMatches(err); matchErr != nil {
		t.Errorf("Ensure mock environment error isn't clobbered by Setenv error:\n%s", matchErr.Error())
	}
	if len(setenvCallLog) != 4 {
		t.Errorf("mockDepsBundle.NativeDeps.Os.Setenv should have been called a fourth time, after an error in the mock environment, but call log shows:\n%#v", setenvCallLog)
	}
	os.Setenv("HOME", realHomeDir)
}

// TestNewDefaultMockDepsRogueHomeDir tests rare cases where home directory
// contains illegal \0 characters.
func TestNewDefaultMockDepsRogueFileIO(t *testing.T) {
	mockDepsBundle := mocks.NewDefaultMockDeps("data", []string{"progname"}, "/home/someone", &map[string]*string{})
	realHomeDir := os.Getenv("HOME")
	argCodeRan := false
	dummyMkdirAllError := errors.New("Dummy MkdirAll() error message")
	mockDepsBundle.NativeDeps.Os.MkdirAll = func(_ string, _ os.FileMode) error {
		return dummyMkdirAllError
	}
	calledInMockEnv := func() error {
		argCodeRan = true
		return nil
	}
	err := mockDepsBundle.InvokeCallInMockedEnv(calledInMockEnv)
	expectedOsErr := testtools.NewErrorSpecFrom(dummyMkdirAllError)
	if matchErr := expectedOsErr.EnsureMatches(err); matchErr != nil {
		t.Error(matchErr.Error())
	}
	if argCodeRan {
		t.Error("closure code should not have run")
	}
	mockDepsBundle.NativeDeps.Os.MkdirAll = os.MkdirAll
	removeAllCallLog := make([]string, 0, 3)
	mockDepsBundle.NativeDeps.Os.RemoveAll = func(path string) error {
		removeAllCallLog = append(removeAllCallLog, path)
		return errors.New("Dummy RemoveAll() error message")
	}
	expectedOsErr.Message = "Dummy RemoveAll() error message"
	err = mockDepsBundle.InvokeCallInMockedEnv(calledInMockEnv)
	if matchErr := expectedOsErr.EnsureMatches(err); matchErr != nil {
		t.Error(matchErr.Error())
	}
	if !argCodeRan {
		t.Error("Error should have occured after closure code ran")
	}
	if len(removeAllCallLog) != 1 {
		t.Errorf("mockDepsBundle.NativeDeps.Os.Setenv should have been called once, but call log shows:\n%#v", removeAllCallLog)
	}
	dummyMockEnvExitError := errors.New("Dummy Mock environment Exit error.")
	err = mockDepsBundle.InvokeCallInMockedEnv(func() error {
		return dummyMockEnvExitError
	})
	if matchErr := testtools.NewErrorSpecFrom(dummyMockEnvExitError).EnsureMatches(err); matchErr != nil {
		t.Errorf("Ensure mock environment error isn't clobbered by Setenv error:\n%s", matchErr.Error())
	}
	if len(removeAllCallLog) != 2 {
		t.Errorf("mockDepsBundle.NativeDeps.Os.Setenv should have been called after an error in the mock environment, but call log shows:\n%#v", removeAllCallLog)
	}
	cleanupErrs := []error{
		os.RemoveAll(anticipatedFakeRootPath),
		os.Setenv("HOME", realHomeDir),
	}
	expectedOsErr = nil // expect no error for last two checks
	for _, cuErr := range cleanupErrs {
		if matchErr := expectedOsErr.EnsureMatches(cuErr); matchErr != nil {
			t.Error(matchErr.Error())
		}
	}
}
