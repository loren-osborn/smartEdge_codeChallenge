package crypt

import (
	"crypto"
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"github.com/smartedge/codechallenge/deps"
	"github.com/smartedge/codechallenge/misc"
	"io"
)

// PkiSettings are the public key settings as specified on the command line.
type PkiSettings struct {
	Algorithm      x509.PublicKeyAlgorithm
	RSAKeyBits     int
	PrivateKeyPath string
	PublicKeyPath  string
}

// AlgorithmPlugin is used to encapsulate algorithm specific code.
type AlgorithmPlugin interface {
	GenKeyPair(randReader io.Reader) (pubKey X509Encoded, privKey X509Encoded, err error)
	InjestPrivateKey(privKey X509Encoded) (signer crypto.Signer, err error)
	VerifySignature(sha256Hash DigestHash, binSig BinarySignature, publicKey crypto.PublicKey) (bool, error)
	GetAlgorithmName() string
}

// CryptoTooling home to all crypto tool state.
type CryptoTooling struct {
	D         *deps.Dependencies
	Settings  *PkiSettings
	AlgPlugin AlgorithmPlugin
	PubKey    PEMEncoded
	PrivKey   PEMEncoded
	Signer    crypto.Signer
}

// GetCryptoTooling returns a home where all the keys, signing and
// verification lives.
func GetCryptoTooling(d *deps.Dependencies, keySettings *PkiSettings) (*CryptoTooling, error) {
	result := CryptoTooling{
		D:        d,
		Settings: keySettings,
	}
	switch result.Settings.Algorithm {
	case x509.ECDSA:
		result.AlgPlugin = &ECDSAPlugin{}
	case x509.RSA:
		result.AlgPlugin = &RSAPlugin{
			KeyLen: result.Settings.RSAKeyBits,
		}
	default:
		return nil, fmt.Errorf("INTERNAL ERROR: Unrecognized algorithm: %#v", result.Settings.Algorithm)
	}
	return &result, nil
}

// GetKeys retrieves the private key from the filesystem, generating keypair
// if necessary.
func (ct *CryptoTooling) GetKeys() error {
	if misc.FileExists(ct.D, ct.Settings.PrivateKeyPath) != misc.FileExists(ct.D, ct.Settings.PublicKeyPath) {
		return fmt.Errorf("Files %s and %s must either both be present or missing", ct.Settings.PrivateKeyPath, ct.Settings.PublicKeyPath)
	}
	if !misc.FileExists(ct.D, ct.Settings.PrivateKeyPath) {
		x509PubKey, x509PrivKey, err := ct.AlgPlugin.GenKeyPair(ct.D.Crypto.Rand.Reader)
		if err != nil {
			return err
		}
		ct.PubKey, err = EncodeAndSaveKey(ct.D, x509PubKey, ct.AlgPlugin.GetAlgorithmName(), PublicKey, ct.Settings.PublicKeyPath, 0444)
		if err != nil {
			return err
		}
		ct.PrivKey, err = EncodeAndSaveKey(ct.D, x509PrivKey, ct.AlgPlugin.GetAlgorithmName(), PrivateKey, ct.Settings.PrivateKeyPath, 0400)
		if err != nil {
			return err
		}
	}
	pemPrivKey, x509PrivKey, err := DecodeAndLoadKey(ct.D, ct.Settings.PrivateKeyPath)
	if err != nil {
		return err
	}
	if ct.PrivKey != nil {
		if ct.PrivKey.String() != pemPrivKey.String() {
			return fmt.Errorf(
				"File %s contents changed between writing and reading: "+
					"Was:\n%s\n\nNow:\n%s",
				ct.Settings.PrivateKeyPath,
				ct.PrivKey.String(),
				pemPrivKey.String())
		}
	} else {
		ct.PrivKey = pemPrivKey
	}
	pemPubKey, _, err := DecodeAndLoadKey(ct.D, ct.Settings.PublicKeyPath)
	if err != nil {
		return err
	}
	if ct.PubKey != nil {
		if ct.PubKey.String() != pemPubKey.String() {
			return fmt.Errorf(
				"File %s contents changed between writing and reading: "+
					"Was:\n%s\n\nNow:\n%s",
				ct.Settings.PrivateKeyPath,
				ct.PubKey.String(),
				pemPubKey.String())
		}
	} else {
		ct.PubKey = pemPubKey
	}
	ct.Signer, err = ct.AlgPlugin.InjestPrivateKey(x509PrivKey)
	if err != nil {
		return err
	}
	return nil
}

// Sign is a thin wrapper over cryptoSigner.Sign() to ease
// type conversions and dependencies.
func (ct *CryptoTooling) Sign(digest DigestHash) (BinarySignature, error) {
	signature, err := ct.Signer.Sign(
		ct.D.Crypto.Rand.Reader,
		[]byte(digest),
		&rsa.PSSOptions{
			SaltLength: rsa.PSSSaltLengthAuto,
			Hash:       crypto.SHA256,
		})
	if err != nil {
		return nil, err
	}
	return BinarySignature(signature), nil
}

// SignMessage simply sighs a hash of the message. It was added for
// consistancy with VerifySignedMessage.
func (ct *CryptoTooling) SignMessage(msg string) (BinarySignature, error) {
	return ct.Sign(NewSHA256DigestHash(msg))
}

// VerifySignedMessage simply sighs a hash of the message. It was added for
// consistancy with VerifySignedMessage.
func (ct *CryptoTooling) VerifySignedMessage(msg string, base64Sig string, pemPubKey string) (bool, error) {
	sig, err := NewBinarySignatureFromBase64(base64Sig)
	if err != nil {
		return false, err
	}
	x509PubKey, err := NewPEMBufferFromString(pemPubKey).DecodeToX509()
	if err != nil {
		return false, err
	}
	genericPubKey, err := x509PubKey.AsGenericPublicKey()
	if err != nil {
		return false, err
	}
	valid, err := ct.AlgPlugin.VerifySignature(NewSHA256DigestHash(msg), sig, genericPubKey)
	if err != nil {
		return false, err
	}
	return valid, nil
}
