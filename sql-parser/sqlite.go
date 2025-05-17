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

	_, err = sqlite.Walk(p, node)
	return err
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
func (p *Parser) Visit(node sqlite.Node) (w sqlite.Visitor, n sqlite.Node, err error) {
	switch n := node.(type) {
	case *sqlite.CreateTableStatement:
		tbName := n.Name.String()
		tb := element.NewTableWithAction(tbName, element.MigrateAddAction)
		p.Migration.AddTable(*tb)
		p.Migration.Using(tbName)

		// TODO: rqlite/sql doesn't support parse constraints
		for i := range n.Columns {
			col := element.Column{
				Node: element.Node{
					Name:   n.Columns[i].Name.String(),
					Action: element.MigrateAddAction,
				},
				CurrentAttr: element.SqlAttr{
					LiteType: n.Columns[i].Type,
					Options:  p.parseSqliteConstrains(tbName, n.Columns[i]),
				},
			}

			p.Migration.AddColumn(tbName, col)
		}

		for _, cons := range n.Constraints {
			switch cons := cons.(type) {
			case *sqlite.UniqueConstraint:
				indexCol := make([]string, len(cons.Columns))
				for i := range cons.Columns {
					indexCol[i] = cons.Columns[i].Collation.Name
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
				CurrentAttr: element.SqlAttr{
					LiteType: n.ColumnDef.Type,
					Options:  p.parseSqliteConstrains(tbName, n.ColumnDef),
				},
			})
		}
	}

	return nil, nil, nil
}

func (p *Parser) parseSqliteConstrains(tbName string, columnDefinition *sqlite.ColumnDefinition) []*ast.ColumnOption {
	// https://www.sqlite.org/syntax/column-constraint.html
	conss := columnDefinition.Constraints
	opts := []*ast.ColumnOption{}
	for _, cons := range conss {
		switch cons := cons.(type) {
		case *sqlite.PrimaryKeyConstraint:
			opts = append(opts, &ast.ColumnOption{Tp: ast.ColumnOptionPrimaryKey})
			if cons.Autoincrement.IsValid() {
				opts = append(opts, &ast.ColumnOption{Tp: ast.ColumnOptionAutoIncrement})
			}

		case *sqlite.NotNullConstraint:
			opts = append(opts, &ast.ColumnOption{Tp: ast.ColumnOptionNotNull})

		case *sqlite.UniqueConstraint:
			opts = append(opts, &ast.ColumnOption{Tp: ast.ColumnOptionUniqKey})

		case *sqlite.CheckConstraint:
			opts = append(opts, &ast.ColumnOption{
				Tp:       ast.ColumnOptionCheck,
				StrValue: cons.Expr.String(),
			})

		case *sqlite.DefaultConstraint:
			opts = append(opts, &ast.ColumnOption{
				Tp:       ast.ColumnOptionDefaultValue,
				StrValue: cons.Expr.String(),
			})
		}
	}

	return opts
}

func (p Parser) VisitEnd(node sqlite.Node) (sqlite.Node, error) {
	return nil, nil
}
