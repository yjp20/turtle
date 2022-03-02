package straw

import (
	"strings"
)

func Filter(in []byte) []byte {
	s := string(in)
	s = strings.ReplaceAll(s, "func", "λ")
	s = strings.ReplaceAll(s, "->", "→")
	s = strings.ReplaceAll(s, "=>", "⇒")
	s = strings.ReplaceAll(s, "for", "∀")
	s = strings.ReplaceAll(s, "each", "∈")
	s = strings.ReplaceAll(s, "..", "‥")
	return []byte(s)
}
