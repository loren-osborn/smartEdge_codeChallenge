package main

import (
	"crypto/rand"
	"fmt"
	"github.com/smartedge/codechallenge/deps"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
)

// DirToFilter is the directory to modify
const (
	DirToFilter = "godoc"
)

// PackageURI is the full package name as Go knows it
const (
	PackageURI = "github.com/smartedge/codechallenge"
)

// GHPagesURL is the URL for the project's GH-Pages
const (
	GHPagesURL = "https://loren-osborn.github.io/smartEdge_codeChallenge"
)

// PackageSourceCodeBrowsableRoot is the URL to browsable source code on
// GitHub. Note that this typically ends in "/blob/" (branch name or commit
// hash). This is the version of the source being browsed.
const (
	PackageSourceCodeBrowsableRoot = "https://github.com/loren-osborn/smartEdge_codeChallenge/blob/master"
)

// LicenseName is the name of the license the code is being licensed as. For
// now I chose MIT License, but I may switch back to the same BSD license that
// Go uses.
const (
	LicenseName = "MIT License"
)

// RealEntryPoint is how main() is loosely bound to RealMain()
var RealEntryPoint func(*deps.Dependencies) = RealMain

// main() calls RealEntryPoint, which defaults to RealMain() in production. At
// testing time, the test harness replaces RealEntryPoint with a stub, so both
// the production Dependencies structure, and production RealMain() can be
// validated independantly.
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

// RealMain traverses filesystem, transforming all files encountered.
func RealMain(d *deps.Dependencies) {
	cwd, err := d.Os.Getwd()
	if err != nil {
		fmt.Fprintf(d.Os.Stderr, "error fetching CWD: %v\n", err)
		d.Os.Exit(1)
	}
	rootPath := filepath.Join(cwd, DirToFilter)
	err = d.Path.FilePath.Walk(
		rootPath,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				fmt.Fprintf(d.Os.Stderr, "Failure accessing a path %q: %v\n", path, err)
				return err
			}
			if info.IsDir() {
				return nil
			}
			origBuf, err := ioutil.ReadFile(path)
			if err != nil {
				fmt.Fprintf(d.Os.Stderr, "Error reading file %q: %v\n", path, err)
				return err
			}
			origStr := string(origBuf)
			filteredStr := FilterFileContent(origStr)
			if filteredStr != origStr {
				err = ioutil.WriteFile(path, []byte(filteredStr), info.Mode())
				if err != nil {
					fmt.Fprintf(d.Os.Stderr, "Error writing file %q: %v\n", path, err)
					return err
				}
				fmt.Printf("modified file: %q\n", path)
			}
			return nil
		})
	if err != nil {
		fmt.Fprintf(d.Os.Stderr, "error walking the path %q: %v\n", rootPath, err)
		d.Os.Exit(2)
	}
}

// TranslationRegExps are the regular expressions that we can precompile.
type TranslationRegExps struct {
	HostnameToReplace    *regexp.Regexp
	LibPathToReplace     *regexp.Regexp
	PkgPathToReplace     *regexp.Regexp
	SrcPathToReplace     *regexp.Regexp
	SrcLineNumToReplace  *regexp.Regexp
	SEqualsToRemove      *regexp.Regexp
	LineNoToModify       *regexp.Regexp
	LicenseURLToReplace  *regexp.Regexp
	LicenseNameToReplace *regexp.Regexp
}

// Patterns only need to be compiled once:
var Patterns = TranslationRegExps{
	HostnameToReplace: regexp.MustCompile(`\bhttp://localhost:6060/`),
	LibPathToReplace:  regexp.MustCompile(`\bhttps://golang\.org/lib/`),
	PkgPathToReplace: regexp.MustCompile(fmt.Sprintf(
		`\bhttps://golang\.org/pkg/%s/`,
		regexp.QuoteMeta(PackageURI))),
	SrcPathToReplace: regexp.MustCompile(fmt.Sprintf(
		`\bhttps://golang\.org/src/%s/`,
		regexp.QuoteMeta(PackageURI))),
	SrcLineNumToReplace: regexp.MustCompile(fmt.Sprintf(
		`\b%s/[^?]+\?s=\d+:\d+#L\d+\b`,
		regexp.QuoteMeta(PackageSourceCodeBrowsableRoot))),
	SEqualsToRemove:      regexp.MustCompile(`\?s=\d+:\d+#L`),
	LineNoToModify:       regexp.MustCompile(`\d+$`),
	LicenseURLToReplace:  regexp.MustCompile(`\bhttps://golang\.org/LICENSE\b`),
	LicenseNameToReplace: regexp.MustCompile(`\bBSD license\b`),
}

// FilterFileContent modifies godoc output to point at github and github pages
// for in package content, and golang.org for anything external to that.
func FilterFileContent(contents string) string {
	if http.DetectContentType([]byte(contents))[0:5] != "text/" {
		// Only modify text files:
		return contents
	}
	contents = Patterns.HostnameToReplace.ReplaceAllString(
		contents,
		`https://golang.org/`)
	contents = Patterns.LibPathToReplace.ReplaceAllString(
		contents,
		GHPagesURL+"/"+DirToFilter+"/lib/")
	contents = Patterns.PkgPathToReplace.ReplaceAllString(
		contents,
		GHPagesURL+"/"+DirToFilter+"/pkg/"+PackageURI+"/")
	contents = Patterns.SrcPathToReplace.ReplaceAllString(
		contents,
		PackageSourceCodeBrowsableRoot+"/")
	contents = Patterns.SrcLineNumToReplace.ReplaceAllStringFunc(
		contents,
		func(inStr string) string {
			outStr := Patterns.SEqualsToRemove.ReplaceAllString(inStr, "#L")
			outStr = Patterns.LineNoToModify.ReplaceAllStringFunc(
				outStr,
				func(lineNumStr string) string {
					line, err := strconv.Atoi(lineNumStr)
					if err != nil {
						panic(fmt.Sprintf(
							"This should be impossible: %s",
							err.Error()))
					}
					return fmt.Sprintf("%d", line+10)
				})
			return outStr
		})
	contents = Patterns.LicenseURLToReplace.ReplaceAllString(
		contents,
		PackageSourceCodeBrowsableRoot+"/LICENSE")
	contents = Patterns.LicenseNameToReplace.ReplaceAllString(contents, LicenseName)
	return contents
}
