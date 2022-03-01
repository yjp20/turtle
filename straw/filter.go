package straw

import (
	"strings"
)

func Filter(in []byte) []byte {
	s := string(in)
	s = strings.ReplaceAll(s, "func", "λ")
	s = strings.ReplaceAll(s, "->", "→")
	s = strings.ReplaceAll(s, "=>", "⇒")
	s = strings.ReplaceAll(s, "for", "⇒")
	return []byte(s)
}
