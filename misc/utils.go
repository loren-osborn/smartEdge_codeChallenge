package misc

import (
	"fmt"
	"github.com/smartedge/codechallenge/deps"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode/utf8"
)

// FileExists reports if a file exists.
func FileExists(d *deps.Dependencies, name string) bool {
	_, err := d.Os.Stat(name)
	return !os.IsNotExist(err)
}

// WriteDirAndFile writes a file at once from a single data buffer. Similar to
// io/ioutil.WriteFile() except ensures all parent directories exist first.
func WriteDirAndFile(d *deps.Dependencies, filename string, data []byte, filePerm os.FileMode, dirPerm os.FileMode) error {
	// This is a hack to keep us from generating root-owned files from within
	// docker.
	//     * Only root can chown files
	//     * Only chown files if environment variable EXT_UID_GID is set
	//     * Only chown files they don't exist
	cleanUpUids := func(err error) error {
		return err
	}
	if (d.Os.Getuid() == 0) && (d.Os.Getenv("EXT_UID_GID") != "") && !FileExists(d, filename) {
		idStrs := strings.Split(d.Os.Getenv("EXT_UID_GID"), ":")
		if len(idStrs) != 2 {
			return fmt.Errorf("Environment variable EXT_UID_GID must have 2 integer ids separated by colons. We found %d", len(idStrs))
		}
		ids := make([]int, len(idStrs))
		idLabels := []string{"user", "group"}
		for i, str := range idStrs {
			intID, err := strconv.Atoi(str)
			if err != nil {
				return fmt.Errorf("%sID %s must be an integer", idLabels[i], str)
			}
			ids[i] = intID
		}
		userID := ids[0]
		groupID := ids[1]
		maxDirDepth := strings.Count(filename, string(os.PathSeparator))
		if maxDirDepth < 3 {
			maxDirDepth = 3
		}
		filesToChown := make([]string, 1, maxDirDepth)
		filesToChown[0] = filename
		for parent := filepath.Dir(filename); (len(parent) > len(string(os.PathSeparator))) && !FileExists(d, parent); parent = filepath.Dir(parent) {
			filesToChown = append(filesToChown, parent)
		}
		cleanUpUids = func(err error) error {
			for _, parentFile := range filesToChown {
				newErr := d.Os.Chown(parentFile, userID, groupID)
				// Only report first error encountered.
				if err == nil {
					err = newErr
				}
			}
			return err
		}
	}
	if !FileExists(d, filepath.Dir(filename)) {
		err := d.Os.MkdirAll(filepath.Dir(filename), dirPerm)
		if err != nil {
			return err
		}
	}
	return cleanUpUids(d.Io.Ioutil.WriteFile(filename, data, filePerm))
}

// TrimRightUTF8Func is based on strings.TrimRightFunc(). It returns a slice of
// the string s with all trailing Unicode code points c satisfying f(c) removed.
// Because UTF-8 isn't one byte per character, we need to slice off one rune
// at a time, instead of one byte
func TrimRightUTF8Func(s string, f func(rune) bool) string {
	c, cSize := utf8.DecodeLastRuneInString(s)
	for (len(s) >= cSize) && f(c) {
		s = s[:len(s)-cSize]
		c, cSize = utf8.DecodeLastRuneInString(s)
	}
	return s
}
