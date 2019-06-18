package deps

import (
	"io"
	"os"
	"path/filepath"
)

// CryptoRandDependencies contains all external dependencies from the crypto/rand package.
type CryptoRandDependencies struct {
	Reader io.Reader
}

// CryptoDependencies contains all external dependencies from the crypto package.
type CryptoDependencies struct {
	Rand CryptoRandDependencies
}

// IoIoutilDependencies contains all external dependencies from the io/ioutils package.
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
	Chdir     func(string) error // unused. For Getwd()
	Chown     func(string, int, int) error
	Exit      func(int)
	Getenv    func(string) string
	Getuid    func() int
	Getwd     func() (string, error) // Used only by buildtools.
	MkdirAll  func(string, os.FileMode) error
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

// Dependencies contains all external dependencies injected by main()
// into RealMain(). Used only by buildtools.
type Dependencies struct {
	Crypto CryptoDependencies
	Io     IoDependencies
	Os     OsDependencies
	Path   PathDependencies // Used only by buildtools.
}
