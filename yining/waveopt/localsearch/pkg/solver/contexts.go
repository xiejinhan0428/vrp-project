package solver

import (
	"time"

	merror "git.garena.com/shopee/bg-logistics/algo/wms/waveopt/localsearch/pkg/error"
)

// internal interfaces

// MoveContext context of a move
type MoveContext struct {
	StepContext *StepContext

	MoveCount       int
	StartTime       time.Time
	SolvingProgress float64

	BeforeMovingScore Score
	Move              Move
	AfterMovingScore  Score
}

// StepContext context of a step which contains multiple moves
type StepContext struct {
	SolverContext *SolverContext

	StepCount       int
	StartTime       time.Time
	SolvingProgress float64

	BeforeStepScore Score
	WinningMove     Move
	AfterStepScore  Score

	BeforeStepSolution Solution
	AfterStepSolution  Solution

	MoveContexts []*MoveContext
}

func (sc *StepContext) NewMoveContext(moveCount int) *MoveContext {
	mc := MoveContext{}
	mc.StepContext = sc
	mc.MoveCount = moveCount
	mc.StartTime = time.Now()
	mc.BeforeMovingScore = sc.BeforeStepScore

	sc.MoveContexts = append(sc.MoveContexts, &mc)

	return &mc
}

// SolverContext context of a solver which contains multiple steps
type SolverContext struct {
	Solver Solver

	InitialScore Score
	BestScore    Score

	InitialSolution Solution
	StartTime       time.Time

	BestSolution          Solution
	BestSolutionStepCount int
	BestSolutionTime      time.Time

	StepContexts []*StepContext
}

func (sc *SolverContext) LastStepScore() Score {
	// do not get the last step's score by simply accessing solverContext.stepContexts[len - 1],
	// because the solverContext.stepContexts[len - 1] may be a freshly created stepContext that has not been evaluated.
	// instead, we can iterate over solverContext.stepContexts from the end, and returen the first non-nil stepContext.afterStepScore
	size := len(sc.StepContexts)
	if size > 0 {
		for i := size - 1; i >= 0; i-- {
			if sc.StepContexts[i].AfterStepScore != nil {
				return sc.StepContexts[i].AfterStepScore
			}
		}
	}

	return sc.InitialScore
}

func (sc *SolverContext) LastStepSolution() (Solution, error) {
	size := len(sc.StepContexts)
	if size > 0 {
		for i := size - 1; i >= 0; i-- {
			if sc.StepContexts[i].AfterStepSolution != nil {
				solutionCopy, err := sc.StepContexts[i].AfterStepSolution.Copy()
				if err != nil {
					return nil, merror.New("SolverContext: Error in copying solution").CausedBy(err)
				}
				return solutionCopy, nil
			}
		}
	}

	return sc.InitialSolution, nil
}

func (sc *SolverContext) NewStepContext(stepCount int) (*StepContext, error) {
	stepCtx := StepContext{}
	stepCtx.SolverContext = sc
	stepCtx.StepCount = stepCount
	stepCtx.StartTime = time.Now()
	stepCtx.MoveContexts = make([]*MoveContext, 0)
	sc.StepContexts = append(sc.StepContexts, &stepCtx)
	stepCtx.BeforeStepScore = sc.LastStepScore()
	lastStepSolution, err := sc.LastStepSolution()
	if err != nil {
		return nil, merror.New("SolverContext: Cannot get last step's solution").CausedBy(err)
	}
	stepCtx.BeforeStepSolution = lastStepSolution
	stepCtx.SolvingProgress = 0.0

	return &stepCtx, nil
}
