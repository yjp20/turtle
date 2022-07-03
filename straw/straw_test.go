package straw

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yjp20/turtle/straw/pkg/ast"
	"github.com/yjp20/turtle/straw/pkg/vm"
	"github.com/yjp20/turtle/straw/pkg/parser"
	"github.com/yjp20/turtle/straw/pkg/generator"
	"github.com/yjp20/turtle/straw/pkg/token"
)

type Test struct {
	name        string
	in          []byte
	out         string
	shouldError bool
}

func TestStraw(t *testing.T) {
	tests := make([]Test, 0)
	entries, err := os.ReadDir("examples")
	if err != nil {
		t.Error(err)
		return
	}

	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".st") {
			continue
		}
		b, err := os.ReadFile(filepath.Join("examples", entry.Name()))
		if err != nil {
			t.Error(err)
			return
		}
		lines := strings.Split(string(b), "\n")
		tests = append(tests, Test{
			name: entry.Name(),
			in: b,
			out: lines[len(lines) - 2][2:],
			shouldError: false,
		})
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			errors := token.NewErrorList()

			file := token.NewFile([]byte(test.in))
			lex := parser.NewLexer(file, &errors)
			par := parser.NewParser(lex, &errors)
			gen := generator.NewGenerator(&errors)

			node := par.ParseProgram()
			code := gen.Generate(node)

			if len(errors) != 0 && !test.shouldError {
				t.Errorf("didn't expect to error")
				for i := 0; i < len(errors); i++ {
					t.Errorf(errors[i].(token.Error).Print(file))
				}
				return
			}
			if len(errors) == 0 && test.shouldError {
				t.Errorf("expected error, but parser didn't throw any\nast: %s", ast.Print(node))
				return
			}

			object := vm.Eval(code, &errors, nil)

			if test.out != object.String() {
				t.Errorf("expected: %s  got: %s\nCODE\n=====\n%s\nAST\n=====\n%s\nIR\n=====\n%s\n", test.out, object.String(), test.in, ast.Print(node), code.String())
			}
		})
	}
}
