package astgen

import "github.com/yjp20/turtle/straw/pkg/token"

type Precedence int

const (
	LOWEST Precedence = iota * 2
	ASSIGN
	EACH
	IF
	COMPARE
	SUM
	PRODUCT
	AS
)

func GetPrecedence(tok token.Token) (Precedence, Precedence) {
	switch tok {
	case token.ASSIGN:
		return ASSIGN, ASSIGN
	case token.EACH:
		return EACH, EACH
	case token.THEN:
		return IF + 1, IF
	case token.EQUAL, token.LESS, token.LESS_EQUAL, token.GREATER, token.GREATER_EQUAL, token.NOT_EQUAL:
		return COMPARE, COMPARE
	case token.ADD, token.SUB:
		return SUM, SUM
	case token.MUL, token.QUO:
		return PRODUCT, PRODUCT
	case token.IDENT:
		return AS, AS
	}
	return LOWEST, LOWEST
}
