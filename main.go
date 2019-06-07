// Package codechallenge implements a tool to sign a short text message,
// creating a key-pair if necessary
package codechallenge

import (
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"unicode"
	"unicode/utf8"
)

// DummyRealMain is the example program supplied as an example for this project
// of working with docker... It will be removed.
func DummyRealMain(d *Dependencies) {
	fmt.Fprintln(d.Os.Stdout, `{
    "message":"theAnswerIs42",
    "signature":"MGUCMCDwlFyVdD620p0hRLtABoJTR7UNgwj8g2r0ipNbWPi4Us57YfxtSQJ3dAkHslyBbwIxAKorQmpWl9QdlBUtACcZm4kEXfL37lJ+gZ/hANcTyuiTgmwcEC0FvEXY35u2bKFwhA==",
    "pubkey":"-----BEGIN PUBLIC KEY-----\nMHYwEAYHKoZIzj0CAQYFK4EEACIDYgAEI5/0zKsIzou9hL3ZdjkvBeVZFKpDwxTb\nfiDVjHpJdu3+qOuaKYgsLLiO9TFfupMYHLa20IqgbJSIv/wjxANH68aewV1q2Wn6\nvLA3yg2mOTa/OHAZEiEf7bVEbnAov+6D\n-----END PUBLIC KEY-----\n"
}`)
}

// RealMain is the program entry-point with all dependencies injected. This
// allows us to test respecting public vs private methods by moving it outside
// the "main" package.
func RealMain(d *Dependencies) {
	fmt.Fprintln(d.Os.Stdout, `{
    "message":"theAnswerIs42",
    "signature":"MGUCMCDwlFyVdD620p0hRLtABoJTR7UNgwj8g2r0ipNbWPi4Us57YfxtSQJ3dAkHslyBbwIxAKorQmpWl9QdlBUtACcZm4kEXfL37lJ+gZ/hANcTyuiTgmwcEC0FvEXY35u2bKFwhA==",
    "pubkey":"-----BEGIN PUBLIC KEY-----\nMHYwEAYHKoZIzj0CAQYFK4EEACIDYgAEI5/0zKsIzou9hL3ZdjkvBeVZFKpDwxTb\nfiDVjHpJdu3+qOuaKYgsLLiO9TFfupMYHLa20IqgbJSIv/wjxANH68aewV1q2Wn6\nvLA3yg2mOTa/OHAZEiEf7bVEbnAov+6D\n-----END PUBLIC KEY-----\n"
}`)
}

// InjestMessage reads all data from dataSource, removing any trailing
// whitespace. Returns an error if the content is longer than 250 characters.
// Input is assumed to be UTF-8. Invalid UTF-8 will also produce an error.
func InjestMessage(dataSource io.Reader, format ContentFormat) (string, error) {
	buff, err := ioutil.ReadAll(dataSource)
	if err != nil {
		return "", err
	}
	msg := string(buff)
	// format is meaningless for an empty string
	if msg == "" {
		return "", nil
	}
	switch format {
	case ASCII, Binary:
		// ASCII is technically only bytes < 127, but related character sets
		// use bytes > 128, so the only difference between ASCII and Binary
		// is the trimming of trailing of trailing whitespace:
		if format == ASCII {
			msg = strings.TrimRightFunc(msg, unicode.IsSpace)
		}
		if len(msg) > 250 {
			return "", fmt.Errorf("Input contains more than 250 bytes:\n%#v", msg)
		}
		return msg, nil
	case UTF8:
		if !utf8.ValidString(msg) {
			return "", fmt.Errorf("Input contains invalid UTF-8 character(s):\n%#v", msg)
		}
		// Because UTF-8 isn't one byte per character, we can't use strings.TrimRightFunc():
		for rChar, rCharLen := utf8.DecodeLastRuneInString(msg); (len(msg) >= rCharLen) && unicode.IsSpace(rChar); {
			msg = msg[:len(msg)-rCharLen]
		}
		charCount := utf8.RuneCountInString(msg)
		if charCount > 250 {
			return "", fmt.Errorf("Input contains more than 250 UTF-8 characters:\n%#v", msg)
		}
		return msg, nil
	}
	return "", fmt.Errorf("INTERNAL ERROR: Unrecognized content format: %#v", format)
}
