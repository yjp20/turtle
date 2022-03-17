package parser

import (
	"strings"
	"fmt"

	"github.com/yjp20/turtle/straw/token"
)

type StrawError struct {
	msg string
	pos token.Pos
	end token.Pos
}

var (
	reset = "\033[0m"
	red   = "\033[31;1;4m"
)

func (se StrawError) Error() string {
	return se.msg
}

func (se StrawError) Print(file *File) string {
	sb := strings.Builder{}
	sb.WriteString(se.Error() + "\n")
	sl := file.StartOfLine(file.SearchLine(se.pos))
	el := file.StartOfLine(file.SearchLine(se.end) + 1)
	sb.Write(file.source[sl:se.pos])
	sb.WriteString(red)
	sb.Write(file.source[se.pos:se.end])
	sb.WriteString(reset)
	sb.Write(file.source[se.end:el])
	sb.WriteString(fmt.Sprintf("%d:%d:%d:%d\n", sl, se.pos, se.end, el))
	return sb.String()
}
