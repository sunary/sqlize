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

// Column ...
type Column struct {
	Node
	MysqlType *types.FieldType
	PgType    *ptypes.T
	LiteType  *sqlite.Type
	Options   []*ast.ColumnOption
	Comment   string
}

// GetType ...
func (c Column) GetType() byte {
	if c.MysqlType != nil {
		return c.MysqlType.Tp
	}

	return 0
}

// HasDefaultValue ...
func (c Column) HasDefaultValue() bool {
	for _, opt := range c.Options {
		if opt.Tp == ast.ColumnOptionDefaultValue {
			return true
		}
	}

	return false
}

func (c Column) hashValue() string {
	strHash := sql.EscapeSqlName(c.Name)
	strHash += c.typeDefinition()
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

		strSql += c.definition()

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
		return []string{fmt.Sprintf(sql.AlterTableModifyColumnStm(), sql.EscapeSqlName(tbName), sql.EscapeSqlName(c.Name)+c.definition())}

	case MigrateRenameAction:
		return []string{fmt.Sprintf(sql.AlterTableRenameColumnStm(), sql.EscapeSqlName(tbName), sql.EscapeSqlName(c.OldName), sql.EscapeSqlName(c.Name))}

	default:
		return nil
	}
}

func (c Column) migrationCommentUp(tbName string) []string {
	if c.Comment == "" || sql.GetDialect() != sql_templates.PostgresDialect {
		return nil
	}

	// apply for postgres only
	return []string{fmt.Sprintf(sql.ColumnComment(), tbName, c.Name, c.Comment)}
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
		return nil

	case MigrateRenameAction:
		c.Name, c.OldName = c.OldName, c.Name

	default:
		return nil
	}

	return c.migrationUp(tbName, after, -1)
}

func (c Column) definition() string {
	strSql := c.typeDefinition()

	for _, opt := range c.Options {
		b := bytes.NewBufferString("")
		var ctx *format.RestoreCtx

		if sql.IsLowercase() {
			ctx = format.NewRestoreCtx(LowerRestoreFlag, b)
		} else {
			ctx = format.NewRestoreCtx(UppercaseRestoreFlag, b)
		}

		if sql.IsPostgres() && opt.Tp == ast.ColumnOptionDefaultValue {
			strSql += " " + b.String()
			continue
		}

		_ = opt.Restore(ctx)
		strSql += " " + b.String()
	}

	return strSql
}

func (c Column) typeDefinition() string {
	switch {
	case sql.IsPostgres() && c.PgType != nil:
		return " " + c.PgType.SQLString()
	case sql.IsSqlite() && c.LiteType != nil:
		return " " + c.LiteType.Name.Name
	case c.MysqlType != nil:
		return " " + c.MysqlType.String()
	}

	return "" // column type is empty
}
