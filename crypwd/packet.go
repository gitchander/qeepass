package crypwd

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
)

const saltSize = 64

type packet struct {
	salt [saltSize]byte      // salt for scrypt
	iv   [aes.BlockSize]byte // init vector
	enc  []byte              // encrypted data
}

func (p *packet) RandomSaltAndIV() error {
	if _, err := rand.Read(p.salt[:]); err != nil {
		return err
	}
	if _, err := rand.Read(p.iv[:]); err != nil {
		return err
	}
	return nil
}

func (p *packet) BlockAES(password string) (cipher.Block, error) {

	key, err := keyScrypt([]byte(password), p.salt[:])
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}

	return block, nil
}

func (p *packet) WriteTo(w io.Writer) error {

	if _, err := w.Write(p.salt[:]); err != nil {
		return err
	}

	if _, err := w.Write(p.iv[:]); err != nil {
		return err
	}

	if _, err := w.Write(p.enc); err != nil {
		return err
	}

	return nil
}

func (p *packet) ReadFrom(r io.Reader) error {

	if _, err := io.ReadFull(r, p.salt[:]); err != nil {
		return err
	}

	if _, err := io.ReadFull(r, p.iv[:]); err != nil {
		return err
	}

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		return err
	}
	p.enc = buf.Bytes()

	return nil
}
