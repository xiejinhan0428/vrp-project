package solver

import (
	"math"
	"sort"
)

func Method03(gPara *GPara, gState *GState, gFun *GFun) {
	searchUnasgTasks(gPara, gState)
	//尝试加车
	if len(gState.InnerUnasgTasks) > 0 {
		addNewResource(gPara, gState, gFun)
	}
}

func GetNewAssigned(gPara *GPara, gState *GState) (assignedTasks []bool) {
	assignedTasks = make([]bool, gPara.NTask)
	for i := 0; i < len(assignedTasks); i++ {
		assignedTasks[i] = true
	}
	for i := 0; i < len(gState.InnerUnasgTasks); i++ {
		assignedTasks[gState.InnerUnasgTasks[i]] = false
	}
	return
}

func addNewResource(gPara *GPara, gState *GState, gFun *GFun) {
	reNewResUse(gPara, gState)
	ok, _ := CheckResCap2(gPara, gState)
	if ok {
		//重新生成线路
		assignedTasks := GetNewAssigned(gPara, gState)
		getNewSeq(gPara, gState, gFun, assignedTasks, 0.3, 0.7)
	}
}

func GetDefaultUnasgSorted(gPara *GPara, sortTask []int) []int {
	// 按照离station远近排序
	var sortedTaskDist []float64
	for i := 0; i < len(sortTask); i++ {
		tmpDist := gPara.Cost[0][0][0][gPara.TaskLoc[sortTask[i]]]
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

func getNewSeq(gPara *GPara, gState *GState, gFun *GFun, assignedTasks []bool, centWeight float64, graphWeight float64) {
	cntAsg := 0
	var next, curr, cnt int = 0, 0, 0
	var capUse CapUse
	var pendingTasks []int
	// proxTask 已存在

	// index sort based on task idx list
	idxSorted := GetDefaultUnasgSorted(gPara, gState.InnerUnasgTasks)

	var sortedUnasgTasks = make([]int, len(gState.InnerUnasgTasks))
	for i := 0; i < len(idxSorted); i++ {
		sortedUnasgTasks[i] = gState.InnerUnasgTasks[idxSorted[i]]
	}

	gFun.InitCapUseTask(&capUse, gPara)

	for cntAsg < len(gState.InnerUnasgTasks) {
		ok, resIdx := gFun.CheckResCap(gPara, gState)
		if !ok {
			break
		}
		s := make([]int, 0)
		gFun.ResetCapUse(&capUse)

		// find first task
		for i := 0; i < len(sortedUnasgTasks); i++ {
			next = sortedUnasgTasks[i]
			if assignedTasks[next] == false {
				assignedTasks[next] = true
				s = append(s, next)
				gFun.UpdateCapUse(gPara, &capUse, resIdx, 0, gPara.TaskLoc[next], next)
				//dist := getSeqDist(s, resIdx)
				//if capUse.F[0] != dist {
				//	fmt.Println("距离不一致:", "dist:", dist, "  capUse-dist:", capUse.F[0])
				//}

				//runtime.ReadMemStats(&ms)
				//fmt.Println("UpdateCapUse2", int(float64(ms.Alloc)*1e-6), "MB")

				cntAsg += 1
				curr = next
				break
			}
		}

		if len(s) == 0 {
			if len(pendingTasks) == 0 {
				break
			}
		}

		// find the following tasks for a res
		var avgCent = []float64{gPara.Nodes[gPara.TaskLoc[curr]][0], gPara.Nodes[gPara.TaskLoc[curr]][1]}
		for i := 1; i <= len(sortedUnasgTasks); i++ {
			cnt = 0
			// mini feat
			var minSumCost float64 = math.MaxFloat64
			for _, j := range gPara.ProxTask[resIdx][curr] {
				if assignedTasks[j] == false && gFun.CheckMaxCap(gPara, &capUse, resIdx, gPara.TaskLoc[curr], gPara.TaskLoc[j], j) {
					//find minimize gPara.Cost(dist + centDis)
					newCentDist := GetCentDist(gPara.Nodes, avgCent, gPara.TaskLoc[j])
					//runtime.ReadMemStats(&ms)
					//fmt.Println("GetCentDist", int(float64(ms.Alloc)*1e-6), "MB")
					var tmpSumCost float64 = centWeight*newCentDist + graphWeight*gPara.Cost[resIdx][0][gPara.TaskLoc[curr]][gPara.TaskLoc[j]]
					if tmpSumCost < minSumCost {
						minSumCost = tmpSumCost
						next = j
						cnt += 1
					}
				}
			}
			// if found
			if cnt > 0 {
				assignedTasks[next] = true
				avgCent = UpdateCent(gPara.Nodes, gPara.TaskLoc, avgCent, len(s), next)
				s = append(s, next)
				gFun.UpdateCapUse(gPara, &capUse, resIdx, gPara.TaskLoc[curr], gPara.TaskLoc[next], next)

				//dist := getSeqDist(s, resIdx)
				//dur := getSeqDur(s, resIdx)
				//if capUse.F[0] != dist {
				//	fmt.Println("距离不一致:", "dist:", dist, "  capUse-dist:", capUse.F[0], "seq:", s)
				//}
				//if capUse.F[3] != dur {
				//	fmt.Println("时间不一致:", "dur:", dur, "  capUse-dur:", capUse.F[3], "seq:", s)
				//}

				//fmt.Println("before init opt:", getSeqDist(s, resIdx), "detail:", &capUse.F, "seq:", s)
				//search4GreedyForInit(gPara, s, &capUse, resIdx, avgCent)
				search3GreedyForInit(gPara, s, &capUse, resIdx)
				search1GreedyForInit(gPara, s, &capUse, resIdx)
				//fmt.Println("after init opt:", getSeqDist(s, resIdx), "detail:", &capUse.F, "seq:", s)
				next = s[len(s)-1]

				cntAsg += 1
				curr = next
			} else {
				//UpdateCapUseReturn2(&capUse, resIdx, gPara.TaskLoc[curr], 0, gPara.Cost)
				if gFun.CheckCap(gPara, &capUse, resIdx) {
					//dist := getSeqDist(s, resIdx)
					//dur := getSeqDur(s, resIdx)
					//if capUse.F[0] != dist {
					//	fmt.Println("距离不一致:", "dist:", dist, "  capUse-dist:", capUse.F[0], "seq:", s)
					//}
					//fmt.Println("dur:", dur, " capUseDur:", capUse.F[3])
					gState.InnerSeqs = append(gState.InnerSeqs, s)
					gState.InnerAsgmts = append(gState.InnerAsgmts, resIdx)
					capUse.F[4], capUse.F[5] = GetSeqMFCostByDist(gPara, resIdx, capUse.F[0])
					gState.InnerSeqDtls = append(gState.InnerSeqDtls, CopySliceF64(capUse.F))
					gState.ResUse = gFun.UpdateResUse(gState.ResUse, resIdx)
					gState.InnerFeats = append(gState.InnerFeats, CopySliceF64(avgCent))
					//更新gState.InnerAsgTasks
					gState.InnerUnasgTasks = UpdateUnasgTasks(gState, s)
				} else {
					pendingTasks = append(pendingTasks, s...)
				}
				break
			}
		}
	}
}

func UpdateUnasgTasks(gState *GState, s []int) (newUnasgTasks []int) {
	newUnasgTasks = make([]int, 0)
	for i := 0; i < len(gState.InnerUnasgTasks); i++ {
		mark := len(gState.InnerUnasgTasks)
		for j := 0; j < len(s); j++ {
			if gState.InnerUnasgTasks[i] == s[j] {
				mark = i
				break
			}
		}
		//不在s里
		if mark == len(gState.InnerUnasgTasks) {
			newUnasgTasks = append(newUnasgTasks, gState.InnerUnasgTasks[i])
		}
	}
	return
}

func reNewResUse(gPara *GPara, gState *GState) {
	var newResUse = make([]int, len(gPara.CapSeq))
	for i := 0; i < len(gState.InnerAsgmts); i++ {
		newResUse[gState.InnerAsgmts[i]] += 1
	}
	gState.ResUse = newResUse
}

func searchUnasgTasks(gPara *GPara, gState *GState) {
	for i := 0; i < len(gState.InnerUnasgTasks); i++ {
		newTask := gState.InnerUnasgTasks[i]
		for j := 0; j < len(gState.InnerSeqs); j++ {
			if insert(gPara, gState, newTask, j) {
				gState.InnerUnasgTasks = append(gState.InnerUnasgTasks[:i], gState.InnerUnasgTasks[i+1:]...)
				break
			}
		}
	}
}

//d[0] -- distance
//d[1] -- parcel
//d[2] -- weight
//d[3] -- duration

func insert(gPara *GPara, gState *GState, t int, seqIdx int) (ok bool) {
	//先判断parcel weight
	conts1 := (gState.InnerSeqDtls[seqIdx][1]+gPara.CapTask[t][0] >= gPara.CapRes[gState.InnerAsgmts[seqIdx]][2]) && (gState.InnerSeqDtls[seqIdx][1]+gPara.CapTask[t][0] <= gPara.CapRes[gState.InnerAsgmts[seqIdx]][3])
	conts2 := (gState.InnerSeqDtls[seqIdx][2]+gPara.CapTask[t][1] >= gPara.CapRes[gState.InnerAsgmts[seqIdx]][4]) && (gState.InnerSeqDtls[seqIdx][2]+gPara.CapTask[t][1] <= gPara.CapRes[gState.InnerAsgmts[seqIdx]][5])

	//再找插入点
	if conts1 && conts2 {
		tIdx, deltaD, deltaDur := getMiniInsert(gPara, gState, t, seqIdx)
		if tIdx > len(gState.InnerSeqs[seqIdx]) {
			return
		}
		conts0 := (gState.InnerSeqDtls[seqIdx][0]+deltaD >= gPara.CapRes[gState.InnerAsgmts[seqIdx]][0]) && (gState.InnerSeqDtls[seqIdx][0]+deltaD <= gPara.CapRes[gState.InnerAsgmts[seqIdx]][1])
		conts3 := (gState.InnerSeqDtls[seqIdx][3]+deltaDur >= gPara.CapRes[gState.InnerAsgmts[seqIdx]][6]) && (gState.InnerSeqDtls[seqIdx][3]+deltaDur <= gPara.CapRes[gState.InnerAsgmts[seqIdx]][7])
		if conts0 && conts3 {
			//update gState.InnerSeqs
			newSeq := make([]int, 0)
			newSeq = append(newSeq, gState.InnerSeqs[seqIdx][:tIdx]...)
			newSeq = append(newSeq, t)
			newSeq = append(newSeq, gState.InnerSeqs[seqIdx][tIdx:]...)
			gState.InnerSeqs[seqIdx] = newSeq
			// update gState.InnerSeqDtls
			gState.InnerSeqDtls[seqIdx][0] += deltaD
			gState.InnerSeqDtls[seqIdx][1] += gPara.CapTask[t][0]
			gState.InnerSeqDtls[seqIdx][2] += gPara.CapTask[t][1]
			gState.InnerSeqDtls[seqIdx][3] += gPara.CapTask[t][2] + deltaDur
			gState.InnerSeqDtls[seqIdx][4], gState.InnerSeqDtls[seqIdx][5] = GetSeqMFCostByDist(gPara, gState.InnerAsgmts[seqIdx], gState.InnerSeqDtls[seqIdx][0])
			//update gState.InnerFeats
			lat1 := (gState.InnerFeats[seqIdx][0]*float64(len(gState.InnerSeqs[seqIdx])-1) + gPara.Nodes[gPara.TaskLoc[t]][0]) / float64(len(gState.InnerSeqs[seqIdx]))
			lng1 := (gState.InnerFeats[seqIdx][1]*float64(len(gState.InnerSeqs[seqIdx])-1) + gPara.Nodes[gPara.TaskLoc[t]][1]) / float64(len(gState.InnerSeqs[seqIdx]))
			gState.InnerFeats[seqIdx][0] = lat1
			gState.InnerFeats[seqIdx][1] = lng1
			ok = true
			return
		}
	}
	return
}

func getMiniInsert(gPara *GPara, gState *GState, t int, seqIdx int) (tIdx int, minD, minDur float64) {
	var nt, n0, n1 int = 0, 0, 0
	tIdx = len(gState.InnerSeqs[seqIdx]) + 1 // Idx
	var minDelta float64 = math.MaxFloat64
	minD, minDur = math.MaxFloat64, math.MaxFloat64
	nt = gPara.TaskLoc[t]
	for i := 0; i < len(gState.InnerSeqs[seqIdx])+1; i++ {
		if i == 0 {
			n0 = 0
		} else {
			n0 = gPara.TaskLoc[gState.InnerSeqs[seqIdx][i-1]]
		}

		if i == len(gState.InnerSeqs[seqIdx]) {
			n1 = 0
		} else {
			n1 = gPara.TaskLoc[gState.InnerSeqs[seqIdx][i]]
		}
		deltaD := gPara.Cost[gState.InnerAsgmts[seqIdx]][0][n0][nt] + gPara.Cost[gState.InnerAsgmts[seqIdx]][0][nt][n1] - gPara.Cost[gState.InnerAsgmts[seqIdx]][0][n0][n1]
		deltaDur := gPara.Cost[gState.InnerAsgmts[seqIdx]][1][n0][nt] + gPara.Cost[gState.InnerAsgmts[seqIdx]][1][nt][n1] - gPara.Cost[gState.InnerAsgmts[seqIdx]][1][n0][n1] + gPara.CapTask[t][2]
		delta := GetInsertMinDelta(gPara, gState, seqIdx, deltaD)
		if delta < minDelta {
			tIdx = i
			minD = deltaD
			minDur = deltaDur
			minDelta = delta
		}
	}
	return
}

func GetInsertMinDelta(gPara *GPara, gState *GState, seqIdx int, deltaD float64) float64 {
	var delta float64 = 0.0
	if gPara.Obj == 0 {
		delta = deltaD
	} else {
		newFC, newMC := GetSeqMFCostByDist(gPara, gState.InnerAsgmts[seqIdx], gState.InnerSeqDtls[seqIdx][0]+deltaD)
		deltaMC := newMC - gState.InnerSeqDtls[seqIdx][5]

		if deltaMC == 0 {
			delta = newFC - gState.InnerSeqDtls[seqIdx][4]
		} else if deltaMC > 0 {
			delta = newFC - gState.InnerSeqDtls[seqIdx][5] + 10000
		} else {
			delta = deltaMC
		}

	}
	return delta
}
