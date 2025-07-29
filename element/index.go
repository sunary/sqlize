package element

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/pingcap/tidb/pkg/parser/ast"
)

// Index ...
type Index struct {
	Node
	Typ       ast.IndexKeyType
	IndexType ast.IndexType
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
				sql.EscapeSqlName(tbName),
				strings.Join(sql.EscapeSqlNames(i.Columns), ", "))}
		}

		switch i.Typ {
		case ast.IndexKeyTypeNone:
			return []string{fmt.Sprintf(sql.CreateIndexStm(i.IndexType.String()),
				sql.EscapeSqlName(i.Name), sql.EscapeSqlName(tbName),
				strings.Join(sql.EscapeSqlNames(i.Columns), ", "))}

		case ast.IndexKeyTypeUnique:
			return []string{fmt.Sprintf(sql.CreateUniqueIndexStm(i.IndexType.String()),
				sql.EscapeSqlName(i.Name), sql.EscapeSqlName(tbName),
				strings.Join(sql.EscapeSqlNames(i.Columns), ", "))}

		default:
			return nil
		}

	case MigrateRemoveAction:
		if i.CnsTyp == ast.ConstraintPrimaryKey {
			return []string{fmt.Sprintf(sql.DropPrimaryKeyStm(),
				sql.EscapeSqlName(tbName))}
		}

		if sql.IsSqlite() {
			return []string{fmt.Sprintf(sql.DropIndexStm(),
				sql.EscapeSqlName(i.Name))}
		}

		return []string{fmt.Sprintf(sql.DropIndexStm(),
			sql.EscapeSqlName(i.Name),
			sql.EscapeSqlName(tbName))}

	case MigrateModifyAction:
		strRems := make([]string, 2)
		i.Action = MigrateRemoveAction
		strRems[0] = i.migrationUp(tbName)[0]
		i.Action = MigrateAddAction
		strRems[1] = i.migrationUp(tbName)[0]
		return strRems

	case MigrateRenameAction:
		return []string{fmt.Sprintf(sql.AlterTableRenameIndexStm(),
			sql.EscapeSqlName(tbName),
			sql.EscapeSqlName(i.OldName),
			sql.EscapeSqlName(i.Name))}

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
