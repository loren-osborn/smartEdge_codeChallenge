package main

import (
	"errors"
	"fmt"
	"github.com/smartedge/codechallenge/testtools"
	"github.com/smartedge/codechallenge/testtools/mocks"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// DirToFilter is the directory to modify
const (
	CSSOrigWithLibPath = `
.treeview-red li { background-image: url(http://localhost:6060/lib/godoc/images/treeview-red-line.gif); }
`
	CSSTranslatedWithLibPath = `
.treeview-red li { background-image: url(https://loren-osborn.github.io/smartEdge_codeChallenge/godoc/lib/godoc/images/treeview-red-line.gif); }
`
)

type TranslationStringPair struct {
	From string
	To   string
}

func GetRepetitions(from, to string) map[string]TranslationStringPair {
	return map[string]TranslationStringPair{
		"Empty string": {
			From: "",
			To:   "",
		},
		"%s Once": {
			From: from,
			To:   to,
		},
		"%s 5 times": {
			From: strings.Repeat(from, 5),
			To:   strings.Repeat(to, 5),
		},
	}
}

// TestWalker traversal of the godoc filesystem.
func TestWalker(t *testing.T) {
	for desc, tc := range map[string]struct {
		homeDir   string
		beforeFs  map[string]*string
		setup     func(mdb *mocks.MockDepsBundle) error
		status    int
		stdOutput testtools.StringMatcher
		stdErr    testtools.StringMatcher
		afterFs   map[string]*string
	}{
		"Simplest test case": {
			homeDir: "/home/foobar",
			beforeFs: map[string]*string{
				"/home/foobar/godoc/fred":   testtools.StringPtr("abc 123"),
				"/home/foobar/godoc/george": testtools.StringPtr(CSSOrigWithLibPath),
			},
			setup: func(mdb *mocks.MockDepsBundle) error {
				return nil
			},
			status:    0,
			stdOutput: testtools.NewStringStringMatcher("modified file: \"/home/foobar/godoc/george\"\n"),
			stdErr:    testtools.NewStringStringMatcher(""),
			afterFs: map[string]*string{
				"/home/foobar/godoc/fred":   testtools.StringPtr("abc 123"),
				"/home/foobar/godoc/george": testtools.StringPtr(CSSTranslatedWithLibPath),
			},
		},
		"Bad Getwd()": {
			homeDir:  "/home/foobar",
			beforeFs: map[string]*string{},
			setup: func(mdb *mocks.MockDepsBundle) error {
				mdb.Deps.Os.Getwd = func() (string, error) {
					return "", errors.New("this is a test")
				}
				return nil
			},
			status:    1,
			stdOutput: testtools.NewStringStringMatcher(""),
			stdErr:    testtools.NewStringStringMatcher("error fetching CWD: this is a test\n"),
			afterFs: map[string]*string{
				"/home/foobar": nil,
			},
		},
		"Bad Walk()": {
			homeDir:  "/home/foobar",
			beforeFs: map[string]*string{},
			setup: func(mdb *mocks.MockDepsBundle) error {
				mdb.Deps.Path.FilePath.Walk = func(root string, walkFn filepath.WalkFunc) error {
					err1 := walkFn(filepath.Join(root, "foo"), nil, errors.New("this is another test"))
					err2 := walkFn(filepath.Join(root, "bar"), &testtools.DummyFileInfo{}, nil)
					return fmt.Errorf("err1: %#v\nerr2: %#v", err1.Error(), err2.Error())
				}
				return nil
			},
			status:    2,
			stdOutput: testtools.NewStringStringMatcher(""),
			stdErr: testtools.NewStringStringMatcher("Failure accessing a path \"/home/foobar/godoc/foo\": this is another test\n" +
				"Error reading file \"/home/foobar/godoc/bar\": open {{FakeFSRoot}}/home/foobar/godoc/bar: no such file or directory\n" +
				"error walking the path \"/home/foobar/godoc\": err1: \"this is another test\"\n" +
				"err2: \"open {{FakeFSRoot}}/home/foobar/godoc/bar: no such file or directory\"\n"),
			afterFs: map[string]*string{
				"/home/foobar": nil,
			},
		},
		"Bad Write": {
			homeDir: "/home/foobar",
			beforeFs: map[string]*string{
				"/home/foobar/godoc/fred":   testtools.StringPtr("abc 123"),
				"/home/foobar/godoc/george": testtools.StringPtr(CSSOrigWithLibPath),
			},
			setup: func(mdb *mocks.MockDepsBundle) error {
				mdb.Deps.Io.Ioutil.WriteFile = func(path string, data []byte, perm os.FileMode) error {
					return errors.New("this is a test")
				}
				return nil
			},
			status:    2,
			stdOutput: testtools.NewStringStringMatcher(""),
			stdErr:    testtools.NewStringStringMatcher("Error writing file \"/home/foobar/godoc/george\": this is a test\nerror walking the path \"/home/foobar/godoc\": this is a test\n"),
			afterFs: map[string]*string{
				"/home/foobar/godoc/fred":   testtools.StringPtr("abc 123"),
				"/home/foobar/godoc/george": testtools.StringPtr(CSSOrigWithLibPath),
			},
		},
	} {
		t.Run(fmt.Sprintf("Subtest: %s", desc), func(tt *testing.T) {
			mockDepsBundle := mocks.NewDefaultMockDeps("", []string{"filterdocs"}, tc.homeDir, &tc.beforeFs)
			err := mockDepsBundle.InvokeCallInMockedEnv(func() error {
				innerErr := tc.setup(mockDepsBundle)
				if innerErr != nil {
					return innerErr
				}
				RealMain(mockDepsBundle.Deps)
				return nil
			})
			if err != nil {
				tt.Errorf("Unexpected error calling mockDepsBundle.InvokeCallInMockedEnv(): %s", err.Error())
			}
			fakeRootPathFixer := strings.NewReplacer(mockDepsBundle.FakeFSRoot, "{{FakeFSRoot}}")
			if exitStatus := mockDepsBundle.GetExitStatus(); exitStatus != tc.status {
				tt.Errorf("RealMain() should have a exit status of %d. Got %#v instead.", tc.status, exitStatus)
			}
			if err := tc.stdOutput.MatchString(fakeRootPathFixer.Replace(mockDepsBundle.OutBuf.String())); err != nil {
				tt.Errorf("Standard Output:\n%#v didn't match:\n%s.", fakeRootPathFixer.Replace(mockDepsBundle.OutBuf.String()), err.Error())
			}
			if err := tc.stdErr.MatchString(fakeRootPathFixer.Replace(mockDepsBundle.ErrBuf.String())); err != nil {
				tt.Errorf("Standard Error:\n%#v didn't match:\n%s.", fakeRootPathFixer.Replace(mockDepsBundle.ErrBuf.String()), err.Error())
			}
			if mockDepsBundle.Files == nil {
				tt.Error("mockDepsBundle.Files is unexpectedly nil")
			} else if !testtools.AreFakeFileSystemsEqual(tc.afterFs, *mockDepsBundle.Files) {
				tt.Errorf("Filesystem doesn't look as expected: we expected:\n%#v\nbut we got:\n%#v", tc.afterFs, mockDepsBundle.Files)
			}
		})
	}
}

// TestConvertContent tests proper translation of different content.
func TestConvertContent(t *testing.T) {
	for _, outerTC := range []struct {
		Description string
		Input       string
		Expected    string
	}{
		{
			Description: "Content containing a http://localhost:6060/ path link",
			Input:       "This is a http://localhost:6060/ test",
			Expected:    "This is a https://golang.org/ test",
		},
		{
			Description: "Content containing a http://localhost:6060/lib/ path link",
			Input:       CSSOrigWithLibPath,
			Expected:    CSSTranslatedWithLibPath,
		},
		{
			Description: "Content containing a http://localhost:6060/src/github.com/smartedge/codechallenge/ path link",
			Input:       `<h2 id="GenerateResponse">func <a href="http://localhost:6060/src/github.com/smartedge/codechallenge/response.go?s=459:576#L9">GenerateResponse</a>`,
			Expected:    `<h2 id="GenerateResponse">func <a href="https://github.com/loren-osborn/smartEdge_codeChallenge/blob/master/response.go#L19">GenerateResponse</a>`,
		},
	} {
		for caseName, tc := range GetRepetitions(outerTC.Input, outerTC.Expected) {
			t.Run(fmt.Sprintf("Testing %s", fmt.Sprintf(caseName, outerTC.Description)), func(tt *testing.T) {
				actual := FilterFileContent(tc.From)
				if tc.To != actual {
					tt.Errorf("File content:\n\t%#v\nnot translated properly. Expected:\n\t%#v\nbut instead got:\n\t%#v", tc.From, tc.To, actual)
				}
			})
		}
	}
}

// // TestRegexps tests proper translation of a localhost:6060/lib/ path.
// func TestRegexps(t *testing.T) {
// 	for _, tc := range []struct {
// 		CaseName    string
// 		Pattern     *regexp.Regexp
// 		Replacement string
// 		From        string
// 		To          string
// 	}{
// 		{
// 			CaseName:    "replace hostname",
// 			Pattern:     Patterns.HostnameToReplace,
// 			Replacement: "MY NEW HOST",
// 			From:        "This is a http://localhost:6060/ test",
// 			To:          "This is a MY NEW HOST test",
// 		},
// 		{
// 			CaseName:    "real-world replace hostname",
// 			Pattern:     Patterns.HostnameToReplace,
// 			Replacement: `https://golang.org/`,
// 			From:        "background-image: url(http://localhost:6060/lib/godoc/images/treeview-red-line.gif); ",
// 			To:          "background-image: url(https://golang.org/lib/godoc/images/treeview-red-line.gif); ",
// 		},
// 	} {
// 		for repeatTestCaseName, rtc := range GetRepetitions(tc.From, tc.To) {
// 			t.Run(fmt.Sprintf(repeatTestCaseName, tc.CaseName), func(tt *testing.T) {
// 				actual := Patterns.HostnameToReplace.ReplaceAllString(rtc.From, tc.Replacement)
// 				if rtc.To != actual {
// 					tt.Errorf("File content:\n\t%#v\nnot translated properly. Expected:\n\t%#v\nbut instead got:\n\t%#v", rtc.From, rtc.To, actual)
// 				}
// 			})
// 		}
// 	}
// }
