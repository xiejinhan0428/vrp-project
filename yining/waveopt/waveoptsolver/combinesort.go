package waveoptsolver

import (
	"fmt"
	"git.garena.com/shopee/bg-logistics/go/gocommon/logger"
	"math"
	"sort"
)

type taskFeature struct {
	id                      string
	groupPriority           int64
	urgentOrderNum          int64
	avgOrderDist            int64
	itemNum                 int64
	orderNum                int64
	isMultiPickerTask       bool
	crossSplitLevelOrderNum int64
}

func combineAndSort(waveSn string, tasks []*SolverWavePickingTask, orphanOrders []*SolverWaveOrder, rule *WaveRule, solverConfig *WaveSolverConfig, groupErrMsg string) *WaveResult {
	logger.LogInfof("WaveOptAlgo - %s: combining and sorting group results", waveSn)
	var err error

	// a distance calculator for avg order distance calculation
	distanceCalculator, err := NewLocationBasedDistanceCalculator(
		solverConfig.ZoneCoeff,
		solverConfig.PathwayCoeff,
		solverConfig.SegmentCoeff,
	)
	if err != nil {
		return errorWaveResult(waveSn, UnknownSolverErrorResult, "fail to init a distance calculator when sorting tasks")
	}

	// convert SolverWavePickingTask to WavePickingTask
	// and extract task features
	idToTaskMap := make(map[string]*WavePickingTask)
	taskFeatures := make([]*taskFeature, 0)
	orderNoToSizeMap := make(map[string]TaskSizeType)
	for _, task := range tasks {
		orderNos := make([]string, 0)
		for _, order := range task.Orders {
			orderNos = append(orderNos, order.Id)
		}

		taskVolume := int64(0)
		taskWeight := int64(0)
		taskSizeMap := make(map[TaskSizeType]bool)

		// convert SolverWaveSubPickingTask to WaveSubPickingTask
		waveSubPickingTasks := make([]*WaveSubPickingTask, 0)
		for _, subTask := range task.SubPickingTasks {
			// convert SolverWaveSku to OrderSku, mark sub task size
			subTaskVolume := int64(0)
			subTaskWeight := int64(0)
			subTaskSizeMap := make(map[TaskSizeType]bool)
			orderSkus := make([]*OrderSku, 0)
			for _, sku := range subTask.Skus {
				orderSku := &OrderSku{
					OrderNo:    sku.Order.Id,
					SkuId:      sku.Id,
					Qty:        sku.Qty,
					LocationId: sku.Location.LocationId,
				}

				orderSkus = append(orderSkus, orderSku)

				subTaskVolume += sku.TotalVolume()
				subTaskWeight += sku.TotalWeight()

				var taskSizeFromSkuSize TaskSizeType
				if sku.SkuSize == ExtraBulkySkuType {
					taskSizeFromSkuSize = ExtraBulkyTaskType
				} else if sku.SkuSize == BulkySkuType {
					taskSizeFromSkuSize = BulkyTaskType
				} else {
					taskSizeFromSkuSize = NonBulkyTaskType
				}

				subTaskSizeMap[taskSizeFromSkuSize] = true
				taskSizeMap[taskSizeFromSkuSize] = true
			}

			// sku size of volume
			if subTaskVolume > rule.CommonRule.MaxBulkyTaskVolume {
				subTaskSizeMap[ExtraBulkyTaskType] = true
			} else if subTaskVolume > rule.CommonRule.MaxNonBulkyTaskVolume && subTaskVolume <= rule.CommonRule.MaxBulkyTaskVolume {
				subTaskSizeMap[BulkyTaskType] = true
			} else {
				subTaskSizeMap[NonBulkyTaskType] = true
			}

			// sku size of weight
			if subTaskWeight > rule.CommonRule.MaxBulkyTaskLoad {
				subTaskSizeMap[ExtraBulkyTaskType] = true
			} else if subTaskWeight > rule.CommonRule.MaxNonBulkyTaskLoad && subTaskVolume <= rule.CommonRule.MaxBulkyTaskLoad {
				subTaskSizeMap[BulkyTaskType] = true
			} else {
				subTaskSizeMap[NonBulkyTaskType] = true
			}

			subPickingTaskSize := NonBulkyTaskType
			if hit, ok := subTaskSizeMap[ExtraBulkyTaskType]; ok && hit {
				subPickingTaskSize = ExtraBulkyTaskType
			} else if hit, ok = subTaskSizeMap[BulkyTaskType]; ok && hit {
				subPickingTaskSize = BulkyTaskType
			}

			waveSubPickingTask := &WaveSubPickingTask{
				Id:                 subTask.Id,
				Skus:               orderSkus,
				SubPickingTaskSize: subPickingTaskSize,
			}

			waveSubPickingTasks = append(waveSubPickingTasks, waveSubPickingTask)

			taskVolume += subTaskVolume
			taskWeight += subTaskWeight
		}

		var pickerMode PickingTaskPickerModeType
		if task.isMultiPickerTask {
			pickerMode = MultiPickerType
		} else {
			pickerMode = SinglePickerType
		}

		if taskVolume > rule.CommonRule.MaxBulkyTaskVolume {
			taskSizeMap[ExtraBulkyTaskType] = true
		} else if taskVolume > rule.CommonRule.MaxNonBulkyTaskVolume && taskVolume <= rule.CommonRule.MaxBulkyTaskVolume {
			taskSizeMap[BulkyTaskType] = true
		} else {
			taskSizeMap[NonBulkyTaskType] = true
		}

		if taskWeight > rule.CommonRule.MaxBulkyTaskLoad {
			taskSizeMap[ExtraBulkyTaskType] = true
		} else if taskWeight > rule.CommonRule.MaxNonBulkyTaskLoad && taskWeight <= rule.CommonRule.MaxBulkyTaskLoad {
			taskSizeMap[BulkyTaskType] = true
		} else {
			taskSizeMap[NonBulkyTaskType] = true
		}

		taskSize := NonBulkyTaskType
		if hit, ok := taskSizeMap[ExtraBulkyTaskType]; ok && hit {
			taskSize = ExtraBulkyTaskType
		} else if hit, ok = taskSizeMap[BulkyTaskType]; ok && hit {
			taskSize = BulkyTaskType
		}

		wavePickingTask := &WavePickingTask{
			PickingTaskId:   task.Id,
			GroupSn:         task.Group.Id,
			GroupType:       task.Group.GroupType,
			OrderNos:        orderNos,
			SubPickingTasks: waveSubPickingTasks,
			PickerMode:      pickerMode,
			PickingTaskSize: taskSize,
		}

		idToTaskMap[wavePickingTask.PickingTaskId] = wavePickingTask

		// new a taskFeature
		feature := &taskFeature{
			id:                      task.Id,
			groupPriority:           task.Group.Priority,
			urgentOrderNum:          0,
			avgOrderDist:            0,
			itemNum:                 0,
			orderNum:                int64(len(task.Orders)),
			isMultiPickerTask:       task.isMultiPickerTask,
			crossSplitLevelOrderNum: 0,
		}

		// count urgent order #
		for _, order := range task.Orders {
			if order.isUrgent {
				feature.urgentOrderNum += 1
			}
		}

		// calculate avg order distance
		totalOrderDist := int64(0)
		for _, subTask := range task.SubPickingTasks {
			totalOrderDist += int64(distanceCalculator.CalculateDistance(subTask))
		}
		feature.avgOrderDist = totalOrderDist / int64(len(task.Orders))

		// count item #
		for _, order := range task.Orders {
			for _, sku := range order.Skus {
				feature.itemNum += sku.Qty
			}
		}

		// skip crossSplitLevelOrderNum counting in SinglePickerOnly mode
		if rule.WavePickerMode != SinglePickerOnly {
			// count crossSplitLevelOrderNum
			var splitLevel SplitLevel
			splitLevel, err = rule.splitLevel()
			if err != nil {
				return errorWaveResult(waveSn, UnreasonableConstraintsResult, err.Error())
			}

			for _, order := range task.Orders {
				crossMap := make(map[string]bool)
				for _, sku := range order.Skus {
					crossLocation, err := extractSplitLevel(sku.Location, splitLevel)
					if err != nil {
						return errorWaveResult(waveSn, UnreasonableConstraintsResult, err.Error())
					}

					crossMap[crossLocation] = true
				}

				if len(crossMap) > 1 {
					feature.crossSplitLevelOrderNum += 1
				}
			}
		}

		taskFeatures = append(taskFeatures, feature)

		// record order size
		for _, order := range task.Orders {
			orderNoToSizeMap[order.Id] = wavePickingTask.PickingTaskSize
		}
	}

	filteredTasks := make([]*WavePickingTask, 0)
	filteredOrphanOrders := make([]string, 0)

	generatedTaskNum := len(idToTaskMap)
	if int64(generatedTaskNum) < rule.CommonRule.MinPickingTaskQtyPerWave {
		return errorWaveResult(waveSn, InfeasibleSolutionResult, fmt.Sprintf("at least %v tasks are required for a wave, but only %v found", rule.CommonRule.MinPickingTaskQtyPerWave, generatedTaskNum))
	}

	switch rule.WavePickerMode {
	case SinglePickerOnly:
		// sort the tasks by features
		sort.Slice(taskFeatures, func(i, j int) bool {
			left := taskFeatures[i]
			right := taskFeatures[j]

			if left.groupPriority != right.groupPriority {
				return left.groupPriority < right.groupPriority
			} else if left.urgentOrderNum != right.urgentOrderNum {
				return left.urgentOrderNum > right.urgentOrderNum
			} else if left.orderNum != right.orderNum {
				return left.orderNum > right.orderNum
			} else if left.itemNum != right.itemNum {
				return left.itemNum > right.itemNum
			} else if left.avgOrderDist != right.avgOrderDist {
				return left.avgOrderDist < right.avgOrderDist
			} else {
				return left.id < right.id
			}
		})

		outputTaskNum := int(math.Min(float64(rule.CommonRule.MaxPickingTaskQtyPerWave), float64(generatedTaskNum)))

		// match real tasks
		for i := 0; i < outputTaskNum; i++ {
			id := taskFeatures[i].id
			if task, ok := idToTaskMap[id]; ok {
				filteredTasks = append(filteredTasks, task)
			} else {
				return errorWaveResult(waveSn, UnknownSolverErrorResult, fmt.Sprintf("cannot find picking task: %v", id))
			}

			delete(idToTaskMap, id)
		}
	case MultiPickerAtMWSWithTotalOrderQty:
		// sort the tasks by features
		sort.Slice(taskFeatures, func(i, j int) bool {
			left := taskFeatures[i]
			right := taskFeatures[j]

			if left.groupPriority != right.groupPriority {
				return left.groupPriority < right.groupPriority
			} else if left.isMultiPickerTask != right.isMultiPickerTask {
				return left.isMultiPickerTask
			} else if left.crossSplitLevelOrderNum != right.crossSplitLevelOrderNum && left.isMultiPickerTask && right.isMultiPickerTask {
				return left.crossSplitLevelOrderNum > right.crossSplitLevelOrderNum
			} else if left.urgentOrderNum != right.urgentOrderNum {
				return left.urgentOrderNum > right.urgentOrderNum
			} else if left.orderNum != right.orderNum {
				return left.orderNum > right.orderNum
			} else if left.itemNum != right.itemNum {
				return left.itemNum > right.itemNum
			} else if left.avgOrderDist != right.avgOrderDist {
				return left.avgOrderDist < right.avgOrderDist
			} else {
				return left.id < right.id
			}
		})

		maxOutputTaskNum := int(math.Min(float64(rule.CommonRule.MaxPickingTaskQtyPerWave), float64(generatedTaskNum)))
		modeRule, ok := rule.ModeRule.(*MultiPickerAtMWSTotalQtyRule)
		if !ok {
			return errorWaveResult(waveSn, UnreasonableConstraintsResult, "peculiar rules in MultiPickerAtMWSWithTotalOrderQty mode must be MultiPickerAtMWSTotalQtyRule")
		}
		maxBulkyNonBulkyOrderNum := modeRule.MaxBacklogAtMWSPerWave

		outputBulkyNonBulkyOrderNum := 0
		outputTaskNum := 0
		for i := 0; i < generatedTaskNum; i++ {
			id := taskFeatures[i].id
			if task, ok := idToTaskMap[id]; ok {
				multiTaskOrderNum := 0
				if task.PickerMode == MultiPickerType {
					multiTaskOrderNum = len(task.OrderNos)

					if int64(outputBulkyNonBulkyOrderNum+len(task.OrderNos)) > maxBulkyNonBulkyOrderNum {
						continue
					}
				}
				if outputTaskNum+1 <= maxOutputTaskNum {
					filteredTasks = append(filteredTasks, task)
					delete(idToTaskMap, id)
					outputBulkyNonBulkyOrderNum += multiTaskOrderNum
					outputTaskNum += 1
				}
			} else {
				return errorWaveResult(waveSn, UnknownSolverErrorResult, fmt.Sprintf("cannot find picking task: %v", id))
			}
		}
	case MultiPickerAtMWSWithRespectiveOrderQty:
		// sort the tasks by features
		sort.Slice(taskFeatures, func(i, j int) bool {
			left := taskFeatures[i]
			right := taskFeatures[j]

			if left.groupPriority != right.groupPriority {
				return left.groupPriority < right.groupPriority
			} else if left.isMultiPickerTask != right.isMultiPickerTask {
				return left.isMultiPickerTask
			} else if left.crossSplitLevelOrderNum != right.crossSplitLevelOrderNum && left.isMultiPickerTask && right.isMultiPickerTask {
				return left.crossSplitLevelOrderNum > right.crossSplitLevelOrderNum
			} else if left.urgentOrderNum != right.urgentOrderNum {
				return left.urgentOrderNum > right.urgentOrderNum
			} else if left.orderNum != right.orderNum {
				return left.orderNum > right.orderNum
			} else if left.itemNum != right.itemNum {
				return left.itemNum > right.itemNum
			} else if left.avgOrderDist != right.avgOrderDist {
				return left.avgOrderDist < right.avgOrderDist
			} else {
				return left.id < right.id
			}
		})

		maxOutputTaskNum := int(math.Min(float64(rule.CommonRule.MaxPickingTaskQtyPerWave), float64(generatedTaskNum)))
		modeRule, ok := rule.ModeRule.(*MultiPickerAtMWSRespectiveQtyRule)
		if !ok {
			return errorWaveResult(waveSn, UnreasonableConstraintsResult, "peculiar rules in MultiPickerAtMWSWithRespectiveOrderQty mode must be MultiPickerAtMWSRespectiveQtyRule")
		}
		maxBulkyOrderNum := modeRule.MaxBulkyBacklogAtMWSPerWave
		maxNonBulkyOrderNum := modeRule.MaxNonBulkyBacklogAtMWSPerWave

		outputBulkyOrderNum := 0
		outputNonBulkyOrderNum := 0
		outputTaskNum := 0
		for i := 0; i < generatedTaskNum; i++ {
			id := taskFeatures[i].id
			if task, ok := idToTaskMap[id]; ok {
				multiTaskBulkyOrderNum := 0
				multiTaskNonBulkyOrderNum := 0
				if task.PickerMode == MultiPickerType {
					for _, orderNo := range task.OrderNos {
						if orderSize, okk := orderNoToSizeMap[orderNo]; okk {
							if orderSize == ExtraBulkyTaskType || orderSize == BulkyTaskType {
								multiTaskBulkyOrderNum += 1
							} else {
								multiTaskNonBulkyOrderNum += 1
							}
						} else {
							return errorWaveResult(waveSn, UnknownSolverErrorResult, fmt.Sprintf("cannot find the size type of order %v", orderNo))
						}
					}

					if int64(outputBulkyOrderNum+multiTaskBulkyOrderNum) > maxBulkyOrderNum || int64(outputNonBulkyOrderNum+multiTaskNonBulkyOrderNum) > maxNonBulkyOrderNum {
						continue
					}
				}

				if outputTaskNum+1 <= maxOutputTaskNum {
					filteredTasks = append(filteredTasks, task)
					delete(idToTaskMap, id)
					outputBulkyOrderNum += multiTaskBulkyOrderNum
					outputNonBulkyOrderNum += multiTaskNonBulkyOrderNum
					outputTaskNum += 1
				}
			} else {
				return errorWaveResult(waveSn, UnknownSolverErrorResult, fmt.Sprintf("cannot find picking task: %v", id))
			}
		}
	case MultiPickerAtMLWithTotalPickingTaskQty:
		// sort the tasks by features
		sort.Slice(taskFeatures, func(i, j int) bool {
			left := taskFeatures[i]
			right := taskFeatures[j]

			if left.groupPriority != right.groupPriority {
				return left.groupPriority < right.groupPriority
			} else if left.isMultiPickerTask != right.isMultiPickerTask {
				return left.isMultiPickerTask
			} else if left.crossSplitLevelOrderNum != right.crossSplitLevelOrderNum && left.isMultiPickerTask && right.isMultiPickerTask {
				return left.crossSplitLevelOrderNum > right.crossSplitLevelOrderNum
			} else if left.urgentOrderNum != right.urgentOrderNum {
				return left.urgentOrderNum > right.urgentOrderNum
			} else if left.orderNum != right.orderNum {
				return left.orderNum > right.orderNum
			} else if left.itemNum != right.itemNum {
				return left.itemNum > right.itemNum
			} else if left.avgOrderDist != right.avgOrderDist {
				return left.avgOrderDist < right.avgOrderDist
			} else {
				return left.id < right.id
			}
		})

		maxOutputTaskNum := int(math.Min(float64(rule.CommonRule.MaxPickingTaskQtyPerWave), float64(generatedTaskNum)))
		modeRule, ok := rule.ModeRule.(*MultiPickerAtMLTotalQtyRule)
		if !ok {
			return errorWaveResult(waveSn, UnreasonableConstraintsResult, "peculiar rules in MultiPickerAtMLWithTotalPickingTaskQty mode must be MultiPickerAtMLTotalQtyRule")
		}
		maxBulkyNonBulkyTaskNum := modeRule.MaxBacklogAtMLPerWave

		outputTaskNum := 0
		outputBulkyNonBulkyTaskNum := 0
		for i := 0; i < generatedTaskNum; i++ {
			id := taskFeatures[i].id
			if task, ok := idToTaskMap[id]; ok {
				multiTaskNum := 0
				if task.PickerMode == MultiPickerType {
					multiTaskNum += 1
					if int64(outputTaskNum+1) > maxBulkyNonBulkyTaskNum {
						continue
					}
				}

				if outputTaskNum+1 <= maxOutputTaskNum {
					filteredTasks = append(filteredTasks, task)
					delete(idToTaskMap, id)
					outputBulkyNonBulkyTaskNum += multiTaskNum
					outputTaskNum += 1
				}
			} else {
				return errorWaveResult(waveSn, UnknownSolverErrorResult, fmt.Sprintf("cannot find picking task: %v", id))
			}
		}
	case MultiPickerAtMLWithRespectivePickingTaskQty:
		// sort the tasks by features
		sort.Slice(taskFeatures, func(i, j int) bool {
			left := taskFeatures[i]
			right := taskFeatures[j]

			if left.groupPriority != right.groupPriority {
				return left.groupPriority < right.groupPriority
			} else if left.isMultiPickerTask != right.isMultiPickerTask {
				return left.isMultiPickerTask
			} else if left.crossSplitLevelOrderNum != right.crossSplitLevelOrderNum && left.isMultiPickerTask && right.isMultiPickerTask {
				return left.crossSplitLevelOrderNum > right.crossSplitLevelOrderNum
			} else if left.urgentOrderNum != right.urgentOrderNum {
				return left.urgentOrderNum > right.urgentOrderNum
			} else if left.orderNum != right.orderNum {
				return left.orderNum > right.orderNum
			} else if left.itemNum != right.itemNum {
				return left.itemNum > right.itemNum
			} else if left.avgOrderDist != right.avgOrderDist {
				return left.avgOrderDist < right.avgOrderDist
			} else {
				return left.id < right.id
			}
		})

		maxOutputTaskNum := int(math.Min(float64(rule.CommonRule.MaxPickingTaskQtyPerWave), float64(generatedTaskNum)))
		modeRule, ok := rule.ModeRule.(*MultiPickerAtMLRespectiveQtyRule)
		if !ok {
			return errorWaveResult(waveSn, UnreasonableConstraintsResult, "peculiar rules in MultiPickerAtMLWithRespectivePickingTaskQty mode must be MultiPickerAtMLRespectiveQtyRule")
		}
		maxBulkyTaskNum := modeRule.MaxBulkyBacklogAtMLPerWave
		maxNonBulkyTaskNum := modeRule.MaxNonBulkyBacklogAtMLPerWave

		outputTaskNum := 0
		outputBulkyTaskNum := 0
		outputNonBulkyTaskNum := 0
		for i := 0; i < generatedTaskNum; i++ {
			id := taskFeatures[i].id
			if task, ok := idToTaskMap[id]; ok {
				multiBulkyTaskNum := 0
				multiNonBulkyTaskNum := 0
				if task.PickerMode == MultiPickerType {
					if task.PickingTaskSize == ExtraBulkyTaskType || task.PickingTaskSize == BulkyTaskType {
						multiBulkyTaskNum += 1
					} else if task.PickingTaskSize == NonBulkyTaskType {
						multiNonBulkyTaskNum += 1
					}

					if int64(outputBulkyTaskNum+1) > maxBulkyTaskNum || int64(outputNonBulkyTaskNum+1) > maxNonBulkyTaskNum {
						continue
					}
				}

				if outputTaskNum+1 <= maxOutputTaskNum {
					filteredTasks = append(filteredTasks, task)
					delete(idToTaskMap, id)
					outputBulkyTaskNum += multiBulkyTaskNum
					outputNonBulkyTaskNum += multiNonBulkyTaskNum
					outputTaskNum += 1
				}
			} else {
				return errorWaveResult(waveSn, UnknownSolverErrorResult, fmt.Sprintf("cannot find picking task: %v", id))
			}
		}
	}

	// combine orphan orders
	for _, order := range orphanOrders {
		filteredOrphanOrders = append(filteredOrphanOrders, order.Id)
	}
	for _, task := range idToTaskMap {
		filteredOrphanOrders = append(filteredOrphanOrders, task.OrderNos...)
	}

	// if # of output tasks is below the MinPickingTaskQtyPerWave, report an infeasible error
	if int64(len(filteredTasks)) < rule.CommonRule.MinPickingTaskQtyPerWave {
		return errorWaveResult(waveSn, InfeasibleSolutionResult, "cannot generate picking tasks")
	}

	waveResult := &WaveResult{
		WaveSn:           waveSn,
		PickingTasks:     filteredTasks,
		UnsolvedOrderNos: filteredOrphanOrders,
		RetCode:          SuccessResult,
		Msg:              groupErrMsg,
	}

	logger.LogInfof("WaveOptAlgo - %s: %d picking tasks and %d unused orders are combined and sorted", waveSn, len(waveResult.PickingTasks), len(waveResult.UnsolvedOrderNos))
	return waveResult
}
