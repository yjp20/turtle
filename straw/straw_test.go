package straw

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yjp20/turtle/straw/ast"
	"github.com/yjp20/turtle/straw/interpreter"
	"github.com/yjp20/turtle/straw/parser"
	"github.com/yjp20/turtle/straw/token"
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
			ps := parser.NewParser(file, &errors)
			as := ps.ParseProgram()
			if len(errors) != 0 && !test.shouldError {
				t.Errorf("didn't expect to error")
				for i := 0; i < len(errors); i++ {
					t.Errorf(errors[i].(token.Error).Print(file))
				}
				return
			}
			if len(errors) == 0 && test.shouldError {
				t.Errorf("expected error, but parser didn't throw any\nast: %s", ast.Sprint(as))
				return
			}

			global := interpreter.NewGlobalFrame(&errors)
			frame := interpreter.NewFunctionFrame(global)
			object := global.Eval(as, frame)

			if test.out != object.Inspect() {
				t.Errorf("expected: %s  got: %s\nast: %s", test.out, object.Inspect(), ast.Sprint(as))
			}
		})
	}
}
