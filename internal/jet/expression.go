package jet

// Expression is common interface for all expressions.
// Can be Bool, Int, Float, String, Date, Time, Timez, Timestamp or Timestampz expressions.
type Expression interface {
	Serializer
	Projection
	GroupByClause
	OrderByClause

	// Test expression whether it is a NULL value.
	IS_NULL() BoolExpression
	// Test expression whether it is a non-NULL value.
	IS_NOT_NULL() BoolExpression

	// Check if this expressions matches any in expressions list
	IN(expressions ...Expression) BoolExpression
	// Check if this expressions is different of all expressions in expressions list
	NOT_IN(expressions ...Expression) BoolExpression

	// The temporary alias name to assign to the expression
	AS(alias string) Projection

	// Expression will be used to sort query result in ascending order
	ASC() OrderByClause
	// Expression will be used to sort query result in ascending order
	DESC() OrderByClause
}

type ExpressionInterfaceImpl struct {
	Parent Expression
}

func (e *ExpressionInterfaceImpl) fromImpl(subQuery SelectTable) Projection {
	return e.Parent
}

func (e *ExpressionInterfaceImpl) IS_NULL() BoolExpression {
	return newPostfixBoolOperatorExpression(e.Parent, "IS NULL")
}

func (e *ExpressionInterfaceImpl) IS_NOT_NULL() BoolExpression {
	return newPostfixBoolOperatorExpression(e.Parent, "IS NOT NULL")
}

func (e *ExpressionInterfaceImpl) IN(expressions ...Expression) BoolExpression {
	return newBinaryBoolOperatorExpression(e.Parent, WRAP(expressions...), "IN")
}

func (e *ExpressionInterfaceImpl) NOT_IN(expressions ...Expression) BoolExpression {
	return newBinaryBoolOperatorExpression(e.Parent, WRAP(expressions...), "NOT IN")
}

func (e *ExpressionInterfaceImpl) AS(alias string) Projection {
	return newAlias(e.Parent, alias)
}

func (e *ExpressionInterfaceImpl) ASC() OrderByClause {
	return newOrderByClause(e.Parent, true)
}

func (e *ExpressionInterfaceImpl) DESC() OrderByClause {
	return newOrderByClause(e.Parent, false)
}

func (e *ExpressionInterfaceImpl) serializeForGroupBy(statement StatementType, out *SQLBuilder) {
	e.Parent.serialize(statement, out, noWrap)
}

func (e *ExpressionInterfaceImpl) serializeForProjection(statement StatementType, out *SQLBuilder) {
	e.Parent.serialize(statement, out, noWrap)
}

func (e *ExpressionInterfaceImpl) serializeForOrderBy(statement StatementType, out *SQLBuilder) {
	e.Parent.serialize(statement, out, noWrap)
}

// Representation of binary operations (e.g. comparisons, arithmetic)
type binaryOperatorExpression struct {
	ExpressionInterfaceImpl

	lhs, rhs        Serializer
	additionalParam Serializer
	operator        string
}

func newBinaryOperatorExpression(lhs, rhs Serializer, operator string, additionalParam ...Expression) *binaryOperatorExpression {
	binaryExpression := &binaryOperatorExpression{
		lhs:      lhs,
		rhs:      rhs,
		operator: operator,
	}

	if len(additionalParam) > 0 {
		binaryExpression.additionalParam = additionalParam[0]
	}

	binaryExpression.ExpressionInterfaceImpl.Parent = binaryExpression

	return binaryExpression
}

func (c *binaryOperatorExpression) serialize(statement StatementType, out *SQLBuilder, options ...SerializeOption) {
	if c.lhs == nil {
		panic("jet: lhs is nil for '" + c.operator + "' operator")
	}
	if c.rhs == nil {
		panic("jet: rhs is nil for '" + c.operator + "' operator")
	}

	wrap := !contains(options, noWrap)

	if wrap {
		out.WriteString("(")
	}

	if serializeOverride := out.Dialect.OperatorSerializeOverride(c.operator); serializeOverride != nil {
		serializeOverrideFunc := serializeOverride(c.lhs, c.rhs, c.additionalParam)
		serializeOverrideFunc(statement, out, options...)
	} else {
		c.lhs.serialize(statement, out)
		out.WriteString(c.operator)
		c.rhs.serialize(statement, out)
	}

	if wrap {
		out.WriteString(")")
	}
}

// A prefix operator Expression
type prefixExpression struct {
	ExpressionInterfaceImpl

	expression Expression
	operator   string
}

func newPrefixOperatorExpression(expression Expression, operator string) *prefixExpression {
	prefixExpression := &prefixExpression{
		expression: expression,
		operator:   operator,
	}
	prefixExpression.ExpressionInterfaceImpl.Parent = prefixExpression

	return prefixExpression
}

func (p *prefixExpression) serialize(statement StatementType, out *SQLBuilder, options ...SerializeOption) {
	out.WriteString("(")
	out.WriteString(p.operator)

	if p.expression == nil {
		panic("jet: nil prefix expression in prefix operator " + p.operator)
	}

	p.expression.serialize(statement, out)

	out.WriteString(")")
}

// A postfix operator Expression
type postfixOpExpression struct {
	ExpressionInterfaceImpl

	expression Expression
	operator   string
}

func newPostfixOperatorExpression(expression Expression, operator string) *postfixOpExpression {
	postfixOpExpression := &postfixOpExpression{
		expression: expression,
		operator:   operator,
	}

	postfixOpExpression.ExpressionInterfaceImpl.Parent = postfixOpExpression

	return postfixOpExpression
}

func (p *postfixOpExpression) serialize(statement StatementType, out *SQLBuilder, options ...SerializeOption) {
	if p.expression == nil {
		panic("jet: nil prefix expression in postfix operator " + p.operator)
	}

	p.expression.serialize(statement, out)

	out.WriteString(p.operator)
}
