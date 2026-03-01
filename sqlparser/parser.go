package sqlparser

import (
	"github.com/sunary/sqlize/element"
	"github.com/sunary/sqlize/sqltemplates"
)

// Parser ...
type Parser struct {
	dialect     sqltemplates.SqlDialect
	Migration   element.Migration
	ignoreOrder bool
}

// NewParser ...
func NewParser(dialect sqltemplates.SqlDialect, lowercase, ignoreOrder bool) *Parser {
	return &Parser{
		dialect:   dialect,
		Migration: element.NewMigration(dialect, lowercase, ignoreOrder),
	}
}

// Parser ...
func (p *Parser) Parser(sql string) error {
	switch p.dialect {
	case sqltemplates.PostgresDialect:
		return p.ParserPostgresql(sql)

	case sqltemplates.SqliteDialect:
		return p.ParserSqlite(sql)

	default:
		// TODO: mysql parser is default for remaining dialects
		return p.ParserMysql(sql)
	}
}

// HashValue ...
func (p *Parser) HashValue() int64 {
	return p.Migration.HashValue()
}

// Diff differ between 2 migrations
func (p *Parser) Diff(old Parser) {
	p.Migration.Diff(old.Migration)
}

// MigrationUp migration up
func (p Parser) MigrationUp() string {
	return p.Migration.MigrationUp()
}

// MigrationDown migration down
func (p Parser) MigrationDown() string {
	return p.Migration.MigrationDown()
}
