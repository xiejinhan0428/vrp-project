package strategy

import (
	"fmt"
	"math"

	"git.garena.com/shopee/bg-logistics/algo/wms/waveopt/localsearch/pkg/solver"

	merror "git.garena.com/shopee/bg-logistics/algo/wms/waveopt/localsearch/pkg/error"
)

type SimulatedAnnealingAcceptor struct {
	*BaseAcceptor
	initialTemperature []float64
	stepCoolingRate    []float64 // zero means constant cooling rate, i.e., Metropolis condition; > 0 && < 1 means fixed cooling rate for each step; otherwize use currentSolverCoolingRate

	currentSolverCoolingRate float64
}

func (sa *SimulatedAnnealingAcceptor) UpdateAtSolverStart(solverContext *solver.SolverContext) error {
	sa.currentSolverCoolingRate = 1.0
	return nil
}

func (sa *SimulatedAnnealingAcceptor) UpdateAtStepStart(stepContext *solver.StepContext) error {
	sa.currentSolverCoolingRate = stepContext.SolvingProgress
	return nil
}

func (sa *SimulatedAnnealingAcceptor) IsAccepted(moveContext *solver.MoveContext) (bool, error) {
	bestScore := moveContext.StepContext.SolverContext.BestScore
	currentScore := moveContext.AfterMovingScore

	diff, err := currentScore.Sub(bestScore)
	if err != nil {
		return false, merror.New("SimulatedAnnealingAcceptor: Fail to get the difference between current score", fmt.Sprint(currentScore), "and best score", fmt.Sprint(bestScore)).CausedBy(err)
	}
	if len(sa.initialTemperature) != len(diff) {
		return false, merror.New("SimulatedAnnealingAcceptor: Score level in SimulatedAnnealingAcceptor do not match with that in the Solver.")
	}

	acceptChance := 1.0
	for i, levelDiff := range diff {
		var levelCoolingRate float64
		if sa.stepCoolingRate[i] <= 0.0 {
			levelCoolingRate = 1.0
		} else if sa.stepCoolingRate[i] > 0.0 && sa.stepCoolingRate[i] < 1.0 {
			levelCoolingRate = math.Pow(sa.stepCoolingRate[i], float64(moveContext.StepContext.StepCount))
		} else {
			levelCoolingRate = sa.currentSolverCoolingRate
		}

		inverseLevelCoolingRate := 1.0 - levelCoolingRate

		trend, err := currentScore.Trend()
		if err != nil {
			return false, err
		}
		if levelDiff*float64(trend) < 0 {
			if levelCoolingRate == 0.0 {
				acceptChance = 0.0
			} else {
				acceptChance *= math.Exp(levelDiff * float64(trend) / (sa.initialTemperature[i] * inverseLevelCoolingRate))
			}
		}
	}

	return solver.GlobalSolverRand.Float64() < acceptChance, nil
}

type SimulatedAnnealingAcceptorBuilder struct {
	initialTemperature []float64
	stepCoolingRate    []float64
}

func NewSimulatedAnnealingAcceptorBuilder() *SimulatedAnnealingAcceptorBuilder {
	return new(SimulatedAnnealingAcceptorBuilder)
}

func (b *SimulatedAnnealingAcceptorBuilder) WithInitialTemperatures(initialTemperature []float64) *SimulatedAnnealingAcceptorBuilder {
	b.initialTemperature = initialTemperature

	return b
}

func (b *SimulatedAnnealingAcceptorBuilder) WithStepCoolingRate(stepCoolingRate []float64) *SimulatedAnnealingAcceptorBuilder {
	b.stepCoolingRate = stepCoolingRate

	return b
}

func (b *SimulatedAnnealingAcceptorBuilder) Build() (*SimulatedAnnealingAcceptor, error) {
	if len(b.initialTemperature) == 0 {
		return nil, merror.New("SimulatedAnnealingAcceptor: It is meaningless to init a SimulatedAnnealingAcceptor with a zero length initialTemperature slice.")
	}

	if len(b.stepCoolingRate) == 0 {
		b.stepCoolingRate = make([]float64, len(b.initialTemperature))
		for i := range b.initialTemperature {
			b.stepCoolingRate[i] = -1.0
		}
	}

	if len(b.initialTemperature) != len(b.stepCoolingRate) {
		return nil, merror.New("SimulatedAnnealingAcceptor: InitialTemperature and StepCoolingRate must have the same level.")
	}

	return &SimulatedAnnealingAcceptor{
		initialTemperature: b.initialTemperature,
		stepCoolingRate:    b.stepCoolingRate,
	}, nil
}
