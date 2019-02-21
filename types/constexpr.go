package types

import (
	"fmt"
	"github.com/iotaledger/iota.go/trinary"
	"strconv"
	"strings"
)

type ConstExpression interface {
	GetConstValue() int
}

type QuplaConstExpr struct {
	Operator string                  `yaml:"operator"`
	LhsWrap  *QuplaExpressionWrapper `yaml:"lhs"`
	RhsWrap  *QuplaExpressionWrapper `yaml:"rhs"`
}

type QuplaConstTerm struct {
	Operator string                  `yaml:"operator"`
	LhsWrap  *QuplaExpressionWrapper `yaml:"lhs"`
	RhsWrap  *QuplaExpressionWrapper `yaml:"rhs"`
}

type QuplaConstTypeName struct {
	TypeName   string `yaml:"typeName"` // not used
	SizeString string `yaml:"size"`
}

type QuplaConstNumber struct {
	Value string `yaml:"value"`
}

type ConstValue struct {
	Value int64
	size  int64
}

func IsConstExpression(e ExpressionInterface) bool {
	if e == nil {
		return false //???????
	}
	switch e.(type) {
	case *QuplaConstExpr:
		return true
	case *QuplaConstTerm:
		return true
	case *QuplaConstTypeName:
		return true
	case *QuplaConstNumber:
		return true
	case *ConstValue:
		return true
	}
	return false
}

func (e *QuplaConstExpr) Analyze(module *QuplaModule, scope *QuplaFuncDef) (ExpressionInterface, error) {
	var err error
	var lei, rei ExpressionInterface
	var ok bool
	if !strings.Contains("+-", e.Operator) {
		return nil, fmt.Errorf("wrong operator symbol %v", e.Operator)
	}
	if lei, err = e.LhsWrap.Analyze(module, scope); err != nil {
		return nil, err
	}
	if rei, err = e.RhsWrap.Analyze(module, scope); err != nil {
		return nil, err
	}
	if !IsConstExpression(e.LhsWrap) || !IsConstExpression(e.RhsWrap) {
		return nil, fmt.Errorf("operands must be constant expression")
	}
	var rv, lv *ConstValue
	if lv, ok = lei.(*ConstValue); !ok {
		return nil, fmt.Errorf("inconsistency I")
	}
	if rv, ok = rei.(*ConstValue); !ok {
		return nil, fmt.Errorf("inconsistency II")
	}
	var ret *ConstValue
	switch e.Operator {
	case "+":
		ret = NewConstValue(lv.Value + rv.Value)
	case "-":
		ret = NewConstValue(lv.Value - rv.Value)
	}
	return ret, nil
}

func (e *QuplaConstExpr) Size() int64 {
	return 0
}

func (e *QuplaConstExpr) RequireSize(_ int64) error {
	return nil
}

func (e *QuplaConstTerm) Analyze(module *QuplaModule, scope *QuplaFuncDef) (ExpressionInterface, error) {
	var err error
	var lei, rei ExpressionInterface
	var ok bool
	if !strings.Contains("*/%", e.Operator) {
		return nil, fmt.Errorf("wrong operator symbol %v", e.Operator)
	}
	if lei, err = e.LhsWrap.Analyze(module, scope); err != nil {
		return nil, err
	}
	if rei, err = e.RhsWrap.Analyze(module, scope); err != nil {
		return nil, err
	}
	if !IsConstExpression(e.LhsWrap) || !IsConstExpression(e.RhsWrap) {
		return nil, fmt.Errorf("operands must be constant expression")
	}
	var rv, lv *ConstValue
	if lv, ok = lei.(*ConstValue); !ok {
		return nil, fmt.Errorf("inconsistency I")
	}
	if rv, ok = rei.(*ConstValue); !ok {
		return nil, fmt.Errorf("inconsistency II")
	}
	var ret *ConstValue
	switch e.Operator {
	case "*":
		ret = NewConstValue(lv.Value * rv.Value)
	case "/":
		if rv.Value != 0 {
			ret = NewConstValue(lv.Value / rv.Value)
		} else {
			return nil, fmt.Errorf("division by 0 in constant expression")
		}
	case "%":
		if rv.Value != 0 {
			ret = NewConstValue(lv.Value % rv.Value)
		} else {
			return nil, fmt.Errorf("division by 0 in constant expression")
		}
	}
	return ret, nil
}

func (e *QuplaConstTerm) Size() int64 {
	return 0
}

func (e *QuplaConstTerm) RequireSize(_ int64) error {
	return nil
}

func (e *QuplaConstTypeName) Analyze(module *QuplaModule, scope *QuplaFuncDef) (ExpressionInterface, error) {
	var err error
	var ret int
	if ret, err = strconv.Atoi(e.SizeString); err != nil {
		return nil, err
	}
	return NewConstValue(int64(ret)), nil
}

func (e *QuplaConstTypeName) Size() int64 {
	return 0
}

func (e *QuplaConstTypeName) RequireSize(_ int64) error {
	return nil
}

func (e *QuplaConstNumber) Analyze(module *QuplaModule, scope *QuplaFuncDef) (ExpressionInterface, error) {
	ret, err := strconv.Atoi(e.Value)
	if err != nil {
		return nil, err
	}
	return NewConstValue(int64(ret)), nil
}

func (e *QuplaConstNumber) Size() int64 {
	return 0
}

func (e *QuplaConstNumber) RequireSize(_ int64) error {
	return nil
}

func (e *ConstValue) Analyze(module *QuplaModule, scope *QuplaFuncDef) (ExpressionInterface, error) {
	return e, nil
}

func (e *ConstValue) Size() int64 {
	return 0
}

func NewConstValue(value int64) *ConstValue {
	return &ConstValue{
		Value: value,
	}
}

func (e *ConstValue) GetTrits() trinary.Trits {
	t := trinary.IntToTrits(e.Value)
	if e.size == 0 {
		return t
	}
	if e.size == int64(len(t)) {
		return t
	}
	ret := make(trinary.Trits, 0, e.size)
	copy(ret, t)
	return ret
}

func (e *ConstValue) RequireSize(size int64) error {
	t := trinary.IntToTrits(e.Value)
	if size > int64(len(t)) {
		return fmt.Errorf("size mismatch in the conditional expression")
	}
	e.size = size
	return nil
}

func GetConstValue(expr ExpressionInterface) (int64, error) {
	cv, ok := expr.(*ConstValue)
	if !ok {
		return 0, fmt.Errorf("not a constant value")
	}
	return cv.Value, nil
}
