package solver

import (
	"math/rand"
	"time"
)

type LSS struct {
	tqty        []float64
	tdur        []float64
	twei        []float64
	ss          [][]int
	ssr         []int
	ssqty       []float64
	ssdur       []float64
	sswei       []float64
	ssdist      []float64
	ssinnerdist []float64 //为 p-shape adjust
	ssmapcost   []float64 //为 p-shape adjust
	rqty        []float64
	rdur        []float64
	rwei        []float64
	rdist       []float64 //车辆max distance
	taskLoc     []int     //task所属nodeIdx
	cost        [][][][]float64
	capRes      [][]float64
	capTask     [][]float64
	CapResCost  [][][]float64
	Latency     int64 //RSO 一个算子第二层一次循环最大用时
}

func LocalSearch(tEndUnix int64, gPara *GPara, gState *GState) {
	if time.Now().Unix() > tEndUnix-5 {
		return
	}

	var lss = &LSS{}
	cpVarLS(lss, gPara, gState)

	//if getConstr(lss) > 0 {
	//	fmt.Println("local search input invalid")
	//	return
	//}

	type Sol struct {
		ss      [][]int
		dist    float64
		cntUn   int
		unTasks []int
	}

	sol0 := Sol{}
	sol0.ss = make([][]int, len(lss.ss))
	for i := 0; i < len(lss.ss); i++ {
		sol0.ss[i] = make([]int, len(lss.ss[i]))
		for j := 0; j < len(lss.ss[i]); j++ {
			sol0.ss[i][j] = lss.ss[i][j]
		}
	}
	sol0.dist = getSeqsDist(lss.ss, lss)
	sol0.unTasks = CopySliceInt(gState.InnerUnasgTasks)
	sol0.cntUn = len(sol0.unTasks)

	var iterLS int = 100
	var iterLNS int = 3000
	var ratioLNS float64 = 0.4

	sol := Sol{}
	sol.ss = make([][]int, len(lss.ss))
	for i := 0; i < len(lss.ss); i++ {
		sol.ss[i] = make([]int, len(lss.ss[i]), cap(lss.ss[i]))
		for j := 0; j < len(lss.ss[i]); j++ {
			sol.ss[i][j] = lss.ss[i][j]
		}
	}
	sol.dist = getSeqsDist(lss.ss, lss)
	sol.unTasks = CopySliceInt(gState.InnerUnasgTasks)
	sol.cntUn = len(sol.unTasks)
	// 这里为止sol0 = sol = lss.ss
	RSO(tEndUnix, lss)
	// lss.ss修改
	if getConstr(lss) > 0 {
		copyTo(sol.ss, lss.ss, nil, nil)
	} else {
		copyTo(lss.ss, sol.ss, nil, nil)
		sol.dist = getSeqsDist(lss.ss, lss)
	}
	//sol = lss.ss

	var d1, d2 float64
	// for 修改lss.ss
	var loopStartTime, loopEndTime int64
	if gState.LSLoopTime < 1 {
		gState.LSLoopTime = 1
	}
	if time.Now().Unix() > tEndUnix-lss.Latency*2 {
		gState.IsTimeEnough = false
	}
	for iter := 0; iter < iterLS && time.Now().Unix() < tEndUnix-5 && gState.IsTimeEnough; iter++ {
		loopStartTime = time.Now().Unix()
		d1 = LNS(tEndUnix, iterLNS, ratioLNS, lss)
		if getConstr(lss) > 0 {
			copyTo(sol.ss, lss.ss, nil, nil)
		} else {
			copyTo(lss.ss, sol.ss, nil, nil)
			sol.dist = getSeqsDist(lss.ss, lss)
		}
		d2 = RSO(tEndUnix, lss)
		if getConstr(lss) > 0 {
			copyTo(sol.ss, lss.ss, nil, nil)
		} else {
			copyTo(lss.ss, sol.ss, nil, nil)
			sol.dist = getSeqsDist(lss.ss, lss)
		}

		if sol.cntUn > 0 {
			unTasks := Reassign(tEndUnix, sol.unTasks, lss)
			if getConstr(lss) > 0 {
				copyTo(sol.ss, lss.ss, nil, nil)
			} else {
				copyTo(lss.ss, sol.ss, nil, nil)
				sol.dist = getSeqsDist(lss.ss, lss)
				sol.unTasks = unTasks
				sol.cntUn = len(sol.unTasks)
			}
		}

		if d1 > -1e-3 && d2 > -1e-3 {
			d1 = LNS(tEndUnix, iterLNS*2, 0.5, lss)
			if getConstr(lss) > 0 {
				copyTo(sol.ss, lss.ss, nil, nil)
			} else {
				copyTo(lss.ss, sol.ss, nil, nil)
				sol.dist = getSeqsDist(lss.ss, lss)
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
		//fmt.Printf("LS1-1: iter:%v   tmpLoopTime:%v    loopMaxTime:%v \n", iter, loopEndTime-loopStartTime, gState.LSLoopTime)
	}

	sol1 := Sol{}
	sol1.ss = copyMatrixI(sol.ss)
	sol1.dist = sol.dist
	sol1.cntUn = sol.cntUn
	sol1.unTasks = CopySliceInt(sol.unTasks)
	// sol1 = sol != lss.ss

	copyTo(sol0.ss, lss.ss, nil, nil)
	// lss.ss = sol0
	sol.dist = sol0.dist
	sol.unTasks = CopySliceInt(gState.InnerUnasgTasks)
	sol.cntUn = len(sol.unTasks)
	//sol = sol0 = lss.ss
	for iter := 0; iter < iterLS && time.Now().Unix() < tEndUnix-5 && gState.IsTimeEnough; iter++ {
		loopStartTime = time.Now().Unix()
		d1 = LNS(tEndUnix, iterLNS, ratioLNS, lss)
		if getConstr(lss) > 0 {
			copyTo(sol.ss, lss.ss, nil, nil)
		} else {
			copyTo(lss.ss, sol.ss, nil, nil)
			sol.dist = getSeqsDist(lss.ss, lss)
		}
		d2 = RSO(tEndUnix, lss)
		if getConstr(lss) > 0 {
			copyTo(sol.ss, lss.ss, nil, nil)
		} else {
			copyTo(lss.ss, sol.ss, nil, nil)
			sol.dist = getSeqsDist(lss.ss, lss)
		}
		if sol.cntUn > 0 {
			unTasks := Reassign(tEndUnix, sol.unTasks, lss)
			if getConstr(lss) > 0 {
				copyTo(sol.ss, lss.ss, nil, nil)
			} else {
				copyTo(lss.ss, sol.ss, nil, nil)
				sol.dist = getSeqsDist(lss.ss, lss)
				sol.unTasks = unTasks
				sol.cntUn = len(sol.unTasks)
			}
		}
		if d1 > -1e-3 && d2 > -1e-3 {
			d1 = LNS(tEndUnix, iterLNS*2, 0.5, lss)
			if getConstr(lss) > 0 {
				copyTo(sol.ss, lss.ss, nil, nil)
			} else {
				copyTo(lss.ss, sol.ss, nil, nil)
				sol.dist = getSeqsDist(lss.ss, lss)
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
		//fmt.Printf("LS1-2: iter:%v   tmpLoopTime:%v    loopMaxTime:%v \n", iter, loopEndTime-loopStartTime, gState.LSLoopTime)
	}
	// 修改lss.ss sol = lss.ss != sol0

	if sol1.cntUn == sol.cntUn {
		if sol1.dist < sol.dist {
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
	for i := 0; i < len(gState.InnerSeqDtls); i++ {
		tmpFCost, tmpMCost := GetSeqMFCostByDist(gPara, gState.InnerAsgmts[i], gState.InnerSeqDtls[i][0])
		gState.InnerSeqDtls[i] = append(gState.InnerSeqDtls[i], tmpFCost)
		gState.InnerSeqDtls[i] = append(gState.InnerSeqDtls[i], tmpMCost)
	}
}

type SSFeat struct {
	dur       []float64
	qty       []float64
	dist      []float64
	mapCost   []float64
	innerDist []float64
}

func LNS(tEndUnix int64, iterLNS int, ratioLNS float64, lss *LSS) (delta float64) {
	if time.Now().Unix() > tEndUnix-5 {
		return
	}

	if len(lss.ss) < 1 || len(lss.tqty) > len(lss.cost[0][0])-1 {
		return
	}

	var d0, d1 float64

	dist0 := getSeqsDist(lss.ss, lss)

	ssbest := copyFull(lss)
	fbest := &SSFeat{}
	fbest.dist = getSeqsDistList(lss.ss, lss)
	fbest.dur = getSeqsDurList(lss.ss, lss)
	fbest.qty = getSeqsQtyList(lss.ss, lss)

	ssnew := make([][]int, len(ssbest))
	for i := 0; i < len(ssnew); i++ {
		ssnew[i] = make([]int, 0, cap(ssbest[i]))
	}
	f := &SSFeat{}
	f.dist = make([]float64, len(fbest.dist))
	f.dur = make([]float64, len(fbest.dist))
	f.qty = make([]float64, len(fbest.dist))

	// copy
	copyTo(ssbest, ssnew, fbest, f)

	var r, pos int
	var c, cbest float64
	var success bool
	var r1, r2 int
	for iter := 0; iter < iterLNS && time.Now().Unix() < tEndUnix-5; iter++ {
		success = true
		d0 = getSeqsDistNew(ssnew, lss)

		var remain []int
		if len(lss.ss) > 1 {
			r1 = rand.Intn(len(lss.ss))
			r2tmp := rand.Intn(len(lss.ss) - 1)
			if r2tmp < r1 {
				r2 = r2tmp
			} else {
				r2 = r2tmp + 1
			}
			//随机选两个route
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
			}
		}

		cbest = 1e8
		for _, t := range remain { //贪心找每个需要插入的点插入的最佳位置
			r, pos = -1, -1
			c, cbest = 0, 1e8
			for i := 0; i < len(ssnew); i++ {
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
			}
			if r >= 0 {
				f.dist[r] += -lss.cost[lss.ssr[r]][0][ssnew[r][pos]][ssnew[r][pos+1]] + lss.cost[lss.ssr[r]][0][ssnew[r][pos]][t] + lss.cost[lss.ssr[r]][0][t][ssnew[r][pos+1]] //+= cbest
				f.qty[r] += lss.tqty[t-1]
				f.dur[r] += lss.tdur[t-1] - lss.cost[lss.ssr[r]][1][ssnew[r][pos]][ssnew[r][pos+1]] + lss.cost[lss.ssr[r]][1][ssnew[r][pos]][t] + lss.cost[lss.ssr[r]][1][t][ssnew[r][pos+1]]
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

		d1 = getSeqsDistNew(ssnew, lss)
		if d1 <= d0 {
			copyTo(ssnew, ssbest, f, fbest)
		} else {
			copyTo(ssbest, ssnew, fbest, f)
		}
	}

	dist1 := getSeqsDistNew(ssbest, lss)
	if dist1 < dist0 {
		copyPartBack(ssbest, lss.ss)
		delta = dist1 - dist0
	}
	return
}

//尝试解决orphan的算子
func Reassign(tEndUnix int64, tasks []int, lss *LSS) (unTasks []int) {
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
	//加更新seqdtl
	return
}

func RSO(tEndUnix int64, lss *LSS) (delta float64) {
	if time.Now().Unix() > tEndUnix-5 {
		return
	}
	lss.Latency = 5
	d0 := getSeqsDist(lss.ss, lss)
	lss.ssqty = getSeqsQtyList(lss.ss, lss)
	lss.ssdur = getSeqsDurList(lss.ss, lss)
	lss.ssdist = getSeqsDistList(lss.ss, lss)

	var d float64
	d += run_global_4(tEndUnix, lss)
	d += run_global_3(tEndUnix, lss)
	d += run_global_2(tEndUnix, lss)
	d += run_global_5(tEndUnix, lss)
	if d < 0 {
		d += run_global_5(tEndUnix, lss)
	}
	d += run_local_1(tEndUnix, lss)
	if d < 0 {
		d += run_global_4(tEndUnix, lss)
	}
	if d < 0 {
		d += run_local_1(tEndUnix, lss)
	}
	if d < 0 {
		d += run_global_2(tEndUnix, lss)
	}
	if d < 0 {
		d += run_local_1(tEndUnix, lss)
	}
	if d < 0 {
		d += run_global_3(tEndUnix, lss)
	}
	if d < 0 {
		d += run_local_1(tEndUnix, lss)
	}

	d += run_global_6(tEndUnix, lss)

	if d < 0 {
		d += run_local_1(tEndUnix, lss)
	}

	if d < 0 {
		d += run_global_2(tEndUnix, lss)
	}
	if d < 0 {
		d += run_local_1(tEndUnix, lss)
	}
	if d < 0 {
		d += run_global_3(tEndUnix, lss)
	}
	if d < 0 {
		d += run_local_1(tEndUnix, lss)
	}
	if d < 0 {
		d += run_global_4(tEndUnix, lss)
	}
	if d < 0 {
		d += run_local_1(tEndUnix, lss)
	}
	if d < 0 {
		d += run_global_5(tEndUnix, lss)
	}
	if d < 0 {
		d += run_local_1(tEndUnix, lss)
	}
	if d < 0 {
		d += run_global_6(tEndUnix, lss)
	}
	if d < 0 {
		d += run_local_1(tEndUnix, lss)
	}

	d += run_global_1(tEndUnix, lss)

	if d < 0 {
		d += run_global_5(tEndUnix, lss)
	}
	if d < 0 {
		d += run_global_4(tEndUnix, lss)
	}
	if d < 0 {
		d += run_global_3(tEndUnix, lss)
	}
	if d < 0 {
		d += run_global_2(tEndUnix, lss)
	}
	if d < 0 {
		d += run_local_1(tEndUnix, lss)
	}

	if d < 0 {
		var d1 float64
		for iter := 0; iter < 10; iter++ {
			d1 = run_global_5(tEndUnix, lss) + run_local_1(tEndUnix, lss)
			d1 += run_global_2(tEndUnix, lss) + run_local_1(tEndUnix, lss)
			d1 += run_global_3(tEndUnix, lss) + run_local_1(tEndUnix, lss)
			d1 += run_global_4(tEndUnix, lss) + run_local_1(tEndUnix, lss)
			d1 += run_global_1(tEndUnix, lss) + run_local_1(tEndUnix, lss)
			d += d1
			if d1 > -1e-3 || time.Now().Unix() > tEndUnix-5 {
				break
			}
		}
	}

	if d < 0 {
		d += run_global_6(tEndUnix, lss)
	}
	if d < 0 {
		d += run_local_1(tEndUnix, lss)
	}
	if d < 0 {
		d += run_global_1(tEndUnix, lss)
	}
	if d < 0 {
		d += run_global_5(tEndUnix, lss)
	}
	if d < 0 {
		d += run_global_4(tEndUnix, lss)
	}
	if d < 0 {
		d += run_global_3(tEndUnix, lss)
	}
	if d < 0 {
		d += run_global_2(tEndUnix, lss)
	}

	d += run_local_2(tEndUnix, lss)

	if d < 0 {
		d += run_global_5(tEndUnix, lss) + run_local_1(tEndUnix, lss)
	}
	if d < 0 {
		d += run_global_2(tEndUnix, lss) + run_local_1(tEndUnix, lss)
	}
	if d < 0 {
		d += run_global_3(tEndUnix, lss) + run_local_1(tEndUnix, lss)
	}
	if d < 0 {
		d += run_global_4(tEndUnix, lss) + run_local_1(tEndUnix, lss)
	}
	if d < 0 {
		d += run_global_1(tEndUnix, lss) + run_local_1(tEndUnix, lss)
	}

	d += run_global_8(tEndUnix, lss)
	if d < 0 {
		d += run_local_2(tEndUnix, lss)
	}
	d += run_global_7(tEndUnix, lss)
	if d < 0 {
		d += run_local_2(tEndUnix, lss)
	}

	if d < 0 {
		d += run_global_5(tEndUnix, lss)
	}
	if d < 0 {
		d += run_local_2(tEndUnix, lss)
	}
	if d < 0 {
		d += run_global_2(tEndUnix, lss)
	}
	if d < 0 {
		d += run_local_2(tEndUnix, lss)
	}
	if d < 0 {
		d += run_global_3(tEndUnix, lss)
	}
	if d < 0 {
		d += run_local_2(tEndUnix, lss)
	}
	if d < 0 {
		d += run_global_4(tEndUnix, lss)
	}
	if d < 0 {
		d += run_local_2(tEndUnix, lss)
	}
	if d < 0 {
		d += run_global_1(tEndUnix, lss)
	}
	if d < 0 {
		d += run_local_2(tEndUnix, lss)
	}

	delta = getSeqsDist(lss.ss, lss) - d0
	return
}

// **************************** local ****************************************
func run_local_1(tEndUnix int64, lss *LSS) (delta_all float64) {
	var delta1, delta2, delta3 float64
	for i := 0; i < len(lss.ss); i++ {
		for iter := 0; iter < 10; iter++ {
			if time.Now().Unix() > tEndUnix-lss.Latency*2 {
				return
			}
			start := time.Now()
			delta1 = run_brute_2opt(i, lss)
			delta2 = run_brute_reloc1(i, lss)
			delta3 = run_brute_swap1(i, lss)
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

func run_local_2(tEndUnix int64, lss *LSS) (delta_all float64) {
	var delta1, delta2, delta3, delta4 float64
	for i := 0; i < len(lss.ss); i++ {
		for iter := 0; iter < 10; iter++ {
			if time.Now().Unix() > tEndUnix-lss.Latency*2 {
				return
			}
			start := time.Now()
			delta1 = run_brute_2opt(i, lss)
			delta2 = run_brute_reloc1(i, lss)
			delta3 = run_brute_swap1(i, lss)
			delta4 = run_brute_swap12(i, lss)
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
func run_global_1(tEndUnix int64, lss *LSS) (delta_all float64) {
	var delta float64
	d_beg := getSeqsDist(lss.ss, lss)
	for iter := 0; iter < 2; iter++ {
		if time.Now().Unix() > tEndUnix-5 {
			return
		}
		delta = run_global_swap1(tEndUnix, lss)
		if delta < -1e-3 {
			delta_all += delta
		} else {
			break
		}
	}
	delta_all = getSeqsDist(lss.ss, lss) - d_beg
	return
}

//组内交换，指定route里某两点交换位置
func run_global_swap1(tEndUnix int64, lss *LSS) (delta float64) {
	for i := 0; i < len(lss.ss); i++ {
		if time.Now().Unix() > tEndUnix-5 {
			return
		}
		run_brute_swap1(i, lss)
	}
	return
}

func run_global_7(tEndUnix int64, lss *LSS) (delta_all float64) {
	var delta float64
	for i := 0; i < len(lss.ss); i++ {
		for j := i + 1; j < len(lss.ss); j++ {
			if time.Now().Unix() > tEndUnix-lss.Latency*2 {
				return
			}
			start := time.Now()
			for m := 0; m < len(lss.ss[i]); m++ {
				for n := 0; n < len(lss.ss[j]); n++ {
					delta = try_swap2(i, j, lss.ssr[i], lss.ssr[j], m, n, lss)
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

func run_global_8(tEndUnix int64, lss *LSS) (delta_all float64) {
	var delta float64
	for i := 0; i < len(lss.ss); i++ {
		for j := i + 1; j < len(lss.ss); j++ {
			if time.Now().Unix() > tEndUnix-lss.Latency*2 {
				return
			}
			start := time.Now()
			for m := 0; m < len(lss.ss[i])-1; m++ {
				for n := 0; n < len(lss.ss[j])-1; n++ {
					delta = try_swap22(i, j, lss.ssr[i], lss.ssr[j], m, n, lss)
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
func run_brute_swap1(s1 int, lss *LSS) (delta float64) {
	for i := 0; i < len(lss.ss[s1])-1; i++ {
		for j := i + 1; j < len(lss.ss[s1]); j++ {
			delta += try_swap1(s1, i, j, lss)
		}
	}
	return
}

//组间交换，route1的连续两个点和route2的一个点交换位置
func run_brute_swap12(s1 int, lss *LSS) (delta float64) {
	for i := 0; i < len(lss.ss[s1])-2; i++ {
		for j := i + 2; j < len(lss.ss[s1]); j++ {
			delta += try_swap12(s1, i, j, lss)
		}
	}
	return
}

//组内交换，指定route里某两点交换位置
func try_swap1(s1 int, i1, i2 int, lss *LSS) (delta float64) {
	delta = d_swap1(lss.ss[s1], i1, i2, lss.ssr[s1], lss)
	if delta < -1e-3 {
		lss.ss[s1][i1], lss.ss[s1][i2] = lss.ss[s1][i2], lss.ss[s1][i1]
		lss.ssdur[s1] = getSeqDur(lss.ss[s1], lss.ssr[s1], lss)
		lss.ssdist[s1] = getSeqDist(lss.ss[s1], lss.ssr[s1], lss)
	} else {
		delta = 0
	}
	return
}

func d_swap1(s1 []int, i1, i2 int, r1 int, lss *LSS) (delta float64) {
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
	return
}

//组间交换，route1的连续两个点和route2的一个点交换位置
func try_swap12(s1 int, i1, i2 int, lss *LSS) (delta float64) {
	delta = d_swap12(lss.ss[s1], i1, i2, lss.ssr[s1], lss)
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
func d_swap12(s1 []int, i1, i2 int, r1 int, lss *LSS) (delta float64) {
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
	}
	return
}

//组间交换，两条route中各任一点交换位置
func try_swap2(s1, s2 int, r1 int, r2 int, i1, i2 int, lss *LSS) (delta float64) {
	delta = d_swap2(s1, s2, i1, i2, r1, r2, lss)
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

//组内交换，两条route中各任一点交换位置
func d_swap2(s1, s2, i1, i2, r1, r2 int, lss *LSS) (delta float64) {
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

	delta = -lss.cost[r1][0][n1[0]][n1[1]] - lss.cost[r1][0][n1[1]][n1[2]] + lss.cost[r1][0][n1[0]][n2[1]] + lss.cost[r1][0][n2[1]][n1[2]] -
		lss.cost[r2][0][n2[0]][n2[1]] - lss.cost[r2][0][n2[1]][n2[2]] + lss.cost[r2][0][n2[0]][n1[1]] + lss.cost[r2][0][n1[1]][n2[2]]

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
func try_swap22(s1, s2 int, r1 int, r2 int, i1, i2 int, lss *LSS) (delta float64) {
	// i1, i1+1, i2, i2+1
	if len(lss.ss[s1]) < 2 || len(lss.ss[s2]) < 2 {
		return
	}
	delta = d_swap22(s1, s2, i1, i2, r1, r2, lss)
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
func d_swap22(s1, s2, i1, i2, r1, r2 int, lss *LSS) (delta float64) {
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

	delta = -lss.cost[r1][0][n1[0]][n1[1]] - lss.cost[r1][0][n1[2]][n1[3]] + lss.cost[r1][0][n1[0]][n2[1]] + lss.cost[r1][0][n2[2]][n1[3]] - lss.cost[r2][0][n2[0]][n2[1]] - lss.cost[r2][0][n2[2]][n2[3]] + lss.cost[r2][0][n2[0]][n1[1]] + lss.cost[r2][0][n1[2]][n2[3]]

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
func run_brute_reloc1(s1 int, lss *LSS) (delta float64) {
	for i := 0; i < len(lss.ss[s1]); i++ {
		for j := 0; j < len(lss.ss[s1]); j++ {
			delta += try_reloc1(s1, i, j, lss)
		}
	}
	return
}

func run_global_2(tEndUnix int64, lss *LSS) (delta_all float64) {
	var delta float64
	for i := 0; i < len(lss.ss); i++ {
		for j := 0; j < len(lss.ss); j++ {
			if time.Now().Unix() > tEndUnix-lss.Latency*2 {
				return
			}
			start := time.Now()
			for m := 0; m < len(lss.ss[i]); m++ {
				for n := 0; n < len(lss.ss[j]); n++ {
					delta = try_reloc2(i, j, m, n, lss)
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

func run_global_3(tEndUnix int64, lss *LSS) (delta_all float64) {
	var delta float64
	for i := 0; i < len(lss.ss); i++ {
		for j := 0; j < len(lss.ss); j++ {
			if time.Now().Unix() > tEndUnix-lss.Latency*2 {
				return
			}
			start := time.Now()
			for m := 0; m < len(lss.ss[i]); m++ {
				for n := 0; n < len(lss.ss[j]); n++ {
					delta = try_reloc3(i, j, m, n, lss)
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

func run_global_4(tEndUnix int64, lss *LSS) (delta_all float64) {
	var delta float64
	for i := 0; i < len(lss.ss); i++ {
		for j := 0; j < len(lss.ss); j++ {
			if time.Now().Unix() > tEndUnix-lss.Latency*2 {
				return
			}
			start := time.Now()
			for m := 0; m < len(lss.ss[i]); m++ {
				for n := 0; n < len(lss.ss[j]); n++ {
					delta = try_reloc4(i, j, m, n, lss)
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
func try_reloc1(s1 int, i1, i2 int, lss *LSS) (delta float64) {
	if len(lss.ss[s1]) < 3 {
		return 0
	}
	if i1 == i2 || i1 == i2+1 {
		return 0
	}
	delta = d_reloc1(lss.ss[s1], i1, i2, lss.ssr[s1], lss)

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
func d_reloc1(s1 []int, i1, i2 int, r1 int, lss *LSS) (delta float64) {
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

	return
}

//组间交换，一个route的1个点插入另一个route中
func try_reloc2(s1, s2 int, i1, i2 int, lss *LSS) (delta float64) {
	if s1 == s2 || i1 >= len(lss.ss[s1]) || i2 >= len(lss.ss[s2]) {
		return 0
	}

	delta = d_reloc2(lss.ss[s1], lss.ss[s2], i1, i2, lss.ssr[s1], lss.ssr[s2], lss.ssqty[s2], lss.ssdur[s2], lss.sswei[s2], lss.ssdist[s2], lss)

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
func d_reloc2(s1, s2 []int, i1, i2 int, r1, r2 int, qty2 float64, dur2 float64, wei2 float64, dist2 float64, lss *LSS) (delta float64) {
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

	delta = -lss.cost[r1][0][n1[0]][n1[1]] - lss.cost[r1][0][n1[1]][n1[2]] - lss.cost[r2][0][n2[0]][n2[1]] +
		lss.cost[r1][0][n1[0]][n1[2]] + lss.cost[r2][0][n2[0]][n1[1]] + lss.cost[r2][0][n1[1]][n2[1]]

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
func try_reloc3(s1, s2 int, i1, i2 int, lss *LSS) (delta float64) {
	if s1 == s2 || i1+1 >= len(lss.ss[s1]) || i2 >= len(lss.ss[s2]) {
		return 0
	}

	var pivot bool
	delta, pivot = d_reloc3(lss.ss[s1], lss.ss[s2], i1, i2, lss.ssr[s1], lss.ssr[s2], lss.ssqty[s2], lss.ssdur[s2], lss.sswei[s2], lss.ssdist[s2], lss)

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
func d_reloc3(s1, s2 []int, i1, i2 int, r1, r2 int, qty2 float64, dur2 float64, wei2 float64, dist2 float64, lss *LSS) (delta float64, pivot bool) {
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

	delta = -lss.cost[r1][0][n1[0]][n1[1]] - lss.cost[r1][0][n1[2]][n1[3]] - lss.cost[r2][0][n2[0]][n2[1]] +
		lss.cost[r1][0][n1[0]][n1[3]] + lss.cost[r2][0][n2[0]][n1[1]] + lss.cost[r2][0][n1[2]][n2[1]]

	if delta > -1e-3 {
		delta = -lss.cost[r1][0][n1[0]][n1[1]] - lss.cost[r1][0][n1[1]][n1[2]] - lss.cost[r1][0][n1[2]][n1[3]] - lss.cost[r2][0][n2[0]][n2[1]] +
			lss.cost[r1][0][n1[0]][n1[3]] + lss.cost[r2][0][n2[0]][n1[2]] + lss.cost[r2][0][n1[2]][n1[1]] + lss.cost[r2][0][n1[1]][n2[1]]
		pivot = true
	}

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
func try_reloc4(s1, s2 int, i1, i2 int, lss *LSS) (delta float64) {
	if s1 == s2 || i1+2 >= len(lss.ss[s1]) || i2 >= len(lss.ss[s2]) {
		return 0
	}
	var pivot bool
	delta, pivot = d_reloc4(lss.ss[s1], lss.ss[s2], i1, i2, lss.ssr[s1], lss.ssr[s2], lss.ssqty[s2], lss.ssdur[s2], lss.sswei[s2], lss.ssdist[s2], lss)

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
func d_reloc4(s1, s2 []int, i1, i2 int, r1, r2 int, qty2 float64, dur2 float64, wei2 float64, dist2 float64, lss *LSS) (delta float64, pivot bool) {
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

	delta = -lss.cost[r1][0][n1[0]][n1[1]] - lss.cost[r1][0][n1[1]][n1[2]] - lss.cost[r1][0][n1[2]][n1[3]] - lss.cost[r1][0][n1[3]][n1[4]] - lss.cost[r2][0][n2[0]][n2[1]] +
		lss.cost[r1][0][n1[0]][n1[4]] + lss.cost[r2][0][n2[0]][n1[1]] + lss.cost[r2][0][n1[1]][n1[2]] + lss.cost[r2][0][n1[2]][n1[3]] + lss.cost[r2][0][n1[3]][n2[1]]

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
func run_brute_2opt(s1 int, lss *LSS) (delta float64) {
	for i := 0; i <= len(lss.ss[s1])-2; i++ {
		for j := i + 1; j < len(lss.ss[s1]); j++ {
			delta += try_2opt(s1, i, j, lss)
		}
	}
	return
}

func run_global_5(tEndUnix int64, lss *LSS) (delta_all float64) {
	var delta float64
	for i := 0; i < len(lss.ss); i++ {
		for j := 0; j < len(lss.ss); j++ {
			if time.Now().Unix() > tEndUnix-lss.Latency*2 {
				return
			}
			start := time.Now()
			for m := -1; m < len(lss.ss[i]); m++ {
				for n := -1; n < len(lss.ss[j]); n++ {
					delta = try_2opt_extra(i, j, m, n, lss)
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

func run_global_6(tEndUnix int64, lss *LSS) (delta_all float64) {
	var delta float64
	for i := 0; i < len(lss.ss); i++ {
		for j := 0; j < len(lss.ss); j++ {
			if time.Now().Unix() > tEndUnix-lss.Latency*2 {
				return
			}
			start := time.Now()
			for m := -1; m < len(lss.ss[i]); m++ {
				for n := -1; n < len(lss.ss[j]); n++ {
					delta = try_2opt_extra_2(i, j, m, n, lss)
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
func try_2opt(s1 int, i1, i2 int, lss *LSS) (delta float64) {
	delta = d_2opt(lss.ss[s1], i1, i2, lss.ssr[s1], lss)
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
func d_2opt(s1 []int, i1, i2 int, r1 int, lss *LSS) (delta float64) {
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
	return
}

//组间交换，两条route各剪一刀，后半段全交换
func try_2opt_extra(s1, s2 int, i1, i2 int, lss *LSS) (delta float64) {
	if s1 == s2 || len(lss.ss[s1]) < 2 || len(lss.ss[s2]) < 1 {
		return
	}
	// s1 (i1, i1+1)  s2 (i2, i2+1)
	var n1, n2 []int
	if i1 == -1 {
		n1 = []int{0, lss.taskLoc[lss.ss[s1][0]]}
	} else if i1 < len(lss.ss[s1])-1 {
		n1 = []int{lss.taskLoc[lss.ss[s1][i1]], lss.taskLoc[lss.ss[s1][i1+1]]}
	} else {
		n1 = []int{lss.taskLoc[lss.ss[s1][i1]], 0}
	}
	if i2 == -1 {
		n2 = []int{0, lss.taskLoc[lss.ss[s2][0]]}
	} else if i2 < len(lss.ss[s2])-1 {
		n2 = []int{lss.taskLoc[lss.ss[s2][i2]], lss.taskLoc[lss.ss[s2][i2+1]]}
	} else {
		n2 = []int{lss.taskLoc[lss.ss[s2][i2]], 0}
	}
	r1, r2 := lss.ssr[s1], lss.ssr[s2]

	delta = -lss.cost[r1][0][n1[0]][n1[1]] - lss.cost[r2][0][n2[0]][n2[1]] + lss.cost[r1][0][n1[0]][n2[1]] + lss.cost[r2][0][n2[0]][n1[1]]

	if delta < -1e-3 {
		qty1 := getSeqQty(lss.ss[s1][:i1+1], lss) + getSeqQty(lss.ss[s2][i2+1:], lss)
		qty2 := getSeqQty(lss.ss[s2][:i2+1], lss) + getSeqQty(lss.ss[s1][i1+1:], lss)
		if qty1 > lss.rqty[r1] || qty2 > lss.rqty[r2] {
			delta = 0
		}
	}
	if delta < -1e-3 {
		s_tmp := make([]int, len(lss.ss[s1])-i1-1)
		s_tmp = CopySliceInt(lss.ss[s1][i1+1:])

		tmp1 := append(CopySliceInt(lss.ss[s1][:i1+1]), lss.ss[s2][i2+1:]...)
		tmp2 := append(CopySliceInt(lss.ss[s2][:i2+1]), s_tmp...)

		dur1 := getSeqDur(tmp1, lss.ssr[s1], lss)
		dur2 := getSeqDur(tmp2, lss.ssr[s2], lss)

		if dur1 > lss.rdur[r1] || dur2 > lss.rdur[r2] {
			delta = 0
		} else {
			dist1 := getSeqDist(tmp1, lss.ssr[s1], lss)
			dist2 := getSeqDist(tmp2, lss.ssr[s1], lss)
			if dist1 > lss.rdist[r1] || dist2 > lss.rdist[r2] {
				delta = 0
			}
		}
	}

	if delta < -1e-3 {
		wei1 := getSeqWei(lss.ss[s1][:i1+1], lss) + getSeqWei(lss.ss[s2][i2+1:], lss)
		wei2 := getSeqWei(lss.ss[s2][:i2+1], lss) + getSeqWei(lss.ss[s1][i1+1:], lss)
		if wei1 > lss.rwei[r1] || wei2 > lss.rwei[r2] {
			delta = 0
		}
	}

	if delta < -1e-3 {
		s_tmp := make([]int, len(lss.ss[s1])-i1-1)
		s_tmp = CopySliceInt(lss.ss[s1][i1+1:])
		lss.ss[s1] = append(lss.ss[s1][:i1+1], lss.ss[s2][i2+1:]...)
		lss.ss[s2] = append(lss.ss[s2][:i2+1], s_tmp...)

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
func try_2opt_extra_2(s1, s2 int, i1, i2 int, lss *LSS) (delta float64) {
	if s1 == s2 || i1+1 > len(lss.ss[s1]) || i2+1 > len(lss.ss[s2]) || len(lss.ss[s1]) < 1 || len(lss.ss[s2]) < 1 {
		return
	}
	// s1 (i1, i1+1)  s2 (i2, i2+1)  -- (i1, i2) and (i1+1, i2+1)
	var n1, n2 []int
	if i1 == -1 {
		n1 = []int{0, lss.taskLoc[lss.ss[s1][0]]}
	} else if i1 < len(lss.ss[s1])-1 {
		n1 = []int{lss.taskLoc[lss.ss[s1][i1]], lss.taskLoc[lss.ss[s1][i1+1]]}
	} else {
		n1 = []int{lss.taskLoc[lss.ss[s1][i1]], 0}
	}
	if i2 == -1 {
		n2 = []int{0, lss.taskLoc[lss.ss[s2][0]]}
	} else if i2 < len(lss.ss[s2])-1 {
		n2 = []int{lss.taskLoc[lss.ss[s2][i2]], lss.taskLoc[lss.ss[s2][i2+1]]}
	} else {
		n2 = []int{lss.taskLoc[lss.ss[s2][i2]], 0}
	}

	r1, r2 := lss.ssr[s1], lss.ssr[s2]
	delta = -lss.cost[r1][0][n1[0]][n1[1]] - lss.cost[r2][0][n2[0]][n2[1]] + lss.cost[r1][0][n1[0]][n2[0]] + lss.cost[r2][0][n1[1]][n2[1]]
	// s2[0-i2] reverse
	for i := 0; i < i2; i++ {
		delta += -lss.cost[r2][0][lss.taskLoc[lss.ss[s2][i]]][lss.taskLoc[lss.ss[s2][i+1]]] + lss.cost[r1][0][lss.taskLoc[lss.ss[s2][i+1]]][lss.taskLoc[lss.ss[s2][i]]]
	}
	if i2 == -1 {
		delta += -lss.cost[r2][0][0][n2[0]] + lss.cost[r1][0][n2[0]][0]
	} else {
		delta += -lss.cost[r2][0][0][lss.taskLoc[lss.ss[s2][0]]] + lss.cost[r1][0][lss.taskLoc[lss.ss[s2][0]]][0]
	}

	// s1[i1+1,...](0) reverse
	for i := i1 + 1; i < len(lss.ss[s1])-1; i++ {
		delta += -lss.cost[r1][0][lss.taskLoc[lss.ss[s1][i]]][lss.taskLoc[lss.ss[s1][i+1]]] + lss.cost[r2][0][lss.taskLoc[lss.ss[s1][i+1]]][lss.taskLoc[lss.ss[s1][i]]]
	}
	if i1 == len(lss.ss[s1])-1 {
		delta += -lss.cost[r1][0][n1[1]][0] + lss.cost[r2][0][0][n1[1]]
	} else {
		delta += -lss.cost[r1][0][lss.taskLoc[lss.ss[s1][len(lss.ss[s1])-1]]][0] + lss.cost[r2][0][0][lss.taskLoc[lss.ss[s1][len(lss.ss[s1])-1]]]
	}

	// use copy to check duration
	var s_tmp_1, s_tmp_2 []int
	var dur1, dur2, dist1, dist2 float64
	if delta < -1e-3 {
		s_tmp_1 = make([]int, 0, i1+i2+2)
		s_tmp_2 = make([]int, 0, len(lss.ss[s1])+len(lss.ss[s2])-i1-i2-2)

		for i := 0; i <= i1; i++ {
			s_tmp_1 = append(s_tmp_1, lss.ss[s1][i])
		}
		for i := i2; i >= 0; i-- {
			s_tmp_1 = append(s_tmp_1, lss.ss[s2][i])
		}
		for i := len(lss.ss[s1]) - 1; i >= i1+1; i-- {
			s_tmp_2 = append(s_tmp_2, lss.ss[s1][i])
		}
		for i := i2 + 1; i < len(lss.ss[s2]); i++ {
			s_tmp_2 = append(s_tmp_2, lss.ss[s2][i])
		}

		dur1 = getSeqDur(s_tmp_1, r1, lss)
		dur2 = getSeqDur(s_tmp_2, r2, lss)

		if dur1 > lss.rdur[r1] || dur2 > lss.rdur[r2] {
			delta = 0
		} else {
			dist1 = getSeqDist(s_tmp_1, r1, lss)
			dist2 = getSeqDist(s_tmp_2, r2, lss)
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
		for i := 0; i < len(s_tmp_1); i++ {
			lss.ss[s1] = append(lss.ss[s1], s_tmp_1[i])
		}
		lss.ss[s2] = lss.ss[s2][:0]
		for i := 0; i < len(s_tmp_2); i++ {
			lss.ss[s2] = append(lss.ss[s2], s_tmp_2[i])
		}
		lss.ssdur[s1] = dur1
		lss.ssdur[s2] = dur2
		lss.ssdist[s1] = dist1
		lss.ssdist[s2] = dist2
		lss.ssqty[s1] = getSeqQty(lss.ss[s1], lss)
		lss.ssqty[s2] = getSeqQty(lss.ss[s2], lss)
		lss.sswei[s1] = getSeqWei(s_tmp_1, lss)
		lss.sswei[s2] = getSeqWei(s_tmp_2, lss)
	} else {
		delta = 0
	}

	return
}

// ***************************************** vars copy ******************************************
func cpVarLS(lss *LSS, gPara *GPara, gState *GState) {
	lss.taskLoc = gPara.TaskLoc
	lss.cost = gPara.Cost
	lss.capRes = gPara.CapRes
	lss.capTask = gPara.CapTask
	lss.CapResCost = gPara.CapResCost

	lss.rqty = make([]float64, len(lss.capRes))
	lss.rdur = make([]float64, len(lss.capRes))
	lss.rwei = make([]float64, len(lss.capRes))
	lss.rdist = make([]float64, len(lss.capRes))
	for i := 0; i < len(lss.capRes); i++ {
		lss.rqty[i] = lss.capRes[i][3]
		lss.rdur[i] = lss.capRes[i][7]
		lss.rwei[i] = lss.capRes[i][5]
		lss.rdist[i] = lss.capRes[i][1]
	}
	lss.tqty = make([]float64, len(lss.capTask))
	lss.tdur = make([]float64, len(lss.capTask))
	lss.twei = make([]float64, len(lss.capTask))
	for i := 0; i < len(lss.tqty); i++ {
		lss.tqty[i] = lss.capTask[i][0]
		lss.twei[i] = lss.capTask[i][1]
		lss.tdur[i] = lss.capTask[i][2]
	}
	lss.ss = make([][]int, len(gState.InnerSeqs))
	lss.ssr = make([]int, len(lss.ss))
	for i := 0; i < len(gState.InnerSeqs); i++ {
		lss.ssr[i] = gState.InnerAsgmts[i]
		lss.ss[i] = make([]int, len(gState.InnerSeqs[i]), int(lss.capRes[lss.ssr[i]][3]))
		for j := 0; j < len(gState.InnerSeqs[i]); j++ {
			lss.ss[i][j] = gState.InnerSeqs[i][j]
		}
	}

	lss.ssqty = make([]float64, len(lss.ss))
	lss.ssdur = make([]float64, len(lss.ss))
	lss.ssdist = make([]float64, len(lss.ss))
	lss.sswei = make([]float64, len(lss.ss))
	for i := 0; i < len(gState.InnerSeqs); i++ {
		lss.ssdist[i] = getSeqDist(lss.ss[i], lss.ssr[i], lss)
		lss.ssqty[i] = getSeqQty(lss.ss[i], lss)
		lss.sswei[i] = getSeqWei(lss.ss[i], lss)
		lss.ssdur[i] = getSeqDur(lss.ss[i], lss.ssr[i], lss)
	}

}

func cpVarLSBack(lss *LSS, gState *GState) {
	var cntEmpty int
	for i := 0; i < len(lss.ss); i++ {
		if len(lss.ss[i]) == 0 {
			cntEmpty += 1
		}
	}
	if cntEmpty > 0 {
		gState.InnerSeqs = make([][]int, len(lss.ss)-cntEmpty)
		gState.InnerSeqDtls = make([][]float64, len(gState.InnerSeqs))
		gState.InnerAsgmts = make([]int, len(gState.InnerSeqs))
		var j int
		for i := 0; i < len(lss.ss); i++ {
			if len(lss.ss[i]) > 0 {
				gState.InnerSeqs[j] = lss.ss[i]
				gState.InnerAsgmts[j] = lss.ssr[i]
				gState.InnerSeqDtls[j] = make([]float64, 4)
				gState.InnerSeqDtls[j][0] = getSeqDist(gState.InnerSeqs[j], lss.ssr[i], lss)
				gState.InnerSeqDtls[j][1] = getSeqQty(gState.InnerSeqs[j], lss)
				gState.InnerSeqDtls[j][2] = getSeqWei(gState.InnerSeqs[j], lss)
				gState.InnerSeqDtls[j][3] = getSeqDur(gState.InnerSeqs[j], lss.ssr[i], lss)
				j += 1
			}
		}
	} else {
		gState.InnerSeqs = lss.ss
		for i := 0; i < len(gState.InnerSeqs); i++ {
			gState.InnerSeqDtls[i] = make([]float64, 4)
			gState.InnerSeqDtls[i][0] = getSeqDist(gState.InnerSeqs[i], lss.ssr[i], lss)
			gState.InnerSeqDtls[i][1] = getSeqQty(gState.InnerSeqs[i], lss)
			gState.InnerSeqDtls[i][2] = getSeqWei(gState.InnerSeqs[i], lss)
			gState.InnerSeqDtls[i][3] = getSeqDur(gState.InnerSeqs[i], lss.ssr[i], lss)
		}
	}
}

// ******************************************************
func getSeqsDist(seqs [][]int, lss *LSS) (d float64) {
	for i := 0; i < len(seqs); i++ {
		d += getSeqDist(seqs[i], lss.ssr[i], lss)
	}
	return
}

func getSeqsDistNew(seqs [][]int, lss *LSS) (d float64) {
	for i := 0; i < len(seqs); i++ {
		d += getSeqDistNew(seqs[i], lss.ssr[i], lss)
	}
	return
}

func getSeqsDur(seqs [][]int, lss *LSS) (d float64) {
	for i := 0; i < len(seqs); i++ {
		d += getSeqDur(seqs[i], lss.ssr[i], lss)
	}
	return
}

func getSeqsDurList(seqs [][]int, lss *LSS) (dlst []float64) {
	dlst = make([]float64, len(seqs))
	for i := 0; i < len(seqs); i++ {
		dlst[i] = getSeqDur(seqs[i], lss.ssr[i], lss)
	}
	return
}

func getSeqsDistList(seqs [][]int, lss *LSS) (dlst []float64) {
	dlst = make([]float64, len(seqs))
	for i := 0; i < len(seqs); i++ {
		dlst[i] = getSeqDist(seqs[i], lss.ssr[i], lss)
	}
	return
}

func getSeqsQtyList(seqs [][]int, lss *LSS) (lst []float64) {
	lst = make([]float64, len(seqs))
	for i := 0; i < len(seqs); i++ {
		for j := 0; j < len(seqs[i]); j++ {
			lst[i] += lss.capTask[seqs[i][j]][0]
		}
	}
	return
}

func getSeqQty(s []int, lss *LSS) (qty float64) {
	for j := 0; j < len(s); j++ {
		qty += lss.capTask[s[j]][0]
	}
	return
}
func getSeqQtyNew(s []int, lss *LSS) (qty float64) {
	for j := 1; j < len(s)-1; j++ {
		qty += lss.capTask[s[j]-1][0]
	}
	return
}

func getSeqWei(s []int, lss *LSS) (wei float64) {
	for j := 0; j < len(s); j++ {
		wei += lss.capTask[s[j]][1]
	}
	return
}

func getSeqDist(seq []int, r int, lss *LSS) (d float64) {
	if len(seq) < 1 {
		return 0
	}
	d += lss.cost[r][0][0][lss.taskLoc[seq[0]]] + lss.cost[r][0][lss.taskLoc[seq[len(seq)-1]]][0]
	for i := 0; i < len(seq)-1; i++ {
		d += lss.cost[r][0][lss.taskLoc[seq[i]]][lss.taskLoc[seq[i+1]]]
	}
	return
}

func getSeqDistNew(seq []int, r int, lss *LSS) (d float64) {
	if len(seq) < 1 {
		return 0
	}
	for i := 0; i < len(seq)-1; i++ {
		d += lss.cost[r][0][seq[i]][seq[i+1]]
	}
	return
}

func getSeqDur(seq []int, r int, lss *LSS) (d float64) {
	if len(seq) < 1 {
		return 0
	}
	d += lss.cost[r][1][0][lss.taskLoc[seq[0]]] + lss.cost[r][1][lss.taskLoc[seq[len(seq)-1]]][0]
	for i := 0; i < len(seq)-1; i++ {
		d += lss.cost[r][1][lss.taskLoc[seq[i]]][lss.taskLoc[seq[i+1]]]
	}
	for i := 0; i < len(seq); i++ {
		d += lss.capTask[seq[i]][2]
	}
	return
}

func getSeqDurNew(seq []int, r int, lss *LSS) (d float64) {
	if len(seq) < 1 {
		return 0
	}
	for i := 0; i < len(seq)-1; i++ {
		d += lss.cost[r][1][seq[i]][seq[i+1]]
	}
	for i := 1; i < len(seq)-1; i++ {
		d += lss.capTask[seq[i]-1][2]
	}
	return
}

func getSeqDurTrim(seq []int, r int, lss *LSS) (d float64) {
	if len(seq) < 1 {
		return 0
	}
	for i := 0; i < len(seq)-1; i++ {
		d += lss.cost[r][1][lss.taskLoc[seq[i]]][lss.taskLoc[seq[i+1]]]
	}
	for i := 0; i < len(seq); i++ {
		d += lss.capTask[seq[i]][2]
	}
	return
}

func getSeqsDistTrim(seqs [][]int, lss *LSS) (d float64) {
	for i := 0; i < len(seqs); i++ {
		d += getSeqDistTrim(seqs[i], lss.ssr[i], lss)
	}
	return
}

func getSeqDistTrim(seq []int, r int, lss *LSS) (d float64) {
	for i := 0; i < len(seq)-1; i++ {
		d += lss.cost[r][0][lss.taskLoc[seq[i]]][lss.taskLoc[seq[i+1]]]
	}
	return
}

func getConstr(lss *LSS) int {
	for i := 0; i < len(lss.ss); i++ {
		if getSeqDur(lss.ss[i], lss.ssr[i], lss) > lss.rdur[lss.ssr[i]] {
			return 1
		}
		if getSeqDist(lss.ss[i], lss.ssr[i], lss) > lss.rdist[lss.ssr[i]] {
			return 2
		}
		if getSeqQty(lss.ss[i], lss) > lss.rqty[lss.ssr[i]] {
			return 3
		}
	}
	return 0
}

func copyFull(lss *LSS) (ssnew [][]int) {
	ssnew = make([][]int, len(lss.ss))
	for i := 0; i < len(lss.ss); i++ {
		ssnew[i] = make([]int, 0, cap(lss.ss[i])+2)
		ssnew[i] = append(ssnew[i], 0)
		for j := 0; j < len(lss.ss[i]); j++ {
			ssnew[i] = append(ssnew[i], lss.taskLoc[lss.ss[i][j]])
		}
		ssnew[i] = append(ssnew[i], 0)
	}
	return
}

//func copyPartBack(ssnew [][]int, ss [][]int) {
//	for i := 0; i < len(ssnew); i++ {
//		ss[i] = ss[i][:0]
//		for j := 1; j < len(ssnew[i])-1; j++ {
//			ss[i] = append(ss[i], ssnew[i][j]-1)
//		}
//	}
//}

func copyTo(m1, m2 [][]int, f1, f2 *SSFeat) {
	for i := 0; i < len(m1); i++ {
		m2[i] = m2[i][:0]
		for j := 0; j < len(m1[i]); j++ {
			m2[i] = append(m2[i], m1[i][j])
		}
	}
	if f1 != nil {
		for i := 0; i < len(m1); i++ {
			f2.dist[i] = f1.dist[i]
			f2.dur[i] = f1.dur[i]
			f2.qty[i] = f1.qty[i]
			if f2.innerDist != nil {
				f2.innerDist[i] = f1.innerDist[i]
			}
			if f2.mapCost != nil {
				f2.mapCost[i] = f1.mapCost[i]
			}
		}
	}
}

//func CopySliceInt(s []int) (sDup []int) {
//	sDup = make([]int, len(s))
//	for i := 0; i < len(s); i++ {
//		sDup[i] = s[i]
//	}
//	return
//}

func CopySliceFloat(s []float64) (sDup []float64) {
	sDup = make([]float64, len(s))
	for i := 0; i < len(s); i++ {
		sDup[i] = s[i]
	}
	return
}

func maxSlice(s []float64) float64 {
	var smax float64
	for i := 0; i < len(s); i++ {
		if smax < s[i] {
			smax = s[i]
		}
	}
	return smax
}

func checkCnt(m [][]int) (cnt int) {
	for i := 0; i < len(m); i++ {
		cnt += len(m[i])
	}
	return
}

func cntTrue(s []bool) (cnt int) {
	for k := 0; k < len(s); k++ {
		if s[k] {
			cnt += 1
		}
	}
	return
}

//func findMin(s []float64) (idx int, v float64) {
//	v = 1e8
//	idx = -1
//	for i := 0; i < len(s); i++ {
//		if s[i] < v {
//			v = s[i]
//			idx = i
//		}
//	}
//	return
//}

func copyMatrixI(m [][]int) (mCp [][]int) {
	mCp = make([][]int, len(m))
	for i := 0; i < len(m); i++ {
		mCp[i] = CopySliceInt(m[i])
	}
	return
}
