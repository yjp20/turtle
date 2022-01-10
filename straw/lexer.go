package straw

import (
	"fmt"
	"unicode"
	"unicode/utf8"

	"github.com/yjp20/straw/token"
)

const eof = -1

type Lexer struct {
	source []byte

	// begin and end represents the [begin, end) of the current unicode rune being read
	begin     int
	end       int
	ch        rune
	errors    []error
	semicolon bool
}

func NewLexer(source []byte, errors []error) *Lexer {
	l := &Lexer{source: source, errors: errors}
	l.readRune()
	return l
}

func (l *Lexer) Next() (token.Token, token.Pos, string) {
	l.skipWhitespace()
	pos := token.Pos(l.begin)

	if isLetter(l.ch) {
		lit := l.readIdentifier()
		tok := token.Lookup(lit)
		if tok == token.BREAK || tok == token.CONTINUE || tok == token.IDENT || tok == token.RETURN {
			l.semicolon = true
		}
		return tok, pos, lit
	}

	if isDecimal(l.ch) {
		lit, hasDecimal := l.readNumber()
		l.semicolon = true
		if hasDecimal {
			return token.FLOAT, pos, lit
		} else {
			return token.INT, pos, lit
		}
	}

	ch := l.ch
	semicolon := l.semicolon
	l.semicolon = false
	l.readRune()

	switch ch {
	case eof:
		if semicolon {
			return token.SEMICOLON, pos, "EOF"
		}
		return token.EOF, pos, "EOF"
	case '\n':
		// semicolon is always true at this point because '\n' would have been
		// consumed by skipWhitespace otherwise
		return token.SEMICOLON, pos, "\n"
	case '#':
		return token.COMMENT, pos, l.readComment()
	case '"':
		l.semicolon = true
		return token.STRING, pos, l.readStringLiteral()
	case '\'':
		l.semicolon = true
		return token.RUNE, pos, l.readRuneLiteral()

	case '+':
		return token.ADD, pos, "+"
	case '-':
		return token.SUB, pos, "-"
	case '*':
		return token.MUL, pos, "*"
	case '/':
		return token.QUO, pos, "/"
	case '%':
		return token.MOD, pos, "%"

	case '&':
		return token.AND, pos, "&"
	case '|':
		return token.OR, pos, "|"
	case '⊕':
		return token.XOR, pos, "⊕"
	case '^':
		return token.XOR, pos, "^"
	case '«':
		return token.SHIFT_LEFT, pos, "«"
	case '»':
		return token.SHIFT_RIGHT, pos, "»"

	case '=':
		return token.EQUAL, pos, "="
	case '<':
		return token.LESS, pos, "<"
	case '>':
		return token.GREATER, pos, ">"
	case '!':
		return token.NOT, pos, "!"
	case '≠':
		return token.NOT_EQUAL, pos, "≠"
	case '≤':
		return token.LESS_EQUAL, pos, "≤"
	case '≥':
		return token.GREATER_EQUAL, pos, "≥"
	case '‥':
		return token.ELIPSIS, pos, "‥"

	case '(':
		return token.LEFT_PAREN, pos, "("
	case ')':
		l.semicolon = true
		return token.RIGHT_PAREN, pos, ")"
	case '[':
		return token.LEFT_BRACK, pos, "["
	case ']':
		l.semicolon = true
		return token.RIGHT_BRACK, pos, "]"
	case '{':
		return token.LEFT_BRACE, pos, "{"
	case '}':
		l.semicolon = true
		return token.RIGHT_BRACE, pos, "}"
	case ',':
		return token.COMMA, pos, ","
	case '.':
		return token.PERIOD, pos, "."
	case ':':
		return token.ASSIGN, pos, ":"
	case ';':
		return token.SEMICOLON, pos, ";"
	case '←':
		return token.LEFT_ARROW, pos, "←"
	case '→':
		return token.RIGHT_ARROW, pos, "→"
	case '?':
		return token.OPTIONAL, pos, "?"

	case '∀':
		return token.FOR, pos, "∀"
	case '∈':
		return token.EACH, pos, "∈"
	case '⇒':
		return token.THEN, pos, "⇒"
	case '~':
		return token.ELSE, pos, "~"
	case '_':
		return token.DEFAULT, pos, "_"
	}

	l.errors = append(l.errors, fmt.Errorf("Invalid token at pos: %d", int(pos)))
	return token.ILLEGAL, pos, string(ch)
}

func (l *Lexer) readRune() {
	width := 1
	if l.end >= len(l.source) {
		l.ch = eof
	} else {
		l.ch, width = utf8.DecodeRune(l.source[l.end:])
	}
	l.begin = l.end
	l.end = l.end + width
}

func (l *Lexer) peek() byte {
	if l.end >= len(l.source) {
		return 0
	}
	return l.source[l.end]
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' && !l.semicolon || l.ch == '\r' {
		l.readRune()
	}
}

func (l *Lexer) readComment() string {
	// # was consumed, so the start is l.begin - 1
	begin := l.begin - 1
	for l.ch != '\n' && l.ch != eof {
		l.readRune()
	}
	return string(l.source[begin:l.begin])
}

func (l *Lexer) readIdentifier() string {
	begin := l.begin
	for isLetter(l.ch) || isDecimal(l.ch) {
		l.readRune()
	}
	return string(l.source[begin:l.begin])
}

func (l *Lexer) readNumber() (string, bool) {
	begin := l.begin
	hasDecimal := false
	for isDecimal(l.ch) || l.ch == '.' && !hasDecimal {
		if l.ch == '.' {
			hasDecimal = true
		}
		l.readRune()
	}
	return string(l.source[begin:l.begin]), hasDecimal
}

func (l *Lexer) readStringLiteral() string {
	// " was consumed, so the start is l.begin - 1
	begin := l.begin - 1
	valid := true
	for {
		l.readRune()
		if l.ch == eof {
			valid = false
		}
		if l.ch == '"' {
			l.readRune()
			break
		}
	}
	if !valid {
		l.errors = append(l.errors, fmt.Errorf("Expected string to be terminated before EOF"))
	}
	return string(l.source[begin:l.begin])
}

func (l *Lexer) readRuneLiteral() string {
	// " was consumed, so the start is l.begin - 1
	begin := l.begin - 1
	valid := true
	for {
		l.readRune()
		if l.ch == eof {
			valid = false
		}
		if l.ch == '\'' {
			l.readRune()
			break
		}
	}
	if !valid {
		l.errors = append(l.errors, fmt.Errorf("Expected rune literal to be terminated before EOF"))
	}
	return string(l.source[begin:l.begin])
}

func isDecimal(ch rune) bool {
	return '0' <= ch && ch <= '9'
}

func isLetter(ch rune) bool {
	return 'A' <= ch && ch <= 'Z' || 'a' <= ch && ch <= 'z' || ch == '_' || ch >= utf8.RuneSelf && unicode.IsLetter(ch)
}
