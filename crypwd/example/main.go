package main

import (
	"bytes"
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/gitchander/qeepass/crypwd"
	"github.com/gitchander/qeepass/genpas"
)

func main() {
	begin := time.Now()

	r := genpas.NewRandom()
	g, err := genpas.NewGenerator(genpas.DefaultParams, r)
	if err != nil {
		log.Fatal(err)
	}

	for i := 0; i < 10; i++ {
		var (
			pass = g.Generate(r.IntRange(4, 11))
			data = make([]byte, r.Intn(100))
		)
		_, err := rand.Read(data)
		if err != nil {
			log.Fatal(err)
		}
		err = checkEncDec(pass, data)
		if err != nil {
			log.Fatal(err)
		}
	}
	fmt.Println(time.Since(begin))
}

func checkEncDec(pass string, data []byte) error {

	var buf bytes.Buffer

	err := crypwd.Encrypt(&buf, pass, data)
	if err != nil {
		return err
	}

	decData, err := crypwd.Decrypt(&buf, pass)
	if err != nil {
		return err
	}

	if !bytes.Equal(data, decData) {
		return errors.New("not equal")
	}

	return nil
}
