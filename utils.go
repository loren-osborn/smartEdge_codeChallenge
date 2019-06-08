package codechallenge

import (
	"os"
	"path/filepath"
)

// FileExists reports if a file exists.
func FileExists(d *Dependencies, name string) bool {
	_, err := d.Os.Stat(name)
	return !os.IsNotExist(err)
}

// WriteDirAndFile writes a file at once from a single data buffer. Similar to
// io/ioutil.WriteFile() except ensures all parent directories exist first.
func WriteDirAndFile(d *Dependencies, filename string, data []byte, filePerm os.FileMode, dirPerm os.FileMode) error {
	if !FileExists(d, filepath.Dir(filename)) {
		err := d.Os.MkdirAll(filepath.Dir(filename), dirPerm)
		if err != nil {
			return err
		}
	}
	return d.Io.Ioutil.WriteFile(filename, data, filePerm)
}
