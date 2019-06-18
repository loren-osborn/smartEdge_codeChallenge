package crypt

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/asn1"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/smartedge/codechallenge/deps"
	"github.com/smartedge/codechallenge/misc"
	"io"
	"math/big"
	"os"
	"strings"
)

// PkiSettings are the public key settings as specified on the command line.
type PkiSettings struct {
	Algorithm      x509.PublicKeyAlgorithm
	RSAKeyBits     int
	PrivateKeyPath string
	PublicKeyPath  string
}

// KeyType indicates if the key is public or private
type KeyType int

// PublicKey indicates a public key
const (
	PublicKey KeyType = iota
	PrivateKey
)

func (t KeyType) String() string {
	nameLookup := map[KeyType]string{
		PublicKey:  "public",
		PrivateKey: "private",
	}
	name, ok := nameLookup[t]
	if !ok {
		return fmt.Sprintf("Unknown KeyType %#v (INTERNAL ERROR)", t)
	}
	return name
}

// AlgorithmPlugin is used to encapsulate algorithm specific code.
type AlgorithmPlugin interface {
	GenKeyPair(randReader io.Reader) (pubKey X509Encoded, privKey X509Encoded, err error)
	InjestPrivateKey(privKey X509Encoded) (signer crypto.Signer, err error)
	HashMessage(message string) DigestHash
	VerifySignature(sha256Hash DigestHash, binSig BinarySignature, pubKey X509Encoded) (bool, error)
	GetAlgorithmName() string
}

// ECDSAPlugin Implementation details for ECDSA.
type ECDSAPlugin struct{}

// GenKeyPair generates a new ECDSA public and private key pair
func (p *ECDSAPlugin) GenKeyPair(randReader io.Reader) (pubKey X509Encoded, privKey X509Encoded, err error) {
	pubkeyCurve := elliptic.P256()
	privatekey, err := ecdsa.GenerateKey(pubkeyCurve, randReader)
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
	return X509Encoded(x509EncodedPub), X509Encoded(x509EncodedPriv), nil
}

// InjestPrivateKey loads a ECDSA private key from a X509Encoded buffer,
func (p *ECDSAPlugin) InjestPrivateKey(privKey X509Encoded) (signer crypto.Signer, err error) {
	return x509.ParseECPrivateKey([]byte(privKey))
}

// HashMessage respecting whatever salting is necessary, Since
// ECDSA needs no special hashing, this is a thin wrapper over
// crypto/sha256.Sum256()
func (p *ECDSAPlugin) HashMessage(message string) DigestHash {
	digest := sha256.Sum256([]byte(message))
	return DigestHash(digest[0:])
}

// VerifySignature verifies a ECDSA signature for a message digest,
func (p *ECDSAPlugin) VerifySignature(sha256Hash DigestHash, binSig BinarySignature, pubKey X509Encoded) (bool, error) {
	// Decode the signature to get R and S
	sigStruct := ecdsaSignature{}
	_, err := asn1.Unmarshal([]byte(binSig), &sigStruct)
	if err != nil {
		return false, err
	}

	// Decode the public key
	genericPublicKey, err := x509.ParsePKIXPublicKey([]byte(pubKey))
	if err != nil {
		return false, err
	}
	publicKey, ok := genericPublicKey.(*ecdsa.PublicKey)
	if !ok {
		return false, fmt.Errorf("Expecting a *ecdsa.PublicKey, but encountered a %T instead", genericPublicKey)
	}

	// Verify signature
	return ecdsa.Verify(publicKey, []byte(sha256Hash), sigStruct.R, sigStruct.S), nil
}

// GetAlgorithmName returns the string "ECDSA"
func (p *ECDSAPlugin) GetAlgorithmName() string {
	return "ECDSA"
}

// RSAPlugin Implementation details for RSA.
type RSAPlugin struct {
	KeyLen int
}

// GenKeyPair generates a new RSA public and private key pair
func (p *RSAPlugin) GenKeyPair(randReader io.Reader) (pubKey X509Encoded, privKey X509Encoded, err error) {
	privateKey, err := rsa.GenerateKey(randReader, p.KeyLen)
	if err != nil {
		return nil, nil, err
	}
	x509EncodedPriv := x509.MarshalPKCS1PrivateKey(privateKey)
	x509EncodedPub, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, nil, err
	}
	return X509Encoded(x509EncodedPub), X509Encoded(x509EncodedPriv), nil
}

// InjestPrivateKey loads a RSA private key from a X509Encoded buffer,
func (p *RSAPlugin) InjestPrivateKey(privKey X509Encoded) (signer crypto.Signer, err error) {
	return x509.ParsePKCS1PrivateKey([]byte(privKey))
}

// HashMessage respecting whatever salting is necessary, Here we
// handle PSS salting
func (p *RSAPlugin) HashMessage(message string) DigestHash {
	pssh := crypto.SHA256.New()
	pssh.Write([]byte(message))
	return DigestHash(pssh.Sum(nil))
}

// VerifySignature verifies a RSA signature for a message digest,
func (p *RSAPlugin) VerifySignature(sha256Hash DigestHash, binSig BinarySignature, pubKey X509Encoded) (bool, error) {
	// Decode the public key
	genericPublicKey, err := x509.ParsePKIXPublicKey([]byte(pubKey))
	if err != nil {
		return false, err
	}
	publicKey, ok := genericPublicKey.(*rsa.PublicKey)
	if !ok {
		return false, fmt.Errorf("Expecting a *rsa.PublicKey, but encountered a %T instead", genericPublicKey)
	}

	// Verify signature
	err = rsa.VerifyPSS(
		publicKey,
		crypto.SHA256,
		[]byte(sha256Hash),
		[]byte(binSig),
		&rsa.PSSOptions{
			SaltLength: rsa.PSSSaltLengthAuto,
			Hash:       crypto.SHA256,
		})
	return err == nil, err
}

// GetAlgorithmName returns the string "RSA"
func (p *RSAPlugin) GetAlgorithmName() string {
	return "RSA"
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

// HashMessage defers to the AlgPlugin to handle hashing.
func (ct *CryptoTooling) HashMessage(msg string) DigestHash {
	return ct.AlgPlugin.HashMessage(msg)
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
	return ct.Sign(ct.HashMessage(msg))
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
	valid, err := ct.AlgPlugin.VerifySignature(ct.HashMessage(msg), sig, x509PubKey)
	if err != nil {
		return false, err
	}
	return valid, nil
}

// X509Encoded data buffer
type X509Encoded []byte

// EncodeToPEM encodes the x509 key as a PEM text block
func (x X509Encoded) EncodeToPEM(algorithm string, kt KeyType) PEMEncoded {
	pemEncodedKey := pem.EncodeToMemory(&pem.Block{
		Type:  strings.ToUpper(fmt.Sprintf("%s %s key", algorithm, kt.String())),
		Bytes: []byte(x),
	})
	return PEMEncoded(pemEncodedKey)
}

// PEMEncoded text data buffer
type PEMEncoded []byte

// NewPEMBufferFromString turns a string into a PEM buffer.
func NewPEMBufferFromString(src string) PEMEncoded {
	return PEMEncoded([]byte(src))
}

// String renders the PEM encoded data as a string.
func (pemBuf PEMEncoded) String() string {
	return string([]byte(pemBuf))
}

// DecodeToX509 decodes the PEM key data block to a x509 buffer
func (pemBuf PEMEncoded) DecodeToX509() (X509Encoded, error) {
	blockPub, _ := pem.Decode([]byte(pemBuf))
	if blockPub == nil {
		return nil, errors.New("No PEM data was found")
	}
	return X509Encoded(blockPub.Bytes), nil
}

// DigestHash data buffer
type DigestHash []byte

// Hex renders the hash digest as a hex string.
// This is primarily for debugging and error messages.
func (hash DigestHash) Hex() string {
	return hex.EncodeToString([]byte(hash))
}

// BinarySignature data buffer
type BinarySignature []byte

// NewBinarySignatureFromBase64 creates a new BinarySignature buffer
// from a base64 string.
func NewBinarySignatureFromBase64(src string) (BinarySignature, error) {
	buf, err := base64.StdEncoding.DecodeString(src)
	if err != nil {
		return nil, err
	}
	return BinarySignature(buf), nil
}

// Base64 renders the signature as a RFC 4648 compliant Base64
// encoded string.
func (sig BinarySignature) Base64() string {
	return base64.StdEncoding.EncodeToString([]byte(sig))
}

// ecdsaSignature is copy of unexported type from ecdsa package for
// unpacking the ECDSA signature.
type ecdsaSignature struct {
	R, S *big.Int
}

// EncodeAndSaveKey PEM encodes a x509 encoded key and writes it to
// a file. Returns the PEM encoded string data.
func EncodeAndSaveKey(
	d *deps.Dependencies,
	keyBuf X509Encoded,
	algorithm string,
	kt KeyType,
	filename string,
	perm os.FileMode,
) (PEMEncoded, error) {
	pemEncodedKey := keyBuf.EncodeToPEM(algorithm, kt)
	var dirPerm os.FileMode = 0700
	if (perm & 0070) != 0 {
		dirPerm += 0050
		if (perm & 0007) != 0 {
			dirPerm += 0005
		}
	}
	err := misc.WriteDirAndFile(d, filename, []byte(pemEncodedKey.String()), perm, dirPerm)
	if err != nil {
		return nil, err
	}
	return pemEncodedKey, nil
}

// DecodeAndLoadKey loads PEM encoded file and decodes it into a
// x509 encoded key block. Returns PEM encoded data with key block.
func DecodeAndLoadKey(d *deps.Dependencies, filename string) (PEMEncoded, X509Encoded, error) {
	pemEncodedKey, err := d.Io.Ioutil.ReadFile(filename)
	if err != nil {
		return nil, nil, err
	}
	pemKey := PEMEncoded(pemEncodedKey)
	x509Key, err := pemKey.DecodeToX509()
	if err != nil {
		return nil, nil, err
	}
	return pemKey, x509Key, err
}
