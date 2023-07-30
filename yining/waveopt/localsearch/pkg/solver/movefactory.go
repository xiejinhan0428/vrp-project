package solver

import (
	merror "git.garena.com/shopee/bg-logistics/algo/wms/waveopt/localsearch/pkg/error"
)

type BaseMoveFactory struct{}

func (t *BaseMoveFactory) UpdateAtSolverStart(context *SolverContext) error { return nil }
func (t *BaseMoveFactory) UpdateAtSolverEnd(context *SolverContext) error   { return nil }
func (t *BaseMoveFactory) UpdateAtStepStart(stepContext *StepContext) error { return nil }
func (t *BaseMoveFactory) UpdateAtStepEnd(stepContext *StepContext) error   { return nil }
func (t *BaseMoveFactory) UpdateAtMoveStart(moveContext *MoveContext) error { return nil }
func (t *BaseMoveFactory) UpdateAtMoveEnd(moveContext *MoveContext) error   { return nil }

// BaseMoveFactoryScheduler the BaseMoveFactoryScheduler implements the SolverLifeCycleListener interface,
// embed it into concrete MoverFactorySchedulers
type BaseMoveFactoryScheduler struct {
	moveFactories []MoveFactory
}

func (s *BaseMoveFactoryScheduler) MoveFactories() ([]MoveFactory, error) {
	return s.moveFactories, nil
}

func (s *BaseMoveFactoryScheduler) UpdateAtSolverStart(context *SolverContext) error { return nil }
func (s *BaseMoveFactoryScheduler) UpdateAtSolverEnd(context *SolverContext) error   { return nil }
func (s *BaseMoveFactoryScheduler) UpdateAtStepStart(stepContext *StepContext) error { return nil }
func (s *BaseMoveFactoryScheduler) UpdateAtStepEnd(stepContext *StepContext) error   { return nil }
func (s *BaseMoveFactoryScheduler) UpdateAtMoveStart(moveContext *MoveContext) error { return nil }
func (s *BaseMoveFactoryScheduler) UpdateAtMoveEnd(moveContext *MoveContext) error   { return nil }

type SequentialMoveFactoryScheduler struct {
	*BaseMoveFactoryScheduler
	idx int
}

func NewSequentialMoveFactoryScheduler(moveFactories []MoveFactory) (*SequentialMoveFactoryScheduler, error) {
	if len(moveFactories) == 0 {
		return nil, merror.New("SequentialMoveFactoryScheduler: Empty input MoveFactory list.")
	}
	return &SequentialMoveFactoryScheduler{
		BaseMoveFactoryScheduler: &BaseMoveFactoryScheduler{moveFactories: moveFactories},
		idx:                      0,
	}, nil
}

func (s *SequentialMoveFactoryScheduler) UpdateAtStepStart(stepContext *StepContext) error {
	s.idx = stepContext.StepCount % len(s.moveFactories)
	return nil
}

// SelectMoveFactory implements the MoveFactoryScheduler interface
func (s *SequentialMoveFactoryScheduler) SelectMoveFactory() (MoveFactory, error) {
	return s.moveFactories[s.idx], nil
}

type RouletteMoveFactoryScheduler struct {
	*BaseMoveFactoryScheduler
	accuMoveFactoryWeights []float64
}

func NewRouletteMoveFactoryScheduler(moveFactoryToWeightMap map[MoveFactory]float64) (*RouletteMoveFactoryScheduler, error) {
	if len(moveFactoryToWeightMap) == 0 {
		return nil, merror.New("RouletteMoveFactoryScheduler: Empty input map of move factories and corresponding weights.")
	}

	// accumulate the weight
	moveFactoryToAccWeightMap := make(map[MoveFactory]float64)
	sum := 0.0
	for moveFactory, weight := range moveFactoryToWeightMap {
		sum += weight
		moveFactoryToAccWeightMap[moveFactory] = sum
	}

	rouletteScheduler := &RouletteMoveFactoryScheduler{
		BaseMoveFactoryScheduler: &BaseMoveFactoryScheduler{moveFactories: make([]MoveFactory, 0)},
		accuMoveFactoryWeights:   make([]float64, 0),
	}

	// normalize the weight
	for moveFactory, accWeight := range moveFactoryToAccWeightMap {
		rouletteScheduler.moveFactories = append(rouletteScheduler.moveFactories, moveFactory)
		rouletteScheduler.accuMoveFactoryWeights = append(rouletteScheduler.accuMoveFactoryWeights, accWeight/sum)
	}

	return rouletteScheduler, nil
}

func (s *RouletteMoveFactoryScheduler) SelectMoveFactory() (MoveFactory, error) {
	if len(s.moveFactories) == 0 {
		return nil, merror.New("RouletteMoveFactoryScheduler: No MoveFactory.")
	}

	rnd := GlobalSolverRand.Float64()
	prev := 0.0
	for i, weight := range s.accuMoveFactoryWeights {
		if rnd >= prev && rnd < weight {
			return s.moveFactories[i], nil
		}
		prev = weight
	}

	rndInt := GlobalSolverRand.Intn(len(s.moveFactories))
	return s.moveFactories[rndInt], nil
}
