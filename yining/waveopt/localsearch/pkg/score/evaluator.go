package score

import "git.garena.com/shopee/bg-logistics/algo/wms/waveopt/localsearch/pkg/solver"

// users should embed this BaseEvaluator struct in any implemented instances anonymously if
// the concrete evaluator does not plan to listen to the solver life-cycle events.
type BaseEvaluator struct{}

func (e *BaseEvaluator) UpdateAtSolverStart(solverContext *solver.SolverContext) error { return nil }
func (e *BaseEvaluator) UpdateAtStepStart(*solver.StepContext) error                   { return nil }
func (e *BaseEvaluator) UpdateAtMoveStart(*solver.MoveContext) error                   { return nil }
func (e *BaseEvaluator) UpdateAtSolverEnd(*solver.SolverContext) error                 { return nil }
func (e *BaseEvaluator) UpdateAtStepEnd(*solver.StepContext) error                     { return nil }
func (e *BaseEvaluator) UpdateAtMoveEnd(*solver.MoveContext) error                     { return nil }
