package main

import (
	"log"
	"time"

	"github.com/atotto/clipboard"
)

func main() {

	password := "Hello, World!"

	err := clipboard.WriteAll(password)
	if err != nil {
		log.Fatal(err)
	}

	<-time.After(10 * time.Second)

	text, err := clipboard.ReadAll()
	if err != nil {
		log.Fatal(err)
	}
	if text == password {
		clipboard.WriteAll("") // clear
	}
}
