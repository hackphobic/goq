package optimize

import . "github.com/lunfardo314/goq/qupla"

// tree of concats optimized into one concat of many expressions

func optimizeConcats(def *Function, stats map[string]int) bool {
	before := StatValue("numOptimizedConcats", stats)
	for _, site := range def.LocalVars {
		if site.NotUsed || site.IsState || site.IsParam || site.NumUses > 1 {
			continue
		}
		site.Assign = optimizeConcatsInExpr(site.Assign, stats)
	}
	def.RetExpr = optimizeConcatsInExpr(def.RetExpr, stats)
	return before != StatValue("numOptimizedConcats", stats)
}

func optimizeConcatsInExpr(expr ExpressionInterface, stats map[string]int) ExpressionInterface {
	_, ok := expr.(*ConcatExpr)
	if !ok {
		return transformSubexpressions(expr, func(se ExpressionInterface) ExpressionInterface {
			return optimizeConcatsInExpr(se, stats)
		})
	}
	subExpr := make([]ExpressionInterface, 0)
	for _, se := range expr.GetSubexpressions() {
		oe := optimizeConcatsInExpr(se, stats)
		if ce, ok := oe.(*ConcatExpr); ok {
			IncStat("numOptimizedConcats", stats)
			for _, e := range ce.GetSubexpressions() {
				subExpr = append(subExpr, e)
			}
		} else {
			subExpr = append(subExpr, oe)
		}
	}
	return NewConcatExpression("optimized", subExpr)
}
