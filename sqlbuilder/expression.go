package sqlbuilder

import (
	"bytes"
	"github.com/dropbox/godropbox/database/sqltypes"
	"github.com/dropbox/godropbox/errors"
)

// An expression
type Expression interface {
	Clause

	As(alias string) Clause
	IsDistinct(expression Expression) BoolExpression
	IsNull() BoolExpression
}

type expressionInterfaceImpl struct {
	parent Expression
}

func (e *expressionInterfaceImpl) As(alias string) Clause {
	return NewAlias(e.parent, alias)
}

func (e *expressionInterfaceImpl) IsDistinct(expression Expression) BoolExpression {
	return nil
}

func (e *expressionInterfaceImpl) IsNull() BoolExpression {
	return nil
}

// Representation of binary operations (e.g. comparisons, arithmetic)
type binaryExpression struct {
	lhs, rhs Expression
	operator []byte
}

func newBinaryExpression(lhs, rhs Expression, operator []byte, parent ...Expression) binaryExpression {
	binaryExpression := binaryExpression{
		lhs:      lhs,
		rhs:      rhs,
		operator: operator,
	}

	return binaryExpression
}

func (c *binaryExpression) SerializeSql(out *bytes.Buffer, options ...serializeOption) (err error) {
	if c.lhs == nil {
		return errors.Newf("nil lhs.  Generated sql: %s", out.String())
	}
	if err = c.lhs.SerializeSql(out); err != nil {
		return
	}

	_, _ = out.Write(c.operator)

	if c.rhs == nil {
		return errors.Newf("nil rhs.  Generated sql: %s", out.String())
	}
	if err = c.rhs.SerializeSql(out); err != nil {
		return
	}

	return nil
}

// A not expression which negates a expression value
type prefixExpression struct {
	expression Expression
	operator   []byte
}

func newPrefixExpression(expression Expression, operator []byte) prefixExpression {
	prefixExpression := prefixExpression{
		expression: expression,
		operator:   operator,
	}

	return prefixExpression
}

func (p *prefixExpression) SerializeSql(out *bytes.Buffer, options ...serializeOption) (err error) {
	_, _ = out.Write(p.operator)

	if p.expression == nil {
		return errors.Newf("nil prefix expression.  Generated sql: %s", out.String())
	}
	if err = p.expression.SerializeSql(out); err != nil {
		return
	}

	return nil
}

// Representation of n-ary conjunctions (AND/OR)
type conjunctExpression struct {
	expressions []BoolExpression
	conjunction []byte
}

func (conj *conjunctExpression) SerializeSql(out *bytes.Buffer, options ...serializeOption) (err error) {
	if len(conj.expressions) == 0 {
		return errors.Newf(
			"Empty conjunction.  Generated sql: %s",
			out.String())
	}

	clauses := make([]Clause, len(conj.expressions), len(conj.expressions))
	for i, expr := range conj.expressions {
		clauses[i] = expr
	}

	useParentheses := len(clauses) > 1
	if useParentheses {
		_ = out.WriteByte('(')
	}

	if err = serializeClauses(clauses, conj.conjunction, out); err != nil {
		return
	}

	if useParentheses {
		_ = out.WriteByte(')')
	}

	return nil
}

//--------------------------------------------------------------

// Representation of an escaped literal
type literalExpression struct {
	expressionInterfaceImpl
	value sqltypes.Value
}

func NewLiteralExpression(value sqltypes.Value) *literalExpression {
	exp := literalExpression{value: value}
	exp.expressionInterfaceImpl.parent = &exp

	return &exp
}

func (c literalExpression) SerializeSql(out *bytes.Buffer, options ...serializeOption) error {
	sqltypes.Value(c.value).EncodeSql(out)
	return nil
}

//------------------------------------------------------//
// Dummy type for select *
type ColumnList []Column

func (cl ColumnList) SerializeSql(out *bytes.Buffer, options ...serializeOption) error {
	for i, column := range cl {
		err := column.SerializeSql(out)

		if err != nil {
			return err
		}

		if i != len(cl)-1 {
			out.WriteString(", ")
		}
	}
	return nil
}

func (e ColumnList) As(alias string) Clause {
	panic("Invalid usage")
}

func (e ColumnList) IsDistinct(expression Expression) BoolExpression {
	panic("Invalid usage")
}

func (e ColumnList) IsNull(expression Expression) BoolExpression {
	panic("Invalid usage")
}
