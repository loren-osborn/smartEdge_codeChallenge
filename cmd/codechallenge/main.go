// Package codechallenge is a stub to call into the main body of the program.
// Having this in a separate main package allows us to test proper use of public
// and private methods, variables and types. It also allows us to properly
// document the public methods with godoc.
package main

import (
	"crypto/rand"
	"github.com/smartedge/codechallenge"
	"github.com/smartedge/codechallenge/deps"
	"io/ioutil"
	"os"
	"path/filepath"
)

// RealEntryPoint is how main() is loosely bound to codechallenge.RealMain()
var RealEntryPoint func(*deps.Dependencies) = codechallenge.RealMain

// main() calls RealEntryPoint, which defaults to codechallenge.RealMain() in
// production. At testing time, the test harness replaces RealEntryPoint with a
// stub, so both the production Dependencies structure, and production
// RealMain() can be validated independantly.
func main() {
	RealEntryPoint(&deps.Dependencies{
		Crypto: deps.CryptoDependencies{
			Rand: deps.CryptoRandDependencies{
				Reader: rand.Reader,
			},
		},
		Io: deps.IoDependencies{
			Ioutil: deps.IoIoutilDependencies{
				ReadFile:  ioutil.ReadFile,
				WriteFile: ioutil.WriteFile,
			},
		},
		Os: deps.OsDependencies{
			Args:      os.Args,
			Chdir:     os.Chdir,
			Chown:     os.Chown,
			Exit:      os.Exit,
			Getenv:    os.Getenv,
			Getuid:    os.Getuid,
			Getwd:     os.Getwd,
			MkdirAll:  os.MkdirAll,
			RemoveAll: os.RemoveAll,
			Setenv:    os.Setenv,
			Stat:      os.Stat,
			Stderr:    os.Stderr,
			Stdin:     os.Stdin,
			Stdout:    os.Stdout,
		},
		Path: deps.PathDependencies{
			FilePath: deps.PathFilePathDependencies{
				Walk: filepath.Walk,
			},
		},
	})
}
