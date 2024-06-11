package sql_parser

import (
	"strings"

	"github.com/auxten/postgresql-parser/pkg/sql/parser"
	"github.com/auxten/postgresql-parser/pkg/sql/sem/tree"
	"github.com/auxten/postgresql-parser/pkg/walk"
	"github.com/pingcap/parser/ast"
	"github.com/sunary/sqlize/element"
)

// ParserPostgresql ...
func (p *Parser) ParserPostgresql(sql string) error {
	w := &walk.AstWalker{
		Fn: p.walker,
	}

	stmts, err := parser.Parse(sql)
	if err != nil {
		return err
	}

	_, err = w.Walk(stmts, nil)
	return err
}

/*
Walk with statements:

CreateTable
ColumnTableDef
CreateIndex
DropIndex
AlterTable
RenameTable
*/
func (p *Parser) walker(ctx interface{}, node interface{}) (stop bool) {
	switch n := node.(type) {
	case *tree.CreateTable:
		tbName := n.Table.Table()

		tb := element.NewTableWithAction(tbName, element.MigrateAddAction)
		p.Migration.AddTable(*tb)
		p.Migration.Using(tbName)

	case *tree.ColumnTableDef:
		col, indexes := postgresColumn(n)
		p.Migration.AddColumn("", col)
		if len(indexes) > 0 {
			for i := range indexes {
				p.Migration.AddIndex("", indexes[i])
			}
		}

	case *tree.CommentOnColumn:
		p.Migration.AddComment(n.TableName.String(), n.Column(), *n.Comment)

	case *tree.CreateIndex:
		p.Migration.AddIndex("", postgresIndex(n))

	case *tree.DropIndex:
		for i := range n.IndexList {
			p.Migration.RemoveIndex("", n.IndexList[i].Index.String())
		}

	case *tree.AlterTable:
		switch nc := n.Cmds[0].(type) {
		case *tree.AlterTableRenameTable:
			p.Migration.RenameTable(n.Table.String(), nc.NewName.TableName.String())

		case *tree.AlterTableRenameColumn:
			p.Migration.RenameColumn(n.Table.String(), nc.Column.String(), nc.NewName.String())

		case *tree.AlterTableRenameConstraint:
			p.Migration.RenameIndex(n.Table.String(), nc.Constraint.String(), nc.NewName.String())

		case *tree.AlterTableAddColumn:
			col, indexes := postgresColumn(nc.ColumnDef)
			p.Migration.AddColumn(n.Table.String(), col)
			if len(indexes) > 0 {
				for i := range indexes {
					p.Migration.AddIndex(n.Table.String(), indexes[i])
				}
			}

		case *tree.AlterTableDropColumn:
			p.Migration.RemoveColumn(n.Table.String(), nc.Column.String())

		case *tree.AlterTableDropNotNull:
			p.Migration.RemoveColumn(n.Table.String(), nc.Column.String())

		case *tree.AlterTableAlterColumnType:
			col := element.Column{
				Node:   element.Node{Name: nc.Column.String(), Action: element.MigrateModifyAction},
				PgType: nc.ToType,
			}
			p.Migration.AddColumn(n.Table.String(), col)

		case *tree.AlterTableSetDefault:
			if nc.Default != nil {
				col := element.Column{
					Node: element.Node{Name: nc.Column.String(), Action: element.MigrateModifyAction},
					Options: []*ast.ColumnOption{{
						Expr:     nil,
						Tp:       ast.ColumnOptionDefaultValue,
						StrValue: nc.Default.String(),
					}},
				}
				p.Migration.AddColumn(n.Table.String(), col)
			}

		case *tree.AlterTableAddConstraint:
			switch nc2 := nc.ConstraintDef.(type) {
			case *tree.UniqueConstraintTableDef:
				p.Migration.AddIndex(n.Table.String(), postgresUnique(nc2))

			case *tree.ForeignKeyConstraintTableDef:
				p.Migration.AddForeignKey(n.Table.String(), postgresForeignKey(nc2))
			}

		case *tree.AlterTableDropConstraint:
			consName := nc.Constraint.String()
			if strings.HasPrefix(strings.ToLower(consName), "fk") { // detect ForeignKey Constraint
				p.Migration.RemoveForeignKey(n.Table.String(), consName)
			} else {
				p.Migration.RemoveIndex(n.Table.String(), consName)
			}
		}

	case *tree.RenameTable:
		p.Migration.RenameTable(n.Name.String(), n.NewName.String())
	}

	return false
}

func postgresColumn(n *tree.ColumnTableDef) (element.Column, []element.Index) {
	opts := make([]*ast.ColumnOption, 0, 1)
	if n.DefaultExpr.Expr != nil {
		opts = append(opts, &ast.ColumnOption{
			Tp:       ast.ColumnOptionDefaultValue,
			StrValue: n.DefaultExpr.Expr.String(),
		})
	}

	indexes := []element.Index{}
	switch {
	case n.PrimaryKey.IsPrimaryKey:
		indexes = append(indexes, element.Index{
			Node:    element.Node{Name: "primary_key", Action: element.MigrateAddAction},
			Typ:     ast.IndexKeyTypeNone,
			CnsTyp:  ast.ConstraintPrimaryKey,
			Columns: []string{n.Name.String()},
		})
	case n.Unique:
		indexes = append(indexes, element.Index{
			Node: element.Node{
				Name:   n.UniqueConstraintName.String(),
				Action: element.MigrateAddAction,
			},
			Typ:     ast.IndexKeyTypeUnique,
			Columns: []string{n.Name.String()},
		})
	case n.References.Table != nil:
	}

	return element.Column{
		Node:    element.Node{Name: n.Name.String(), Action: element.MigrateAddAction},
		PgType:  n.Type,
		Options: opts,
	}, indexes
}

func postgresUnique(n *tree.UniqueConstraintTableDef) element.Index {
	cols := make([]string, len(n.Columns))
	for i := range n.Columns {
		cols[i] = n.Columns[i].Column.String()
	}

	if n.PrimaryKey {
		return element.Index{
			Node:    element.Node{Name: "primary_key", Action: element.MigrateAddAction},
			Typ:     ast.IndexKeyTypeNone,
			CnsTyp:  ast.ConstraintPrimaryKey,
			Columns: cols,
		}
	}

	return element.Index{
		Node: element.Node{
			Name:   n.Name.String(),
			Action: element.MigrateAddAction,
		},
		Typ:     ast.IndexKeyTypeUnique,
		Columns: cols,
	}
}

func postgresIndex(n *tree.CreateIndex) element.Index {
	cols := make([]string, len(n.Columns))
	for i := range n.Columns {
		cols[i] = n.Columns[i].Column.String()
	}

	typ := ast.IndexKeyTypeNone
	if n.Unique {
		typ = ast.IndexKeyTypeUnique
	}

	return element.Index{
		Node: element.Node{
			Name:   n.Name.String(),
			Action: element.MigrateAddAction,
		},
		Typ:     typ,
		Columns: cols,
	}
}

func postgresForeignKey(n *tree.ForeignKeyConstraintTableDef) element.ForeignKey {
	return element.ForeignKey{
		Node:       element.Node{Name: n.Name.String(), Action: element.MigrateAddAction},
		Table:      "",
		Column:     n.FromCols[0].String(),
		RefTable:   n.Table.Table(),
		RefColumn:  n.ToCols[0].String(),
		Constraint: "",
	}
}
