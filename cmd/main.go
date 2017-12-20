package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/gitchander/qeepass"
	"github.com/gitchander/qeepass/crypwd"
	"github.com/howeyc/gopass"
)

func main() {
	Main()
	//makeTest()
}

func Main() {
	if len(os.Args) < 2 {
		log.Fatal("need an argument")
	}
	filename := os.Args[1]

	data, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Print("Enter master password ")
	pass, err := gopass.GetPasswd()
	if err != nil {
		log.Fatal(err)
	}

	dec, err := crypwd.Decrypt(bytes.NewReader(data), string(pass))
	if err != nil {
		log.Fatal(err)
	}

	root, err := parseLines(dec)
	if err != nil {
		log.Fatal(err)
	}

	_ = root

	var br = bufio.NewReader(os.Stdin)

mainLoop:
	for {
		fmt.Print("> ")

		line, _, err := br.ReadLine()
		if err != nil {
			log.Fatal(err)
		}

		//strings.HasPrefix()

		parts := strings.Split(string(line), " ")
		//parts := bytes.Split(line, []byte{' '})
		if len(parts) == 0 {
			continue
		}

		fmt.Println(len(parts), parts)

		cmd := string(parts[0])
		switch cmd {
		case "quit", "exit":
			break mainLoop
		}
	}
}

func parseLines(data []byte) (*qeepass.Group, error) {

	var recs []qeepass.Record

	var scanner = bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		var rec qeepass.Record
		err := json.Unmarshal(scanner.Bytes(), &rec)
		if err != nil {
			return nil, err
		}
		recs = append(recs, rec)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	g := qeepass.NewGroupFromRecords(recs)
	printGroup(g, 0, 0)

	return g, nil
}

func printGroup(g *qeepass.Group, i int, tab int) {
	tabData := make([]byte, tab)
	for i := range tabData {
		tabData[i] = '\t'
	}
	fmt.Printf("%s%d> %s\n", string(tabData), i, g.Name)

	for i, entry := range g.Entrys {
		printEntry(entry, i, tab+1)
	}

	for i, child := range g.Childs {
		printGroup(child, i, tab+1)
	}
}

func printEntry(e *qeepass.Entry, i int, tab int) {
	tabData := make([]byte, tab)
	for i := range tabData {
		tabData[i] = '\t'
	}
	fmt.Printf("%s%d] %s\n", string(tabData), i, e.Title)
}

// func split()
