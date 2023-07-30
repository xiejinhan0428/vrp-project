package waveoptsolver

import (
	"fmt"
	"git.garena.com/shopee/bg-logistics/algo/wms/waveopt/localsearch/pkg/solver"
)

type WaveOptSolutionSingle struct {
	PickingTasks []*PickingTaskSingle

	normalTasks  []*PickingTaskSingle
	orphanOrders []*SolverWaveOrder
}

func (sol *WaveOptSolutionSingle) Copy() (solver.Solution, error) {
	newSolution := new(WaveOptSolutionSingle)

	for _, pickingTask := range sol.PickingTasks {
		tmpPickingTask := &PickingTaskSingle{
			Orders: append(make([]*SolverWaveOrder, 0), pickingTask.Orders...),
		}
		newSolution.PickingTasks = append(newSolution.PickingTasks, tmpPickingTask)
	}

	for _, nt := range sol.normalTasks {
		tmpNormalTask := &PickingTaskSingle{
			Orders: append(make([]*SolverWaveOrder, 0), nt.Orders...),
		}
		newSolution.normalTasks = append(newSolution.normalTasks, tmpNormalTask)
	}

	for _, orphanOrder := range sol.orphanOrders {
		newSolution.orphanOrders = append(newSolution.orphanOrders, orphanOrder)
	}

	return newSolution, nil
}

type PickingTaskSingle struct {
	Orders              []*SolverWaveOrder
	id                  int
	OrderQty            int64
	SkuQty              int64
	isUrgent            bool
	Distance            int64
	ConstraintViolation int64
	IfCrossZoneCluster  bool
	Volume              int64
	SplitLevelNum       int64
}

func (sol *WaveOptSolutionSingle) String() string {
	return fmt.Sprintf("PickingTasks:{%v}\nnormalTasks:{%v}\norphanOrders:{%v}", sol.PickingTasks, sol.normalTasks, sol.orphanOrders)
}

func (pt *PickingTaskSingle) String() string {
	return fmt.Sprintf("Orders:{%v}", pt.Orders)
}
