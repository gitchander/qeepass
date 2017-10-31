package genpas

import (
	"bufio"
	"crypto/rand"
	"encoding/binary"
	"io"
	"math"
)

const (
	sizeofUint32 = 4
	sizeofUint64 = 8
)

type Random struct {
	source io.Reader
	buf    [sizeofUint64]byte
}

func NewRandom() *Random {
	return &Random{
		source: bufio.NewReader(rand.Reader),
	}
}

func NewRandomReader(source io.Reader) *Random {
	return &Random{source: source}
}

func (r *Random) Uint32() uint32 {
	buf := r.buf[:sizeofUint32]
	_, err := io.ReadFull(r.source, buf)
	if err != nil {
		panic(err)
	}
	return binary.BigEndian.Uint32(buf)
}

func (r *Random) Uint64() uint64 {
	buf := r.buf[:sizeofUint64]
	_, err := io.ReadFull(r.source, buf)
	if err != nil {
		panic(err)
	}
	return binary.BigEndian.Uint64(buf)
}

// To avoid modulo bias:
// Make sure random is below the largest number where (random % n) == 0

func (r *Random) Uint32n(n uint32) uint32 {
	const max uint32 = math.MaxUint32
	var limit = max - (max % n)
	for {
		x := r.Uint32()
		if x < limit {
			return x % n
		}
	}
}

func (r *Random) Uint64n(n uint64) uint64 {
	const max uint64 = math.MaxUint64
	var limit = max - (max % n)
	for {
		x := r.Uint64()
		if x < limit {
			return x % n
		}
	}
}

// pseudo random number [0, n)
func (r *Random) Intn(n int) int {
	if n <= 0 {
		panic("invalid argument to Intn")
	}
	if n <= math.MaxInt32 {
		return int(r.Uint32n(uint32(n)))
	}
	return int(r.Uint64n(uint64(n)))
}

// pseudo random number [min, max)
func (r *Random) IntRange(min, max int) int {
	if min > max {
		min, max = max, min
	}
	return min + r.Intn(max-min)
}
