package main

import (
	"fmt"
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

// main traverses filesystem, transforming all files encountered.
func main() {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error fetching CWD: %v\n", err)
		os.Exit(1)
	}
	rootPath := filepath.Join(cwd, DirToFilter)
	err = filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failure accessing a path %q: %v\n", path, err)
			return err
		}
		if info.IsDir() {
			return nil
		}
		origBuf, err := ioutil.ReadFile(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading file %q: %v\n", path, err)
			return err
		}
		origStr := string(origBuf)
		filteredStr := FilterFileContent(origStr)
		if filteredStr != origStr {
			err = ioutil.WriteFile(path, []byte(filteredStr), info.Mode())
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error writing file %q: %v\n", path, err)
				return err
			}
			fmt.Printf("modified file: %q\n", path)
		}
		return nil
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error walking the path %q: %v\n", rootPath, err)
		os.Exit(2)
	}
}

// FilterFileContent modifies godoc output to point at github and github pages
// for in package content, and golang.org for anything external to that.
func FilterFileContent(contents string) string {
	if http.DetectContentType([]byte(contents))[0:4] != "text/" {
		// Only modify text files:
		return contents
	}
	hostnameToReplace := regexp.MustCompile(`\bhttp://localhost:6060/`)
	contents = hostnameToReplace.ReplaceAllString(contents, `https://golang.org/`)
	libPathToReplace := regexp.MustCompile(`\bhttps://golang\.org/lib/`)
	contents = libPathToReplace.ReplaceAllString(contents, GHPagesURL+"/"+DirToFilter+"/lib/")
	pkgPathToReplace := regexp.MustCompile(fmt.Sprintf(`\bhttps://golang\.org/pkg/%s/`, regexp.QuoteMeta(PackageURI)))
	contents = pkgPathToReplace.ReplaceAllString(contents, GHPagesURL+"/"+DirToFilter+"/pkg/"+PackageURI+"/")
	srcPathToReplace := regexp.MustCompile(fmt.Sprintf(`\bhttps://golang\.org/src/%s/`, regexp.QuoteMeta(PackageURI)))
	contents = srcPathToReplace.ReplaceAllString(contents, PackageSourceCodeBrowsableRoot+"/")
	srcLineNumToReplace := regexp.MustCompile(fmt.Sprintf(`\b%s/[^?]+\?s=\d+:\d+#L\d+\b`, regexp.QuoteMeta(PackageSourceCodeBrowsableRoot)))
	sEqualsToRemove := regexp.MustCompile(`\?s=\d+:\d+#L`)
	lineNoToModify := regexp.MustCompile(`\d+$`)
	contents = srcLineNumToReplace.ReplaceAllStringFunc(
		contents,
		func(inStr string) string {
			outStr := sEqualsToRemove.ReplaceAllString(inStr, "#L")
			outStr = lineNoToModify.ReplaceAllStringFunc(
				outStr,
				func(lineNumStr string) string {
					line, err := strconv.Atoi(lineNumStr)
					if err != nil {
						panic(fmt.Sprintf("This should be impossible: %s", err.Error()))
					}
					return fmt.Sprintf("%d", line+10)
				})
			return outStr
		})
	return contents
}
