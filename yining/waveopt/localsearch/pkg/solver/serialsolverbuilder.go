package solver

import (
	"sort"

	merror "git.garena.com/shopee/bg-logistics/algo/wms/waveopt/localsearch/pkg/error"
)

// builder of SerialSolver, all fields are the same
// fields with zero values in SerialSolverBuilder will be replaced by a default value defined in the builder

const DefaultMovesPerStep = 500
const DefaultSolverRandSeed = 2021112920211129

type SerialSolverBuilder struct {
	name string

	solution Solution

	params *Params

	acceptors   []Acceptor
	selector    Selector
	terminators []Terminator

	evaluator Evaluator

	moveFactories          []MoveFactory
	moveFactoryToWeightMap map[MoveFactory]float64
	moveFactoryScheduler   MoveFactoryScheduler

	seed         int64
	isStableRand bool
	isSeedSet    bool

	isLoggingToConsole bool
}

func NewSerialSolverBuilder() *SerialSolverBuilder {
	return &SerialSolverBuilder{
		acceptors:              make([]Acceptor, 0),
		terminators:            make([]Terminator, 0),
		moveFactoryToWeightMap: make(map[MoveFactory]float64),
	}
}

func (ssb *SerialSolverBuilder) WithName(name string) *SerialSolverBuilder {
	ssb.name = name
	return ssb
}

func (ssb *SerialSolverBuilder) WithSolution(solution Solution) *SerialSolverBuilder {
	ssb.solution = solution
	return ssb
}

func (ssb *SerialSolverBuilder) WithParams(params *Params) *SerialSolverBuilder {
	ssb.params = params
	return ssb
}

func (ssb *SerialSolverBuilder) WithAcceptor(acceptor Acceptor) *SerialSolverBuilder {
	ssb.acceptors = append(ssb.acceptors, acceptor)
	return ssb
}

func (ssb *SerialSolverBuilder) WithSelector(selector Selector) *SerialSolverBuilder {
	ssb.selector = selector
	return ssb
}

func (ssb *SerialSolverBuilder) WithTerminator(terminator Terminator) *SerialSolverBuilder {
	ssb.terminators = append(ssb.terminators, terminator)
	return ssb
}

func (ssb *SerialSolverBuilder) WithEvaluator(evaluator Evaluator) *SerialSolverBuilder {
	ssb.evaluator = evaluator
	return ssb
}

func (ssb *SerialSolverBuilder) WithRouletteMoveFactories(moveFactoryToWeightMap map[MoveFactory]float64) *SerialSolverBuilder {
	ssb.moveFactoryToWeightMap = moveFactoryToWeightMap

	return ssb
}

func (ssb *SerialSolverBuilder) WithSequentialMoveFactories(moveFactories []MoveFactory) *SerialSolverBuilder {
	ssb.moveFactories = moveFactories

	return ssb
}

func (ssb *SerialSolverBuilder) WithCustomizedMoveFactoryScheduler(moveFactoryScheduler MoveFactoryScheduler) *SerialSolverBuilder {
	ssb.moveFactoryScheduler = moveFactoryScheduler

	return ssb
}

func (ssb *SerialSolverBuilder) WithStableRandSeed(seed int64) *SerialSolverBuilder {
	ssb.seed = seed
	ssb.isStableRand = true
	ssb.isSeedSet = true
	return ssb
}

func (ssb *SerialSolverBuilder) IsLoggingToConsole(flag bool) *SerialSolverBuilder {
	ssb.isLoggingToConsole = flag
	return ssb
}

func (ssb *SerialSolverBuilder) Build() (*SerialSolver, error) {
	var err error

	lifeCycleListeners := make([]SolverLifeCycleListener, 0)

	if ssb.solution == nil {
		return nil, merror.New("SerialSolverBuilder: Nil solution for SerialSolver.")
	}

	if ssb.params == nil {
		ssb.params = &Params{
			MovesPerStep: DefaultMovesPerStep,
		}
	} else {
		if ssb.params.MovesPerStep <= 0 {
			return nil, merror.New("SerialSolverBuilder: Non-positive MovesPerStep for SerialSolver, must be positive.")
		}
	}

	if len(ssb.acceptors) == 0 {
		return nil, merror.New("SerialSolverBuilder: Nil Acceptor for SerialSolver.")
	}

	if ssb.selector == nil {
		return nil, merror.New("SerialSolverBuilder, Nil Selector for SerialSolver.")
	}

	if len(ssb.terminators) == 0 {
		return nil, merror.New("SerialSolverBuilder: Nil Terminator fro SerialSolver.")
	}

	if ssb.evaluator == nil {
		return nil, merror.New("SerialSolverBuilder: Nil Evaluator for SerialSolver.")
	}

	if len(ssb.moveFactoryToWeightMap) == 0 && len(ssb.moveFactories) == 0 && ssb.moveFactoryScheduler == nil {
		return nil, merror.New("SerialSolverBuilder: No moveFactory provided")
	}

	if len(ssb.moveFactoryToWeightMap) != 0 && len(ssb.moveFactories) != 0 {
		return nil, merror.New("SerialSolverBuilder: Both roulette and sequential MoveFactory scheduler is specified, consider to split them into 2 solvers.")
	}

	var moveFactoryScheduler MoveFactoryScheduler
	if len(ssb.moveFactories) != 0 {
		moveFactoryScheduler, err = NewSequentialMoveFactoryScheduler(ssb.moveFactories)
		if err != nil {
			return nil, merror.New("SerialSolverBuilder: Fail to init the MoveFactoryScheduler").CausedBy(err)
		}
	}
	if len(ssb.moveFactoryToWeightMap) != 0 {
		moveFactoryScheduler, err = NewRouletteMoveFactoryScheduler(ssb.moveFactoryToWeightMap)
		if err != nil {
			return nil, merror.New("SerialSolverBuilder: Fail to init the MoveFactoryScheduler").CausedBy(err)
		}
	}

	if moveFactoryScheduler != nil && ssb.moveFactoryScheduler != nil {
		return nil, merror.New("SerialSolverBuilder: Customized MoveFactoryScheduler will override the built-in one")
	}

	if moveFactoryScheduler == nil && ssb.moveFactoryScheduler != nil {
		moveFactoryScheduler = ssb.moveFactoryScheduler
	}

	// add SolverLifeCycleListeners
	lifeCycleListeners = append(lifeCycleListeners, moveFactoryScheduler)
	moveFactories, err := moveFactoryScheduler.MoveFactories()
	if err != nil {
		return nil, err
	}
	for _, moveFactory := range moveFactories {
		lifeCycleListeners = append(lifeCycleListeners, moveFactory)
	}
	for _, acceptor := range ssb.acceptors {
		lifeCycleListeners = append(lifeCycleListeners, acceptor)
	}
	for _, terminator := range ssb.terminators {
		lifeCycleListeners = append(lifeCycleListeners, terminator)
	}
	lifeCycleListeners = append(lifeCycleListeners, ssb.selector)
	lifeCycleListeners = append(lifeCycleListeners, ssb.evaluator)

	// sort the lifeCycleListener slice, put the Terminator at the head, since Acceptor may depend on StepContext updating by Terminator
	sort.Slice(lifeCycleListeners, func(i, j int) bool {
		if _, isTerminator := lifeCycleListeners[i].(Terminator); isTerminator {
			return true
		} else if _, isMoveFactoryScheduer := lifeCycleListeners[i].(MoveFactoryScheduler); isMoveFactoryScheduer {
			return false
		} else if _, isMoveFactoryScheduer := lifeCycleListeners[j].(MoveFactoryScheduler); isMoveFactoryScheduer {
			return true
		}

		return false
	})

	ss := &SerialSolver{
		name:                 ssb.name,
		lifeCycleListeners:   lifeCycleListeners,
		solution:             ssb.solution,
		params:               ssb.params,
		acceptors:            ssb.acceptors,
		selector:             ssb.selector,
		evaluator:            ssb.evaluator,
		terminators:          ssb.terminators,
		moveFactoryScheduler: moveFactoryScheduler,
		isLoggingToConsole:   ssb.isLoggingToConsole,
		asyncTerminate:       make(chan int, 1),
	}

	if ssb.isStableRand {
		if ssb.isSeedSet {
			InitReproducibleSolverRand(ssb.seed)
		} else {
			InitReproducibleSolverRand(DefaultSolverRandSeed)
		}
	} else {
		InitTrueSolverRand()
	}

	return ss, nil
}
