package waveoptsolver

import (
	"fmt"
)

type FloatConstraint interface {
	Name() string
	SetName(string)
	SetLevel(int)
	Level() int
	Penalize(*WaveOptSolution) (float64, error)
}

type BaseConstraint struct {
	name  string
	level int
}

func (c *BaseConstraint) Name() string {
	return c.name
}

func (c *BaseConstraint) SetName(name string) {
	c.name = name
}

func (c *BaseConstraint) Level() int {
	return c.level
}

func (c *BaseConstraint) SetLevel(level int) {
	c.level = level
}

type NotInitOrderConstraint struct {
	*BaseConstraint
}

func (c *NotInitOrderConstraint) Penalize(solution *WaveOptSolution) (float64, error) {
	penalty := 0.0
	for _, order := range solution.Orders {
		if order.PickingTask == nil {
			penalty += 1
		}
	}

	return penalty, nil
}

func NewNotInitOrderConstraint(level int) (*NotInitOrderConstraint, error) {
	if level < 0 {
		return nil, fmt.Errorf("level of constraint %v is out of bound", level)
	}

	baseCons := new(BaseConstraint)
	baseCons.SetName("NotInitOrderConstraint")
	baseCons.SetLevel(level)

	return &NotInitOrderConstraint{baseCons}, nil
}

type PickingTaskCrossZoneClusterConstraint struct {
	*BaseConstraint
	IsCrossZoneCluster bool
}

func (c *PickingTaskCrossZoneClusterConstraint) Penalize(solution *WaveOptSolution) (float64, error) {
	penalty := 0.0
	for _, task := range solution.Tasks {
		if task.IsReserved || len(task.Orders) <= 0 {
			continue
		}
		taskPenalty, err := penalizeCrossCluster(task, c.IsCrossZoneCluster)
		if err != nil {
			return 0.0, nil
		}
		penalty += taskPenalty
	}
	return penalty, nil
}

func penalizeCrossCluster(task *SolverWavePickingTask, isCrossCluster bool) (float64, error) {
	clusterSet := make(map[string]bool)
	for _, order := range task.Orders {
		for _, sku := range order.Skus {
			skuZoneClusterId := sku.Location.ZoneClusterId
			if len(skuZoneClusterId) == 0 {
				return 0.0, fmt.Errorf("sku %v has blank ZoneClusterId", sku.Id)
			}

			clusterSet[skuZoneClusterId] = true
		}
	}

	if !isCrossCluster && len(clusterSet) > 1 {
		return float64(len(clusterSet)) - 1, nil
	}

	return 0.0, nil
}

func NewPickingTaskCrossZoneClusterConstraint(isCrossZoneCluster bool, level int) (*PickingTaskCrossZoneClusterConstraint, error) {
	if level < 0 {
		return nil, fmt.Errorf("level of constraint %v is out of bound", level)
	}

	baseCons := new(BaseConstraint)
	baseCons.SetName("PickingTaskCrossZoneClusterConstraint")
	baseCons.SetLevel(level)
	return &PickingTaskCrossZoneClusterConstraint{
		BaseConstraint:     baseCons,
		IsCrossZoneCluster: isCrossZoneCluster,
	}, nil
}

type PickingTaskOrderQtyConstraint struct {
	*BaseConstraint
	maxOrderQty int64
	minOrderQty int64
}

func (c *PickingTaskOrderQtyConstraint) Penalize(solution *WaveOptSolution) (float64, error) {
	penalty := 0.0

	for _, task := range solution.Tasks {
		if len(task.Orders) <= 0 || task.IsReserved {
			continue
		}

		taskPenalty, _ := penalizeOrderQtyPerTask(task, c.maxOrderQty, c.minOrderQty)
		penalty += taskPenalty
	}

	return penalty, nil
}

func penalizeOrderQtyPerTask(task *SolverWavePickingTask, max, min int64) (float64, error) {
	penalty := 0.0
	orderQty := int64(len(task.Orders))
	if orderQty < min && !task.IsReserved {
		penalty += float64(min - orderQty)
	} else if orderQty > max {
		penalty += float64(orderQty - max)
	}

	return penalty, nil
}

func NewPickingTaskOrderQtyConstraint(minOrderQty int64, maxOrderQty int64, level int) (*PickingTaskOrderQtyConstraint, error) {
	if level < 0 {
		return nil, fmt.Errorf("level of constraint %v is out of bound", level)
	}

	baseCons := new(BaseConstraint)
	baseCons.SetName("PickingTaskQtyConstraint")
	baseCons.SetLevel(level)
	return &PickingTaskOrderQtyConstraint{
		BaseConstraint: baseCons,
		minOrderQty:    minOrderQty,
		maxOrderQty:    maxOrderQty,
	}, nil
}

type SubPickingTaskItemQtyConstraint struct {
	*BaseConstraint
	maxItemQty int64
	minItemQty int64
}

func (c *SubPickingTaskItemQtyConstraint) Penalize(solution *WaveOptSolution) (float64, error) {
	penalty := 0.0

	for _, task := range solution.Tasks {
		if task.IsReserved || len(task.Orders) <= 0 {
			continue
		}
		taskPenalty, _ := penalizeItemQtyPerSubPickingTask(task, c.maxItemQty, c.minItemQty)
		penalty += taskPenalty
	}

	return penalty, nil
}

func penalizeItemQtyPerSubPickingTask(task *SolverWavePickingTask, max, min int64) (float64, error) {
	penalty := 0.0
	for _, subTask := range task.SubPickingTasks {
		itemQty := int64(0)
		for _, sku := range subTask.Skus {
			itemQty += sku.Qty
		}

		if itemQty < min {
			penalty += float64(min - itemQty)
		} else if itemQty > max {
			penalty += float64(itemQty - max)
		}
	}

	return penalty, nil
}

func NewSubPickingTaskItemQtyConstraint(minItemQty int64, maxItemQty int64, level int) (*SubPickingTaskItemQtyConstraint, error) {
	if level < 0 {
		return nil, fmt.Errorf("level of constraint %v is out of bound", level)
	}
	if minItemQty < 0 {
		return nil, fmt.Errorf("minimal quantity of items in a sub picking task cannot be negative: %v", minItemQty)
	}
	if maxItemQty < 0 {
		return nil, fmt.Errorf("maximal quantity of items in a sub picking task cannot be negative: %v", maxItemQty)
	}
	if maxItemQty < minItemQty {
		return nil, fmt.Errorf("unreasonable bounds of quantity of items in a sub picking task: max is %v, min is %v", maxItemQty, minItemQty)
	}

	baseCons := new(BaseConstraint)
	baseCons.SetName("SubPickingTaskItemQtyConstraint")
	baseCons.SetLevel(level)
	return &SubPickingTaskItemQtyConstraint{
		BaseConstraint: baseCons,
		minItemQty:     minItemQty,
		maxItemQty:     maxItemQty,
	}, nil
}

type SubPickingTaskItemVolumeConstraint struct {
	*BaseConstraint
	minItemVolume int64
	maxItemVolume int64
}

func (c *SubPickingTaskItemVolumeConstraint) Penalize(solution *WaveOptSolution) (float64, error) {
	penalty := 0.0

	for _, task := range solution.Tasks {
		if task.IsReserved || len(task.Orders) <= 0 {
			continue
		}
		taskPenalty, _ := penalizeItemVolumePerSubPickingTask(task, c.maxItemVolume, c.minItemVolume)
		penalty += taskPenalty
	}

	return penalty, nil
}

func penalizeItemVolumePerSubPickingTask(task *SolverWavePickingTask, max, min int64) (float64, error) {
	penalty := 0.0
	for _, subTask := range task.SubPickingTasks {
		totalSkuVolume := int64(0)
		isVolumeConsDisabled := false
		for _, sku := range subTask.Skus {
			skuVolume := sku.SingleVolume()
			if skuVolume <= 0 {
				isVolumeConsDisabled = true
				break
			}
			totalSkuVolume += skuVolume * sku.Qty
		}

		if isVolumeConsDisabled {
			continue
		}

		if totalSkuVolume < min {
			penalty += float64(min - totalSkuVolume)
		} else if totalSkuVolume > max {
			penalty += float64(totalSkuVolume - max)
		}
	}

	return penalty, nil
}

func NewSubPickingTaskItemVolumeConstraint(minItemVolume int64, maxItemVolume int64, level int) (*SubPickingTaskItemVolumeConstraint, error) {
	if level < 0 {
		return nil, fmt.Errorf("level of constraint %v is out of bound", level)
	}
	if minItemVolume < 0 {
		return nil, fmt.Errorf("minimal volume of items in a sub picking task cannot be negative: %v", minItemVolume)
	}
	if maxItemVolume < 0 {
		return nil, fmt.Errorf("maximal volume of items in a sub picking task cannot be negative: %v", maxItemVolume)
	}
	if maxItemVolume < minItemVolume {
		return nil, fmt.Errorf("unreasonable bounds of volume of items in a sub picking task: max is %v, min is %v", maxItemVolume, minItemVolume)
	}
	baseCons := new(BaseConstraint)
	baseCons.SetName("SubPickingTaskItemVolumeConstraint")
	baseCons.SetLevel(level)
	return &SubPickingTaskItemVolumeConstraint{
		BaseConstraint: baseCons,
		minItemVolume:  minItemVolume,
		maxItemVolume:  maxItemVolume,
	}, nil
}
