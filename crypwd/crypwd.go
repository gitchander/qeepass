package crypwd

import (
	"bytes"
	"crypto/cipher"
	"crypto/sha256"
	"errors"
	"io"
)

var ErrDecrypt = errors.New("decrypt error")

func Encrypt(w io.Writer, password string, data []byte) error {

	var p packet

	p.RandomSaltAndIV()

	block, err := p.BlockAES(password)
	if err != nil {
		return err
	}

	data = appendSHA256(data)

	p.enc = encrypt_CFB(data, block, p.iv[:])
	return p.WriteTo(w)
}

func Decrypt(r io.Reader, password string) ([]byte, error) {

	var p packet

	if err := p.ReadFrom(r); err != nil {
		return nil, err
	}

	block, err := p.BlockAES(password)
	if err != nil {
		return nil, err
	}

	data := decrypt_CFB(p.enc, block, p.iv[:])

	data, err = checkAndCutSHA256(data)
	if err != nil {
		return nil, ErrDecrypt
	}

	return data, nil
}

func encrypt_CFB(src []byte, block cipher.Block, iv []byte) (dst []byte) {

	src = appendPadding(src, block.BlockSize())
	dst = make([]byte, len(src))

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(dst, src)

	return dst
}

func decrypt_CFB(src []byte, block cipher.Block, iv []byte) (dst []byte) {

	dst = make([]byte, len(src))

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(dst, src)

	dst = cutPadding(dst, block.BlockSize())

	return dst
}

func appendSHA256(data []byte) []byte {
	sum := sha256.Sum256(data)
	return append(data, sum[:]...)
}

func checkAndCutSHA256(data []byte) ([]byte, error) {
	if len(data) < sha256.Size {
		return nil, errors.New("short data len")
	}
	n := len(data) - sha256.Size
	sum := sha256.Sum256(data[:n])
	if !bytes.Equal(sum[:], data[n:]) {
		return nil, errors.New("checksum not valid")
	}
	return data[:n], nil
}
