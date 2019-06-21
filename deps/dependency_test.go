package deps_test

import (
	"crypto/rand"
	"fmt"
	"github.com/smartedge/codechallenge/deps"
	"github.com/smartedge/codechallenge/testtools"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// TestEntryPoint verifies that the injected dependencies are properly bound
// to the main code.
func TestDependencies(t *testing.T) {
	for _, tc := range []struct {
		OrigName string
		Orig     interface{}
		DepName  string
		Dep      interface{}
	}{
		{
			OrigName: "rand.Reader",
			Orig:     rand.Reader,
			DepName:  "deps.Defaults.Crypto.Rand.Reader",
			Dep:      deps.Defaults.Crypto.Rand.Reader,
		},
		{
			OrigName: "ioutil.ReadFile",
			Orig:     ioutil.ReadFile,
			DepName:  "deps.Defaults.Io.Ioutil.ReadFile",
			Dep:      deps.Defaults.Io.Ioutil.ReadFile,
		},
		{
			OrigName: "ioutil.WriteFile",
			Orig:     ioutil.WriteFile,
			DepName:  "deps.Defaults.Io.Ioutil.WriteFile",
			Dep:      deps.Defaults.Io.Ioutil.WriteFile,
		},
		{
			OrigName: "os.Args",
			Orig:     os.Args,
			DepName:  "deps.Defaults.Os.Args",
			Dep:      deps.Defaults.Os.Args,
		},
		{
			OrigName: "os.Chdir",
			Orig:     os.Chdir,
			DepName:  "deps.Defaults.Os.Chdir",
			Dep:      deps.Defaults.Os.Chdir,
		},
		{
			OrigName: "os.Chown",
			Orig:     os.Chown,
			DepName:  "deps.Defaults.Os.Chown",
			Dep:      deps.Defaults.Os.Chown,
		},
		{
			OrigName: "os.Exit",
			Orig:     os.Exit,
			DepName:  "deps.Defaults.Os.Exit",
			Dep:      deps.Defaults.Os.Exit,
		},
		{
			OrigName: "os.Getenv",
			Orig:     os.Getenv,
			DepName:  "deps.Defaults.Os.Getenv",
			Dep:      deps.Defaults.Os.Getenv,
		},
		{
			OrigName: "os.Getuid",
			Orig:     os.Getuid,
			DepName:  "deps.Defaults.Os.Getuid",
			Dep:      deps.Defaults.Os.Getuid,
		},
		{
			OrigName: "os.Getwd",
			Orig:     os.Getwd,
			DepName:  "deps.Defaults.Os.Getwd",
			Dep:      deps.Defaults.Os.Getwd,
		},
		{
			OrigName: "os.MkdirAll",
			Orig:     os.MkdirAll,
			DepName:  "deps.Defaults.Os.MkdirAll",
			Dep:      deps.Defaults.Os.MkdirAll,
		},
		{
			OrigName: "os.Open",
			Orig:     os.Open,
			DepName:  "deps.Defaults.Os.Open",
			Dep:      deps.Defaults.Os.Open,
		},
		{
			OrigName: "os.RemoveAll",
			Orig:     os.RemoveAll,
			DepName:  "deps.Defaults.Os.RemoveAll",
			Dep:      deps.Defaults.Os.RemoveAll,
		},
		{
			OrigName: "os.Setenv",
			Orig:     os.Setenv,
			DepName:  "deps.Defaults.Os.Setenv",
			Dep:      deps.Defaults.Os.Setenv,
		},
		{
			OrigName: "os.Stat",
			Orig:     os.Stat,
			DepName:  "deps.Defaults.Os.Stat",
			Dep:      deps.Defaults.Os.Stat,
		},
		{
			OrigName: "os.Stderr",
			Orig:     os.Stderr,
			DepName:  "deps.Defaults.Os.Stderr",
			Dep:      deps.Defaults.Os.Stderr,
		},
		{
			OrigName: "os.Stdin",
			Orig:     os.Stdin,
			DepName:  "deps.Defaults.Os.Stdin",
			Dep:      deps.Defaults.Os.Stdin,
		},
		{
			OrigName: "os.Stdout",
			Orig:     os.Stdout,
			DepName:  "deps.Defaults.Os.Stdout",
			Dep:      deps.Defaults.Os.Stdout,
		},
		{
			OrigName: "filepath.Walk",
			Orig:     filepath.Walk,
			DepName:  "deps.Defaults.Path.FilePath.Walk",
			Dep:      deps.Defaults.Path.FilePath.Walk,
		},
		{
			OrigName: "runtime.Caller",
			Orig:     runtime.Caller,
			DepName:  "deps.Defaults.Runtime.Caller",
			Dep:      deps.Defaults.Runtime.Caller,
		},
	} {
		t.Run(fmt.Sprintf("Verifying %s", tc.DepName), func(tt *testing.T) {
			// There are three types of dependancies:
			same := false
			switch origTyped := tc.Orig.(type) {
			case *os.File:
				// This is for os.Stdin, os.Stdout and os.Stderr:
				if depTyped, ok := tc.Dep.(*os.File); ok {
					same = (origTyped == depTyped)
				}
			case []string:
				// This is for os.Args:
				if depTyped, ok := tc.Dep.([]string); ok {
					same = testtools.AreStringSlicesEqual(depTyped, origTyped)
				}
			case io.Reader:
				// This is for rand.Reader:
				if depTyped, ok := tc.Dep.(io.Reader); ok {
					same = (origTyped == depTyped)
				}
			default:
				// This is for everything else, that should be a function.
				if eq, err := testtools.AreFuncsEqual(tc.Orig, tc.Dep); err != nil {
					tt.Error(err.Error())
					return
				} else {
					same = eq
				}
				if !same {
					tt.Errorf("%s should evaluate to %s", tc.DepName, tc.OrigName)
				}
			}
		})

	}
}
