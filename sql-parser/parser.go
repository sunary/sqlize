package sql_parser

import (
	"github.com/sunary/sqlize/element"
)

type Parser struct {
	isPostgres bool
	Migration  element.Migration
}

func NewParser(isPostgres, isLower bool) *Parser {
	return &Parser{
		isPostgres: isPostgres,
		Migration:  element.NewMigration(isPostgres, isLower),
	}
}

func (p *Parser) Parser(sql string) error {
	if p.isPostgres {
		return p.ParserPostgresql(sql)
	}

	return p.ParserMysql(sql)
}

func (p *Parser) Diff(old Parser) {
	p.Migration.Diff(old.Migration)
}

func (p Parser) MigrationUp() string {
	return p.Migration.MigrationUp()
}

func (p Parser) MigrationDown() string {
	return p.Migration.MigrationDown()
}
