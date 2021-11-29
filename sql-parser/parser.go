package sql_parser

import (
	"github.com/sunary/sqlize/element"
)

// Parser ...
type Parser struct {
	isPostgres bool
	Migration  element.Migration
}

// NewParser ...
func NewParser(isPostgres, isLower bool) *Parser {
	return &Parser{
		isPostgres: isPostgres,
		Migration:  element.NewMigration(isPostgres, isLower),
	}
}

// Parser ...
func (p *Parser) Parser(sql string) error {
	if p.isPostgres {
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
