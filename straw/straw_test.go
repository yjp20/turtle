package straw

import (
	"reflect"
	"testing"

	"github.com/yjp20/turtle/straw/ast"
	"github.com/yjp20/turtle/straw/kind"
	"github.com/yjp20/turtle/straw/interpreter"
	"github.com/yjp20/turtle/straw/parser"
)

type Test struct {
	name        string
	in          string
	out         interpreter.Object
	shouldError bool
}

var tests = []Test{
	{
		"order of operations",
		`1 + 2 * 3 + 4`,
		&interpreter.I64{Value: 11},
		false,
	},
	{
		"match",
		`k: 3
		match k {
			3 ⇒ 5
			2 ⇒ 5
			_ ⇒ 7
		}`,
		&interpreter.I64{Value: 5},
		false,
	},
	{
		"match default",
		`k: 10
		match k {
			3 ⇒ 5
			2 ⇒ 5
			_ ⇒ 7
		}`,
		&interpreter.I64{Value: 7},
		false,
	},
	{
		"closures",
		`λ f (x i64) → λ g (y i64) → x + y
		.{ .f 3 } 5`,
		&interpreter.I64{Value: 8},
		false,
	},
	{
		"if true",
		`k: 3
		k = 3 ⇒ 1 ~ 2`,
		&interpreter.I64{Value: 1},
		false,
	},
	{
		"if false",
		`k: 1
		k = 3 ⇒ 1 ~ 2`,
		&interpreter.I64{Value: 2},
		false,
	},
	{
		"if chain true false",
		`j: 1, k: 3
		j = 1 ⇒ 3 ~ k = 2 ⇒ 4 ~ 5`,
		&interpreter.I64{Value: 3},
		false,
	},
	{
		"if chain false true",
		`j: 0, k: 2
		j = 1 ⇒ 3 ~ k = 2 ⇒ 4 ~ 5`,
		&interpreter.I64{Value: 4},
		false,
	},
	{
		"if chain false false",
		`j: 0, k: 3
		j = 1 ⇒ 3 ~ k = 2 ⇒ 4 ~ 5`,
		&interpreter.I64{Value: 5},
		false,
	},
	{
		"tuple",
		`f: λ (a i64, b i64) → (a+b, a-b)
		(x, y): .f 8 2
		(y, x)`,
		&interpreter.Tuple{
			Fields: []interpreter.Field{
				{Name: "0", Type: interpreter.Type{ObjectKind: kind.I64}, Value: &interpreter.I64{Value: 6}},
				{Name: "1", Type: interpreter.Type{ObjectKind: kind.I64}, Value: &interpreter.I64{Value: 10}},
			},
		},
		false,
	},
	{
		"return 1",
		`f: λ (i i64) → {
			i = 10 ⇒ return 100
			return i
		}
		.f 10`,
		&interpreter.I64{Value: 100},
		false,
	},
	{
		"return 2",
		`f: λ (i i64) → {
			i = 10 ⇒ return 100
			return i
		}
		.f 1`,
		&interpreter.I64{Value: 1},
		false,
	},
	{
		"fibo recursive 20",
		`fibo: λ (n i64) → {
			match n {
				0 ⇒ 0
				1 ⇒ 1
				_ ⇒ .fibo { n - 1 } + .fibo { n - 2 }
			}
		}
		.fibo 20`,
		&interpreter.I64{Value: 6765},
		false,
	},
	{
		"fibo array 40",
		`fibo: λ (n i64) → {
			a: .make array[i64] { n+1 }
			a[0]: 0
			a[1]: 1
			∀ i ∈ range[2‥n] → {
				a[i]: a[i-2] + a[i-1]
			}
			a[n]
		}
		.fibo 40`,
		&interpreter.I64{Value: 102334155},
		false,
	},
	{
		"default argument functions",
		`f: λ (n i64: 40) → n
		.f`,
		&interpreter.I64{Value: 40},
		false,
	},
	{
		"array constructor",
		`■ array[i64] (1,2,3,4,5)`,
		interpreter.NULL,
		false,
	},
}

func TestStraw(t *testing.T) {
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			errors := []error{}
			file := parser.NewFile([]byte(test.in))
			ps := parser.NewParser(file, &errors)
			as := ps.ParseProgram()
			if len(errors) != 0 && !test.shouldError {
				t.Errorf("didn't expect to error")
				for i:=0; i<len(errors); i++ {
					t.Errorf(errors[i].(parser.StrawError).Print(file))
				}
				return
			}
			if len(errors) == 0 && test.shouldError {
				t.Errorf("expected error, but parser didn't throw any\nast: %s", ast.Sprint(as))
				return
			}

			global := interpreter.NewGlobalFrame(&errors)
			frame := interpreter.NewFunctionFrame(global)
			object := interpreter.Eval(as, frame)

			if !reflect.DeepEqual(object, test.out) {
				t.Errorf("expected: %s  got: %s\nast: %s", test.out.Inspect(), object.Inspect(), ast.Sprint(as))
			}
		})
	}
}
