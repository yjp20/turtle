package parser

import (
	"fmt"
	"testing"

	"github.com/yjp20/turtle/straw/token"
)

func expect(l *Lexer, t *testing.T, tok token.Token, lit string) {
	to, _, li := l.Next()
	if to != tok {
		t.Errorf("expected '%v', got '%v' for token", tok, to)
		return
	}
	if li != lit {
		t.Errorf("expected '%v', got '%v' for lit ", lit, li)
		return
	}
}

func TestLexer(t *testing.T) {
	errors := []error{}
	l := NewLexer(source, &errors)
	for {
		tok, _, lit := l.Next()
		if lit == "\n" {
			lit = "newline"
		}
		fmt.Printf("%-11s '%v'\n", tok.String(), lit)
		if tok == token.EOF {
			break
		}
	}
}
