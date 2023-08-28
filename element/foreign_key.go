package element

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/sunary/sqlize/utils"
)

// ForeignKey ...
type ForeignKey struct {
	Node
	Table      string
	Column     string
	RefTable   string
	RefColumn  string
	Constraint string
}

func (fk ForeignKey) hashValue() string {
	strHash := strings.Join(fk.migrationUp(fk.Name), ";")
	hash := md5.Sum([]byte(strHash))
	return hex.EncodeToString(hash[:])
}

func (fk ForeignKey) migrationUp(tbName string) []string {
	switch fk.Action {
	case MigrateNoAction:
		return nil

	case MigrateAddAction:
		return []string{fmt.Sprintf(sql.CreateForeignKeyStm(),
			utils.EscapeSqlName(sql.GetDialect(), tbName), utils.EscapeSqlName(sql.GetDialect(), fk.Name), utils.EscapeSqlName(sql.GetDialect(), fk.Column),
			utils.EscapeSqlName(sql.GetDialect(), fk.RefTable), utils.EscapeSqlName(sql.GetDialect(), fk.RefColumn))}

	case MigrateRemoveAction:
		return []string{fmt.Sprintf(sql.DropForeignKeyStm(),
			utils.EscapeSqlName(sql.GetDialect(), tbName), utils.EscapeSqlName(sql.GetDialect(), fk.Name))}

	case MigrateModifyAction:
		return nil

	case MigrateRenameAction:
		return nil

	default:
		return nil
	}
}

func (fk ForeignKey) migrationDown(tbName string) []string {
	switch fk.Action {
	case MigrateNoAction:
		return nil

	case MigrateAddAction:
		fk.Action = MigrateRemoveAction

	case MigrateRemoveAction:
		fk.Action = MigrateAddAction

	case MigrateModifyAction:

	case MigrateRenameAction:

	default:
		return nil
	}

	return fk.migrationUp(tbName)
}
