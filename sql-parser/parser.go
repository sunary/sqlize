package sql_parser

import (
	"github.com/sunary/sqlize/element"
	sql_templates "github.com/sunary/sqlize/sql-templates"
)

// Parser ...
type Parser struct {
	dialect   sql_templates.SqlDialect
	Migration element.Migration
}

// NewParser ...
func NewParser(dialect sql_templates.SqlDialect, lowercase bool) *Parser {
	return &Parser{
		dialect:   dialect,
		Migration: element.NewMigration(dialect, lowercase),
	}
}

// Parser ...
func (p *Parser) Parser(sql string) error {
	if p.dialect == sql_templates.PostgresDialect {
		return p.ParserPostgresql(sql)
	}

	return p.ParserMysql(sql)
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
