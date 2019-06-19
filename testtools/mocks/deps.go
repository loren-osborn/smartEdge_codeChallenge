package mocks

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"github.com/smartedge/codechallenge/deps"
	mathRand "math/rand"
	"os"
	"path/filepath"
)

// AllItems tells os.*File.ReadDir to read all items from directory
const (
	AllItems = -1
)

// MockDepsBundle is a bundle of dependencies along with a mock environment it
// talks to.
type MockDepsBundle struct {
	Deps        *deps.Dependencies
	NativeDeps  *deps.Dependencies
	OutBuf      *bytes.Buffer
	ErrBuf      *bytes.Buffer
	exitHarness *OsExitHarness
	argList     []string
	homeDirPath string
	prevCwd     string
	MapPathIn   func(string) (string, error)
	MapPathOut  func(string) (string, error)
	FakeFSRoot  string
	hiddenFiles *map[string]*string
	Files       *map[string]*string
}

// NewDefaultMockDeps generates a mock environment, along with a
// *deps.Dependencies that operates in this mock environment.
// Due to language constraints, it's turned out to be more practical to map
// filesystem calls into a temporary filesystem directory rather than
// simulating filesystem activity in memory.
func NewDefaultMockDeps(stdinContent string, cmdLnArgs []string, homeDir string, files *map[string]*string) *MockDepsBundle {
	fakeStdout := &bytes.Buffer{}
	fakeStderr := &bytes.Buffer{}
	osExitHarness := NewOsExitMockHarness()
	CopyOfDefaultDeps := *deps.Defaults
	if files == nil || *files == nil {
		localMap := make(map[string]*string, 1)
		if files == nil {
			files = &localMap
		} else {
			*files = localMap
		}
	}
	return &MockDepsBundle{
		Deps: &deps.Dependencies{
			Crypto: deps.CryptoDependencies{
				Rand: deps.CryptoRandDependencies{
					Reader: mathRand.New(mathRand.NewSource(0x0123456789abcdef)),
				},
			},
			Io: deps.IoDependencies{
				Ioutil: deps.IoIoutilDependencies{
					ReadFile:  nil,
					WriteFile: nil,
				},
			},
			Os: deps.OsDependencies{
				Args:      cmdLnArgs,
				Chdir:     nil,
				Chown:     nil,
				Exit:      osExitHarness.GetMock(),
				Getenv:    os.Getenv,
				Getuid:    os.Getuid,
				Getwd:     nil,
				MkdirAll:  nil,
				Open:      nil,
				RemoveAll: nil,
				Setenv:    os.Setenv,
				Stat:      nil,
				Stderr:    fakeStderr,
				Stdin:     bytes.NewBufferString(stdinContent),
				Stdout:    fakeStdout,
			},
			Path: deps.PathDependencies{
				FilePath: deps.PathFilePathDependencies{
					Walk: nil,
				},
			},
		},
		NativeDeps:  &CopyOfDefaultDeps,
		OutBuf:      fakeStdout,
		ErrBuf:      fakeStderr,
		exitHarness: osExitHarness,
		argList:     cmdLnArgs,
		homeDirPath: homeDir,
		prevCwd:     "",
		MapPathIn:   nil,
		MapPathOut:  nil,
		hiddenFiles: files,
		Files:       files,
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
	mdb.prevCwd, outErr = os.Getwd()
	if outErr != nil {
		return
	}
	fakeRootPath := filepath.Join(os.TempDir(), "tmpfs", fmt.Sprintf("sm_codechallenge_test_%d", os.Getpid()))
	mdb.FakeFSRoot = fakeRootPath
	outErr = mdb.NativeDeps.Os.MkdirAll(filepath.Join(fakeRootPath, mdb.homeDirPath), 0755)
	if outErr != nil {
		return
	}
	outErr = mdb.NativeDeps.Os.Chdir(filepath.Join(fakeRootPath, mdb.homeDirPath))
	if outErr != nil {
		return
	}
	mdb.MapPathIn = func(path string) (string, error) {
		return filepath.Join(fakeRootPath, path), nil
	}
	mdb.MapPathOut = func(path string) (string, error) {
		cleanFakeRoot := filepath.Clean(fakeRootPath)
		cleanPath := filepath.Clean(path)
		if cleanFakeRoot == cleanPath {
			return "/", nil
		}
		cleanFakeRoot = fmt.Sprintf("%s/", cleanFakeRoot)
		if (len(cleanPath) < len(cleanFakeRoot)) || (cleanPath[0:len(cleanFakeRoot)] != cleanFakeRoot) {
			return "", fmt.Errorf("File %#v not inside fake root %#v when trying to map it outside", cleanPath, cleanFakeRoot)
		}
		return cleanPath[len(cleanFakeRoot)-1:], nil
	}
	mdb.Deps.Io.Ioutil.ReadFile = func(path string) ([]byte, error) {
		realPath, err := mdb.MapPathIn(path)
		if err != nil {
			return nil, err
		}
		return mdb.NativeDeps.Io.Ioutil.ReadFile(realPath)
	}
	mdb.Deps.Io.Ioutil.WriteFile = func(path string, data []byte, perm os.FileMode) error {
		realPath, err := mdb.MapPathIn(path)
		if err != nil {
			return err
		}
		return mdb.NativeDeps.Io.Ioutil.WriteFile(realPath, data, perm)
	}
	mdb.Deps.Os.Chdir = func(dir string) error {
		realDir, err := mdb.MapPathIn(dir)
		if err != nil {
			return err
		}
		return mdb.NativeDeps.Os.Chdir(realDir)
	}
	mdb.Deps.Os.Chown = func(path string, uid, gid int) error {
		realPath, err := mdb.MapPathIn(path)
		if err != nil {
			return err
		}
		return mdb.NativeDeps.Os.Chown(realPath, uid, gid)
	}
	mdb.Deps.Os.Getwd = func() (string, error) {
		realDir, err := mdb.NativeDeps.Os.Getwd()
		if err != nil {
			return "", err
		}
		dir, err := mdb.MapPathOut(realDir)
		if err != nil {
			return "", err
		}
		return dir, nil
	}
	mdb.Deps.Os.MkdirAll = func(path string, perm os.FileMode) error {
		realPath, err := mdb.MapPathIn(path)
		if err != nil {
			return err
		}
		return mdb.NativeDeps.Os.MkdirAll(realPath, perm)
	}
	mdb.Deps.Os.Open = func(path string) (*os.File, error) {
		realPath, err := mdb.MapPathIn(path)
		if err != nil {
			return nil, err
		}
		return mdb.NativeDeps.Os.Open(realPath)
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
	mdb.Deps.Path.FilePath.Walk = func(root string, walkFn filepath.WalkFunc) error {
		realRoot, err := mdb.MapPathIn(root)
		if err != nil {
			return err
		}
		realWalkFunc := filepath.WalkFunc(func(realPath string, info os.FileInfo, err error) error {
			path, err := mdb.MapPathOut(realPath)
			if err != nil {
				return err
			}
			return walkFn(path, info, err)
		})
		return mdb.NativeDeps.Path.FilePath.Walk(realRoot, realWalkFunc)
	}
	// Populate fake file system
	mdb.hiddenFiles = mdb.Files
	// Not kept in sync with mock environment. Set to nil to prevent access.
	mdb.Files = nil
	for path, content := range *mdb.hiddenFiles {
		realPath, err := mdb.MapPathIn(path)
		if err != nil {
			return err
		}
		if content == nil {
			err = mdb.NativeDeps.Os.MkdirAll(realPath, 0755)
			if err != nil {
				return err
			}
		} else {
			err := mdb.NativeDeps.Os.MkdirAll(filepath.Dir(realPath), 0755)
			if err != nil {
				return err
			}
			err = mdb.NativeDeps.Io.Ioutil.WriteFile(realPath, []byte(*content), 0644)
			if err != nil {
				return err
			}
		}
	}

	// Teardown fake filesystem
	defer func() {
		// Restore nil file map to restore visability
		mdb.Files = mdb.hiddenFiles
		*mdb.Files = make(map[string]*string)
		newErr := mdb.NativeDeps.Path.FilePath.Walk(fakeRootPath, func(realPath string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			path, err := mdb.MapPathOut(realPath)
			if err != nil {
				return err
			}
			if info.IsDir() {
				dirFileHandle, err := mdb.NativeDeps.Os.Open(realPath)
				if err != nil {
					return err
				}
				dirFileNames, err := dirFileHandle.Readdirnames(AllItems)
				if err != nil {
					return err
				}
				childCount := 0
				for _, name := range dirFileNames {
					if (name != ".") && (name != "..") {
						childCount++
					}
				}
				if childCount == 0 {
					(*mdb.Files)[path] = nil
				}
				return nil
			}
			fileBuf, err := mdb.NativeDeps.Io.Ioutil.ReadFile(realPath)
			if err != nil {
				return err
			}
			content := string(fileBuf)
			(*mdb.Files)[path] = &content
			return nil
		})
		if outErr == nil {
			outErr = newErr
		}
		newErr = mdb.NativeDeps.Os.Chdir(mdb.prevCwd)
		if outErr == nil {
			outErr = newErr
		}
		newErr = mdb.NativeDeps.Os.RemoveAll(fakeRootPath)
		if outErr == nil {
			outErr = newErr
		}
	}()

	// Run the code requested:
	return mdb.exitHarness.InvokeCallThatMightExit(wrapped)
}
