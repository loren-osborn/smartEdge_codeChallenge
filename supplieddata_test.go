package codechallenge_test

import (
	"crypto/x509"
	"fmt"
	"github.com/smartedge/codechallenge/crypt"
	"testing"
)

// TestValidationOfSampleData tests (hopefully vaild) sample data
// against our validation routine.
func TestValidationOfSampleData(t *testing.T) {
	for i, tc := range []struct {
		Algorithm x509.PublicKeyAlgorithm
		message   string
		signature string
		publicKey string
	}{
		// From docker example main.go
		{
			Algorithm: x509.ECDSA,
			message:   "theAnswerIs42",
			signature: "MGUCMCDwlFyVdD620p0hRLtABoJTR7UNgwj8g2r0ipNbWPi4Us57YfxtSQJ3dAkHslyBbwIxAKorQmpWl9QdlBUtACcZm4kEXfL37lJ+gZ/hANcTyuiTgmwcEC0FvEXY35u2bKFwhA==",
			publicKey: "-----BEGIN PUBLIC KEY-----\nMHYwEAYHKoZIzj0CAQYFK4EEACIDYgAEI5/0zKsIzou9hL3ZdjkvBeVZFKpDwxTb\nfiDVjHpJdu3+qOuaKYgsLLiO9TFfupMYHLa20IqgbJSIv/wjxANH68aewV1q2Wn6\nvLA3yg2mOTa/OHAZEiEf7bVEbnAov+6D\n-----END PUBLIC KEY-----\n",
		},
		// Example from project spec page
		{
			Algorithm: x509.ECDSA,
			message:   "your@email.com",
			signature: "MGUCMGrxqpS689zQEi5yoBElG41u6U7eKX7ZzaXmXr0C5HgNXlJbiiVQYUS0ZOBxsLU4UgIxAL9AAgkRBUQ7/3EKQag4MjRflAxbfpbGmxb6ar9d4bGZ8FDQkUe6cnCIRleaxFnu2A==",
			publicKey: "-----BEGIN PUBLIC KEY-----\nMHYwEAYHKoZIzj0CAQYFK4EEACIDYgAEDUlT2XxqQAR3PBjeL2D8pQJdghFyBXWI\n/7RvD8Tsdv1YVFwqkJNEC3lNS4Gp7a19JfcrI/8fabLI+yPZBPZjtvuwRoauvGC6\nwdBrL2nzrZxZL4ZsUVNbWnG4SmqQ1f2k\n-----END PUBLIC KEY-----\n",
		},
	} {
		t.Run(fmt.Sprintf("Subtest %d", i+1), func(tt *testing.T) {
			settings := crypt.PkiSettings{
				Algorithm:      tc.Algorithm,
				RSAKeyBits:     2048,
				PrivateKeyPath: "",
				PublicKeyPath:  "",
			}
			tooling, err := crypt.GetCryptoTooling(nil, &settings)
			if err != nil {
				tt.Errorf("Unexpected error calling codechallenge.GetCryptoTooling(): %s", err.Error())
			}
			matched, err := tooling.VerifySignedMessage(tc.message, tc.signature, tc.publicKey)
			if err != nil {
				tt.Errorf("Unexpected error calling codechallenge.(PEMEncoded).DecodeToX509(): %s", err.Error())
			}
			if !matched {
				tt.Errorf("Code signature invalid for message: %#v", tc.message)
			}
		})
	}
}
