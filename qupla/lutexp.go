package qupla

import (
	. "github.com/iotaledger/iota.go/trinary"
)

type LutExpr struct {
	ExpressionBase
	LutDef *LutDef
}

func (e *LutExpr) Size() int {
	if e == nil {
		return 0
	}
	return e.LutDef.Size()
}

func (e *LutExpr) InlineCopy(funExpr *FunctionExpr) ExpressionInterface {
	return &LutExpr{
		ExpressionBase: e.inlineCopyBase(funExpr),
		LutDef:         e.LutDef,
	}
}

func (e *LutExpr) Eval(frame *EvalFrame, result Trits) bool {
	var buf [3]int8 // no more than 3 inputs
	for i, a := range e.subExpr {
		if a.Eval(frame, buf[i:i+1]) {
			return true
		}
	}
	lutArg := buf[:e.LutDef.InputSize]
	null := e.LutDef.Lookup(result, lutArg)
	return null
}
