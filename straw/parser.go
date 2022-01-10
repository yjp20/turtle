package straw

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/yjp20/straw/ast"
	"github.com/yjp20/straw/token"
)

const NoPos = token.Pos(-1)

type Parser struct {
	lexer *Lexer

	tok token.Token
	pos token.Pos
	lit string

	errors       []error
	commentGroup *ast.CommentGroup
}

func NewParser(source []byte) *Parser {
	errors := make([]error, 0)
	p := &Parser{
		lexer:  NewLexer(source, errors),
		errors: errors,
	}
	p.next()
	return p
}

func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{
		Statements: p.parseStatements(),
	}
	if p.tok != token.EOF {
		// panic("DIDNT FINISH")
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

func (p *Parser) parseStatements() []ast.Statement {
	statement := p.parseStatement()
	statements := []ast.Statement{statement}
	for p.tok == token.SEMICOLON || p.tok == token.COMMA {
		p.consumeSemi()
		if p.tok == token.RIGHT_BRACE || p.tok == token.RIGHT_PAREN {
			break
		}
		statements = append(statements, p.parseStatement())
	}
	return statements
}

func (p *Parser) parseStatement() ast.Statement {
	compileTime := false
	if p.tok == token.COMPILE_TIME {
		compileTime = true
		p.consume(token.COMPILE_TIME)
	}

	var statement ast.Statement
	switch p.tok {
	case token.BREAK, token.CONTINUE:
		statement = p.consumeBranchStatement()
	case token.FOR:
		statement = p.consumeForStatement()
	case token.SEMICOLON:
		statement = p.consumeEmptyStatement()
	default:
		statement = p.consumeFlexStatement()
	}

	if compileTime {
		// TODO
	}

	return statement
}

func (p *Parser) parseExpression(precedence Precedence) ast.Expression {
	var left ast.Expression = p.parseAtomicExpression()
	if left == nil {
		return nil
	}

	for {
		switch p.tok {
		case token.ADD, token.SUB, token.MUL, token.QUO, token.EQUAL, token.LESS_EQUAL, token.LESS, token.GREATER_EQUAL, token.GREATER, token.NOT_EQUAL, token.PERIOD:
			lp, rp := GetPrecedence(p.tok)
			p.nextToken()
			if precedence > lp {
				return left
			}
			left = &ast.Infix{
				Operator:    p.tok,
				OperatorPos: p.pos,
				Left:        left,
				Right:       p.parseExpression(rp),
			}
		case token.THEN:
			lp, rp := GetPrecedence(p.tok)
			if precedence > lp {
				return left
			}
			p.consume(token.THEN)
			var t, f ast.Expression
			t = p.parseExpression(LOWEST)
			if p.tok == token.ELSE {
				p.consume(token.ELSE)
				f = p.parseExpression(rp)
			}
			left = &ast.If{
				Conditional: left,
				True:        t,
				False:       f,
			}
		default:
			lp, rp := GetPrecedence(token.IDENT)
			if precedence > lp {
				return left
			}
			expr := p.parseExpression(rp)
			if expr == nil {
				return left
			}
			left = &ast.Infix{
				Operator:    p.tok,
				OperatorPos: p.pos,
				Left:        left,
				Right:       expr,
			}
		}
	}
}

// Parses an atomic expression, which is an expression that is not joined by
// infix operators or is a part of an expression list. Returns nil if the
// current cursor is not an atomic expression.
func (p *Parser) parseAtomicExpression() ast.Expression {
	var expression ast.Expression
	switch p.tok {
	case token.NOT, token.SUB, token.MUL, token.AND:
		tok := p.tok
		pos := p.pos
		p.next()
		expression = &ast.Prefix{
			Operator:    tok,
			OperatorPos: pos,
			Expression:  p.parseAtomicExpression(),
		}
	case token.LEFT_BRACE:
		expression = p.consumeBlock()
	case token.LEFT_PAREN:
		expression = p.consumeTuple()
	case token.INT:
		expression = p.consumeIntLiteral()
	case token.FLOAT:
		expression = p.consumeFloatLiteral()
	case token.STRING:
		expression = p.consumeStringLiteral()
	case token.RUNE:
		expression = p.consumeRuneLiteral()
	case token.IDENT:
		expression = p.consumeIdentifier()
	case token.RANGE:
		expression = p.consumeRangeLiteral()
	case token.FUNC:
		expression = p.consumeFunctionDefinition()
	case token.ELIPSIS:
		expression = p.consumeSpread()
	case token.INTERFACE, token.STRUCT:
		expression = p.consumeTypeSpec()
	default:
		return nil
	}

	for p.tok == token.LEFT_BRACK {
		expression = &ast.Indexor{
			Expression: expression,
			Index:      p.consumeBrackTuple(),
		}
	}

	return expression
}

// ---
// Statements

func (p *Parser) consumeFlexStatement() ast.Statement {
	left := p.parseExpression(LOWEST)
	if p.tok == token.ASSIGN {
		p.consume(token.ASSIGN)
		right := p.parseExpression(LOWEST)
		return &ast.AssignStatement{Left: left, Right: right}
	}
	if p.tok == token.EACH {
		p.consume(token.EACH)
		right := p.parseExpression(LOWEST)
		return &ast.EachStatement{Left: left, Right: right}
	}
	return &ast.ExpressionStatement{Expression: left}
}

func (p *Parser) consumeEmptyStatement() *ast.EmptyStatement {
	return &ast.EmptyStatement{}
}

func (p *Parser) consumeBranchStatement() *ast.BranchStatement {
	return &ast.BranchStatement{
		Keyword:    p.tok,
		KeywordPos: p.consume(p.tok),
		Label:      p.consumeIdentifier(),
	}
}

func (p *Parser) consumeForStatement() *ast.ForStatement {
	print("AAA")
	defer print("BBB")
	pos := p.consume(token.FOR)
	clauses := p.consumeForClauses()
	p.consume(token.RIGHT_ARROW)
	return &ast.ForStatement{
		For:        pos,
		Statements: clauses,
		Expression: p.parseExpression(LOWEST),
	}
}

func (p *Parser) consumeForClauses() []ast.Statement {
	statements := []ast.Statement{}
	for p.tok != token.RIGHT_ARROW {
		statements = append(statements, p.parseStatement())
		p.consumeSemi()
	}
	return statements
}

// ---
// Expressions

func (p *Parser) consumeTuple() *ast.Tuple {
	return &ast.Tuple{
		LeftParen:  p.consume(token.LEFT_PAREN),
		Statements: p.parseStatements(),
		RightParen: p.consume(token.RIGHT_PAREN),
	}
}

func (p *Parser) consumeBrackTuple() *ast.Tuple {
	return &ast.Tuple{
		LeftParen:  p.consume(token.LEFT_BRACK),
		Statements: p.parseStatements(),
		RightParen: p.consume(token.RIGHT_BRACK),
	}
}

func (p *Parser) consumeBlock() *ast.Block {
	return &ast.Block{
		LeftBrace:  p.consume(token.LEFT_BRACE),
		Statements: p.parseStatements(),
		RightBrace: p.consume(token.RIGHT_BRACE),
	}
}

func (p *Parser) consumeIdentifier() *ast.Identifier {
	lit := p.lit
	pos := p.consume(token.IDENT)
	return &ast.Identifier{
		NamePos: pos,
		Name:    lit,
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
		IntPos:  pos,
		Literal: lit,
		Value:   value,
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
		FloatPos: pos,
		Literal:  lit,
		Value:    value,
	}
}

func (p *Parser) consumeStringLiteral() *ast.StringLiteral {
	lit := p.lit
	pos := p.consume(token.STRING)
	// FIXME, currently assumes that the string is well formed, which it is not
	// guaranteed to be. Also, there is no support for escape characters which
	// will eventually be needed
	return &ast.StringLiteral{
		StringPos: pos,
		Value:     lit[1 : len(lit)-1],
	}
}

func (p *Parser) consumeRuneLiteral() *ast.RuneLiteral {
	lit := p.lit
	pos := p.consume(token.RUNE)
	// FIXME, currently assumes that the rune is well formed, which it is not
	// guaranteed to be. Also, there is no support for escape characters which
	// will eventually be needed
	return &ast.RuneLiteral{
		RunePos: pos,
		Value:   lit[1 : len(lit)-1],
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
		p.errors = append(p.errors, fmt.Errorf("Expected to range literal to have either '[' or '(' to start the range"))
	}

	rl.Left = p.parseAtomicExpression()
	p.consume(token.ELIPSIS)
	rl.Right = p.parseAtomicExpression()

	switch p.tok {
	case token.RIGHT_PAREN:
		rl.LeftInclusive = false
		p.consume(token.RIGHT_PAREN)
	case token.RIGHT_BRACK:
		rl.LeftInclusive = true
		p.consume(token.RIGHT_BRACK)
	default:
		p.errors = append(p.errors, fmt.Errorf("Expected to range literal to have either ')' or ']' to end the range"))
	}
	return rl
}

func (p *Parser) consumeFunctionDefinition() *ast.FunctionDefinition {
	fd := &ast.FunctionDefinition{}
	fd.Func = p.consume(token.FUNC)
	if p.tok == token.IDENT {
		fd.Identifier = p.consumeIdentifier()
	}
	if p.tok == token.LEFT_BRACK {
		fd.Params = p.consumeBrackTuple()
	}
	fd.Args = p.consumeTuple()
	if p.tok != token.LEFT_BRACE {
		fd.Return = p.parseAtomicExpression()
	}
	if p.tok == token.LEFT_BRACE {
		fd.Block = p.consumeBlock()
	}
	return fd
}

func (p *Parser) consumeSpread() *ast.Spread {
	return &ast.Spread{
		Elipsis:    p.consume(token.ELIPSIS),
		Expression: p.parseAtomicExpression(),
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

// ---
// Misc

func (p *Parser) consumeCommentGroup() *ast.CommentGroup {
	if p.tok != token.COMMENT {
		return nil
	}
	group := &ast.CommentGroup{Lines: make([]*ast.Comment, 0)}
	for p.tok == token.COMMENT {
		group.Lines = append(group.Lines, &ast.Comment{
			TextPos: p.pos,
			Text:    p.lit,
		})
		p.nextToken()
	}
	return group
}

func (p *Parser) consume(tok token.Token) token.Pos {
	if p.tok != tok {
		p.errors = append(p.errors, fmt.Errorf("Expected '%s' got '%s'", tok.String(), p.tok.String()))
		return NoPos
	}
	p.next()
	return p.pos
}

func (p *Parser) consumeSemi() {
	if p.tok == token.SEMICOLON || p.tok == token.COMMA {
		p.consume(p.tok)
	}
}

func debug(something interface{}) {
	res, _ := json.MarshalIndent(something, "", "| ")
	println(string(res))
}
