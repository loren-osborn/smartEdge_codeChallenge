package deps

import (
	"io"
	"os"
)

// OsDependencies contains all external dependencies from the os package.
type OsDependencies struct {
	Args      []string
	Stdin     io.Reader
	Stdout    io.Writer
	Stderr    io.Writer
	Exit      func(int)
	Getenv    func(string) string
	Setenv    func(string, string) error
	MkdirAll  func(string, os.FileMode) error
	RemoveAll func(string) error
	Stat      func(string) (os.FileInfo, error)
	Chown     func(string, int, int) error
	Getuid    func() int
}

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
	WriteFile func(string, []byte, os.FileMode) error
	ReadFile  func(string) ([]byte, error)
}

// IoDependencies contains all external dependencies from the io package.
type IoDependencies struct {
	Ioutil IoIoutilDependencies
}

// Dependencies contains all external dependencies injected by main()
// into RealMain()
type Dependencies struct {
	Os     OsDependencies
	Crypto CryptoDependencies
	Io     IoDependencies
}
