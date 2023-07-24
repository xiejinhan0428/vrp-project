package solver

import (
	"sort"
)

const (
	MaxProx = 5000
)

func GetInnerK(x1, x2, y1, y2 float64) float64 {
	var innerK float64
	innerK = (y2 - y1) / (x2 - x1)
	return innerK
}

func GetCapResMap(input *SolverInput) []int {
	var capResMap = make([]int, len(input.Resources))
	if input.Obj == 0 {
		for i := 0; i < len(input.Resources); i++ {
			capResMap[i] = i
		}
	} else {
		capResMap = SortRes(input)
	}
	return capResMap
}

func SortRes(input *SolverInput) []int {
	var resSortInfo []float64
	for r := 0; r < len(input.Resources); r++ {
		tmpCost := input.Resources[r].CostTier[0][2] * CostCoef / input.Resources[r].Capacity[3]
		resSortInfo = append(resSortInfo, tmpCost)
	}
	//resSortInfo2 := float64ToFloat64(resSortInfo)
	var tmpIdx = make([]int, len(resSortInfo))
	sli := sort.Float64Slice(resSortInfo)
	sorter := NewIndexSorter(sli, tmpIdx)
	sort.Sort(sorter)
	//sort.Sort(sorter)
	indexer := sorter.GetIndex()
	return indexer
}

func InitParaS2(solverInput *SolverInput) GPara {
	//solver入参转solver内部公共数据结构，采用指针赋值
	//task
	gPara := GPara{}
	gPara.NTask = len(solverInput.Tasks)
	gPara.Obj = solverInput.Obj

	//车辆信息
	//车辆排序
	gPara.CapResMap = GetCapResMap(solverInput)
	//CapResCost[resIdx][t][0]:LowerBound
	//CapResCost[resIdx][t][1]:UpperBound
	//CapResCost[resIdx][t][2]:outerK
	//CapResCost[resIdx][t][3]:LowerCost
	//CapResCost[resIdx][t][4]:UpperCost
	//CapResCost[resIdx][t][5]:innerK
	gPara.CapSeq = make([]int, len(solverInput.Resources), cap(solverInput.Resources))
	gPara.CapRes = make([][]float64, len(solverInput.Resources))
	gPara.CapResCost = make([][][]float64, len(solverInput.Resources))
	for i := 0; i < len(gPara.CapResMap); i++ {
		gPara.CapSeq[i] = solverInput.Resources[gPara.CapResMap[i]].Quantity
		gPara.CapRes[i] = solverInput.Resources[gPara.CapResMap[i]].Capacity
		gPara.CapResCost[i] = make([][]float64, 1)
		gPara.CapResCost[i][0] = []float64{-1, 0, 0, 0, 0, 0}
		for t := 0; t < len(solverInput.Resources[gPara.CapResMap[i]].CostTier); t++ {
			var tierInfo = make([]float64, 6)
			tierInfo[0] = solverInput.Resources[gPara.CapResMap[i]].CostTier[t][0]
			tierInfo[1] = solverInput.Resources[gPara.CapResMap[i]].CostTier[t][1]
			tierInfo[2] = solverInput.Resources[gPara.CapResMap[i]].CostTier[t][3]
			tierInfo[4] = solverInput.Resources[gPara.CapResMap[i]].CostTier[t][2] * CostCoef
			if t-1 < 0 {
				tierInfo[3] = 0
			} else {
				tierInfo[3] = solverInput.Resources[gPara.CapResMap[i]].CostTier[t-1][2] * CostCoef
			}
			tierInfo[5] = GetInnerK(tierInfo[0], tierInfo[1], tierInfo[3], tierInfo[4])
			gPara.CapResCost[i] = append(gPara.CapResCost[i], tierInfo)
		}
	}

	//Nodes信息
	gPara.Nodes = make([][]float64, len(solverInput.Nodes))
	for i := 0; i < len(solverInput.Nodes); i++ {
		var nodeLatLng = make([]float64, 2, 2)
		nodeLatLng[0] = solverInput.Nodes[i].Lat
		nodeLatLng[1] = solverInput.Nodes[i].Lng
		gPara.Nodes[i] = nodeLatLng
	}

	//tasks信息
	gPara.CapTask = make([][]float64, len(solverInput.Tasks))
	for i := 0; i < len(solverInput.Tasks); i++ {
		gPara.CapTask[i] = solverInput.Tasks[i].ReqCap
	}

	//taskNode信息
	gPara.TaskLoc = make([]int, len(solverInput.Tasks))
	for i := 0; i < len(solverInput.Tasks); i++ {
		gPara.TaskLoc[i] = solverInput.Tasks[i].NodeIdx
	}

	//cost信息
	gPara.Cost = make([][][][]float64, len(solverInput.EdgeInfos))
	for i := 0; i < len(solverInput.EdgeInfos); i++ {
		var edgeGraph = make([][][]float64, 2)
		edgeGraph[0] = solverInput.EdgeInfos[i].Distance
		edgeGraph[1] = solverInput.EdgeInfos[i].Duration
		gPara.Cost[i] = edgeGraph
	}
	return gPara
}

func InitProxS2(nTask int) [][]int {
	//初始化每个task id的全部neighbor
	//目前全联接
	var proxTasks = make([][]int, nTask)
	for i := 0; i < nTask; i++ {
		proxTasks[i] = make([]int, nTask)
		for j := 0; j < nTask; j++ {
			proxTasks[i][j] = j
		}
	}
	return proxTasks
}

func GetProxTask2(gPara *GPara) [][][]int { //3000
	//得到经过筛选的 task id 的部分neighbor
	//暂定逻辑为初始化
	var proxTasks = make([][][]int, len(gPara.CapRes))
	if gPara.NTask <= MaxProx {
		for i := 0; i < len(gPara.CapRes); i++ {
			proxTasks[i] = make([][]int, gPara.NTask)
			for x := 0; x < gPara.NTask; x++ {
				proxTasks[i][x] = make([]int, gPara.NTask)
				for y := 0; y < gPara.NTask; y++ {
					proxTasks[i][x][y] = y
				}
			}
		}
	} else {
		for i := 0; i < len(gPara.CapRes); i++ {
			proxTasks[i] = make([][]int, gPara.NTask)
			var tmpIdx = make([]int, gPara.NTask)
			//根据资源类型排序
			for x := 0; x < gPara.NTask; x++ {
				proxTasks[i][x] = make([]int, gPara.NTask)
				var tmpSlice = CopySliceF64(gPara.Cost[i][0][gPara.TaskLoc[x]][1:])
				//tmpSlice2 := float64ToFloat64(tmpSlice)
				sli := sort.Float64Slice(tmpSlice)
				sorter := NewIndexSorter(sli, tmpIdx)
				sort.Sort(sorter)
				indexer := sorter.GetIndex()
				proxTasks[i][x] = indexer[:MaxProx]
			}
		}
	}

	return proxTasks
}

func GetDefaultIdxSorted2(gPara *GPara) []int {
	// 按照离station远近排序
	var sortedTaskDist []float64
	for i := 0; i < gPara.NTask; i++ {
		tmpDist := gPara.Cost[0][0][0][gPara.TaskLoc[i]]
		sortedTaskDist = append(sortedTaskDist, tmpDist)
	}
	//sortedTaskDist2 := float64ToFloat64(sortedTaskDist)
	var tmpIdx = make([]int, len(sortedTaskDist))
	sli := sort.Float64Slice(sortedTaskDist)
	sorter := NewIndexSorter(sli, tmpIdx)
	sort.Sort(sort.Reverse(sorter))
	//sort.Sort(sorter)
	indexer := sorter.GetIndex()

	return indexer
}

func InitCapUseTask2(capUse *CapUse, gPara *GPara) {
	//初始化各个capacity含义
	//capUse.F[0] -- distance
	//capUse.F[1] -- parcel
	//capUse.F[2] -- weight
	//capUse.F[3] -- duration
	//capUse.F[4] -- FitCost
	//capUse.F[5] -- MapCost
	if gPara.CapRes != nil {
		capUse.F = make([]float64, len(gPara.CapRes[0])/2+2)
		for i := 0; i < len(gPara.CapRes[0])/2+2; i++ {
			capUse.F[i] = 0
		}
	} else {
		capUse.F = make([]float64, 0)
	}
}

func CheckResCap2(gPara *GPara, gState *GState) (bool, int) {
	var ok bool = false
	var resIdx int = 0
	if (gState.ResUse != nil) && (gPara.CapSeq != nil) {
		for i := 0; i < len(gState.ResUse); i++ {
			if gState.ResUse[i] < gPara.CapSeq[i] {
				resIdx = i
				ok = true
				break
			}
		}
	}
	return ok, resIdx
}

func InitResUse2(gPara *GPara, gState *GState) []int {
	if gPara.CapSeq != nil {
		gState.ResUse = make([]int, len(gPara.CapSeq), cap(gPara.CapSeq))
		for i := 0; i < len(gPara.CapSeq); i++ {
			gState.ResUse[i] = 0
		}
	} else {
		gState.ResUse = make([]int, 0)
	}
	return gState.ResUse
}

func UpdateResUse2(resUse []int, resIdx int) []int {
	resUse[resIdx] += 1
	return resUse
}

func ResetCapUse2(capUse *CapUse) {
	//重置capacity 数值
	for i := 0; i < len(capUse.F); i++ {
		capUse.F[i] = 0
	}
}

func UpdateCapUseReturn2(capUse *CapUse, resIdx int, preNodeIdx int, newNodeIdx int, graphs [][][][]float64) {
	//更新已用资源
	capUse.F[0] += graphs[resIdx][0][preNodeIdx][newNodeIdx] //update distance
	capUse.F[3] += graphs[resIdx][1][preNodeIdx][newNodeIdx] //update preIdx-newIdx edge duration
}

// CapUse.F[0] distance
// CapUse.F[1] parcel
// CapUse.F[2] weight
// CapUse.F[3] duration
// CapUse.F[4] FitCost
// CapUse.F[5] MapCost
func UpdateCapUse2(gPara *GPara, capUse *CapUse, resIdx int, preNodeIdx int, newNodeIdx int, newTaskIdx int) {
	//更新已用资源
	if preNodeIdx == 0 {
		capUse.F[0] += gPara.Cost[resIdx][0][preNodeIdx][newNodeIdx] + gPara.Cost[resIdx][0][newNodeIdx][0] //update distance
		capUse.F[3] += gPara.Cost[resIdx][1][preNodeIdx][newNodeIdx] + gPara.Cost[resIdx][1][newNodeIdx][0] //update preIdx-newIdx edge duration
	} else {
		capUse.F[0] += -gPara.Cost[resIdx][0][preNodeIdx][0] + gPara.Cost[resIdx][0][preNodeIdx][newNodeIdx] + gPara.Cost[resIdx][0][newNodeIdx][0] //update distance
		capUse.F[3] += -gPara.Cost[resIdx][1][preNodeIdx][0] + gPara.Cost[resIdx][1][preNodeIdx][newNodeIdx] + gPara.Cost[resIdx][1][newNodeIdx][0] //update preIdx-newIdx edge duration
	}
	//过程加回程
	capUse.F[1] += gPara.CapTask[newTaskIdx][0] // update parcel
	capUse.F[2] += gPara.CapTask[newTaskIdx][1] // update weight
	capUse.F[3] += gPara.CapTask[newTaskIdx][2] //update newIdx service duration
	capUse.F[4], capUse.F[5] = GetSeqMFCostByDist(gPara, resIdx, capUse.F[0])
}

func CheckMaxCap2(para *GPara, capUse *CapUse, resIdx int, preNodeIdx int, newNodeIdx int, newTaskIdx int) bool {
	//校验加入task后是否满足约束
	var statisfy bool
	var tmpDis, tmpParcel, tmpWeight, tmpDur float64 = 0.0, 0.0, 0.0, 0.0

	if capUse != nil {
		tmpDis = capUse.F[0] + para.Cost[resIdx][0][preNodeIdx][newNodeIdx] + para.Cost[resIdx][0][newNodeIdx][0] - para.Cost[resIdx][0][preNodeIdx][0] //校验加回程
		tmpParcel = capUse.F[1] + para.CapTask[newTaskIdx][0]
		tmpWeight = capUse.F[2] + para.CapTask[newTaskIdx][1]
		tmpDur = capUse.F[3] + para.Cost[resIdx][1][preNodeIdx][newNodeIdx] + para.CapTask[newTaskIdx][2] + para.Cost[resIdx][1][newNodeIdx][0] - para.Cost[resIdx][1][preNodeIdx][0] //校验加回程
	} else {
		// for checking infeasible tasks
		tmpDis = para.Cost[resIdx][0][preNodeIdx][newNodeIdx] + para.Cost[resIdx][0][newNodeIdx][0] //校验加回程
		tmpParcel = para.CapTask[newTaskIdx][0]
		tmpWeight = para.CapTask[newTaskIdx][1]
		tmpDur = para.Cost[resIdx][1][preNodeIdx][newNodeIdx] + para.CapTask[newTaskIdx][2] + para.Cost[resIdx][1][newNodeIdx][0] //校验加回程
	}
	tmpDisBool := (tmpDis >= 0) && (tmpDis <= para.CapRes[resIdx][1]) //防止不连通距离相加小于0
	tmpParcelBool := (tmpParcel >= 0) && (tmpParcel <= para.CapRes[resIdx][3])
	tmpWeightBool := (tmpWeight >= 0) && (tmpWeight <= para.CapRes[resIdx][5])
	tmpDurBool := (tmpDur >= 0) && (tmpDur <= para.CapRes[resIdx][7])

	if tmpDisBool && tmpParcelBool && tmpWeightBool && tmpDurBool {
		statisfy = true
	} else {
		statisfy = false
	}
	return statisfy
}

func CheckCap2(para *GPara, capUse *CapUse, resIdx int) bool {
	//校验加入task后是否满足约束
	var statisfy bool
	//tmpDis 加station到第一个点距离
	//capUse.F[0] -- distance
	//capUse.F[1] -- parcel
	//capUse.F[2] -- weight
	//capUse.F[3] -- duration

	tmpDisBool := capUse.F[0] >= para.CapRes[resIdx][0] && capUse.F[0] <= para.CapRes[resIdx][1] //取消校验mini
	tmpParcelBool := capUse.F[1] >= para.CapRes[resIdx][2] && capUse.F[1] <= para.CapRes[resIdx][3]
	tmpWeightBool := capUse.F[2] >= para.CapRes[resIdx][4] && capUse.F[2] <= para.CapRes[resIdx][5]
	tmpDurBool := capUse.F[3] >= para.CapRes[resIdx][6] && capUse.F[3] <= para.CapRes[resIdx][7]

	if tmpDisBool && tmpParcelBool && tmpWeightBool && tmpDurBool {
		statisfy = true
	} else {
		statisfy = false
	}
	return statisfy
}

func UpdateCent(nodes [][]float64, taskLoc []int, avgCent []float64, num int, task int) []float64 {
	var tmpAvgLat float64 = 0.0
	var tmpAvgLng float64 = 0.0
	tmpLat := nodes[taskLoc[task]][0]
	tmpLng := nodes[taskLoc[task]][1]
	tmpAvgLat = (avgCent[0]*float64(num) + tmpLat) / float64(num+1)
	tmpAvgLng = (avgCent[1]*float64(num) + tmpLng) / float64(num+1)

	cent := []float64{tmpAvgLat, tmpAvgLng}
	return cent
}

func GetCentDist(nodes [][]float64, centLatLng []float64, nodeId int) float64 {
	var taskCentDist float64
	taskCentDist = GreatCircleDistance(centLatLng, nodes[nodeId])
	return taskCentDist
}

func ResetAssigned(pendingTasks []int, assigned []bool) []bool {
	var newAssigned = assigned
	for i := 0; i < len(pendingTasks); i++ {
		newAssigned[pendingTasks[i]] = false
	}
	return newAssigned
}

func GetUnassignedTasks2(assigned []bool) []int {
	var unassignedTasks = make([]int, 0)
	for i := 0; i < len(assigned); i++ {
		if !assigned[i] {
			unassignedTasks = append(unassignedTasks, i)
		}
	}
	return unassignedTasks
}

func CopySliceF64(s []float64) []float64 {
	var newSlice []float64
	for i := 0; i < len(s); i++ {
		newSlice = append(newSlice, s[i])
	}
	return newSlice
}

func PostProcessMethod(state *GState, para *GPara) {
	state.Seqs = make([][][]int, 1)
	state.Asgmts = make([][]int, 1)
	state.InfeaTasks = make([][]int, 1)
	state.UnasgTasks = make([][]int, 1)
	state.Feats = make([][][]float64, 1)
	state.Seqs[0] = state.BestInnerSeqs
	var tmpAsg = make([]int, len(state.BestInnerAsgmts))
	for i := 0; i < len(state.BestInnerAsgmts); i++ {
		tmpAsg[i] = para.CapResMap[state.BestInnerAsgmts[i]]
	}
	state.Asgmts[0] = tmpAsg
	state.InfeaTasks[0] = state.BestInnerInfeaTasks
	state.UnasgTasks[0] = state.BestInnerUnasgTasks
	state.Feats[0] = state.InnerFeats
}
