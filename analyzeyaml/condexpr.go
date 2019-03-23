package analyzeyaml

import (
	. "fmt"
	. "github.com/lunfardo314/goq/abstract"
	. "github.com/lunfardo314/goq/qupla"
	. "github.com/lunfardo314/quplayaml/quplayaml"
)

func AnalyzeCondExpr(exprYAML *QuplaCondExprYAML, module ModuleInterface, scope FuncDefInterface) (*QuplaCondExpr, error) {
	module.IncStat("numCond")

	ret := &QuplaCondExpr{
		QuplaExprBase: NewQuplaExprBase(exprYAML.Source),
	}
	if ifExpr, err := module.AnalyzeExpression(exprYAML.If, scope); err != nil {
		return nil, err
	} else {
		ret.AppendSubExpr(ifExpr)
	}
	if ret.NumSubExpr() != 1 {
		return nil, Errorf("condition size must be 1 trit, funDef %v: '%v'", scope.GetName(), ret.GetSource())
	}
	if thenExpr, err := module.AnalyzeExpression(exprYAML.Then, scope); err != nil {
		return nil, err
	} else {
		ret.AppendSubExpr(thenExpr)
	}
	if elseExpr, err := module.AnalyzeExpression(exprYAML.Else, scope); err != nil {
		return nil, err
	} else {
		ret.AppendSubExpr(elseExpr)
	}
	s1 := ret.GetSubExpr(1)
	s2 := ret.GetSubExpr(2)
	if IsNullExpr(s1) && IsNullExpr(s2) {
		return nil, Errorf("can't be both branches null. Dunc def '%v': '%v'", scope.GetName(), ret.GetSource())
	}
	if IsNullExpr(s1) {
		s1.(*QuplaNullExpr).SetSize(s1.Size())
	}
	if IsNullExpr(s2) {
		s2.(*QuplaNullExpr).SetSize(s1.Size())
	}
	return ret, nil
}
