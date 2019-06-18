package codechallenge

import (
	"encoding/json"
	"errors"
	"github.com/smartedge/codechallenge/crypt"
	"github.com/smartedge/codechallenge/deps"
)

// SignedMessage the final response to be rendered to JSON.
type SignedMessage struct {
	Message   string `json:"message"`
	Signature string `json:"signature"`
	Pubkey    string `json:"pubkey"`
}

// GenerateResponse takes the message, signature and public key and writes them
// in JSON format to d.Os.Stdout
func GenerateResponse(d *deps.Dependencies, message string, sig crypt.BinarySignature, pubKey crypt.PEMEncoded) error {
	response := SignedMessage{
		Message:   message,
		Signature: sig.Base64(),
		Pubkey:    pubKey.String(),
	}
	buff, err := json.MarshalIndent(&response, "", "")
	if err != nil {
		return err
	}
	n, err := d.Os.Stdout.Write(buff)
	if err != nil {
		return err
	}
	if n < len(buff) {
		return errors.New("failed to write all of response. Should have produced an error to explain failure")
	}
	return nil
}
