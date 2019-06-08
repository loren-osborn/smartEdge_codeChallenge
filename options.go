package codechallenge

import (
	"crypto/x509"
	"errors"
	"flag"
	"fmt"
	"path/filepath"
	"strings"
)

// ContentFormat the data format of the message to be signed
type ContentFormat int

// Content character encodings
const (
	UTF8 ContentFormat = iota
	ASCII
	Binary
)

// PkiSettings are the public key settings as specified on the command line.
type PkiSettings struct {
	Algorithm      x509.PublicKeyAlgorithm
	RSAKeyBits     int
	PrivateKeyPath string
	PublicKeyPath  string
}

// RunConfig program's running config as specified on the command line.
type RunConfig struct {
	Format         ContentFormat
	PubKeySettings PkiSettings
}

// ParseArgs parses the runtime configuration from the command line arguemnts.
func ParseArgs(d *Dependencies) (*RunConfig, error) {
	defaultKeyDir := filepath.Join(d.Os.Getenv("HOME"), ".smartEdge")
	flag.CommandLine.SetOutput(d.Os.Stderr)
	result := RunConfig{
		Format: UTF8, // default
		PubKeySettings: PkiSettings{
			Algorithm:      x509.ECDSA, // default
			RSAKeyBits:     2048,       //default
			PrivateKeyPath: filepath.Join(defaultKeyDir, "id_{{algorithm}}.priv"),
			PublicKeyPath:  filepath.Join(defaultKeyDir, "id_{{algorithm}}.pub"),
		},
	}
	type namedFlagValPair struct {
		name    string
		present *bool
	}
	algorithmFlags := map[x509.PublicKeyAlgorithm]namedFlagValPair{
		x509.RSA: {
			name:    "rsa",
			present: flag.Bool("rsa", false, "Causes the mesage to be signed with an RSA key-pair"),
		},
		x509.ECDSA: {
			name:    "ecdsa",
			present: flag.Bool("ecdsa", false, "Causes the mesage to be signed with an ECDSA key-pair [default]"),
		},
	}
	formatFlags := map[ContentFormat]namedFlagValPair{
		UTF8: {
			name:    "utf8",
			present: flag.Bool("utf8", false, "This specifies that the message is UTF-8 content [default]"),
		},
		ASCII: {
			name:    "ascii",
			present: flag.Bool("ascii", false, "This specifies that the message is ASCII content"),
		},
		Binary: {
			name:    "binary",
			present: flag.Bool("binary", false, "This specifies that the message is raw binary content"),
		},
	}
	overridePrivateKeyPath := flag.String("private", "", "filepath of the private key file. Defaults to ~/.smartEdge/id_rsa.priv for RSA and ~/.smartEdge/id_ecdsa.priv for ECDSA.")
	overridePublicKeyPath := flag.String("public", "", "filepath of the private key file. Defaults to ~/.smartEdge/id_rsa.pub for RSA and ~/.smartEdge/id_ecdsa.pub for ECDSA.")
	rsaKeyBits := flag.Uint("bits", 0, "Bit length of the RSA key [default=2048]")
	if err := flag.CommandLine.Parse(d.Os.Args[1:]); err != nil {
		return nil, err
	}
	mutuallyExclusiveFlagCount := 0
	lastNamedOption := ""
	for val, flagPair := range algorithmFlags {
		if *(flagPair.present) {
			if mutuallyExclusiveFlagCount > 0 {
				return nil, fmt.Errorf("Options -%s and -%s may not be used together", lastNamedOption, flagPair.name)
			}
			mutuallyExclusiveFlagCount++
			lastNamedOption = flagPair.name
			result.PubKeySettings.Algorithm = val
		}
	}
	mutuallyExclusiveFlagCount = 0
	lastNamedOption = ""
	for val, flagPair := range formatFlags {
		if *(flagPair.present) {
			if mutuallyExclusiveFlagCount > 0 {
				return nil, fmt.Errorf("Options -%s and -%s may not be used together", lastNamedOption, flagPair.name)
			}
			mutuallyExclusiveFlagCount++
			lastNamedOption = flagPair.name
			result.Format = val
		}
	}
	// we only want to replace the "{{algorithm}}" token in the defaults, not in
	// the command arguments.
	result.PubKeySettings.PrivateKeyPath = strings.Replace(
		result.PubKeySettings.PrivateKeyPath,
		"{{algorithm}}",
		algorithmFlags[result.PubKeySettings.Algorithm].name,
		0)
	result.PubKeySettings.PublicKeyPath = strings.Replace(
		result.PubKeySettings.PublicKeyPath,
		"{{algorithm}}",
		algorithmFlags[result.PubKeySettings.Algorithm].name,
		0)
	if result.PubKeySettings.Algorithm != x509.RSA {
		if *rsaKeyBits != 0 {
			return nil, errors.New("Options -bits is only valid for RSA")
		}
	} else {
		if *rsaKeyBits < 256 {
			// 2048 is the least currently considered "secure through 2030."
			// 256 bits is 2.791 * 10^539 times weaker than that.
			return nil, fmt.Errorf("Options -bits less than 256 not allowed. Saw -bits=%d", *rsaKeyBits)
		}
		result.PubKeySettings.RSAKeyBits = int(*rsaKeyBits)
	}

	// Replace if we don't see the default value of empty string
	if *overridePrivateKeyPath != "" {
		result.PubKeySettings.PrivateKeyPath = *overridePrivateKeyPath
	}
	if *overridePublicKeyPath != "" {
		result.PubKeySettings.PublicKeyPath = *overridePublicKeyPath
	}
	return &result, nil
}
