package solver

import (
	"math"
	"runtime"
)

var ms runtime.MemStats

func searchInfeaTasks(gState *GState, gPara *GPara, assigned []bool) []bool {
	for i := 0; i < gPara.NTask; i++ {
		ifInfea := true
		for r := 0; r < len(gPara.CapRes); r++ {
			ok := gPara.CapSeq[r] > 0 && CheckMaxCap2(gPara, nil, r, 0, gPara.TaskLoc[i], i)
			if ok {
				ifInfea = false
				break
			}
		}
		if ifInfea {
			assigned[i] = true
			gState.InnerInfeaTasks = append(gState.InnerInfeaTasks, i)
		}
	}
	return assigned
}

func searchOrigin(gState *GState, gFun *GFun, gPara *GPara, centWeight float64, graphWeight float64) ([]int, []bool, *CapUse, []int) {
	cntTask := gPara.NTask
	assigned := make([]bool, cntTask)
	cntAsg := 0

	var capUse CapUse
	var curr, next, cnt int
	gState.InnerAsgmts = make([]int, 0)

	gState.InnerInfeaTasks = make([]int, 0)
	gState.InnerUnasgTasks = make([]int, 0)
	pendingTasks := make([]int, 0)

	// task proximity
	gPara.ProxTask = gFun.GetProxTask(gPara)
	// index sort based on task idx list
	idxSorted := gFun.GetDefaultIdxSorted(gPara)

	gFun.InitCapUseTask(&capUse, gPara)
	gState.ResUse = gFun.InitResUse(gPara, gState)

	gState.InnerSeqs = make([][]int, 0)
	gState.InnerSeqDtls = make([][]float64, 0)
	gState.InnerFeats = make([][]float64, 0, cap(gState.InnerSeqs))

	assigned = searchInfeaTasks(gState, gPara, assigned)

	for cntAsg < cntTask {
		ok, resIdx := gFun.CheckResCap(gPara, gState)
		if !ok {
			break
		}
		s := make([]int, 0)
		gFun.ResetCapUse(&capUse)

		// find first task
		for i := 0; i < cntTask; i++ {
			//对于每一辆车，从离站点最远的开始，找到第一个合适加入route
			next = idxSorted[i]
			if assigned[next] == false {
				if !gFun.CheckMaxCap(gPara, &capUse, resIdx, 0, gPara.TaskLoc[next], next) {
					if resIdx == 0 {
						gState.InnerInfeaTasks = append(gState.InnerInfeaTasks, next)
					} else {
						gState.InnerUnasgTasks = append(gState.InnerUnasgTasks, next)
					}
					assigned[next] = true
					cntAsg += 1
					continue
				}
				assigned[next] = true
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
		for i := 1; i <= cntTask; i++ {
			cnt = 0
			// mini feat
			var minSumCost float64 = math.MaxFloat64
			//ProxTask是排完序后的近邻点序列，3000个
			//对每条route加新点，加的是最新点的(遍历寻找)最近的合法点
			for _, j := range gPara.ProxTask[resIdx][curr] {
				if assigned[j] == false && gFun.CheckMaxCap(gPara, &capUse, resIdx, gPara.TaskLoc[curr], gPara.TaskLoc[j], j) {
					//find minimize cost(dist + centDis)
					newCentDist := GetCentDist(gPara.Nodes, avgCent, gPara.TaskLoc[j])
					//runtime.ReadMemStats(&ms)
					//fmt.Println("GetCentDist", int(float64(ms.Alloc)*1e-6), "MB")
					tmpSumCost := centWeight*newCentDist + graphWeight*gPara.Cost[resIdx][0][gPara.TaskLoc[curr]][gPara.TaskLoc[j]]
					if tmpSumCost < minSumCost {
						minSumCost = tmpSumCost
						next = j
						cnt += 1
					}
				}
			}
			// if found 加入本次route
			if cnt > 0 {
				assigned[next] = true
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
				//capUse.F[4], capUse.F[5] = GetSeqMFCostByDist(gPara, resIdx, capUse.F[0])
				//fmt.Println("after init opt:", getSeqDist(s, resIdx), "detail:", &capUse.F, "seq:", s)
				next = s[len(s)-1]

				cntAsg += 1
				curr = next
			} else { //not found ，该route结束。
				//UpdateCapUseReturn2(&capUse, resIdx, taskLoc[curr], 0, cost)
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
				} else {
					pendingTasks = append(pendingTasks, s...)
				}
				break
			}
		}
	}

	//var totalSeqTask = 0
	//for i := 0; i < len(innerSeqs); i++ {
	//	totalSeqTask += len(innerSeqs[i])
	//}
	//fmt.Println("totalSeqTask:", totalSeqTask, "  nTask", nTask)
	for i := 0; i < len(gState.InnerUnasgTasks); i++ {
		pendingTasks = append(pendingTasks, gState.InnerUnasgTasks[i])
		assigned[gState.InnerUnasgTasks[i]] = false
	}
	return pendingTasks, assigned, &capUse, idxSorted
}

//对于orphan点，一个个尝试插入route各个位置
func searchPending(gFun *GFun, gPara *GPara, gState *GState, pendingTasks []int, assigned []bool, capUse *CapUse, idxSorted []int) []int {
	var next, curr, cnt int
	cntTask := gPara.NTask
	var newPendingTasks = make([]int, 0)

	if len(pendingTasks) > 0 {
		cntAsg := 0
		//assigned = ResetAssigned(pendingTasks, assigned)

		for cntAsg < len(gState.InnerUnasgTasks) {
			ok, resIdx := gFun.CheckResCap(gPara, gState)
			if !ok {
				break
			}
			s := make([]int, 0)

			gFun.ResetCapUse(capUse)

			// find first task
			for i := 0; i < cntTask; i++ {
				next = idxSorted[i]
				if assigned[next] == false {
					assigned[next] = true
					s = append(s, next)
					gFun.UpdateCapUse(gPara, capUse, resIdx, 0, gPara.TaskLoc[next], next)
					cntAsg += 1
					curr = next
					break
				}
			}

			// find the following tasks for a res
			var avgCent = []float64{gPara.Nodes[gPara.TaskLoc[curr]][0], gPara.Nodes[gPara.TaskLoc[curr]][1]}
			for i := 1; i <= cntTask; i++ {
				cnt = 0
				var minSumCost float64 = math.MaxFloat64
				for _, j := range gPara.ProxTask[resIdx][curr] {
					if assigned[j] == false && gFun.CheckMaxCap(gPara, capUse, resIdx, gPara.TaskLoc[curr], gPara.TaskLoc[j], j) {
						//find minimize cost(dist + centDis)
						newCentDist := GetCentDist(gPara.Nodes, avgCent, gPara.TaskLoc[j])
						tmpSumCost := newCentDist + gPara.Cost[resIdx][0][gPara.TaskLoc[curr]][gPara.TaskLoc[j]]
						if tmpSumCost < minSumCost {
							minSumCost = tmpSumCost
							next = j
							cnt += 1
						}
					}
				}
				// if found
				if cnt > 0 {
					assigned[next] = true
					avgCent = UpdateCent(gPara.Nodes, gPara.TaskLoc, avgCent, len(s), next)
					s = append(s, next)
					gFun.UpdateCapUse(gPara, capUse, resIdx, gPara.TaskLoc[curr], gPara.TaskLoc[next], next)
					cntAsg += 1
					curr = next
				} else {
					//UpdateCapUseReturn2(capUse, resIdx, taskLoc[curr], 0, cost)
					if gFun.CheckCap(gPara, capUse, resIdx) {
						gState.InnerSeqs = append(gState.InnerSeqs, s)
						capUse.F[4], capUse.F[5] = GetSeqMFCostByDist(gPara, resIdx, capUse.F[0])
						gState.InnerSeqDtls = append(gState.InnerSeqDtls, CopySliceF64(capUse.F))
						gState.InnerAsgmts = append(gState.InnerAsgmts, resIdx)
						gState.ResUse = gFun.UpdateResUse(gState.ResUse, resIdx)
						gState.InnerFeats = append(gState.InnerFeats, CopySliceF64(avgCent))
					} else {
						newPendingTasks = append(newPendingTasks, s...)
					}
					break
				}
			}
		}
	}
	return newPendingTasks
}

func InitTask2(gState *GState, gFun *GFun, gPara *GPara, centWight float64, graphWeight float64) {
	//startTime := time.Now().UnixNano()
	pendingTasks, assigned, capUse, idxSorted := searchOrigin(gState, gFun, gPara, centWight, graphWeight)
	assigned = ResetAssigned(pendingTasks, assigned)
	gState.InnerUnasgTasks = GetUnassignedTasks2(assigned)

	newPd := searchPending(gFun, gPara, gState, pendingTasks, assigned, capUse, idxSorted)
	assigned = ResetAssigned(newPd, assigned)
	gState.InnerUnasgTasks = GetUnassignedTasks2(assigned)

	//endTime := time.Now().UnixNano()
	//iterCost := (endTime - startTime) / 1e6
	//fmt.Println("Time cost:", iterCost)
}

func InitSols(tEndUnix int64, gState *GState, gFun *GFun, gPara *GPara) {
	//var centWeight, graphWeight float64
	InitTask2(gState, gFun, gPara, 1.0, 0.0)
	gState.BestInnerSeqs, gState.BestInnerAsgmts, gState.BestInnerUnasgTasks, gState.BestInnerInfeaTasks = CopyResult(gState)
	bestInnerFeats := GenerateFeats(gPara, gState.BestInnerSeqs)
	gState.BestInnerSeqDtls = GenerateSeqDtls(gState.BestInnerSeqs, gState.BestInnerAsgmts, gPara)
	//for centWeight = 0.9; centWeight >= 0; centWeight -= 0.1 {
	//	if time.Now().Unix() > tEndUnix {
	//		break
	//	}
	//	graphWeight = 1.0 - centWeight
	//	InitTask2(gState, gFun, gPara, centWeight, graphWeight)
	//	RenewInit(&gState.BestInnerSeqs, &gState.BestInnerAsgmts, &gState.BestInnerUnasgTasks, &gState.InnerInfeaTasks, gState, gPara)
	//}
	gState.InnerSeqs, gState.InnerAsgmts, gState.InnerUnasgTasks, gState.InnerInfeaTasks, gState.InnerFeats, gState.InnerSeqDtls = gState.BestInnerSeqs, gState.BestInnerAsgmts, gState.BestInnerUnasgTasks, gState.BestInnerInfeaTasks, bestInnerFeats, gState.BestInnerSeqDtls
}

func RenewInit(BestInnerSeqs *[][]int, BestInnerAsgmts, BestInnerUnasgTasks, InnerInfeaTasks *[]int, gState *GState, gPara *GPara) {
	if len(gState.InnerUnasgTasks) > len(*BestInnerUnasgTasks) || !CheckDataOpt(gState, gPara) {
		return
	}
	if GetDistance(gState.InnerSeqs, gState.InnerAsgmts, gPara) < GetDistance(*BestInnerSeqs, *BestInnerAsgmts, gPara) {
		*BestInnerSeqs, *BestInnerAsgmts, *BestInnerUnasgTasks, *InnerInfeaTasks = CopyResult(gState)
		gState.BestInnerSeqDtls = GenerateSeqDtls(*BestInnerSeqs, *BestInnerAsgmts, gPara)
	}
}
