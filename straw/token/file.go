package token

type File struct {
	Lines  []Pos
	Source []byte
}

func NewFile(source []byte) *File {
	return &File{
		Lines:  make([]Pos, 1),
		Source: source,
	}
}

func (f *File) SearchLine(pos Pos) int {
	l, r := 0, len(f.Lines)
	for r - l > 1 {
		m := (l + r) / 2
		if f.Lines[m] < pos {
			l = m
		} else {
			r = m
		}
	}
	return l
}

func (f *File) StartOfLine(line int) Pos {
	if line >= len(f.Lines) {
		return Pos(len(f.Source))
	}
	return f.Lines[line]
}
