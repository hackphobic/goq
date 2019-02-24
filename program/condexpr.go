package program

import (
	"fmt"
	"github.com/iotaledger/iota.go/trinary"
)

type QuplaCondExpr struct {
	If   *QuplaExpressionWrapper `yaml:"if"`
	Then *QuplaExpressionWrapper `yaml:"then"`
	Else *QuplaExpressionWrapper `yaml:"else"`
	//--
	ifExpr   ExpressionInterface
	thenExpr ExpressionInterface
	elseExpr ExpressionInterface
}

func (e *QuplaCondExpr) Analyze(module *QuplaModule, scope *QuplaFuncDef) (ExpressionInterface, error) {
	var err error
	module.IncStat("numCond")

	if e.ifExpr, err = e.If.Analyze(module, scope); err != nil {
		return nil, err
	}
	if e.ifExpr.Size() != 1 {
		return nil, fmt.Errorf("condition size must be 1 trit: scope %v", scope.GetName())
	}
	if e.thenExpr, err = e.Then.Analyze(module, scope); err != nil {
		return nil, err
	}
	if e.elseExpr, err = e.Else.Analyze(module, scope); err != nil {
		return nil, err
	}
	if IsNullExpr(e.thenExpr) && IsNullExpr(e.elseExpr) {
		return nil, fmt.Errorf("can't be both branches null: scope %v", scope.GetName())
	}
	if IsNullExpr(e.thenExpr) {
		e.thenExpr.(*QuplaNullExpr).SetSize(e.elseExpr.Size())
	}
	if IsNullExpr(e.elseExpr) {
		e.elseExpr.(*QuplaNullExpr).SetSize(e.thenExpr.Size())
	}
	return e, nil
}

func (e *QuplaCondExpr) Size() int64 {
	if e == nil {
		return 0
	}
	return e.thenExpr.Size()
}

func (e *QuplaCondExpr) Eval(callFrame *CallFrame, result trinary.Trits) bool {
	var buf [1]int8
	null := e.ifExpr.Eval(callFrame, buf[:])
	if null {
		return true
	}
	switch buf[0] {
	case 1:
		return e.thenExpr.Eval(callFrame, result)
	case -1:
		return e.elseExpr.Eval(callFrame, result)
	}
	panic("trit value")
}
