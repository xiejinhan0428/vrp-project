package solver

// a single-thread implementation of Solver interface
import (
	"fmt"
	"git.garena.com/shopee/bg-logistics/go/gocommon/logger"
	"time"

	merror "git.garena.com/shopee/bg-logistics/algo/wms/waveopt/localsearch/pkg/error"
)

type SerialSolver struct {
	name               string
	lifeCycleListeners []SolverLifeCycleListener

	solution Solution

	params *Params

	acceptors   []Acceptor
	selector    Selector
	terminators []Terminator

	evaluator Evaluator

	moveFactoryScheduler MoveFactoryScheduler

	isLoggingToConsole bool

	asyncTerminate chan int
	isStopped      bool
}

// implements the Solver interface

func (ss *SerialSolver) Solve() (Solution, Score, error) {
	ss.isStopped = false
	defer func() {
		ss.isStopped = true
	}()
	solverCtx, err := ss.newSolverContext()
	if err != nil {
		return nil, nil, merror.New("SerialSolver: Fail to init solver context").CausedBy(err)
	}

	err = ss.updateAtSolverStart(solverCtx)
	if err != nil {
		return solverCtx.BestSolution, solverCtx.BestScore, merror.New("SerialSolver: Fail to update at solving start").CausedBy(err)
	}

	isStepTerminated := false
	for stepCount := 0; ; stepCount++ {
		isTerminated, err := ss.isTerminated()
		if err != nil {
			return solverCtx.BestSolution, solverCtx.BestScore, merror.New("SerialSolver: Unexpected termination").CausedBy(err)
		}
		if isTerminated {
			break
		}

		stepCtx, err := solverCtx.NewStepContext(stepCount)
		if err != nil {
			return solverCtx.BestSolution, solverCtx.BestScore, merror.New("SerialSolver: Fail to create new step").CausedBy(err)
		}
		err = ss.updateAtStepStart(stepCtx)
		if err != nil {
			return solverCtx.BestSolution, solverCtx.BestScore, merror.New("SerialSolver: Fail to update at step start").CausedBy(err)
		}

		stepSolution, err := stepCtx.BeforeStepSolution.Copy()
		if err != nil {
			return solverCtx.BestSolution, solverCtx.BestScore, merror.New("SerialSolver: Fail to update at step start").CausedBy(err)
		}

		for moveCount := 0; moveCount < ss.params.MovesPerStep; moveCount++ {
			moveCtx := stepCtx.NewMoveContext(moveCount)
			err = ss.updateAtMoveStart(moveCtx)
			if err != nil {
				return solverCtx.BestSolution, solverCtx.BestScore, err
			}

			isTerminated, err := ss.isTerminated()
			if err != nil {
				return solverCtx.BestSolution, solverCtx.BestScore, merror.New("SerialSolver: Unexpected termination").CausedBy(err)
			}
			if isTerminated {
				isStepTerminated = true
				break
			}

			moveFactory, err := ss.moveFactoryScheduler.SelectMoveFactory()
			if err != nil {
				return solverCtx.BestSolution, solverCtx.BestScore, err
			}
			move, err := moveFactory.CreateMove(stepSolution)
			moveCtx.Move = move
			if err != nil {
				return solverCtx.BestSolution, solverCtx.BestScore, err
			}
			reverseMove, err := move.Do(stepSolution)
			if err != nil {
				return solverCtx.BestSolution, solverCtx.BestScore, err
			}
			moveScore, err := ss.evaluator.Evaluate(stepSolution)
			if err != nil {
				return solverCtx.BestSolution, solverCtx.BestScore, err
			}
			moveCtx.AfterMovingScore = moveScore
			isMoveAccepted, err := ss.acceptMove(moveCtx)
			if err != nil {
				return solverCtx.BestSolution, solverCtx.BestScore, merror.New("SerialSolver: Error in accepting move", fmt.Sprint(move)).CausedBy(err)
			}
			if isMoveAccepted {
				err = ss.selector.AddCandidateMove(moveCtx)
				if err != nil {
					return solverCtx.BestSolution, solverCtx.BestScore, merror.New("SerialSolver: Error in adding candidate move to selector").CausedBy(err)
				}
			}
			_, err = reverseMove.Do(stepSolution)
			if err != nil {
				return solverCtx.BestSolution, solverCtx.BestScore, merror.New("SerialSolver: Error in doing move", fmt.Sprint(reverseMove)).CausedBy(err)
			}

			err = ss.updateAtMoveEnd(moveCtx)
			if err != nil {
				return solverCtx.BestSolution, solverCtx.BestScore, merror.New("SerialSolver: Fail to update at move end").CausedBy(err)
			}
		}

		if isStepTerminated {
			break
		}

		winningMoveCtx, err := ss.selector.Select()
		if err != nil {
			return solverCtx.BestSolution, solverCtx.BestScore, merror.New("SerialSolver: Error in selecting the winning move").CausedBy(err)
		}
		winningMove := winningMoveCtx.Move
		_, err = winningMove.Do(stepSolution)
		if err != nil {
			return solverCtx.BestSolution, solverCtx.BestScore, merror.New("SerialSolver: Error in doing winning move").CausedBy(err)
		}
		winningScore, err := ss.evaluator.Evaluate(stepSolution)
		if err != nil {
			return solverCtx.BestSolution, solverCtx.BestScore, merror.New("SerialSolver: Error in evaluating the winning move").CausedBy(err)
		}

		stepCtx.WinningMove = winningMove
		stepCtx.AfterStepScore = winningScore
		stepCtx.AfterStepSolution = stepSolution

		err = ss.updateAtStepEnd(stepCtx)
		if err != nil {
			return solverCtx.BestSolution, solverCtx.BestScore, merror.New("SerialSolver: Fail to update at step end").CausedBy(err)
		}

		if ss.isLoggingToConsole && stepCount%100 == 0 {
			logger.LogInfof("WaveOptAlgo - %v  %v - Best Score: %v  Step(%v)-%.0f%%  Score: %v  Winning Move: %v \n", time.Now().Format("2006-01-02 15:04:05"), ss.name, solverCtx.BestScore, stepCount, stepCtx.SolvingProgress*100.0, winningScore, winningMove)
		}
	}

	err = ss.updateAtSolverEnd(solverCtx)
	if err != nil {
		return solverCtx.BestSolution, solverCtx.BestScore, merror.New("SerialSolver: Fail to update at solver end").CausedBy(err)
	}

	return solverCtx.BestSolution, solverCtx.BestScore, nil
}

func (ss *SerialSolver) AsyncTerminate() {
	if ss.isStopped {
		if ss.isLoggingToConsole {
			logger.LogDebug("WaveOptSolver - The solver is already terminated, cannot be interrupted again.")
		}
	} else {
		if ss.isLoggingToConsole {
			logger.LogDebug("WaveOptSolver - Attempt to terminate the solver early...")
		}
		ss.asyncTerminate <- 0
	}
}

func (ss *SerialSolver) newSolverContext() (*SolverContext, error) {
	solverStartTime := time.Now()
	initScore, err := ss.evaluator.Evaluate(ss.solution)
	if err != nil {
		return nil, err
	}

	solutionCopy, err := ss.solution.Copy()
	if err != nil {
		return nil, err
	}

	sc := &SolverContext{
		Solver:           ss,
		InitialScore:     initScore,
		BestScore:        initScore,
		StepContexts:     make([]*StepContext, 0),
		InitialSolution:  solutionCopy,
		BestSolution:     solutionCopy,
		StartTime:        solverStartTime,
		BestSolutionTime: solverStartTime,
	}

	return sc, nil
}

func (ss *SerialSolver) isTerminated() (bool, error) {
	select {
	case <-ss.asyncTerminate:
		return true, nil
	default:
		for _, terminator := range ss.terminators {
			isTerminated, err := terminator.IsTerminated()
			if err != nil {
				return true, err
			}
			if isTerminated {
				return true, nil
			}
		}

		return false, nil
	}
}

func (ss *SerialSolver) acceptMove(moveContext *MoveContext) (bool, error) {
	for _, acceptor := range ss.acceptors {
		isAccepted, err := acceptor.IsAccepted(moveContext)
		if err != nil {
			return false, err
		}
		if !isAccepted {
			return false, nil
		}
	}

	return true, nil
}

func (ss *SerialSolver) updateAtSolverStart(solverContext *SolverContext) error {
	for _, listener := range ss.lifeCycleListeners {
		err := listener.UpdateAtSolverStart(solverContext)
		if err != nil {
			return err
		}
	}

	return nil
}

func (ss *SerialSolver) updateAtSolverEnd(solverContext *SolverContext) error {
	for _, listener := range ss.lifeCycleListeners {
		err := listener.UpdateAtSolverEnd(solverContext)
		if err != nil {
			return err
		}
	}

	return nil
}

func (ss *SerialSolver) updateAtStepStart(stepContext *StepContext) error {
	for _, terminator := range ss.terminators {
		progress, err := terminator.SolvingProgress(stepContext)
		if err != nil {
			return err
		}
		if progress >= stepContext.SolvingProgress {
			stepContext.SolvingProgress = progress
		}
	}

	for _, listener := range ss.lifeCycleListeners {
		err := listener.UpdateAtStepStart(stepContext)
		if err != nil {
			return err
		}
	}

	return nil
}

func (ss *SerialSolver) updateAtStepEnd(stepContext *StepContext) error {
	cmp, err := stepContext.AfterStepScore.CompareToScore(stepContext.SolverContext.BestScore)
	if err != nil {
		return err
	}
	if cmp > 0 {
		stepContext.SolverContext.BestScore = stepContext.AfterStepScore
		stepContext.SolverContext.BestSolution, err = stepContext.AfterStepSolution.Copy()
		if err != nil {
			return err
		}
		stepContext.SolverContext.BestSolutionStepCount = stepContext.StepCount
		stepContext.SolverContext.BestSolutionTime = time.Now()
	}

	for _, listener := range ss.lifeCycleListeners {
		err := listener.UpdateAtStepEnd(stepContext)
		if err != nil {
			return err
		}
	}

	return nil
}

func (ss *SerialSolver) updateAtMoveStart(moveContext *MoveContext) error {
	for _, listener := range ss.lifeCycleListeners {
		err := listener.UpdateAtMoveStart(moveContext)
		if err != nil {
			return err
		}
	}

	return nil
}

func (ss *SerialSolver) updateAtMoveEnd(moveContext *MoveContext) error {
	for _, listener := range ss.lifeCycleListeners {
		err := listener.UpdateAtMoveEnd(moveContext)
		if err != nil {
			return err
		}
	}

	return nil
}
