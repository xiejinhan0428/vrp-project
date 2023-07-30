package waveoptsolver

import (
	"errors"
	"fmt"
	"git.garena.com/shopee/bg-logistics/algo/wms/waveopt/localsearch/pkg/solver"
)

type OrderChangeMove struct {
	order   *SolverWaveOrder
	newTask *SolverWavePickingTask
	oldTask *SolverWavePickingTask
}

func (move *OrderChangeMove) String() string {
	return fmt.Sprintf("Change  %v: %v -> %v", move.order, move.oldTask, move.newTask)
}

func (move *OrderChangeMove) Do(solution solver.Solution) (solver.Move, error) {
	if move.oldTask != nil {
		err := move.oldTask.removeOrder(move.order)
		if err != nil {
			return new(solver.NoChangeMove), err
		}
		move.oldTask.SubPickingTasks = move.oldTask.Divider.Divide(move.oldTask)
	}

	if move.newTask != nil {
		move.newTask.addOrder(move.order)
		move.newTask.SubPickingTasks = move.newTask.Divider.Divide(move.newTask)
	}

	return &OrderChangeMove{
		order:   move.order,
		newTask: move.oldTask,
		oldTask: move.newTask,
	}, nil
}

func (move *OrderChangeMove) MovedVariables() ([]solver.Variable, error) {
	return append(make([]solver.Variable, 0), move.order), nil
}

func (move *OrderChangeMove) ToValues() ([]solver.Value, error) {
	return append(make([]solver.Value, 0), move.newTask), nil
}

// RandomOrderChangeMoveFactory moves a random order to a random task
type RandomOrderChangeMoveFactory struct {
	*solver.BaseMoveFactory
}

func (mf *RandomOrderChangeMoveFactory) CreateMove(solution solver.Solution) (solver.Move, error) {
	sol, ok := solution.(*WaveOptSolution)
	if !ok {
		return nil, errors.New("cannot perform RandomOrderChangeMove on a solution of unknown type")
	}

	orders := sol.Orders
	if len(orders) <= 0 {
		return &solver.NoChangeMove{}, nil
	}

	orderIdx := solver.GlobalSolverRand.Intn(len(orders))
	order := orders[orderIdx]

	tasks := sol.Tasks
	if len(tasks) <= 0 {
		return &solver.NoChangeMove{}, nil
	}

	newTaskIdx := -1
	for i := 0; newTaskIdx < 0 && i < 10; i++ {
		tmpIdx := solver.GlobalSolverRand.Intn(len(tasks))
		tmpTask := sol.Tasks[tmpIdx]
		if tmpTask.Id == order.PickingTask.Id {
			continue
		}
		newTaskIdx = tmpIdx
		break
	}
	if newTaskIdx < 0 {
		return new(solver.NoChangeMove), nil
	}

	newTask := tasks[newTaskIdx]

	return &OrderChangeMove{
		order:   order,
		newTask: newTask,
		oldTask: order.PickingTask,
	}, nil
}

func NewRandomOrderChangeMoveFactory() *RandomOrderChangeMoveFactory {
	return new(RandomOrderChangeMoveFactory)
}

type BestFitConstructionMoveFactory struct {
	*solver.BaseMoveFactory
	orderIdx int
	taskIdx  int
}

func (mf *BestFitConstructionMoveFactory) UpdateAtStepStart(stepContext *solver.StepContext) error {
	mf.taskIdx = 0
	return nil
}

func (mf *BestFitConstructionMoveFactory) UpdateAtStepEnd(stepContext *solver.StepContext) error {
	mf.orderIdx++
	return nil
}

func (mf *BestFitConstructionMoveFactory) UpdateAtMoveEnd(moveContext *solver.MoveContext) error {
	mf.taskIdx++
	return nil
}

func (mf *BestFitConstructionMoveFactory) CreateMove(solution solver.Solution) (solver.Move, error) {
	sol, _ := solution.(*WaveOptSolution)

	return &OrderChangeMove{
		order:   sol.Orders[mf.orderIdx],
		newTask: sol.Tasks[mf.taskIdx],
		oldTask: nil,
	}, nil
}

func NewBestFitConstructionMoveFactory() *BestFitConstructionMoveFactory {
	return &BestFitConstructionMoveFactory{
		orderIdx: 0,
		taskIdx:  0,
	}
}

type OrderSwapMove struct {
	leftOrder  *SolverWaveOrder
	rightOrder *SolverWaveOrder

	leftTask  *SolverWavePickingTask
	rightTask *SolverWavePickingTask
}

func (move *OrderSwapMove) String() string {
	return fmt.Sprintf("Swap  %v: %v  <-->  %v: %v", move.leftOrder, move.leftTask, move.rightOrder, move.rightTask)
}

func (move *OrderSwapMove) Do(solution solver.Solution) (solver.Move, error) {
	err := move.leftTask.removeOrder(move.leftOrder)
	if err != nil {
		return new(solver.NoChangeMove), err
	}

	err = move.rightTask.removeOrder(move.rightOrder)
	if err != nil {
		return new(solver.NoChangeMove), err
	}

	move.leftTask.addOrder(move.rightOrder)
	move.rightTask.addOrder(move.leftOrder)

	// re-divide the left and right tasks
	move.leftTask.SubPickingTasks = move.leftTask.Divider.Divide(move.leftTask)
	move.rightTask.SubPickingTasks = move.rightTask.Divider.Divide(move.rightTask)

	return &OrderSwapMove{
		leftOrder:  move.leftOrder,
		rightOrder: move.rightOrder,
		leftTask:   move.rightTask,
		rightTask:  move.leftTask,
	}, nil
}

func (move *OrderSwapMove) MovedVariables() ([]solver.Variable, error) {
	if move.leftOrder == nil {
		return nil, errors.New("left order of OrderSwapMove is nil")
	}
	if move.rightOrder == nil {
		return nil, errors.New("right order of OrderSwapMove is nil")
	}

	return append(make([]solver.Variable, 0), move.leftOrder, move.rightOrder), nil
}

func (move *OrderSwapMove) ToValues() ([]solver.Value, error) {
	if move.leftTask == nil {
		return nil, errors.New("left task of OrderSwapMove is nil")
	}
	if move.rightTask == nil {
		return nil, errors.New("right task of OrderSwapMove is nil")
	}

	return append(make([]solver.Value, 0), move.leftTask, move.rightTask), nil
}

type RandomOrderSwapMoveFactory struct {
	*solver.BaseMoveFactory
}

func (move *RandomOrderSwapMoveFactory) CreateMove(solution solver.Solution) (solver.Move, error) {
	sol, ok := solution.(*WaveOptSolution)
	if !ok {
		return new(solver.NoChangeMove), errors.New("cannot perform RandomOrderSwapMove on a solution of unknown type")
	}

	taskNum := len(sol.Tasks)
	if taskNum <= 1 {
		return new(solver.NoChangeMove), nil
	}

	leftTaskIdx := -1
	for i := 0; leftTaskIdx < 0 && i < 10; i++ {
		tmpIdx := solver.GlobalSolverRand.Intn(taskNum)
		if len(sol.Tasks[tmpIdx].Orders) <= 0 {
			continue
		}
		leftTaskIdx = tmpIdx
		break
	}

	rightTaskIdx := -1
	for i := 0; rightTaskIdx < 0 && i < 10; i++ {
		tmpIdx := solver.GlobalSolverRand.Intn(taskNum)
		if len(sol.Tasks[tmpIdx].Orders) <= 0 || tmpIdx == leftTaskIdx {
			continue
		}
		rightTaskIdx = tmpIdx
		break
	}

	if leftTaskIdx < 0 || rightTaskIdx < 0 {
		return new(solver.NoChangeMove), nil
	}

	leftTask := sol.Tasks[leftTaskIdx]
	rightTask := sol.Tasks[rightTaskIdx]

	if len(leftTask.Orders) <= 0 || len(rightTask.Orders) <= 0 {
		return new(solver.NoChangeMove), nil
	}

	leftOrder := leftTask.Orders[solver.GlobalSolverRand.Intn(len(leftTask.Orders))]
	rightOrder := rightTask.Orders[solver.GlobalSolverRand.Intn(len(rightTask.Orders))]

	return &OrderSwapMove{
		leftOrder:  leftOrder,
		rightOrder: rightOrder,
		leftTask:   leftTask,
		rightTask:  rightTask,
	}, nil
}

func NewRandomOrderSwapMoveFactory() *RandomOrderSwapMoveFactory {
	return new(RandomOrderSwapMoveFactory)
}

type MultipleOrderChangeMove struct {
	originalTasks []*SolverWavePickingTask

	orders []*SolverWaveOrder

	destinationTasks []*SolverWavePickingTask
}

func (move *MultipleOrderChangeMove) Do(solution solver.Solution) (solver.Move, error) {
	if len(move.orders) <= 0 {
		return &solver.NoChangeMove{}, nil
	}

	if len(move.orders) != len(move.destinationTasks) {
		return &solver.NoChangeMove{}, nil
	}

	for i, order := range move.orders {
		originalTask := move.originalTasks[i]
		destTask := move.destinationTasks[i]
		err := originalTask.removeOrder(order)
		if err != nil {
			return nil, err
		}
		originalTaskDivider := originalTask.Divider
		newOriginalSubTasks := originalTaskDivider.Divide(originalTask)
		originalTask.SubPickingTasks = newOriginalSubTasks

		destTask.addOrder(order)
		destTaskDivider := destTask.Divider
		newDestSubTasks := destTaskDivider.Divide(destTask)
		destTask.SubPickingTasks = newDestSubTasks
	}

	return &MultipleOrderChangeMove{
		originalTasks:    move.destinationTasks,
		orders:           move.orders,
		destinationTasks: move.originalTasks,
	}, nil
}

func (move *MultipleOrderChangeMove) MovedVariables() ([]solver.Variable, error) {
	vars := make([]solver.Variable, len(move.orders))
	for i, order := range move.orders {
		vars[i] = order
	}
	return vars, nil
}

func (move *MultipleOrderChangeMove) ToValues() ([]solver.Value, error) {
	vals := make([]solver.Value, len(move.destinationTasks))
	for i, task := range move.destinationTasks {
		vals[i] = task
	}

	return vals, nil
}

type TaskDestroyMoveFactory struct {
	*solver.BaseMoveFactory
}

func (mf *TaskDestroyMoveFactory) CreateMove(solution solver.Solution) (solver.Move, error) {
	sol, _ := solution.(*WaveOptSolution)

	nonEmptyNormalTasks := make([]*SolverWavePickingTask, 0)
	for _, task := range sol.NormalTasks {
		if len(task.Orders) > 0 {
			nonEmptyNormalTasks = append(nonEmptyNormalTasks, task)
		}
	}

	if len(nonEmptyNormalTasks) <= 0 {
		return &solver.NoChangeMove{}, nil
	}

	destroyTaskIdx := solver.GlobalSolverRand.Intn(len(nonEmptyNormalTasks))
	destroyTask := nonEmptyNormalTasks[destroyTaskIdx]

	candidateNormalTasks := make([]*SolverWavePickingTask, 0)
	candidateNormalTasks = append(candidateNormalTasks, nonEmptyNormalTasks[:destroyTaskIdx]...)
	candidateNormalTasks = append(candidateNormalTasks, nonEmptyNormalTasks[destroyTaskIdx+1:]...)

	if len(candidateNormalTasks) <= 0 {
		return &solver.NoChangeMove{}, nil
	}

	destinationTasks := make([]*SolverWavePickingTask, 0)
	for i := 0; i < len(destroyTask.Orders); i++ {
		destinationTaskIdx := solver.GlobalSolverRand.Intn(len(candidateNormalTasks))
		destinationTasks = append(destinationTasks, candidateNormalTasks[destinationTaskIdx])
	}

	originalTasks := make([]*SolverWavePickingTask, 0)
	for i := 0; i < len(destroyTask.Orders); i++ {
		originalTasks = append(originalTasks, destroyTask)
	}

	return &MultipleOrderChangeMove{
		originalTasks:    originalTasks,
		orders:           destroyTask.Orders,
		destinationTasks: destinationTasks,
	}, nil
}

func NewTaskDestroyMoveFactory() *TaskDestroyMoveFactory {
	return new(TaskDestroyMoveFactory)
}
