package main

import (
	"bufio"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io"
)

const (
	bitsCount = 3
	randMax   = ^(((^uint32(0)) >> bitsCount) << bitsCount)

	sizeofUint32 = 4
)

var random = func() func() uint32 {
	br := bufio.NewReader(rand.Reader)
	buffer := make([]byte, sizeofUint32)
	return func() uint32 {
		_, err := io.ReadFull(br, buffer)
		if err != nil {
			panic(err)
		}
		v := binary.BigEndian.Uint32(buffer)
		return v & randMax
	}
}()

func randIntnBad(n uint32) uint32 {
	return random() % n
}

// To avoid modulo bias:
// Make sure random is below the largest number where (random % n) == 0
func randIntnGood(n uint32) uint32 {
	limit := randMax - (randMax % n)
	for {
		x := random()
		if x < limit {
			return x % n
		}
	}
}

func test(randIntn func(n uint32) uint32) {
	n := 3
	var total int
	m := make(map[int]int)
	for i := 0; i < 10000000; i++ {
		r := randIntn(uint32(n))
		m[int(r)]++
		total++
	}
	for i := 0; i < n; i++ {
		fmt.Printf("%4d: %.2f%%\n", i, float64(100*m[i])/float64(total))
	}
}

func main() {
	//fmt.Println("randMax:", randMax)

	fmt.Println("Test \"Bad\" rand.Intn:")
	test(randIntnBad)

	fmt.Println("Test \"Good\" rand.Intn:")
	test(randIntnGood)
}
