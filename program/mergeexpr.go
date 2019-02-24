package program

import (
	"fmt"
	. "github.com/iotaledger/iota.go/trinary"
)

type QuplaMergeExpr struct {
	LhsWrap *QuplaExpressionWrapper `yaml:"lhs"`
	RhsWrap *QuplaExpressionWrapper `yaml:"rhs"`
	//----
	lhsExpr ExpressionInterface
	rhsExpr ExpressionInterface
}

func (e *QuplaMergeExpr) Analyze(module *QuplaModule, scope *QuplaFuncDef) (ExpressionInterface, error) {
	var err error
	module.IncStat("numMergeExpr")

	e.lhsExpr, err = e.LhsWrap.Analyze(module, scope)
	if err != nil {
		return nil, err
	}
	if IsNullExpr(e.lhsExpr) {
		return nil, fmt.Errorf("constant null in merge expression, scope %v", scope.GetName())
	}
	e.rhsExpr, err = e.RhsWrap.Analyze(module, scope)
	if err != nil {
		return nil, err
	}
	if IsNullExpr(e.rhsExpr) {
		return nil, fmt.Errorf("constant null in merge expression, scope %v", scope.GetName())
	}
	if e.lhsExpr.Size() != e.rhsExpr.Size() {
		return nil, fmt.Errorf("operand sizes must be equal in merge expression, scope %v", scope.GetName())
	}
	return e, nil
}

func (e *QuplaMergeExpr) Size() int64 {
	if e == nil {
		return 0
	}
	return e.lhsExpr.Size()
}

func (e *QuplaMergeExpr) Eval(buffer Trits) bool {
	null := e.lhsExpr.Eval(buffer)
	if null {
		return e.rhsExpr.Eval(buffer)
	}
	return false
}
