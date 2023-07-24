package solver

import (
	"fmt"
	"math"
	"math/rand"
	"time"
)

func getSeqInnerDist(seq []int, resId int, lss *LSS) float64 {
	var innerDist float64 = 0.0
	for i := 0; i < len(seq); i++ {
		for j := 0; j < len(seq); j++ {
			innerDist += lss.cost[resId][0][lss.taskLoc[seq[i]]][lss.taskLoc[seq[j]]]
		}
	}
	return innerDist / float64(len(seq)*len(seq))
	//return innerDist
}

func getSeqsInnerDist(seqs [][]int, lss *LSS) (d float64) {
	for i := 0; i < len(seqs); i++ {
		d += getSeqInnerDist(seqs[i], lss.ssr[i], lss)
	}
	return
}
func getSeqInnerDist2(seq []int, resId int, lss *LSS) float64 {
	average := 0.0
	for i := 0; i < len(seq); i++ {
		innerDist := 0.0
		for j := 0; j < len(seq); j++ {
			innerDist += lss.cost[resId][0][lss.taskLoc[seq[i]]][lss.taskLoc[seq[j]]]
		}
		if len(seq) != 1 {
			innerDist /= float64(len(seq) - 1)
		} else {
			innerDist = 0
		}
		mina := 999999999.0
		flag := false
		for k := 0; k < len(lss.ss); k++ {
			if SameSeq(lss.ss[k], seq) {
				flag = true
				continue
			}
			dis := 0.0
			for l := 0; l < len(lss.ss[k]); l++ {
				dis += lss.cost[resId][0][lss.taskLoc[seq[i]]][lss.taskLoc[lss.ss[k][l]]]
			}
			dis /= float64(len(lss.ss[k]))
			if dis < mina {
				mina = dis
			}
		}
		if flag == false {
			fmt.Printf("error\n\n\n")
		}
		average += (mina - innerDist) / math.Max(innerDist, mina)
	}
	average /= float64(len(seq))
	return (1 - average) * 10000
}

func getSeqsInnerDist2(seqs [][]int, lss *LSS) (d float64) {
	if len(seqs) == 0 {
		return 0
	}
	for i := 0; i < len(seqs); i++ {
		d += getSeqInnerDist2(seqs[i], lss.ssr[i], lss)
	}
	d /= float64(len(seqs))
	return
}
func SameSeq(a, b []int) bool {
	lenth := len(a)
	if lenth != len(b) {
		return false
	}
	for i := 0; i < lenth; i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func getSeqsMapCost(seqs [][]int, lss *LSS) (d float64) {
	for i := 0; i < len(seqs); i++ {
		dist := getSeqDist(seqs[i], lss.ssr[i], lss)
		d += getSeqMapCost(lss.ssr[i], lss, dist)
	}
	return
}

func getSeqMapCost(resId int, lss *LSS, dist float64) float64 {
	var seqCost float64 = 0.0
	for t := 0; t < len(lss.CapResCost[resId]); t++ {
		if dist >= lss.CapResCost[resId][t][0] && dist < lss.CapResCost[resId][t][1] {
			seqCost = lss.CapResCost[resId][t][4]
			break
		}
	}
	if dist >= lss.CapResCost[resId][len(lss.CapResCost[resId])-1][1] {
		seqCost = lss.CapResCost[resId][len(lss.CapResCost[resId])-1][4]
	}
	return seqCost
}

func getSeqsInnerDistList(seqs [][]int, lss *LSS) (innerDist []float64) {
	innerDist = make([]float64, len(seqs))
	for i := 0; i < len(seqs); i++ {
		innerDist[i] = getSeqInnerDist(seqs[i], lss.ssr[i], lss)
	}
	return
}

func getSeqsMapCostList(seqs [][]int, lss *LSS) (mapCost []float64) {
	mapCost = make([]float64, len(seqs))
	for i := 0; i < len(seqs); i++ {
		dist := getSeqDist(seqs[i], lss.ssr[i], lss)
		mapCost[i] = getSeqMapCost(lss.ssr[i], lss, dist)
	}
	return
}

func getSeqsMapCostNew(seqs [][]int, lss *LSS) (c float64) {
	for i := 0; i < len(seqs); i++ {
		c += getSeqMapCostNew(seqs[i], lss.ssr[i], lss)
	}
	return
}

func getSeqMapCostNew(seq []int, resId int, lss *LSS) (c float64) {
	dist := getSeqDistNew(seq, resId, lss)
	c = getSeqMapCost(resId, lss, dist)
	return
}

func getSeqsInnerDistNew(seqs [][]int, lss *LSS) (d float64) {
	for i := 0; i < len(seqs); i++ {
		d += getSeqInnerDistNew(seqs[i], lss.ssr[i], lss)
	}
	return
}

func getSeqInnerDistNew(seq []int, resId int, lss *LSS) (d float64) {
	for i := 1; i < len(seq)-1; i++ {
		for j := 1; j < len(seq)-1; j++ {
			d += lss.cost[resId][0][seq[i]][seq[j]]
		}
	}
	//return d
	return d / float64((len(seq)-1)*(len(seq)-1))
}

//***********************LS OPT******************************//
func ShapeAdjustLocalSearch(tEndUnix int64, gPara *GPara, gState *GState) {
	if time.Now().Unix() > tEndUnix-5 {
		return
	}

	var lss = &LSS{}
	cpVarLS(lss, gPara, gState)

	type SolAdj struct {
		ss        [][]int
		dist      float64
		mapCost   float64
		innerDist float64
		cntUn     int
		unTasks   []int
	}

	sol0 := SolAdj{}
	sol0.ss = make([][]int, len(lss.ss))
	for i := 0; i < len(lss.ss); i++ {
		sol0.ss[i] = make([]int, len(lss.ss[i]))
		for j := 0; j < len(lss.ss[i]); j++ {
			sol0.ss[i][j] = lss.ss[i][j]
		}
	}
	sol0.dist = getSeqsDist(lss.ss, lss)
	// todo add cal cost & innerDist
	sol0.mapCost = getSeqsMapCost(lss.ss, lss)
	sol0.innerDist = getSeqsInnerDist(lss.ss, lss)

	sol0.unTasks = CopySliceInt(gState.InnerUnasgTasks)
	sol0.cntUn = len(sol0.unTasks)

	var iterLS int = 100
	var iterLNS int = 3000
	var ratioLNS float64 = 0.4

	sol := SolAdj{}
	sol.ss = make([][]int, len(lss.ss))
	for i := 0; i < len(lss.ss); i++ {
		sol.ss[i] = make([]int, len(lss.ss[i]), cap(lss.ss[i]))
		for j := 0; j < len(lss.ss[i]); j++ {
			sol.ss[i][j] = lss.ss[i][j]
		}
	}
	sol.dist = getSeqsDist(lss.ss, lss)
	// todo add cal cost & innerDist
	sol.mapCost = getSeqsMapCost(lss.ss, lss)
	sol.innerDist = getSeqsInnerDist(lss.ss, lss)
	sol.unTasks = CopySliceInt(gState.InnerUnasgTasks)
	sol.cntUn = len(sol.unTasks)

	//innerDist := GetInnerDistance(lss.ss, lss.ssr, gPara)
	//dist := GetDistance(lss.ss, lss.ssr, gPara)
	//cost := GetSeqsMapCost(lss.ss, lss.ssr, gPara)
	//fmt.Println("RSOShapeAdjust前  cost:", cost, "dist:", dist, "innerDist:", innerDist)
	RSOShapeAdjust(tEndUnix, lss, gPara)
	//fmt.Printf("RSOShapeAdjust后 %v\n", lss.Latency)
	//innerDist = GetInnerDistance(lss.ss, lss.ssr, gPara)
	//dist = GetDistance(lss.ss, lss.ssr, gPara)
	//cost = GetSeqsMapCost(lss.ss, lss.ssr, gPara)
	//fmt.Println("RSOShapeAdjust后  cost:", cost, "dist:", dist, "innerDist:", innerDist)

	if getConstr(lss) > 0 {
		copyTo(sol.ss, lss.ss, nil, nil)
	} else {
		copyTo(lss.ss, sol.ss, nil, nil)
		sol.dist = getSeqsDist(lss.ss, lss)
		// todo add cal cost & innerDist
		sol.mapCost = getSeqsMapCost(lss.ss, lss)
		sol.innerDist = getSeqsInnerDist(lss.ss, lss)
	}
	var d1, d2 float64
	var loopStartTime, loopEndTime int64
	if gState.LSLoopTime < 1 {
		gState.LSLoopTime = 1
	}
	if time.Now().Unix() > tEndUnix-lss.Latency*2 {
		gState.IsTimeEnough = false
	}
	for iter := 0; iter < iterLS && time.Now().Unix() < tEndUnix-5 && gState.IsTimeEnough; iter++ {
		loopStartTime = time.Now().Unix()
		//innerDist := GetInnerDistance(lss.ss, lss.ssr, gPara)
		//dist := GetDistance(lss.ss, lss.ssr, gPara)
		//cost := GetSeqsMapCost(lss.ss, lss.ssr, gPara)
		//fmt.Println("LNSShapeAdjust前  cost:", cost, "dist:", dist, "innerDist:", innerDist)
		d1 = LNSShapeAdjust(tEndUnix, iterLNS, ratioLNS, lss)
		//innerDist = GetInnerDistance(lss.ss, lss.ssr, gPara)
		//dist = GetDistance(lss.ss, lss.ssr, gPara)
		//cost = GetSeqsMapCost(lss.ss, lss.ssr, gPara)
		//fmt.Println("LNSShapeAdjust后  cost:", cost, "dist:", dist, "innerDist:", innerDist)

		if getConstr(lss) > 0 {
			copyTo(sol.ss, lss.ss, nil, nil)
		} else {
			copyTo(lss.ss, sol.ss, nil, nil)
			sol.dist = getSeqsDist(lss.ss, lss)
			// todo add cal cost & innerDist
			sol.mapCost = getSeqsMapCost(lss.ss, lss)
			sol.innerDist = getSeqsInnerDist(lss.ss, lss)
		}
		d2 = RSOShapeAdjust(tEndUnix, lss, gPara)
		if getConstr(lss) > 0 {
			copyTo(sol.ss, lss.ss, nil, nil)
		} else {
			copyTo(lss.ss, sol.ss, nil, nil)
			sol.dist = getSeqsDist(lss.ss, lss)
			// todo add cal cost & innerDist
			sol.mapCost = getSeqsMapCost(lss.ss, lss)
			sol.innerDist = getSeqsInnerDist(lss.ss, lss)
		}

		if sol.cntUn > 0 {
			//innerDist = GetInnerDistance(lss.ss, lss.ssr, gPara)
			//dist = GetDistance(lss.ss, lss.ssr, gPara)
			//cost = GetSeqsMapCost(lss.ss, lss.ssr, gPara)
			//fmt.Println("ReassignShapeAdjust前  cost:", cost, "dist:", dist, "innerDist:", innerDist)
			unTasks := ReassignShapeAdjust(tEndUnix, sol.unTasks, lss)
			//innerDist = GetInnerDistance(lss.ss, lss.ssr, gPara)
			//dist = GetDistance(lss.ss, lss.ssr, gPara)
			//cost = GetSeqsMapCost(lss.ss, lss.ssr, gPara)
			//fmt.Println("ReassignShapeAdjust后  cost:", cost, "dist:", dist, "innerDist:", innerDist)

			if getConstr(lss) > 0 {
				copyTo(sol.ss, lss.ss, nil, nil)
			} else {
				copyTo(lss.ss, sol.ss, nil, nil)
				sol.dist = getSeqsDist(lss.ss, lss)
				// todo add cal cost & innerDist
				sol.mapCost = getSeqsMapCost(lss.ss, lss)
				sol.innerDist = getSeqsInnerDist(lss.ss, lss)
				sol.unTasks = unTasks
				sol.cntUn = len(sol.unTasks)
			}
		}

		if d1 > -1e-3 && d2 > -1e-3 {
			//innerDist = GetInnerDistance(lss.ss, lss.ssr, gPara)
			//dist = GetDistance(lss.ss, lss.ssr, gPara)
			//cost = GetSeqsMapCost(lss.ss, lss.ssr, gPara)
			//fmt.Println("LNSShapeAdjust前  cost:", cost, "dist:", dist, "innerDist:", innerDist)
			d1 = LNSShapeAdjust(tEndUnix, iterLNS*2, 0.5, lss)
			//innerDist = GetInnerDistance(lss.ss, lss.ssr, gPara)
			//dist = GetDistance(lss.ss, lss.ssr, gPara)
			//cost = GetSeqsMapCost(lss.ss, lss.ssr, gPara)
			//fmt.Println("LNSShapeAdjust后  cost:", cost, "dist:", dist, "innerDist:", innerDist)
			if getConstr(lss) > 0 {
				copyTo(sol.ss, lss.ss, nil, nil)
			} else {
				copyTo(lss.ss, sol.ss, nil, nil)
				sol.dist = getSeqsDist(lss.ss, lss)
				// todo add cal cost & innerDist
				sol.mapCost = getSeqsMapCost(lss.ss, lss)
				sol.innerDist = getSeqsInnerDist(lss.ss, lss)
			}
			if d1 > -1e-3 {
				break
			}
		}
		loopEndTime = time.Now().Unix()
		if loopEndTime-loopStartTime > gState.LSLoopTime {
			gState.LSLoopTime = loopEndTime - loopStartTime
		}
		if 2*(tEndUnix-loopEndTime) < 3*gState.LSLoopTime {
			gState.IsTimeEnough = false
		}
		//fmt.Printf("LS2-1: iter:%v   tmpLoopTime:%v    loopMaxTime:%v \n", iter, loopEndTime-loopStartTime, gState.LSLoopTime)
	}

	sol1 := SolAdj{}
	sol1.ss = copyMatrixI(sol.ss)
	sol1.dist = sol.dist
	// todo add mapCost & innerDist
	sol1.mapCost = sol.mapCost
	sol1.innerDist = sol.innerDist
	sol1.cntUn = sol.cntUn
	sol1.unTasks = CopySliceInt(sol.unTasks)

	copyTo(sol0.ss, lss.ss, nil, nil)
	sol.dist = sol0.dist
	// todo add mapCost & innerDist
	sol.mapCost = sol0.mapCost
	sol.innerDist = sol0.innerDist
	sol.unTasks = CopySliceInt(gState.InnerUnasgTasks)
	sol.cntUn = len(sol.unTasks)
	//fmt.Println("开始新迭代！！！！！")
	for iter := 0; iter < iterLS && time.Now().Unix() < tEndUnix-5 && gState.IsTimeEnough; iter++ {
		loopStartTime = time.Now().Unix()
		d1 = LNSShapeAdjust(tEndUnix, iterLNS, ratioLNS, lss)
		if getConstr(lss) > 0 {
			copyTo(sol.ss, lss.ss, nil, nil)
		} else {
			copyTo(lss.ss, sol.ss, nil, nil)
			sol.dist = getSeqsDist(lss.ss, lss)
			// todo add cal cost & innerDist
			sol.mapCost = getSeqsMapCost(lss.ss, lss)
			sol.innerDist = getSeqsInnerDist(lss.ss, lss)
		}
		d2 = RSOShapeAdjust(tEndUnix, lss, gPara)
		if getConstr(lss) > 0 {
			copyTo(sol.ss, lss.ss, nil, nil)
		} else {
			copyTo(lss.ss, sol.ss, nil, nil)
			sol.dist = getSeqsDist(lss.ss, lss)
			// todo add cal cost & innerDist
			sol.mapCost = getSeqsMapCost(lss.ss, lss)
			sol.innerDist = getSeqsInnerDist(lss.ss, lss)
		}
		if sol.cntUn > 0 {
			unTasks := ReassignShapeAdjust(tEndUnix, sol.unTasks, lss)
			if getConstr(lss) > 0 {
				copyTo(sol.ss, lss.ss, nil, nil)
			} else {
				copyTo(lss.ss, sol.ss, nil, nil)
				sol.dist = getSeqsDist(lss.ss, lss)
				// todo add cal cost & innerDist
				sol.mapCost = getSeqsMapCost(lss.ss, lss)
				sol.innerDist = getSeqsInnerDist(lss.ss, lss)
				sol.unTasks = unTasks
				sol.cntUn = len(sol.unTasks)
			}
		}
		if d1 > -1e-3 && d2 > -1e-3 {
			d1 = LNSShapeAdjust(tEndUnix, iterLNS*2, 0.5, lss)
			if getConstr(lss) > 0 {
				copyTo(sol.ss, lss.ss, nil, nil)
			} else {
				copyTo(lss.ss, sol.ss, nil, nil)
				sol.dist = getSeqsDist(lss.ss, lss)
				// todo add cal cost & innerDist
				sol.mapCost = getSeqsMapCost(lss.ss, lss)
				sol.innerDist = getSeqsInnerDist(lss.ss, lss)
			}
			if d1 > -1e-3 {
				break
			}
		}
		loopEndTime = time.Now().Unix()
		if loopEndTime-loopStartTime > gState.LSLoopTime {
			gState.LSLoopTime = loopEndTime - loopStartTime
		}
		if 2*(tEndUnix-loopEndTime) < 3*gState.LSLoopTime {
			gState.IsTimeEnough = false
		}
		//fmt.Printf("LS2-2: iter:%v   tmpLoopTime:%v    loopMaxTime:%v \n", iter, loopEndTime-loopStartTime, gState.LSLoopTime)
	}

	//evaluate the solution by unAsg, mapCost, innerDist
	if sol1.cntUn == sol.cntUn {
		// todo 改innerdist
		if sol1.mapCost == sol.mapCost && sol1.innerDist < sol.innerDist {
			copyTo(sol1.ss, lss.ss, nil, nil)
			gState.InnerUnasgTasks = CopySliceInt(sol1.unTasks)
		} else if sol1.mapCost < sol.mapCost {
			copyTo(sol1.ss, lss.ss, nil, nil)
			gState.InnerUnasgTasks = CopySliceInt(sol1.unTasks)
		} else {
			copyTo(sol.ss, lss.ss, nil, nil)
			gState.InnerUnasgTasks = CopySliceInt(sol.unTasks)
		}
	} else if sol1.cntUn < sol.cntUn {
		copyTo(sol1.ss, lss.ss, nil, nil)
		gState.InnerUnasgTasks = CopySliceInt(sol1.unTasks)
	} else {
		copyTo(sol.ss, lss.ss, nil, nil)
		gState.InnerUnasgTasks = CopySliceInt(sol.unTasks)
	}

	cpVarLSBack(lss, gState)
	//for i := 0; i < len(gState.InnerSeqDtls); i++ {
	//	deltaDis, err := OneRouteTidy(5000, gState, gPara, i, gState.InnerAsgmts[i])
	//	if err != nil {
	//		fmt.Printf("OneRouteTidy Error :%v: %v\n", i, err)
	//	}
	//	fmt.Printf("%v: %f\n", i, deltaDis)
	//}
	for i := 0; i < len(gState.InnerSeqDtls); i++ {
		tmpFCost, tmpMCost := GetSeqMFCostByDist(gPara, gState.InnerAsgmts[i], gState.InnerSeqDtls[i][0])
		gState.InnerSeqDtls[i] = append(gState.InnerSeqDtls[i], tmpFCost)
		gState.InnerSeqDtls[i] = append(gState.InnerSeqDtls[i], tmpMCost)
	}
}

func LNSShapeAdjust(tEndUnix int64, iterLNS int, ratioLNS float64, lss *LSS) (delta float64) {
	if time.Now().Unix() > tEndUnix-5 {
		return
	}

	if len(lss.ss) < 1 || len(lss.tqty) > len(lss.cost[0][0])-1 {
		return
	}

	var c0, c1, innerD0, innerD1 float64

	//dist0 := getSeqsDist(lss.ss, lss)
	//todo 计算innerdist & cost
	mapCost0 := getSeqsMapCost(lss.ss, lss)
	innerDist0 := getSeqsInnerDist(lss.ss, lss)
	ssbest := copyFull(lss)
	fbest := &SSFeat{}
	fbest.dist = getSeqsDistList(lss.ss, lss)
	fbest.dur = getSeqsDurList(lss.ss, lss)
	fbest.qty = getSeqsQtyList(lss.ss, lss)
	fbest.innerDist = getSeqsInnerDistList(lss.ss, lss)
	fbest.mapCost = getSeqsMapCostList(lss.ss, lss)

	ssnew := make([][]int, len(ssbest))
	for i := 0; i < len(ssnew); i++ {
		ssnew[i] = make([]int, 0, cap(ssbest[i]))
	}
	f := &SSFeat{}
	f.dist = make([]float64, len(fbest.dist))
	f.dur = make([]float64, len(fbest.dist))
	f.qty = make([]float64, len(fbest.dist))
	f.innerDist = make([]float64, len(fbest.dist))
	f.mapCost = make([]float64, len(fbest.dist))

	// copy
	copyTo(ssbest, ssnew, fbest, f)

	var r, pos int
	var c, cbest float64
	var success bool
	var r1, r2 int

	for iter := 0; iter < iterLNS && time.Now().Unix() < tEndUnix-5; iter++ {
		success = true
		//d0 = getSeqsDistNew(ssnew, lss)
		//todo 计算innerdist & cost
		c0 = getSeqsMapCostNew(ssnew, lss)
		innerD0 = getSeqsInnerDistNew(ssnew, lss)

		var remain []int
		if len(lss.ss) > 1 {
			r1 = rand.Intn(len(lss.ss))
			r2tmp := rand.Intn(len(lss.ss) - 1)
			if r2tmp < r1 {
				r2 = r2tmp
			} else {
				r2 = r2tmp + 1
			}

			nRm := int(float64(len(ssnew[r1])+len(ssnew[r2])-4) * ratioLNS)
			n1 := rand.Intn(nRm + 1)
			if n1 > len(ssnew[r1])-2 {
				n1 = len(ssnew[r1]) - 2
			}
			n2 := nRm - n1
			if n2 > len(ssnew[r2])-2 {
				n2 = len(ssnew[r2]) - 2
				n1 = nRm - n2
			}
			i1 := 1 + rand.Intn(len(ssnew[r1])-n1-1) // rm [i1, i1+n1-1]
			i2 := 1 + rand.Intn(len(ssnew[r2])-n2-1)

			remain = make([]int, 0, nRm)
			for i := i1; i < i1+n1; i++ {
				remain = append(remain, ssnew[r1][i])
			}
			for i := i2; i < i2+n2; i++ {
				remain = append(remain, ssnew[r2][i])
			}
			rand.Seed(time.Now().UnixNano())
			rand.Shuffle(len(remain), func(i, j int) { remain[i], remain[j] = remain[j], remain[i] })

			ssnew[r1] = append(ssnew[r1][:i1], ssnew[r1][i1+n1:]...)
			ssnew[r2] = append(ssnew[r2][:i2], ssnew[r2][i2+n2:]...)
			for _, r := range []int{r1, r2} {
				f.dist[r] = getSeqDistNew(ssnew[r], lss.ssr[r], lss)
				f.qty[r] = getSeqQtyNew(ssnew[r], lss)
				f.dur[r] = getSeqDurNew(ssnew[r], lss.ssr[r], lss)
				f.innerDist[r] = getSeqInnerDistNew(ssnew[r], lss.ssr[r], lss)
				f.mapCost[r] = getSeqMapCostNew(ssnew[r], lss.ssr[r], lss)
			}
		} else {
			r1 = rand.Intn(len(lss.ss))

			nRm := int(float64(len(ssnew[r1])-2) * ratioLNS)
			n1 := rand.Intn(nRm + 1)
			i1 := 1 + rand.Intn(len(ssnew[r1])-n1-1) // rm [i1, i1+n1-1]

			remain = make([]int, 0, nRm)
			for i := i1; i < i1+n1; i++ {
				remain = append(remain, ssnew[r1][i])
			}
			rand.Seed(time.Now().UnixNano())
			rand.Shuffle(len(remain), func(i, j int) { remain[i], remain[j] = remain[j], remain[i] })

			ssnew[r1] = append(ssnew[r1][:i1], ssnew[r1][i1+n1:]...)
			for _, r := range []int{r1} {
				f.dist[r] = getSeqDistNew(ssnew[r], lss.ssr[r], lss)
				f.qty[r] = getSeqQtyNew(ssnew[r], lss)
				f.dur[r] = getSeqDurNew(ssnew[r], lss.ssr[r], lss)
				f.innerDist[r] = getSeqInnerDistNew(ssnew[r], lss.ssr[r], lss)
				f.mapCost[r] = getSeqMapCost(lss.ssr[r], lss, f.dist[r])
			}
		}

		cbest = 1e8
		for _, t := range remain {
			r, pos = -1, -1
			c, cbest = 0, 1e8
			for i := 0; i < len(ssnew); i++ {
				//计算inner dist
				for j := 0; j < len(ssnew[i])-1; j++ {
					c = -lss.cost[lss.ssr[i]][0][ssnew[i][j]][ssnew[i][j+1]] + lss.cost[lss.ssr[i]][0][ssnew[i][j]][t] + lss.cost[lss.ssr[i]][0][t][ssnew[i][j+1]]
					if c < cbest {
						if f.qty[i]+lss.tqty[t-1] <= lss.rqty[lss.ssr[i]] && f.dist[i]+c <= lss.rdist[lss.ssr[i]] && f.dur[i]+lss.tdur[t-1]-lss.cost[lss.ssr[i]][1][ssnew[i][j]][ssnew[i][j+1]]+lss.cost[lss.ssr[i]][1][ssnew[i][j]][t]+lss.cost[lss.ssr[i]][1][t][ssnew[i][j+1]] <= lss.rdur[lss.ssr[i]] {
							cbest = c
							r = i
							pos = j
						}
					}
				}
				//记录最好 idx
			}
			if r >= 0 {
				f.dist[r] += -lss.cost[lss.ssr[r]][0][ssnew[r][pos]][ssnew[r][pos+1]] + lss.cost[lss.ssr[r]][0][ssnew[r][pos]][t] + lss.cost[lss.ssr[r]][0][t][ssnew[r][pos+1]] //+= cbest
				f.qty[r] += lss.tqty[t-1]
				f.dur[r] += lss.tdur[t-1] - lss.cost[lss.ssr[r]][1][ssnew[r][pos]][ssnew[r][pos+1]] + lss.cost[lss.ssr[r]][1][ssnew[r][pos]][t] + lss.cost[lss.ssr[r]][1][t][ssnew[r][pos+1]]
				f.innerDist[r] = getSeqInnerDistNew(ssnew[r], lss.ssr[r], lss)
				f.mapCost[r] = getSeqMapCost(lss.ssr[r], lss, f.dist[r])
				tmp := CopySliceInt(ssnew[r][pos+1:])
				ssnew[r] = append(ssnew[r][:pos+1], t)
				ssnew[r] = append(ssnew[r], tmp...)
			} else {
				success = false
				break
			}
		}

		if !success {
			copyTo(ssbest, ssnew, fbest, f)
			continue
		}

		//d1 = getSeqsDistNew(ssnew, lss)
		//todo 计算innerdist & cost
		c1 = getSeqsMapCostNew(ssnew, lss)
		innerD1 = getSeqsInnerDistNew(ssnew, lss)

		//todo 计算 innerdist + cost1 < cost2 && innerdist
		//if c1 == c0 && innerD1 < innerD0 {
		//	copyTo(ssnew, ssbest, f, fbest)
		//} else if c1 < c0 {
		//	copyTo(ssnew, ssbest, f, fbest)
		//} else {
		//	copyTo(ssbest, ssnew, fbest, f)
		//}

		if c1 <= c0 && innerD1 < innerD0 {
			copyTo(ssnew, ssbest, f, fbest)
		} else {
			copyTo(ssbest, ssnew, fbest, f)
		}
	}

	//dist1 := getSeqsDistNew(ssbest, lss)
	innerDist1 := getSeqsInnerDistNew(ssbest, lss)
	mapCost1 := getSeqsMapCostNew(ssbest, lss)
	//计算 inner dist
	// todo dist1 < dist0 改成 cost1 < cost0
	//if mapCost1 == mapCost0 && innerDist1 < innerDist0 {
	//	copyPartBack(ssbest, lss.ss)
	//	delta = innerDist1 - innerDist0
	//} else if mapCost1 < mapCost0 {
	//	copyPartBack(ssbest, lss.ss)
	//	delta = mapCost1 - mapCost0
	//}
	if mapCost1 <= mapCost0 && innerDist1 < innerDist0 {
		copyPartBack(ssbest, lss.ss)
		delta = innerDist1 - innerDist0
	}
	return
}

func ReassignShapeAdjust(tEndUnix int64, tasks []int, lss *LSS) (unTasks []int) {
	//处理orphan
	if time.Now().Unix() > tEndUnix-5 {
		unTasks = tasks
		return
	}

	if len(lss.tqty) > len(lss.cost[0][0])-1 {
		return
	}

	ssnew := copyFull(lss)
	fbest := &SSFeat{}
	fbest.dist = getSeqsDistList(lss.ss, lss)
	fbest.dur = getSeqsDurList(lss.ss, lss)
	fbest.qty = getSeqsQtyList(lss.ss, lss)

	// 生成nodes和标记位
	nodes := make([]int, len(tasks))
	unassigned := make([]bool, len(tasks))
	for i := 0; i < len(tasks); i++ {
		nodes[i] = lss.taskLoc[tasks[i]]
		unassigned[i] = true
	}

	// each task insertion lss.cost
	seqpos := make([][2]int, len(nodes))    // （node) best route and best pos
	priceMin := make([]float64, len(nodes)) // （node） best price
	price := make([][]float64, len(nodes))  // （node, seq）

	var t, tbest, kbest, pos int
	var c, cbest float64

	f := &SSFeat{}
	f.dist = CopySliceFloat(fbest.dist)
	f.dur = CopySliceFloat(fbest.dur)
	f.qty = CopySliceFloat(fbest.qty)

	cbest = 1e8
	kbest = -1
	for k := 0; k < len(nodes) && time.Now().Unix() < tEndUnix-5; k++ {
		priceMin[k] = 1e8
		t = nodes[k]
		price[k] = make([]float64, len(ssnew))
		for i := 0; i < len(ssnew); i++ {
			// best seq
			price[k][i] = 1e8
			pos = -1
			for j := 0; j < len(ssnew[i])-1; j++ {
				c = -lss.cost[lss.ssr[i]][0][ssnew[i][j]][ssnew[i][j+1]] + lss.cost[lss.ssr[i]][0][ssnew[i][j]][t] + lss.cost[lss.ssr[i]][0][t][ssnew[i][j+1]]
				if c < price[k][i] {
					if f.qty[i]+lss.tqty[t-1] <= lss.rqty[lss.ssr[i]] && f.dist[i]+c <= lss.rdist[lss.ssr[i]] && f.dur[i]+lss.tdur[t-1]-lss.cost[lss.ssr[i]][1][ssnew[i][j]][ssnew[i][j+1]]+lss.cost[lss.ssr[i]][1][ssnew[i][j]][t]+lss.cost[lss.ssr[i]][1][t][ssnew[i][j+1]] <= lss.rdur[lss.ssr[i]] {
						price[k][i] = c
						pos = j
					}
				}
			}
			if price[k][i] < priceMin[k] {
				priceMin[k] = price[k][i]
				seqpos[k][0] = i
				seqpos[k][1] = pos
			}
		}
		// check global best
		if priceMin[k] < cbest {
			cbest = priceMin[k]
			kbest = k
		}
	}

	var rnew int
	for q := 0; q < len(unassigned) && time.Now().Unix() < tEndUnix-5; q++ {
		if kbest < 0 {
			break
		}
		tbest = nodes[kbest]
		// insert best node
		unassigned[kbest] = false
		r := seqpos[kbest][0]
		p := seqpos[kbest][1]
		if p < 0 { // need to find best pos
			cbest = 1e8
			for j := 0; j < len(ssnew[r])-1; j++ {
				c = -lss.cost[lss.ssr[r]][0][ssnew[r][j]][ssnew[r][j+1]] + lss.cost[lss.ssr[r]][0][ssnew[r][j]][tbest] + lss.cost[lss.ssr[r]][0][tbest][ssnew[r][j+1]]
				if c < cbest {
					cbest = c
					p = j
				}
			}
		}
		f.qty[r] += lss.tqty[t-1]
		f.dist[r] += -lss.cost[lss.ssr[r]][0][ssnew[r][p]][ssnew[r][p+1]] + lss.cost[lss.ssr[r]][0][ssnew[r][p]][tbest] + lss.cost[lss.ssr[r]][0][tbest][ssnew[r][p+1]]
		f.dur[r] += lss.tdur[tbest-1] - lss.cost[lss.ssr[r]][1][ssnew[r][p]][ssnew[r][p+1]] + lss.cost[lss.ssr[r]][1][ssnew[r][p]][tbest] + lss.cost[lss.ssr[r]][1][tbest][ssnew[r][p+1]]
		tmp := CopySliceInt(ssnew[r][p+1:])
		ssnew[r] = append(ssnew[r][:p+1], tbest)
		ssnew[r] = append(ssnew[r], tmp...)

		// update info (cbest, kbest, seqpos)
		cbest = 1e8
		kbest = -1
		for k := 0; k < len(nodes); k++ {
			if !unassigned[k] {
				continue
			}
			t = nodes[k]
			price[k][r] = 1e8
			pos = -1
			for j := 0; j < len(ssnew[r])-1; j++ {
				c = -lss.cost[lss.ssr[r]][0][ssnew[r][j]][ssnew[r][j+1]] + lss.cost[lss.ssr[r]][0][ssnew[r][j]][t] + lss.cost[lss.ssr[r]][0][t][ssnew[r][j+1]]
				if c < price[k][r] {
					if f.qty[r]+lss.tqty[t-1] <= lss.rqty[lss.ssr[r]] && f.dist[r]+c <= lss.rdist[lss.ssr[r]] && f.dur[r]+lss.tdur[t-1]-lss.cost[lss.ssr[r]][1][ssnew[r][j]][ssnew[r][j+1]]+lss.cost[lss.ssr[r]][1][ssnew[r][j]][t]+lss.cost[lss.ssr[r]][1][t][ssnew[r][j+1]] <= lss.rdur[lss.ssr[r]] {
						price[k][r] = c
						pos = j
					}
				}
			}
			rnew, priceMin[k] = findMin(price[k])
			if rnew == r { // need update
				seqpos[k][0] = r
				seqpos[k][1] = pos
			} else if seqpos[k][0] == r {
				seqpos[k][0] = rnew
				seqpos[k][1] = -1 // lazy way
			}
			if priceMin[k] < cbest {
				cbest = priceMin[k]
				kbest = k
			}
		}
	}

	unTasks = make([]int, 0)
	for i := 0; i < len(tasks); i++ {
		if unassigned[i] {
			unTasks = append(unTasks, tasks[i])
		}
	}

	copyPartBack(ssnew, lss.ss)
	return
}

func RSOShapeAdjust(tEndUnix int64, lss *LSS, gPara *GPara) (delta float64) {
	if time.Now().Unix() > tEndUnix-5 {
		return
	}

	//d0 := getSeqsDist(lss.ss, lss)
	c0 := getSeqsMapCost(lss.ss, lss)
	innerD0 := getSeqsInnerDist(lss.ss, lss)
	lss.ssqty = getSeqsQtyList(lss.ss, lss)
	lss.ssdur = getSeqsDurList(lss.ss, lss)
	lss.ssdist = getSeqsDistList(lss.ss, lss)

	var d float64
	lss.Latency = 5
	//innerDist := GetInnerDistance(lss.ss, lss.ssr, gPara)
	//dist := GetDistance(lss.ss, lss.ssr, gPara)
	//cost := GetSeqsMapCost(lss.ss, lss.ssr, gPara)
	//fmt.Println("run_global_4_sa前  cost:", cost, "dist:", dist, "innerDist:", innerDist)
	d += run_global_4_sa(tEndUnix, lss)
	//innerDist = GetInnerDistance(lss.ss, lss.ssr, gPara)
	//dist = GetDistance(lss.ss, lss.ssr, gPara)
	//cost = GetSeqsMapCost(lss.ss, lss.ssr, gPara)
	//fmt.Println("run_global_4_sa后  cost:", cost, "dist:", dist, "innerDist:", innerDist)

	//innerDist = GetInnerDistance(lss.ss, lss.ssr, gPara)
	//dist = GetDistance(lss.ss, lss.ssr, gPara)
	//cost = GetSeqsMapCost(lss.ss, lss.ssr, gPara)
	//fmt.Println("run_global_3_sa前  cost:", cost, "dist:", dist, "innerDist:", innerDist)
	d += run_global_3_sa(tEndUnix, lss)
	//innerDist = GetInnerDistance(lss.ss, lss.ssr, gPara)
	//dist = GetDistance(lss.ss, lss.ssr, gPara)
	//cost = GetSeqsMapCost(lss.ss, lss.ssr, gPara)
	//fmt.Println("run_global_3_sa后  cost:", cost, "dist:", dist, "innerDist:", innerDist)

	//innerDist = GetInnerDistance(lss.ss, lss.ssr, gPara)
	//dist = GetDistance(lss.ss, lss.ssr, gPara)
	//cost = GetSeqsMapCost(lss.ss, lss.ssr, gPara)
	//fmt.Println("run_global_2_sa前  cost:", cost, "dist:", dist, "innerDist:", innerDist)
	d += run_global_2_sa(tEndUnix, lss)
	//innerDist = GetInnerDistance(lss.ss, lss.ssr, gPara)
	//dist = GetDistance(lss.ss, lss.ssr, gPara)
	//cost = GetSeqsMapCost(lss.ss, lss.ssr, gPara)
	//fmt.Println("run_global_2_sa后  cost:", cost, "dist:", dist, "innerDist:", innerDist)

	//innerDist = GetInnerDistance(lss.ss, lss.ssr, gPara)
	//dist = GetDistance(lss.ss, lss.ssr, gPara)
	//cost = GetSeqsMapCost(lss.ss, lss.ssr, gPara)
	//fmt.Println("run_global_5_sa前  cost:", cost, "dist:", dist, "innerDist:", innerDist)
	d += run_global_5_sa(tEndUnix, lss)
	//innerDist = GetInnerDistance(lss.ss, lss.ssr, gPara)
	//dist = GetDistance(lss.ss, lss.ssr, gPara)
	//cost = GetSeqsMapCost(lss.ss, lss.ssr, gPara)
	//fmt.Println("run_global_5_sa后  cost:", cost, "dist:", dist, "innerDist:", innerDist)

	if d < 0 {
		d += run_global_5_sa(tEndUnix, lss)
	}
	d += run_local_1_sa(tEndUnix, lss)
	if d < 0 {
		d += run_global_4_sa(tEndUnix, lss)
	}
	if d < 0 {
		d += run_local_1_sa(tEndUnix, lss)
	}
	if d < 0 {
		d += run_global_2_sa(tEndUnix, lss)
	}
	if d < 0 {
		d += run_local_1_sa(tEndUnix, lss)
	}
	if d < 0 {
		d += run_global_3_sa(tEndUnix, lss)
	}
	if d < 0 {
		d += run_local_1_sa(tEndUnix, lss)
	}
	//
	//innerDist = GetInnerDistance(lss.ss, lss.ssr, gPara)
	//dist = GetDistance(lss.ss, lss.ssr, gPara)
	//cost = GetSeqsMapCost(lss.ss, lss.ssr, gPara)
	//fmt.Println("run_global_6_sa前  cost:", cost, "dist:", dist, "innerDist:", innerDist)
	d += run_global_6_sa(tEndUnix, lss)
	//innerDist = GetInnerDistance(lss.ss, lss.ssr, gPara)
	//dist = GetDistance(lss.ss, lss.ssr, gPara)
	//cost = GetSeqsMapCost(lss.ss, lss.ssr, gPara)
	//fmt.Println("run_global_6_sa后  cost:", cost, "dist:", dist, "innerDist:", innerDist)
	//
	if d < 0 {
		d += run_local_1_sa(tEndUnix, lss)
	}

	if d < 0 {
		d += run_global_2_sa(tEndUnix, lss)
	}
	if d < 0 {
		d += run_local_1_sa(tEndUnix, lss)
	}
	if d < 0 {
		d += run_global_3_sa(tEndUnix, lss)
	}
	if d < 0 {
		d += run_local_1_sa(tEndUnix, lss)
	}
	if d < 0 {
		d += run_global_4_sa(tEndUnix, lss)
	}
	if d < 0 {
		d += run_local_1_sa(tEndUnix, lss)
	}
	if d < 0 {
		d += run_global_5_sa(tEndUnix, lss)
	}
	if d < 0 {
		d += run_local_1_sa(tEndUnix, lss)
	}
	if d < 0 {
		d += run_global_6_sa(tEndUnix, lss)
	}
	if d < 0 {
		d += run_local_1_sa(tEndUnix, lss)
	}
	//
	//innerDist = GetInnerDistance(lss.ss, lss.ssr, gPara)
	//dist = GetDistance(lss.ss, lss.ssr, gPara)
	//cost = GetSeqsMapCost(lss.ss, lss.ssr, gPara)
	//fmt.Println("run_global_1_sa前  cost:", cost, "dist:", dist, "innerDist:", innerDist)
	d += run_global_1_sa(tEndUnix, lss)
	//innerDist = GetInnerDistance(lss.ss, lss.ssr, gPara)
	//dist = GetDistance(lss.ss, lss.ssr, gPara)
	//cost = GetSeqsMapCost(lss.ss, lss.ssr, gPara)
	//fmt.Println("run_global_1_sa后  cost:", cost, "dist:", dist, "innerDist:", innerDist)
	//
	if d < 0 {
		d += run_global_5_sa(tEndUnix, lss)
	}
	if d < 0 {
		d += run_global_4_sa(tEndUnix, lss)
	}
	if d < 0 {
		d += run_global_3_sa(tEndUnix, lss)
	}
	if d < 0 {
		d += run_global_2_sa(tEndUnix, lss)
	}
	if d < 0 {
		d += run_local_1_sa(tEndUnix, lss)
	}

	if d < 0 {
		var d1 float64
		for iter := 0; iter < 10; iter++ {
			d1 = run_global_5_sa(tEndUnix, lss) + run_local_1_sa(tEndUnix, lss)
			d1 += run_global_2_sa(tEndUnix, lss) + run_local_1_sa(tEndUnix, lss)
			d1 += run_global_3_sa(tEndUnix, lss) + run_local_1_sa(tEndUnix, lss)
			d1 += run_global_4_sa(tEndUnix, lss) + run_local_1_sa(tEndUnix, lss)
			d1 += run_global_1_sa(tEndUnix, lss) + run_local_1_sa(tEndUnix, lss)
			d += d1
			if d1 > -1e-3 || time.Now().Unix() > tEndUnix-lss.Latency {
				break
			}
		}
	}

	if d < 0 {
		d += run_global_6_sa(tEndUnix, lss)
	}
	if d < 0 {
		d += run_local_1_sa(tEndUnix, lss)
	}
	if d < 0 {
		d += run_global_1_sa(tEndUnix, lss)
	}
	if d < 0 {
		d += run_global_5_sa(tEndUnix, lss)
	}
	if d < 0 {
		d += run_global_4_sa(tEndUnix, lss)
	}
	if d < 0 {
		d += run_global_3_sa(tEndUnix, lss)
	}
	if d < 0 {
		d += run_global_2_sa(tEndUnix, lss)
	}

	d += run_local_2_sa(tEndUnix, lss)

	if d < 0 {
		d += run_global_5_sa(tEndUnix, lss) + run_local_1_sa(tEndUnix, lss)
	}
	if d < 0 {
		d += run_global_2_sa(tEndUnix, lss) + run_local_1_sa(tEndUnix, lss)
	}
	if d < 0 {
		d += run_global_3_sa(tEndUnix, lss) + run_local_1_sa(tEndUnix, lss)
	}
	if d < 0 {
		d += run_global_4_sa(tEndUnix, lss) + run_local_1_sa(tEndUnix, lss)
	}
	if d < 0 {
		d += run_global_1_sa(tEndUnix, lss) + run_local_1_sa(tEndUnix, lss)
	}
	//
	//innerDist = GetInnerDistance(lss.ss, lss.ssr, gPara)
	//dist = GetDistance(lss.ss, lss.ssr, gPara)
	//cost = GetSeqsMapCost(lss.ss, lss.ssr, gPara)
	//fmt.Println("run_global_8_sa前  cost:", cost, "dist:", dist, "innerDist:", innerDist)
	d += run_global_8_sa(tEndUnix, lss)
	//innerDist = GetInnerDistance(lss.ss, lss.ssr, gPara)
	//dist = GetDistance(lss.ss, lss.ssr, gPara)
	//cost = GetSeqsMapCost(lss.ss, lss.ssr, gPara)
	//fmt.Println("run_global_8_sa后  cost:", cost, "dist:", dist, "innerDist:", innerDist)
	//
	if d < 0 {
		d += run_local_2_sa(tEndUnix, lss)
	}
	//
	//innerDist = GetInnerDistance(lss.ss, lss.ssr, gPara)
	//dist = GetDistance(lss.ss, lss.ssr, gPara)
	//cost = GetSeqsMapCost(lss.ss, lss.ssr, gPara)
	//fmt.Println("run_global_7_sa前  cost:", cost, "dist:", dist, "innerDist:", innerDist)
	d += run_global_7_sa(tEndUnix, lss)
	//innerDist = GetInnerDistance(lss.ss, lss.ssr, gPara)
	//dist = GetDistance(lss.ss, lss.ssr, gPara)
	//cost = GetSeqsMapCost(lss.ss, lss.ssr, gPara)
	//fmt.Println("run_global_7_sa后  cost:", cost, "dist:", dist, "innerDist:", innerDist)
	//
	if d < 0 {
		d += run_local_2_sa(tEndUnix, lss)
	}

	if d < 0 {
		d += run_global_5_sa(tEndUnix, lss)
	}
	if d < 0 {
		d += run_local_2_sa(tEndUnix, lss)
	}
	if d < 0 {
		d += run_global_2_sa(tEndUnix, lss)
	}
	if d < 0 {
		d += run_local_2_sa(tEndUnix, lss)
	}
	if d < 0 {
		d += run_global_3_sa(tEndUnix, lss)
	}
	if d < 0 {
		d += run_local_2_sa(tEndUnix, lss)
	}
	if d < 0 {
		d += run_global_4_sa(tEndUnix, lss)
	}
	if d < 0 {
		d += run_local_2_sa(tEndUnix, lss)
	}
	if d < 0 {
		d += run_global_1_sa(tEndUnix, lss)
	}
	if d < 0 {
		d += run_local_2_sa(tEndUnix, lss)
	}

	c1 := getSeqsMapCost(lss.ss, lss)
	innerD1 := getSeqsInnerDist(lss.ss, lss)

	if c1 < c0 {
		delta = c1 - c0
	} else {
		delta = innerD1 - innerD0
	}
	return
}

// **************************** local ****************************************
func run_local_1_sa(tEndUnix int64, lss *LSS) (delta_all float64) {
	var delta1, delta2, delta3 float64
	for i := 0; i < len(lss.ss); i++ {
		for iter := 0; iter < 10; iter++ {
			if time.Now().Unix() > tEndUnix-lss.Latency*2 {
				return
			}
			start := time.Now()
			delta1 = run_brute_2opt_sa(i, lss)
			delta2 = run_brute_reloc1_sa(i, lss)
			delta3 = run_brute_swap1_sa(i, lss)
			if delta1 < -1e-3 || delta2 < -1e-3 || delta3 < -1e-3 {
				delta_all += (delta1 + delta2 + delta3)
			} else {
				break
			}
			elapsed := time.Since(start)
			tmpLatency := int64(elapsed/time.Second) + 1
			if tmpLatency > lss.Latency {
				lss.Latency = tmpLatency
			}
		}
	}
	return
}

func run_local_2_sa(tEndUnix int64, lss *LSS) (delta_all float64) {
	var delta1, delta2, delta3, delta4 float64
	for i := 0; i < len(lss.ss); i++ {
		for iter := 0; iter < 10; iter++ {
			if time.Now().Unix() > tEndUnix-lss.Latency*2 {
				return
			}
			start := time.Now()
			delta1 = run_brute_2opt_sa(i, lss)
			delta2 = run_brute_reloc1_sa(i, lss)
			delta3 = run_brute_swap1_sa(i, lss)
			delta4 = run_brute_swap12_sa(i, lss)
			if delta1 < -1e-3 || delta2 < -1e-3 || delta3 < -1e-3 || delta4 < -1e-3 {
				delta_all += (delta1 + delta2 + delta3 + delta4)
			} else {
				break
			}
			elapsed := time.Since(start)
			tmpLatency := int64(elapsed/time.Second) + 1
			if tmpLatency > lss.Latency {
				lss.Latency = tmpLatency
			}
		}
	}
	return
}

// *************************** swap *****************************************
func run_global_1_sa(tEndUnix int64, lss *LSS) (delta_all float64) {
	var delta float64
	d_beg := getSeqsDist(lss.ss, lss)
	for iter := 0; iter < 2; iter++ {
		if time.Now().Unix() > tEndUnix-lss.Latency*2 {
			return
		}
		start := time.Now()
		delta = run_global_swap1_sa(tEndUnix, lss)
		if delta < -1e-3 {
			delta_all += delta
		} else {
			break
		}
		elapsed := time.Since(start)
		tmpLatency := int64(elapsed/time.Second) + 1
		if tmpLatency > lss.Latency {
			lss.Latency = tmpLatency
		}
	}
	delta_all = getSeqsDist(lss.ss, lss) - d_beg
	return
}

//组内交换，指定route里某两点交换位置
func run_global_swap1_sa(tEndUnix int64, lss *LSS) (delta float64) {
	for i := 0; i < len(lss.ss); i++ {
		if time.Now().Unix() > tEndUnix-5 {
			return
		}
		run_brute_swap1_sa(i, lss)
	}
	return
}

func run_global_7_sa(tEndUnix int64, lss *LSS) (delta_all float64) {
	var delta float64
	for i := 0; i < len(lss.ss); i++ {
		for j := i + 1; j < len(lss.ss); j++ {
			if time.Now().Unix() > tEndUnix-lss.Latency*2 {
				return
			}
			start := time.Now()
			for m := 0; m < len(lss.ss[i]); m++ {
				for n := 0; n < len(lss.ss[j]); n++ {
					delta = try_swap2_sa(i, j, lss.ssr[i], lss.ssr[j], m, n, lss)
					delta_all += delta
				}
			}
			elapsed := time.Since(start)
			tmpLatency := int64(elapsed/time.Second) + 1
			if tmpLatency > lss.Latency {
				lss.Latency = tmpLatency
			}
		}
	}
	return
}

func run_global_8_sa(tEndUnix int64, lss *LSS) (delta_all float64) {
	var delta float64
	for i := 0; i < len(lss.ss); i++ {
		for j := i + 1; j < len(lss.ss); j++ {
			if time.Now().Unix() > tEndUnix-lss.Latency*2 {
				return
			}
			start := time.Now()
			for m := 0; m < len(lss.ss[i])-1; m++ {
				for n := 0; n < len(lss.ss[j])-1; n++ {
					delta = try_swap22_sa(i, j, lss.ssr[i], lss.ssr[j], m, n, lss)
					delta_all += delta
				}
			}
			elapsed := time.Since(start)
			tmpLatency := int64(elapsed/time.Second) + 1
			if tmpLatency > lss.Latency {
				lss.Latency = tmpLatency
			}
		}
	}
	return
}

//组内交换，指定route里某两点交换位置
func run_brute_swap1_sa(s1 int, lss *LSS) (delta float64) {
	for i := 0; i < len(lss.ss[s1])-1; i++ {
		for j := i + 1; j < len(lss.ss[s1]); j++ {
			delta += try_swap1_sa(s1, i, j, lss)
		}
	}
	return
}

//组间交换，route1的连续两个点和route2的一个点交换位置
func run_brute_swap12_sa(s1 int, lss *LSS) (delta float64) {
	for i := 0; i < len(lss.ss[s1])-2; i++ {
		for j := i + 2; j < len(lss.ss[s1]); j++ {
			delta += try_swap12_sa(s1, i, j, lss)
		}
	}
	return
}

//组内交换，指定route里某两点交换位置
func try_swap1_sa(s1 int, i1, i2 int, lss *LSS) (delta float64) {
	delta = d_swap1_sa(lss.ss[s1], i1, i2, lss.ssr[s1], lss)
	if delta < -1e-3 {
		lss.ss[s1][i1], lss.ss[s1][i2] = lss.ss[s1][i2], lss.ss[s1][i1]
		lss.ssdur[s1] = getSeqDur(lss.ss[s1], lss.ssr[s1], lss)
		lss.ssdist[s1] = getSeqDist(lss.ss[s1], lss.ssr[s1], lss)
	} else {
		delta = 0
	}
	return
}

func d_swap1_sa(s1 []int, i1, i2 int, r1 int, lss *LSS) (delta float64) {
	if i1 > i2 {
		i1, i2 = i2, i1
	}

	n1 := make([]int, 3)
	n2 := make([]int, 3)
	if i1 > 0 {
		n1[0] = lss.taskLoc[s1[i1-1]]
	}
	n1[1] = lss.taskLoc[s1[i1]]
	n1[2] = lss.taskLoc[s1[i1+1]]
	n2[0] = lss.taskLoc[s1[i2-1]]
	n2[1] = lss.taskLoc[s1[i2]]
	if i2 < len(s1)-1 {
		n2[2] = lss.taskLoc[s1[i2+1]]
	}

	if i1+1 < i2 {
		delta = -lss.cost[r1][0][n1[0]][n1[1]] - lss.cost[r1][0][n1[1]][n1[2]] - lss.cost[r1][0][n2[0]][n2[1]] - lss.cost[r1][0][n2[1]][n2[2]] +
			lss.cost[r1][0][n1[0]][n2[1]] + lss.cost[r1][0][n2[1]][n1[2]] + lss.cost[r1][0][n2[0]][n1[1]] + lss.cost[r1][0][n1[1]][n2[2]]
	} else if i1+1 == i2 {
		delta = -lss.cost[r1][0][n1[0]][n1[1]] - lss.cost[r1][0][n1[1]][n1[2]] - lss.cost[r1][0][n2[1]][n2[2]] +
			lss.cost[r1][0][n1[0]][n2[1]] + lss.cost[r1][0][n2[1]][n1[1]] + lss.cost[r1][0][n1[1]][n2[2]]
	}
	//delta = 100 //delta > 0
	return
}

//组间交换，route1的连续两个点和route2的一个点交换位置
func try_swap12_sa(s1 int, i1, i2 int, lss *LSS) (delta float64) {
	delta = d_swap12_sa(lss.ss[s1], i1, i2, lss.ssr[s1], lss)
	if delta < -1e-3 {
		tmp := lss.ss[s1][i1]
		lss.ss[s1][i1] = lss.ss[s1][i2]
		lss.ss[s1][i1+1], lss.ss[s1][i2] = lss.ss[s1][i2], lss.ss[s1][i1+1]
		for i := i1 + 1; i < i2-1; i++ {
			lss.ss[s1][i] = lss.ss[s1][i+1]
		}
		lss.ss[s1][i2-1] = tmp
		lss.ssdur[s1] = getSeqDur(lss.ss[s1], lss.ssr[s1], lss)
		lss.ssdist[s1] = getSeqDist(lss.ss[s1], lss.ssr[s1], lss)
	} else {
		delta = 0
	}
	return
}

//组间交换，route1的连续两个点和route2的一个点交换位置
func d_swap12_sa(s1 []int, i1, i2 int, r1 int, lss *LSS) (delta float64) {
	if i1+2 < i2 {
		n1 := make([]int, 4)
		n2 := make([]int, 3)
		if i1 > 0 {
			n1[0] = lss.taskLoc[s1[i1-1]]
		}
		n1[1] = lss.taskLoc[s1[i1]]
		n1[2] = lss.taskLoc[s1[i1+1]]
		n1[3] = lss.taskLoc[s1[i1+2]]
		n2[0] = lss.taskLoc[s1[i2-1]]
		n2[1] = lss.taskLoc[s1[i2]]
		if i2 < len(s1)-1 {
			n2[2] = lss.taskLoc[s1[i2+1]]
		}
		delta = -lss.cost[r1][0][n1[0]][n1[1]] - lss.cost[r1][0][n1[2]][n1[3]] - lss.cost[r1][0][n2[0]][n2[1]] - lss.cost[r1][0][n2[1]][n2[2]] +
			lss.cost[r1][0][n1[0]][n2[1]] + lss.cost[r1][0][n2[1]][n1[3]] + lss.cost[r1][0][n2[0]][n1[1]] + lss.cost[r1][0][n1[2]][n2[2]]
		// delta = 100
	}
	return
}

//组间交换，两条route中各任一点交换位置
func try_swap2_sa(s1, s2 int, r1 int, r2 int, i1, i2 int, lss *LSS) (delta float64) {
	delta = d_swap2_sa(s1, s2, i1, i2, r1, r2, lss)
	if delta < -1e-3 {
		lss.ssqty[s1] = lss.ssqty[s1] - lss.tqty[lss.ss[s1][i1]] + lss.tqty[lss.ss[s2][i2]]
		lss.ssqty[s2] = lss.ssqty[s2] + lss.tqty[lss.ss[s1][i1]] - lss.tqty[lss.ss[s2][i2]]
		lss.sswei[s1] = lss.sswei[s1] - lss.twei[lss.ss[s1][i1]] + lss.twei[lss.ss[s2][i2]]
		lss.sswei[s2] = lss.sswei[s2] + lss.twei[lss.ss[s1][i1]] - lss.twei[lss.ss[s2][i2]]
		lss.ss[s1][i1], lss.ss[s2][i2] = lss.ss[s2][i2], lss.ss[s1][i1]
		lss.ssdur[s1] = getSeqDur(lss.ss[s1], lss.ssr[s1], lss)
		lss.ssdur[s2] = getSeqDur(lss.ss[s2], lss.ssr[s2], lss)
		lss.ssdist[s1] = getSeqDist(lss.ss[s1], lss.ssr[s1], lss)
		lss.ssdist[s2] = getSeqDist(lss.ss[s2], lss.ssr[s2], lss)
	}
	return delta
}

//组间交换，两条route中各任一点交换位置
func d_swap2_sa(s1, s2, i1, i2, r1, r2 int, lss *LSS) (delta float64) {
	n1 := make([]int, 3)
	n2 := make([]int, 3)

	if i1 > 0 {
		n1[0] = lss.taskLoc[lss.ss[s1][i1-1]]
	}
	n1[1] = lss.taskLoc[lss.ss[s1][i1]]
	if i1 < len(lss.ss[s1])-1 {
		n1[2] = lss.taskLoc[lss.ss[s1][i1+1]]
	}

	if i2 > 0 {
		n2[0] = lss.taskLoc[lss.ss[s2][i2-1]]
	}
	n2[1] = lss.taskLoc[lss.ss[s2][i2]]
	if i2 < len(lss.ss[s2])-1 {
		n2[2] = lss.taskLoc[lss.ss[s2][i2+1]]
	}

	cost1 := getSeqMapCost(r1, lss, lss.ssdist[s1])
	cost2 := getSeqMapCost(r2, lss, lss.ssdist[s2])
	sumCost := cost1 + cost2

	var tmpS1 = make([]int, 0)
	tmpS1 = append(tmpS1, lss.ss[s1][:i1]...)
	tmpS1 = append(tmpS1, lss.ss[s2][i2])
	tmpS1 = append(tmpS1, lss.ss[s1][i1+1:]...)
	var tmpS2 = make([]int, 0)
	tmpS2 = append(tmpS2, lss.ss[s2][:i2]...)
	tmpS2 = append(tmpS2, lss.ss[s1][i1])
	tmpS2 = append(tmpS2, lss.ss[s2][i2+1:]...)

	tmpDist1 := getSeqDist(tmpS1, r1, lss)
	tmpDist2 := getSeqDist(tmpS2, r2, lss)
	tmpCost1 := getSeqMapCost(r1, lss, tmpDist1)
	tmpCost2 := getSeqMapCost(r2, lss, tmpDist2)
	tmpSumCost := tmpCost1 + tmpCost2

	if tmpSumCost <= sumCost {
		innerD1 := getSeqInnerDist(lss.ss[s1], r1, lss)
		innerD2 := getSeqInnerDist(lss.ss[s2], r2, lss)
		oldS1 := CopySliceInt(lss.ss[s1])
		oldS2 := CopySliceInt(lss.ss[s2])
		lss.ss[s1] = tmpS1
		lss.ss[s2] = tmpS2
		tmpInnerD1 := getSeqInnerDist(tmpS1, r1, lss)
		tmpInnerD2 := getSeqInnerDist(tmpS2, r2, lss)
		lss.ss[s1] = oldS1
		lss.ss[s2] = oldS2
		delta = tmpInnerD1 + tmpInnerD2 - innerD1 - innerD2
	} else {
		delta = 100
	}

	//delta = -lss.cost[r1][0][n1[0]][n1[1]] - lss.cost[r1][0][n1[1]][n1[2]] + lss.cost[r1][0][n1[0]][n2[1]] + lss.cost[r1][0][n2[1]][n1[2]] -
	//	lss.cost[r2][0][n2[0]][n2[1]] - lss.cost[r2][0][n2[1]][n2[2]] + lss.cost[r2][0][n2[0]][n1[1]] + lss.cost[r2][0][n1[1]][n2[2]]

	if delta < -1e-3 {
		if lss.ssqty[s1]-lss.tqty[lss.ss[s1][i1]]+lss.tqty[lss.ss[s2][i2]] > lss.rqty[r1] || lss.ssqty[s2]+lss.tqty[lss.ss[s1][i1]]-lss.tqty[lss.ss[s2][i2]] > lss.rqty[r2] {
			delta = 0
		} else if lss.ssdur[s1]-lss.tdur[lss.ss[s1][i1]]+lss.tdur[lss.ss[s2][i2]]-lss.cost[r1][1][n1[0]][n1[1]]-lss.cost[r1][1][n1[1]][n1[2]]+lss.cost[r1][1][n1[0]][n2[1]]+lss.cost[r1][1][n2[1]][n1[2]] > lss.rdur[r1] {
			delta = 0
		} else if lss.ssdur[s2]+lss.tdur[lss.ss[s1][i1]]-lss.tdur[lss.ss[s2][i2]]-lss.cost[r2][1][n2[0]][n2[1]]-lss.cost[r2][1][n2[1]][n2[2]]+lss.cost[r2][1][n2[0]][n1[1]]+lss.cost[r2][1][n1[1]][n2[2]] > lss.rdur[r2] {
			delta = 0
		}
	} else {
		delta = 0
	}

	if delta < -1e-3 {
		if lss.sswei[s1]-lss.twei[lss.ss[s1][i1]]+lss.twei[lss.ss[s2][i2]] > lss.rwei[r1] || lss.sswei[s2]+lss.twei[lss.ss[s1][i1]]-lss.twei[lss.ss[s2][i2]] > lss.rwei[r2] {
			delta = 0
		} else if lss.ssdist[s1]-lss.cost[r1][0][n1[0]][n1[1]]-lss.cost[r1][0][n1[1]][n1[2]]+lss.cost[r1][0][n1[0]][n2[1]]+lss.cost[r1][0][n2[1]][n1[2]] > lss.rdist[r1] {
			delta = 0
		} else if lss.ssdist[s2]-lss.cost[r2][0][n2[0]][n2[1]]-lss.cost[r2][0][n2[1]][n2[2]]+lss.cost[r2][0][n2[0]][n1[1]]+lss.cost[r2][0][n1[1]][n2[2]] > lss.rdist[r2] {
			delta = 0
		}
	}
	return
}

//组间交换，两条route中，各选取一段线段(连续两点)，交换位置
func try_swap22_sa(s1, s2 int, r1 int, r2 int, i1, i2 int, lss *LSS) (delta float64) {
	// i1, i1+1, i2, i2+1
	if len(lss.ss[s1]) < 2 || len(lss.ss[s2]) < 2 {
		return
	}
	delta = d_swap22_sa(s1, s2, i1, i2, r1, r2, lss)
	if delta < -1e-3 {
		lss.ssqty[s1] = lss.ssqty[s1] - lss.tqty[lss.ss[s1][i1]] - lss.tqty[lss.ss[s1][i1+1]] + lss.tqty[lss.ss[s2][i2]] + lss.tqty[lss.ss[s2][i2+1]]
		lss.ssqty[s2] = lss.ssqty[s2] + lss.tqty[lss.ss[s1][i1]] + lss.tqty[lss.ss[s1][i1+1]] - lss.tqty[lss.ss[s2][i2]] - lss.tqty[lss.ss[s2][i2+1]]
		lss.sswei[s1] = lss.sswei[s1] - lss.twei[lss.ss[s1][i1]] - lss.twei[lss.ss[s1][i1+1]] + lss.twei[lss.ss[s2][i2]] + lss.twei[lss.ss[s2][i2+1]]
		lss.sswei[s2] = lss.sswei[s2] + lss.twei[lss.ss[s1][i1]] + lss.twei[lss.ss[s1][i1+1]] - lss.twei[lss.ss[s2][i2]] - lss.twei[lss.ss[s2][i2+1]]
		lss.ss[s1][i1], lss.ss[s2][i2] = lss.ss[s2][i2], lss.ss[s1][i1]
		lss.ss[s1][i1+1], lss.ss[s2][i2+1] = lss.ss[s2][i2+1], lss.ss[s1][i1+1]
		lss.ssdur[s1] = getSeqDur(lss.ss[s1], lss.ssr[s1], lss)
		lss.ssdur[s2] = getSeqDur(lss.ss[s2], lss.ssr[s2], lss)
		lss.ssdist[s1] = getSeqDist(lss.ss[s1], lss.ssr[s1], lss)
		lss.ssdist[s2] = getSeqDist(lss.ss[s2], lss.ssr[s2], lss)
	}
	return delta
}

//组间交换，两条route中，各选取一段线段(连续两点)，交换位置
func d_swap22_sa(s1, s2, i1, i2, r1, r2 int, lss *LSS) (delta float64) {
	n1 := make([]int, 4)
	n2 := make([]int, 4)

	if i1 > 0 {
		n1[0] = lss.taskLoc[lss.ss[s1][i1-1]]
	}
	n1[1] = lss.taskLoc[lss.ss[s1][i1]]
	n1[2] = lss.taskLoc[lss.ss[s1][i1+1]]
	if i1+1 < len(lss.ss[s1])-1 {
		n1[3] = lss.taskLoc[lss.ss[s1][i1+2]]
	}

	if i2 > 0 {
		n2[0] = lss.taskLoc[lss.ss[s2][i2-1]]
	}
	n2[1] = lss.taskLoc[lss.ss[s2][i2]]
	n2[2] = lss.taskLoc[lss.ss[s2][i2+1]]
	if i2+1 < len(lss.ss[s2])-1 {
		n2[3] = lss.taskLoc[lss.ss[s2][i2+2]]
	}

	cost1 := getSeqMapCost(r1, lss, lss.ssdist[s1])
	cost2 := getSeqMapCost(r2, lss, lss.ssdist[s2])
	sumCost := cost1 + cost2

	var tmpS1 = make([]int, 0)
	tmpS1 = append(tmpS1, lss.ss[s1][:i1]...)
	tmpS1 = append(tmpS1, lss.ss[s2][i2:i2+2]...)
	tmpS1 = append(tmpS1, lss.ss[s1][i1+2:]...)
	var tmpS2 = make([]int, 0)
	tmpS2 = append(tmpS2, lss.ss[s2][:i2]...)
	tmpS2 = append(tmpS2, lss.ss[s1][i1:i1+2]...)
	tmpS2 = append(tmpS2, lss.ss[s2][i2+2:]...)

	tmpDist1 := lss.ssdist[s1] - lss.cost[r1][0][n1[0]][n1[1]] - lss.cost[r1][0][n1[1]][n1[2]] - lss.cost[r1][0][n1[2]][n1[3]] +
		lss.cost[r1][0][n1[0]][n2[1]] + lss.cost[r1][0][n2[1]][n2[2]] + lss.cost[r1][0][n2[2]][n1[3]]
	tmpDist2 := lss.ssdist[s2] - lss.cost[r2][0][n2[0]][n2[1]] - lss.cost[r2][0][n2[1]][n2[2]] - lss.cost[r2][0][n2[2]][n2[3]] +
		lss.cost[r2][0][n2[0]][n1[1]] + lss.cost[r2][0][n1[1]][n1[2]] + lss.cost[r2][0][n1[2]][n2[3]]
	tmpCost1 := getSeqMapCost(r1, lss, tmpDist1)
	tmpCost2 := getSeqMapCost(r2, lss, tmpDist2)
	tmpSumCost := tmpCost1 + tmpCost2

	if tmpSumCost <= sumCost {
		innerD1 := getSeqInnerDist(lss.ss[s1], r1, lss)
		innerD2 := getSeqInnerDist(lss.ss[s2], r2, lss)
		//oldS1 := CopySliceInt(lss.ss[s1])
		//oldS2 := CopySliceInt(lss.ss[s2])
		//lss.ss[s1] = tmpS1
		//lss.ss[s2] = tmpS2
		tmpInnerD1 := getSeqInnerDist(tmpS1, r1, lss)
		tmpInnerD2 := getSeqInnerDist(tmpS2, r2, lss)
		//lss.ss[s1] = oldS1
		//lss.ss[s2] = oldS2
		delta = tmpInnerD1 + tmpInnerD2 - innerD1 - innerD2
	} else {
		delta = 100
	}

	//delta = -lss.cost[r1][0][n1[0]][n1[1]] - lss.cost[r1][0][n1[2]][n1[3]] + lss.cost[r1][0][n1[0]][n2[1]] + lss.cost[r1][0][n2[2]][n1[3]] - lss.cost[r2][0][n2[0]][n2[1]] - lss.cost[r2][0][n2[2]][n2[3]] + lss.cost[r2][0][n2[0]][n1[1]] + lss.cost[r2][0][n1[2]][n2[3]]

	if delta < -1e-3 {
		if lss.ssqty[s1]-lss.tqty[lss.ss[s1][i1]]-lss.tqty[lss.ss[s1][i1+1]]+lss.tqty[lss.ss[s2][i2]]+lss.tqty[lss.ss[s2][i2+1]] > lss.rqty[r1] || lss.ssqty[s2]+lss.tqty[lss.ss[s1][i1]]+lss.tqty[lss.ss[s1][i1+1]]-lss.tqty[lss.ss[s2][i2]]-lss.tqty[lss.ss[s2][i2+1]] > lss.rqty[r2] {
			delta = 0
		} else if lss.ssdur[s1]-lss.tdur[lss.ss[s1][i1]]-lss.tdur[lss.ss[s1][i1+1]]+lss.tdur[lss.ss[s2][i2]]+lss.tdur[lss.ss[s2][i2+1]]-
			lss.cost[r1][1][n1[0]][n1[1]]-lss.cost[r1][1][n1[1]][n1[2]]-lss.cost[r1][1][n1[2]][n1[3]]+
			lss.cost[r1][1][n1[0]][n2[1]]+lss.cost[r1][1][n2[1]][n2[2]]+lss.cost[r1][1][n2[2]][n1[3]] > lss.rdur[r1] {
			delta = 0
		} else if lss.ssdur[s2]+lss.tdur[lss.ss[s1][i1]]+lss.tdur[lss.ss[s1][i1+1]]-lss.tdur[lss.ss[s2][i2]]-lss.tdur[lss.ss[s2][i2+1]]-
			lss.cost[r2][1][n2[0]][n2[1]]-lss.cost[r2][1][n2[1]][n2[2]]-lss.cost[r2][1][n2[2]][n2[3]]+
			lss.cost[r2][1][n2[0]][n1[1]]+lss.cost[r2][1][n1[1]][n1[2]]+lss.cost[r2][1][n1[2]][n2[3]] > lss.rdur[r2] {
			delta = 0
		}
	} else {
		delta = 0
	}

	if delta < -1e-3 {
		if lss.sswei[s1]-lss.twei[lss.ss[s1][i1]]-lss.twei[lss.ss[s1][i1+1]]+lss.twei[lss.ss[s2][i2]]+lss.twei[lss.ss[s2][i2+1]] > lss.rwei[r1] || lss.sswei[s2]+lss.twei[lss.ss[s1][i1]]+lss.twei[lss.ss[s1][i1+1]]-lss.twei[lss.ss[s2][i2]]-lss.twei[lss.ss[s2][i2+1]] > lss.rwei[r2] {
			delta = 0
		} else if lss.ssdist[s1]-lss.cost[r1][0][n1[0]][n1[1]]-lss.cost[r1][0][n1[1]][n1[2]]-lss.cost[r1][0][n1[2]][n1[3]]+
			lss.cost[r1][0][n1[0]][n2[1]]+lss.cost[r1][0][n2[1]][n2[2]]+lss.cost[r1][0][n2[2]][n1[3]] > lss.rdist[r1] {
			delta = 0
		} else if lss.ssdist[s2]-lss.cost[r2][0][n2[0]][n2[1]]-lss.cost[r2][0][n2[1]][n2[2]]-lss.cost[r2][0][n2[2]][n2[3]]+
			lss.cost[r2][0][n2[0]][n1[1]]+lss.cost[r2][0][n1[1]][n1[2]]+lss.cost[r2][0][n1[2]][n2[3]] > lss.rdist[r2] {
			delta = 0
		}
	}

	return
}

// **************************************** reloc ****************************************
//组内交换，指定route里某一点插入其他两点间
func run_brute_reloc1_sa(s1 int, lss *LSS) (delta float64) {
	for i := 0; i < len(lss.ss[s1]); i++ {
		for j := 0; j < len(lss.ss[s1]); j++ {
			delta += try_reloc1_sa(s1, i, j, lss)
		}
	}
	return
}

func run_global_2_sa(tEndUnix int64, lss *LSS) (delta_all float64) {
	var delta float64
	for i := 0; i < len(lss.ss); i++ {
		for j := 0; j < len(lss.ss); j++ {
			if time.Now().Unix() > tEndUnix-lss.Latency*2 {
				return
			}
			start := time.Now()
			for m := 0; m < len(lss.ss[i]); m++ {
				for n := 0; n < len(lss.ss[j]); n++ {
					delta = try_reloc2_sa(i, j, m, n, lss)
					delta_all += delta
				}
			}
			elapsed := time.Since(start)
			tmpLatency := int64(elapsed/time.Second) + 1
			if tmpLatency > lss.Latency {
				lss.Latency = tmpLatency
			}
		}
	}
	return
}

func run_global_3_sa(tEndUnix int64, lss *LSS) (delta_all float64) {
	var delta float64
	for i := 0; i < len(lss.ss); i++ {
		for j := 0; j < len(lss.ss); j++ {
			if time.Now().Unix() > tEndUnix-lss.Latency*2 {
				return
			}
			start := time.Now()
			for m := 0; m < len(lss.ss[i]); m++ {
				for n := 0; n < len(lss.ss[j]); n++ {
					delta = try_reloc3_sa(i, j, m, n, lss)
					delta_all += delta
				}
			}
			elapsed := time.Since(start)
			tmpLatency := int64(elapsed/time.Second) + 1
			if tmpLatency > lss.Latency {
				lss.Latency = tmpLatency
			}
		}
	}
	return
}

func run_global_4_sa(tEndUnix int64, lss *LSS) (delta_all float64) {
	var delta float64
	for i := 0; i < len(lss.ss); i++ {
		for j := 0; j < len(lss.ss); j++ {
			if time.Now().Unix() > tEndUnix-lss.Latency*2 {
				return
			}
			start := time.Now()
			for m := 0; m < len(lss.ss[i]); m++ {
				for n := 0; n < len(lss.ss[j]); n++ {
					delta = try_reloc4_sa(i, j, m, n, lss)
					delta_all += delta
				}
			}
			elapsed := time.Since(start)
			tmpLatency := int64(elapsed/time.Second) + 1
			if tmpLatency > lss.Latency {
				lss.Latency = tmpLatency
			}
		}
	}
	return
}

//组内交换，指定route里某一点插入其他两点间
func try_reloc1_sa(s1 int, i1, i2 int, lss *LSS) (delta float64) {
	if len(lss.ss[s1]) < 3 {
		return 0
	}
	if i1 == i2 || i1 == i2+1 {
		return 0
	}
	delta = d_reloc1_sa(lss.ss[s1], i1, i2, lss.ssr[s1], lss)

	if delta < -1e-3 {
		if i1 < i2 {
			tmp := lss.ss[s1][i1]
			for i := i1; i < i2; i++ {
				lss.ss[s1][i] = lss.ss[s1][i+1]
			}
			lss.ss[s1][i2] = tmp
		} else {
			tmp := lss.ss[s1][i1]
			for i := i1; i > i2+1; i-- {
				lss.ss[s1][i] = lss.ss[s1][i-1]
			}
			lss.ss[s1][i2+1] = tmp
		}

		lss.ssdur[s1] = getSeqDur(lss.ss[s1], lss.ssr[s1], lss)
		lss.ssdist[s1] = getSeqDist(lss.ss[s1], lss.ssr[s1], lss)

	} else {
		delta = 0
	}
	return
}

//组内交换，指定route里某一点插入其他两点间
func d_reloc1_sa(s1 []int, i1, i2 int, r1 int, lss *LSS) (delta float64) {
	if i1 == i2+1 {
		return
	}

	n1 := make([]int, 3)
	n2 := make([]int, 2)
	if i1 > 0 {
		n1[0] = lss.taskLoc[s1[i1-1]]
	}
	n1[1] = lss.taskLoc[s1[i1]]
	if i1 < len(s1)-1 {
		n1[2] = lss.taskLoc[s1[i1+1]]
	}
	n2[0] = lss.taskLoc[s1[i2]]
	if i2 < len(s1)-1 {
		n2[1] = lss.taskLoc[s1[i2+1]]
	}

	delta = -lss.cost[r1][0][n1[0]][n1[1]] - lss.cost[r1][0][n1[1]][n1[2]] - lss.cost[r1][0][n2[0]][n2[1]] +
		lss.cost[r1][0][n1[0]][n1[2]] + lss.cost[r1][0][n2[0]][n1[1]] + lss.cost[r1][0][n1[1]][n2[1]]
	//delta = 100

	return
}

//组间交换，一个route的1个点插入另一个route中
func try_reloc2_sa(s1, s2 int, i1, i2 int, lss *LSS) (delta float64) {
	if s1 == s2 || i1 >= len(lss.ss[s1]) || i2 >= len(lss.ss[s2]) {
		return 0
	}

	delta = d_reloc2_sa(s1, s2, i1, i2, lss.ssr[s1], lss.ssr[s2], lss.ssqty[s2], lss.ssdur[s2], lss.sswei[s2], lss.ssdist[s2], lss)

	if delta < -1e-3 {
		lss.ssqty[s1] = lss.ssqty[s1] - lss.tqty[lss.ss[s1][i1]]
		lss.ssqty[s2] = lss.ssqty[s2] + lss.tqty[lss.ss[s1][i1]]
		lss.sswei[s1] = lss.sswei[s1] - lss.twei[lss.ss[s1][i1]]
		lss.sswei[s2] = lss.sswei[s2] + lss.twei[lss.ss[s1][i1]]

		lss.ss[s2] = append(lss.ss[s2], 0)
		for i := len(lss.ss[s2]) - 1; i > i2+1; i-- {
			lss.ss[s2][i] = lss.ss[s2][i-1]
		}
		lss.ss[s2][i2+1] = lss.ss[s1][i1]
		lss.ss[s1] = append(lss.ss[s1][:i1], lss.ss[s1][i1+1:]...)

		lss.ssdur[s1] = getSeqDur(lss.ss[s1], lss.ssr[s1], lss)
		lss.ssdur[s2] = getSeqDur(lss.ss[s2], lss.ssr[s2], lss)
		lss.ssdist[s1] = getSeqDist(lss.ss[s1], lss.ssr[s1], lss)
		lss.ssdist[s2] = getSeqDist(lss.ss[s2], lss.ssr[s2], lss)

	} else {
		delta = 0
	}
	return
}

//组间交换，一个route的1个点插入另一个route中
func d_reloc2_sa(ss1, ss2 int, i1, i2 int, r1, r2 int, qty2 float64, dur2 float64, wei2 float64, dist2 float64, lss *LSS) (delta float64) {
	s1 := lss.ss[ss1]
	s2 := lss.ss[ss2]
	n1 := make([]int, 3)
	if i1 > 0 {
		n1[0] = lss.taskLoc[s1[i1-1]]
	}
	n1[1] = lss.taskLoc[s1[i1]]
	if i1 < len(s1)-1 {
		n1[2] = lss.taskLoc[s1[i1+1]]
	}

	n2 := make([]int, 2)
	n2[0] = lss.taskLoc[s2[i2]]
	if i2 < len(s2)-1 {
		n2[1] = lss.taskLoc[s2[i2+1]]
	}

	dist1 := getSeqDist(s1, r1, lss)
	cost1 := getSeqMapCost(r1, lss, dist1)
	cost2 := getSeqMapCost(r2, lss, dist2)
	sumCost := cost1 + cost2

	var tmpS1 = make([]int, 0)
	tmpS1 = append(tmpS1, s1[:i1]...)
	tmpS1 = append(tmpS1, s1[i1+1:]...)
	var tmpS2 = make([]int, 0)
	tmpS2 = append(tmpS2, s2[:i2+1]...)
	tmpS2 = append(tmpS2, s1[i1])
	tmpS2 = append(tmpS2, s2[i2+1:]...)

	tmpDist1 := getSeqDist(tmpS1, r1, lss)
	tmpDist2 := dist2 - lss.cost[r2][0][n2[0]][n2[1]] + lss.cost[r2][0][n2[0]][n1[1]] + lss.cost[r2][0][n1[1]][n2[1]]
	tmpCost1 := getSeqMapCost(r1, lss, tmpDist1)
	tmpCost2 := getSeqMapCost(r2, lss, tmpDist2)
	tmpSumCost := tmpCost1 + tmpCost2

	if tmpSumCost <= sumCost {
		innerD1 := getSeqInnerDist(s1, r1, lss)
		innerD2 := getSeqInnerDist(s2, r2, lss)
		//oldS1 := CopySliceInt(lss.ss[ss1])
		//oldS2 := CopySliceInt(lss.ss[ss2])
		//lss.ss[ss1] = tmpS1
		//lss.ss[ss2] = tmpS2
		tmpInnerD1 := getSeqInnerDist(tmpS1, r1, lss)
		tmpInnerD2 := getSeqInnerDist(tmpS2, r2, lss)
		//lss.ss[ss1] = oldS1
		//lss.ss[ss2] = oldS2
		delta = tmpInnerD1 + tmpInnerD2 - innerD1 - innerD2
	} else {
		delta = 100
	}

	//delta = -lss.cost[r1][0][n1[0]][n1[1]] - lss.cost[r1][0][n1[1]][n1[2]] - lss.cost[r2][0][n2[0]][n2[1]] +
	//	lss.cost[r1][0][n1[0]][n1[2]] + lss.cost[r2][0][n2[0]][n1[1]] + lss.cost[r2][0][n1[1]][n2[1]]
	//
	if delta < -1e-3 {
		if qty2+lss.tqty[s1[i1]] > lss.rqty[r2] {
			delta = 0
		} else if dur2+lss.tdur[s1[i1]]-lss.cost[r2][1][n2[0]][n2[1]]+lss.cost[r2][1][n2[0]][n1[1]]+lss.cost[r2][1][n1[1]][n2[1]] > lss.rdur[r2] {
			delta = 0
		}
	}
	if delta < -1e-3 {
		if wei2+lss.twei[s1[i1]] > lss.rwei[r2] {
			delta = 0
		} else if dist2-lss.cost[r2][0][n2[0]][n2[1]]+lss.cost[r2][0][n2[0]][n1[1]]+lss.cost[r2][0][n1[1]][n2[1]] > lss.rdist[r2] {
			delta = 0
		}
	}
	return
}

//组间交换，一个route中点一段线段(连续两点)，插入另一route中
func try_reloc3_sa(s1, s2 int, i1, i2 int, lss *LSS) (delta float64) {
	if s1 == s2 || i1+1 >= len(lss.ss[s1]) || i2 >= len(lss.ss[s2]) {
		return 0
	}

	var pivot bool
	delta, pivot = d_reloc3_sa(s1, s2, i1, i2, lss.ssr[s1], lss.ssr[s2], lss.ssqty[s2], lss.ssdur[s2], lss.sswei[s2], lss.ssdist[s2], lss)

	if delta < -1e-3 {
		lss.ssqty[s1] = lss.ssqty[s1] - lss.tqty[lss.ss[s1][i1]] - lss.tqty[lss.ss[s1][i1+1]]
		lss.ssqty[s2] = lss.ssqty[s2] + lss.tqty[lss.ss[s1][i1]] + lss.tqty[lss.ss[s1][i1+1]]
		lss.sswei[s1] = lss.sswei[s1] - lss.twei[lss.ss[s1][i1]] - lss.twei[lss.ss[s1][i1+1]]
		lss.sswei[s2] = lss.sswei[s2] + lss.twei[lss.ss[s1][i1]] + lss.twei[lss.ss[s1][i1+1]]

		lss.ss[s2] = append(lss.ss[s2], []int{0, 0}...)
		for i := len(lss.ss[s2]) - 1; i > i2+2; i-- {
			lss.ss[s2][i] = lss.ss[s2][i-2]
		}
		if !pivot {
			lss.ss[s2][i2+1] = lss.ss[s1][i1]
			lss.ss[s2][i2+2] = lss.ss[s1][i1+1]
		} else {
			lss.ss[s2][i2+1] = lss.ss[s1][i1+1]
			lss.ss[s2][i2+2] = lss.ss[s1][i1]
		}

		lss.ss[s1] = append(lss.ss[s1][:i1], lss.ss[s1][i1+2:]...)

		lss.ssdur[s1] = getSeqDur(lss.ss[s1], lss.ssr[s1], lss)
		lss.ssdur[s2] = getSeqDur(lss.ss[s2], lss.ssr[s2], lss)
		lss.ssdist[s1] = getSeqDist(lss.ss[s1], lss.ssr[s1], lss)
		lss.ssdist[s2] = getSeqDist(lss.ss[s2], lss.ssr[s2], lss)

	} else {
		delta = 0
	}
	return
}

//组间交换，一个route中点一段线段(连续两点)，插入另一route中
func d_reloc3_sa(ss1, ss2 int, i1, i2 int, r1, r2 int, qty2 float64, dur2 float64, wei2 float64, dist2 float64, lss *LSS) (delta float64, pivot bool) {
	s1 := lss.ss[ss1]
	s2 := lss.ss[ss2]
	n1 := make([]int, 4)
	n2 := make([]int, 2)

	if i1 > 0 {
		n1[0] = lss.taskLoc[s1[i1-1]]
	}
	n1[1] = lss.taskLoc[s1[i1]]
	n1[2] = lss.taskLoc[s1[i1+1]]
	if i1+1 < len(s1)-1 {
		n1[3] = lss.taskLoc[s1[i1+2]]
	}

	n2[0] = lss.taskLoc[s2[i2]]
	if i2 < len(s2)-1 {
		n2[1] = lss.taskLoc[s2[i2+1]]
	}

	dist1 := getSeqDist(s1, r1, lss)
	cost1 := getSeqMapCost(r1, lss, dist1)
	cost2 := getSeqMapCost(r2, lss, dist2)
	sumCost := cost1 + cost2

	var tmpS1 = make([]int, 0)
	tmpS1 = append(tmpS1, s1[:i1]...)
	tmpS1 = append(tmpS1, s1[i1+2:]...)
	var tmpS2 = make([]int, 0)
	tmpS2 = append(tmpS2, s2[:i2+1]...)
	tmpS2 = append(tmpS2, s1[i1:i1+2]...)
	tmpS2 = append(tmpS2, s2[i2+1:]...)

	tmpDist1 := dist1 - lss.cost[r1][0][n1[0]][n1[1]] - lss.cost[r1][0][n1[2]][n1[3]] + lss.cost[r1][0][n1[0]][n1[3]]
	tmpDist2 := dist2 - lss.cost[r2][0][n2[0]][n2[1]] + lss.cost[r2][0][n2[0]][n1[1]] + lss.cost[r2][0][n1[1]][n1[2]] + lss.cost[r2][0][n1[2]][n2[1]]
	tmpCost1 := getSeqMapCost(r1, lss, tmpDist1)
	tmpCost2 := getSeqMapCost(r2, lss, tmpDist2)
	tmpSumCost := tmpCost1 + tmpCost2

	if tmpSumCost <= sumCost {
		innerD1 := getSeqInnerDist(s1, r1, lss)
		innerD2 := getSeqInnerDist(s2, r2, lss)
		//oldS1 := CopySliceInt(lss.ss[ss1])
		//oldS2 := CopySliceInt(lss.ss[ss2])
		//lss.ss[ss1] = tmpS1
		//lss.ss[ss2] = tmpS2
		tmpInnerD1 := getSeqInnerDist(tmpS1, r1, lss)
		tmpInnerD2 := getSeqInnerDist(tmpS2, r2, lss)
		//lss.ss[ss1] = oldS1
		//lss.ss[ss2] = oldS2
		delta = tmpInnerD1 + tmpInnerD2 - innerD1 - innerD2
	} else {
		delta = 100
	}

	//delta = -lss.cost[r1][0][n1[0]][n1[1]] - lss.cost[r1][0][n1[2]][n1[3]] - lss.cost[r2][0][n2[0]][n2[1]] +
	//	lss.cost[r1][0][n1[0]][n1[3]] + lss.cost[r2][0][n2[0]][n1[1]] + lss.cost[r2][0][n1[2]][n2[1]]
	//
	//if delta > -1e-3 {
	//	delta = -lss.cost[r1][0][n1[0]][n1[1]] - lss.cost[r1][0][n1[1]][n1[2]] - lss.cost[r1][0][n1[2]][n1[3]] - lss.cost[r2][0][n2[0]][n2[1]] +
	//		lss.cost[r1][0][n1[0]][n1[3]] + lss.cost[r2][0][n2[0]][n1[2]] + lss.cost[r2][0][n1[2]][n1[1]] + lss.cost[r2][0][n1[1]][n2[1]]
	//	pivot = true
	//}

	if delta < -1e-3 {
		if qty2+lss.tqty[s1[i1]]+lss.tqty[s1[i1+1]] > lss.rqty[r2] {
			delta = 0
		} else if !pivot {
			if dur2+lss.tdur[s1[i1]]+lss.tdur[s1[i1+1]]-lss.cost[r2][1][n2[0]][n2[1]]+lss.cost[r2][1][n2[0]][n1[1]]+lss.cost[r2][1][n1[1]][n1[2]]+lss.cost[r2][1][n1[2]][n2[1]] > lss.rdur[r2] {
				delta = 0
			}
		} else if pivot {
			if dur2+lss.tdur[s1[i1]]+lss.tdur[s1[i1+1]]-lss.cost[r2][1][n2[0]][n2[1]]+lss.cost[r2][1][n2[0]][n1[2]]+lss.cost[r2][1][n1[2]][n1[1]]+lss.cost[r2][1][n1[1]][n2[1]] > lss.rdur[r2] {
				delta = 0
			}
		}
	}
	if delta < -1e-3 {
		if wei2+lss.twei[s1[i1]]+lss.twei[s1[i1+1]] > lss.rwei[r2] {
			delta = 0
		} else if !pivot {
			if dist2-lss.cost[r2][0][n2[0]][n2[1]]+lss.cost[r2][0][n2[0]][n1[1]]+lss.cost[r2][0][n1[1]][n1[2]]+lss.cost[r2][0][n1[2]][n2[1]] > lss.rdist[r2] {
				delta = 0
			}
		} else if pivot {
			if dist2-lss.cost[r2][0][n2[0]][n2[1]]+lss.cost[r2][0][n2[0]][n1[2]]+lss.cost[r2][0][n1[2]][n1[1]]+lss.cost[r2][0][n1[1]][n2[1]] > lss.rdist[r2] {
				delta = 0
			}
		}
	}
	return
}

//选取route1的连续5个点，route2的连续2个点，将route1中间3个点插入route2的两个点中
func try_reloc4_sa(s1, s2 int, i1, i2 int, lss *LSS) (delta float64) {
	if s1 == s2 || i1+2 >= len(lss.ss[s1]) || i2 >= len(lss.ss[s2]) {
		return 0
	}
	var pivot bool
	delta, pivot = d_reloc4_sa(s1, s2, i1, i2, lss.ssr[s1], lss.ssr[s2], lss.ssqty[s2], lss.ssdur[s2], lss.sswei[s2], lss.ssdist[s2], lss)

	if delta < -1e-3 {
		lss.ssqty[s1] = lss.ssqty[s1] - lss.tqty[lss.ss[s1][i1]] - lss.tqty[lss.ss[s1][i1+1]] - lss.tqty[lss.ss[s1][i1+2]]
		lss.ssqty[s2] = lss.ssqty[s2] + lss.tqty[lss.ss[s1][i1]] + lss.tqty[lss.ss[s1][i1+1]] + lss.tqty[lss.ss[s1][i1+2]]
		lss.sswei[s1] = lss.sswei[s1] - lss.twei[lss.ss[s1][i1]] - lss.twei[lss.ss[s1][i1+1]] - lss.twei[lss.ss[s1][i1+2]]
		lss.sswei[s2] = lss.sswei[s2] + lss.twei[lss.ss[s1][i1]] + lss.twei[lss.ss[s1][i1+1]] + lss.twei[lss.ss[s1][i1+2]]

		lss.ss[s2] = append(lss.ss[s2], []int{0, 0, 0}...)
		for i := len(lss.ss[s2]) - 1; i > i2+3; i-- {
			lss.ss[s2][i] = lss.ss[s2][i-3]
		}
		if !pivot {
			lss.ss[s2][i2+1] = lss.ss[s1][i1]
			lss.ss[s2][i2+2] = lss.ss[s1][i1+1]
			lss.ss[s2][i2+3] = lss.ss[s1][i1+2]
		} else {
			lss.ss[s2][i2+1] = lss.ss[s1][i1+2]
			lss.ss[s2][i2+2] = lss.ss[s1][i1+1]
			lss.ss[s2][i2+3] = lss.ss[s1][i1]
		}

		lss.ss[s1] = append(lss.ss[s1][:i1], lss.ss[s1][i1+3:]...)

		lss.ssdur[s1] = getSeqDur(lss.ss[s1], lss.ssr[s1], lss)
		lss.ssdur[s2] = getSeqDur(lss.ss[s2], lss.ssr[s2], lss)
		lss.ssdist[s1] = getSeqDist(lss.ss[s1], lss.ssr[s1], lss)
		lss.ssdist[s2] = getSeqDist(lss.ss[s2], lss.ssr[s2], lss)

	} else {
		delta = 0
	}
	return
}

//选取route1的连续5个点，route2的连续2个点，将route1中间3个点插入route2的两个点中
func d_reloc4_sa(ss1, ss2 int, i1, i2 int, r1, r2 int, qty2 float64, dur2 float64, wei2 float64, dist2 float64, lss *LSS) (delta float64, pivot bool) {
	s1 := lss.ss[ss1]
	s2 := lss.ss[ss2]
	n1 := make([]int, 5)
	n2 := make([]int, 2)

	if i1 > 0 {
		n1[0] = lss.taskLoc[s1[i1-1]]
	}
	n1[1] = lss.taskLoc[s1[i1]]
	n1[2] = lss.taskLoc[s1[i1+1]]
	n1[3] = lss.taskLoc[s1[i1+2]]
	if i1+2 < len(s1)-1 {
		n1[4] = lss.taskLoc[s1[i1+3]]
	}

	n2[0] = lss.taskLoc[s2[i2]]
	if i2 < len(s2)-1 {
		n2[1] = lss.taskLoc[s2[i2+1]]
	}

	dist1 := getSeqDist(s1, r1, lss)
	cost1 := getSeqMapCost(r1, lss, dist1)
	cost2 := getSeqMapCost(r2, lss, dist2)
	sumCost := cost1 + cost2

	var tmpS1 = make([]int, 0)
	tmpS1 = append(tmpS1, s1[:i1]...)
	tmpS1 = append(tmpS1, s1[i1+3:]...)
	var tmpS2 = make([]int, 0)
	tmpS2 = append(tmpS2, s2[:i2+1]...)
	tmpS2 = append(tmpS2, s1[i1:i1+3]...)
	tmpS2 = append(tmpS2, s2[i2+1:]...)

	tmpDist1 := dist1 - lss.cost[r1][0][n1[0]][n1[1]] - lss.cost[r1][0][n1[1]][n1[2]] - lss.cost[r1][0][n1[2]][n1[3]] - lss.cost[r1][0][n1[3]][n1[4]] + lss.cost[r1][0][n1[0]][n1[4]]
	tmpDist2 := dist2 - lss.cost[r2][0][n2[0]][n2[1]] + lss.cost[r2][0][n2[0]][n1[1]] + lss.cost[r2][0][n1[1]][n1[2]] + lss.cost[r2][0][n1[2]][n1[3]] + lss.cost[r2][0][n1[3]][n2[1]]
	tmpCost1 := getSeqMapCost(r1, lss, tmpDist1)
	tmpCost2 := getSeqMapCost(r2, lss, tmpDist2)
	tmpSumCost := tmpCost1 + tmpCost2

	if tmpSumCost <= sumCost {
		innerD1 := getSeqInnerDist(s1, r1, lss)
		innerD2 := getSeqInnerDist(s2, r2, lss)
		//oldS1 := CopySliceInt(lss.ss[ss1])
		//oldS2 := CopySliceInt(lss.ss[ss2])
		//lss.ss[ss1] = tmpS1
		//lss.ss[ss2] = tmpS2
		tmpInnerD1 := getSeqInnerDist(tmpS1, r1, lss)
		tmpInnerD2 := getSeqInnerDist(tmpS2, r2, lss)
		//lss.ss[ss1] = oldS1
		//lss.ss[ss2] = oldS2
		delta = tmpInnerD1 + tmpInnerD2 - innerD1 - innerD2
	} else {
		delta = 100
	}

	//delta = -lss.cost[r1][0][n1[0]][n1[1]] - lss.cost[r1][0][n1[1]][n1[2]] - lss.cost[r1][0][n1[2]][n1[3]] - lss.cost[r1][0][n1[3]][n1[4]] - lss.cost[r2][0][n2[0]][n2[1]] +
	//	lss.cost[r1][0][n1[0]][n1[4]] + lss.cost[r2][0][n2[0]][n1[1]] + lss.cost[r2][0][n1[1]][n1[2]] + lss.cost[r2][0][n1[2]][n1[3]] + lss.cost[r2][0][n1[3]][n2[1]]
	//
	if delta < -1e-3 {
		if qty2+lss.tqty[s1[i1]]+lss.tqty[s1[i1+1]]+lss.tqty[s1[i1+2]] > lss.rqty[r2] {
			delta = 0
		} else if !pivot {
			if dur2+lss.tdur[s1[i1]]+lss.tdur[s1[i1+1]]+lss.tdur[s1[i1+2]]-lss.cost[r2][1][n2[0]][n2[1]]+lss.cost[r2][1][n2[0]][n1[1]]+lss.cost[r2][1][n1[1]][n1[2]]+lss.cost[r2][1][n1[2]][n1[3]]+lss.cost[r2][1][n1[3]][n2[1]] > lss.rdur[r2] {
				delta = 0
			}
		}
	}
	if delta < -1e-3 {
		if wei2+lss.twei[s1[i1]]+lss.twei[s1[i1+1]]+lss.twei[s1[i1+2]] > lss.rwei[r2] {
			delta = 0
		} else if !pivot {
			if dist2-lss.cost[r2][0][n2[0]][n2[1]]+lss.cost[r2][0][n2[0]][n1[1]]+lss.cost[r2][0][n1[1]][n1[2]]+lss.cost[r2][0][n1[2]][n1[3]]+lss.cost[r2][0][n1[3]][n2[1]] > lss.rdist[r2] {
				delta = 0
			}
		}
	}
	return
}

// **************************** 2opt ****************************************
//在指定route上选两个点，这两点间的派送方向翻转
func run_brute_2opt_sa(s1 int, lss *LSS) (delta float64) {
	for i := 0; i <= len(lss.ss[s1])-2; i++ {
		for j := i + 1; j < len(lss.ss[s1]); j++ {
			delta += try_2opt_sa(s1, i, j, lss)
		}
	}
	return
}

func run_global_5_sa(tEndUnix int64, lss *LSS) (delta_all float64) {
	var delta float64
	for i := 0; i < len(lss.ss); i++ {
		for j := 0; j < len(lss.ss); j++ {
			if time.Now().Unix() > tEndUnix-lss.Latency*2 {
				return
			}
			start := time.Now()
			for m := -1; m < len(lss.ss[i]); m++ {
				for n := -1; n < len(lss.ss[j]); n++ {
					delta = try_2opt_extra_sa(i, j, m, n, lss)
					delta_all += delta
				}
			}
			elapsed := time.Since(start)
			tmpLatency := int64(elapsed/time.Second) + 1
			if tmpLatency > lss.Latency {
				lss.Latency = tmpLatency
			}
		}
	}
	return
}

func run_global_6_sa(tEndUnix int64, lss *LSS) (delta_all float64) {
	var delta float64
	for i := 0; i < len(lss.ss); i++ {
		for j := 0; j < len(lss.ss); j++ {
			if time.Now().Unix() > tEndUnix-lss.Latency*2 {
				return
			}
			start := time.Now()
			for m := -1; m < len(lss.ss[i]); m++ {
				for n := -1; n < len(lss.ss[j]); n++ {
					delta = try_2opt_extra_2_sa(i, j, m, n, lss)
					delta_all += delta
				}
			}
			elapsed := time.Since(start)
			tmpLatency := int64(elapsed/time.Second) + 1
			if tmpLatency > lss.Latency {
				lss.Latency = tmpLatency
			}
		}
	}
	return
}

//在指定route上选两个点，这两点间的派送方向翻转
func try_2opt_sa(s1 int, i1, i2 int, lss *LSS) (delta float64) {
	delta = d_2opt_sa(lss.ss[s1], i1, i2, lss.ssr[s1], lss)
	if delta < -1e-3 {
		for p, q := i1, i2; p < q; p, q = p+1, q-1 {
			lss.ss[s1][p], lss.ss[s1][q] = lss.ss[s1][q], lss.ss[s1][p]
		}
		lss.ssdur[s1] = getSeqDur(lss.ss[s1], lss.ssr[s1], lss)
		lss.ssdist[s1] = getSeqDist(lss.ss[s1], lss.ssr[s1], lss)
	} else {
		delta = 0
	}
	return
}

//在指定route上选两个点，这两点间的派送方向翻转
func d_2opt_sa(s1 []int, i1, i2 int, r1 int, lss *LSS) (delta float64) {
	// (i-1,i) (j,j+1)
	n1 := make([]int, 2)
	n2 := make([]int, 2)
	if i1 == 0 {
		n1[0] = 0
	} else {
		n1[0] = lss.taskLoc[s1[i1-1]]
	}
	n1[1] = lss.taskLoc[s1[i1]]
	n2[0] = lss.taskLoc[s1[i2]]
	if i2 == len(s1)-1 {
		n2[1] = 0
	} else {
		n2[1] = lss.taskLoc[s1[i2+1]]
	}

	// first alss.ssume d12 = d21 for simplicity, then check full path
	delta = -lss.cost[r1][0][n1[0]][n1[1]] - lss.cost[r1][0][n2[0]][n2[1]] + lss.cost[r1][0][n1[0]][n2[0]] + lss.cost[r1][0][n1[1]][n2[1]]
	if delta < 1e8 {
		for i := i1; i < i2; i++ {
			delta += -lss.cost[r1][0][lss.taskLoc[s1[i]]][lss.taskLoc[s1[i+1]]] + lss.cost[r1][0][lss.taskLoc[s1[i+1]]][lss.taskLoc[s1[i]]]
		}
	}
	//delta = 100
	return
}

//组间交换，两条route各剪一刀，后半段全交换
func try_2opt_extra_sa(s1, s2 int, i1, i2 int, lss *LSS) (delta float64) {
	if s1 == s2 || len(lss.ss[s1]) < 2 || len(lss.ss[s2]) < 1 {
		return
	}
	// s1 (i1, i1+1)  s2 (i2, i2+1)
	//var n1, n2 []int
	//if i1 == -1 {
	//	n1 = []int{0, lss.taskLoc[lss.ss[s1][i1+1]]}
	//} else if i1 < len(lss.ss[s1])-1 {
	//	n1 = []int{lss.taskLoc[lss.ss[s1][i1]], lss.taskLoc[lss.ss[s1][i1+1]]}
	//} else {
	//	n1 = []int{lss.taskLoc[lss.ss[s1][i1]], 0}
	//}
	//if i2 == -1 {
	//	n2 = []int{0, lss.taskLoc[lss.ss[s2][i2+1]]}
	//} else if i2 < len(lss.ss[s2])-1 {
	//	n2 = []int{lss.taskLoc[lss.ss[s2][i2]], lss.taskLoc[lss.ss[s2][i2+1]]}
	//} else {
	//	n2 = []int{lss.taskLoc[lss.ss[s2][i2]], 0}
	//}
	r1, r2 := lss.ssr[s1], lss.ssr[s2]

	//innerD1 := getSeqInnerDist(lss.ss[s1], r1, lss)
	//innerD2 := getSeqInnerDist(lss.ss[s2], r2, lss)

	cost1 := getSeqMapCost(r1, lss, lss.ssdist[s1])
	cost2 := getSeqMapCost(r2, lss, lss.ssdist[s2])
	sumCost := cost1 + cost2

	s_tmp := make([]int, len(lss.ss[s1])-i1-1)
	s_tmp = CopySliceInt(lss.ss[s1][i1+1:])

	tmpS1 := append(CopySliceInt(lss.ss[s1][:i1+1]), lss.ss[s2][i2+1:]...)
	tmpS2 := append(CopySliceInt(lss.ss[s2][:i2+1]), s_tmp...)

	tmpDist1 := getSeqDist(tmpS1, r1, lss)
	tmpDist2 := getSeqDist(tmpS2, r2, lss)
	tmpCost1 := getSeqMapCost(r1, lss, tmpDist1)
	tmpCost2 := getSeqMapCost(r2, lss, tmpDist2)
	tmpSumCost := tmpCost1 + tmpCost2

	if tmpSumCost <= sumCost {
		innerD1 := getSeqInnerDist(lss.ss[s1], r1, lss)
		innerD2 := getSeqInnerDist(lss.ss[s2], r2, lss)
		//oldS1 := CopySliceInt(lss.ss[s1])
		//oldS2 := CopySliceInt(lss.ss[s2])
		//lss.ss[s1] = tmpS1
		//lss.ss[s2] = tmpS2
		tmpInnerD1 := getSeqInnerDist(tmpS1, r1, lss)
		tmpInnerD2 := getSeqInnerDist(tmpS2, r2, lss)
		//lss.ss[s1] = oldS1
		//lss.ss[s2] = oldS2
		delta = tmpInnerD1 + tmpInnerD2 - innerD1 - innerD2
	} else {
		delta = 100
	}

	//delta = -lss.cost[r1][0][n1[0]][n1[1]] - lss.cost[r2][0][n2[0]][n2[1]] + lss.cost[r1][0][n1[0]][n2[1]] + lss.cost[r2][0][n2[0]][n1[1]]

	if delta < -1e-3 {
		qty1 := getSeqQty(tmpS1, lss)
		qty2 := getSeqQty(tmpS2, lss)
		if qty1 > lss.rqty[r1] || qty2 > lss.rqty[r2] {
			delta = 0
		}
	}
	if delta < -1e-3 {
		dur1 := getSeqDur(tmpS1, lss.ssr[s1], lss)
		dur2 := getSeqDur(tmpS2, lss.ssr[s2], lss)

		if dur1 > lss.rdur[r1] || dur2 > lss.rdur[r2] {
			delta = 0
		} else {
			dist1 := getSeqDist(tmpS1, lss.ssr[s1], lss)
			dist2 := getSeqDist(tmpS2, lss.ssr[s2], lss)
			if dist1 > lss.rdist[r1] || dist2 > lss.rdist[r2] {
				delta = 0
			}
		}
	}

	if delta < -1e-3 {
		wei1 := getSeqWei(tmpS1, lss)
		wei2 := getSeqWei(tmpS2, lss)
		if wei1 > lss.rwei[r1] || wei2 > lss.rwei[r2] {
			delta = 0
		}
	}

	if delta < -1e-3 {
		lss.ss[s1] = tmpS1
		lss.ss[s2] = tmpS2

		lss.ssqty[s1] = getSeqQty(lss.ss[s1], lss)
		lss.ssqty[s2] = getSeqQty(lss.ss[s2], lss)
		lss.sswei[s1] = getSeqWei(lss.ss[s1], lss)
		lss.sswei[s2] = getSeqWei(lss.ss[s2], lss)
		lss.ssdur[s1] = getSeqDur(lss.ss[s1], lss.ssr[s1], lss)
		lss.ssdur[s2] = getSeqDur(lss.ss[s2], lss.ssr[s2], lss)
		lss.ssdist[s1] = getSeqDist(lss.ss[s1], lss.ssr[s1], lss)
		lss.ssdist[s2] = getSeqDist(lss.ss[s2], lss.ssr[s2], lss)
	} else {
		delta = 0
	}
	return
}

//组间交换，两条route各剪一刀，route1前半段接route2前半段翻转(按route1的车)，route1后半段翻转接route2后半段(按route2的车)
func try_2opt_extra_2_sa(s1, s2 int, i1, i2 int, lss *LSS) (delta float64) {
	if s1 == s2 || i1+1 > len(lss.ss[s1]) || i2+1 > len(lss.ss[s2]) || len(lss.ss[s1]) < 1 || len(lss.ss[s2]) < 1 {
		return
	}
	// s1 (i1, i1+1)  s2 (i2, i2+1)  -- (i1, i2) and (i1+1, i2+1)
	//var n1, n2 []int
	//if i1 == -1 {
	//	n1 = []int{0, lss.taskLoc[lss.ss[s1][0]]}
	//} else if i1 < len(lss.ss[s1])-1 {
	//	n1 = []int{lss.taskLoc[lss.ss[s1][i1]], lss.taskLoc[lss.ss[s1][i1+1]]}
	//} else {
	//	n1 = []int{lss.taskLoc[lss.ss[s1][i1]], 0}
	//}
	//if i2 == -1 {
	//	n2 = []int{0, lss.taskLoc[lss.ss[s2][0]]}
	//} else if i2 < len(lss.ss[s2])-1 {
	//	n2 = []int{lss.taskLoc[lss.ss[s2][i2]], lss.taskLoc[lss.ss[s2][i2+1]]}
	//} else {
	//	n2 = []int{lss.taskLoc[lss.ss[s2][i2]], 0}
	//}

	r1, r2 := lss.ssr[s1], lss.ssr[s2]

	cost1 := getSeqMapCost(r1, lss, lss.ssdist[s1])
	cost2 := getSeqMapCost(r2, lss, lss.ssdist[s2])
	sumCost := cost1 + cost2

	tmpS1 := make([]int, 0, i1+i2+2)
	tmpS2 := make([]int, 0, len(lss.ss[s1])+len(lss.ss[s2])-i1-i2-2)

	for i := 0; i <= i1; i++ {
		tmpS1 = append(tmpS1, lss.ss[s1][i])
	}
	for i := i2; i >= 0; i-- {
		tmpS1 = append(tmpS1, lss.ss[s2][i])
	}
	for i := len(lss.ss[s1]) - 1; i >= i1+1; i-- {
		tmpS2 = append(tmpS2, lss.ss[s1][i])
	}
	for i := i2 + 1; i < len(lss.ss[s2]); i++ {
		tmpS2 = append(tmpS2, lss.ss[s2][i])
	}

	tmpDist1 := getSeqDist(tmpS1, r1, lss)
	tmpDist2 := getSeqDist(tmpS2, r2, lss)
	tmpCost1 := getSeqMapCost(r1, lss, tmpDist1)
	tmpCost2 := getSeqMapCost(r2, lss, tmpDist2)
	tmpSumCost := tmpCost1 + tmpCost2

	if tmpSumCost <= sumCost {
		innerD1 := getSeqInnerDist(lss.ss[s1], r1, lss)
		innerD2 := getSeqInnerDist(lss.ss[s2], r2, lss)
		//oldS1 := CopySliceInt(lss.ss[s1])
		//oldS2 := CopySliceInt(lss.ss[s2])
		//lss.ss[s1] = tmpS1
		//lss.ss[s2] = tmpS2
		tmpInnerD1 := getSeqInnerDist(tmpS1, r1, lss)
		tmpInnerD2 := getSeqInnerDist(tmpS2, r2, lss)
		//lss.ss[s1] = oldS1
		//lss.ss[s2] = oldS2
		delta = tmpInnerD1 + tmpInnerD2 - innerD1 - innerD2
	} else {
		delta = 100
	}
	//delta = -lss.cost[r1][0][n1[0]][n1[1]] - lss.cost[r2][0][n2[0]][n2[1]] + lss.cost[r1][0][n1[0]][n2[0]] + lss.cost[r2][0][n1[1]][n2[1]]
	// s2[0-i2] reverse
	//for i := 0; i < i2; i++ {
	//	delta += -lss.cost[r2][0][lss.taskLoc[lss.ss[s2][i]]][lss.taskLoc[lss.ss[s2][i+1]]] + lss.cost[r1][0][lss.taskLoc[lss.ss[s2][i+1]]][lss.taskLoc[lss.ss[s2][i]]]
	//}
	//if i2 == -1 {
	//	delta += -lss.cost[r2][0][0][n2[0]] + lss.cost[r1][0][n2[0]][0]
	//} else {
	//	delta += -lss.cost[r2][0][0][lss.taskLoc[lss.ss[s2][0]]] + lss.cost[r1][0][lss.taskLoc[lss.ss[s2][0]]][0]
	//}

	// s1[i1+1,...](0) reverse
	//for i := i1 + 1; i < len(lss.ss[s1])-1; i++ {
	//	delta += -lss.cost[r1][0][lss.taskLoc[lss.ss[s1][i]]][lss.taskLoc[lss.ss[s1][i+1]]] + lss.cost[r2][0][lss.taskLoc[lss.ss[s1][i+1]]][lss.taskLoc[lss.ss[s1][i]]]
	//}
	//if i1 == len(lss.ss[s1])-1 {
	//	delta += -lss.cost[r1][0][n1[1]][0] + lss.cost[r2][0][0][n1[1]]
	//} else {
	//	delta += -lss.cost[r1][0][lss.taskLoc[lss.ss[s1][len(lss.ss[s1])-1]]][0] + lss.cost[r2][0][0][lss.taskLoc[lss.ss[s1][len(lss.ss[s1])-1]]]
	//}

	// use copy to check duration
	//var s_tmp_1, s_tmp_2 []int
	var dur1, dur2, dist1, dist2 float64
	if delta < -1e-3 {
		//s_tmp_1 = make([]int, 0, i1+i2+2)
		//s_tmp_2 = make([]int, 0, len(lss.ss[s1])+len(lss.ss[s2])-i1-i2-2)
		//
		//for i := 0; i <= i1; i++ {
		//	s_tmp_1 = append(s_tmp_1, lss.ss[s1][i])
		//}
		//for i := i2; i >= 0; i-- {
		//	s_tmp_1 = append(s_tmp_1, lss.ss[s2][i])
		//}
		//for i := len(lss.ss[s1]) - 1; i >= i1+1; i-- {
		//	s_tmp_2 = append(s_tmp_2, lss.ss[s1][i])
		//}
		//for i := i2 + 1; i < len(lss.ss[s2]); i++ {
		//	s_tmp_2 = append(s_tmp_2, lss.ss[s2][i])
		//}

		dur1 = getSeqDur(tmpS1, r1, lss)
		dur2 = getSeqDur(tmpS2, r2, lss)

		if dur1 > lss.rdur[r1] || dur2 > lss.rdur[r2] {
			delta = 0
		} else {
			dist1 = getSeqDist(tmpS1, r1, lss)
			dist2 = getSeqDist(tmpS2, r2, lss)
			if dist1 > lss.rdist[r1] || dist2 > lss.rdist[r2] {
				delta = 0
			}
		}
	}

	if delta < -1e-3 {
		wei1 := getSeqWei(lss.ss[s1][:i1+1], lss) + getSeqWei(lss.ss[s2][:i2+1], lss)
		wei2 := getSeqWei(lss.ss[s1][i1+1:], lss) + getSeqWei(lss.ss[s2][i2+1:], lss)
		if wei1 > lss.rwei[r1] || wei2 > lss.rwei[r2] {
			delta = 0
		}
	}

	if delta < -1e-3 {
		qty1 := getSeqQty(lss.ss[s1][:i1+1], lss) + getSeqQty(lss.ss[s2][:i2+1], lss)
		qty2 := getSeqQty(lss.ss[s1][i1+1:], lss) + getSeqQty(lss.ss[s2][i2+1:], lss)
		if qty1 > lss.rqty[r1] || qty2 > lss.rqty[r2] {
			delta = 0
		}
	}

	if delta < -1e-3 {
		lss.ss[s1] = lss.ss[s1][:0]
		for i := 0; i < len(tmpS1); i++ {
			lss.ss[s1] = append(lss.ss[s1], tmpS1[i])
		}
		lss.ss[s2] = lss.ss[s2][:0]
		for i := 0; i < len(tmpS2); i++ {
			lss.ss[s2] = append(lss.ss[s2], tmpS2[i])
		}
		lss.ssdur[s1] = dur1
		lss.ssdur[s2] = dur2
		lss.ssdist[s1] = dist1
		lss.ssdist[s2] = dist2
		lss.ssqty[s1] = getSeqQty(lss.ss[s1], lss)
		lss.ssqty[s2] = getSeqQty(lss.ss[s2], lss)
		lss.sswei[s1] = getSeqWei(tmpS1, lss)
		lss.sswei[s2] = getSeqWei(tmpS2, lss)
	} else {
		delta = 0
	}

	return
}
