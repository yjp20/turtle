package ast

// Yes, a lot of the right end points of the expressions are off by one, and
// will also error out because of null exceptions in probably half of the
// cases. I will come back to this later but I seriously cannot be arsed right
// now.

import (
	"strings"

	"github.com/yjp20/turtle/straw/pkg/token"
)

type Node interface {
	Pos() token.Pos
	End() token.Pos
}

type Field struct {
	Name  string
	Type  Node
	Value Node
}

// ---
// General Nodes

// Root of a file's AST
type Program struct {
	Nodes []Node
}

func (p *Program) Pos() token.Pos { return 0 }
func (p *Program) End() token.Pos {
	if len(p.Nodes) == 0 {
		return 0
	}
	return p.Nodes[len(p.Nodes)-1].End()
}

// Comment in the form of `// .*`
type Comment struct {
	TextPos token.Pos
	Text    string
}

func (c *Comment) Pos() token.Pos { return c.TextPos }
func (c *Comment) End() token.Pos { return c.TextPos + token.Pos(len(c.Text)) }

// Multiple consecutive comments in the form of `// .*`
type CommentGroup struct {
	Lines []*Comment
}

func (cg *CommentGroup) Pos() token.Pos { return cg.Lines[0].Pos() }
func (cg *CommentGroup) End() token.Pos {
	if len(cg.Lines) == 0 {
		return 0
	}
	return cg.Lines[len(cg.Lines)-1].End()
}
func (cg *CommentGroup) Text() (string, error) {
	sb := strings.Builder{}
	// TODO: smarter handling of concatenating comment strings
	for _, comment := range cg.Lines {
		_, err := sb.WriteString(comment.Text[2:])
		if err != nil {
			return "", err
		}
	}
	return sb.String(), nil
}

// Assignment expression in the form of `(left): (right)`
type Assign struct {
	Left  Node
	Right Node
}

func (ls *Assign) Pos() token.Pos { return ls.Left.Pos() }
func (ls *Assign) End() token.Pos { return ls.Right.End() }

// Each expression in the form of `(left) ∈ (right)`
type Each struct {
	Left  Node
	Right Node
}

func (es *Each) Pos() token.Pos { return es.Left.Pos() }
func (es *Each) End() token.Pos { return es.Right.End() }

type Branch struct {
	Keyword    token.Token
	KeywordPos token.Pos
	Label      *Identifier
}

func (bs *Branch) Pos() token.Pos { return bs.KeywordPos }
func (bs *Branch) End() token.Pos { return bs.Label.End() }

type Return struct {
	KeywordPos token.Pos
	Body       Node
}

func (rs *Return) Pos() token.Pos { return rs.KeywordPos }
func (rs *Return) End() token.Pos { return rs.Body.End() }

type For struct {
	KeywordPos    token.Pos
	Clause        Node
	RightArrowPos token.Pos
	Body          Node
}

func (fs *For) Pos() token.Pos { return fs.KeywordPos }
func (fs *For) End() token.Pos { return fs.Body.End() }

type Identifier struct {
	WordPos token.Pos
	Value   string
}

func (id *Identifier) Pos() token.Pos { return id.WordPos }
func (id *Identifier) End() token.Pos { return id.WordPos + token.Pos(len(id.Value)) }

type Call struct {
	CallPos   token.Pos
	Procedure Node
	Arguments []Node
}

func (c *Call) Pos() token.Pos { return c.Arguments[0].Pos() }
func (c *Call) End() token.Pos { return c.Arguments[0].End() }

type Construct struct {
	Construct token.Pos
	Type      Node
	Value     *Tuple
}

func (c *Construct) Pos() token.Pos { return c.Construct }
func (c *Construct) End() token.Pos { return c.Value.End() }

type Selector struct {
	Node      Node
	Selection *Identifier
}

func (s *Selector) Pos() token.Pos { return s.Node.Pos() }
func (s *Selector) End() token.Pos { return s.Selection.End() }

type Indexor struct {
	Node  Node
	Index Node
}

func (i *Indexor) Pos() token.Pos { return i.Node.Pos() }
func (i *Indexor) End() token.Pos { return i.Index.End() }

type Tuple struct {
	LeftPos  token.Pos
	Nodes    []Node
	RightPos token.Pos
}

func (t *Tuple) Pos() token.Pos { return t.LeftPos }
func (t *Tuple) End() token.Pos { return t.RightPos }

type Block struct {
	LeftPos  token.Pos
	Nodes    []Node
	RightPos token.Pos
}

func (b *Block) Pos() token.Pos { return b.LeftPos }
func (b *Block) End() token.Pos { return b.RightPos }

type If struct {
	Condition Node
	TrueBody  Node
	FalseBody Node
}

func (i *If) Pos() token.Pos { return i.TrueBody.Pos() }
func (i *If) End() token.Pos { return i.FalseBody.End() }

type DefaultLiteral struct {
	KeywordPos token.Pos
}

func (dl *DefaultLiteral) Pos() token.Pos { return dl.KeywordPos }
func (dl *DefaultLiteral) End() token.Pos { return dl.KeywordPos + 1 }

type IntLiteral struct {
	LiteralPos token.Pos
	Literal    string
	Value      int64
}

func (il *IntLiteral) Pos() token.Pos { return il.LiteralPos }
func (il *IntLiteral) End() token.Pos { return il.LiteralPos + token.Pos(len(il.Literal)) }

type FloatLiteral struct {
	LiteralPos token.Pos
	Literal    string
	Value      float64
}

func (fl *FloatLiteral) Pos() token.Pos { return fl.LiteralPos }
func (fl *FloatLiteral) End() token.Pos { return fl.LiteralPos + token.Pos(len(fl.Literal)) }

type StringLiteral struct {
	LiteralPos token.Pos
	Value      string
}

func (sl *StringLiteral) Pos() token.Pos { return sl.LiteralPos }
func (sl *StringLiteral) End() token.Pos { return sl.LiteralPos + token.Pos(len(sl.Value)) }

type RuneLiteral struct {
	LiteralPos token.Pos
	Value      string
}

func (rl *RuneLiteral) Pos() token.Pos { return rl.LiteralPos }
func (rl *RuneLiteral) End() token.Pos { return rl.LiteralPos + token.Pos(len(rl.Value)) }

type RangeLiteral struct {
	RangePos       token.Pos
	Left           Node
	LeftInclusive  bool
	Right          Node
	RightInclusive bool
	RightPos       token.Pos
}

func (rl *RangeLiteral) Pos() token.Pos { return rl.RangePos }
func (rl *RangeLiteral) End() token.Pos { return rl.RightPos + 1 }

type TrueLiteral struct {
	LiteralPos token.Pos
}

func (tl *TrueLiteral) Pos() token.Pos { return tl.LiteralPos }
func (tl *TrueLiteral) End() token.Pos { return tl.LiteralPos + 4 }

type FalseLiteral struct {
	LiteralPos token.Pos
}

func (fl *FalseLiteral) Pos() token.Pos { return fl.LiteralPos }
func (fl *FalseLiteral) End() token.Pos { return fl.LiteralPos + 5 }

type ProcedureType struct {
	KeywordPos token.Pos
	Name       *Identifier
	Params     []Field
	Arguments  []Field
	ReturnType Node
}

func (ft *ProcedureType) Pos() token.Pos { return ft.KeywordPos }
func (ft *ProcedureType) End() token.Pos { return -1 }

type ProcedureDefinition struct {
	ProcedureType *ProcedureType
	KeywordPos    token.Pos
	Body          Node
}

func (fd *ProcedureDefinition) Pos() token.Pos { return fd.ProcedureType.Pos() }
func (fd *ProcedureDefinition) End() token.Pos { return fd.Body.End() }

type Spread struct {
	KeywordPos token.Pos
	Node       Node
}

func (s *Spread) Pos() token.Pos { return s.KeywordPos }
func (s *Spread) End() token.Pos { return s.Node.End() }

type Prefix struct {
	Operator    token.Token
	OperatorPos token.Pos
	Node        Node
}

func (p *Prefix) Pos() token.Pos { return p.OperatorPos }
func (p *Prefix) End() token.Pos { return p.Node.End() }

type Infix struct {
	Operator    token.Token
	OperatorPos token.Pos
	Left        Node
	Right       Node
}

func (i *Infix) Pos() token.Pos { return i.Left.Pos() }
func (i *Infix) End() token.Pos { return i.Right.End() }

type TypeSpec struct {
	Type    token.Token
	TypePos token.Pos
	Params  *Tuple
	Spec    *Tuple
}

func (ts *TypeSpec) Pos() token.Pos { return ts.TypePos }
func (ts *TypeSpec) End() token.Pos { return ts.Spec.End() }

type As struct {
	Node Node
	Type Node
}

func (a *As) Pos() token.Pos { return a.Node.Pos() }
func (a *As) End() token.Pos { return a.Type.End() }

type Match struct {
	KeywordPos token.Pos
	Node       Node
	Tuple      *Tuple
}

func (m *Match) Pos() token.Pos { return m.KeywordPos }
func (m *Match) End() token.Pos { return m.Tuple.End() }
