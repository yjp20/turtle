package main

import (
	"io/ioutil"
	"os"

	"github.com/yjp20/turtle/straw/pkg/ast"
	"github.com/yjp20/turtle/straw/pkg/compiler/rv64"
	"github.com/yjp20/turtle/straw/pkg/generator"
	"github.com/yjp20/turtle/straw/pkg/parser"
	"github.com/yjp20/turtle/straw/pkg/token"
)

func main() {
	b, _ := ioutil.ReadAll(os.Stdin)

	errors := token.NewErrorList()
	file := token.NewFile(b)
	lex := parser.NewLexer(file, &errors)
	par := parser.NewParser(lex, &errors)
	gen := generator.NewGenerator(&errors)

	node := par.ParseProgram()
	code := gen.Generate(node)

	println(ast.Print(node))
	println(code.String())

	os.Stdout.WriteString(rv64.Compile(code))
}
