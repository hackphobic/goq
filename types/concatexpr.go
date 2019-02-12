package types

type QuplaConcatExpr struct {
	Lhs *QuplaExpressionWrapper `yaml:"lhs"`
	Rhs *QuplaExpressionWrapper `yaml:"rhs"`
}

func (e *QuplaConcatExpr) Analyze(module *QuplaModule) error {
	if err := e.Lhs.Analyze(module); err != nil {
		return err
	}
	return e.Rhs.Analyze(module)
}
