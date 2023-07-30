package waveoptsolver

import (
	"fmt"
	"git.garena.com/shopee/bg-logistics/algo/wms/waveopt/localsearch/pkg/solver"
	"math"
	"math/rand"
	"sort"
	"time"
)

type WavePickingSingleChangeMoveFactory struct {
	*solver.BaseMoveFactory
	rule   *WaveRule
	config *WaveSolverConfig
}

type WavePickingSingleSwapChangeMove struct {
	i           int //SelectedPickingTask1
	j           int //SelectedPickingTask2
	k           int //SelectedOrder1
	l           int //SelectedOrder2
	OldSolution *WaveOptSolutionSingle
}

func (cm *WavePickingSingleSwapChangeMove) String() string {
	return fmt.Sprintf("{PiakcingTask[%v]-Order[%v]} -> {PiakcingTask[%v]-Order[%v]}", cm.i, cm.k, cm.j, cm.l)
}

// implements Move interface

func (cm *WavePickingSingleSwapChangeMove) Do(solution solver.Solution) (solver.Move, error) {

	sol := cm.OldSolution
	sol.PickingTasks[cm.i].Orders[cm.k], sol.PickingTasks[cm.j].Orders[cm.l] = sol.PickingTasks[cm.j].Orders[cm.l], sol.PickingTasks[cm.i].Orders[cm.k]

	//fmt.Println(cm.String())
	//fmt.Println("new_solution_len:")
	//fmt.Println(len(cm.OldSolution.PickingTasks))

	return cm, nil
}

func (cm *WavePickingSingleSwapChangeMove) MovedVariables() ([]solver.Variable, error) {
	vars := make([]solver.Variable, 0)
	return vars, nil
}

func (cm *WavePickingSingleSwapChangeMove) ToValues() ([]solver.Value, error) {
	vals := make([]solver.Value, 0)
	return vals, nil
}

func (mf *WavePickingSingleChangeMoveFactory) CreateMove(solution solver.Solution) (solver.Move, error) {

	oldSol, _ := solution.(*WaveOptSolutionSingle)
	//fmt.Println("old_solution_len:")
	//fmt.Println(len(oldSol.PickingTasks))

	move := SwapOperator(oldSol)

	return move, nil
}

// SwapOperator 交换算子
func SwapOperator(sol *WaveOptSolutionSingle) solver.Move {
	rand.Seed(time.Now().UnixNano())
	unNilTaskIdx := make([]int, 0)
	nilTaskNum := 0
	for idx, task := range sol.PickingTasks {
		if len(task.Orders) == 0 {
			nilTaskNum += 1
		} else {
			unNilTaskIdx = append(unNilTaskIdx, idx)
		}
	}
	if nilTaskNum == len(sol.PickingTasks) {
		ncm := new(solver.NoChangeMove)
		return ncm
	} else {
		i := unNilTaskIdx[rand.Intn(len(unNilTaskIdx))]
		j := unNilTaskIdx[rand.Intn(len(unNilTaskIdx))]

		selectedPickingTask1 := sol.PickingTasks[i].Orders
		selectedPickingTask2 := sol.PickingTasks[j].Orders
		//fmt.Println(selectedPickingTask1, selectedPickingTask2)
		k := rand.Intn(len(selectedPickingTask1))
		l := rand.Intn(len(selectedPickingTask2))
		//fmt.Println(i, j, k, l)

		move := &WavePickingSingleSwapChangeMove{
			i:           i,
			j:           j,
			k:           k,
			l:           l,
			OldSolution: sol,
		}

		return move

	}
}

// destroy & repair move factory config

type WavePickingSingleDnrChangeMoveFactory struct {
	*solver.BaseMoveFactory

	rule        *WaveRule
	config      *WaveSolverConfig
	greedyIdx   int
	greedyInfos []*DestroyInfo
}

func (mf *WavePickingSingleDnrChangeMoveFactory) UpdateAtStepStart(stepContext *solver.StepContext) error {
	var err error

	sol_ := stepContext.BeforeStepSolution
	sol := sol_.(*WaveOptSolutionSingle)

	GreedyDestroyInfos := make([]*DestroyInfo, 0)
	tmpSolution_, _ := sol.Copy()

	for pickingTaskIndex, pickingTask := range sol.PickingTasks {
		oldSolutionDistance := PickingTaskDistanceCalculate(sol.PickingTasks[pickingTaskIndex], mf.config, mf.rule)["distance"]
		oldSolutionConstraint := PickingTaskDistanceCalculate(sol.PickingTasks[pickingTaskIndex], mf.config, mf.rule)["constraint"]
		for orderIndex := range pickingTask.Orders {
			tmpSolution := tmpSolution_.(*WaveOptSolutionSingle)
			tmpSol := tmpSolution.PickingTasks[pickingTaskIndex].Orders

			destroyOrder := tmpSol[orderIndex]
			tmpSol = append(tmpSol[:orderIndex], tmpSol[orderIndex+1:]...)
			tmpGreedyInfos := &DestroyInfo{
				ReduceDistance:      oldSolutionDistance - PickingTaskDistanceCalculate(tmpSolution.PickingTasks[pickingTaskIndex], mf.config, mf.rule)["distance"],
				ReduceConstraint:    oldSolutionConstraint - PickingTaskDistanceCalculate(tmpSolution.PickingTasks[pickingTaskIndex], mf.config, mf.rule)["constraint"],
				DestroyedOrderIndex: orderIndex,
				DestroyedPtIndex:    pickingTaskIndex,
				DestroyOrder:        destroyOrder,
			}
			// restore the solution which is destroyed
			tmpOrders := make([]*SolverWaveOrder, 0)
			tmpOrders = append(tmpOrders, tmpSol[:orderIndex]...)
			tmpOrders = append(tmpOrders, destroyOrder)
			tmpOrders = append(tmpOrders, tmpSol[orderIndex:]...)

			tmpSolution.PickingTasks[pickingTaskIndex].Orders = tmpOrders

			// put the solution into the greedyInfo for repairing
			tmpGreedyInfos.Solution = tmpSolution
			GreedyDestroyInfos = append(GreedyDestroyInfos, tmpGreedyInfos)

		}
	}
	sort.Slice(GreedyDestroyInfos, func(i, j int) bool {
		if GreedyDestroyInfos[i].ReduceConstraint != GreedyDestroyInfos[j].ReduceConstraint {
			return GreedyDestroyInfos[i].ReduceConstraint > GreedyDestroyInfos[j].ReduceConstraint
		} else {
			return GreedyDestroyInfos[i].ReduceDistance > GreedyDestroyInfos[j].ReduceDistance
		}
	})
	mf.greedyInfos = GreedyDestroyInfos
	mf.greedyIdx = 0
	return err
}

// destroy&repair info

type WavePickingSingleDnrChangeMove struct {
	DestroyedOrderIndex int
	DestroyedPtIndex    int
	RepairPtIndex       int
}

func (cm *WavePickingSingleDnrChangeMove) String() string {
	return fmt.Sprintf("{DestoryPt[%v]-Order[%v]} -> {ReapirPiakcingTask[%v]}",
		cm.DestroyedPtIndex, cm.DestroyedOrderIndex, cm.RepairPtIndex)
}

// implements Move interface

func (cm *WavePickingSingleDnrChangeMove) Do(solution solver.Solution) (solver.Move, error) {

	solution_ := solution.(*WaveOptSolutionSingle)

	repairPt := solution_.PickingTasks[cm.RepairPtIndex]
	destroyPt := solution_.PickingTasks[cm.DestroyedPtIndex]
	destroyOrder := destroyPt.Orders[cm.DestroyedOrderIndex]
	newOrders := make([]*SolverWaveOrder, 0)
	newOrders = append(newOrders, destroyPt.Orders[:cm.DestroyedOrderIndex]...)
	newOrders = append(newOrders, destroyPt.Orders[cm.DestroyedOrderIndex+1:]...)
	destroyPt.Orders = newOrders
	repairPt.Orders = append(repairPt.Orders, destroyOrder)

	return &WavePickingSingleDnrChangeMove{
		DestroyedPtIndex:    cm.RepairPtIndex,
		RepairPtIndex:       cm.DestroyedPtIndex,
		DestroyedOrderIndex: len(repairPt.Orders) - 1, // 插入原来repair的最后一位，所以取长度即可
	}, nil
}

func (cm *WavePickingSingleDnrChangeMove) MovedVariables() ([]solver.Variable, error) {
	vars := make([]solver.Variable, 0)
	return vars, nil
}

func (cm *WavePickingSingleDnrChangeMove) ToValues() ([]solver.Value, error) {
	vals := make([]solver.Value, 0)
	return vals, nil
}

// the evaluation about destroy&repair calculator

func PickingTaskDistanceCalculate(pt *PickingTaskSingle, config *WaveSolverConfig, rule *WaveRule) map[string]int64 {
	// 计算pickingtask指标
	tmpPosition := make([]*SolverSkuLocation, 0)
	resultMap := make(map[string]int64)
	var volume int64
	var constraintViolation int64

	for _, order := range pt.Orders {
		for _, sku := range order.Skus {
			tmpPosition = append(tmpPosition, sku.Location)
			volume += sku.totalVolume
		}
	}

	//distance := CalculateDistance(tmpPosition, config)

	distance := CalculateDistance(tmpPosition, config)
	if volume > rule.CommonRule.MaxItemVolumePerSubPickingTask {
		constraintViolation += pt.Volume - rule.CommonRule.MaxItemVolumePerSubPickingTask
	}
	if volume < rule.CommonRule.MinItemVolumePerSubPickingTask {
		constraintViolation += rule.CommonRule.MinItemVolumePerSubPickingTask - pt.Volume
	}
	resultMap["distance"] = int64(distance)
	resultMap["constraint"] = constraintViolation

	return resultMap
}

// destroy calculator information

type DestroyInfo struct {
	ReduceDistance      int64
	ReduceConstraint    int64
	Solution            *WaveOptSolutionSingle
	DestroyOrder        *SolverWaveOrder
	DestroyedOrderIndex int
	DestroyedPtIndex    int
}

// repair calculator information

type RepairInfo struct {
	IncreaseDistance   int64
	IncreaseConstraint int64
	RepairPtIdx        int
}

// random destroy calculator

func RandomDestroy(sol *WaveOptSolutionSingle, unNilTaskIdx []int) *DestroyInfo {
	tmpSolution_, _ := sol.Copy()
	tmpSolution := tmpSolution_.(*WaveOptSolutionSingle)
	destroyPtIdx := unNilTaskIdx[rand.Intn(len(unNilTaskIdx))]
	destroyPt := tmpSolution.PickingTasks[destroyPtIdx]
	destroyOrderIdx := rand.Intn(len(destroyPt.Orders))
	destroyOrder := destroyPt.Orders[destroyOrderIdx]
	//destroyPt.Orders = append(destroyPt.Orders[:destroyOrderIdx], destroyPt.Orders[destroyOrderIdx+1:]...)
	destroyInfo := &DestroyInfo{
		ReduceDistance:      0.0,
		Solution:            tmpSolution,
		DestroyedOrderIndex: destroyOrderIdx,
		DestroyedPtIndex:    destroyPtIdx,
		DestroyOrder:        destroyOrder,
	}

	return destroyInfo
}

// greedy destroy calculator

func GreedyDestroy(sol *WaveOptSolutionSingle, mf *WavePickingSingleDnrChangeMoveFactory, unNilTaskIdx []int) *DestroyInfo {

	//GreedyDestroyInfos := make([]*DestroyInfo, 0)

	//bestGreedyInfo := &DestroyInfo{
	//	ReduceDistance:      math.MinInt64,
	//	Solution:            nil,
	//	DestroyOrder:        nil,
	//	DestroyedOrderIndex: 0,
	//	DestroyedPtIndex:    0,
	//}
	//tmpSolution_, _ := sol.Copy()
	//for pickingTaskIndex, pickingTask := range sol.PickingTasks {
	//	oldSolutionDistance := PickingTaskDistanceCalculate(sol.PickingTasks[pickingTaskIndex], config)
	//	for orderIndex := range pickingTask.Orders {
	//		tmpSolution := tmpSolution_.(*WaveOptSolutionSingle)
	//		tmpSol := tmpSolution.PickingTasks[pickingTaskIndex].Orders
	//		//fmt.Println("before destroy")
	//		//fmt.Println(tmpSol)
	//		destroyOrder := tmpSol[orderIndex]
	//		tmpSol = append(tmpSol[:orderIndex], tmpSol[orderIndex+1:]...)
	//		tmpGreedyInfos := &DestroyInfo{
	//			ReduceDistance:      oldSolutionDistance - PickingTaskDistanceCalculate(tmpSolution.PickingTasks[pickingTaskIndex], config),
	//			DestroyedOrderIndex: orderIndex,
	//			DestroyedPtIndex:    pickingTaskIndex,
	//			DestroyOrder:        destroyOrder,
	//		}
	//		// restore the solution which is destroyed
	//		tmpOrders := make([]*SolverWaveOrder, 0)
	//		tmpOrders = append(tmpOrders, tmpSol[:orderIndex]...)
	//		tmpOrders = append(tmpOrders, destroyOrder)
	//		tmpOrders = append(tmpOrders, tmpSol[orderIndex:]...)
	//		tmpSolution.PickingTasks[pickingTaskIndex].Orders = tmpOrders
	//
	//		// put the solution into the greedyInfo for repairing
	//		tmpGreedyInfos.Solution = tmpSolution
	//		//fmt.Println("after destroy")
	//		//fmt.Println(tmpSolution.PickingTasks[pickingTaskIndex].Orders)
	//
	//		if tmpGreedyInfos.ReduceDistance > bestGreedyInfo.ReduceDistance {
	//			bestGreedyInfo = tmpGreedyInfos
	//		}
	//	}
	//}
	////sort.Slice(GreedyDestroyInfos, func(i, j int) bool {
	////	return GreedyDestroyInfos[i].ReduceDistance > GreedyDestroyInfos[j].ReduceDistance
	////})
	var greedyInfo *DestroyInfo
	if mf.greedyIdx >= len(mf.greedyInfos) {
		greedyInfo = RandomDestroy(sol, unNilTaskIdx)
	} else {
		greedyInfo = mf.greedyInfos[mf.greedyIdx]
		mf.greedyIdx++
	}
	return greedyInfo
}

// random repair calculator

func RandomRepair(destroyInfo *DestroyInfo) *WavePickingSingleDnrChangeMove {
	sol := destroyInfo.Solution
	repairPtIndex := rand.Intn(len(sol.PickingTasks))
	for {
		if repairPtIndex != destroyInfo.DestroyedPtIndex {
			break
		}
		repairPtIndex = rand.Intn(len(sol.PickingTasks))
	}

	dnrInfo := &WavePickingSingleDnrChangeMove{
		DestroyedOrderIndex: destroyInfo.DestroyedOrderIndex,
		DestroyedPtIndex:    destroyInfo.DestroyedPtIndex,
		RepairPtIndex:       repairPtIndex,
	}
	return dnrInfo
}

// 贪心插入算子

func GreedyRepair(destroyInfo *DestroyInfo, mf *WavePickingSingleDnrChangeMoveFactory) *WavePickingSingleDnrChangeMove {
	sol_ := destroyInfo.Solution
	destroyOrder := destroyInfo.DestroyOrder
	//repairInfos := make([]*RepairInfo, 0)
	bestRepairInfo := &RepairInfo{
		IncreaseDistance:   math.MaxInt64,
		IncreaseConstraint: math.MaxInt64,
		RepairPtIdx:        0,
	}
	tmpSolution_, _ := sol_.Copy()
	tmpSolution := tmpSolution_.(*WaveOptSolutionSingle)

	//make a new slice to save the destroyed picking task
	destroyPt := tmpSolution.PickingTasks[destroyInfo.DestroyedPtIndex]
	//fmt.Println("order_len:")
	//fmt.Println(len(destroyPt.Orders))
	//fmt.Println("destroy_order")
	//fmt.Println(destroyInfo.DestroyedOrderIndex)
	//fmt.Println(destroyInfo.DestroyedPtIndex)
	tmpOrders := make([]*SolverWaveOrder, 0)
	tmpOrders = append(tmpOrders, destroyPt.Orders[:destroyInfo.DestroyedOrderIndex]...)
	tmpOrders = append(tmpOrders, destroyPt.Orders[destroyInfo.DestroyedOrderIndex+1:]...)
	tmpSolution.PickingTasks[destroyInfo.DestroyedPtIndex].Orders = tmpOrders

	for pickingTaskIndex := range tmpSolution.PickingTasks {
		//tmpSolution_, _ := sol.Copy()
		if pickingTaskIndex == destroyInfo.DestroyedPtIndex {
			continue
		}
		tmpSolution.PickingTasks[pickingTaskIndex].Orders = append(tmpSolution.PickingTasks[pickingTaskIndex].Orders, destroyOrder)
		tmpRepairInfo := &RepairInfo{
			IncreaseDistance: PickingTaskDistanceCalculate(tmpSolution.PickingTasks[pickingTaskIndex], mf.config, mf.rule)["distance"] -
				PickingTaskDistanceCalculate(sol_.PickingTasks[pickingTaskIndex], mf.config, mf.rule)["distance"],
			IncreaseConstraint: PickingTaskDistanceCalculate(tmpSolution.PickingTasks[pickingTaskIndex], mf.config, mf.rule)["constraint"] -
				PickingTaskDistanceCalculate(sol_.PickingTasks[pickingTaskIndex], mf.config, mf.rule)["constraint"],
			RepairPtIdx: pickingTaskIndex,
		}

		if tmpRepairInfo.IncreaseConstraint < bestRepairInfo.IncreaseConstraint {
			bestRepairInfo = tmpRepairInfo
		}
		if tmpRepairInfo.IncreaseConstraint == bestRepairInfo.IncreaseConstraint {
			if tmpRepairInfo.IncreaseDistance < bestRepairInfo.IncreaseDistance {
				bestRepairInfo = tmpRepairInfo
			}
		}
		tmpSolution.PickingTasks[pickingTaskIndex].Orders = tmpSolution.PickingTasks[pickingTaskIndex].Orders[:len(tmpSolution.PickingTasks[pickingTaskIndex].Orders)-1]
	}
	//sort.Slice(repairInfos, func(i, j int) bool {
	//	return repairInfos[i].IncreaseDistance < repairInfos[j].IncreaseDistance
	//})
	dnrInfo := &WavePickingSingleDnrChangeMove{
		DestroyedOrderIndex: destroyInfo.DestroyedOrderIndex,
		DestroyedPtIndex:    destroyInfo.DestroyedPtIndex,
		RepairPtIndex:       bestRepairInfo.RepairPtIdx,
	}
	return dnrInfo
}

// destroy&repair

func (mf *WavePickingSingleDnrChangeMoveFactory) CreateMove(solution solver.Solution) (solver.Move, error) {
	//fmt.Println("move starting")
	//fmt.Println(time.Now().Clock())
	destroySettings := append(make([]string, 0), "Random", "Greedy")
	repairSettings := append(make([]string, 0), "Random", "Greedy")
	solution_ := solution.(*WaveOptSolutionSingle)
	// 检验是否有空任务
	unNilTaskIdx := make([]int, 0)
	nilTaskNum := 0
	for idx, task := range solution_.PickingTasks {
		if len(task.Orders) == 0 {
			nilTaskNum += 1
		} else {
			unNilTaskIdx = append(unNilTaskIdx, idx)
		}
	}
	if nilTaskNum == len(solution_.PickingTasks) {
		ncm := new(solver.NoChangeMove)
		return ncm, nil
	}
	// 对初始解进行破坏
	ranNum := rand.Intn(len(destroySettings))
	selectedDestroyCalculator := destroySettings[ranNum]

	destroyInfo := new(DestroyInfo)
	if selectedDestroyCalculator == "Random" {
		destroyInfo = RandomDestroy(solution_, unNilTaskIdx)
	} else {
		destroyInfo = GreedyDestroy(solution_, mf, unNilTaskIdx)
	}

	//destroyInfo = GreedyDestroy(solution_, mf.config)
	//fmt.Println("**********")

	//for _, sku := range destroyInfo.DestroyOrder.Skus {
	//	//fmt.Println(sku.Location)
	//}
	//fmt.Println("**********")

	ranNum = rand.Intn(len(repairSettings))
	selectedInsertCalculator := repairSettings[ranNum]

	move := new(WavePickingSingleDnrChangeMove)
	//move = GreedyRepair(destroyInfo, mf.config)
	if selectedInsertCalculator == "Random" {
		move = RandomRepair(destroyInfo)
	} else {
		move = GreedyRepair(destroyInfo, mf)
	}

	//fmt.Println("CreateMove:" + move.String())
	//fmt.Println("move ending")

	//fmt.Println(time.Now().Clock())

	return move, nil
}
