package crypto

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/alexedwards/argon2id"
)

type crypto struct {
}

func NewCrypto() *crypto {
	return &crypto{}
}

func (c *crypto) EncryptPassword(password string) (string, error) {
	res, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		return "", fmt.Errorf("could not encrypt password: %w", err)
	}

	return res, nil
}

func (c *crypto) ArePasswordAndHashEqual(password string, hash string) (bool, error) {
	return argon2id.ComparePasswordAndHash(password, hash)
}

func (c *crypto) StringToSha256(str string) string {
	sum := sha256.Sum256([]byte(str))
	hex := hex.EncodeToString(sum[:]) // convert to hex because raw sha256 includes binary data

	return hex
}
