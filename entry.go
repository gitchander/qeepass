package qeepass

import (
	"strings"
)

type Entry struct {
	// Group    string
	Title    string
	Username string
	Password string
	URL      string
	Notes    string
}

type Group struct {
	Name   string
	Entrys []*Entry
	Childs []*Group
}

type Record struct {
	Group string // "root/one/two"
	//	Title    string
	//	Username string
	//	Password string
	//	URL      string
	//	Notes    string

	Entry
}

func AppendRecord(g *Group, r Record) {

	names := strings.Split(r.Group, "/")
	for _, name := range names {
		var cg *Group
		for _, child := range g.Childs {
			if child.Name == name {
				cg = child
				break
			}
		}
		if cg == nil {
			cg = &Group{Name: name}
			g.Childs = append(g.Childs, cg)
		}
		g = cg
	}

	var e Entry = r.Entry
	g.Entrys = append(g.Entrys, &e)
}

func NewGroupFromRecords(rs []Record) *Group {
	var g = new(Group)
	for _, r := range rs {
		AppendRecord(g, r)
	}
	return g
}
