package deps

import (
	"crypto/rand"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
)

// CryptoRandDependencies contains all external dependencies from the crypto/rand package.
type CryptoRandDependencies struct {
	Reader io.Reader
}

// CryptoDependencies contains all external dependencies from the crypto package.
type CryptoDependencies struct {
	Rand CryptoRandDependencies
}

// IoIoutilDependencies contains all external dependencies from the io/ioutil package.
type IoIoutilDependencies struct {
	ReadFile  func(string) ([]byte, error)
	WriteFile func(string, []byte, os.FileMode) error
}

// IoDependencies contains all external dependencies from the io package.
type IoDependencies struct {
	Ioutil IoIoutilDependencies
}

// OsDependencies contains all external dependencies from the os package.
type OsDependencies struct {
	Args      []string
	Chdir     func(string) error // Used only by testtools.
	Chown     func(string, int, int) error
	Exit      func(int)
	Getenv    func(string) string
	Getuid    func() int
	Getwd     func() (string, error) // Used only by buildtools.
	MkdirAll  func(string, os.FileMode) error
	Open      func(string) (*os.File, error) // Used only by testtools.
	RemoveAll func(string) error
	Setenv    func(string, string) error
	Stat      func(string) (os.FileInfo, error)
	Stderr    io.Writer
	Stdin     io.Reader
	Stdout    io.Writer
}

// PathFilePathDependencies contains all external dependencies from the path/filepath package.
type PathFilePathDependencies struct {
	Walk func(string, filepath.WalkFunc) error
}

// PathDependencies contains all external dependencies from the path package.
// Used only by buildtools.
type PathDependencies struct {
	FilePath PathFilePathDependencies
}

// RuntimeDependencies contains all external dependencies from the runtime
// package. Used only by testtools.
type RuntimeDependencies struct {
	Caller func(int) (uintptr, string, int, bool)
}

// Dependencies contains all external dependencies injected by main()
// into RealMain(). Used only by buildtools.
type Dependencies struct {
	Crypto  CryptoDependencies
	Io      IoDependencies
	Os      OsDependencies
	Path    PathDependencies    // Used only by buildtools.
	Runtime RuntimeDependencies // Used only by testtools.
}

// Defaults is the default set of injected dependencies
var Defaults = &Dependencies{
	Crypto: CryptoDependencies{
		Rand: CryptoRandDependencies{
			Reader: rand.Reader,
		},
	},
	Io: IoDependencies{
		Ioutil: IoIoutilDependencies{
			ReadFile:  ioutil.ReadFile,
			WriteFile: ioutil.WriteFile,
		},
	},
	Os: OsDependencies{
		Args:      os.Args,
		Chdir:     os.Chdir,
		Chown:     os.Chown,
		Exit:      os.Exit,
		Getenv:    os.Getenv,
		Getuid:    os.Getuid,
		Getwd:     os.Getwd,
		MkdirAll:  os.MkdirAll,
		Open:      os.Open,
		RemoveAll: os.RemoveAll,
		Setenv:    os.Setenv,
		Stat:      os.Stat,
		Stderr:    os.Stderr,
		Stdin:     os.Stdin,
		Stdout:    os.Stdout,
	},
	Path: PathDependencies{
		FilePath: PathFilePathDependencies{
			Walk: filepath.Walk,
		},
	},
	Runtime: RuntimeDependencies{
		Caller: runtime.Caller,
	},
}
