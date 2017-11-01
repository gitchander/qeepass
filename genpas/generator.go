package genpas

import (
	"errors"
)

const (
	upperLetters   = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	lowerLetters   = "abcdefghijklmnopqrstuvwxyz"
	digits         = "0123456789"
	specialSymbols = "!@#$%^&*?"
	similarSymbols = "0OolIi1"
)

type Params struct {
	Upper   bool // [A..Z]
	Lower   bool // [a..z]
	Digits  bool // [0..9]
	Special bool // Special Symbols

	ExcludeSimilar bool // Exclude similar symbols: (0,O,o) and (l,I,i,1)
	HasEveryGroup  bool // AllSetGroups - All marked groups
}

var DefaultParams = Params{
	Upper:   true,
	Lower:   true,
	Digits:  true,
	Special: true,

	ExcludeSimilar: true,
	HasEveryGroup:  true,
}

// Password Generator
type Generator struct {
	generate func([]rune)
}

func NewGenerator(p Params, r *Random) (*Generator, error) {

	var groups = makeGroups(p)
	if len(groups) == 0 {
		return nil, errors.New("has no groups")
	}

	var table []rune
	for _, group := range groups {
		for _, r := range group {
			table = append(table, r)
		}
	}

	g := new(Generator)

	if p.HasEveryGroup {
		g.generate = func(rs []rune) {
			if len(rs) < len(groups) {
				panic("insufficient password length")
			}
			for i, group := range groups {
				rs[i] = group[r.Intn(len(group))]
			}
			for i := len(groups); i < len(rs); i++ {
				rs[i] = table[r.Intn(len(table))]
			}
			for i := len(rs) - 1; i > 0; i-- {
				if j := r.Intn(i + 1); i != j {
					rs[i], rs[j] = rs[j], rs[i]
				}
			}
		}
	} else {
		g.generate = func(rs []rune) {
			for i := range rs {
				rs[i] = table[r.Intn(len(table))]
			}
		}
	}

	return g, nil
}

func (g *Generator) Generate(n int) string {
	rs := make([]rune, n)
	g.generate(rs)
	return string(rs)
}

func makeGroups(p Params) (groups [][]rune) {

	prepare := func(s string) []rune {
		return []rune(s)
	}
	if p.ExcludeSimilar {
		// prepare not similar
		m := make(map[rune]struct{})
		for _, r := range similarSymbols {
			m[r] = struct{}{}
		}
		prepare = func(s string) (rs []rune) {
			for _, r := range s {
				if _, ok := m[r]; !ok {
					rs = append(rs, r)
				}
			}
			return
		}
	}

	if p.Upper {
		rs := prepare(upperLetters)
		groups = append(groups, rs)
	}
	if p.Lower {
		rs := prepare(lowerLetters)
		groups = append(groups, rs)
	}
	if p.Digits {
		rs := prepare(digits)
		groups = append(groups, rs)
	}
	if p.Special {
		rs := prepare(specialSymbols)
		groups = append(groups, rs)
	}

	return groups
}
