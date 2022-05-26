package token

import (
	"fmt"
	"strings"
)

type Error struct {
	msg string
	pos Pos
	end Pos
}

var (
	reset = "\033[0m"
	red   = "\033[31;1;4m"
)

func NewError(msg string, pos Pos, end Pos) Error {
	return Error{msg, pos, end}
}

func (se Error) Error() string {
	return se.msg
}

func (se Error) Print(file *File) string {
	sb := strings.Builder{}
	sb.WriteString(se.Error() + "\n")
	sl := file.StartOfLine(file.SearchLine(se.pos))
	el := file.StartOfLine(file.SearchLine(se.end) + 1)
	sb.Write(file.Source[sl:se.pos])
	sb.WriteString(red)
	sb.Write(file.Source[se.pos:se.end])
	sb.WriteString(reset)
	sb.Write(file.Source[se.end:el])
	sb.WriteString(fmt.Sprintf("%d:%d:%d:%d\n", sl, se.pos, se.end, el))
	return sb.String()
}
