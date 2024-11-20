package element

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strings"

	ptypes "github.com/auxten/postgresql-parser/pkg/sql/types"
	"github.com/pingcap/parser/ast"
	"github.com/pingcap/parser/format"
	"github.com/pingcap/parser/types"
	sqlite "github.com/rqlite/sql"
	sql_templates "github.com/sunary/sqlize/sql-templates"
)

const (
	// UppercaseRestoreFlag ...
	UppercaseRestoreFlag = format.RestoreStringSingleQuotes | format.RestoreKeyWordUppercase | format.RestoreNameUppercase | format.RestoreNameBackQuotes
	// LowerRestoreFlag ...
	LowerRestoreFlag = format.RestoreStringSingleQuotes | format.RestoreKeyWordLowercase | format.RestoreNameLowercase | format.RestoreNameBackQuotes
)

type SqlAttr struct {
	MysqlType *types.FieldType
	PgType    *ptypes.T
	LiteType  *sqlite.Type
	Options   []*ast.ColumnOption
	Comment   string
}

// Column ...
type Column struct {
	Node

	CurrentAttr  SqlAttr
	PreviousAttr SqlAttr
}

// GetType ...
func (c Column) GetType() byte {
	if c.CurrentAttr.MysqlType != nil {
		return c.CurrentAttr.MysqlType.Tp
	}

	return 0
}

// HasDefaultValue ...
func (c Column) HasDefaultValue() bool {
	for _, opt := range c.CurrentAttr.Options {
		if opt.Tp == ast.ColumnOptionDefaultValue {
			return true
		}
	}

	return false
}

func (c Column) hashValue() string {
	strHash := sql.EscapeSqlName(c.Name)
	strHash += c.typeDefinition(false)
	hash := md5.Sum([]byte(strHash))
	return hex.EncodeToString(hash[:])
}

func (c Column) migrationUp(tbName, after string, ident int) []string {
	switch c.Action {
	case MigrateNoAction:
		return nil

	case MigrateAddAction:
		strSql := sql.EscapeSqlName(c.Name)

		if ident > len(c.Name) {
			strSql += strings.Repeat(" ", ident-len(c.Name))
		}

		strSql += c.definition(false)

		if ident < 0 {
			if after != "" {
				return []string{fmt.Sprintf(sql.AlterTableAddColumnAfterStm(), sql.EscapeSqlName(tbName), strSql, sql.EscapeSqlName(after))}
			}
			return []string{fmt.Sprintf(sql.AlterTableAddColumnFirstStm(), sql.EscapeSqlName(tbName), strSql)}
		}

		return append([]string{strSql}, c.migrationCommentUp(tbName)...)

	case MigrateRemoveAction:
		if sql.IsSqlite() {
			return nil
		}

		return []string{fmt.Sprintf(sql.AlterTableDropColumnStm(), sql.EscapeSqlName(tbName), sql.EscapeSqlName(c.Name))}

	case MigrateModifyAction:
		def, isPk := c.pkDefinition(false)
		if isPk {
			if _, isPrevPk := c.pkDefinition(true); isPrevPk {
				// avoid repeat define primary key
				def = strings.Replace(def, sql.PrimaryOption(), "", 1)
			}
		}

		return []string{fmt.Sprintf(sql.AlterTableModifyColumnStm(), sql.EscapeSqlName(tbName), sql.EscapeSqlName(c.Name)+def)}

	case MigrateRevertAction:
		prevDef, isPrevPk := c.pkDefinition(true)
		if isPrevPk {
			if _, isPk := c.pkDefinition(false); isPk {
				// avoid repeat define primary key
				prevDef = strings.Replace(prevDef, sql.PrimaryOption(), "", 1)
			}
		}

		return []string{fmt.Sprintf(sql.AlterTableModifyColumnStm(), sql.EscapeSqlName(tbName), sql.EscapeSqlName(c.Name)+prevDef)}

	case MigrateRenameAction:
		return []string{fmt.Sprintf(sql.AlterTableRenameColumnStm(), sql.EscapeSqlName(tbName), sql.EscapeSqlName(c.OldName), sql.EscapeSqlName(c.Name))}

	default:
		return nil
	}
}

func (c Column) migrationCommentUp(tbName string) []string {
	if c.CurrentAttr.Comment == "" || sql.GetDialect() != sql_templates.PostgresDialect {
		return nil
	}

	// apply for postgres only
	return []string{fmt.Sprintf(sql.ColumnComment(), tbName, c.Name, c.CurrentAttr.Comment)}
}

func (c Column) migrationDown(tbName, after string) []string {
	switch c.Action {
	case MigrateNoAction:
		return nil

	case MigrateAddAction:
		c.Action = MigrateRemoveAction

	case MigrateRemoveAction:
		c.Action = MigrateAddAction

	case MigrateModifyAction:
		c.Action = MigrateRevertAction

	case MigrateRenameAction:
		c.Name, c.OldName = c.OldName, c.Name

	default:
		return nil
	}

	return c.migrationUp(tbName, after, -1)
}

func (c Column) pkDefinition(isPrev bool) (string, bool) {
	attr := c.CurrentAttr
	if isPrev {
		attr = c.PreviousAttr
	}
	strSql := c.typeDefinition(isPrev)

	isPrimaryKey := false
	for _, opt := range attr.Options {
		if opt.Tp == ast.ColumnOptionPrimaryKey {
			isPrimaryKey = true
		}

		b := bytes.NewBufferString("")
		var ctx *format.RestoreCtx

		if sql.IsLowercase() {
			ctx = format.NewRestoreCtx(LowerRestoreFlag, b)
		} else {
			ctx = format.NewRestoreCtx(UppercaseRestoreFlag, b)
		}

		if sql.IsPostgres() && opt.Tp == ast.ColumnOptionDefaultValue {
			strSql += " " + opt.StrValue
			continue
		}

		_ = opt.Restore(ctx)
		strSql += " " + b.String()
	}

	return strSql, isPrimaryKey
}

func (c Column) definition(isPrev bool) string {
	def, _ := c.pkDefinition(isPrev)
	return def
}

func (c Column) typeDefinition(isPrev bool) string {
	attr := c.CurrentAttr
	if isPrev {
		attr = c.PreviousAttr
	}

	switch {
	case sql.IsPostgres() && attr.PgType != nil:
		return " " + attr.PgType.SQLString()
	case sql.IsSqlite() && attr.LiteType != nil:
		return " " + attr.LiteType.Name.Name
	case attr.MysqlType != nil:
		return " " + attr.MysqlType.String()
	}

	return "" // column type is empty
}
