package sql_parser

import (
	"github.com/pingcap/parser"
	"github.com/pingcap/parser/ast"
	"github.com/sunary/sqlize/element"
)

// ParserMysql ...
func (p *Parser) ParserMysql(sql string) error {
	ps := parser.New()
	stmtNodes, _, err := ps.Parse(sql, "", "")
	if err != nil {
		return err
	}

	for _, n := range stmtNodes {
		switch n.(type) {
		case ast.DDLNode:
			n.Accept(p)
			break
		}
	}

	return nil
}

// Enter ...
func (p *Parser) Enter(in ast.Node) (ast.Node, bool) {
	// get Table name
	if tb, ok := in.(*ast.TableName); ok {
		p.Migration.Using(tb.Name.O)
	}

	// drop Table
	if tb, ok := in.(*ast.DropTableStmt); ok {
		for i := range tb.Tables {
			p.Migration.RemoveTable(tb.Tables[i].Name.O)
		}
	}

	// alter Table
	if alter, ok := in.(*ast.AlterTableStmt); ok {
		for i := range alter.Specs {
			switch alter.Specs[i].Tp {
			case ast.AlterTableAddColumns:
				if alter.Specs[i].Position != nil {
					p.Migration.SetColumnPosition("", alter.Specs[i].Position)
				}

			case ast.AlterTableDropColumn:
				p.Migration.RemoveColumn(alter.Table.Name.O, alter.Specs[i].OldColumnName.Name.O)

			case ast.AlterTableModifyColumn:
				if len(alter.Specs[i].NewColumns) > 0 {
					for j := range alter.Specs[i].NewColumns {
						col := element.Column{
							Node: element.Node{Name: alter.Specs[i].NewColumns[j].Name.Name.O, Action: element.MigrateModifyAction},
							Typ:  alter.Specs[i].NewColumns[j].Tp,
						}
						p.Migration.AddColumn(alter.Table.Name.O, col)
					}
				}

			case ast.AlterTableRenameColumn:
				p.Migration.RenameColumn(alter.Table.Name.O, alter.Specs[i].OldColumnName.Name.O, alter.Specs[i].NewColumnName.Name.O)

			case ast.AlterTableRenameTable:
				p.Migration.RenameTable(alter.Table.Name.O, alter.Specs[i].NewTable.Name.O)

			case ast.AlterTableRenameIndex:
				p.Migration.RenameIndex(alter.Table.Name.O, alter.Specs[i].FromKey.O, alter.Specs[i].ToKey.O)
			}
		}
	}

	// drop Index
	if idx, ok := in.(*ast.DropIndexStmt); ok {
		p.Migration.RemoveIndex(idx.Table.Name.O, idx.IndexName)
	}

	// create Table
	if tab, ok := in.(*ast.CreateTableStmt); ok {
		tbName := tab.Table.Name.O
		tb := element.NewTable(tbName)

		p.Migration.Using(tbName)
		for i := range tab.Constraints {
			cols := make([]string, len(tab.Constraints[i].Keys))
			for j, key := range tab.Constraints[i].Keys {
				cols[j] = key.Column.Name.O
			}

			switch tab.Constraints[i].Tp {
			case ast.ConstraintPrimaryKey:
				if tab.Constraints[i].Keys != nil {
					tb.AddIndex(element.Index{
						Node:    element.Node{Name: "primary_key", Action: element.MigrateAddAction},
						Typ:     ast.IndexKeyTypeNone,
						CnsTyp:  ast.ConstraintPrimaryKey,
						Columns: cols,
					})
				} else {
					tb.AddColumn(element.Column{
						Node: element.Node{Name: cols[0], Action: element.MigrateAddAction},
						Typ:  nil,
						Options: []*ast.ColumnOption{
							{
								Tp: ast.ColumnOptionPrimaryKey,
							},
						},
					})
				}

			case ast.ConstraintKey, ast.ConstraintIndex:
				tb.AddIndex(element.Index{
					Node: element.Node{
						Name:   tab.Constraints[i].Name,
						Action: element.MigrateAddAction,
					},
					Typ:       ast.IndexKeyTypeNone,
					IndexType: tab.Constraints[i].Option.Tp,
					Columns:   cols,
				})

			case ast.ConstraintUniqKey, ast.ConstraintUniqIndex:
				tb.AddIndex(element.Index{
					Node: element.Node{
						Name:   tab.Constraints[i].Name,
						Action: element.MigrateAddAction,
					},
					Typ:       ast.IndexKeyTypeUnique,
					IndexType: tab.Constraints[i].Option.Tp,
					Columns:   cols,
				})
			}
		}

		p.Migration.AddTable(*tb)
	}

	// define Column
	if def, ok := in.(*ast.ColumnDef); ok {
		column := element.Column{
			Node:    element.Node{Name: def.Name.Name.O, Action: element.MigrateAddAction},
			Typ:     def.Tp,
			Options: def.Options,
		}
		p.Migration.AddColumn("", column)
	}

	// create Index
	if idx, ok := in.(*ast.CreateIndexStmt); ok {
		cols := make([]string, len(idx.IndexPartSpecifications))
		for i := range idx.IndexPartSpecifications {
			cols[i] = idx.IndexPartSpecifications[i].Column.Name.O
		}

		p.Migration.AddIndex(idx.Table.Name.O, element.Index{
			Node: element.Node{
				Name:   idx.IndexName,
				Action: element.MigrateAddAction,
			},
			Typ:     idx.KeyType,
			Columns: cols,
		})
	}

	return in, false
}

// Leave ...
func (p *Parser) Leave(in ast.Node) (ast.Node, bool) {
	return in, true
}
