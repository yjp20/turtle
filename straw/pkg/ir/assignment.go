package ir

import (
	"fmt"

	"github.com/yjp20/turtle/straw/pkg/kind"
)

type Assignment int

func (a Assignment) String() string {
	return fmt.Sprintf("%%%d", a)
}

type Field struct {
	Name  string
	Type  Type
	Value Assignment
}

type Type struct {
	Kind    kind.Kind
	Extra   []Field
	Returns []Field
}
