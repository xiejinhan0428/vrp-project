package strategy

import (
	merror "git.garena.com/shopee/bg-logistics/algo/wms/waveopt/localsearch/pkg/error"
	"git.garena.com/shopee/bg-logistics/algo/wms/waveopt/localsearch/pkg/solver"
)

// an abstract instance of the Acceptor and SolverLifeCycleListener interface,
// should be embedded to concrete Acceptors.
type BaseAcceptor struct{}

// implements the Acceptor interface
func (a *BaseAcceptor) IsAccepted(moveContext solver.MoveContext) (bool, error) {
	return false, merror.New("No concrete instance of Acceptor.")
}

// implements the SolverLifeCycleListener interface
func (a *BaseAcceptor) UpdateAtSolverStart(solverContext *solver.SolverContext) error { return nil }
func (a *BaseAcceptor) UpdateAtSolverEnd(solverContext *solver.SolverContext) error   { return nil }
func (a *BaseAcceptor) UpdateAtStepStart(stepContext *solver.StepContext) error       { return nil }
func (a *BaseAcceptor) UpdateAtStepEnd(stepContext *solver.StepContext) error         { return nil }
func (a *BaseAcceptor) UpdateAtMoveStart(moveContext *solver.MoveContext) error       { return nil }
func (a *BaseAcceptor) UpdateAtMoveEnd(moveContext *solver.MoveContext) error         { return nil }
