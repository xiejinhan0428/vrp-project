package waveoptsolver

import (
	"errors"

	"git.garena.com/shopee/bg-logistics/algo/wms/waveopt/localsearch/pkg/score"
	"git.garena.com/shopee/bg-logistics/algo/wms/waveopt/localsearch/pkg/solver"
)

type WaveOptEvaluator struct {
	*score.BaseEvaluator
	constraints        []FloatConstraint
	distanceCalculator DistanceCalculator
	constraintLevel    int
	objectiveLevel     int
}

func (e *WaveOptEvaluator) AddConstraints(constraints ...FloatConstraint) {
	e.constraints = append(e.constraints, constraints...)
}

func (e *WaveOptEvaluator) AddDistanceCalculator(distanceCalculator DistanceCalculator) {
	e.distanceCalculator = distanceCalculator
}

func (e *WaveOptEvaluator) Evaluate(solution solver.Solution) (solver.Score, error) {
	var err error
	sol, ok := solution.(*WaveOptSolution)
	if !ok {
		return nil, errors.New("not evaluating a WaveOptSolution")
	}

	score_, err := e.NewScore()
	if err != nil {
		return nil, err
	}
	scr, ok := score_.(*score.FloatScore)
	if !ok {
		return nil, errors.New("not creating a FloatScore")
	}

	// check constraints
	for _, cons := range e.constraints {
		consLevel := cons.Level()
		consScore, err := cons.Penalize(sol)
		if err != nil {
			return nil, err
		}
		scr.ConstraintScores[consLevel] = consScore
	}

	// calculate distance
	pickingDistance := 0.0
	if e.distanceCalculator != nil {
		for _, task := range sol.Tasks {
			for _, subTask := range task.SubPickingTasks {
				subTaskDistance := e.distanceCalculator.CalculateDistance(subTask)
				if task.IsReserved {
					pickingDistance += reservedOrderDistanceCoeff * subTaskDistance * float64(len(subTask.Skus))
				} else {
					pickingDistance += subTaskDistance
				}
			}
		}
	}
	scr.ObjectiveScores[0] = pickingDistance

	// count orders in normal picking task
	orderNum := 0
	for _, task := range sol.Tasks {
		if task.IsReserved {
			continue
		}
		taskOrderNum := len(task.Orders)
		orderNum += taskOrderNum
	}
	scr.ObjectiveScores[1] = float64(-orderNum)

	return scr, nil
}

func (e *WaveOptEvaluator) NewScore() (solver.Score, error) {
	var err error
	score, err := score.NewFloatScore(e.constraintLevel, e.objectiveLevel, solver.DownScore)
	if err != nil {
		return nil, errors.New("not a FloatScore")
	}

	return &score, nil
}

type WaveOptScenario string

const (
	PureSinglePicker                           WaveOptScenario = "SinglePicker"
	MultiPickerWithMergingWhileSortingStation  WaveOptScenario = "MultiPickerWithMWS"
	SinglePickerWithMergingWhileSortingStation WaveOptScenario = "SinglePickerWithMWS"
	MultiPickerWithMergingLane                 WaveOptScenario = "MultiPickerWithMergingLane"
	SinglePickerWithMergingLane                WaveOptScenario = "SinglePickerWithMergingLane"
)

type EvaluatorFactory struct{}

func (f *EvaluatorFactory) createGeneralEvaluator(
	isCrossZoneCluster bool,
	minOrderQtyPerPickingTask int64,
	maxOrderQtyPerPickingTask int64,
	minItemQtyPerSubPickingTask int64,
	maxItemQtyPerSubPickingTask int64,
	minItemVolumePerSubPickingTask int64,
	maxItemVolumePerSubPickingTask int64,
	zoneWeight float64,
	pathwayWeight float64,
	segmentWeight float64,
) (solver.Evaluator, error) {
	var err error

	isCrossZoneClusterCons, err := NewPickingTaskCrossZoneClusterConstraint(isCrossZoneCluster, 0)
	if err != nil {
		return nil, err
	}
	orderQtyPerTaskCons, err := NewPickingTaskOrderQtyConstraint(minOrderQtyPerPickingTask, maxOrderQtyPerPickingTask, 1)
	if err != nil {
		return nil, err
	}
	itemQtyPerSubPickingTaskCons, err := NewSubPickingTaskItemQtyConstraint(minItemQtyPerSubPickingTask, maxItemQtyPerSubPickingTask, 2)
	if err != nil {
		return nil, err
	}
	itemVolumePerSubPickingTaskCons, err := NewSubPickingTaskItemVolumeConstraint(minItemVolumePerSubPickingTask, maxItemVolumePerSubPickingTask, 3)
	if err != nil {
		return nil, err
	}

	evaluator := new(WaveOptEvaluator)
	evaluator.AddConstraints(isCrossZoneClusterCons)
	evaluator.AddConstraints(orderQtyPerTaskCons)
	evaluator.AddConstraints(itemQtyPerSubPickingTaskCons)
	evaluator.AddConstraints(itemVolumePerSubPickingTaskCons)

	distanceCalculator, err := NewLocationBasedDistanceCalculator(zoneWeight, pathwayWeight, segmentWeight)
	if err != nil {
		return nil, err
	}
	evaluator.AddDistanceCalculator(distanceCalculator)

	evaluator.constraintLevel = 4
	evaluator.objectiveLevel = 2

	return evaluator, nil
}

func (f *EvaluatorFactory) CreateSinglePickerEvaluator(
	isCrossZoneCluster bool,
	minOrderQtyPerPickingTask int64,
	maxOrderQtyPerPickingTask int64,
	minItemQtyPerSubPickingTask int64,
	maxItemQtyPerSubPickingTask int64,
	minItemVolumePerSubPickingTask int64,
	maxItemVolumePerSubPickingTask int64,
	zoneWeight float64,
	pathwayWeight float64,
	segmentWeight float64,
) (solver.Evaluator, error) {
	var err error
	evaluator, err := f.createGeneralEvaluator(
		isCrossZoneCluster,
		minOrderQtyPerPickingTask,
		maxOrderQtyPerPickingTask,
		minItemQtyPerSubPickingTask,
		maxItemQtyPerSubPickingTask,
		minItemVolumePerSubPickingTask,
		maxItemVolumePerSubPickingTask,
		zoneWeight,
		pathwayWeight,
		segmentWeight,
	)
	if err != nil {
		return nil, err
	}

	return evaluator, nil
}

func (f *EvaluatorFactory) CreateMultiPickerWithMergingWhileSortingStationEvaluator(
	isCrossZoneCluster bool,
	minOrderQtyPerPickingTask int64,
	maxOrderQtyPerPickingTask int64,
	minItemQtyPerSubPickingTask int64,
	maxItemQtyPerSubPickingTask int64,
	minItemVolumePerSubPickingTask int64,
	maxItemVolumePerSubPickingTask int64,
	zoneWeight float64,
	pathwayWeight float64,
	segmentWeight float64,
) (solver.Evaluator, error) {
	var err error
	evaluator, err := f.createGeneralEvaluator(
		isCrossZoneCluster,
		minOrderQtyPerPickingTask,
		maxOrderQtyPerPickingTask,
		minItemQtyPerSubPickingTask,
		maxItemQtyPerSubPickingTask,
		minItemVolumePerSubPickingTask,
		maxItemVolumePerSubPickingTask,
		zoneWeight,
		pathwayWeight,
		segmentWeight,
	)
	if err != nil {
		return nil, err
	}

	return evaluator, nil
}

func (f *EvaluatorFactory) CreateSinglePickerWithMergingWhileSortingStationEvaluator(
	isCrossZoneCluster bool,
	minOrderQtyPerPickingTask int64,
	maxOrderQtyPerPickingTask int64,
	minItemQtyPerSubPickingTask int64,
	maxItemQtyPerSubPickingTask int64,
	minItemVolumePerSubPickingTask int64,
	maxItemVolumePerSubPickingTask int64,
	zoneWeight float64,
	pathwayWeight float64,
	segmentWeight float64,
) (solver.Evaluator, error) {
	var err error
	evaluator, err := f.createGeneralEvaluator(
		isCrossZoneCluster,
		minOrderQtyPerPickingTask,
		maxOrderQtyPerPickingTask,
		minItemQtyPerSubPickingTask,
		maxItemQtyPerSubPickingTask,
		minItemVolumePerSubPickingTask,
		maxItemVolumePerSubPickingTask,
		zoneWeight,
		pathwayWeight,
		segmentWeight,
	)
	if err != nil {
		return nil, err
	}

	return evaluator, nil
}

func (f *EvaluatorFactory) CreateMultiPickerWithMergingLaneEvaluator(
	isCrossZoneCluster bool,
	minOrderQtyPerPickingTask int64,
	maxOrderQtyPerPickingTask int64,
	minItemQtyPerSubPickingTask int64,
	maxItemQtyPerSubPickingTask int64,
	minItemVolumePerSubPickingTask int64,
	maxItemVolumePerSubPickingTask int64,
	zoneWeight float64,
	pathwayWeight float64,
	segmentWeight float64,
) (solver.Evaluator, error) {
	var err error
	evaluator, err := f.createGeneralEvaluator(
		isCrossZoneCluster,
		minOrderQtyPerPickingTask,
		maxOrderQtyPerPickingTask,
		minItemQtyPerSubPickingTask,
		maxItemQtyPerSubPickingTask,
		minItemVolumePerSubPickingTask,
		maxItemVolumePerSubPickingTask,
		zoneWeight,
		pathwayWeight,
		segmentWeight,
	)
	if err != nil {
		return nil, err
	}

	return evaluator, nil
}

func (f *EvaluatorFactory) CreateSinglePickerWithMergingLaneEvaluator(
	isCrossZoneCluster bool,
	minOrderQtyPerPickingTask int64,
	maxOrderQtyPerPickingTask int64,
	minItemQtyPerSubPickingTask int64,
	maxItemQtyPerSubPickingTask int64,
	minItemVolumePerSubPickingTask int64,
	maxItemVolumePerSubPickingTask int64,
	zoneWeight float64,
	pathwayWeight float64,
	segmentWeight float64,
) (solver.Evaluator, error) {
	var err error
	evaluator, err := f.createGeneralEvaluator(
		isCrossZoneCluster,
		minOrderQtyPerPickingTask,
		maxOrderQtyPerPickingTask,
		minItemQtyPerSubPickingTask,
		maxItemQtyPerSubPickingTask,
		minItemVolumePerSubPickingTask,
		maxItemVolumePerSubPickingTask,
		zoneWeight,
		pathwayWeight,
		segmentWeight,
	)
	if err != nil {
		return nil, err
	}

	return evaluator, nil
}

func (f *EvaluatorFactory) CreateInitEvaluator(
	isCrossZoneCluster bool,
	minOrderQtyPerPickingTask int64,
	maxOrderQtyPerPickingTask int64,
	minItemQtyPerSubPickingTask int64,
	maxItemQtyPerSubPickingTask int64,
	minItemVolumePerSubPickingTask int64,
	maxItemVolumePerSubPickingTask int64,
	zoneWeight float64,
	pathwayWeight float64,
	segmentWeight float64,
) (solver.Evaluator, error) {
	var err error

	notInitOrderCons, err := NewNotInitOrderConstraint(0)
	if err != nil {
		return nil, err
	}
	isCrossZoneClusterCons, err := NewPickingTaskCrossZoneClusterConstraint(isCrossZoneCluster, 1)
	if err != nil {
		return nil, err
	}
	orderQtyPerTaskCons, err := NewPickingTaskOrderQtyConstraint(minOrderQtyPerPickingTask, maxOrderQtyPerPickingTask, 2)
	if err != nil {
		return nil, err
	}
	itemQtyPerSubPickingTaskCons, err := NewSubPickingTaskItemQtyConstraint(minItemQtyPerSubPickingTask, maxItemQtyPerSubPickingTask, 3)
	if err != nil {
		return nil, err
	}
	itemVolumePerSubPickingTaskCons, err := NewSubPickingTaskItemVolumeConstraint(minItemVolumePerSubPickingTask, maxItemVolumePerSubPickingTask, 4)
	if err != nil {
		return nil, err
	}

	evaluator := new(WaveOptEvaluator)
	evaluator.AddConstraints(notInitOrderCons)
	evaluator.AddConstraints(isCrossZoneClusterCons)
	evaluator.AddConstraints(orderQtyPerTaskCons)
	evaluator.AddConstraints(itemQtyPerSubPickingTaskCons)
	evaluator.AddConstraints(itemVolumePerSubPickingTaskCons)

	//distanceCalculator, err := NewLocationBasedDistanceCalculator(zoneWeight, pathwayWeight, segmentWeight)
	//if err != nil {
	//	return nil, err
	//}
	//evaluator.AddDistanceCalculator(distanceCalculator)

	evaluator.constraintLevel = 5
	evaluator.objectiveLevel = 2

	return evaluator, nil
}
