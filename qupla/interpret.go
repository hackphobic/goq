package qupla

import (
	"fmt"
	. "github.com/iotaledger/iota.go/trinary"
	"github.com/lunfardo314/goq/abra"
	. "github.com/lunfardo314/goq/cfg"
	"github.com/lunfardo314/goq/utils"
)

type QuplaSite struct {
	Name     string
	Analyzed bool
	Idx      int
	Offset   int
	Size     int
	SliceEnd int // Offset + size precalculated
	IsState  bool
	IsParam  bool
	Assign   ExpressionInterface
	NumUses  int  // number of times referenced in the scope (by slice expressions)
	NotUsed  bool // optimized away
}

type EvalFrame struct {
	prev    *EvalFrame
	buffer  Trits
	context *FunctionExpr
}

type ExpressionInterface interface {
	Size() int
	Eval(*EvalFrame, Trits) bool
	References(string) bool
	HasState() bool
	Copy() ExpressionInterface // shallow copy
	GetSubexpressions() []ExpressionInterface
	SetSubexpressions([]ExpressionInterface)
	GetSource() string
	GetAbraSite(branch *abra.Branch, codeUnit *abra.CodeUnit, lookupName string) *abra.Site
}

const (
	notEvaluated    = int8(100)
	evaluatedToNull = int8(101)
)

func newEvalFrame(expr *FunctionExpr, prev *EvalFrame) EvalFrame {
	ret := EvalFrame{
		prev:    prev,
		buffer:  make(Trits, expr.FuncDef.BufLen, expr.FuncDef.BufLen),
		context: expr,
	}
	for _, vi := range expr.FuncDef.Sites {
		ret.buffer[vi.Offset] = notEvaluated
	}
	return ret
}

func (vi *QuplaSite) IncNumUses() {
	vi.NumUses++
}

func (frame *EvalFrame) getCallTrace() []uint8 {
	ret := make([]uint8, 0, 40)
	f := frame
	for ; f != nil; f = f.prev {
		if f.prev != nil {
			ret = append(ret, f.context.callIndex)
		} else {
			ret = append(ret, []uint8(frame.context.FuncDef.Name)...)
		}
	}
	return ret
}

func (vi *QuplaSite) GetAbraLookupName() string {
	return "var_site_" + vi.Name
}

func (vi *QuplaSite) Eval(frame *EvalFrame) (Trits, bool) {
	result := frame.buffer[vi.Offset:vi.SliceEnd]
	if result[0] == evaluatedToNull {
		if frame.context.FuncDef.traceLevel > 1 {
			vi.trace(frame, nil, true, true)
		}
		return nil, true
	}
	if result[0] != notEvaluated {
		if frame.context.FuncDef.traceLevel > 1 {
			vi.trace(frame, result, true, false)
		}
		return result, false
	}
	if vi.IsParam {
		// evaluate in the context of the previous call
		if frame.context.subExpr[vi.Idx].Eval(frame.prev, result) {
			result[0] = evaluatedToNull
			if frame.context.FuncDef.traceLevel > 1 {
				vi.trace(frame, nil, false, true)
			}
			return nil, true
		}
		return result, false
	}
	if vi.IsState {
		// for state variables (latches) we return value, retrieved from the key/value storage
		// at the module level.
		// the key is frame.getCallTrace(). It return all 0 f not present
		// but calculated value stays in the buffer
		result = frame.context.FuncDef.StateHashMap.getValue(frame.getCallTrace(), len(result))
		if frame.context.FuncDef.traceLevel > 1 {
			vi.trace(frame, result, false, false)
		}
		return result, false
	}
	if vi.Assign.Eval(frame, result) {
		result[0] = evaluatedToNull
		if frame.context.FuncDef.traceLevel > 1 {
			vi.trace(frame, nil, false, true)
		}
		return nil, true
	}
	if frame.context.FuncDef.traceLevel > 1 {
		vi.trace(frame, result, false, false)
	}
	return result, false
}

func (vi *QuplaSite) trace(frame *EvalFrame, result Trits, cached, null bool) {
	var s string
	if cached {
		s = "cached value "
	} else {
		s = "evaluated value "
	}
	if null {
		s += "null"
	} else {
		//bi, _ := utils.TritsToBigInt(result)
		res := utils.TritsToString(result)
		reslen := len(res)
		if len(res) > 100 {
			res = res[:100] + "..."
		}
		s += fmt.Sprintf("'%v' len=%v", res, reslen)
	}
	if vi.IsState {
		s += fmt.Sprintf(" (state with call trace '%v)' ",
			frame.getCallTrace())
	}
	Logf(frame.context.FuncDef.traceLevel, "trace var %v.%v: %v",
		frame.context.FuncDef.Name, vi.Name, s)
}

func (frame *EvalFrame) SaveStateVariables() {
	if frame == nil || !frame.context.FuncDef.HasStateVariables {
		return
	}
	Logf(7, "SaveStateVariables for '%v'", frame.context.FuncDef.Name)
	var val Trits
	for _, vi := range frame.context.FuncDef.Sites {
		if !vi.IsState {
			continue
		}
		val = make(Trits, vi.Assign.Size(), vi.Assign.Size())
		if !vi.Assign.Eval(frame, val) {
			frame.context.FuncDef.StateHashMap.storeValue(frame.getCallTrace(), val)
		}
	}
}

func MatchSizes(e1, e2 ExpressionInterface) error {
	s1 := e1.Size()
	s2 := e2.Size()

	if s1 != s2 {
		return fmt.Errorf("sizes doesn't match: %v != %v", s1, s2)
	}
	return nil
}

func RequireSize(e ExpressionInterface, size int) error {
	s := e.Size()

	if s != size {
		return fmt.Errorf("sizes doesn't match: required %v != %v", size, s)
	}
	return nil
}
