package crypt

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/x509"
	"encoding/asn1"
	"fmt"
	"io"
)

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

// VerifySignature verifies a ECDSA signature for a message digest,
func (p *ECDSAPlugin) VerifySignature(sha256Hash DigestHash, binSig BinarySignature, publicKey crypto.PublicKey) (bool, error) {
	// Decode the signature to get R and S
	sigStruct := ecdsaSignature{}
	_, err := asn1.Unmarshal([]byte(binSig), &sigStruct)
	if err != nil {
		return false, err
	}
	ecdsaPublicKey, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return false, fmt.Errorf("Expecting a *ecdsa.PublicKey, but encountered a %T instead", publicKey)
	}

	// Verify signature
	return ecdsa.Verify(ecdsaPublicKey, []byte(sha256Hash), sigStruct.R, sigStruct.S), nil
}

// GetAlgorithmName returns the string "ECDSA"
func (p *ECDSAPlugin) GetAlgorithmName() string {
	return "ECDSA"
}
