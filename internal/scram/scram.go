package scram

import (
	"crypto/hmac"
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
)

const (
	defaultIterationCount = 4096
	defaultSaltLength     = 16
)

type ScramSecret struct {
	iterationCount int
	salt           []byte
	storedKey      []byte
	serverKey      []byte
}

func GenerateScramSHA256Secret(password string, salt []byte, iterationCount int) ScramSecret {
	if salt == nil {
		salt := make([]byte, defaultSaltLength)
		if _, err := rand.Read(salt); err != nil {
			panic(err)
		}
	}
	if iterationCount == 0 {
		iterationCount = defaultIterationCount
	}

	saltedPassword, err := pbkdf2.Key(sha256.New, password, salt, iterationCount, sha256.Size)
	if err != nil {
		panic(err)
	}

	clientKeyHmac := hmac.New(sha256.New, saltedPassword)
	clientKeyHmac.Write([]byte("Client Key"))
	clientKey := clientKeyHmac.Sum(nil)
	storedKey := sha256.Sum256(clientKey)

	serverKeyHmac := hmac.New(sha256.New, saltedPassword)
	serverKeyHmac.Write([]byte("Server Key"))
	serverKey := serverKeyHmac.Sum(nil)

	return ScramSecret{
		iterationCount: iterationCount,
		salt:           salt,
		storedKey:      storedKey[:],
		serverKey:      serverKey,
	}
}

func (s *ScramSecret) EncodeToBase64() string {
	return fmt.Sprintf(
		"SCRAM-SHA-256$%d:%s$%s:%s",
		s.iterationCount,
		base64.StdEncoding.EncodeToString(s.salt),
		base64.StdEncoding.EncodeToString(s.storedKey),
		base64.StdEncoding.EncodeToString(s.serverKey),
	)
}
