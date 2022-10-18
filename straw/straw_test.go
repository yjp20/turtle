package straw

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yjp20/turtle/straw/pkg/ast"
	"github.com/yjp20/turtle/straw/pkg/astgen"
	"github.com/yjp20/turtle/straw/pkg/irgen"
	"github.com/yjp20/turtle/straw/pkg/token"
	"github.com/yjp20/turtle/straw/pkg/vm"
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
			name:        entry.Name(),
			in:          b,
			out:         lines[len(lines)-2][2:],
			shouldError: false,
		})
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			errors := token.NewErrorList()

			file := token.NewFile([]byte(test.in))
			lex := astgen.NewLexer(file, &errors)
			par := astgen.NewParser(lex, &errors)
			gen := irgen.NewGenerator(&errors)

			t.Log(string(test.in))
			node := par.ParseProgram()
			t.Log(ast.Print(node))
			code := gen.Generate(node)
			t.Log(code.String())

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

			frame := vm.NewFrame(nil)
			object := vm.Eval(code, &errors, frame)

			if object == nil {
				t.Errorf("expected: %s  got: nil\n", test.out)
			} else if test.out != object.String() {
				t.Errorf("expected: %s  got: %s\n", test.out, object.String())
			}
		})
	}
}
