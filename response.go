package codechallenge

// SignedMessage the final response to be rendered to JSON.
type SignedMessage struct {
	Message   string `json:"message"`
	Signature string `json:"signature"`
	Pubkey    string `json:"pubkey"`
}
