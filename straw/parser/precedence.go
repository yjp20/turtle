package parser

import "github.com/yjp20/turtle/straw/token"

type Precedence int

const (
	AS Precedence = iota * 2
	LOWEST
	IF
	EQUAL
	COMPARE
	SUM
	PRODUCT
)

func GetPrecedence(tok token.Token) (Precedence, Precedence) {
	switch tok {
	case token.IDENT:
		return AS, AS
	case token.THEN:
		return IF + 1, IF
	case token.EQUAL:
		return EQUAL, EQUAL
	case token.LESS, token.LESS_EQUAL, token.GREATER, token.GREATER_EQUAL, token.NOT_EQUAL:
		return COMPARE, COMPARE
	case token.ADD, token.SUB:
		return SUM, SUM
	case token.MUL, token.QUO:
		return PRODUCT, PRODUCT
	}
	return LOWEST, LOWEST
}
