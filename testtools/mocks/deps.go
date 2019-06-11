package mocks

import (
	"bytes"
	cryptoRand "crypto/rand"
	"errors"
	"flag"
	"fmt"
	"github.com/smartedge/codechallenge"
	"io/ioutil"
	mathRand "math/rand"
	"os"
	"path/filepath"
)

// MockDepsBundle is a bundle of dependencies along with a mock environment it
// talks to.
type MockDepsBundle struct {
	Deps        *codechallenge.Dependencies
	NativeDeps  *codechallenge.Dependencies
	OutBuf      *bytes.Buffer
	ErrBuf      *bytes.Buffer
	exitHarness *OsExitHarness
	argList     []string
	homeDirPath string
	MapPathIn   func(string) (string, error)
	MapPathOut  func(string) (string, error)
}

// NewDefaultMockDeps generates a mock environment, along with a
// *codechallenge.Dependencies that operates in this mock environment.
// Due to language constraints, it's turned out to be more practical to map
// filesystem calls into a temporary filesystem directory rather than
// simulating filesystem activity in memory.
func NewDefaultMockDeps(stdinContent string, cmdLnArgs []string, homeDir string, _ *map[string]*string) *MockDepsBundle {
	fakeStdout := &bytes.Buffer{}
	fakeStderr := &bytes.Buffer{}
	osExitHarness := NewOsExitMockHarness()
	return &MockDepsBundle{
		Deps: &codechallenge.Dependencies{
			Os: codechallenge.OsDependencies{
				Args:      cmdLnArgs,
				Stdin:     bytes.NewBufferString(stdinContent),
				Stdout:    fakeStdout,
				Stderr:    fakeStderr,
				Exit:      osExitHarness.GetMock(),
				Getenv:    os.Getenv,
				Setenv:    os.Setenv,
				MkdirAll:  nil,
				RemoveAll: nil,
				Stat:      nil,
				Chown:     nil,
				Getuid:    os.Getuid,
			},
			Crypto: codechallenge.CryptoDependencies{
				Rand: codechallenge.CryptoRandDependencies{
					Reader: mathRand.New(mathRand.NewSource(0x0123456789abcdef)),
				},
			},
			Io: codechallenge.IoDependencies{
				Ioutil: codechallenge.IoIoutilDependencies{
					WriteFile: nil,
					ReadFile:  nil,
				},
			},
		},
		NativeDeps: &codechallenge.Dependencies{
			Os: codechallenge.OsDependencies{
				Args:      os.Args,
				Stdin:     os.Stdin,
				Stdout:    os.Stdout,
				Stderr:    os.Stderr,
				Exit:      os.Exit,
				Getenv:    os.Getenv,
				Setenv:    os.Setenv,
				MkdirAll:  os.MkdirAll,
				RemoveAll: os.RemoveAll,
				Stat:      os.Stat,
				Chown:     os.Chown,
				Getuid:    os.Getuid,
			},
			Crypto: codechallenge.CryptoDependencies{
				Rand: codechallenge.CryptoRandDependencies{
					Reader: cryptoRand.Reader,
				},
			},
			Io: codechallenge.IoDependencies{
				Ioutil: codechallenge.IoIoutilDependencies{
					WriteFile: ioutil.WriteFile,
					ReadFile:  ioutil.ReadFile,
				},
			},
		},
		OutBuf:      fakeStdout,
		ErrBuf:      fakeStderr,
		exitHarness: osExitHarness,
		argList:     cmdLnArgs,
		homeDirPath: homeDir,
	}
}

// GetExitStatus returns the value that was passed to mock of os.Exit() or 0 if none.
func (mdb *MockDepsBundle) GetExitStatus() int {
	return mdb.exitHarness.GetExitStatus()
}

// InvokeCallInMockedEnv run passed function, responding liked mocked
// environment.
func (mdb *MockDepsBundle) InvokeCallInMockedEnv(wrapped func() error) (outErr error) {
	// Save command line argument state:
	realOsArgsList := os.Args
	realFlagCommandLineUsage := flag.CommandLine.Usage
	realFlagCommandLine := flag.CommandLine
	realFlagErrHelp := flag.ErrHelp
	realFlagUsage := flag.Usage
	realHomeDir := mdb.NativeDeps.Os.Getenv("HOME")

	// Restore command line argument state: (before return)
	defer func() {
		setenvErr := mdb.NativeDeps.Os.Setenv("HOME", realHomeDir)
		if outErr == nil {
			outErr = setenvErr
		}
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
	outErr = mdb.NativeDeps.Os.Setenv("HOME", mdb.homeDirPath)
	if outErr != nil {
		return
	}

	// Setup fake filesystem
	fakeRootPath := filepath.Join(os.TempDir(), "tmpfs", fmt.Sprintf("sm_codechallenge_test_%d", os.Getpid()))
	outErr = mdb.NativeDeps.Os.MkdirAll(filepath.Join(fakeRootPath, mdb.homeDirPath), 0755)
	if outErr != nil {
		return
	}
	mdb.MapPathIn = func(path string) (string, error) {
		return filepath.Join(fakeRootPath, path), nil
	}
	mdb.MapPathOut = func(path string) (string, error) {
		cleanFakeRoot := filepath.Clean(fakeRootPath) + "/"
		cleanPath := filepath.Clean(path)
		if cleanPath[0:len(cleanFakeRoot)] != cleanFakeRoot {
			return "", fmt.Errorf("File %#v not inside fake root %#v when trying to map it outside", cleanPath, cleanFakeRoot)
		}
		return cleanPath[len(cleanFakeRoot)-1:], nil
	}
	mdb.Deps.Os.MkdirAll = func(path string, perm os.FileMode) error {
		realPath, err := mdb.MapPathIn(path)
		if err != nil {
			return err
		}
		return mdb.NativeDeps.Os.MkdirAll(realPath, perm)
	}
	mdb.Deps.Os.RemoveAll = func(path string) error {
		realPath, err := mdb.MapPathIn(path)
		if err != nil {
			return err
		}
		return mdb.NativeDeps.Os.RemoveAll(realPath)
	}
	mdb.Deps.Os.Stat = func(path string) (os.FileInfo, error) {
		realPath, err := mdb.MapPathIn(path)
		if err != nil {
			return nil, err
		}
		return mdb.NativeDeps.Os.Stat(realPath)
	}
	mdb.Deps.Os.Chown = func(path string, uid, gid int) error {
		realPath, err := mdb.MapPathIn(path)
		if err != nil {
			return err
		}
		return mdb.NativeDeps.Os.Chown(realPath, uid, gid)
	}
	mdb.Deps.Io.Ioutil.WriteFile = func(path string, data []byte, perm os.FileMode) error {
		realPath, err := mdb.MapPathIn(path)
		if err != nil {
			return err
		}
		return mdb.NativeDeps.Io.Ioutil.WriteFile(realPath, data, perm)
	}
	mdb.Deps.Io.Ioutil.ReadFile = func(path string) ([]byte, error) {
		realPath, err := mdb.MapPathIn(path)
		if err != nil {
			return nil, err
		}
		return mdb.NativeDeps.Io.Ioutil.ReadFile(realPath)
	}

	// Teardown fake filesystem
	defer func() {
		removeAllErr := mdb.NativeDeps.Os.RemoveAll(fakeRootPath)
		if outErr == nil {
			outErr = removeAllErr
		}
	}()

	// Run the code requested:
	return mdb.exitHarness.InvokeCallThatMightExit(wrapped)
}
