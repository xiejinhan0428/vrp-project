package waveoptsolver

import (
	"errors"
	"fmt"
	"sort"
)

type DivideGranularity string

const (
	DivideByNone        DivideGranularity = "None"
	DivideByZone        DivideGranularity = "Zone"
	DivideByZoneCluster DivideGranularity = "ZoneCluster"
	DivideByZoneSector  DivideGranularity = "ZoneSector"
)

type Divider interface {
	Divide(*SolverWavePickingTask) []*SolverWaveSubPickingTask
}

type LocationBasedDivider struct {
	Granularity                 DivideGranularity
	MaxSubPickingTaskItemQty    int64
	MinSubPickingTaskItemQty    int64
	MaxSubPickingTaskItemVolume int64
	MinSubPickingTaskItemVolume int64
	EmptySubPickingTaskQty      int64
}

func NewLocationBasedDivider(granularity DivideGranularity, maxSubPickingTaskItemQty int64, minSubPickingTaskItemQty int64, maxSubPickingTaskItemVolume int64, minSubPickingTaskItemVolume int64) (*LocationBasedDivider, error) {
	if maxSubPickingTaskItemQty < 1 {
		return nil, errors.New("cannot init a LocationBasedDivider with maxSubPickingTaskItemQty < 1")
	}

	if maxSubPickingTaskItemVolume < 1 {
		return nil, errors.New("cannot init a LocationBasedDivider with non-positive maxSubPickingTaskItemVolume")
	}

	return &LocationBasedDivider{
		Granularity:                 granularity,
		MaxSubPickingTaskItemQty:    maxSubPickingTaskItemQty,
		MinSubPickingTaskItemQty:    minSubPickingTaskItemQty,
		MaxSubPickingTaskItemVolume: maxSubPickingTaskItemVolume,
		MinSubPickingTaskItemVolume: minSubPickingTaskItemVolume,
	}, nil
}

func NewDivideByNoneDivider() (*LocationBasedDivider, error) {
	return NewLocationBasedDivider(DivideByNone, 2, 0, 2, 0)
}

func (d *LocationBasedDivider) Divide(pickingTask *SolverWavePickingTask) []*SolverWaveSubPickingTask {
	if d.Granularity == DivideByNone {
		skus := make([]*SolverWaveSku, 0)
		for _, order := range pickingTask.Orders {
			skus = append(skus, order.Skus...)
		}

		subTask := &SolverWaveSubPickingTask{
			PickingTask: pickingTask,
			Skus:        append(make([]*SolverWaveSku, 0), skus...),
		}

		return append(make([]*SolverWaveSubPickingTask, 0), subTask)
	}

	taskSkus := make([]*SolverWaveSku, 0)
	for _, order := range pickingTask.Orders {
		taskSkus = append(taskSkus, order.Skus...)
	}

	granularityToSkusMap := groupByDivideGranularity(taskSkus, d.Granularity)
	subTasks := make([]*SolverWaveSubPickingTask, 0)
	for _, skus := range granularityToSkusMap {

		// init empty subtasks
		// subtask # == sku #
		subTaskNum := len(skus)
		granularitySubTasks := make([]*SolverWaveSubPickingTask, subTaskNum)
		for i := 0; i < subTaskNum; i++ {
			granularitySubTasks[i] = &SolverWaveSubPickingTask{
				PickingTask: pickingTask,
				Skus:        make([]*SolverWaveSku, 0),
			}
		}

		// sort skus by location
		sort.SliceStable(skus, func(i, j int) bool {
			left := skus[i]
			right := skus[j]
			if left.Location.ZoneSectorId != right.Location.ZoneSectorId {
				return left.Location.ZoneSectorId < right.Location.ZoneSectorId
			} else if left.Location.ZoneClusterId != right.Location.ZoneClusterId {
				return left.Location.ZoneClusterId < right.Location.ZoneClusterId
			} else if left.Location.ZoneId != right.Location.ZoneId {
				return left.Location.ZoneId < right.Location.ZoneId
			} else if left.Location.PathwayId != right.Location.PathwayId {
				return left.Location.PathwayId < right.Location.PathwayId
			} else if left.Location.SegmentId != right.Location.SegmentId {
				return left.Location.SegmentId < right.Location.SegmentId
			} else {
				return left.Location.LocationId < right.Location.LocationId
			}
		})

		// insert sku one by one into subtasks, first fit
		for _, sku := range skus {
			for _, subTask := range granularitySubTasks {
				subTaskTotalSkuQty := subTask.TotalSkuQty()
				if sku.Qty+subTaskTotalSkuQty <= d.MaxSubPickingTaskItemQty && sku.TotalVolume()+subTask.TotalVolume() <= d.MaxSubPickingTaskItemVolume {
					subTask.Skus = append(subTask.Skus, sku)
					break
				}
			}
		}

		// put non-empty subtask into subTasks and name them

		for _, subTask := range granularitySubTasks {
			if len(subTask.Skus) > 0 {
				subTasks = append(subTasks, subTask)
			}
		}
	}

	taskId := pickingTask.Id
	subTaskIdx := 0
	for _, subTask := range subTasks {
		subTaskId := fmt.Sprintf("%s_SubPickingTask_%d", taskId, subTaskIdx)
		subTask.Id = subTaskId
		subTaskIdx += 1
	}

	return subTasks
}

func groupByDivideGranularity(skus []*SolverWaveSku, granularity DivideGranularity) map[string][]*SolverWaveSku {
	granularityToSkusMap := make(map[string][]*SolverWaveSku)
	for _, sku := range skus {
		skuGranularity := extractSkuDivideGranularity(sku, granularity)

		_, ok := granularityToSkusMap[skuGranularity]
		if !ok {
			granularityToSkusMap[skuGranularity] = make([]*SolverWaveSku, 0)
		}

		granularityToSkusMap[skuGranularity] = append(granularityToSkusMap[skuGranularity], sku)
	}

	return granularityToSkusMap
}

func extractSkuDivideGranularity(sku *SolverWaveSku, granularity DivideGranularity) string {
	switch granularity {
	case DivideByZone:
		return sku.Location.ZoneId
	case DivideByZoneCluster:
		return sku.Location.ZoneClusterId
	case DivideByZoneSector:
		return sku.Location.ZoneSectorId
	default:
		panic(fmt.Sprintf("WaveOptSolver: Unknown DivideGranular %v of picking task", string(granularity)))
	}
}

//type SinglePickerDivider struct{}
//
//func NewSinglePickerDivider() *SinglePickerDivider {
//	return &SinglePickerDivider{}
//}
//
//func (d *SinglePickerDivider) Divide(pickingTask *SolverWavePickingTask) []*SolverWaveSubPickingTask {
//	skus := make([]*SolverWaveSku, 0)
//	for _, order := range pickingTask.OrderSkus {
//		skus = append(skus, order.Skus...)
//	}
//
//	subTask := &SolverWaveSubPickingTask{
//		PickingTask: pickingTask,
//		Skus:        append(make([]*SolverWaveSku, 0), skus...),
//	}
//
//	return append(make([]*SolverWaveSubPickingTask, 0), subTask)
//}
