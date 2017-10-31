package crypwd

import (
	"golang.org/x/crypto/scrypt"
)

const keyLen = 32

// Scrypt
func keyScrypt(password, salt []byte) (key [keyLen]byte, err error) {
	keyData, err := scrypt.Key(password, salt, 16384, 8, 1, keyLen)
	copy(key[:], keyData)
	return
}
