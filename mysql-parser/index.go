package mysql_parser

import (
	"fmt"
	"strings"

	"github.com/pingcap/parser/ast"
	"github.com/sunary/sqlize/mysql-templates"
	"github.com/sunary/sqlize/utils"
)

type Index struct {
	Node
	Typ     ast.IndexKeyType
	Columns []string
}

func (i Index) migrationUp(tbName string) []string {
	switch i.Action {
	case MigrateNoAction:
		return nil

	case MigrateAddAction:
		switch i.Typ {
		case ast.IndexKeyTypeNone:
			return []string{fmt.Sprintf(mysql_templates.CreateIndexStm(isLower), utils.EscapeSqlName(i.Name), utils.EscapeSqlName(tbName), strings.Join(utils.EscapeSqlNames(i.Columns), ", "))}

		case ast.IndexKeyTypeUnique:
			return []string{fmt.Sprintf(mysql_templates.CreateUniqueIndexStm(isLower), utils.EscapeSqlName(i.Name), utils.EscapeSqlName(tbName), strings.Join(utils.EscapeSqlNames(i.Columns), ", "))}

		default:
			return nil
		}

	case MigrateRemoveAction:
		return []string{fmt.Sprintf(mysql_templates.DropIndexStm(isLower), utils.EscapeSqlName(i.Name), utils.EscapeSqlName(tbName))}

	case MigrateModifyAction:
		strRems := make([]string, 2)
		i.Action = MigrateRemoveAction
		strRems[0] = i.migrationUp(tbName)[0]
		i.Action = MigrateAddAction
		strRems[1] = i.migrationUp(tbName)[0]
		return strRems

	default:
		return nil
	}
}

func (i Index) migrationDown(tbName string) []string {
	switch i.Action {
	case MigrateNoAction:
		return nil

	case MigrateAddAction:
		i.Action = MigrateRemoveAction
		break

	case MigrateRemoveAction:
		i.Action = MigrateAddAction
		break

	case MigrateModifyAction:
		break

	default:
		return nil
	}

	return i.migrationUp(tbName)
}
