package qupla

import (
	"fmt"
	. "github.com/iotaledger/iota.go/trinary"
)

type ExpressionFactory interface {
	AnalyzeExpression(interface{}, *QuplaModule, *QuplaFuncDef) (ExpressionInterface, error)
}

type ExpressionInterface interface {
	Size() int64
	Eval(*CallFrame, Trits) bool
}

type QuplaTypeDef struct {
	Size   string                            `yaml:"end"`
	Fields map[string]*struct{ Size string } `yaml:"fields,omitempty"`
}

type QuplaNullExpr struct {
	size int64
}

func IsNullExpr(e interface{}) bool {
	_, ok := e.(*QuplaNullExpr)
	return ok
}

func (e *QuplaNullExpr) Analyze(module *QuplaModule, scope *QuplaFuncDef) (ExpressionInterface, error) {
	return e, nil
}

func (e *QuplaNullExpr) Size() int64 {
	return e.size
}

func (e *QuplaNullExpr) Eval(_ *CallFrame, _ Trits) bool {
	return true
}

func (e *QuplaNullExpr) SetSize(size int64) {
	e.size = size
}

func MatchSizes(e1, e2 ExpressionInterface) error {
	s1 := e1.Size()
	s2 := e2.Size()

	if s1 != s2 {
		return fmt.Errorf("sizes doesn't match: %v != %v", s1, s2)
	}
	return nil
}

func RequireSize(e ExpressionInterface, size int64) error {
	s := e.Size()

	if s != size {
		return fmt.Errorf("sizes doesn't match: required %v != %v", size, s)
	}
	return nil
}