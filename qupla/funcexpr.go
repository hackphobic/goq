package qupla

import (
	"fmt"
	. "github.com/iotaledger/iota.go/trinary"
	. "github.com/lunfardo314/goq/abstract"
	. "github.com/lunfardo314/goq/quplayaml"
)

type QuplaFuncExpr struct {
	QuplaExprBase
	source  string
	name    string
	funcDef *QuplaFuncDef
}

func AnalyzeFuncExpr(exprYAML *QuplaFuncExprYAML, module ModuleInterface, scope FuncDefInterface) (*QuplaFuncExpr, error) {
	var err error
	ret := &QuplaFuncExpr{
		QuplaExprBase: NewQuplaExprBase(exprYAML.Source),
		name:          exprYAML.Name,
	}
	var fdi FuncDefInterface
	fdi, err = module.FindFuncDef(exprYAML.Name)
	if err != nil {
		return nil, err
	}
	var ok bool
	if ret.funcDef, ok = fdi.(*QuplaFuncDef); !ok {
		return nil, fmt.Errorf("inconsistency with types in %v", exprYAML.Name)
	}

	var fe ExpressionInterface
	module.IncStat("numFuncExpr")

	for _, arg := range exprYAML.Args {
		if fe, err = module.AnalyzeExpression(arg, scope); err != nil {
			return nil, err
		}
		ret.AppendSubExpr(fe)
	}
	err = ret.funcDef.checkArgSizes(ret.subexpr)
	return ret, err
}

func (e *QuplaFuncExpr) HasState() bool {
	if e.funcDef.HasState() {
		return true
	}
	for _, arg := range e.subexpr {
		if arg.HasState() {
			return true
		}
	}
	return false
}

func (e *QuplaFuncExpr) Size() int64 {
	if e == nil {
		return 0
	}
	return e.funcDef.Size()
}

func (e *QuplaFuncExpr) NewCallFrame(parent *CallFrame) *CallFrame {
	numVars := len(e.funcDef.localVars)
	return &CallFrame{
		context:  e,
		parent:   parent,
		buffer:   make(Trits, e.funcDef.bufLen, e.funcDef.bufLen),
		valueTag: make([]uint8, numVars, numVars),
	}
}

func (e *QuplaFuncExpr) Eval(proc ProcessorInterface, result Trits) bool {
	return proc.Eval(e.funcDef.retExpr, result)
}
