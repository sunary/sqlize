package sql_templates

import (
	"fmt"
	"strings"
)

type SqlDialect string

const (
	MysqlDialect     = "mysql"
	PostgresDialect  = "postgres"
	SqlserverDialect = "sqlserver"
	SqliteDialect    = "sqlite3"
)

// PrimaryOption ...
func (s Sql) PrimaryOption() string {
	return s.apply("PRIMARY KEY")
}

// AutoIncrementOption ...
func (s Sql) AutoIncrementOption() string {
	switch s.dialect {
	case PostgresDialect:
		return s.apply("SERIAL")

	default:
		// TODO: mysql template is default for other dialects
		return s.apply("AUTO_INCREMENT")
	}
}

// DefaultOption ...
func (s Sql) DefaultOption() string {
	return s.apply("DEFAULT %s")
}

// NotNullValue ...
func (s Sql) NotNullValue() string {
	return s.apply("NOT NULL")
}

// NullValue ...
func (s Sql) NullValue() string {
	return s.apply("NULL")
}

// ColumnComment ...
func (s Sql) ColumnComment() string {
	switch s.dialect {
	case PostgresDialect:
		return s.apply("COMMENT ON COLUMN %s.%s IS '%s';")

	default:
		return s.apply("COMMENT '%s'")
	}
}

// TableComment ...
func (s Sql) TableComment() string {
	switch s.dialect {
	case PostgresDialect:
		return s.apply("COMMENT ON TABLE %s IS '%s';")

	default:
		return s.apply("COMMENT '%s'")
	}
}

// EscapeSqlName ...
func (s Sql) EscapeSqlName(name string) string {
	if name == "" {
		return name
	}

	escapeChar := "`"
	if s.dialect == PostgresDialect {
		escapeChar = "\""
	}

	return fmt.Sprintf("%s%s%s", escapeChar, strings.Trim(name, escapeChar), escapeChar)
}

// EscapeSqlNames ...
func (s Sql) EscapeSqlNames(names []string) []string {
	ns := make([]string, len(names))
	for i := range names {
		ns[i] = s.EscapeSqlName(names[i])
	}

	return ns
}
