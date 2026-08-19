package crypto

import (
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type crypto struct {
}

func NewCrypto() *crypto {
	return &crypto{
	}
}

func (c *crypto) EncryptPassword(password string) ([]byte, error) {
	res, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("could not encrypt password: %w", err)
	}

	return res, nil
}

func (c *crypto) ArePasswordAndHashEqual(password string, hash string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return false, nil
		}

		return false, fmt.Errorf("could not compare password and hash: %w", err)
	}

	return true, nil
}
