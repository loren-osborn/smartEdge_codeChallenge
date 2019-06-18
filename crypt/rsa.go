package crypt

import (
	"crypto"
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"io"
)

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
func (p *RSAPlugin) VerifySignature(sha256Hash DigestHash, binSig BinarySignature, publicKey crypto.PublicKey) (bool, error) {
	rsaPublicKey, ok := publicKey.(*rsa.PublicKey)
	if !ok {
		return false, fmt.Errorf("Expecting a *rsa.PublicKey, but encountered a %T instead", publicKey)
	}

	// Verify signature
	err := rsa.VerifyPSS(
		rsaPublicKey,
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
