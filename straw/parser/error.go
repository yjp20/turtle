package parser

import "github.com/yjp20/turtle/straw/token"

type StrawError struct {
	msg string
	pos token.Pos
	end token.Pos
}

func (se StrawError) Error() string {
	return se.msg
}
