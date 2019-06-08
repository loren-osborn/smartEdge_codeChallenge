package codechallenge

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rsa"
	"crypto/x509"
	"encoding/asn1"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
)

// AlgoSpecificOps is used to encapsulate algorithm specific code.
type AlgoSpecificOps struct {
	genPubPrivKeys  func() (x509EncodedPubKey []byte, x509EncodedPrivKey []byte, err error)
	injestPrivKey   func(x509EncodedPrivKey []byte) (signer crypto.Signer, err error)
	verifySignature func(sha256 []byte, binarySignature []byte, x509EncodedPubKey []byte) (bool, error)
	algoName        string
}

// ecdsaSignature is copy of unexported type from ecdsa package for
// unpacking the ECDSA signature.
type ecdsaSignature struct {
	R, S *big.Int
}

// GetKeys retrieves the private key from the filesystem, generating keypair
// if necessary.
func GetKeys(d *Dependencies, keySettings *PkiSettings) (crypto.Signer, error) {
	if FileExists(d, keySettings.PrivateKeyPath) != FileExists(d, keySettings.PublicKeyPath) {
		return nil, fmt.Errorf("Files %s and %s must either both be present or missing", keySettings.PrivateKeyPath, keySettings.PublicKeyPath)
	}
	// keyGenLoader encapsulates the algorithm specific parts of key
	// generation, saving and loading.
	keyGenLoader := AlgoSpecificOps{}
	switch keySettings.Algorithm {
	case x509.ECDSA:
		keyGenLoader.genPubPrivKeys = func() (x509EncodedPubKey []byte, x509EncodedPrivKey []byte, err error) {
			pubkeyCurve := elliptic.P256()
			privatekey, err := ecdsa.GenerateKey(pubkeyCurve, d.Crypto.Rand.Reader)
			if err != nil {
				return nil, nil, err
			}
			x509EncodedPriv, err := x509.MarshalECPrivateKey(privatekey)
			if err != nil {
				return nil, nil, err
			}
			x509EncodedPub, err := x509.MarshalPKIXPublicKey(&privatekey.PublicKey)
			if err != nil {
				return nil, nil, err
			}
			return x509EncodedPub, x509EncodedPriv, nil
		}
		keyGenLoader.injestPrivKey = func(x509EncodedPrivKey []byte) (signer crypto.Signer, err error) {
			return x509.ParseECPrivateKey(x509EncodedPrivKey)
		}
		keyGenLoader.verifySignature = func(sha256 []byte, binarySignature []byte, x509EncodedPubKey []byte) (bool, error) {
			// Decode the signature to get R and S
			sig := ecdsaSignature{}
			_, err := asn1.Unmarshal(binarySignature, sig)

			// Decode the public key
			genericPublicKey, err := x509.ParsePKIXPublicKey(x509EncodedPubKey)
			if err != nil {
				return false, err
			}
			publicKey, ok := genericPublicKey.(*ecdsa.PublicKey)
			if !ok {
				return false, fmt.Errorf("Expecting a *ecdsa.PublicKey, but encountered a %T instead", genericPublicKey)
			}

			// Verify signature
			return ecdsa.Verify(publicKey, sha256, sig.R, sig.S), nil
		}
		keyGenLoader.algoName = "ECDSA"
	case x509.RSA:
		keyGenLoader.genPubPrivKeys = func() (x509EncodedPubKey []byte, x509EncodedPrivKey []byte, err error) {
			privateKey, err := rsa.GenerateKey(d.Crypto.Rand.Reader, keySettings.RSAKeyBits)
			if err != nil {
				return nil, nil, err
			}
			x509EncodedPriv := x509.MarshalPKCS1PrivateKey(privateKey)
			x509EncodedPub, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
			if err != nil {
				return nil, nil, err
			}
			return x509EncodedPub, x509EncodedPriv, nil
		}
		keyGenLoader.injestPrivKey = func(x509EncodedPrivKey []byte) (signer crypto.Signer, err error) {
			return x509.ParsePKCS1PrivateKey(x509EncodedPrivKey)
		}
		keyGenLoader.verifySignature = func(sha256 []byte, binarySignature []byte, x509EncodedPubKey []byte) (bool, error) {
			// Decode the signature to get R and S
			sig := ecdsaSignature{}
			_, err := asn1.Unmarshal(binarySignature, sig)

			// Decode the public key
			genericPublicKey, err := x509.ParsePKIXPublicKey(x509EncodedPubKey)
			if err != nil {
				return false, err
			}
			publicKey, ok := genericPublicKey.(*rsa.PublicKey)
			if !ok {
				return false, fmt.Errorf("Expecting a *rsa.PublicKey, but encountered a %T instead", genericPublicKey)
			}

			// Verify signature
			err = rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, sha256, binarySignature)
			return err == nil, err
		}
		keyGenLoader.algoName = "RSA"
	default:
		return nil, fmt.Errorf("INTERNAL ERROR: Unrecognized key type: %#v", keySettings.Algorithm)
	}

	if !FileExists(d, keySettings.PrivateKeyPath) {
		switch keySettings.Algorithm {
		case x509.ECDSA:
			pubkeyCurve := elliptic.P256()
			_ /* privatekey */, err := ecdsa.GenerateKey(pubkeyCurve, d.Crypto.Rand.Reader)
			if err != nil {
				return nil, err
			}
		case x509.RSA:
		}
		return nil, fmt.Errorf("INTERNAL ERROR: Unrecognized key type: %#v", keySettings.Algorithm)
	}
	return nil, fmt.Errorf("INTERNAL ERROR: nothing should get here")
}

// SerializeECDSAKeyPair wip
func SerializeECDSAKeyPair(privateKey *ecdsa.PrivateKey, publicKey *ecdsa.PublicKey) (string, string) {
	x509Encoded, _ := x509.MarshalECPrivateKey(privateKey)
	pemEncoded := pem.EncodeToMemory(&pem.Block{Type: "ECDSA PRIVATE KEY", Bytes: x509Encoded})

	x509EncodedPub, _ := x509.MarshalPKIXPublicKey(publicKey)
	pemEncodedPub := pem.EncodeToMemory(&pem.Block{Type: "ECDSA PUBLIC KEY", Bytes: x509EncodedPub})

	return string(pemEncoded), string(pemEncodedPub)
}

func decode(pemEncoded string, pemEncodedPub string) (*ecdsa.PrivateKey, *ecdsa.PublicKey) {
	block, _ := pem.Decode([]byte(pemEncoded))
	x509Encoded := block.Bytes
	privateKey, _ := x509.ParseECPrivateKey(x509Encoded)

	blockPub, _ := pem.Decode([]byte(pemEncodedPub))
	x509EncodedPub := blockPub.Bytes
	genericPublicKey, _ := x509.ParsePKIXPublicKey(x509EncodedPub)
	publicKey := genericPublicKey.(*ecdsa.PublicKey)

	return privateKey, publicKey
}

// EncodeAndSaveKey PEM encodes a x509 encoded key and writes it to
// a file. Returns the PEM encoded string data.
func EncodeAndSaveKey(
	d *Dependencies,
	x509EncodedKey []byte,
	label string,
	filename string,
	perm os.FileMode,
) (string, error) {
	pemEncodedKey := pem.EncodeToMemory(&pem.Block{
		Type:  label,
		Bytes: x509EncodedKey,
	})
	if !FileExists(d, filepath.Dir(filename)) {
		var basePerms os.FileMode = 0700
		if (perm & 0070) != 0 {
			basePerms += 0050
			if (perm & 0007) != 0 {
				basePerms += 0005
			}
		}
		err := d.Os.MkdirAll(filepath.Dir(filename), basePerms)
		if err != nil {
			return "", err
		}
	}
	if err := d.Io.Ioutil.WriteFile(filename, pemEncodedKey, perm); err != nil {
		return "", err
	}
	return string(pemEncodedKey), nil
}

// DecodeAndLoadKey loads PEM encoded file and decodes it into a
// x509 encoded key block. Returns PEM encoded data with key block.
func DecodeAndLoadKey(d *Dependencies, filename string) (string, []byte, error) {
	pemEncodedKey, err := d.Io.Ioutil.ReadFile(filename)
	if err != nil {
		return "", nil, err
	}
	blockPub, _ := pem.Decode([]byte(pemEncodedKey))
	if blockPub == nil {
		return "", nil, errors.New("No PEM data was found")
	}
	return string(pemEncodedKey), blockPub.Bytes, nil
}
