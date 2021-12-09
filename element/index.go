package element

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/pingcap/parser/ast"
	"github.com/pingcap/parser/model"
	"github.com/sunary/sqlize/utils"
)

// Index ...
type Index struct {
	Node
	Typ       ast.IndexKeyType
	IndexType model.IndexType
	CnsTyp    ast.ConstraintType
	Columns   []string
}

func (i Index) hashValue() string {
	strHash := strings.Join(i.migrationUp(""), ";")
	hash := md5.Sum([]byte(strHash))
	return hex.EncodeToString(hash[:])
}

func (i Index) migrationUp(tbName string) []string {
	switch i.Action {
	case MigrateNoAction:
		return nil

	case MigrateAddAction:
		if i.CnsTyp == ast.ConstraintPrimaryKey {
			return []string{fmt.Sprintf(sql.CreatePrimaryKeyStm(),
				utils.EscapeSqlName(sql.IsPostgres, tbName),
				strings.Join(utils.EscapeSqlNames(sql.IsPostgres, i.Columns), ", "))}
		}

		switch i.Typ {
		case ast.IndexKeyTypeNone:
			return []string{fmt.Sprintf(sql.CreateIndexStm(i.IndexType.String()),
				utils.EscapeSqlName(sql.IsPostgres, i.Name), utils.EscapeSqlName(sql.IsPostgres, tbName),
				strings.Join(utils.EscapeSqlNames(sql.IsPostgres, i.Columns), ", "))}

		case ast.IndexKeyTypeUnique:
			return []string{fmt.Sprintf(sql.CreateUniqueIndexStm(i.IndexType.String()),
				utils.EscapeSqlName(sql.IsPostgres, i.Name), utils.EscapeSqlName(sql.IsPostgres, tbName),
				strings.Join(utils.EscapeSqlNames(sql.IsPostgres, i.Columns), ", "))}

		default:
			return nil
		}

	case MigrateRemoveAction:
		if i.CnsTyp == ast.ConstraintPrimaryKey {
			return []string{fmt.Sprintf(sql.DropPrimaryKeyStm(),
				utils.EscapeSqlName(sql.IsPostgres, tbName))}
		}

		return []string{fmt.Sprintf(sql.DropIndexStm(),
			utils.EscapeSqlName(sql.IsPostgres, i.Name),
			utils.EscapeSqlName(sql.IsPostgres, tbName))}

	case MigrateModifyAction:
		strRems := make([]string, 2)
		i.Action = MigrateRemoveAction
		strRems[0] = i.migrationUp(tbName)[0]
		i.Action = MigrateAddAction
		strRems[1] = i.migrationUp(tbName)[0]
		return strRems

	case MigrateRenameAction:
		return []string{fmt.Sprintf(sql.AlterTableRenameIndexStm(),
			utils.EscapeSqlName(sql.IsPostgres, tbName),
			utils.EscapeSqlName(sql.IsPostgres, i.OldName),
			utils.EscapeSqlName(sql.IsPostgres, i.Name))}

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

	case MigrateRemoveAction:
		i.Action = MigrateAddAction

	case MigrateModifyAction:

	case MigrateRenameAction:
		i.Name, i.OldName = i.OldName, i.Name

	default:
		return nil
	}

	return i.migrationUp(tbName)
}
