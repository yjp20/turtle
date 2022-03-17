package ast

// Yes, a lot of the right end points of the statements are off by one, and
// will also error out because of null exceptions in probably half of the
// cases. I will come back to this later but I seriously cannot be arsed right
// now.

import (
	"strings"

	"github.com/yjp20/turtle/straw/token"
)

type Node interface {
	Pos() token.Pos
	End() token.Pos
}

type Statement interface {
	Node
	statementNode()
}

type Expression interface {
	Node
	expressionNode()
}

// ---
// General Nodes

type Program struct {
	Statements []Statement
}

func (p *Program) Pos() token.Pos { return 0 }
func (p *Program) End() token.Pos { return 0 }

type Comment struct {
	TextPos token.Pos
	Text    string
}

func (c *Comment) Pos() token.Pos { return c.TextPos }
func (c *Comment) End() token.Pos { return c.TextPos + token.Pos(len(c.Text)) }

type CommentGroup struct {
	Lines []*Comment
}

func (cg *CommentGroup) Pos() token.Pos { return cg.Lines[0].Pos() }
func (cg *CommentGroup) End() token.Pos { return cg.Lines[len(cg.Lines)-1].End() }
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

// ---
// Statements

type AssignStatement struct {
	Left  Expression
	Right Expression
}

func (ls *AssignStatement) statementNode() {}
func (ls *AssignStatement) Pos() token.Pos { return ls.Left.Pos() }
func (ls *AssignStatement) End() token.Pos { return ls.Right.End() }

type EachStatement struct {
	Left  Expression
	Right Expression
}

func (es *EachStatement) statementNode() {}
func (es *EachStatement) Pos() token.Pos { return es.Left.Pos() }
func (es *EachStatement) End() token.Pos { return es.Right.End() }

type ExpressionStatement struct {
	Expression Expression
}

func (es *ExpressionStatement) statementNode() {}
func (es *ExpressionStatement) Pos() token.Pos { return es.Expression.Pos() }
func (es *ExpressionStatement) End() token.Pos { return es.Expression.End() }

type BranchStatement struct {
	Keyword    token.Token
	KeywordPos token.Pos
	Label      *Identifier
}

func (bs *BranchStatement) statementNode() {}
func (bs *BranchStatement) Pos() token.Pos { return bs.KeywordPos }
func (bs *BranchStatement) End() token.Pos { return bs.Label.End() }

type ReturnStatement struct {
	Return     token.Pos
	Expression Expression
}

func (rs *ReturnStatement) statementNode() {}
func (rs *ReturnStatement) Pos() token.Pos { return rs.Return }
func (rs *ReturnStatement) End() token.Pos { return rs.Expression.End() }

type ForStatement struct {
	For        token.Pos
	Clauses    []Statement
	Expression Expression
}

func (fs *ForStatement) statementNode() {}
func (fs *ForStatement) Pos() token.Pos { return fs.For }
func (fs *ForStatement) End() token.Pos { return fs.Expression.End() }

// FIXME: Think about how positions work in empty statements. Maybe look at how go does it in their parser?
type EmptyStatement struct{}

func (es *EmptyStatement) statementNode() {}
func (es *EmptyStatement) Pos() token.Pos { return -1 }
func (es *EmptyStatement) End() token.Pos { return -1 }

// ---
// Expresions

type Identifier struct {
	NamePos token.Pos
	Name    string
}

func (id *Identifier) expressionNode() {}
func (id *Identifier) Pos() token.Pos  { return id.NamePos }
func (id *Identifier) End() token.Pos  { return id.NamePos + token.Pos(len(id.Name)) }

type CallExpression struct {
	Expressions []Expression
}

func (cl *CallExpression) expressionNode() {}
func (cl *CallExpression) Pos() token.Pos  { return cl.Expressions[0].Pos() }
func (cl *CallExpression) End() token.Pos  { return cl.Expressions[0].End() }

type ConstructExpression struct {
	Construct token.Pos
	Type      Expression
	Value     *Tuple
}

func (ce *ConstructExpression) expressionNode() {}
func (ce *ConstructExpression) Pos() token.Pos  { return ce.Construct }
func (ce *ConstructExpression) End() token.Pos  { return ce.Value.End() }

type Selector struct {
	Expression Expression
	Selection  *Identifier
}

func (s *Selector) expressionNode() {}
func (s *Selector) Pos() token.Pos  { return s.Expression.Pos() }
func (s *Selector) End() token.Pos  { return s.Selection.End() }

type Indexor struct {
	Expression Expression
	Index      Expression
}

func (i *Indexor) expressionNode() {}
func (i *Indexor) Pos() token.Pos  { return i.Expression.Pos() }
func (i *Indexor) End() token.Pos  { return i.Index.End() }

type Tuple struct {
	Left       token.Pos
	Right      token.Pos
	Statements []Statement
}

func (t *Tuple) expressionNode() {}
func (t *Tuple) Pos() token.Pos  { return t.Left }
func (t *Tuple) End() token.Pos  { return t.Right }

type Block struct {
	LeftBrace  token.Pos
	Statements []Statement
	RightBrace token.Pos
}

func (b *Block) expressionNode() {}
func (b *Block) Pos() token.Pos  { return b.LeftBrace }
func (b *Block) End() token.Pos  { return b.RightBrace }

type If struct {
	Conditional Expression
	True        Statement
	False       Statement
}

func (i *If) expressionNode() {}
func (i *If) Pos() token.Pos  { return i.True.Pos() }
func (i *If) End() token.Pos  { return i.False.End() }

type DefaultLiteral struct {
	DefaultPos token.Pos
}

func (dl *DefaultLiteral) expressionNode() {}
func (dl *DefaultLiteral) Pos() token.Pos  { return dl.DefaultPos }
func (dl *DefaultLiteral) End() token.Pos  { return dl.DefaultPos + 1 }

type IntLiteral struct {
	IntPos  token.Pos
	Literal string
	Value   int64
}

func (il *IntLiteral) expressionNode() {}
func (il *IntLiteral) Pos() token.Pos  { return il.IntPos }
func (il *IntLiteral) End() token.Pos  { return il.IntPos + token.Pos(len(il.Literal)) }

type FloatLiteral struct {
	FloatPos token.Pos
	Literal  string
	Value    float64
}

func (fl *FloatLiteral) expressionNode() {}
func (fl *FloatLiteral) Pos() token.Pos  { return fl.FloatPos }
func (fl *FloatLiteral) End() token.Pos  { return fl.FloatPos + token.Pos(len(fl.Literal)) }

type StringLiteral struct {
	StringPos token.Pos
	Value     string
}

func (sl *StringLiteral) expressionNode() {}
func (sl *StringLiteral) Pos() token.Pos  { return sl.StringPos }
func (sl *StringLiteral) End() token.Pos  { return sl.StringPos + token.Pos(len(sl.Value)) }

type RuneLiteral struct {
	RunePos token.Pos
	Value   string
}

func (rl *RuneLiteral) expressionNode() {}
func (rl *RuneLiteral) Pos() token.Pos  { return rl.RunePos }
func (rl *RuneLiteral) End() token.Pos  { return rl.RunePos + token.Pos(len(rl.Value)) }

type RangeLiteral struct {
	RangePos       token.Pos
	RightPos       token.Pos
	Left           Expression
	LeftInclusive  bool
	Right          Expression
	RightInclusive bool
}

func (rl *RangeLiteral) expressionNode() {}
func (rl *RangeLiteral) Pos() token.Pos  { return rl.RangePos }
func (rl *RangeLiteral) End() token.Pos  { return rl.RightPos + 1 }

type TrueLiteral struct {
	True token.Pos
}

func (tl *TrueLiteral) expressionNode() {}
func (tl *TrueLiteral) Pos() token.Pos  { return tl.True }
func (tl *TrueLiteral) End() token.Pos  { return tl.True + 4 }

type FalseLiteral struct {
	False token.Pos
}

func (fl *FalseLiteral) expressionNode() {}
func (fl *FalseLiteral) Pos() token.Pos  { return fl.False }
func (fl *FalseLiteral) End() token.Pos  { return fl.False + 5 }

type FunctionDefinition struct {
	Func       token.Pos
	Identifier *Identifier
	Args       *Tuple
	Params     *Tuple
	Return     Expression
	Body       Statement
}

func (fd *FunctionDefinition) expressionNode() {}
func (fd *FunctionDefinition) Pos() token.Pos  { return fd.Func }
func (fd *FunctionDefinition) End() token.Pos  { return fd.Body.End() }

type Spread struct {
	Elipsis    token.Pos
	Expression Expression
}

func (s *Spread) expressionNode() {}
func (s *Spread) Pos() token.Pos  { return s.Elipsis }
func (s *Spread) End() token.Pos  { return s.Expression.End() }

type Prefix struct {
	Operator    token.Token
	OperatorPos token.Pos
	Expression  Expression
}

func (p *Prefix) expressionNode() {}
func (p *Prefix) Pos() token.Pos  { return p.OperatorPos }
func (p *Prefix) End() token.Pos  { return p.Expression.End() }

type Infix struct {
	Operator    token.Token
	OperatorPos token.Pos
	Left        Expression
	Right       Expression
}

func (i *Infix) expressionNode() {}
func (i *Infix) Pos() token.Pos  { return i.Left.Pos() }
func (i *Infix) End() token.Pos  { return i.Right.End() }

type Bind struct {
	Left  Expression
	Right Expression
}

func (b *Bind) expressionNode() {}
func (b *Bind) Pos() token.Pos  { return b.Left.Pos() }
func (b *Bind) End() token.Pos  { return b.Right.End() }

type TypeSpec struct {
	Type    token.Token
	TypePos token.Pos
	Params  *Tuple
	Spec    *Tuple
}

func (ts *TypeSpec) expressionNode() {}
func (ts *TypeSpec) Pos() token.Pos  { return ts.TypePos }
func (ts *TypeSpec) End() token.Pos  { return ts.Spec.End() }

type As struct {
	Value Expression
	Type  Expression
}

func (a *As) expressionNode() {}
func (a *As) Pos() token.Pos  { return a.Value.Pos() }
func (a *As) End() token.Pos  { return a.Type.End() }

type Match struct {
	Match      token.Pos
	Left       token.Pos
	Right      token.Pos
	Item       Expression
	Conditions []Expression
	Bodies     []Statement
}

func (m *Match) expressionNode() {}
func (m *Match) Pos() token.Pos  { return m.Left }
func (m *Match) End() token.Pos  { return m.Right + 1 }
