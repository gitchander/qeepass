package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"

	"qeepass"
	"qeepass/crypwd"
)

func makeTest() {
	var buf bytes.Buffer
	for _, r := range testRecords {
		data, err := json.Marshal(r)
		checkError(err)
		_, err = buf.Write(data)
		checkError(err)
		buf.WriteByte('\n')
	}
	fmt.Println(buf.String())

	err := ioutil.WriteFile("test.txt", buf.Bytes(), 0664)
	checkError(err)

	const password = "HaqmprmJ?5hhT53B1Q?&*?j?R"

	var ew bytes.Buffer
	err = crypwd.Encrypt(&ew, password, buf.Bytes())
	checkError(err)

	err = ioutil.WriteFile("test.edb", ew.Bytes(), 0664)
	checkError(err)
}

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

var testRecords = []qeepass.Record{
	{
		Group: "top/level/my",
		Entry: qeepass.Entry{
			Title:    "one",
			Username: "Chab-Chab",
			Password: "Ws!k^BK2WqdEgEs",
			URL:      "chab-chab.com",
			Notes:    "...",
		},
	},
	{
		Group: "hello/my/friend",
		Entry: qeepass.Entry{
			Title:    "two",
			Username: "Mefodiy",
			Password: "wPR$VANNVDfoZgoM5*2A",
			URL:      "ertywer.com",
			Notes:    "--//--",
		},
	},
	{
		Group: "hello/job",
		Entry: qeepass.Entry{
			Title:    "job-fi",
			Username: "Guloq",
			Password: "?Nh4SkHE?m",
			URL:      "inhotep.sov",
			Notes:    "?",
		},
	},
}
