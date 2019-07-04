package crypt_test

import (
	"crypto/x509"
	"errors"
	"fmt"
	"github.com/smartedge/codechallenge/crypt"
	"github.com/smartedge/codechallenge/testtools"
	"github.com/smartedge/codechallenge/testtools/mocks"
	"os"
	"path/filepath"
	"testing"
)

// TestGetCryptoTooling tests GetCryptoTooling().
func TestGetCryptoTooling(t *testing.T) {
	for _, tc := range []struct {
		desc            string
		settings        *crypt.PkiSettings
		expectedError   *testtools.ErrorSpec
		expectedAlgName string
	}{
		{
			desc:     "Invalid algorithm",
			settings: nil,
			expectedError: &testtools.ErrorSpec{
				Type:    "*errors.errorString",
				Message: "INTERNAL ERROR: No settings provided",
			},
			expectedAlgName: "",
		},
		{
			desc:     "Invalid algorithm",
			settings: &crypt.PkiSettings{},
			expectedError: &testtools.ErrorSpec{
				Type:    "*errors.errorString",
				Message: "INTERNAL ERROR: Unrecognized algorithm: 0",
			},
			expectedAlgName: "",
		},
		{
			desc: "ecdsa",
			settings: &crypt.PkiSettings{
				Algorithm: x509.ECDSA,
			},
			expectedError:   nil,
			expectedAlgName: "ECDSA",
		},
		{
			desc: "rsa",
			settings: &crypt.PkiSettings{
				Algorithm: x509.RSA,
			},
			expectedError:   nil,
			expectedAlgName: "RSA",
		},
	} {
		t.Run(fmt.Sprintf("Subtest: %s", tc.desc), func(tt *testing.T) {
			mockDepsBundle := mocks.NewDefaultMockDeps("", []string{"progname"}, "/home/user", nil)
			returnedNormally := false
			var actualTooling *crypt.CryptoTooling
			var actualErr error
			err := mockDepsBundle.InvokeCallInMockedEnv(func() error {
				actualTooling, actualErr = crypt.GetCryptoTooling(mockDepsBundle.Deps, tc.settings)
				returnedNormally = true
				return nil
			})
			if err != nil {
				tt.Errorf("Unexpected error calling mockDepsBundle.InvokeCallInMockedEnv(): %s", err.Error())
			}
			if exitStatus := mockDepsBundle.GetExitStatus(); (exitStatus != 0) || !returnedNormally {
				tt.Error("EncodeAndSaveKey() should not have paniced or called os.Exit.")
			}
			if (mockDepsBundle.OutBuf.String() != "") || (mockDepsBundle.ErrBuf.String() != "") {
				tt.Errorf("EncodeAndSaveKey() should not have output any data. Saw stdout:\n%s\nstderr:\n%s", mockDepsBundle.OutBuf.String(), mockDepsBundle.ErrBuf.String())
			}
			if !mockDepsBundle.Files.IsEqualTo(testtools.FakeFileSystem{"/home/user": nil}) {
				tt.Errorf("Unexpected change in filesystem state. Expected:\n%s\nActual:\n%s", testtools.FakeFileSystem{"/home/user": nil}.String(), mockDepsBundle.Files.String())
			}
			if err := tc.expectedError.EnsureMatches(actualErr); err != nil {
				tt.Error(err.Error())
			}
			if tc.expectedError != nil {
				if actualTooling != nil {
					tt.Errorf("GetCryptoTooling() should have returned a nil *crypt.CryptoTooling.")
				}
			} else {
				if (actualTooling == nil) || (actualTooling.AlgPlugin == nil) {
					tt.Errorf("GetCryptoTooling() should have returned a *crypt.CryptoTooling.")
				} else {
					if actualTooling.AlgPlugin.GetAlgorithmName() != tc.expectedAlgName {
						tt.Errorf("Expected tooling for algorithm %s, but got tooling for algorithm %s", tc.expectedAlgName, actualTooling.AlgPlugin.GetAlgorithmName())
					}
				}
			}
		})
	}
}

// TestPopulateKeys tests PopulateKeys().
func TestPopulateKeys(t *testing.T) {
	for _, tc := range []struct {
		desc            string
		settings        *crypt.PkiSettings
		fileSystemState testtools.FakeFileSystem
		setup           func(mdb *mocks.MockDepsBundle) error
		expectedError   *testtools.ErrorSpec
	}{
		{
			desc: "ecdsa",
			settings: &crypt.PkiSettings{
				Algorithm:      x509.ECDSA,
				PrivateKeyPath: ".prog/ecdsa_priv.key",
				PublicKeyPath:  ".prog/ecdsa.pub",
			},
			fileSystemState: nil,
			setup: func(mdb *mocks.MockDepsBundle) error {
				return nil
			},
			expectedError: nil,
		},
		{
			desc: "rsa",
			settings: &crypt.PkiSettings{
				Algorithm:      x509.RSA,
				RSAKeyBits:     2048,
				PrivateKeyPath: ".prog/rsa_priv.key",
				PublicKeyPath:  ".prog/rsa.pub",
			},
			fileSystemState: nil,
			setup: func(mdb *mocks.MockDepsBundle) error {
				return nil
			},
			expectedError: nil,
		},
		{
			desc: "existing ecdsa key",
			settings: &crypt.PkiSettings{
				Algorithm:      x509.ECDSA,
				PrivateKeyPath: ".prog/ecdsa_priv.key",
				PublicKeyPath:  ".prog/ecdsa.pub",
			},
			fileSystemState: testtools.FakeFileSystem{
				"/home/user/.prog/ecdsa_priv.key": testtools.StringPtr("-----BEGIN ECDSA PRIVATE KEY-----\n" +
					"MHcCAQEEIHhN6EluNZ8M83qtHKNw/SXxXZgr1C5r0R/L/ND2IUgjoAoGCCqGSM49\n" +
					"AwEHoUQDQgAE7WzVjtn9Gk+WHr5xbv8XMvooqU25BhgNjZ/vHZLBdVtCOjk4KxjS\n" +
					"1UBfQm0c3TRxWBl3hj2AmnJbCrnGofMHBQ==\n" +
					"-----END ECDSA PRIVATE KEY-----\n"),
				"/home/user/.prog/ecdsa.pub": testtools.StringPtr("-----BEGIN ECDSA PUBLIC KEY-----\n" +
					"MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE7WzVjtn9Gk+WHr5xbv8XMvooqU25\n" +
					"BhgNjZ/vHZLBdVtCOjk4KxjS1UBfQm0c3TRxWBl3hj2AmnJbCrnGofMHBQ==\n" +
					"-----END ECDSA PUBLIC KEY-----\n"),
			},
			setup: func(mdb *mocks.MockDepsBundle) error {
				return nil
			},
			expectedError: nil,
		},
		{
			desc: "existing rsa key",
			settings: &crypt.PkiSettings{
				Algorithm:      x509.RSA,
				RSAKeyBits:     2048,
				PrivateKeyPath: ".prog/rsa_priv.key",
				PublicKeyPath:  ".prog/rsa.pub",
			},
			fileSystemState: testtools.FakeFileSystem{
				"/home/user/.prog/rsa_priv.key": testtools.StringPtr("-----BEGIN RSA PRIVATE KEY-----\n" +
					"MIIEpgIBAAKCAQEAzCTTFKQBHfTN8jW6q8PTHNZKWnRPxSt9kpgWmyqFaZnEUipg\n" +
					"oKGAxSIsVrl2PJSm5OlgkVzx+MY+LWM64VKMbRpUUGJR3zdMNhwZQX0hjOpLpVJv\n" +
					"UwD78utVs8vijrU7sH48usFiaZQYjy4m4hQh63/x4h3KVz7YqUnlRMzYJFT43+Aw\n" +
					"YzYuEpzWRxtW7IObJPtjtmYVoqva98fF6aj5uHAsvaAgZGBalHXmCiPzKiGU/hal\n" +
					"zXSPvyJ2Cqz2aUqMHgwi/2Ip4z/mrfX+mUTaS+LyBy7GgqJ5vbkGArMagJIc0eAR\n" +
					"F60r6Uf483xh17oniABdLJy4qlLf6PcEU+utEwIDAQABAoIBAQCXoQp/nEmQHJMT\n" +
					"wsDCcBNQqgJUKWxOwTzmM70mx3CMvT/K39shtJPW2MkiKWMfIDLOeGHX1reL1oO8\n" +
					"ZqYHUq8nIpVZl43ERGiBEGHZ+L2A004Yn6A8gNCi4BWqFFhVM1wAfeNRu+4DCZMs\n" +
					"VlVfOyDusPvSvdna770yEMcQUS6B3J427cDvovA/VFyu97bXpTbeU3Ycsm9PMYQy\n" +
					"mbXHVERtUNuFBUSqiBArJEZMQPo2jCp1Qphcny3RzvTWBIG/lgC2Jv4WfiFtJBNR\n" +
					"EXqmRwrZg6e2PQRqxUZITlEiuDYTKEIfGR5hplZWKaQdI/LTfq5K0LTn0F9jAfaM\n" +
					"0Rb9jgehAoGBAPw/SkojyqhVc1r0OgP3gJdqcc4aMEq1Q4Ea2XhKR2SN9KAQf2/y\n" +
					"PjAFDuXE1Gk+/6NCHLbZl8R1iJ5Cb+jqHeOdKZEtbtvypgo3jG0CuG6awRj+paOx\n" +
					"8rD/KldYdXLrHIgeKw0cr0raouEuSKrANBDXM9sEwCgzBI5Xn0Np1HsZAoGBAM8u\n" +
					"U94UEKg8o2A7jziEmBMI8AdEl2vwvM3Vv1vj4qQalUBY2kBNjDS+8RtMqqpDno6y\n" +
					"6iiDAXU2dDSsQ4h5UTJPjoSqKMdlauqf2Xjkksno/b5D7XREuVQ80u/p8R7qO8GT\n" +
					"WM5dmwpmtop6Af8PRQkKcHn9rWJjttF5Z84h3NsLAoGBAMD4SgTdzLNqa40xORC/\n" +
					"zwgGznk1X6xHbxTdTXDQoj0yu+mXtWYWk6x4siTkpvq8zyQ992mKnKgWoiUv/hzY\n" +
					"vXTbTmlZsG1i+9LlG/BpHF7A1OgiJuVLxLXS/rlDWtZHNtSK/7RQNWm2SNSra7v0\n" +
					"veAEQg9TWw1luh6KubQAyiRBAoGBALcA5AH5VVFV5rYdRgAVV0MRFPxGgT5OMmfa\n" +
					"06H2ZH6yIH3rPjWoih0ZQF3t1Z56BjdkIGPSfFot1G2mcCy/hJJdJbVXnJespMlE\n" +
					"k1MvC94f2OrUk42tGssmwug6i8rT+h6d6ca3dji0y6774IGM2l0HBJ0tD5cmHxlf\n" +
					"FOtGjBBLAoGBALxzAKtXZVGom4zFXazmhD9ZhwctUvwMlVv8DbayJDBBdp2T9uZw\n" +
					"Vln95ZncZjCcIlmlHXKZ363B5q2BksWDEeOMOteyjv/pUF+5SUMNDXAefzakbzib\n" +
					"2M9Ugg1FtqD97FQpvz4Q26DnIZweyF+4Nfp6AfpDj/Dxgtk/kwE6SWbR\n" +
					"-----END RSA PRIVATE KEY-----\n"),
				"/home/user/.prog/rsa.pub": testtools.StringPtr("-----BEGIN RSA PUBLIC KEY-----\n" +
					"MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAzCTTFKQBHfTN8jW6q8PT\n" +
					"HNZKWnRPxSt9kpgWmyqFaZnEUipgoKGAxSIsVrl2PJSm5OlgkVzx+MY+LWM64VKM\n" +
					"bRpUUGJR3zdMNhwZQX0hjOpLpVJvUwD78utVs8vijrU7sH48usFiaZQYjy4m4hQh\n" +
					"63/x4h3KVz7YqUnlRMzYJFT43+AwYzYuEpzWRxtW7IObJPtjtmYVoqva98fF6aj5\n" +
					"uHAsvaAgZGBalHXmCiPzKiGU/halzXSPvyJ2Cqz2aUqMHgwi/2Ip4z/mrfX+mUTa\n" +
					"S+LyBy7GgqJ5vbkGArMagJIc0eARF60r6Uf483xh17oniABdLJy4qlLf6PcEU+ut\n" +
					"EwIDAQAB\n" +
					"-----END RSA PUBLIC KEY-----\n"),
			},
			setup: func(mdb *mocks.MockDepsBundle) error {
				return nil
			},
			expectedError: nil,
		},
		{
			desc: "no public key",
			settings: &crypt.PkiSettings{
				Algorithm:      x509.ECDSA,
				PrivateKeyPath: ".prog/ecdsa_priv.key",
				PublicKeyPath:  ".prog/ecdsa.pub",
			},
			fileSystemState: testtools.FakeFileSystem{
				"/home/user/.prog/ecdsa_priv.key": testtools.StringPtr("nonsense"),
			},
			setup: func(mdb *mocks.MockDepsBundle) error {
				return nil
			},
			expectedError: &testtools.ErrorSpec{
				Type:    "*errors.errorString",
				Message: "Files .prog/ecdsa_priv.key and .prog/ecdsa.pub must either both be present or missing",
			},
		},
		{
			desc: "no private key",
			settings: &crypt.PkiSettings{
				Algorithm:      x509.RSA,
				PrivateKeyPath: "foobar/rsa_priv.key",
				PublicKeyPath:  "foobar/rsa.pub",
			},
			fileSystemState: testtools.FakeFileSystem{
				"/home/user/foobar/rsa.pub": testtools.StringPtr("nonsense"),
			},
			setup: func(mdb *mocks.MockDepsBundle) error {
				return nil
			},
			expectedError: &testtools.ErrorSpec{
				Type:    "*errors.errorString",
				Message: "Files foobar/rsa_priv.key and foobar/rsa.pub must either both be present or missing",
			},
		},
		{
			desc: "ecdsa bad random data",
			settings: &crypt.PkiSettings{
				Algorithm:      x509.ECDSA,
				PrivateKeyPath: ".prog/ecdsa_priv.key",
				PublicKeyPath:  ".prog/ecdsa.pub",
			},
			fileSystemState: nil,
			setup: func(mdb *mocks.MockDepsBundle) error {
				mdb.Deps.Crypto.Rand.Reader = testtools.ReaderFunc(func(p []byte) (n int, err error) {
					return 0, errors.New("Fake I/O Error")
				})
				return nil
			},
			expectedError: &testtools.ErrorSpec{
				Type:    "*errors.errorString",
				Message: "Fake I/O Error",
			},
		},
		{
			desc: "rsa bad random data",
			settings: &crypt.PkiSettings{
				Algorithm:      x509.RSA,
				RSAKeyBits:     2048,
				PrivateKeyPath: ".prog/rsa_priv.key",
				PublicKeyPath:  ".prog/rsa.pub",
			},
			fileSystemState: nil,
			setup: func(mdb *mocks.MockDepsBundle) error {
				mdb.Deps.Crypto.Rand.Reader = testtools.ReaderFunc(func(p []byte) (n int, err error) {
					return 0, errors.New("Fake I/O Error")
				})
				return nil
			},
			expectedError: &testtools.ErrorSpec{
				Type:    "*errors.errorString",
				Message: "Fake I/O Error",
			},
		},
		{
			desc: "bad public key write",
			settings: &crypt.PkiSettings{
				Algorithm:      x509.RSA,
				RSAKeyBits:     2048,
				PrivateKeyPath: ".prog/rsa_priv.key",
				PublicKeyPath:  ".prog/rsa.pub",
			},
			fileSystemState: nil,
			setup: func(mdb *mocks.MockDepsBundle) error {
				mdb.Deps.Io.Ioutil.WriteFile = func(path string, data []byte, perm os.FileMode) error {
					return errors.New("Fake Public key write failure")
				}
				return nil
			},
			expectedError: &testtools.ErrorSpec{
				Type:    "*errors.errorString",
				Message: "Fake Public key write failure",
			},
		},
		{
			desc: "bad private key write",
			settings: &crypt.PkiSettings{
				Algorithm:      x509.RSA,
				RSAKeyBits:     2048,
				PrivateKeyPath: ".prog/rsa_priv.key",
				PublicKeyPath:  ".prog/rsa.pub",
			},
			fileSystemState: nil,
			setup: func(mdb *mocks.MockDepsBundle) error {
				origWriteFile := mdb.Deps.Io.Ioutil.WriteFile
				counter := 0
				mdb.Deps.Io.Ioutil.WriteFile = func(path string, data []byte, perm os.FileMode) error {
					counter++
					if counter < 2 {
						return origWriteFile(path, data, perm)
					}
					return errors.New("Fake Private key write failure")
				}
				return nil
			},
			expectedError: &testtools.ErrorSpec{
				Type:    "*errors.errorString",
				Message: "Fake Private key write failure",
			},
		},
		{
			desc: "bad private key read",
			settings: &crypt.PkiSettings{
				Algorithm:      x509.RSA,
				RSAKeyBits:     2048,
				PrivateKeyPath: ".prog/rsa_priv.key",
				PublicKeyPath:  ".prog/rsa.pub",
			},
			fileSystemState: nil,
			setup: func(mdb *mocks.MockDepsBundle) error {
				mdb.Deps.Io.Ioutil.ReadFile = func(path string) ([]byte, error) {
					return nil, errors.New("Fake Private key read failure")
				}
				return nil
			},
			expectedError: &testtools.ErrorSpec{
				Type:    "*errors.errorString",
				Message: "Fake Private key read failure",
			},
		},
		{
			desc: "private key changed after write",
			settings: &crypt.PkiSettings{
				Algorithm: x509.ECDSA, // Even with constant pseudo-random input,
				// generated RSA keys seem to vary over time.
				PrivateKeyPath: ".prog/ecdsa_priv.key",
				PublicKeyPath:  ".prog/ecdsa.pub",
			},
			fileSystemState: nil,
			setup: func(mdb *mocks.MockDepsBundle) error {
				mdb.Deps.Io.Ioutil.ReadFile = func(path string) ([]byte, error) {
					return []byte("-----BEGIN ECDSA PRIVATE KEY-----\n" +
						"NotWhatWeWrote00\n" +
						"-----END ECDSA PRIVATE KEY-----\n"), nil
				}
				return nil
			},
			expectedError: &testtools.ErrorSpec{
				Type: "*errors.errorString",
				Message: "File .prog/ecdsa_priv.key contents changed between writing and reading: Was:\n" +
					"-----BEGIN ECDSA PRIVATE KEY-----\n" +
					"MHcCAQEEIHhN6EluNZ8M83qtHKNw/SXxXZgr1C5r0R/L/ND2IUgjoAoGCCqGSM49\n" +
					"AwEHoUQDQgAE7WzVjtn9Gk+WHr5xbv8XMvooqU25BhgNjZ/vHZLBdVtCOjk4KxjS\n" +
					"1UBfQm0c3TRxWBl3hj2AmnJbCrnGofMHBQ==\n" +
					"-----END ECDSA PRIVATE KEY-----\n" +
					"\n" +
					"\n" +
					"Now:\n" +
					"-----BEGIN ECDSA PRIVATE KEY-----\n" +
					"NotWhatWeWrote00\n" +
					"-----END ECDSA PRIVATE KEY-----\n",
			},
		},
		{
			desc: "bad public key read",
			settings: &crypt.PkiSettings{
				Algorithm:      x509.RSA,
				RSAKeyBits:     2048,
				PrivateKeyPath: ".prog/rsa_priv.key",
				PublicKeyPath:  ".prog/rsa.pub",
			},
			fileSystemState: nil,
			setup: func(mdb *mocks.MockDepsBundle) error {
				origReadFile := mdb.Deps.Io.Ioutil.ReadFile
				counter := 0
				mdb.Deps.Io.Ioutil.ReadFile = func(path string) ([]byte, error) {
					counter++
					if counter < 2 {
						return origReadFile(path)
					}
					return nil, errors.New("Fake Public key read failure")
				}
				return nil
			},
			expectedError: &testtools.ErrorSpec{
				Type:    "*errors.errorString",
				Message: "Fake Public key read failure",
			},
		},
		{
			desc: "private key changed after write",
			settings: &crypt.PkiSettings{
				Algorithm: x509.ECDSA, // Even with constant pseudo-random input,
				// generated RSA keys seem to vary over time.
				PrivateKeyPath: ".prog/ecdsa_priv.key",
				PublicKeyPath:  ".prog/ecdsa.pub",
			},
			fileSystemState: nil,
			setup: func(mdb *mocks.MockDepsBundle) error {
				origReadFile := mdb.Deps.Io.Ioutil.ReadFile
				counter := 0
				mdb.Deps.Io.Ioutil.ReadFile = func(path string) ([]byte, error) {
					counter++
					if counter < 2 {
						return origReadFile(path)
					}
					return []byte("-----BEGIN ECDSA PUBLIC KEY-----\n" +
						"NotWhatWeWrote00\n" +
						"-----END ECDSA PUBLIC KEY-----\n"), nil
				}
				return nil
			},
			expectedError: &testtools.ErrorSpec{
				Type: "*errors.errorString",
				Message: "File .prog/ecdsa_priv.key contents changed between writing and reading: Was:\n" +
					"-----BEGIN ECDSA PUBLIC KEY-----\n" +
					"MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE7WzVjtn9Gk+WHr5xbv8XMvooqU25\n" +
					"BhgNjZ/vHZLBdVtCOjk4KxjS1UBfQm0c3TRxWBl3hj2AmnJbCrnGofMHBQ==\n" +
					"-----END ECDSA PUBLIC KEY-----\n" +
					"\n" +
					"\n" +
					"Now:\n" +
					"-----BEGIN ECDSA PUBLIC KEY-----\n" +
					"NotWhatWeWrote00\n" +
					"-----END ECDSA PUBLIC KEY-----\n",
			},
		},
		{
			desc: "rsa mode with existing ecdsa keys",
			settings: &crypt.PkiSettings{
				Algorithm:      x509.RSA,
				RSAKeyBits:     2048,
				PrivateKeyPath: ".prog/ecdsa_priv.key",
				PublicKeyPath:  ".prog/ecdsa.pub",
			},
			fileSystemState: testtools.FakeFileSystem{
				"/home/user/.prog/ecdsa_priv.key": testtools.StringPtr("-----BEGIN ECDSA PRIVATE KEY-----\n" +
					"MHcCAQEEIHhN6EluNZ8M83qtHKNw/SXxXZgr1C5r0R/L/ND2IUgjoAoGCCqGSM49\n" +
					"AwEHoUQDQgAE7WzVjtn9Gk+WHr5xbv8XMvooqU25BhgNjZ/vHZLBdVtCOjk4KxjS\n" +
					"1UBfQm0c3TRxWBl3hj2AmnJbCrnGofMHBQ==\n" +
					"-----END ECDSA PRIVATE KEY-----\n"),
				"/home/user/.prog/ecdsa.pub": testtools.StringPtr("-----BEGIN ECDSA PUBLIC KEY-----\n" +
					"MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE7WzVjtn9Gk+WHr5xbv8XMvooqU25\n" +
					"BhgNjZ/vHZLBdVtCOjk4KxjS1UBfQm0c3TRxWBl3hj2AmnJbCrnGofMHBQ==\n" +
					"-----END ECDSA PUBLIC KEY-----\n"),
			},
			setup: func(mdb *mocks.MockDepsBundle) error {
				return nil
			},
			expectedError: &testtools.ErrorSpec{
				Type: "asn1.StructuralError",
				Message: "asn1: structure " +
					"error: tags don't match (2 vs {" +
					"class:0 " +
					"tag:4 " +
					"length:32 " +
					"isCompound:false}) {" +
					"optional:false " +
					"explicit:false " +
					"application:false " +
					"defaultValue:<nil> " +
					"tag:<nil> " +
					"stringType:0 " +
					"timeType:0 " +
					"set:false " +
					"omitEmpty:false}  @5",
			},
		},
	} {
		t.Run(fmt.Sprintf("Subtest: %s", tc.desc), func(tt *testing.T) {
			curFileSysState := tc.fileSystemState.Clone()
			mockDepsBundle := mocks.NewDefaultMockDeps("", []string{"progname"}, "/home/user", &curFileSysState)
			returnedNormally := false
			var tooling *crypt.CryptoTooling
			var actualErr error
			err := mockDepsBundle.InvokeCallInMockedEnv(func() error {
				innerErr := tc.setup(mockDepsBundle)
				if innerErr != nil {
					return innerErr
				}
				var toolingErr error
				tooling, toolingErr = crypt.GetCryptoTooling(mockDepsBundle.Deps, tc.settings)
				if toolingErr != nil {
					return toolingErr
				}
				actualErr = tooling.PopulateKeys()
				returnedNormally = true
				return nil
			})
			if err != nil {
				tt.Errorf("Unexpected error calling mockDepsBundle.InvokeCallInMockedEnv(): %s", err.Error())
			}
			if exitStatus := mockDepsBundle.GetExitStatus(); (exitStatus != 0) || !returnedNormally {
				tt.Error("EncodeAndSaveKey() should not have paniced or called os.Exit.")
			}
			if (mockDepsBundle.OutBuf.String() != "") || (mockDepsBundle.ErrBuf.String() != "") {
				tt.Errorf("EncodeAndSaveKey() should not have output any data. Saw stdout:\n%s\nstderr:\n%s", mockDepsBundle.OutBuf.String(), mockDepsBundle.ErrBuf.String())
			}
			if err := tc.expectedError.EnsureMatches(actualErr); err != nil {
				tt.Error(err.Error())
			}
			if tc.expectedError == nil {
				expectedFileSysState := tc.fileSystemState.Clone()
				if expectedFileSysState == nil {
					expectedFileSysState = make(testtools.FakeFileSystem, 2)
				}
				expectedFileSysState[filepath.Join("/home/user", tc.settings.PrivateKeyPath)] = testtools.StringPtr(tooling.PrivKey.String())
				expectedFileSysState[filepath.Join("/home/user", tc.settings.PublicKeyPath)] = testtools.StringPtr(tooling.PubKey.String())
				if !expectedFileSysState.IsEqualTo(*mockDepsBundle.Files) {
					tt.Errorf("Unexpected change in filesystem state. Expected:\n%s\nActual:\n%s", expectedFileSysState.String(), mockDepsBundle.Files.String())
				}
			}
		})
	}
}

// TestSignMessage tests SignMessage().
func TestSignMessage(t *testing.T) {
	for _, tc := range []struct {
		desc            string
		settings        *crypt.PkiSettings
		fileSystemState testtools.FakeFileSystem
		setup           func(mdb *mocks.MockDepsBundle, setupDone *bool) error
		messageToSign   string
		expectedError   *testtools.ErrorSpec
	}{
		{
			desc: "ecdsa",
			settings: &crypt.PkiSettings{
				Algorithm:      x509.ECDSA,
				PrivateKeyPath: ".prog/ecdsa_priv.key",
				PublicKeyPath:  ".prog/ecdsa.pub",
			},
			fileSystemState: nil,
			setup: func(mdb *mocks.MockDepsBundle, setupDone *bool) error {
				return nil
			},
			messageToSign: "This is a test message",
			expectedError: nil,
		},
		{
			desc: "rsa",
			settings: &crypt.PkiSettings{
				Algorithm:      x509.RSA,
				RSAKeyBits:     2048,
				PrivateKeyPath: ".prog/rsa_priv.key",
				PublicKeyPath:  ".prog/rsa.pub",
			},
			fileSystemState: nil,
			setup: func(mdb *mocks.MockDepsBundle, setupDone *bool) error {
				return nil
			},
			messageToSign: "This is a different test message",
			expectedError: nil,
		},
		{
			desc: "signing failure",
			settings: &crypt.PkiSettings{
				Algorithm:      x509.ECDSA,
				PrivateKeyPath: ".prog/ecdsa_priv.key",
				PublicKeyPath:  ".prog/ecdsa.pub",
			},
			fileSystemState: nil,
			setup: func(mdb *mocks.MockDepsBundle, setupDone *bool) error {
				origCryptoRandReader := mdb.Deps.Crypto.Rand.Reader
				mdb.Deps.Crypto.Rand.Reader = testtools.ReaderFunc(func(p []byte) (n int, err error) {
					if *setupDone {
						return 0, errors.New("Fake I/O Error")
					}
					return origCryptoRandReader.Read(p)
				})
				return nil
			},
			messageToSign: "some other message",
			expectedError: &testtools.ErrorSpec{
				Type:    "*errors.errorString",
				Message: "Fake I/O Error",
			},
		},
	} {
		t.Run(fmt.Sprintf("Subtest: %s", tc.desc), func(tt *testing.T) {
			curFileSysState := tc.fileSystemState.Clone()
			mockDepsBundle := mocks.NewDefaultMockDeps("", []string{"progname"}, "/home/user", &curFileSysState)
			returnedNormally := false
			var tooling *crypt.CryptoTooling
			var actualErr error
			var actualBinSignature crypt.BinarySignature
			err := mockDepsBundle.InvokeCallInMockedEnv(func() error {
				setupComplete := false
				innerErr := tc.setup(mockDepsBundle, &setupComplete)
				if innerErr != nil {
					return innerErr
				}
				var toolingErr error
				tooling, toolingErr = crypt.GetCryptoTooling(mockDepsBundle.Deps, tc.settings)
				if toolingErr != nil {
					return toolingErr
				}
				popKeysErr := tooling.PopulateKeys()
				if popKeysErr != nil {
					return popKeysErr
				}
				setupComplete = true
				actualBinSignature, actualErr = tooling.SignMessage(tc.messageToSign)
				returnedNormally = true
				return nil
			})
			if err != nil {
				tt.Errorf("Unexpected error calling mockDepsBundle.InvokeCallInMockedEnv(): %s", err.Error())
			}
			if exitStatus := mockDepsBundle.GetExitStatus(); (exitStatus != 0) || !returnedNormally {
				tt.Error("EncodeAndSaveKey() should not have paniced or called os.Exit.")
			}
			if (mockDepsBundle.OutBuf.String() != "") || (mockDepsBundle.ErrBuf.String() != "") {
				tt.Errorf("EncodeAndSaveKey() should not have output any data. Saw stdout:\n%s\nstderr:\n%s", mockDepsBundle.OutBuf.String(), mockDepsBundle.ErrBuf.String())
			}
			if err := tc.expectedError.EnsureMatches(actualErr); err != nil {
				tt.Error(err.Error())
			}
			if tc.expectedError == nil {
				expectedFileSysState := tc.fileSystemState.Clone()
				if expectedFileSysState == nil {
					expectedFileSysState = make(testtools.FakeFileSystem, 2)
				}
				expectedFileSysState[filepath.Join("/home/user", tc.settings.PrivateKeyPath)] = testtools.StringPtr(tooling.PrivKey.String())
				expectedFileSysState[filepath.Join("/home/user", tc.settings.PublicKeyPath)] = testtools.StringPtr(tooling.PubKey.String())
				if !expectedFileSysState.IsEqualTo(*mockDepsBundle.Files) {
					tt.Errorf("Unexpected change in filesystem state. Expected:\n%s\nActual:\n%s", expectedFileSysState.String(), mockDepsBundle.Files.String())
				}
				valid, verifyErr := tooling.VerifySignedMessage(tc.messageToSign, actualBinSignature.Base64(), tooling.PubKey.String())
				if verifyErr != nil {
					tt.Errorf("Unexpected error validating signature: %s", verifyErr.Error())
				}
				if !valid {
					tt.Error("Signature not valid")
				}
			}
		})
	}
}

// TestVerifySignedMessage tests VerifySignedMessage() error conditions.
// Happy path was already tested above.
func TestVerifySignedMessage(t *testing.T) {
	for _, tc := range []struct {
		desc             string
		settings         *crypt.PkiSettings
		setup            func(mdb *mocks.MockDepsBundle, setupDone *bool) error
		messageToSign    string
		base64Signature  string
		PEMPublicKey     string
		expectedError    *testtools.ErrorSpec
		expectedValidity bool
	}{
		{
			desc: "invalid base64 signature",
			settings: &crypt.PkiSettings{
				Algorithm:      x509.ECDSA,
				PrivateKeyPath: ".prog/ecdsa_priv.key",
				PublicKeyPath:  ".prog/ecdsa.pub",
			},
			setup: func(mdb *mocks.MockDepsBundle, setupDone *bool) error {
				return nil
			},
			messageToSign:   "some other message",
			base64Signature: "@#$^&*()_",
			PEMPublicKey:    "",
			expectedError: &testtools.ErrorSpec{
				Type:    "base64.CorruptInputError",
				Message: "illegal base64 data at input byte 0",
			},
			expectedValidity: false,
		},
		{
			desc: "empty PEM key",
			settings: &crypt.PkiSettings{
				Algorithm:      x509.ECDSA,
				PrivateKeyPath: ".prog/ecdsa_priv.key",
				PublicKeyPath:  ".prog/ecdsa.pub",
			},
			setup: func(mdb *mocks.MockDepsBundle, setupDone *bool) error {
				return nil
			},
			messageToSign:   "some other message",
			base64Signature: "abcdefgh",
			PEMPublicKey:    "",
			expectedError: &testtools.ErrorSpec{
				Type:    "*errors.errorString",
				Message: "No PEM data was found",
			},
			expectedValidity: false,
		},
		{
			desc: "bad key data",
			settings: &crypt.PkiSettings{
				Algorithm:      x509.ECDSA,
				PrivateKeyPath: ".prog/ecdsa_priv.key",
				PublicKeyPath:  ".prog/ecdsa.pub",
			},
			setup: func(mdb *mocks.MockDepsBundle, setupDone *bool) error {
				return nil
			},
			messageToSign:   "some other message",
			base64Signature: "abcdefgh",
			PEMPublicKey: "-----BEGIN INVALID DATA-----\n" +
				"MTIzNDU2Nzg5MGFiY2RlZmdoaWprbG1ub3BxcnN0dXZ3eHl6\n" +
				"-----END INVALID DATA-----\n",
			expectedError: &testtools.ErrorSpec{
				Type: "asn1.StructuralError",
				Message: "asn1: structure " +
					"error: tags don't match (16 vs {class:0 " +
					"tag:17 " +
					"length:50 " +
					"isCompound:true}) {optional:false " +
					"explicit:false " +
					"application:false " +
					"defaultValue:<nil> " +
					"tag:<nil> " +
					"stringType:0 " +
					"timeType:0 " +
					"set:false " +
					"omitEmpty:false} publicKeyInfo @2",
			},
			expectedValidity: false,
		},
		{
			desc: "invalid signature",
			settings: &crypt.PkiSettings{
				Algorithm:      x509.ECDSA,
				PrivateKeyPath: ".prog/ecdsa_priv.key",
				PublicKeyPath:  ".prog/ecdsa.pub",
			},
			setup: func(mdb *mocks.MockDepsBundle, setupDone *bool) error {
				return nil
			},
			messageToSign:   "some other message",
			base64Signature: "abcdefgh",
			PEMPublicKey: "-----BEGIN ECDSA PUBLIC KEY-----\n" +
				"MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE7WzVjtn9Gk+WHr5xbv8XMvooqU25\n" +
				"BhgNjZ/vHZLBdVtCOjk4KxjS1UBfQm0c3TRxWBl3hj2AmnJbCrnGofMHBQ==\n" +
				"-----END ECDSA PUBLIC KEY-----\n",
			expectedError: &testtools.ErrorSpec{
				Type:    "asn1.SyntaxError",
				Message: "asn1: syntax error: truncated tag or length",
			},
			expectedValidity: false,
		},
		{
			desc: "ecdsa key for rsa mode",
			settings: &crypt.PkiSettings{
				Algorithm:      x509.RSA,
				PrivateKeyPath: ".prog/ecdsa_priv.key",
				PublicKeyPath:  ".prog/ecdsa.pub",
			},
			setup: func(mdb *mocks.MockDepsBundle, setupDone *bool) error {
				return nil
			},
			messageToSign:   "some other message",
			base64Signature: "N3SuIdWI7XlXDteTmcOZUd2OBacyUWY+/+A8SC4QUBz9rXnldBqXha6YyGwnTuizxuy6quQ2QDFdtW16dj7EQk3lozfngskyhc2r86q3AUbdFDvrQVphMQhzsgBhHVoMjCL/YRfvtzCTWhBxegjVMLraLDCBb8IZTIqcMYafYyeJTvAnjBuntlZ+14TDuTt14Uqz85T04CXxBEqlIXMMKpTc01ST4Jsxz5HLO+At1htXp5eHOUFtQSilm3G7iO8ynhgPcXHDWfMAWu6VySUoHWCG70pJaCq6ehF7223t0UFOCqAyDyyQyP9yeUHj8F75SPSxfJm8iKXGx2LND/qLYw==",
			PEMPublicKey: "-----BEGIN RSA PUBLIC KEY-----\n" +
				"MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE7WzVjtn9Gk+WHr5xbv8XMvooqU25\n" +
				"BhgNjZ/vHZLBdVtCOjk4KxjS1UBfQm0c3TRxWBl3hj2AmnJbCrnGofMHBQ==\n" +
				"-----END RSA PUBLIC KEY-----\n",
			expectedError: &testtools.ErrorSpec{
				Type:    "*errors.errorString",
				Message: "Expecting a *rsa.PublicKey, but encountered a *ecdsa.PublicKey instead",
			},
			expectedValidity: false,
		},
		{
			desc: "rsa key for ecdsa mode",
			settings: &crypt.PkiSettings{
				Algorithm:      x509.ECDSA,
				PrivateKeyPath: ".prog/ecdsa_priv.key",
				PublicKeyPath:  ".prog/ecdsa.pub",
			},
			setup: func(mdb *mocks.MockDepsBundle, setupDone *bool) error {
				return nil
			},
			messageToSign:   "some other message",
			base64Signature: "MEYCIQDPM0fc/PFauoZzpltH3RpWtlaqRnL0gFk5WFiLMrFqrwIhAIDvlBozU6Ky2UC9xOSq3YZ5iFuO356t9RnHOElaaXFJ",
			PEMPublicKey: "-----BEGIN RSA PUBLIC KEY-----\n" +
				"MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAzCTTFKQBHfTN8jW6q8PT\n" +
				"HNZKWnRPxSt9kpgWmyqFaZnEUipgoKGAxSIsVrl2PJSm5OlgkVzx+MY+LWM64VKM\n" +
				"bRpUUGJR3zdMNhwZQX0hjOpLpVJvUwD78utVs8vijrU7sH48usFiaZQYjy4m4hQh\n" +
				"63/x4h3KVz7YqUnlRMzYJFT43+AwYzYuEpzWRxtW7IObJPtjtmYVoqva98fF6aj5\n" +
				"uHAsvaAgZGBalHXmCiPzKiGU/halzXSPvyJ2Cqz2aUqMHgwi/2Ip4z/mrfX+mUTa\n" +
				"S+LyBy7GgqJ5vbkGArMagJIc0eARF60r6Uf483xh17oniABdLJy4qlLf6PcEU+ut\n" +
				"EwIDAQAB\n" +
				"-----END RSA PUBLIC KEY-----\n",
			expectedError: &testtools.ErrorSpec{
				Type:    "*errors.errorString",
				Message: "Expecting a *ecdsa.PublicKey, but encountered a *rsa.PublicKey instead",
			},
			expectedValidity: false,
		},
	} {
		t.Run(fmt.Sprintf("Subtest: %s", tc.desc), func(tt *testing.T) {
			mockDepsBundle := mocks.NewDefaultMockDeps("", []string{"progname"}, "/home/user", nil)
			returnedNormally := false
			var tooling *crypt.CryptoTooling
			var actualErr error
			var actualValidity bool
			err := mockDepsBundle.InvokeCallInMockedEnv(func() error {
				setupComplete := false
				innerErr := tc.setup(mockDepsBundle, &setupComplete)
				if innerErr != nil {
					return innerErr
				}
				var toolingErr error
				tooling, toolingErr = crypt.GetCryptoTooling(mockDepsBundle.Deps, tc.settings)
				if toolingErr != nil {
					return toolingErr
				}
				setupComplete = true
				actualValidity, actualErr = tooling.VerifySignedMessage(tc.messageToSign, tc.base64Signature, tc.PEMPublicKey)
				returnedNormally = true
				return nil
			})
			if err != nil {
				tt.Errorf("Unexpected error calling mockDepsBundle.InvokeCallInMockedEnv(): %s", err.Error())
			}
			if exitStatus := mockDepsBundle.GetExitStatus(); (exitStatus != 0) || !returnedNormally {
				tt.Error("EncodeAndSaveKey() should not have paniced or called os.Exit.")
			}
			if (mockDepsBundle.OutBuf.String() != "") || (mockDepsBundle.ErrBuf.String() != "") {
				tt.Errorf("EncodeAndSaveKey() should not have output any data. Saw stdout:\n%s\nstderr:\n%s", mockDepsBundle.OutBuf.String(), mockDepsBundle.ErrBuf.String())
			}
			if err := tc.expectedError.EnsureMatches(actualErr); err != nil {
				tt.Error(err.Error())
			}
			if tc.expectedError == nil {
				if actualValidity != tc.expectedValidity {
					tt.Errorf("Signature is %#v when %#v expected", actualValidity, tc.expectedValidity)
				}
			} else {
				if tc.expectedValidity {
					tt.Error("TEST CASE INVALID. Should not expect \"valid\".")
				}
				if actualValidity {
					tt.Error("Error was expected. Should not report \"valid\".")
				}
			}
		})
	}
}
