package solver

import (
	"math"
	"sort"
)

type TmpSol struct {
	TmpSeqs   [][]int
	TmpAsgmt  []int
	TmpSeqDtl [][]float64
}

func Method04ForBest(gPara *GPara, gState *GState) bool {
	//换车
	var totalTrans = false
	reNewResUseForBest(gPara, gState)
	if len(gPara.CapSeq) > 1 {
		// 尝试换便宜的车
		totalTrans = TranResourceForBest(gPara, gState)
	}
	return totalTrans
}

func reNewResUseForBest(gPara *GPara, gState *GState) {
	var newResUse = make([]int, len(gPara.CapSeq))
	for i := 0; i < len(gState.BestInnerAsgmts); i++ {
		newResUse[gState.BestInnerAsgmts[i]] += 1
	}
	gState.ResUse = newResUse
}

func TranResourceForBest(gPara *GPara, gState *GState) bool {
	var totalTrans = false
	for sIdx := 0; sIdx < len(gState.BestInnerSeqs); sIdx++ {
		for rIdx := 0; rIdx < len(gPara.CapSeq); rIdx++ {
			if gState.BestInnerAsgmts[sIdx] == rIdx {
				continue
			}
			var ifTrans bool = false
			var newCost float64 = 0.0
			if gPara.CapSeq[rIdx]-gState.ResUse[rIdx] > 0 {
				// rIdx有可用车 尝试换对应区间的车
				ifTrans, newCost = IfTransForBest(rIdx, sIdx, gPara, gState)
			}

			if ifTrans {
				//如果 sIdx可以换成rIdx，换车动作
				TransResForSeqForBest(rIdx, sIdx, newCost, gPara, gState)
				totalTrans = true
			}
		}
	}
	return totalTrans
}

func TransResForSeqForBest(rIdx, sIdx int, newCost float64, gPara *GPara, gState *GState) {
	//update asg, dtl[4]、[5], ResUse
	gState.BestInnerAsgmts[sIdx] = rIdx
	gState.BestInnerSeqDtls[sIdx][4] = GetSeqFitCost(gPara, rIdx, gState.BestInnerSeqDtls[sIdx][0])
	gState.BestInnerSeqDtls[sIdx][5] = newCost
	reNewResUseForBest(gPara, gState)
}

func IfTransForBest(rIdx, sIdx int, gPara *GPara, gState *GState) (bool, float64) {
	var ifTrans bool = false
	var tmpCost float64 = 999999999999
	for t := 0; t < len(gPara.CapResCost[rIdx]); t++ {
		if gState.BestInnerSeqDtls[sIdx][0] >= gPara.CapResCost[rIdx][t][0] && gState.BestInnerSeqDtls[sIdx][0] < gPara.CapResCost[rIdx][t][1] {
			tmpCost = gPara.CapResCost[rIdx][t][4]
			break
		}
	}
	if gState.BestInnerSeqDtls[sIdx][0] >= gPara.CapResCost[rIdx][len(gPara.CapResCost[rIdx])-1][1] {
		tmpCost = gPara.CapResCost[rIdx][len(gPara.CapResCost[rIdx])-1][4]
	}

	if gState.BestInnerSeqDtls[sIdx][5] > tmpCost {
		// rIdx 更便宜，校验约束
		ifCont := CheckTransForBest(sIdx, rIdx, gPara, gState)
		if ifCont {
			ifTrans = true
		}
	}
	return ifTrans, tmpCost
}

func CheckTransForBest(sIdx, rIdx int, para *GPara, state *GState) (ok bool) {
	ok = true
	cont0 := state.BestInnerSeqDtls[sIdx][0] >= para.CapRes[rIdx][0]
	cont1 := state.BestInnerSeqDtls[sIdx][0] <= para.CapRes[rIdx][1]
	cont2 := state.BestInnerSeqDtls[sIdx][1] >= para.CapRes[rIdx][2]
	cont3 := state.BestInnerSeqDtls[sIdx][1] <= para.CapRes[rIdx][3]
	cont4 := state.BestInnerSeqDtls[sIdx][2] >= para.CapRes[rIdx][4]
	cont5 := state.BestInnerSeqDtls[sIdx][2] <= para.CapRes[rIdx][5]
	cont6 := state.BestInnerSeqDtls[sIdx][3] >= para.CapRes[rIdx][6]
	cont7 := state.BestInnerSeqDtls[sIdx][3] <= para.CapRes[rIdx][7]

	ok = cont0 && cont1 && cont2 && cont3 && cont4 && cont5 && cont6 && cont7

	return
}

//////for normal below/////////

func Method04(gPara *GPara, gState *GState) bool {
	//换车
	var totalTrans = false
	reNewResUse(gPara, gState)
	if len(gPara.CapSeq) > 1 {
		// 尝试换便宜的车
		totalTrans = TranResource(gPara, gState)
	}
	return totalTrans
}

func TranResource(gPara *GPara, gState *GState) bool {
	var totalTrans = false
	for sIdx := 0; sIdx < len(gState.InnerSeqs); sIdx++ {
		for rIdx := 0; rIdx < len(gPara.CapSeq); rIdx++ {
			if gState.InnerAsgmts[sIdx] == rIdx {
				continue
			}
			var ifTrans bool = false
			var newCost float64 = 0.0
			if gPara.CapSeq[rIdx]-gState.ResUse[rIdx] > 0 {
				// rIdx有可用车 尝试换对应区间的车
				ifTrans, newCost = IfTrans(rIdx, sIdx, gPara, gState)
			}

			if ifTrans {
				//如果 sIdx可以换成rIdx，换车动作
				TransResForSeq(rIdx, sIdx, newCost, gPara, gState)
				totalTrans = true
			}
		}
	}
	return totalTrans
}

func TransResForSeq(rIdx, sIdx int, newCost float64, gPara *GPara, gState *GState) {
	//update asg, dtl[4]、[5], ResUse
	gState.InnerAsgmts[sIdx] = rIdx
	gState.InnerSeqDtls[sIdx][4] = GetSeqFitCost(gPara, rIdx, gState.InnerSeqDtls[sIdx][0])
	gState.InnerSeqDtls[sIdx][5] = newCost
	reNewResUse(gPara, gState)
}

func IfTrans(rIdx, sIdx int, gPara *GPara, gState *GState) (bool, float64) {
	var ifTrans bool = false
	var tmpCost float64 = math.MaxFloat64
	for t := 0; t < len(gPara.CapResCost[rIdx]); t++ {
		if gState.InnerSeqDtls[sIdx][0] >= gPara.CapResCost[rIdx][t][0] && gState.InnerSeqDtls[sIdx][0] < gPara.CapResCost[rIdx][t][1] {
			tmpCost = gPara.CapResCost[rIdx][t][4]
			break
		}
	}
	if gState.InnerSeqDtls[sIdx][0] >= gPara.CapResCost[rIdx][len(gPara.CapResCost[rIdx])-1][1] {
		tmpCost = gPara.CapResCost[rIdx][len(gPara.CapResCost[rIdx])-1][4]
	}

	if gState.InnerSeqDtls[sIdx][5] > tmpCost {
		// rIdx 更便宜，校验约束
		ifCont := CheckTranVehSeqCont(sIdx, rIdx, gPara, gState)
		if ifCont {
			ifTrans = true
		}
	}
	return ifTrans, tmpCost
}

func CheckTranVehSeqCont(sIdx, rIdx int, para *GPara, state *GState) (ok bool) {
	ok = true
	cont0 := state.InnerSeqDtls[sIdx][0] >= para.CapRes[rIdx][0]
	cont1 := state.InnerSeqDtls[sIdx][0] <= para.CapRes[rIdx][1]
	cont2 := state.InnerSeqDtls[sIdx][1] >= para.CapRes[rIdx][2]
	cont3 := state.InnerSeqDtls[sIdx][1] <= para.CapRes[rIdx][3]
	cont4 := state.InnerSeqDtls[sIdx][2] >= para.CapRes[rIdx][4]
	cont5 := state.InnerSeqDtls[sIdx][2] <= para.CapRes[rIdx][5]
	cont6 := state.InnerSeqDtls[sIdx][3] >= para.CapRes[rIdx][6]
	cont7 := state.InnerSeqDtls[sIdx][3] <= para.CapRes[rIdx][7]

	ok = cont0 && cont1 && cont2 && cont3 && cont4 && cont5 && cont6 && cont7

	return
}

func Method05(gPara *GPara, gState *GState) {
	//减车
	//for ss := 0; ss < len(gState.InnerSeqDtls); ss++ {
	//	if gState.InnerSeqDtls[ss][5] == 0 {
	//		fmt.Println("stop")
	//	}
	//}
	ifSearch, sIdx := SearchSeq(gPara, gState)
	if ifSearch {
		ResourceOpt(sIdx, gPara, gState)
	}
}

func ResourceOpt(sIdx int, gPara *GPara, gState *GState) bool {
	var ifOpt bool = false
	initTmpSol := CopySolFromGState(gState)
	initDist, initFC, initMC := GetTmpSolObjInfo(initTmpSol)
	initResNum := len(initTmpSol.TmpAsgmt)

	finalTmpSol := searchTmpTasks(sIdx, initTmpSol, gPara)
	finalDist, finalFC, finalMC := GetTmpSolObjInfo(finalTmpSol)
	finalResNum := len(finalTmpSol.TmpAsgmt)

	delta := GetTotalDelta(gPara, initDist, finalDist, initFC, finalFC, initMC, finalMC)
	if finalResNum-initResNum < 0 {
		CopyGStateFromSol(gState, finalTmpSol)
		ifOpt = true
	} else if finalResNum-initResNum == 0 && delta < 0 {
		CopyGStateFromSol(gState, finalTmpSol)
		ifOpt = true
	}
	return ifOpt

}

func GetTmpSolObjInfo(sol TmpSol) (float64, float64, float64) {
	var totalDist, totalFC, totalMC float64 = 0.0, 0.0, 0.0
	for i := 0; i < len(sol.TmpSeqDtl); i++ {
		totalDist += sol.TmpSeqDtl[i][0]
		totalFC += sol.TmpSeqDtl[i][4]
		totalMC += sol.TmpSeqDtl[i][5]
	}
	return totalDist, totalFC, totalMC
}

func GetSortedFromStat(gPara *GPara, seq []int) []int {
	// 按照离station远近排序
	var sortedTaskDist []float64
	for i := 0; i < len(seq); i++ {
		tmpDist := gPara.Cost[0][0][0][seq[i]]
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

func searchTmpTasks(sIdx int, tmpSol TmpSol, gPara *GPara) TmpSol {
	var newTotalSol = CopyTmpSol(tmpSol)
	var oldTaskIdList = make([]int, len(tmpSol.TmpSeqs[sIdx]))
	oldTaskIDIdxList := GetSortedFromStat(gPara, tmpSol.TmpSeqs[sIdx])
	for s := 0; s < len(oldTaskIDIdxList); s++ {
		oldTaskIdList[s] = newTotalSol.TmpSeqs[sIdx][oldTaskIDIdxList[s]]
	}

	for i := 0; i < len(oldTaskIdList); i++ {
		newTask := oldTaskIdList[i]
		var newTaskIDList = make([]int, len(newTotalSol.TmpSeqs[sIdx]))
		newTaskIDIdxList := GetSortedFromStat(gPara, newTotalSol.TmpSeqs[sIdx])
		for s := 0; s < len(newTaskIDIdxList); s++ {
			newTaskIDList[s] = newTotalSol.TmpSeqs[sIdx][newTaskIDIdxList[s]]
		}

		oriTaskIdx := -1
		for tt := 0; tt < len(newTaskIDList); tt++ {
			if newTaskIDList[tt] == newTask {
				oriTaskIdx = newTaskIDIdxList[tt]
				break
			}
		}
		if oriTaskIdx == -1 {
			continue
		}

		var ifInsert bool = false
		var bestSeqIdx, bestPosIdx, bestSeqPosIdx int = -1, -1, -1
		var minSeqDelta, minDelta float64 = math.MaxFloat64, math.MaxFloat64
		for j := 0; j < len(newTotalSol.TmpSeqs); j++ {
			if j == sIdx {
				continue
			}
			ifInsert, bestSeqPosIdx, minSeqDelta = TryInsert(newTask, j, newTotalSol, gPara)
			if ifInsert {
				// 如果能insert 记录 obj sIdx, 位置 最后找最小的
				if minSeqDelta < minDelta {
					minDelta = minSeqDelta
					bestSeqIdx = j
					bestPosIdx = bestSeqPosIdx
				}
			}
		}
		//所有线路的遍历完，执行最佳位置变换动作
		if ifInsert {
			newSol := DoInsert(sIdx, newTask, oriTaskIdx, bestSeqIdx, bestPosIdx, newTotalSol, gPara)
			newTotalSol = CopyTmpSol(newSol)
		}
	}
	return newTotalSol
}

func DoInsert(sIdx, newTask, oriTaskIdx, bestSeqIdx, bestPosIdx int, tmpSol TmpSol, gPara *GPara) TmpSol {
	var newSol TmpSol
	var newSeqs = make([][]int, 0)
	var newAsgmt = make([]int, 0)
	var newSeqsDtl = make([][]float64, 0)

	//修改bestSeqIdx 序列
	newInsertSeq := make([]int, 0)
	if bestPosIdx == 0 {
		newInsertSeq = append(newInsertSeq, newTask)
		newInsertSeq = append(newInsertSeq, tmpSol.TmpSeqs[bestSeqIdx]...)
	} else if bestPosIdx == len(tmpSol.TmpSeqs[bestSeqIdx]) {
		newInsertSeq = append(newInsertSeq, tmpSol.TmpSeqs[bestSeqIdx]...)
		newInsertSeq = append(newInsertSeq, newTask)
	} else {
		newInsertSeq = append(newInsertSeq, tmpSol.TmpSeqs[bestSeqIdx][:bestPosIdx]...)
		newInsertSeq = append(newInsertSeq, newTask)
		newInsertSeq = append(newInsertSeq, tmpSol.TmpSeqs[bestSeqIdx][bestPosIdx:]...)
	}

	//修改sIdx 序列 (有可能len() == 0)
	newDestroySeq := make([]int, 0)
	if len(tmpSol.TmpSeqs[sIdx])-1 != 0 {
		newDestroySeq = append(newDestroySeq, tmpSol.TmpSeqs[sIdx][:oriTaskIdx]...)
		if oriTaskIdx != len(tmpSol.TmpSeqs[sIdx])-1 {
			newDestroySeq = append(newDestroySeq, tmpSol.TmpSeqs[sIdx][oriTaskIdx+1:]...)
		}
	}

	//修改 asgmts
	for i := 0; i < len(tmpSol.TmpAsgmt); i++ {
		if i == sIdx && len(newDestroySeq) == 0 {
			continue
		}
		newAsgmt = append(newAsgmt, tmpSol.TmpAsgmt[i])
	}

	//修改 seqs
	for i := 0; i < len(tmpSol.TmpSeqs); i++ {
		tmpNewSeq := make([]int, 0)
		if i == sIdx && len(newDestroySeq) == 0 {
			continue
		} else if i == sIdx && len(newDestroySeq) > 0 {
			tmpNewSeq = append(tmpNewSeq, newDestroySeq...)
		} else if i == bestSeqIdx {
			tmpNewSeq = append(tmpNewSeq, newInsertSeq...)
		} else {
			tmpNewSeq = append(tmpNewSeq, tmpSol.TmpSeqs[i]...)
		}
		newSeqs = append(newSeqs, tmpNewSeq)
	}

	// 修改 dtl
	for i := 0; i < len(tmpSol.TmpSeqDtl); i++ {
		if i == sIdx && len(newDestroySeq) == 0 {
			continue
		} else if i == sIdx && len(newDestroySeq) > 0 {
			newDestroyDtl := GetSeqDtl(newDestroySeq, i, tmpSol.TmpAsgmt, gPara)
			newSeqsDtl = append(newSeqsDtl, newDestroyDtl)
		} else if i == bestSeqIdx {
			newInsertDtl := GetSeqDtl(newInsertSeq, i, tmpSol.TmpAsgmt, gPara)
			newSeqsDtl = append(newSeqsDtl, newInsertDtl)
		} else {
			newSeqDtl := make([]float64, 0)
			newSeqDtl = append(newSeqDtl, tmpSol.TmpSeqDtl[i]...)
			newSeqsDtl = append(newSeqsDtl, newSeqDtl)
		}
	}

	newSol = TmpSol{
		TmpSeqs:   newSeqs,
		TmpAsgmt:  newAsgmt,
		TmpSeqDtl: newSeqsDtl,
	}

	return newSol
}

func TryInsert(newTask, trysIdx int, tmpSol TmpSol, gPara *GPara) (bool, int, float64) {
	var ifInsert bool = false
	var bestPosIdx, tIdx int = -1, -1
	var bestDelta, deltaD, deltaDur, delta float64 = math.MaxFloat64, math.MaxFloat64, math.MaxFloat64, math.MaxFloat64
	//先判断parcel weight
	conts1 := (tmpSol.TmpSeqDtl[trysIdx][1]+gPara.CapTask[newTask][0] >= gPara.CapRes[tmpSol.TmpAsgmt[trysIdx]][2]) && (tmpSol.TmpSeqDtl[trysIdx][1]+gPara.CapTask[newTask][0] <= gPara.CapRes[tmpSol.TmpAsgmt[trysIdx]][3])
	conts2 := (tmpSol.TmpSeqDtl[trysIdx][2]+gPara.CapTask[newTask][1] >= gPara.CapRes[tmpSol.TmpAsgmt[trysIdx]][4]) && (tmpSol.TmpSeqDtl[trysIdx][2]+gPara.CapTask[newTask][1] <= gPara.CapRes[tmpSol.TmpAsgmt[trysIdx]][5])
	//再找插入点
	if conts1 && conts2 {
		//for ss := 0; ss < len(tmpSol.TmpSeqDtl); ss++ {
		//	if tmpSol.TmpSeqDtl[ss][5] == 0 {
		//		fmt.Println("stop")
		//	}
		//}
		tIdx, deltaD, deltaDur, delta = getMiniInsertPos(gPara, tmpSol, newTask, trysIdx)
		//tIdx为插入obj增加最小的位置，然后校验距离 时间约束
		var conts0, conts3 = false, false
		if tIdx != -1 {
			conts0 = (tmpSol.TmpSeqDtl[trysIdx][0]+deltaD >= gPara.CapRes[tmpSol.TmpAsgmt[trysIdx]][0]) && (tmpSol.TmpSeqDtl[trysIdx][0]+deltaD <= gPara.CapRes[tmpSol.TmpAsgmt[trysIdx]][1])
			conts3 = (tmpSol.TmpSeqDtl[trysIdx][3]+deltaDur >= gPara.CapRes[tmpSol.TmpAsgmt[trysIdx]][6]) && (tmpSol.TmpSeqDtl[trysIdx][3]+deltaDur <= gPara.CapRes[tmpSol.TmpAsgmt[trysIdx]][7])
		}
		if conts0 && conts3 {
			//满足约束可以插入
			ifInsert = true
			bestPosIdx = tIdx
			bestDelta = delta
		}
	}
	return ifInsert, bestPosIdx, bestDelta
}

func getMiniInsertPos(gPara *GPara, tmpSol TmpSol, newTask int, isIdx int) (tIdx int, minD, minDur, minDelta float64) {
	var nt, n0, n1 int = 0, 0, 0
	tIdx = -1 // Idx
	minDelta, minD, minDur = math.MaxFloat64, math.MaxFloat64, math.MaxFloat64
	nt = gPara.TaskLoc[newTask]
	for i := 0; i < len(tmpSol.TmpSeqs[isIdx])+1; i++ {
		if i == 0 {
			n0 = 0 //station
		} else {
			n0 = gPara.TaskLoc[tmpSol.TmpSeqs[isIdx][i-1]]
		}

		if i == len(tmpSol.TmpSeqs[isIdx]) {
			n1 = 0
		} else {
			n1 = gPara.TaskLoc[tmpSol.TmpSeqs[isIdx][i]]
		}
		deltaD := gPara.Cost[tmpSol.TmpAsgmt[isIdx]][0][n0][nt] + gPara.Cost[tmpSol.TmpAsgmt[isIdx]][0][nt][n1] - gPara.Cost[tmpSol.TmpAsgmt[isIdx]][0][n0][n1]
		deltaDur := gPara.Cost[tmpSol.TmpAsgmt[isIdx]][1][n0][nt] + gPara.Cost[tmpSol.TmpAsgmt[isIdx]][1][nt][n1] - gPara.Cost[tmpSol.TmpAsgmt[isIdx]][1][n0][n1] + gPara.CapTask[newTask][2]
		//for ss := 0; ss < len(tmpSol.TmpSeqDtl); ss++ {
		//	if tmpSol.TmpSeqDtl[ss][5] == 0 {
		//		fmt.Println("stop")
		//	}
		//}
		delta := GetInsertMinDeltaForPos(gPara, tmpSol, isIdx, deltaD)
		if delta < minDelta {
			tIdx = i
			minD = deltaD
			minDur = deltaDur
			minDelta = delta
		}
	}
	return
}

func GetInsertMinDeltaForPos(gPara *GPara, tmpSol TmpSol, seqIdx int, deltaD float64) float64 {
	var delta float64 = math.MaxFloat64
	if gPara.Obj == 0 {
		delta = deltaD
	} else {
		newFC, newMC := GetSeqMFCostByDist(gPara, tmpSol.TmpAsgmt[seqIdx], tmpSol.TmpSeqDtl[seqIdx][0]+deltaD)
		deltaMC := newMC - tmpSol.TmpSeqDtl[seqIdx][5]

		if deltaMC == 0 {
			delta = newFC - tmpSol.TmpSeqDtl[seqIdx][4]
		} else if deltaMC > 0 {
			delta = newFC - tmpSol.TmpSeqDtl[seqIdx][5] + 10000
		} else {
			delta = deltaMC
		}
	}
	return delta
}

func CopySolFromGState(gState *GState) TmpSol {
	//Copy innerseqs/ innerasg / innerSeqDtl
	var tmpSeqs = make([][]int, len(gState.InnerSeqs))
	var tmpAsgmts = make([]int, len(gState.InnerAsgmts))
	var tmpSeqDtl = make([][]float64, len(gState.InnerSeqDtls))
	for i := 0; i < len(gState.InnerSeqs); i++ {
		tmpSeqs[i] = make([]int, len(gState.InnerSeqs[i]))
		tmpAsgmts[i] = gState.InnerAsgmts[i]
		tmpSeqDtl[i] = make([]float64, len(gState.InnerSeqDtls[i]))
		for j := 0; j < len(gState.InnerSeqs[i]); j++ {
			tmpSeqs[i][j] = gState.InnerSeqs[i][j]
		}
		for j := 0; j < len(gState.InnerSeqDtls[i]); j++ {
			tmpSeqDtl[i][j] = gState.InnerSeqDtls[i][j]
		}
	}
	tmpSol := TmpSol{
		TmpSeqs:   tmpSeqs,
		TmpAsgmt:  tmpAsgmts,
		TmpSeqDtl: tmpSeqDtl,
	}
	return tmpSol
}

func SearchSeq(gPara *GPara, gState *GState) (bool, int) {
	if len(gState.InnerSeqs) < 1 {
		return false, 0
	}

	var minSeqIdx int = 0
	var ifSearch bool = false
	var avgLen = (gPara.NTask - len(gState.InnerUnasgTasks) - len(gState.InnerInfeaTasks)) / len(gState.InnerSeqs)
	for s := 0; s < len(gState.InnerSeqs); s++ {
		if len(gState.InnerSeqs[minSeqIdx]) > len(gState.InnerSeqs[s]) {
			minSeqIdx = s
		}
	}
	if len(gState.InnerSeqs[minSeqIdx]) < avgLen {
		ifSearch = true
	}
	return ifSearch, minSeqIdx
}

func CopyTmpSol(sol TmpSol) TmpSol {
	var newSol TmpSol
	var newTmpSeqs = make([][]int, len(sol.TmpSeqs))
	var newTmpAsgmt = make([]int, len(sol.TmpAsgmt))
	var newTmpSeqsDtl = make([][]float64, len(sol.TmpSeqDtl))

	for i := 0; i < len(sol.TmpSeqs); i++ {
		var newSeq = make([]int, len(sol.TmpSeqs[i]))
		for j := 0; j < len(sol.TmpSeqs[i]); j++ {
			newSeq[j] = sol.TmpSeqs[i][j]
		}
		newTmpSeqs[i] = newSeq

		newTmpAsgmt[i] = sol.TmpAsgmt[i]
		var newSeqDtl = make([]float64, len(sol.TmpSeqDtl[i]))
		for j := 0; j < len(sol.TmpSeqDtl[i]); j++ {
			newSeqDtl[j] = sol.TmpSeqDtl[i][j]
		}
		newTmpSeqsDtl[i] = newSeqDtl
	}
	newSol = TmpSol{
		TmpSeqs:   newTmpSeqs,
		TmpAsgmt:  newTmpAsgmt,
		TmpSeqDtl: newTmpSeqsDtl,
	}
	return newSol
}

func CopyGStateFromSol(gState *GState, sol TmpSol) {
	gState.InnerSeqs = sol.TmpSeqs
	gState.InnerAsgmts = sol.TmpAsgmt
	gState.InnerSeqDtls = sol.TmpSeqDtl
}
