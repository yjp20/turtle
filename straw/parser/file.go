package parser

import (
	"github.com/yjp20/turtle/straw/token"
)

type File struct {
	lines  []token.Pos
	source []byte
}

func NewFile(source []byte) *File {
	return &File{
		lines:  make([]token.Pos, 1),
		source: source,
	}
}

func (f *File) SearchLine(pos token.Pos) int {
	l, r := 0, len(f.lines)
	for r - l > 1 {
		m := (l + r) / 2
		if f.lines[m] < pos {
			l = m
		} else {
			r = m
		}
	}
	return l
}

func (f *File) StartOfLine(line int) token.Pos {
	if line >= len(f.lines) {
		return token.Pos(len(f.source))
	}
	return f.lines[line]
}
