package strategy

import (
	"sort"

	merror "git.garena.com/shopee/bg-logistics/algo/wms/waveopt/localsearch/pkg/error"
	"git.garena.com/shopee/bg-logistics/algo/wms/waveopt/localsearch/pkg/solver"
)

type BaseSelector struct {
	moveContexts []*solver.MoveContext
}

func (t *BaseSelector) UpdateAtSolverStart(solverContext *solver.SolverContext) error { return nil }

func (t *BaseSelector) UpdateAtSolverEnd(solverContext *solver.SolverContext) error { return nil }

func (t *BaseSelector) UpdateAtStepStart(stepContext *solver.StepContext) error {
	t.moveContexts = nil
	t.moveContexts = make([]*solver.MoveContext, 0)
	return nil
}

func (t *BaseSelector) UpdateAtStepEnd(stepContext *solver.StepContext) error { return nil }

func (t *BaseSelector) UpdateAtMoveStart(moveContext *solver.MoveContext) error { return nil }

func (t *BaseSelector) UpdateAtMoveEnd(moveContext *solver.MoveContext) error { return nil }

func (s *BaseSelector) AddCandidateMove(moveContext *solver.MoveContext) error {
	if moveContext == nil {
		return merror.New("BaseSelector: Nil input move.")
	}

	s.moveContexts = append(s.moveContexts, moveContext)

	return nil
}

type GreedySelector struct {
	*BaseSelector
}

func (s *GreedySelector) Select() (*solver.MoveContext, error) {
	if len(s.moveContexts) == 0 {
		return &solver.MoveContext{
			Move: &solver.NoChangeMove{},
		}, nil
	}

	sort.Slice(s.moveContexts, func(i, j int) bool {
		cmp, _ := s.moveContexts[i].AfterMovingScore.CompareToScore(s.moveContexts[j].AfterMovingScore)
		return cmp > 0
	})

	return s.moveContexts[0], nil
}

func NewGreedySelector() *GreedySelector {
	return &GreedySelector{
		BaseSelector: &BaseSelector{
			moveContexts: make([]*solver.MoveContext, 0),
		},
	}
}

type EpsilonGreedySelector struct {
	*BaseSelector
	epsilon float64
}

func (egs *EpsilonGreedySelector) Select() (*solver.MoveContext, error) {
	if egs.epsilon <= 0 || egs.epsilon > 1 {
		return nil, merror.New("EpsilonGreedySelector: epsilon must be between (0, 1].")
	}

	if len(egs.moveContexts) == 0 {
		return &solver.MoveContext{
			Move: &solver.NoChangeMove{},
		}, nil
	}

	sort.Slice(egs.moveContexts, func(i, j int) bool {
		cmp, _ := egs.moveContexts[i].AfterMovingScore.CompareToScore(egs.moveContexts[j].AfterMovingScore)
		return cmp > 0
	})

	currentBestMoveCtx := egs.moveContexts[0]
	currentBestScore := currentBestMoveCtx.AfterMovingScore
	globalBestScore := currentBestMoveCtx.StepContext.SolverContext.BestScore
	cmp, err := currentBestScore.CompareToScore(globalBestScore)
	if err != nil {
		return nil, err
	}

	if cmp > 0 {
		return currentBestMoveCtx, nil
	}

	for _, moveCtx := range egs.moveContexts {
		if solver.GlobalSolverRand.Float64() > egs.epsilon {
			return moveCtx, nil
		}
	}

	rndIdx := solver.GlobalSolverRand.Intn(len(egs.moveContexts))
	return egs.moveContexts[rndIdx], nil
}

func NewEpsilonGreedSelector(epsilon float64) (*EpsilonGreedySelector, error) {
	if epsilon <= 0 || epsilon > 1 {
		return nil, merror.New("EpsilonGreedySelector: epsilon must be between (0, 1].")
	}

	return &EpsilonGreedySelector{
		epsilon: epsilon,
		BaseSelector: &BaseSelector{
			moveContexts: make([]*solver.MoveContext, 0),
		},
	}, nil
}
