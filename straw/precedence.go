package straw

import "github.com/yjp20/straw/token"

type Precedence int

const (
	LOWEST Precedence = iota * 2
	IF
	EQUAL
	COMPARE
	SUM
	PRODUCT
	CALL
	SELECTOR
)

func GetPrecedence(tok token.Token) (Precedence, Precedence) {
	switch tok {
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
	case token.IDENT:
		return CALL, CALL
	case token.PERIOD:
		return SELECTOR, SELECTOR
	default:
		return LOWEST, LOWEST
	}
}
