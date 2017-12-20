package main

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"

	"github.com/gitchander/qeepass/crypwd"
)

//const pass = "qwerty"
const pass = "HaqmprmJ?5hhT53B1Q?&*?j?R"

const fileName = "../cmd/test.edb"

type Record struct {
	Group    string
	Title    string
	Username string
	Password string
	URL      string
	Notes    string
}

func main() {
	decodeFile()
}

func parse_CSV() {

	data, err := ioutil.ReadFile("ps.csv")
	if err != nil {
		log.Fatal(err)
	}

	rdr := csv.NewReader(bytes.NewReader(data))

	var buf bytes.Buffer

	for {
		record, err := rdr.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatal(err)
		}
		//fmt.Println(len(record))

		rec := Record{
			Group:    record[0],
			Title:    record[1],
			Username: record[2],
			Password: record[3],
			URL:      record[4],
			Notes:    record[5],
		}

		text, err := json.Marshal(rec)
		if err != nil {
			log.Fatal(err)
		}

		//fmt.Println(string(text))

		fmt.Fprintln(&buf, string(text))
	}

	err = ioutil.WriteFile("ps.json", buf.Bytes(), 0600)
	if err != nil {
		log.Fatal(err)
	}
}

func parse_JSON() {
	data, err := ioutil.ReadFile("ps.json")
	if err != nil {
		log.Fatal(err)
	}
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		var rec Record
		err = json.Unmarshal(scanner.Bytes(), &rec)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(rec.Password)
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

func encodeFile() {

	data, err := ioutil.ReadFile("ps.json")
	if err != nil {
		log.Fatal(err)
	}

	var buf bytes.Buffer
	err = crypwd.Encrypt(&buf, pass, data)
	if err != nil {
		log.Fatal(err)
	}

	err = ioutil.WriteFile("chap.edb", buf.Bytes(), 0600)
	if err != nil {
		log.Fatal(err)
	}
}

func decodeFile() {
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Fatal(err)
	}

	text, err := crypwd.Decrypt(bytes.NewReader(data), pass)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(text))
}

func decodeFileAndEncode() {
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Fatal(err)
	}

	text, err := crypwd.Decrypt(bytes.NewReader(data), pass)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(text))

	//------------------------------------

	var buf bytes.Buffer
	err = crypwd.Encrypt(&buf, pass, text)
	if err != nil {
		log.Fatal(err)
	}

	err = ioutil.WriteFile(fileName, buf.Bytes(), 0600)
	if err != nil {
		log.Fatal(err)
	}
}
