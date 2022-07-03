package parser

import (
	"encoding/json"
	"testing"

	"github.com/yjp20/turtle/straw/pkg/token"
)

func TestParser(t *testing.T) {
	errors := token.NewErrorList()
	p := NewParser(source, &errors)
	prog := p.ParseProgram()

	res, err := json.MarshalIndent(prog, "", "| ")
	if err != nil {
		t.Error(err)
	}
	println(string(res))
}
