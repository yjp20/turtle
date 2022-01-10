package straw

import (
	"encoding/json"
	"testing"
)

func TestParser(t *testing.T) {
	p := NewParser(source)
	prog := p.ParseProgram()

	res, err := json.MarshalIndent(prog, "", "| ")
	if err != nil {
		t.Error(err)
	}
	println(string(res))
}
