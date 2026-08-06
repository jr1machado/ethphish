package models

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"math/big"
)

const tokenAlphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// generateOpaqueToken returns a random alphanumeric token of length n,
// suitable for magic links and session tokens. Follows the same
// crypto/rand approach as generateResultId (models/result.go), sized up
// for tokens that grant a session or a decision rather than just
// identifying a tracked result.
func generateOpaqueToken(n int) (string, error) {
	k := make([]byte, n)
	for i := range k {
		idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(tokenAlphabet))))
		if err != nil {
			return "", err
		}
		k[i] = tokenAlphabet[idx.Int64()]
	}
	return string(k), nil
}

// hashToken returns the SHA-256 hex digest of a token, so bearer tokens
// for approval magic links and client sessions are never stored plaintext.
func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
