package token

//go:generate stringer -type=Token

type Token int

const (
	ILLEGAL Token = iota
	EOF
	COMMENT

	IDENT  // sheeeeeeeeeeeeeesh
	INT    // 123
	FLOAT  // 123. 123.45 0.34
	RUNE   // 'a'
	STRING // "abc"
	TRUE   // true
	FALSE  // false

	ADD // +
	SUB // -
	MUL // *
	QUO // %
	MOD // $

	INDEX // /

	AND         // &
	OR          // |
	XOR         // ⊕
	EXPONENT    // ^
	SHIFT_LEFT  // «
	SHIFT_RIGHT // »

	ASSIGN        // :
	NOT           // !
	LOGICAL_AND   // and
	LOGICAL_OR    // or
	LOGICAL_XOR   // xor
	EQUAL         // =
	LESS          // <
	GREATER       // >
	NOT_EQUAL     // ≠
	LESS_EQUAL    // ≤
	GREATER_EQUAL // ≥
	ELIPSIS       // ‥

	LEFT_PAREN  // (
	RIGHT_PAREN // )
	LEFT_BRACK  // [
	RIGHT_BRACK // ]
	LEFT_BRACE  // {
	RIGHT_BRACE // }
	COMMA       // ,
	PERIOD      // .
	SEMICOLON   // ;
	LEFT_ARROW  // ←
	RIGHT_ARROW // →
	OPTIONAL    // ?

	FUNC      // λ
	FOR       // ∀
	EACH      // ∈
	THEN      // ⇒
	ELSE      // ~
	CONSTRUCT // ■
	BREAK     // break
	CONTINUE  // continue
	RETURN    // return
	DEFAULT   // _
	GO        // go
	SELECT    // select
	MATCH     // match

	MUTABLE      // μ
	COMPILE_TIME // σ

	RANGE     // range
	CHAN      // chan
	INTERFACE // interface
	STRUCT    // struct
)

func Lookup(lit string) Token {
	switch lit {
	case "λ":
		return FUNC
	case "μ":
		return MUTABLE
	case "σ":
		return COMPILE_TIME
	case "■":
		return CONSTRUCT

	case "true":
		return TRUE
	case "false":
		return FALSE

	case "or":
		return OR
	case "xor":
		return XOR
	case "and":
		return AND
	case "break":
		return BREAK
	case "continue":
		return CONTINUE
	case "return":
		return RETURN
	case "go":
		return GO
	case "select":
		return SELECT
	case "match":
		return MATCH

	case "range":
		return RANGE
	case "chan":
		return CHAN
	case "interface":
		return INTERFACE
	case "struct":
		return STRUCT
	case "_":
		return DEFAULT
	default:
		return IDENT
	}
}
