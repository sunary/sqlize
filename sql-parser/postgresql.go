package sql_parser

import (
	"github.com/auxten/postgresql-parser/pkg/sql/sem/tree"
	"github.com/auxten/postgresql-parser/pkg/walk"
	"github.com/pingcap/parser/ast"
	"github.com/sunary/sqlize/element"
)

func (p *Parser) ParserPostgresql(sql string) error {
	w := &walk.AstWalker{
		Fn: p.walker,
	}
	_, err := w.Walk(sql, nil)
	return err
}

func (p *Parser) walker(ctx interface{}, node interface{}) (stop bool) {
	switch n := node.(type) {
	case *tree.RenameTable:
		p.Migration.RenameTable(n.Name.String(), n.NewName.String())

	case *tree.CreateTable:
		tbName := n.Table.Table()

		tb := element.NewTable(tbName)
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

	case *tree.UniqueConstraintTableDef:
		p.Migration.AddIndex("", postgresUnique(n))

	case *tree.IndexTableDef:
		cols := make([]string, len(n.Columns))
		for i := range n.Columns {
			cols[i] = n.Columns[i].Column.String()
		}

		p.Migration.AddIndex("", element.Index{
			Node: element.Node{
				Name:   n.Name.String(),
				Action: element.MigrateAddAction,
			},
			Typ:     ast.IndexKeyTypeNone,
			Columns: cols,
		})

	case *tree.DropIndex:
		p.Migration.RemoveIndex("", n.String())

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
				Node:  element.Node{Name: nc.Column.String(), Action: element.MigrateModifyAction},
				PgTyp: &nc.ToType.InternalType,
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
				// TODO

			}
		case *tree.AlterTableDropConstraint:
			p.Migration.RemoveIndex(n.Table.String(), nc.Constraint.String())
		}
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
	if n.PrimaryKey.IsPrimaryKey {
		indexes = append(indexes, element.Index{
			Node:    element.Node{Name: "primary_key", Action: element.MigrateAddAction},
			Typ:     ast.IndexKeyTypeNone,
			CnsTyp:  ast.ConstraintPrimaryKey,
			Columns: []string{n.Name.String()},
		})
	} else if n.Unique {
		indexes = append(indexes, element.Index{
			Node: element.Node{
				Name:   n.UniqueConstraintName.String(),
				Action: element.MigrateAddAction,
			},
			Typ:     ast.IndexKeyTypeUnique,
			Columns: []string{n.Name.String()},
		})
	}

	return element.Column{
		Node:    element.Node{Name: n.Name.String(), Action: element.MigrateAddAction},
		PgTyp:   &n.Type.InternalType,
		Options: opts,
	}, indexes
}

func postgresUnique(n *tree.UniqueConstraintTableDef) element.Index {
	cols := make([]string, len(n.Columns))
	for i := range n.Columns {
		cols[i] = n.Columns[i].Column.String()
	}

	var index element.Index
	if n.PrimaryKey {
		index = element.Index{
			Node:    element.Node{Name: "primary_key", Action: element.MigrateAddAction},
			Typ:     ast.IndexKeyTypeNone,
			CnsTyp:  ast.ConstraintPrimaryKey,
			Columns: cols,
		}
	} else {
		index = element.Index{
			Node: element.Node{
				Name:   n.Name.String(),
				Action: element.MigrateAddAction,
			},
			Typ:     ast.IndexKeyTypeUnique,
			Columns: cols,
		}
	}
	return index
}
