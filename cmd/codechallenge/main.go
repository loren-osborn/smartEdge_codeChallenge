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
)

// RealEntryPoint is how main() is loosely bound to codechallenge.RealMain()
var RealEntryPoint func(*deps.Dependencies) = codechallenge.RealMain

// main() calls RealEntryPoint (which defaults to codechallenge.RealMain())
// which is RealMain() is called in production, but at testing time, the
// test harness replaces RealEntryPoint with a stub, so both the production
// Dependencies structure, and production RealMain() can be validated
// independantly
func main() {
	RealEntryPoint(&deps.Dependencies{
		Os: deps.OsDependencies{
			Args:      os.Args,
			Stdin:     os.Stdin,
			Stdout:    os.Stdout,
			Stderr:    os.Stderr,
			Exit:      os.Exit,
			Getenv:    os.Getenv,
			Setenv:    os.Setenv,
			MkdirAll:  os.MkdirAll,
			RemoveAll: os.RemoveAll,
			Stat:      os.Stat,
			Chown:     os.Chown,
			Getuid:    os.Getuid,
		},
		Crypto: deps.CryptoDependencies{
			Rand: deps.CryptoRandDependencies{
				Reader: rand.Reader,
			},
		},
		Io: deps.IoDependencies{
			Ioutil: deps.IoIoutilDependencies{
				WriteFile: ioutil.WriteFile,
				ReadFile:  ioutil.ReadFile,
			},
		},
	})
}
