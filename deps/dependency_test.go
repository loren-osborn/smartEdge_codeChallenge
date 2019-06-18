package deps_test

import (
	"crypto/rand"
	"github.com/smartedge/codechallenge/deps"
	"github.com/smartedge/codechallenge/testtools"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

// TestEntryPoint verifies that the injected dependencies are properly bound
// to the main code.
func TestDependencies(t *testing.T) {
	if rand.Reader != deps.Defaults.Crypto.Rand.Reader {
		t.Error("deps.Defaults.Crypto.Rand.Reader should resolve to rand.Reader")
	}
	if eq, actualErr := testtools.AreFuncsEqual(ioutil.ReadFile, deps.Defaults.Io.Ioutil.ReadFile); actualErr != nil {
		t.Error(actualErr.Error())
	} else if !eq {
		t.Error("deps.Defaults.Io.Ioutil.ReadFile should evaluate to ioutil.ReadFile")
	}
	if eq, actualErr := testtools.AreFuncsEqual(ioutil.WriteFile, deps.Defaults.Io.Ioutil.WriteFile); actualErr != nil {
		t.Error(actualErr.Error())
	} else if !eq {
		t.Error("deps.Defaults.Io.Ioutil.WriteFile should evaluate to ioutil.WriteFile")
	}
	if !testtools.AreStringSlicesEqual(deps.Defaults.Os.Args, os.Args) {
		t.Error("deps.Defaults.Os.Args should be equal to os.Args")
	}
	if eq, actualErr := testtools.AreFuncsEqual(os.Chdir, deps.Defaults.Os.Chdir); actualErr != nil {
		t.Error(actualErr.Error())
	} else if !eq {
		t.Error("deps.Defaults.Os.Chdir should evaluate to os.Chdir")
	}
	if eq, actualErr := testtools.AreFuncsEqual(os.Chown, deps.Defaults.Os.Chown); actualErr != nil {
		t.Error(actualErr.Error())
	} else if !eq {
		t.Error("deps.Defaults.Os.Chown should evaluate to os.Chown")
	}
	if eq, actualErr := testtools.AreFuncsEqual(os.Exit, deps.Defaults.Os.Exit); actualErr != nil {
		t.Error(actualErr.Error())
	} else if !eq {
		t.Error("deps.Defaults.Os.Exit should evaluate to os.Exit")
	}
	if eq, actualErr := testtools.AreFuncsEqual(os.Getenv, deps.Defaults.Os.Getenv); actualErr != nil {
		t.Error(actualErr.Error())
	} else if !eq {
		t.Error("deps.Defaults.Os.Getenv should evaluate to os.Getenv")
	}
	if eq, actualErr := testtools.AreFuncsEqual(os.Getuid, deps.Defaults.Os.Getuid); actualErr != nil {
		t.Error(actualErr.Error())
	} else if !eq {
		t.Error("deps.Defaults.Os.Getuid should evaluate to os.Getuid")
	}
	if eq, actualErr := testtools.AreFuncsEqual(os.Getwd, deps.Defaults.Os.Getwd); actualErr != nil {
		t.Error(actualErr.Error())
	} else if !eq {
		t.Error("deps.Defaults.Os.Getwd should evaluate to os.Getwd")
	}
	if eq, actualErr := testtools.AreFuncsEqual(os.MkdirAll, deps.Defaults.Os.MkdirAll); actualErr != nil {
		t.Error(actualErr.Error())
	} else if !eq {
		t.Error("deps.Defaults.Os.MkdirAll should evaluate to os.MkdirAll")
	}
	if eq, actualErr := testtools.AreFuncsEqual(os.Open, deps.Defaults.Os.Open); actualErr != nil {
		t.Error(actualErr.Error())
	} else if !eq {
		t.Error("deps.Defaults.Os.Open should evaluate to os.Open")
	}
	if eq, actualErr := testtools.AreFuncsEqual(os.RemoveAll, deps.Defaults.Os.RemoveAll); actualErr != nil {
		t.Error(actualErr.Error())
	} else if !eq {
		t.Error("deps.Defaults.Os.RemoveAll should evaluate to os.RemoveAll")
	}
	if eq, actualErr := testtools.AreFuncsEqual(os.Setenv, deps.Defaults.Os.Setenv); actualErr != nil {
		t.Error(actualErr.Error())
	} else if !eq {
		t.Error("deps.Defaults.Os.Setenv should evaluate to os.Setenv")
	}
	if eq, actualErr := testtools.AreFuncsEqual(os.Stat, deps.Defaults.Os.Stat); actualErr != nil {
		t.Error(actualErr.Error())
	} else if !eq {
		t.Error("deps.Defaults.Os.Stat should evaluate to os.Stat")
	}
	if os.Stderr != deps.Defaults.Os.Stderr {
		t.Error("deps.Defaults.Os.Stderr should resolve to os.Stderr")
	}
	if os.Stdin != deps.Defaults.Os.Stdin {
		t.Error("deps.Defaults.Os.Stdin should resolve to os.Stdin")
	}
	if os.Stdout != deps.Defaults.Os.Stdout {
		t.Error("deps.Defaults.Os.Stdout should resolve to os.Stdout")
	}
	if eq, actualErr := testtools.AreFuncsEqual(filepath.Walk, deps.Defaults.Path.FilePath.Walk); actualErr != nil {
		t.Error(actualErr.Error())
	} else if !eq {
		t.Error("deps.Defaults.Path.FilePath.Walk should evaluate to filepath.Walk")
	}
}
