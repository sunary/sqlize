package postgres_parser

import (
	"bufio"
	"bytes"
	"io"
)

const eof = rune(0)

// Scanner represents a lexical scanner.
type Scanner struct {
	r  *bufio.Reader
	ch rune // for peek
	l  uint
	c  uint
}

// NewScanner returns a new instance of Scanner.
func NewScanner(r io.Reader) *Scanner {
	s := &Scanner{r: bufio.NewReader(r), l: 1, c: 0}
	s.next()
	return s
}

// Next return next token and literal value
func (s *Scanner) Read() (tok Token, lit string) {
	for isWhitespace(s.ch) {
		s.next()
	}

	// Otherwise read the individual character.
	switch {
	case isLetter(s.ch):
		return s.scanIdent()

	case isDigit(s.ch):
		return s.scanNumber()

	default:
		ch := s.ch
		lit := string(ch)
		s.next()
		switch ch {
		case eof:
			return EOF, ""

		case '-':
			if s.ch == '-' {
				return COMMENT, s.scanComment(s.ch)
			}

			return SUB, lit

		case '<':
			return LT, lit

		case '>':
			return GT, lit

		case '(':
			return LPAREN, lit

		case '[':
			return LBRACK, lit

		case '{':
			return LBRACE, lit

		case ')':
			return RPAREN, lit

		case ']':
			return RBRACK, lit

		case '}':
			return RBRACE, lit

		case ';':
			return SEMICOLON, lit

		case ':':
			return COLON, lit

		case ',':
			return COMMA, lit

		case '.':
			return PERIOD, lit

		case '`':
			return s.scanExpression()

		case '\'', '"':
			return s.scanString(ch)

		case '/':
			if s.ch == '/' {
				return COMMENT, s.scanComment(s.ch)
			}

			return ILLEGAL, string(ch)

		case '#':
			return COMMENT, s.scanComment(0)
		}

		return ILLEGAL, string(ch)
	}
}

func (s *Scanner) scanComment(nextChar rune) string {
	var buf bytes.Buffer
	if nextChar != 0 {
		buf.WriteRune(nextChar)
	}

	for s.ch != '\n' && s.ch != eof {
		buf.WriteRune(s.ch)
		s.next()
	}

	return buf.String()
}

func (s *Scanner) scanNumber() (Token, string) {
	var buf bytes.Buffer
	countDot := 0
	for isDigit(s.ch) || (s.ch == '.' && countDot < 2) {
		if s.ch == '.' {
			countDot++
		}

		buf.WriteRune(s.ch)
		s.next()
	}

	if countDot < 1 {
		return INTEGER, buf.String()
	} else if countDot > 1 {
		return ILLEGAL, buf.String()
	}

	return NUMBER, buf.String()
}

func (s *Scanner) scanString(quo rune) (Token, string) {
	switch quo {
	case '"':
		lit, ok := s.scanTo(quo)
		if ok {
			return DSTRING, lit
		}
		return ILLEGAL, lit

	case '\'':
		if s.ch != '\'' {
			lit, ok := s.scanTo(quo)
			if ok {
				return STRING, lit
			}
			return ILLEGAL, lit
		}

		// Handle Triple quote string
		var buf bytes.Buffer
		s.next()
		if s.ch == '\'' { // triple quote string
			s.next()
			count := 0
			for count < 3 {
				switch s.ch {
				case '\'':
					count++
				case eof:
					return ILLEGAL, buf.String()
				}
				buf.WriteRune(s.ch)
				s.next()
			}
			return TSTRING, buf.String()[:buf.Len()-count]
		}
		return ILLEGAL, buf.String()

	default:
		return ILLEGAL, string(eof)
	}
}

func (s *Scanner) scanExpression() (Token, string) {
	lit, ok := s.scanTo('`')
	if ok {
		return EXPR, lit
	}

	return ILLEGAL, lit
}

func (s *Scanner) scanTo(stop rune) (string, bool) {
	var buf bytes.Buffer
	for {
		switch s.ch {
		case stop:
			s.next()
			return buf.String(), true

		case '\n', eof:
			return buf.String(), false

		default:
			buf.WriteRune(s.ch)
			s.next()
		}
	}
}

func (s *Scanner) scanIdent() (tok Token, lit string) {
	var buf bytes.Buffer
	for {
		buf.WriteRune(s.ch)
		s.next()

		if !isLetter(s.ch) && !isDigit(s.ch) && s.ch != '_' && s.ch != '.' {
			break
		}
	}

	return Lookup(buf.String()), buf.String()
}

func (s *Scanner) next() {
	ch, _, err := s.r.ReadRune()
	if err != nil {
		s.ch = eof
		return
	}

	if ch == '\n' {
		s.l++
		s.c = 0
	}

	s.c++
	s.ch = ch
}

// LineInfo return line info
func (s *Scanner) LineInfo() (uint, uint) {
	return s.l, s.c
}

// isWhitespace returns true if the rune is a space, tab, or newline.
func isWhitespace(ch rune) bool { return ch == ' ' || ch == '\t' || ch == '\n' }

// isLetter returns true if the rune is a letter.
func isLetter(ch rune) bool { return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') }

// isDigit returns true if the rune is a digit.
func isDigit(ch rune) bool { return ch >= '0' && ch <= '9' }

// isIdentChar returns true if the rune can be used in an unquoted identifier.
func isIdentChar(ch rune) bool { return isLetter(ch) || isDigit(ch) || ch == '_' }

// isIdentFirstChar returns true if the rune can be used as the first char in an unquoted identifer.
func isIdentFirstChar(ch rune) bool { return isLetter(ch) || ch == '_' }
