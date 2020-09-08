package mysql_parser

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/pingcap/parser/ast"
	"github.com/pingcap/parser/format"
	"github.com/pingcap/parser/types"
	"github.com/sunary/sqlize/mysql-templates"
	"github.com/sunary/sqlize/utils"
)

const (
	UppercaseRestoreFlag = format.RestoreStringSingleQuotes | format.RestoreKeyWordUppercase | format.RestoreNameUppercase | format.RestoreNameBackQuotes
	LowerRestoreFlag     = format.RestoreStringSingleQuotes | format.RestoreKeyWordLowercase | format.RestoreNameLowercase | format.RestoreNameBackQuotes
)

type Column struct {
	Node
	Typ     *types.FieldType
	Options []*ast.ColumnOption
}

func (c Column) HasDefaultValue() bool {
	for _, opt := range c.Options {
		if opt.Tp == ast.ColumnOptionDefaultValue {
			return true
		}
	}

	return false
}

func (c Column) migrationUp(tbName, after string, ident int) []string {
	switch c.Action {
	case MigrateNoAction:
		return nil

	case MigrateAddAction:
		strSql := utils.EscapeSqlName(c.Name)

		if ident > len(c.Name) {
			strSql += strings.Repeat(" ", ident-len(c.Name))
		}
		if c.Typ != nil {
			strSql += " " + c.Typ.String()
		}

		for _, opt := range c.Options {
			b := bytes.NewBufferString("")
			var ctx *format.RestoreCtx
			if isLower {
				ctx = format.NewRestoreCtx(LowerRestoreFlag, b)
			} else {
				ctx = format.NewRestoreCtx(UppercaseRestoreFlag, b)
			}
			_ = opt.Restore(ctx)
			strSql += " " + b.String()
		}

		if ident < 0 {
			if after != "" {
				return []string{fmt.Sprintf(mysql_templates.AlterTableAddColumnAfterStm(isLower), utils.EscapeSqlName(tbName), strSql, utils.EscapeSqlName(after))}
			} else {
				return []string{fmt.Sprintf(mysql_templates.AlterTableAddColumnFirstStm(isLower), utils.EscapeSqlName(tbName), strSql)}
			}
		}

		return []string{strSql}

	case MigrateRemoveAction:
		return []string{fmt.Sprintf(mysql_templates.AlterTableDropColumnStm(isLower), tbName, c.Name)}

	case MigrateModifyAction:
		return nil

	default:
		return nil
	}
}

func (c Column) migrationDown(tbName, after string) []string {
	switch c.Action {
	case MigrateNoAction:
		return nil

	case MigrateAddAction:
		c.Action = MigrateRemoveAction
		break

	case MigrateRemoveAction:
		c.Action = MigrateAddAction
		break

	case MigrateModifyAction:
		break

	default:
		return nil
	}

	return c.migrationUp(tbName, after, -1)
}
