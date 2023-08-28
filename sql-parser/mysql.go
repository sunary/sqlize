package sql_parser

import (
	"github.com/pingcap/parser"
	"github.com/pingcap/parser/ast"
	"github.com/pingcap/parser/model"
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

			case ast.AlterTableAddConstraint:
				switch alter.Specs[i].Constraint.Tp {
				case ast.ConstraintPrimaryKey:
					cols := make([]string, len(alter.Specs[i].Constraint.Keys))
					for j, key := range alter.Specs[i].Constraint.Keys {
						cols[j] = key.Column.Name.O
					}

					if alter.Specs[i].Constraint.Keys != nil {
						p.Migration.AddIndex(alter.Table.Text(), element.Index{
							Node:    element.Node{Name: "primary_key", Action: element.MigrateAddAction},
							Typ:     ast.IndexKeyTypeNone,
							CnsTyp:  ast.ConstraintPrimaryKey,
							Columns: cols,
						})
					} else {
						p.Migration.AddColumn(alter.Table.Text(), element.Column{
							Node: element.Node{Name: cols[0], Action: element.MigrateAddAction},
							Typ:  nil,
							Options: []*ast.ColumnOption{
								{
									Tp: ast.ColumnOptionPrimaryKey,
								},
							},
						})
					}

				case ast.ConstraintForeignKey:
					cols := make([]string, len(alter.Specs[i].Constraint.Keys))
					for j, key := range alter.Specs[i].Constraint.Keys {
						cols[j] = key.Column.Name.O
					}

					p.Migration.AddForeignKey(alter.Table.Text(), element.ForeignKey{
						Node:       element.Node{Name: alter.Specs[i].Constraint.Name, Action: element.MigrateAddAction},
						Table:      alter.Table.Text(),
						Column:     cols[0],
						RefTable:   alter.Specs[i].Constraint.Refer.Table.Name.String(),
						RefColumn:  alter.Specs[i].Constraint.Refer.IndexPartSpecifications[0].Column.Name.String(),
						Constraint: "",
					})
				}

			case ast.AlterTableDropColumn:
				p.Migration.RemoveColumn(alter.Table.Name.O, alter.Specs[i].OldColumnName.Name.O)

			case ast.AlterTableDropPrimaryKey:

			case ast.AlterTableDropIndex:

			case ast.AlterTableDropForeignKey:
				p.Migration.RemoveForeignKey(alter.Table.Name.O, alter.Specs[i].Constraint.Name)

			case ast.AlterTableModifyColumn:
				if len(alter.Specs[i].NewColumns) > 0 {
					for j := range alter.Specs[i].NewColumns {
						col := element.Column{
							Node:    element.Node{Name: alter.Specs[i].NewColumns[j].Name.Name.O, Action: element.MigrateModifyAction},
							Typ:     alter.Specs[i].NewColumns[j].Tp,
							Comment: alter.Specs[i].Comment,
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
		tb := element.NewTableWithAction(tbName, element.MigrateAddAction)

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
				indexType := model.IndexTypeBtree
				if tab.Constraints[i].Option != nil {
					indexType = tab.Constraints[i].Option.Tp
				}
				tb.AddIndex(element.Index{
					Node: element.Node{
						Name:   tab.Constraints[i].Name,
						Action: element.MigrateAddAction,
					},
					Typ:       ast.IndexKeyTypeNone,
					IndexType: indexType,
					Columns:   cols,
				})

			case ast.ConstraintUniqKey, ast.ConstraintUniqIndex:
				indexType := model.IndexTypeBtree
				if tab.Constraints[i].Option != nil {
					indexType = tab.Constraints[i].Option.Tp
				}
				tb.AddIndex(element.Index{
					Node: element.Node{
						Name:   tab.Constraints[i].Name,
						Action: element.MigrateAddAction,
					},
					Typ:       ast.IndexKeyTypeUnique,
					IndexType: indexType,
					Columns:   cols,
				})
			}
		}

		p.Migration.AddTable(*tb)
	}

	// define Column
	if def, ok := in.(*ast.ColumnDef); ok {
		comment := ""
		for i := range def.Options {
			if def.Options[i].Tp == ast.ColumnOptionComment {
				comment = def.Options[i].StrValue
			}
		}
		column := element.Column{
			Node:    element.Node{Name: def.Name.Name.O, Action: element.MigrateAddAction},
			Typ:     def.Tp,
			Options: def.Options,
			Comment: comment,
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
