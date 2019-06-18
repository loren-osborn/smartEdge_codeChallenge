// Package codechallenge implements a tool to sign a short text message,
// creating a key-pair if necessary
package codechallenge

import (
	"errors"
	"flag"
	"fmt"
	"github.com/smartedge/codechallenge/crypt"
	"github.com/smartedge/codechallenge/deps"
	"github.com/smartedge/codechallenge/misc"
	"io"
	"io/ioutil"
	"strings"
	"unicode"
	"unicode/utf8"
)

// HandleError displays an error message with Usage information to Stderr,
// and exits with an error code.
func HandleError(d *deps.Dependencies, err error, exitStatus int) {
	fmt.Fprintln(d.Os.Stderr, err.Error())
	flag.CommandLine.Usage()
	d.Os.Exit(exitStatus)
}

// RealMain is the program entry-point with all dependencies injected. This
// allows us to test respecting public vs private methods by moving it outside
// the "main" package.
func RealMain(d *deps.Dependencies) {
	config, err := ParseArgs(d)
	if err != nil {
		HandleError(d, err, 1)
	}
	if config.HelpMode {
		flag.CommandLine.SetOutput(d.Os.Stdout)
		flag.CommandLine.Usage()
		d.Os.Exit(0)
	}
	message, err := InjestMessage(d.Os.Stdin, config.Format)
	if err != nil {
		HandleError(d, err, 2)
	}
	cryptStuff, err := crypt.GetCryptoTooling(d, &config.PubKeySettings)
	if err != nil {
		HandleError(d, err, 3)
	}
	err = cryptStuff.GetKeys()
	if err != nil {
		HandleError(d, err, 4)
	}
	binSig, err := cryptStuff.SignMessage(message)
	if err != nil {
		HandleError(d, err, 5)
	}
	// Verify with a round trip:
	valid, err := cryptStuff.VerifySignedMessage(message, binSig.Base64(), cryptStuff.PubKey.String())
	if err != nil {
		HandleError(d, err, 6)
	}
	if !valid {
		HandleError(d, errors.New("round trip verification of signature failed"), 7)
	}
	err = GenerateResponse(d, message, binSig, cryptStuff.PubKey)
	if err != nil {
		HandleError(d, err, 8)
	}
}

// InjestMessage reads all data from dataSource, removing any trailing
// whitespace. Returns an error if the content is longer than 250 characters.
// Input is allowed to be ASCII, Binary or UTF-8: ASCII and Binary data have a
// 250 byte limit, while UTF-8 has a 250 character limit with up to 4 bytes per
// character. ASCII and UTF-8 inputs are both trimmed of trailing whitespace.
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
			return "", fmt.Errorf("Input contains more than 250 bytes (exactly %d):\n%#v", len(msg), msg)
		}
		return msg, nil
	case UTF8:
		if !utf8.ValidString(msg) {
			return "", fmt.Errorf("Input contains invalid UTF-8 character(s):\n%#v", msg)
		}
		msg = misc.TrimRightUTF8Func(msg, unicode.IsSpace)
		charCount := utf8.RuneCountInString(msg)
		if charCount > 250 {
			return "", fmt.Errorf("Input contains more than 250 UTF-8 characters:\n%#v", msg)
		}
		return msg, nil
	}
	return "", fmt.Errorf("INTERNAL ERROR: Unrecognized content format: %#v", format)
}
