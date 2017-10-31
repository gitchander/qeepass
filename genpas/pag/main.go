package main

import (
	"errors"
	"flag"
	"fmt"
	"log"

	"qeepass/genpas"
)

// ./pag -len=100 -count=1000 -groups="ulds" >> passwords.txt

func main() {

	var (
		pCount  = flag.Int("count", 1, "count of password")
		pLen    = flag.Int("len", 8, "password length")
		pGroups = flag.String("groups", "ulds", "u:upper, l:lower, d:digits, s:special")
		pFlags  = flag.String("flags", "ea", "e:exclude similar symbols, a:all set groups")
	)

	flag.Parse()

	err := execute(*pCount, *pLen, *pGroups, *pFlags)
	if err != nil {
		log.Fatal(err)
	}
}

func execute(count int, length int, groups string, flags string) error {
	if length < 0 {
		return errors.New("length is below the zero")
	}
	var p genpas.Params
	err := parseGroups(groups, &p)
	if err != nil {
		return err
	}
	err = parseFlags(flags, &p)
	if err != nil {
		return err
	}

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

func parseFlags(flags string, p *genpas.Params) error {
	for _, r := range flags {
		switch r {
		case 'e':
			p.ExcludeSimilar = true
		case 'a':
			p.AllSetGroups = true
		default:
			return errors.New("invalid flag type")
		}
	}
	return nil
}
