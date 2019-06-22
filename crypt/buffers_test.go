package crypt_test

import (
	"fmt"
	"github.com/smartedge/codechallenge/crypt"
	"strings"
	"testing"
)

// TestKeyTypeString tests how KeyTypes render to strings.
func TestKeyTypeString(t *testing.T) {
	for desc, tc := range map[string]struct {
		Value    int
		Expected string
	}{
		"public": {
			Value:    0,
			Expected: "public",
		},
		"private": {
			Value:    1,
			Expected: "private",
		},
		"invalid": {
			Value:    2,
			Expected: "Unknown KeyType 2 (INTERNAL ERROR)",
		},
	} {
		t.Run(fmt.Sprintf("Subtest: %s", desc), func(tt *testing.T) {
			typeValue := crypt.KeyType(tc.Value)
			actual := typeValue.String()
			if tc.Expected != actual {
				tt.Errorf("Unexpected value for crypt.KeyType(%d).String(): Actual: %#v Expected: %#v", tc.Value, actual, tc.Expected)
			}
		})
	}
}

// TestX509EncodedEncodeToPEM tests how x509 buffers are serialized to PEMs.
func TestX509EncodedEncodeToPEM(t *testing.T) {
	for i, tc := range []struct {
		Data          string
		AlgorithmName string
		Type          crypt.KeyType
		Expected      string
	}{
		{
			Data:          "abc123",
			AlgorithmName: "Foo",
			Type:          crypt.PublicKey,
			Expected:      "-----BEGIN FOO PUBLIC KEY-----\nYWJjMTIz\n-----END FOO PUBLIC KEY-----\n",
		},
		{
			Data:          "XYZ9876543210",
			AlgorithmName: "bar",
			Type:          crypt.PublicKey,
			Expected:      "-----BEGIN BAR PUBLIC KEY-----\nWFlaOTg3NjU0MzIxMA==\n-----END BAR PUBLIC KEY-----\n",
		},
	} {
		t.Run(fmt.Sprintf("Subtest: %d", i+1), func(tt *testing.T) {
			x509Buffer := crypt.X509Encoded([]byte(tc.Data))
			actual := x509Buffer.EncodeToPEM(tc.AlgorithmName, tc.Type)
			if tc.Expected != string(actual) {
				tt.Errorf("Unexpected value for crypt.X509Encoded([]byte(%#v)).EncodeToPEM(%#v, crypt.%sKey): Actual: %#v Expected: %#v", tc.Data, tc.AlgorithmName, strings.Title(tc.Type.String()), string([]byte(actual)), tc.Expected)
			}
		})
	}
}
