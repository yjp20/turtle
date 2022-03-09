package parser

import (
	"encoding/json"
	"testing"
)

func TestParser(t *testing.T) {
	errors := make([]error, 0)
	p := NewParser(source, &errors)
	prog := p.ParseProgram()

	res, err := json.MarshalIndent(prog, "", "| ")
	if err != nil {
		t.Error(err)
	}
	println(string(res))
}
