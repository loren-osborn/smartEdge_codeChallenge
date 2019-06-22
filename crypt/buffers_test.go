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
		Desc          string
		PEMData       string
		ExpectedKey   string
		ExpectedError *testtools.ErrorSpec
	}{
		{
			Desc: "Invalid input",
			PEMData: "-----BEGIN FOO PUBLIC KEY-----\n" +
				"YWJjMTIz\n" +
				"-----END FOO PUBLIC KEY-----\n",
			ExpectedKey: "<nil>",
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
			ExpectedError: nil,
		},
	} {
		t.Run(fmt.Sprintf("Subtest %d: %s", i+1, tc.Desc), func(tt *testing.T) {
			pemBuffer := crypt.NewPEMBufferFromString(tc.PEMData)
			if pemBuffer.String() != tc.PEMData {
				tt.Errorf("PEM buffer:\n%#v\ndidn't match expected:\n%#v", pemBuffer.String(), tc.PEMData)
			}
			x509Buf, err := pemBuffer.DecodeToX509()
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