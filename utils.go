package codechallenge

import (
	"os"
)

// FileExists reports if a file exists.
func FileExists(d *Dependencies, name string) bool {
	_, err := d.Os.Stat(name)
	return !os.IsNotExist(err)
}
