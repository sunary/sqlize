package mermaidjs

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/sunary/sqlize/element"
)

const (
	erdTag              = "erDiagram"
	liveUrl             = "https://mermaid.ink/img/"
	defaultRelationType = "}o--||" // n-1
)

type MermaidJs struct {
	entities  []string
	relations []string
}

func NewMermaidJs(tables []element.Table) *MermaidJs {
	mm := &MermaidJs{}
	for i := range tables {
		mm.AddTable(tables[i])
	}

	return mm
}

func (m *MermaidJs) AddTable(table element.Table) {
	normEntityName := func(s string) string {
		return strings.ToUpper(s)
	}

	tab := " "
	ent := []string{tab + normEntityName(table.Name) + " {"}

	tab = "  "
	for _, c := range table.Columns {
		dataType := c.DataType()
		if strings.HasPrefix(strings.ToLower(dataType), "enum") {
			dataType = dataType[:4]
		}

		constraint := c.Constraint()

		cmt := c.CurrentAttr.Comment
		if cmt != "" {
			cmt = "\"" + cmt + "\""
		}

		ent = append(ent, fmt.Sprintf("%s%s %s %s %s", tab, dataType, c.Name, constraint, cmt))
	}

	tab = " "
	ent = append(ent, tab+"}")
	m.entities = append(m.entities, strings.Join(ent, "\n"))

	uniRel := map[string]bool{}
	for _, rel := range table.ForeignKeys {
		if _, ok := uniRel[rel.RefTable+"-"+rel.Table]; ok {
			continue
		}

		relType := defaultRelationType
		m.relations = append(m.relations, fmt.Sprintf("%s%s %s %s: %s", tab, normEntityName(rel.Table), relType, normEntityName(rel.RefTable), rel.Column))
		uniRel[rel.Table+"-"+rel.RefTable] = true
	}
}

func (m MermaidJs) String() string {
	return erdTag + "\n" + strings.Join(m.entities, "\n") + "\n" + strings.Join(m.relations, "\n")
}

func (m MermaidJs) Live() string {
	mmParam := base64.URLEncoding.EncodeToString([]byte(m.String()))
	return liveUrl + string(mmParam)
}
