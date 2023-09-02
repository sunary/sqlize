package sql_parser

import (
	"strings"

	"github.com/pingcap/parser/ast"
	sqlite "github.com/rqlite/sql"
	"github.com/sunary/sqlize/element"
)

// ParserSqlite ...
func (p *Parser) ParserSqlite(sql string) error {
	ps := sqlite.NewParser(strings.NewReader(sql))

	node, err := ps.ParseStatement()
	if err != nil {
		return err
	}

	return sqlite.Walk(p, node)
}

/*
Walk with statements:

CreateTableStatement
CreateIndexStatement
DropTableStatement
DropIndexStatement
AlterTableStatement

sqlite does not support drop column
*/
func (p Parser) Visit(node sqlite.Node) (w sqlite.Visitor, err error) {
	switch n := node.(type) {
	case *sqlite.CreateTableStatement:
		tbName := n.Table.String()
		tb := element.NewTableWithAction(tbName, element.MigrateAddAction)
		p.Migration.AddTable(*tb)
		p.Migration.Using(tbName)

		for i := range n.Columns {
			col := element.Column{
				Node: element.Node{
					Name:   n.Columns[i].Name.String(),
					Action: element.MigrateAddAction,
				},
				LiteType: n.Columns[i].Type,
				Options:  p.parseSqliteConstrains(tbName, n.Columns[i].Constraints),
			}

			p.Migration.AddColumn(tbName, col)
		}

		for _, cons := range n.Constraints {
			switch cons := cons.(type) {
			case *sqlite.UniqueConstraint:
				indexCol := make([]string, len(cons.Columns))
				for i := range cons.Columns {
					indexCol[i] = cons.Columns[i].Name
				}

				p.Migration.AddIndex(tbName, element.Index{
					Node: element.Node{
						Name:   cons.Name.Name,
						Action: element.MigrateAddAction,
					},
					Columns: indexCol,
					Typ:     ast.IndexKeyTypeUnique,
				})

			case *sqlite.ForeignKeyConstraint:
				// p.Migration.AddForeignKey(tbName, element.ForeignKey{})
			}
		}

	case *sqlite.CreateIndexStatement:
		tbName := n.Table.String()

		indexCol := make([]string, len(n.Columns))
		for i := range n.Columns {
			indexCol[i] = n.Columns[i].String()
		}

		typ := ast.IndexKeyTypeNone
		if n.Unique.IsValid() {
			typ = ast.IndexKeyTypeUnique
		}

		p.Migration.AddIndex(tbName, element.Index{
			Node: element.Node{
				Name:   n.Name.Name,
				Action: element.MigrateAddAction,
			},
			Columns: indexCol,
			Typ:     typ,
		})

	case *sqlite.DropTableStatement:
		tbName := n.Table.String()
		p.Migration.RemoveTable(tbName)

	case *sqlite.DropIndexStatement:

	case *sqlite.AlterTableStatement:
		tbName := n.Table.String()
		switch {
		case n.Rename.IsValid():
			p.Migration.RenameTable(tbName, n.NewName.Name)

		case n.RenameColumn.IsValid():
			p.Migration.RenameColumn(tbName, n.ColumnName.String(), n.NewColumnName.String())

		case n.AddColumn.IsValid():
			p.Migration.AddColumn(tbName, element.Column{
				Node: element.Node{
					Name:   n.ColumnDef.Name.String(),
					Action: element.MigrateAddAction,
				},
				LiteType: n.ColumnDef.Type,
				Options:  p.parseSqliteConstrains(tbName, n.ColumnDef.Constraints),
			})
		}
	}

	return nil, nil
}

func (p *Parser) parseSqliteConstrains(tbName string, conss []sqlite.Constraint) []*ast.ColumnOption {
	opts := []*ast.ColumnOption{}
	for _, cons := range conss {
		switch cons := cons.(type) {
		case *sqlite.PrimaryKeyConstraint:
			opts = append(opts, &ast.ColumnOption{Tp: ast.ColumnOptionPrimaryKey})

		case *sqlite.NotNullConstraint:
			opts = append(opts, &ast.ColumnOption{Tp: ast.ColumnOptionNotNull})

		case *sqlite.UniqueConstraint:
			indexCol := make([]string, len(cons.Columns))
			for i := range cons.Columns {
				indexCol[i] = cons.Columns[i].Name
			}

			p.Migration.AddIndex(tbName, element.Index{
				Node: element.Node{
					Name:   cons.Name.Name,
					Action: element.MigrateAddAction,
				},
				Columns: indexCol,
				Typ:     ast.IndexKeyTypeUnique,
			})

		case *sqlite.CheckConstraint:

		case *sqlite.DefaultConstraint:
			opts = append(opts, &ast.ColumnOption{
				Tp:       ast.ColumnOptionDefaultValue,
				StrValue: cons.Default.String()})
		}
	}

	return opts
}

func (p Parser) VisitEnd(node sqlite.Node) error {
	return nil
}
