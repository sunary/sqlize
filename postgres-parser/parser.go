package postgres_parser

import (
	"fmt"
	"strings"

	"github.com/pingcap/parser/ast"
	"github.com/sunary/sqlize/element"
)

// Parser declaration
type Parser struct {
	Migration element.Migration
	s         *Scanner

	// current token & literal
	token Token
	lit   string
}

// NewParser ...
func NewParser() *Parser {
	return &Parser{
		token: ILLEGAL,
		lit:   "",
	}
}

// Parse ...
func (p *Parser) Parse(sql string) error {
	p.s = NewScanner(strings.NewReader(sql))

	for {
		p.next()
		switch p.token {
		case CREATE:
			err := p.parseCreateTable()
			if err != nil {
				return err
			}

		case ALTER:

		case EOF:
			return nil

		default:
			return p.expectErr()
		}
	}
}

var _ = `
CREATE TABLE IF NOT EXISTS inventory
(
    id                 SERIAL,
    sku                VARCHAR        NOT NULL,
    warehouse          VARCHAR        NOT NULL,
    available_quantity FLOAT8         NOT NULL DEFAULT 0,
    group_id           INTEGER        NOT NULL DEFAULT 0, -- for partition purpose
    created_at         timestamptz(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at         timestamptz(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at         timestamptz(6),
    PRIMARY KEY (id, group_id)
) PARTITION BY LIST (group_id);
`

func (p *Parser) parseCreateTable() error {
	p.next()
	if p.token != TABLE {
		return p.expectErr(TABLE)
	}

	p.next()
	if p.token == IF {
		err := p.expectedNext(NOT, EXISTS)
		if err != nil {
			return err
		}

		p.next()
	}

	p.Migration.AddTable(*element.NewTable(p.lit))
	p.Migration.Using(p.lit)

	err := p.expectedNext(LBRACE)
	if err != nil {
		return err
	}

	for {
		p.next()
		switch p.token {
		case INDEX:
			err := p.parseIndex()
			if err != nil {
				return err
			}

		case PRIMARY:
			err := p.expectedNext(KEY)
			if err != nil {
				return err
			}
			err = p.parseIndex()
			if err != nil {
				return err
			}

		case RBRACE:
			return nil

		default:
			p.next()
			if p.token == COMMENT {
				_, err = p.parseString()
				if err != nil {
					return err
				}

				p.next()
			} else {
				err := p.parseColumn(p.lit)
				if err != nil {
					return err
				}
			}
		}
	}
}

func (p *Parser) parseIndex() error {
	idx := element.Index{
		Node: element.Node{
			Name: "",
		},
	}

	if p.token == LPAREN {
		p.next()
		for p.token == IDENT {
			idx.Columns = append(idx.Columns, p.lit)
			p.next()
			if p.token == COMMA {
				p.next()
			}
		}

		if p.token != RPAREN {
			return p.expectErr(RPAREN)
		}
	} else if p.token == IDENT {
		idx.Columns = append(idx.Columns, p.lit)
	} else {
		return p.expectErr(COLUMN_NAME)
	}

	p.next()

	if p.token == LBRACK {
		commaAllowed := false

		for {
			p.next()
			switch {
			case p.token == UNIQUE:
				idx.Typ = ast.IndexKeyTypeUnique

			case p.token == COMMA:
				if !commaAllowed {
					return p.expectErr(INDEX_NAME)
				}

			case p.token == RBRACK:
				p.next()
				return nil

			default:
				return p.expectErr(PK, UNIQUE)
			}
			commaAllowed = !commaAllowed
		}
	}

	p.Migration.AddIndex("", idx)
	return nil
}

func (p *Parser) parseColumn(name string) error {
	col := element.Column{
		Node: element.Node{
			Name: name,
		},
	}
	if p.token != IDENT {
		return p.expectErr(INTEGER, VARCHAR)
	}

	col.StrTyp = p.lit
	p.next()

	// parse for type
	switch p.token {
	case LPAREN:
		p.next()
		if p.token != INTEGER {
			return p.expectErr(INTEGER)
		}

		col.StrTyp = fmt.Sprintf("%s(%s)", col.StrTyp, p.lit)
		p.next()
		if p.token != RPAREN {
			return p.expectErr(RPAREN)
		}

		p.next()
		if p.token != LBRACK {
			break
		}
		fallthrough

	case LBRACK:
		//handle parseColumn
		err := p.parseColumnSettings(&col)
		if err != nil {
			return err
		}

		p.next() // remove ']'
	}

	p.Migration.AddColumn("", col)
	return nil
}

func (p *Parser) parseColumnSettings(col *element.Column) error {
	commaAllowed := false

	for {
		p.next()
		switch p.token {
		case PK:
			col.Options = append(col.Options, &ast.ColumnOption{
				Tp: ast.ColumnOptionPrimaryKey,
			})

		case PRIMARY:
			p.next()
			if p.token != KEY {
				return p.expectErr(KEY)
			}
			col.Options = append(col.Options, &ast.ColumnOption{
				Tp: ast.ColumnOptionPrimaryKey,
			})

		case NOT:
			p.next()
			if p.token != NULL {
				return p.expectErr(NULL)
			}
			col.Options = append(col.Options, &ast.ColumnOption{
				Tp: ast.ColumnOptionNotNull,
			})

		case UNIQUE:
			col.Options = append(col.Options, &ast.ColumnOption{
				Tp: ast.ColumnOptionUniqKey,
			})

		case INCR:
			col.Options = append(col.Options, &ast.ColumnOption{
				Tp: ast.ColumnOptionAutoIncrement,
			})

		case DEFAULT:
			p.next()
			if p.token != COLON {
				return p.expectErr(COLON)
			}
			p.next()
			switch p.token {
			case STRING, DSTRING, TSTRING, INTEGER, NUMBER, EXPR:
				col.Options = append(col.Options, &ast.ColumnOption{
					Tp:       ast.ColumnOptionDefaultValue,
					StrValue: p.lit,
				})

			default:
				return p.expectErr(DEFAULT)
			}
		case COMMA:
			if !commaAllowed {
				return p.expectErr()
			}

		case RBRACK:
			return nil

		default:
			return p.expectErr(PRIMARY, PK, UNIQUE)
		}

		commaAllowed = !commaAllowed
	}
}

func (p *Parser) parseString() (string, error) {
	p.next()
	switch p.token {
	case STRING, DSTRING, TSTRING:
		return p.lit, nil

	default:
		return "", p.expectErr(STRING, DSTRING, TSTRING)
	}
}

func (p *Parser) next() {
	for {
		p.token, p.lit = p.s.Read()
		if p.token != COMMENT {
			break
		}
	}
}

func (p *Parser) expectedNext(keywords ...Token) error {
	for _, k := range keywords {
		p.next()
		if p.token != k {
			return p.expectErr(k)
		}
	}

	return nil
}

func (p *Parser) expectErr(toks ...Token) error {
	l, c := p.s.LineInfo()
	expected := make([]string, len(toks))
	for i := range toks {
		quote := "'"
		if tokens[toks[i]] == "'" {
			quote = "\""
		}
		expected[i] = quote + tokens[toks[i]] + quote
	}

	return fmt.Errorf("[%d:%d] invalid token '%s', expected: %s", l, c, p.lit, strings.Join(expected, ","))
}
