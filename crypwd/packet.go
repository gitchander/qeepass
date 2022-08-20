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

func (p *packet) WriteTo(w io.Writer) (int64, error) {

	var n int64

	m, err := w.Write(p.salt[:])
	n += int64(m)
	if err != nil {
		return n, err
	}

	m, err = w.Write(p.iv[:])
	n += int64(m)
	if err != nil {
		return n, err
	}

	m, err = w.Write(p.enc)
	n += int64(m)
	if err != nil {
		return n, err
	}

	return n, nil
}

func (p *packet) ReadFrom(r io.Reader) (int64, error) {

	var n int64

	m, err := io.ReadFull(r, p.salt[:])
	n += int64(m)
	if err != nil {
		return n, err
	}

	m, err = io.ReadFull(r, p.iv[:])
	n += int64(m)
	if err != nil {
		return n, err
	}

	var buf bytes.Buffer
	wn, err := io.Copy(&buf, r)
	n += wn
	if err != nil {
		return n, err
	}
	p.enc = buf.Bytes()

	return n, nil
}
