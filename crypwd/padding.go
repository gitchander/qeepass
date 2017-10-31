package crypwd

// https://en.wikipedia.org/wiki/Padding_(cryptography)#Bit_padding

func appendPadding(data []byte, blockSize int) []byte {
	n := blockSize - (len(data) % blockSize)
	padding := make([]byte, n)
	for i := range padding {
		padding[i] = byte(n)
	}
	return append(data, padding...)
}

func cutPadding(data []byte, blockSize int) []byte {

	n := int(data[len(data)-1])

	if n > len(data) {
		panic("wrong padding len")
	}

	return data[:len(data)-n]
}
