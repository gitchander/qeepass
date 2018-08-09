package main

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/gitchander/qeepass/genpas"
)

// https://golang.org/doc/articles/wiki/

func main() {
	http.HandleFunc("/", handler)
	http.HandleFunc("/generate/", generateHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

type Page struct {
	Title string
	Body  []byte
}

func loadPage(title string) (*Page, error) {
	filename := title + ".txt"
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}

type params struct {
	count   int
	passLen int

	genParams genpas.Params
}

func getParamsMap(p params) map[string]string {

	m := make(map[string]string)

	m["Count"] = strconv.Itoa(p.count)
	m["PasswordLength"] = strconv.Itoa(p.passLen)

	const on = "checked"

	boolToChecked := func(b bool) string {
		if b {
			return "checked"
		}
		return ""
	}

	m["UpperLetters"] = boolToChecked(p.genParams.Upper)
	m["LowerLetters"] = boolToChecked(p.genParams.Lower)
	m["Digits"] = boolToChecked(p.genParams.Digits)
	m["SpecialSymbols"] = boolToChecked(p.genParams.Special)

	if p.genParams.ExcludeSimilar {
		m["ExcludeSimilar"] = on
	}
	if p.genParams.HasEveryGroup {
		m["HasEveryGroup"] = on
	}

	return m
}

func parseCheckbox(form url.Values, key string) bool {
	if vs, ok := form[key]; ok {
		if (len(vs) > 0) && (vs[0] == "on") {
			return true
		}
	}
	return false
}

func getIntValue(form url.Values, key string) (int, error) {
	vs, ok := form[key]
	if !ok || (len(vs) == 0) {
		return 0, errors.New("no value")
	}
	return strconv.Atoi(vs[0])
}

func generateHandler(w http.ResponseWriter, r *http.Request) {

	err := r.ParseForm()
	checkError(err)

	//	for key, value := range r.Form {
	//		fmt.Printf("%s = %v\n", key, value)
	//	}

	count, _ := getIntValue(r.Form, "count")
	count = cropInt(count, 1, 100)

	passLen, _ := getIntValue(r.Form, "password_length")
	passLen = cropInt(passLen, 8, 64)

	p := params{
		count:   count,
		passLen: passLen,
		genParams: genpas.Params{
			Upper:   parseCheckbox(r.Form, "upper_letters"),
			Lower:   parseCheckbox(r.Form, "lower_letters"),
			Digits:  parseCheckbox(r.Form, "digits"),
			Special: parseCheckbox(r.Form, "special_symbols"),

			ExcludeSimilar: parseCheckbox(r.Form, "exclude_similar"),
			HasEveryGroup:  parseCheckbox(r.Form, "has_every_group"),
		},
	}
	checkGroups(&(p.genParams))

	g, err := genpas.NewGenerator(p.genParams, genpas.NewRandom())
	checkError(err)

	passwords := make([]string, p.count)
	for i := range passwords {
		passwords[i] = g.Generate(p.passLen)
	}

	//	for _, pass := range passwords {
	//		fmt.Println(pass)
	//	}

	t, err := template.ParseFiles(
		"index.txt",
		"generate.txt",
	)
	checkError(err)

	m := getParamsMap(p)
	m["Result"] = strings.Join(passwords, "\n")

	var buf bytes.Buffer
	err = t.ExecuteTemplate(&buf, "index", m)
	checkError(err)
	_, err = w.Write(buf.Bytes())
	checkError(err)
}

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func cropInt(x, min, max int) int {
	if min > max {
		min, max = max, min
	}
	if x < min {
		x = min
	}
	if x > max {
		x = max
	}
	return x
}

func checkGroups(p *genpas.Params) {
	if (p.Upper == false) &&
		(p.Lower == false) &&
		(p.Digits == false) &&
		(p.Special == false) {

		p.Upper = true
		p.Lower = true
		p.Digits = true
		p.Special = true
	}
}
