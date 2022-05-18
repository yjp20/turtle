package parser

// This package deals with the static analysis of straw files like lexing and
// parsing.

import (
	"fmt"
	"strconv"

	"github.com/yjp20/turtle/straw/ast"
	"github.com/yjp20/turtle/straw/token"
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

func NewParser(file *File, errors *token.ErrorList) *Parser {
	p := &Parser{
		lexer:  NewLexer(file, errors),
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

func (p *Parser) parseStatements() []ast.Statement {
	statement := p.parseStatement()
	statements := []ast.Statement{statement}
	for p.tok == token.SEMICOLON || p.tok == token.COMMA {
		p.consumeSemi()
		if p.tok == token.RIGHT_BRACE || p.tok == token.RIGHT_PAREN {
			break
		}
		statement := p.parseStatement()
		if _, ok := statement.(*ast.EmptyStatement); !ok {
			statements = append(statements, statement)
		}
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
	case token.RETURN:
		statement = p.consumeReturn()
	default:
		statement = p.consumeFlexStatement()
	}

	if compileTime {
		// TODO
	}

	return statement
}

func (p *Parser) parseExpression(precedence Precedence) ast.Expression {
	left := p.parseAtomicExpression()

	lp, rp := GetPrecedence(p.tok)
	if precedence > lp {
		return left
	}

	for {
		switch p.tok {
		case token.ADD, token.SUB, token.MUL, token.QUO, token.EQUAL, token.LESS_EQUAL,
			token.LESS, token.GREATER_EQUAL, token.GREATER, token.NOT_EQUAL:
			tok := p.tok
			pos := p.consume(p.tok)
			expr := p.parseExpression(rp)
			left = &ast.Infix{
				Operator:    tok,
				OperatorPos: pos,
				Left:        left,
				Right:       expr,
			}
		case token.THEN:
			p.consume(token.THEN)
			var t, f ast.Statement
			t = p.parseStatement()
			if p.tok == token.ELSE {
				p.consume(token.ELSE)
				f = p.parseStatement()
			}
			left = &ast.IfExpression{
				Conditional: left,
				True:        t,
				False:       f,
			}
		case token.IDENT:
			t := p.parseExpression(LOWEST)
			left = &ast.As{Value: left, Type: t}

		default:
			return left
		}
	}
}

func (p *Parser) consumeCallExpression() ast.Expression {
	p.consume(token.PERIOD)
	expressions := make([]ast.Expression, 0)
	for {
		expr := p.parseAtomicExpression()
		if expr == nil {
			break
		}
		expressions = append(expressions, expr)
	}
	return &ast.CallExpression{Expressions: expressions}
}

// Parses an atomic expression, which is an expression that is not joined by
// infix operators or is a part of an expression list, ie. expressions that are
// prefix, postfix, or literals. Returns nil if the current cursor is not an
// atomic expression.
func (p *Parser) parseAtomicExpression() ast.Expression {
	var expression ast.Expression
	switch p.tok {
	case token.NOT, token.SUB, token.MUL, token.AND:
		tok := p.tok
		pos := p.consume(p.tok)
		expr := p.parseAtomicExpression()
		expression = &ast.Prefix{
			Operator:    tok,
			OperatorPos: pos,
			Expression:  expr,
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
	case token.TRUE:
		expression = p.consumeTrueLiteral()
	case token.FALSE:
		expression = p.consumeFalseLiteral()
	case token.MATCH:
		expression = p.consumeMatchExpression()
	case token.PERIOD:
		expression = p.consumeCallExpression()
	case token.CONSTRUCT:
		expression = p.consumeConstructExpression()
	default:
		return nil
	}

	for {
		switch p.tok {
		case token.INDEX:
			p.consume(token.INDEX)
			identifier := p.consumeIdentifier()
			expression = &ast.IndexExpression{
				Expression: expression,
				Index:      identifier,
			}
		case token.LEFT_BRACK:
			tuple := p.consumeBrackTuple()
			expression = &ast.IndexExpression{
				Expression: expression,
				Index:      tuple,
			}
		default:
			return expression
		}
	}
}

func (p *Parser) consumeConstructExpression() ast.Expression {
	pos := p.consume(token.CONSTRUCT)
	return &ast.ConstructExpression{
		Construct: pos,
		Type:      p.parseAtomicExpression(),
		Value:     p.consumeTuple(),
	}
}

// ---
// Statements

func (p *Parser) consumeFlexStatement() ast.Statement {
	left := p.parseExpression(LOWEST)
	if left == nil {
		return &ast.EmptyStatement{}
	}
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
	tok := p.tok
	pos := p.consume(p.tok)
	var label *ast.Identifier
	if p.tok == token.IDENT {
		label = p.consumeIdentifier()
	}
	return &ast.BranchStatement{
		Keyword:    tok,
		KeywordPos: pos,
		Label:      label,
	}
}

func (p *Parser) consumeForStatement() *ast.ForStatement {
	pos := p.consume(token.FOR)
	clauses := p.parseStatements()
	p.consume(token.RIGHT_ARROW)
	expr := p.parseExpression(LOWEST)
	return &ast.ForStatement{
		For:        pos,
		Clauses:    clauses,
		Expression: expr,
	}
}

// ---
// Expressions

func (p *Parser) consumeTuple() *ast.Tuple {
	lp := p.consume(token.LEFT_PAREN)
	stmts := p.parseStatements()
	rp := p.consume(token.RIGHT_PAREN)
	return &ast.Tuple{
		Left:       lp,
		Statements: stmts,
		Right:      rp,
	}
}

func (p *Parser) consumeReturn() *ast.ReturnStatement {
	pos := p.consume(token.RETURN)
	e := p.parseExpression(LOWEST)
	return &ast.ReturnStatement{
		Return:     pos,
		Expression: e,
	}
}

func (p *Parser) consumeDefaultLiteral() *ast.DefaultLiteral {
	pos := p.consume(token.DEFAULT)
	return &ast.DefaultLiteral{
		DefaultPos: pos,
	}
}

func (p *Parser) consumeBrackTuple() *ast.Tuple {
	lp := p.consume(token.LEFT_BRACK)
	stmts := p.parseStatements()
	rp := p.consume(token.RIGHT_BRACK)
	return &ast.Tuple{
		Left:       lp,
		Statements: stmts,
		Right:      rp,
	}
}

func (p *Parser) consumeBlock() *ast.Block {
	lp := p.consume(token.LEFT_BRACE)
	stmts := p.parseStatements()
	rp := p.consume(token.RIGHT_BRACE)
	return &ast.Block{
		LeftBrace:  lp,
		Statements: stmts,
		RightBrace: rp,
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
		p.appendError("Expected to range literal to have either '[' or '(' to start the range", p.pos, p.pos+1)
	}

	rl.Left = p.parseAtomicExpression()
	p.consume(token.ELIPSIS)
	rl.Right = p.parseAtomicExpression()

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

func (p *Parser) consumeTrueLiteral() *ast.TrueLiteral {
	pos := p.consume(token.TRUE)
	return &ast.TrueLiteral{True: pos}
}

func (p *Parser) consumeFalseLiteral() *ast.FalseLiteral {
	pos := p.consume(token.FALSE)
	return &ast.FalseLiteral{False: pos}
}

func (p *Parser) consumeMatchExpression() *ast.MatchExpression {
	match := p.consume(token.MATCH)
	m := p.parseExpression(LOWEST)
	left := p.consume(token.LEFT_BRACE)
	conditionals := []ast.Expression{}
	bodies := []ast.Statement{}
	i := 0
	for p.tok != token.RIGHT_BRACE && p.tok != token.EOF && p.tok != token.SEMICOLON {
		conditionals = append(conditionals, p.parseExpression(EQUAL))
		p.consume(token.THEN)
		bodies = append(bodies, p.parseStatement())
		p.consume(token.SEMICOLON)
		i += 1
	}
	right := p.consume(token.RIGHT_BRACE)
	return &ast.MatchExpression{
		Match:      match,
		Left:       left,
		Right:      right,
		Item:       m,
		Conditions: conditionals,
		Bodies:     bodies,
	}
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
	if p.tok != token.RIGHT_ARROW {
		fd.Return = p.parseAtomicExpression()
	}
	if p.tok == token.RIGHT_ARROW {
		p.consume(token.RIGHT_ARROW)
		fd.Body = p.parseStatement()
	}
	return fd
}

func (p *Parser) consumeSpread() *ast.Spread {
	pos := p.consume(token.ELIPSIS)
	expr := p.parseAtomicExpression()
	return &ast.Spread{
		Elipsis:    pos,
		Expression: expr,
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
		lit := p.lit
		pos := p.consume(token.COMMENT)
		group.Lines = append(group.Lines, &ast.Comment{
			TextPos: pos,
			Text:    lit,
		})
	}
	return group
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

func (p *Parser) consumeSemi() {
	if p.tok == token.SEMICOLON || p.tok == token.COMMA {
		p.consume(p.tok)
	}
}

func (p *Parser) appendError(msg string, pos token.Pos, end token.Pos) {
	*p.errors = append(*p.errors, StrawError{
		msg: "[parser] " + msg,
		pos: pos,
		end: end,
	})
}
