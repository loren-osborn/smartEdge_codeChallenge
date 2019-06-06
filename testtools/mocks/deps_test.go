package mocks_test

import (
	"bytes"
	"github.com/smartedge/codechallenge"
	"github.com/smartedge/codechallenge/testtools/mocks"
	"io"
	"io/ioutil"
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
		homeDirPath:    "/root/",
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
		mockDepsBundle.InvokeCallInMockedEnv(func() {
			mockDepsBundle.Deps.Os.Exit(tc.fakeExitStatus)
			codeAfterExitExecuted = true
		})
		if mockDepsBundle.GetExitStatus() != tc.fakeExitStatus {
			t.Errorf("mockDepsBundle.GetExitStatus() should report value of %d passed to mockDepsBundle.Deps.Os.Exit(). Got %#v instead.", tc.fakeExitStatus, mockDepsBundle.GetExitStatus())
		}
		if codeAfterExitExecuted {
			t.Error("Call to mockDepsBundle.Deps.Os.Exit() didn't interrupt flow of execution.")
		}
	}
}
