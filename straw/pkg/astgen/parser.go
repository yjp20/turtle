package astgen

// This package deals with the static analysis of straw files like lexing and
// parsing.

import (
	"fmt"
	"strconv"

	"github.com/yjp20/turtle/straw/pkg/ast"
	"github.com/yjp20/turtle/straw/pkg/token"
)

const NoPos = token.Pos(-1)

type Parser struct {
	lexer *Lexer

	tok token.Token
	pos token.Pos
	lit string

	errors       *token.ErrorList
	commentGroup *ast.CommentGroup
}

func NewParser(lexer *Lexer, errors *token.ErrorList) *Parser {
	p := &Parser{lexer: lexer, errors: errors}
	p.next()
	return p
}

func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{Nodes: p.parseNodes()}
	if p.tok != token.EOF {
		p.appendError("Didn't consume all tokens in the lexer", p.pos, p.pos+token.Pos(len(p.lit)))
	}
	return program
}

// Increments the token currently looked at by the parser. It is different
// from nextToken in that it "skips" and consumes comment groups.
func (p *Parser) next() {
	p.commentGroup = nil
	p.nextToken()
	p.commentGroup = p.consumeCommentGroup()
}

// Increments the token currently looked at by the parser.
func (p *Parser) nextToken() {
	p.tok, p.pos, p.lit = p.lexer.Next()
}

// Consumes multiple nodes until we don't have a valid atomic node, and returns
// the list
func (p *Parser) parseNodes() []ast.Node {
	exprs := []ast.Node{}
	for p.tok != token.SEMICOLON && p.tok != token.COMMA {
		node := p.parseNode(LOWEST)
		if node == nil {
			break
		}
		p.consumeSemi()
		exprs = append(exprs, node)
	}
	return exprs
}

// Attempts to part an expression, which is either an atomic node or some
// combination of atomic nodes through operators
func (p *Parser) parseNode(precedence Precedence) ast.Node {
	left := p.parseAtomicNode()
	if left == nil {
		return nil
	}

	fmt.Printf("%s\n", p.tok.String())
	lp, rp := GetPrecedence(p.tok)
	if precedence > lp {
		return left
	}

	for {
		switch p.tok {
		case token.ADD, token.SUB, token.MUL, token.QUO, token.EQUAL, token.LESS_EQUAL, token.LESS, token.GREATER_EQUAL, token.GREATER, token.NOT_EQUAL, token.AND, token.OR, token.XOR:
			tok := p.tok
			pos := p.consume(p.tok)
			expr := p.parseNode(rp)
			left = &ast.Infix{
				Operator:    tok,
				OperatorPos: pos,
				Left:        left,
				Right:       expr,
			}
		case token.ASSIGN:
			p.consume(p.tok)
			expr := p.parseNode(rp)
			left = &ast.Assign{
				Left:  left,
				Right: expr,
			}
		case token.EACH:
			p.consume(p.tok)
			expr := p.parseNode(rp)
			left = &ast.Each{
				Left:  left,
				Right: expr,
			}
		case token.THEN:
			i := &ast.If{}
			i.Condition = left
			p.consume(token.THEN)
			i.TrueBody = p.parseNode(IF)
			if p.tok == token.ELSE {
				p.consume(token.ELSE)
				i.FalseBody = p.parseNode(LOWEST)
			}
			left = i
		case token.IDENT:
			t := p.parseNode(LOWEST)
			left = &ast.As{Node: left, Type: t}

		default:
			return left
		}
	}
}

// Attempts to parse an atomic node, which is an expression that is not joined
// by infix operators or is a part of an expression list, ie.  expressions that
// are prefix, postfix, or literals. Returns nil if the current cursor is not
// an atomic expression.
func (p *Parser) parseAtomicNode() ast.Node {
	var expression ast.Node
	switch p.tok {
	case token.NOT, token.SUB, token.MUL, token.AND:
		tok := p.tok
		expression = &ast.Prefix{
			Operator:    tok,
			OperatorPos: p.consume(p.tok),
			Node:        p.parseAtomicNode(),
		}
	case token.LEFT_BRACE:
		expression = p.consumeBlock()
	case token.LEFT_PAREN:
		expression = p.consumeTuple()
	case token.DEFAULT:
		expression = p.consumeDefaultLiteral()
	case token.INT:
		expression = p.consumeIntLiteral()
	case token.FLOAT:
		expression = p.consumeFloatLiteral()
	case token.STRING:
		expression = p.consumeStringLiteral()
	case token.RUNE:
		expression = p.consumeRuneLiteral()
	case token.RANGE:
		expression = p.consumeRangeLiteral()
	case token.TRUE:
		expression = &ast.TrueLiteral{p.consume(token.TRUE)}
	case token.FALSE:
		expression = &ast.FalseLiteral{p.consume(token.FALSE)}
	case token.IDENT:
		expression = p.consumeIdentifier()
	case token.FUNC:
		expression = p.consumeProcedure()
	case token.ELIPSIS:
		expression = p.consumeSpread()
	case token.INTERFACE, token.STRUCT:
		expression = p.consumeTypeSpec()
	case token.MATCH:
		expression = p.consumeMatch()
	case token.PERIOD:
		expression = p.consumeCallNode()
	case token.CONSTRUCT:
		expression = p.consumeConstruct()
	case token.BREAK, token.CONTINUE:
		expression = p.consumeBranch()
	case token.FOR:
		expression = p.consumeFor()
	case token.RETURN:
		expression = p.consumeReturn()
	default:
		return nil
	}

	for {
		switch p.tok {
		case token.INDEX:
			p.consume(token.INDEX)
			identifier := p.consumeIdentifier()
			expression = &ast.Indexor{
				Node:  expression,
				Index: identifier,
			}
		case token.LEFT_BRACK:
			tuple := p.consumeBrackTuple()
			expression = &ast.Indexor{
				Node:  expression,
				Index: tuple,
			}
		default:
			return expression
		}
	}
}

func (p *Parser) consumeBlock() *ast.Block {
	return &ast.Block{
		LeftPos:  p.consume(token.LEFT_BRACE),
		Nodes:    p.parseNodes(),
		RightPos: p.consume(token.RIGHT_BRACE),
	}
}

func (p *Parser) consumeTuple() *ast.Tuple {
	return &ast.Tuple{
		LeftPos:  p.consume(token.LEFT_PAREN),
		Nodes:    p.parseNodes(),
		RightPos: p.consume(token.RIGHT_PAREN),
	}
}

func (p *Parser) consumeBrackTuple() *ast.Tuple {
	return &ast.Tuple{
		LeftPos:  p.consume(token.LEFT_BRACK),
		Nodes:    p.parseNodes(),
		RightPos: p.consume(token.RIGHT_BRACK),
	}
}

func (p *Parser) consumeConstruct() *ast.Construct {
	return &ast.Construct{
		Construct: p.consume(token.CONSTRUCT),
		Type:      p.parseAtomicNode(),
		Value:     p.consumeTuple(),
	}
}

func (p *Parser) consumeDefaultLiteral() *ast.DefaultLiteral {
	return &ast.DefaultLiteral{
		KeywordPos: p.consume(token.DEFAULT),
	}
}

func (p *Parser) consumeIntLiteral() *ast.IntLiteral {
	value, err := strconv.ParseInt(p.lit, 10, 64)
	if err != nil {
		// TODO handle error
	}
	lit := p.lit
	pos := p.consume(token.INT)
	return &ast.IntLiteral{
		LiteralPos: pos,
		Literal:    lit,
		Value:      value,
	}
}

func (p *Parser) consumeFloatLiteral() *ast.FloatLiteral {
	value, err := strconv.ParseFloat(p.lit, 64)
	if err != nil {
		// TODO handle error
	}
	lit := p.lit
	pos := p.consume(token.FLOAT)
	return &ast.FloatLiteral{
		LiteralPos: pos,
		Literal:    lit,
		Value:      value,
	}
}

func (p *Parser) consumeStringLiteral() *ast.StringLiteral {
	lit := p.lit
	pos := p.consume(token.STRING)
	// FIXME, currently assumes that the string is well formed, which it is not
	// guaranteed to be. Also, there is no support for escape characters which
	// will eventually be needed
	return &ast.StringLiteral{
		LiteralPos: pos,
		Value:      lit[1 : len(lit)-1],
	}
}

func (p *Parser) consumeRuneLiteral() *ast.RuneLiteral {
	lit := p.lit
	pos := p.consume(token.RUNE)
	// FIXME, currently assumes that the rune is well formed, which it is not
	// guaranteed to be. Also, there is no support for escape characters which
	// will eventually be needed
	return &ast.RuneLiteral{
		LiteralPos: pos,
		Value:      lit[1 : len(lit)-1],
	}
}

func (p *Parser) consumeRangeLiteral() *ast.RangeLiteral {
	rl := &ast.RangeLiteral{RangePos: p.consume(token.RANGE)}

	switch p.tok {
	case token.LEFT_PAREN:
		rl.LeftInclusive = false
		p.consume(token.LEFT_PAREN)
	case token.LEFT_BRACK:
		rl.LeftInclusive = true
		p.consume(token.LEFT_BRACK)
	default:
		p.appendError("Expected to range literal to have either '[' or '(' to start the range", p.pos, p.pos+1)
	}

	rl.Left = p.parseAtomicNode()
	p.consume(token.ELIPSIS)
	rl.Right = p.parseAtomicNode()

	switch p.tok {
	case token.RIGHT_PAREN:
		rl.RightInclusive = false
		p.consume(token.RIGHT_PAREN)
	case token.RIGHT_BRACK:
		rl.RightInclusive = true
		p.consume(token.RIGHT_BRACK)
	default:
		p.appendError("Expected to range literal to have either ']' or ')' to end the range", p.pos, p.pos+1)
	}
	return rl
}

func (p *Parser) tryConsumeIdentifier() *ast.Identifier {
	if p.tok == token.IDENT {
		return p.consumeIdentifier()
	}
	return nil
}

func (p *Parser) consumeIdentifier() *ast.Identifier {
	lit := p.lit
	return &ast.Identifier{
		Value:   lit,
		WordPos: p.consume(token.IDENT),
	}
}

func (p *Parser) consumeProcedure() ast.Node {
	var node ast.Node
	pos := p.consume(token.FUNC)
	pt := &ast.ProcedureType{KeywordPos: pos}
	node = pt
	if p.tok == token.IDENT {
		pt.Name = p.tryConsumeIdentifier()
	}
	if p.tok == token.LEFT_BRACK {
		bt := p.consumeBrackTuple()
		pt.Params = toFields(bt)
	}
	t := p.consumeTuple()
	pt.Arguments = toFields(t)
	pt.ReturnType = p.parseAtomicNode()
	if p.tok == token.RIGHT_ARROW {
		node = &ast.ProcedureDefinition{
			ProcedureType: pt,
			KeywordPos:    p.consume(token.RIGHT_ARROW),
			Body:          p.parseNode(EACH),
		}
	}

	return node
}

func (p *Parser) consumeSpread() *ast.Spread {
	var (
		pos  = p.consume(token.ELIPSIS)
		expr = p.parseAtomicNode()
	)
	return &ast.Spread{
		KeywordPos: pos,
		Node:       expr,
	}
}

func (p *Parser) consumeTypeSpec() *ast.TypeSpec {
	if p.tok != token.INTERFACE && p.tok != token.STRUCT {
		// TODO handle error
	}
	ts := &ast.TypeSpec{
		Type:    p.tok,
		TypePos: p.consume(p.tok),
	}
	if p.tok == token.LEFT_BRACK {
		ts.Params = p.consumeBrackTuple()
	}
	ts.Spec = p.consumeTuple()
	return ts
}

func (p *Parser) consumeMatch() *ast.Match {
	var (
		match = p.consume(token.MATCH)
		node  = p.parseNode(LOWEST)
		tuple = p.consumeTuple()
	)
	return &ast.Match{
		KeywordPos: match,
		Node:       node,
		Tuple:      tuple,
	}
}

func (p *Parser) consumeCallNode() ast.Node {
	p.consume(token.PERIOD)
	arguments := make([]ast.Node, 0)
	proc := p.parseAtomicNode()
	for {
		expr := p.parseAtomicNode()
		if expr == nil {
			break
		}
		arguments = append(arguments, expr)
	}
	return &ast.Call{Procedure: proc, Arguments: arguments}
}

func (p *Parser) consumeBranch() *ast.Branch {
	return &ast.Branch{
		Keyword:    p.tok,
		KeywordPos: p.consume(p.tok),
		Label:      p.tryConsumeIdentifier(),
	}
}

func (p *Parser) consumeFor() *ast.For {
	var (
		KeywordPos    = p.consume(token.FOR)
		Clause        = p.parseNode(EACH)
		RightArrowPos = p.consume(token.RIGHT_ARROW)
		Body          = p.parseNode(LOWEST)
	)
	return &ast.For{
		KeywordPos,
		Clause,
		RightArrowPos,
		Body,
	}
}

func (p *Parser) consumeReturn() *ast.Return {
	return &ast.Return{
		KeywordPos: p.consume(token.RETURN),
		Body:       p.parseNode(LOWEST),
	}
}

func (p *Parser) consumeCommentGroup() *ast.CommentGroup {
	if p.tok != token.COMMENT {
		return nil
	}
	group := &ast.CommentGroup{Lines: make([]*ast.Comment, 0)}
	for p.tok == token.COMMENT {
		lit := p.lit
		pos := p.consume(token.COMMENT)
		group.Lines = append(group.Lines, &ast.Comment{
			TextPos: pos,
			Text:    lit,
		})
	}
	return group
}

func (p *Parser) consumeSemi() {
	if p.tok == token.SEMICOLON || p.tok == token.COMMA {
		p.consume(p.tok)
	}
}

func (p *Parser) consume(tok token.Token) token.Pos {
	if p.tok != tok {
		p.appendError(
			fmt.Sprintf("Expected '%s' got '%s'", tok.String(), p.tok.String()),
			p.pos,
			p.pos+token.Pos(len(p.lit)),
		)
		return NoPos
	}
	p.next()
	return p.pos
}

func (p *Parser) appendError(msg string, pos token.Pos, end token.Pos) {
	*p.errors = append(*p.errors, token.NewError("[parser] "+msg, pos, end))
}

func toFields(tuple *ast.Tuple) []ast.Field {
	fields := make([]ast.Field, 0)
	for _, n := range tuple.Nodes {
		switch n := n.(type) {
		case *ast.As:
			fields = append(fields, ast.Field{
				Name: n.Node.(*ast.Identifier).Value,
				Type: n.Type,
			})

		case *ast.Assign:
			switch l := n.Left.(type) {
			case *ast.As:
				fields = append(fields, ast.Field{
					Name:  l.Node.(*ast.Identifier).Value,
					Type:  l.Type,
					Value: n.Right,
				})
			}
		}
	}
	return fields
}
