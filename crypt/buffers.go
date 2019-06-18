package crypt

import (
	"crypto"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/smartedge/codechallenge/deps"
	"github.com/smartedge/codechallenge/misc"
	"math/big"
	"os"
	"strings"
)

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

// AsGenericPublicKey decodes the public key (if it is one)
func (x X509Encoded) AsGenericPublicKey() (crypto.PublicKey, error) {
	genericPublicKey, err := x509.ParsePKIXPublicKey([]byte(x))
	if err != nil {
		return false, err
	}
	return crypto.PublicKey(genericPublicKey), nil
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
