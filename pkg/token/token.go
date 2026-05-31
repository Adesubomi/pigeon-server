package token

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base32"
	"encoding/base64"
	"encoding/hex"
	"io"
	"strings"
)

func GenerateURLSafe(byteLength int) (string, error) {
	buf := make([]byte, byteLength)
	if _, err := io.ReadFull(rand.Reader, buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func GeneratePairingCode() (string, error) {
	buf := make([]byte, 5)
	if _, err := io.ReadFull(rand.Reader, buf); err != nil {
		return "", err
	}
	code := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(buf)
	if len(code) > 8 {
		code = code[:8]
	}
	return strings.ToUpper(code), nil
}

func Hash(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func CompareHash(raw, hash string) bool {
	expected := Hash(raw)
	return subtle.ConstantTimeCompare([]byte(expected), []byte(hash)) == 1
}
