package main

import (
	"fmt"
	"regexp"
	"strings"
	"testing"
)

// DirToFilter is the directory to modify
const (
	CSSOrigWithLibPath = `
.treeview-red li { background-image: url(http://localhost:6060/lib/godoc/images/treeview-red-line.gif); }
`
	CSSTranslatedWithLibPath = `
.treeview-red li { background-image: url(https://loren-osborn.github.io/smartEdge_codeChallenge/godoc/lib/godoc/images/treeview-red-line.gif); }
`
)

type TranslationStringPair struct {
	From string
	To   string
}

func GetRepetitions(from, to string) map[string]TranslationStringPair {
	return map[string]TranslationStringPair{
		"Empty string": {
			From: "",
			To:   "",
		},
		"%s Once": {
			From: from,
			To:   to,
		},
		"%s 5 times": {
			From: strings.Repeat(from, 5),
			To:   strings.Repeat(to, 5),
		},
	}
}

// TestConvertLibPath tests proper translation of a localhost:6060/lib/ path.
func TestConvertLibPath(t *testing.T) {
	// t.Skip("Skiping currently known-broken tests.")
	for caseName, tc := range GetRepetitions(CSSOrigWithLibPath, CSSTranslatedWithLibPath) {
		t.Run(fmt.Sprintf("Testing %s", fmt.Sprintf(caseName, "/lib/ path")), func(tt *testing.T) {
			actual := FilterFileContent(tc.From)
			if tc.To != actual {
				tt.Errorf("File content:\n\t%#v\nnot translated properly. Expected:\n\t%#v\nbut instead got:\n\t%#v", tc.From, tc.To, actual)
			}
		})
	}
}

// TestRegexps tests proper translation of a localhost:6060/lib/ path.
func TestRegexps(t *testing.T) {
	for _, tc := range []struct {
		CaseName    string
		Pattern     *regexp.Regexp
		Replacement string
		From        string
		To          string
	}{
		{
			CaseName:    "replace hostname",
			Pattern:     Patterns.HostnameToReplace,
			Replacement: "MY NEW HOST",
			From:        "This is a http://localhost:6060/ test",
			To:          "This is a MY NEW HOST test",
		},
		{
			CaseName:    "real-world replace hostname",
			Pattern:     Patterns.HostnameToReplace,
			Replacement: `https://golang.org/`,
			From:        "background-image: url(http://localhost:6060/lib/godoc/images/treeview-red-line.gif); ",
			To:          "background-image: url(https://golang.org/lib/godoc/images/treeview-red-line.gif); ",
		},
	} {
		for repeatTestCaseName, rtc := range GetRepetitions(tc.From, tc.To) {
			t.Run(fmt.Sprintf(repeatTestCaseName, tc.CaseName), func(tt *testing.T) {
				actual := Patterns.HostnameToReplace.ReplaceAllString(rtc.From, tc.Replacement)
				if rtc.To != actual {
					tt.Errorf("File content:\n\t%#v\nnot translated properly. Expected:\n\t%#v\nbut instead got:\n\t%#v", rtc.From, rtc.To, actual)
				}
			})
		}
	}
}
