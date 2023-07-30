package strategy

import (
	"git.garena.com/shopee/bg-logistics/algo/wms/waveopt/localsearch/pkg/solver"
)

type TabuSearchAcceptor struct {
	*BaseAcceptor

	variableTabuSize int // non-positive if variables are not tabu
	valueTabuSize    int // non-positive if values are not tabu

	variableTabuTable map[interface{}]int
	valueTabuTable    map[interface{}]int
}

func (tsa *TabuSearchAcceptor) IsAccepted(moveContext *solver.MoveContext) (bool, error) {
	// aspiration criteria: the move with a score higher than the incumbent best score is accepted despite it breaks tabu rules
	moveScore := moveContext.AfterMovingScore
	bestScore := moveContext.StepContext.SolverContext.BestScore
	cmp, err := moveScore.CompareToScore(bestScore)
	if err != nil {
		return false, err
	}
	if cmp >= 0 {
		return true, nil
	}

	move := moveContext.Move
	variables, err := move.MovedVariables()
	if err != nil {
		return false, err
	}
	values, err := move.ToValues()
	if err != nil {
		return false, err
	}
	moveStepCount := moveContext.StepContext.StepCount

	if tsa.variableTabuSize > 0 {
		for _, vari := range variables {
			stepCount, ok := tsa.variableTabuTable[vari.Identifier()]
			if !ok {
				continue
			}

			if stepCount+tsa.variableTabuSize > moveStepCount {
				return false, nil
			}
		}
	}

	if tsa.valueTabuSize > 0 {
		for _, val := range values {
			stepCount, ok := tsa.valueTabuTable[val.Identifier()]
			if !ok {
				continue
			}

			if stepCount+tsa.valueTabuSize > moveStepCount {
				return false, nil
			}
		}
	}

	return true, nil
}

func (tsa *TabuSearchAcceptor) UpdateAtStepEnd(stepContext *solver.StepContext) error {
	winningMove := stepContext.WinningMove

	movedVariables, err := winningMove.MovedVariables()
	if err != nil {
		return err
	}
	toValues, err := winningMove.ToValues()
	if err != nil {
		return err
	}

	stepCount := stepContext.StepCount

	for _, vari := range movedVariables {
		tsa.variableTabuTable[vari.Identifier()] = stepCount
	}

	for _, val := range toValues {
		tsa.valueTabuTable[val.Identifier()] = stepCount
	}

	return nil
}

type TabuSearchAcceptorBuilder struct {
	variableTabuSize int
	valueTabuSize    int
}

func NewTabuSearchAcceptorBuilder() *TabuSearchAcceptorBuilder {
	return &TabuSearchAcceptorBuilder{
		variableTabuSize: -1,
		valueTabuSize:    -1,
	}
}

func (b *TabuSearchAcceptorBuilder) WithVariableTabuSize(variableTabuSize int) *TabuSearchAcceptorBuilder {
	b.variableTabuSize = variableTabuSize
	return b
}

func (b *TabuSearchAcceptorBuilder) WithValueTabuSize(valueTabuSize int) *TabuSearchAcceptorBuilder {
	b.valueTabuSize = valueTabuSize
	return b
}

func (b *TabuSearchAcceptorBuilder) Build() (*TabuSearchAcceptor, error) {
	return &TabuSearchAcceptor{
		variableTabuSize:  b.variableTabuSize,
		valueTabuSize:     b.valueTabuSize,
		variableTabuTable: make(map[interface{}]int),
		valueTabuTable:    make(map[interface{}]int),
	}, nil
}
