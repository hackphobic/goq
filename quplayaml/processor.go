package quplayaml

import . "github.com/iotaledger/iota.go/trinary"
import . "github.com/lunfardo314/goq/abstract"

type CallFrame struct {
	context   *QuplaFuncExpr // which function called
	parent    *CallFrame     // context where it was called
	buffer    Trits          // buffer to place all params and variables
	evaluated []bool         // flag if respective variable was evaluated
	isNull    []bool         // flag if value was evaluated to null
	result    Trits          // slice where to put result
}

type StackProcessor struct {
	curFrame *CallFrame
}

func NewStackProcessor() *StackProcessor {
	return &StackProcessor{}
}

func (proc *StackProcessor) Eval(expr ExpressionInterface, result Trits) bool {
	funExpr, isFunction := expr.(*QuplaFuncExpr)
	if isFunction {
		proc.curFrame = funExpr.NewCallFrame(proc.curFrame)
	}
	null := expr.Eval(proc, result)
	if isFunction {
		proc.curFrame = proc.curFrame.parent
	}
	return null
}

func (proc *StackProcessor) EvalVar(idx int64) bool {
	var null bool
	if proc.curFrame == nil {
		panic("variable can't be evaluated in nil context")
	}
	vi := proc.curFrame.context.funcDef.VarByIdx(idx)
	if vi == nil {
		panic("wrong var idx")
	}

	res := proc.curFrame.buffer[vi.Offset : vi.Offset+vi.Size]
	if vi.IsParam {
		saveCurFrame := proc.curFrame
		proc.curFrame = proc.curFrame.parent
		null = proc.Eval(proc.curFrame.context.args[vi.Idx], res)
		proc.curFrame = saveCurFrame
	} else {
		null = proc.Eval(vi.Expr, res)
	}

	return null
}

func (proc *StackProcessor) Slice(offset, size int64) Trits {
	return proc.curFrame.buffer[offset : offset+size]
}
