package crypt_test

import (
	"fmt"
	"github.com/smartedge/codechallenge/crypt"
	"github.com/smartedge/codechallenge/testtools"
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
			Expected: "-----BEGIN FOO PUBLIC KEY-----\n" +
				"YWJjMTIz\n" +
				"-----END FOO PUBLIC KEY-----\n",
		},
		{
			Data:          "XYZ9876543210",
			AlgorithmName: "bar",
			Type:          crypt.PublicKey,
			Expected: "-----BEGIN BAR PUBLIC KEY-----\n" +
				"WFlaOTg3NjU0MzIxMA==\n" +
				"-----END BAR PUBLIC KEY-----\n",
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

// TestX509EncodedAsGenericPublicKey tests how x509 buffers are decoded into public keys.
func TestX509EncodedAsGenericPublicKey(t *testing.T) {
	for i, tc := range []struct {
		Desc                 string
		PEMData              string
		ExpectedKey          string
		ExpectPEMDecodeError bool
		ExpectedError        *testtools.ErrorSpec
	}{
		{
			Desc:                 "Invalid PEM input",
			PEMData:              "abc123\n",
			ExpectedKey:          "", // n/a
			ExpectPEMDecodeError: true,
			ExpectedError: &testtools.ErrorSpec{
				Type:    "*errors.errorString",
				Message: "No PEM data was found",
			},
		},
		{
			Desc: "Invalid x509 input",
			PEMData: "-----BEGIN FOO PUBLIC KEY-----\n" +
				"YWJjMTIz\n" +
				"-----END FOO PUBLIC KEY-----\n",
			ExpectedKey:          "<nil>",
			ExpectPEMDecodeError: false,
			ExpectedError: &testtools.ErrorSpec{
				Type: "asn1.StructuralError",
				Message: "asn1: structure error: tags don't match " +
					"(16 vs {class:1 tag:1 length:98 isCompound:true}) " +
					"{" +
					"optional:false " +
					"explicit:false " +
					"application:false " +
					"defaultValue:<nil> " +
					"tag:<nil> " +
					"stringType:0 " +
					"timeType:0 " +
					"set:false " +
					"omitEmpty:false" +
					"} publicKeyInfo @2",
			},
		},
		{
			Desc: "Valid input",
			PEMData: "-----BEGIN RSA PUBLIC KEY-----\n" +
				"MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA8QfemZPYmChA2Rnm2pja\n" +
				"JuxjpzWa16RAgV8mhNiAMyGRIvMQ1ec7zgL8j9eCrUJb+RovVkN/ANmM9qBZ4SKC\n" +
				"K0rxIQHBQHKzNTLPPas3PHw47F2HsW3I6XolvqAMWJoXFk9/o9U0qk8zkXWkv3pM\n" +
				"sdBuod++FKI11qabXIobIbR40kFWdF2TpKnwLGjtX2ade8/TUFUv/PQ/YBnVXTAw\n" +
				"ilYsvvTG1JijkoYNyeLrUhcdibE9XAWdU1NicDxU/x6CGb/ALo/WlW+aozsB7OAH\n" +
				"yINoMQ5wPx5zblsh2SUUn2fbcLVu3fLAv6FnaxAiaFo3wK/dbLy2EUduEwdHFyFQ\n" +
				"IQIDAQAB\n" +
				"-----END RSA PUBLIC KEY-----\n",
			ExpectedKey: "&rsa.PublicKey{N:30427312107142716373358747838974387706" +
				"2164806841138110090410812546353424602164777241125504339706669063" +
				"8964347194528922643654238848939468225578324074329107430962098774" +
				"1773388478658307728442616725919594201188330984901975733761247876" +
				"2859024847778100036251828251780950901516886026755599828907940034" +
				"2595708949293461365405393347671925097562881388810550179560551162" +
				"2803696924829582307949181899636200794896203740456253783344451197" +
				"4127651164398730320586847027974392966884965912021628848046743184" +
				"1766701590200578211474964612687407014185219512140269663909935873" +
				"8279934060503875651084074546544931657240541911067692979624712687" +
				"649, E:65537}",
			ExpectPEMDecodeError: false,
			ExpectedError:        nil,
		},
	} {
		t.Run(fmt.Sprintf("Subtest %d: %s", i+1, tc.Desc), func(tt *testing.T) {
			pemBuffer := crypt.NewPEMBufferFromString(tc.PEMData)
			if pemBuffer.String() != tc.PEMData {
				tt.Errorf("PEM buffer:\n%#v\ndidn't match expected:\n%#v", pemBuffer.String(), tc.PEMData)
			}
			x509Buf, err := pemBuffer.DecodeToX509()
			if tc.ExpectPEMDecodeError {
				if tc.ExpectedError == nil {
					tt.Error("Invalid test, ExpectedError must be non-nil when ExpectPEMDecodeError is true")
				} else if matchErr := tc.ExpectedError.EnsureMatches(err); matchErr != nil {
					tt.Error(matchErr.Error())
				}
				return
			}
			if err != nil {
				tt.Errorf("Error decoding valid PEM buffer: %T %s", err, err.Error())
				return
			}
			actualKey, actualErr := x509Buf.AsGenericPublicKey()
			if actualKeyAsString := fmt.Sprintf("%#v", actualKey); tc.ExpectedKey != actualKeyAsString {
				tt.Errorf("Expected %s to match expected %s", actualKeyAsString, tc.ExpectedKey)
			}
			if err := tc.ExpectedError.EnsureMatches(actualErr); err != nil {
				tt.Error(err.Error())
			}
			if (actualErr != nil) == (actualKey != nil) {
				tt.Errorf("Exactly one of key and err should be nil: key: %#v; error: %T %s", actualKey, actualErr, actualErr.Error())
			}
		})
	}
}

// TestDigestHash tests how the hash buffer output of SHA256 behaves.
func TestDigestHash(t *testing.T) {
	for i, tc := range []struct {
		Desc              string
		TestData          string
		ExpectedSHA256    string
		ExpectedSHA256Hex string
	}{
		{
			Desc:              "Data to hash",
			TestData:          "abc123",
			ExpectedSHA256:    "\x6c\xa1\x3d\x52\xca\x70\xc8\x83\xe0\xf0\xbb\x10\x1e\x42\x5a\x89\xe8\x62\x4d\xe5\x1d\xb2\xd2\x39\x25\x93\xaf\x6a\x84\x11\x80\x90",
			ExpectedSHA256Hex: "6ca13d52ca70c883e0f0bb101e425a89e8624de51db2d2392593af6a84118090",
		},
		{
			Desc:              "Other data to hash",
			TestData:          "456def",
			ExpectedSHA256:    "\x5e\x68\x9f\xe2\xea\xed\x09\x7f\x8a\x9d\xc7\x4e\x6d\x2a\xaa\xef\x59\x68\x8d\x82\x0e\xcb\x4c\xef\xc7\xb1\xa6\x5c\xc7\x8f\xa0\xa2",
			ExpectedSHA256Hex: "5e689fe2eaed097f8a9dc74e6d2aaaef59688d820ecb4cefc7b1a65cc78fa0a2",
		},
	} {
		t.Run(fmt.Sprintf("Subtest %d: %s", i+1, tc.Desc), func(tt *testing.T) {
			digestHash := crypt.NewSHA256DigestHash(tc.TestData)
			if string([]byte(digestHash)) != tc.ExpectedSHA256 {
				tt.Errorf("Test data:\n"+
					"%#v\n"+
					"does not have expected SHA256 string:\n"+
					"%#v\n"+
					"instead got string:\n"+
					"%#v",
					tc.TestData,
					tc.ExpectedSHA256,
					string([]byte(digestHash)))
			}
			if digestHash.Hex() != tc.ExpectedSHA256Hex {
				tt.Errorf("Test data:\n"+
					"%#v\n"+
					"does not have expected Hexidecimal string:\n"+
					"%#v\n"+
					"instead got string:\n"+
					"%#v",
					tc.TestData,
					tc.ExpectedSHA256Hex,
					digestHash.Hex())
			}
		})
	}
}

// TestBinarySignature tests how the binary signature buffer behaves.
func TestBinarySignature(t *testing.T) {
	for i, tc := range []struct {
		Desc          string
		Base64Input   string
		ExpectedError *testtools.ErrorSpec
		ExpectedData  string
	}{
		{
			Desc:        "bad data",
			Base64Input: "!@#$^&*()Punctuation!@#$^&*()",
			ExpectedError: &testtools.ErrorSpec{
				Type:    "base64.CorruptInputError",
				Message: "illegal base64 data at input byte 0",
			},
			ExpectedData: "",
		},
		{
			Desc:          "Exactly 6 bytes",
			Base64Input:   "45/7de+g",
			ExpectedError: nil,
			ExpectedData:  "\xe3\x9f\xfb\x75\xef\xa0",
		},
		{
			Desc:        "invalid padding",
			Base64Input: "45/7de+g=",
			ExpectedError: &testtools.ErrorSpec{
				Type:    "base64.CorruptInputError",
				Message: "illegal base64 data at input byte 8",
			},
			ExpectedData: "",
		},
		{
			Desc:          "data with padding",
			Base64Input:   "456def8=",
			ExpectedError: nil,
			ExpectedData:  "\xe3\x9e\x9d\x79\xff",
		},
	} {
		t.Run(fmt.Sprintf("Subtest %d: %s", i+1, tc.Desc), func(tt *testing.T) {
			binSig, actualErr := crypt.NewBinarySignatureFromBase64(tc.Base64Input)
			if err := tc.ExpectedError.EnsureMatches(actualErr); err != nil {
				tt.Error(err.Error())
			}
			if tc.ExpectedError != nil {
				if binSig != nil {
					tt.Errorf("If there is an error, the returned buffer "+
						"should be nil. Instead it's:\n"+
						"%#v",
						[]byte(binSig))
				}
				return
			}
			if string([]byte(binSig)) != tc.ExpectedData {
				tt.Errorf("Test base64 data:\n"+
					"%#v\n"+
					"does not have expected value:\n"+
					"%#v\n"+
					"instead got:\n"+
					"%#v",
					tc.Base64Input,
					[]byte(tc.ExpectedData),
					binSig)
			}
			if binSig.Base64() != tc.Base64Input {
				tt.Errorf("Test binary data:\n"+
					"%#v\n"+
					"does not have base64 value:\n"+
					"%#v\n"+
					"instead got:\n"+
					"%#v",
					[]byte(tc.ExpectedData),
					tc.Base64Input,
					binSig.Base64())
			}
		})
	}
}
