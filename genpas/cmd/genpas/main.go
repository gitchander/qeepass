package main

import (
	"errors"
	"flag"
	"fmt"
	"log"

	"qeepass/genpas"
)

// ./genpas -len=100 -count=1000 -groups="ulds" >> passwords.txt

func main() {

	var (
		count          = flag.Int("count", 1, "count of password")
		length         = flag.Int("len", 8, "password length")
		groups         = flag.String("groups", "ulds", "u:upper, l:lower, d:digits, s:special")
		excludeSimilar = flag.Bool("exclude-similar", false, "exclude similar symbols")
		everyGroup     = flag.Bool("every-group", false, "has every group")
	)

	flag.Parse()

	err := execute(*count, *length, *groups, *excludeSimilar, *everyGroup)
	if err != nil {
		log.Fatal(err)
	}
}

func execute(count int, length int, groups string, excludeSimilar bool, everyGroup bool) error {
	if length < 0 {
		return errors.New("length is below the zero")
	}
	var p genpas.Params
	err := parseGroups(groups, &p)
	if err != nil {
		return err
	}
	p.ExcludeSimilar = excludeSimilar
	p.HasEveryGroup = everyGroup

	g, err := genpas.NewGenerator(p, genpas.NewRandom())
	if err != nil {
		return err
	}
	for i := 0; i < count; i++ {
		fmt.Println(g.Generate(length))
	}
	return nil
}

func parseGroups(groups string, p *genpas.Params) error {
	for _, r := range groups {
		switch r {
		case 'u':
			p.Upper = true
		case 'l':
			p.Lower = true
		case 'd':
			p.Digits = true
		case 's':
			p.Special = true
		default:
			return errors.New("invalid group type")
		}
	}
	return nil
}
