package solver

import (
	"math/rand"
	"time"
)

// **************************************** draft **********************************************************
// **************************************** add accept *******************************************************
//
//func GetSingleDelta(gPara *GPara, oldDtl, newDtl []float64) float64 {
//	var delta float64
//	if gPara.Obj == 0 {
//		delta = newDtl[0] - oldDtl[0]
//	} else {
//		deltaMC := newDtl[5] - oldDtl[5]
//		if deltaMC < 0 {
//			delta = deltaMC
//		} else if deltaMC == 0 {
//			deltaFC := newDtl[4] - oldDtl[4]
//			delta = deltaFC
//		} else {
//			delta = 100000
//		}
//	}
//	return delta
//}
//
//func GetGroupDelta(gPara *GPara, oldDtl1, oldDtl2, newDtl1, newDtl2 []float64) float64 {
//	var delta float64
//	if gPara.Obj == 0 {
//		delta = (newDtl1[0] + newDtl2[0]) - (oldDtl1[0] + oldDtl2[0])
//	} else {
//		deltaMC := (newDtl1[5] + newDtl2[5]) - (oldDtl1[5] + oldDtl2[5])
//		if deltaMC < 0 {
//			delta = deltaMC
//		} else if deltaMC == 0 {
//			deltaFC := (newDtl1[4] + newDtl2[4]) - (oldDtl1[4] + oldDtl2[4])
//			delta = deltaFC
//		} else {
//			delta = 100000
//		}
//	}
//	return delta
//}
//func SelectSeqs(k1 int, feats [][]float64) (k2 int) {
//	kCent := feats[k1]
//	var minD float64 = math.Maxfloat64
//	k2 = k1
//	for i := 0; i < len(feats); i++ {
//		if i == k1 {
//			continue
//		}
//		dist := GreatCircleDistance(kCent, feats[i])
//		if dist < minD {
//			minD = dist
//			k2 = i
//		}
//	}
//	return
//}
//
//// apply ops
//func Method01(optEndTime int64, gPara *GPara, gState *GState) {
//	//cntIter := []int{2e4, 2e4}
//	if time.Now().Unix() < optEndTime-5 {
//		search1Weighted(gPara, gState)
//		search3Weighted(gPara, gState)
//		search4Weighted(gPara, gState)
//	}
//
//	if time.Now().Unix() < optEndTime-5 {
//		search1Greedy(gPara, gState)
//		search3Greedy(gPara, gState)
//		search4Greedy(gPara, gState)
//	}
//}
//
//func search1Greedy(gPara *GPara, gState *GState) {
//	for i := 0; i < len(gState.InnerSeqs); i++ {
//		//preDist := gState.InnerSeqDtls[i][0]
//		//fmt.Println("search4-", i, ":", gState.InnerSeqDtls[i])
//		//fmt.Println(i, ":", gState.InnerSeqDtls[i])
//
//		for i1 := 0; i1 < len(gState.InnerSeqs[i])-1; i1++ {
//			for i2 := i1 + 1; i2 < len(gState.InnerSeqs[i]); i2++ {
//				search1(gPara, gState.InnerSeqs[i], gState.InnerAsgmts[i], i1, i2, gState.InnerSeqDtls[i])
//				//if ok1 {
//				//	fmt.Println("search1-", i, "success", i1, " ", i2, ":", gState.InnerSeqDtls[i])
//				//	postDist := gState.InnerSeqDtls[i][0]
//				//	improvePerc := (preDist - postDist) / preDist * 100
//				//	fmt.Println("search1-", i, "improvePerc:", improvePerc)
//				//	//if improvePerc == 0 {
//				//	//	fmt.Println("stop")
//				//	//}
//				//}
//			}
//		}
//		//fmt.Println("search3-", i, ":", gState.InnerSeqDtls[i])
//		//postDist := getObj()
//		//improvePerc := (preDist - postDist) / preDist * 100
//		//fmt.Println("search3-", i, "improvePerc:", improvePerc)
//	}
//}
//
//func search3Greedy(gPara *GPara, gState *GState) {
//	for i := 0; i < len(gState.InnerSeqs); i++ {
//		//preDist := gState.InnerSeqDtls[i][0]
//		//fmt.Println("search4-", i, ":", gState.InnerSeqDtls[i])
//		//fmt.Println(i, ":", gState.InnerSeqDtls[i])
//
//		for i1 := 0; i1 < len(gState.InnerSeqs[i])-1; i1++ {
//			for i2 := i1 + 1; i2 < len(gState.InnerSeqs[i]); i2++ {
//				search3(gPara, gState.InnerSeqs[i], gState.InnerAsgmts[i], i1, i2, gState.InnerSeqDtls[i])
//				//if ok1 {
//				//	fmt.Println("search3-", i, "success", i1, " ", i2, ":", gState.InnerSeqDtls[i])
//				//	postDist := gState.InnerSeqDtls[i][0]
//				//	improvePerc := (preDist - postDist) / preDist * 100
//				//	fmt.Println("search3-", i, "improvePerc:", improvePerc)
//				//	//if improvePerc == 0 {
//				//	//	fmt.Println("stop")
//				//	//}
//				//}
//			}
//		}
//		//fmt.Println("search3-", i, ":", gState.InnerSeqDtls[i])
//		//postDist := getObj()
//		//improvePerc := (preDist - postDist) / preDist * 100
//		//fmt.Println("search3-", i, "improvePerc:", improvePerc)
//	}
//}
//
//func search4Greedy(gPara *GPara, gState *GState) {
//	for i := 0; i < len(gState.InnerSeqs); i++ {
//		//preDist := gState.InnerSeqDtls[i][0]
//		//fmt.Println("search4-", i, ":", gState.InnerSeqDtls[i])
//		//fmt.Println(i, ":", gState.InnerSeqDtls[i])
//
//		for i1 := 0; i1 < len(gState.InnerSeqs[i])-1; i1++ {
//			for i2 := i1 + 1; i2 < len(gState.InnerSeqs[i]); i2++ {
//				search4(gPara, gState.InnerSeqs[i], gState.InnerAsgmts[i], i1, i2, gState.InnerSeqDtls[i], gState.InnerFeats[i])
//				//if ok4 {
//				//	fmt.Println("search4-", i, "success", i1, " ", i2, ":", gState.InnerSeqDtls[i])
//				//	postDist := gState.InnerSeqDtls[i][0]
//				//	improvePerc := (preDist - postDist) / preDist * 100
//				//	fmt.Println("search4-", i, "improvePerc:", improvePerc)
//				//	//if improvePerc == 0 {
//				//	//	fmt.Println("stop")
//				//	//}
//				//}
//			}
//		}
//		//fmt.Println("search3-", i, ":", gState.InnerSeqDtls[i])
//		//postDist := getObj()
//		//improvePerc := (preDist - postDist) / preDist * 100
//		//fmt.Println("search3-", i, "improvePerc:", improvePerc)
//	}
//}
//
//// **************************************** search **********************************************************
//// swap two gPara.Nodes in a seq
//func search1(gPara *GPara, s []int, r int, i1, i2 int, d []float64) bool {
//	if i1 < 0 {
//		i1 = rand.Intn(len(s) - 1)
//		i2 = i1 + 1 + rand.Intn(len(s)-1-i1)
//	}
//
//	delta, newD, ok := dMove01(gPara, s, i1, i2, r, d, nil)
//
//	if ok && delta < 0 {
//		s[i1], s[i2] = s[i2], s[i1]
//		for i := 0; i < len(d); i++ {
//			d[i] = newD[i]
//		}
//	}
//	return ok && delta < 0
//}
//
//// swap two gPara.Nodes between two different seqs
//func search2(gPara *GPara, s1 []int, s2 []int, r1 int, r2 int, i1, i2 int, d1 []float64, d2 []float64, f1 []float64, f2 []float64) bool {
//	if i1 < 0 {
//		i1 = rand.Intn(len(s1) - 1)
//		i2 = rand.Intn(len(s2) - 1)
//	}
//
//	delta, ok, newD1, newD2 := dMove02(gPara, s1, s2, i1, i2, r1, r2, d1, d2, f1, f2)
//	if ok && delta < 0 {
//		//update cent
//		lat1 := (f1[0]*float64(len(s1)) - gPara.Nodes[gPara.TaskLoc[s1[i1]]][0] + gPara.Nodes[gPara.TaskLoc[s2[i2]]][0]) / float64(len(s1))
//		lng1 := (f1[1]*float64(len(s1)) - gPara.Nodes[gPara.TaskLoc[s1[i1]]][1] + gPara.Nodes[gPara.TaskLoc[s2[i2]]][1]) / float64(len(s1))
//
//		lat2 := (f2[0]*float64(len(s2)) - gPara.Nodes[gPara.TaskLoc[s2[i2]]][0] + gPara.Nodes[gPara.TaskLoc[s1[i1]]][0]) / float64(len(s2))
//		lng2 := (f2[1]*float64(len(s2)) - gPara.Nodes[gPara.TaskLoc[s2[i2]]][1] + gPara.Nodes[gPara.TaskLoc[s1[i1]]][1]) / float64(len(s2))
//
//		f1[0] = lat1
//		f1[1] = lng1
//		f2[0] = lat2
//		f2[1] = lng2
//
//		s1[i1], s2[i2] = s2[i2], s1[i1]
//		for i := 0; i < len(d1); i++ {
//			d1[i] = newD1[i]
//			d2[i] = newD2[i]
//		}
//
//	}
//	return ok && delta < 0
//}
//
//// move i1 to after i2 in s
//func search3(gPara *GPara, s []int, r int, i1, i2 int, d []float64) bool {
//	if len(s) < 3 {
//		return false
//	}
//	if i1 < 0 {
//		i1 = rand.Intn(len(s) - 1)
//		i2 = i1 + 1 + rand.Intn(len(s)-1-i1)
//		if rand.Intn(2) > 0 {
//			i1, i2 = i2, i1
//		}
//	}
//
//	delta, ok, newD := dMove03(gPara, s, i1, i2, r, d)
//
//	if ok && delta < 0 {
//		if i1 < i2 {
//			//fmt.Println(s[i1], s[i2])
//			tmp := s[i1]
//			for i := i1; i < i2; i++ {
//				s[i] = s[i+1]
//			}
//			s[i2] = tmp
//			//fmt.Println(s[i2], s[i2-1])
//		} else {
//			//fmt.Println(s[i1], s[i2])
//			//fmt.Println("i1:", i1, "i2:", i2)
//			tmp := s[i1]
//			for i := i1; i > i2+1; i-- {
//				s[i] = s[i-1]
//			}
//			s[i2+1] = tmp
//			//fmt.Println(s[i2+1], s[i2])
//		}
//		for i := 0; i < len(d); i++ {
//			d[i] = newD[i]
//		}
//	}
//	return ok && delta < 0
//}
//
//// apply 2-opt in s
//func search4(gPara *GPara, s []int, r int, i1, i2 int, d []float64, f []float64) bool {
//	if i1 < 0 {
//		i1 = rand.Intn(len(s) - 3)
//		i2 = i1 + 2 + rand.Intn(len(s)-1-i1-2)
//	}
//
//	delta, ok, newD := dMove04(gPara, s, i1, i2, r, d, f)
//
//	if ok && delta < 0 {
//		//fmt.Println("before search4 opt:", getSeqObj(s, r), "detail:", d, "seq:", s)
//		reverse(s, i1, i2)
//		//fmt.Println("reverse:", i1, "--", i2)
//		for i := 0; i < len(d); i++ {
//			d[i] = newD[i]
//		}
//		//dist := getSeqDist(s, r)
//		//if d[0] != dist {
//		//	fmt.Println("距离不一致:", "dist:", dist, "  d[0]:", d[0])
//		//}
//		//fmt.Println("after search4 opt:", getSeqObj(s, r), "detail:", d, "seq:", s)
//	}
//	return ok && delta < 0
//}
//
//// **************************************** check delta **********************************************************
//// swap two gPara.Nodes s1[i1] and s1[i2]
//func dMove01(gPara *GPara, s1 []int, i1, i2 int, r1 int, d1 []float64, feat1 []float64) (delta float64, newD []float64, ok bool) {
//	if i1 > i2 {
//		i1, i2 = i2, i1
//	}
//
//	n1 := make([]int, 3)
//	n2 := make([]int, 3)
//	if i1 > 0 {
//		n1[0] = gPara.TaskLoc[s1[i1-1]]
//	}
//	n1[1] = gPara.TaskLoc[s1[i1]]
//	n1[2] = gPara.TaskLoc[s1[i1+1]]
//	n2[0] = gPara.TaskLoc[s1[i2-1]]
//	n2[1] = gPara.TaskLoc[s1[i2]]
//	if i2 < len(s1)-1 {
//		n2[2] = gPara.TaskLoc[s1[i2+1]]
//	}
//	// acceptance
//	ok, newD, delta = checkMove01(gPara, i1, i2, r1, n1, n2, d1)
//	return
//}
//
////CapResCost[resIdx][t][0]:LowerBound
////CapResCost[resIdx][t][1]:UpperBound
////CapResCost[resIdx][t][2]:OuterK
////CapResCost[resIdx][t][3]:LowerCost
////CapResCost[resIdx][t][4]:UpperCost
////CapResCost[resIdx][t][5]:InnerK
//func checkMove01(gPara *GPara, i1, i2 int, r1 int, n1 []int, n2 []int, d []float64) (ok bool, newD []float64, delta float64) {
//	newD = make([]float64, len(d))
//	var dtD float64 = 0.0
//	var dtDrt float64 = 0.0
//	//distance & duration
//	if i1+1 < i2 {
//		dtD = -gPara.Cost[r1][0][n1[0]][n1[1]] - gPara.Cost[r1][0][n1[1]][n1[2]] - gPara.Cost[r1][0][n2[0]][n2[1]] - gPara.Cost[r1][0][n2[1]][n2[2]] +
//			gPara.Cost[r1][0][n1[0]][n2[1]] + gPara.Cost[r1][0][n2[1]][n1[2]] + gPara.Cost[r1][0][n2[0]][n1[1]] + gPara.Cost[r1][0][n1[1]][n2[2]]
//
//		dtDrt = -gPara.Cost[r1][1][n1[0]][n1[1]] - gPara.Cost[r1][1][n1[1]][n1[2]] - gPara.Cost[r1][1][n2[0]][n2[1]] - gPara.Cost[r1][1][n2[1]][n2[2]] +
//			gPara.Cost[r1][1][n1[0]][n2[1]] + gPara.Cost[r1][1][n2[1]][n1[2]] + gPara.Cost[r1][1][n2[0]][n1[1]] + gPara.Cost[r1][1][n1[1]][n2[2]]
//	} else if i1 < i2 {
//		dtD = -gPara.Cost[r1][0][n1[0]][n1[1]] - gPara.Cost[r1][0][n1[1]][n1[2]] - gPara.Cost[r1][0][n2[1]][n2[2]] +
//			gPara.Cost[r1][0][n1[0]][n2[1]] + gPara.Cost[r1][0][n2[1]][n1[1]] + gPara.Cost[r1][0][n1[1]][n2[2]]
//
//		dtDrt = -gPara.Cost[r1][1][n1[0]][n1[1]] - gPara.Cost[r1][1][n1[1]][n1[2]] - gPara.Cost[r1][1][n2[1]][n2[2]] +
//			gPara.Cost[r1][1][n1[0]][n2[1]] + gPara.Cost[r1][1][n2[1]][n1[1]] + gPara.Cost[r1][1][n1[1]][n2[2]]
//	}
//
//	newD[0] = d[0] + dtD
//	//if (d[0]-newD[0])/d[0]*100 >= 30 {
//	//	fmt.Println("stop")
//	//}
//	newD[3] = d[3] + dtDrt
//	newD[1] = d[1]
//	newD[2] = d[2]
//	newD[4], newD[5] = GetSeqMFCostByDist(gPara, r1, newD[0])
//
//	delta = GetSingleDelta(gPara, d, newD)
//
//	dD := (newD[0] <= gPara.CapRes[r1][1]) && (newD[0] >= gPara.CapRes[r1][0])
//	dDrt := (newD[3] <= gPara.CapRes[r1][7]) && (newD[3] >= gPara.CapRes[r1][6])
//
//	ok = dD && dDrt
//	return
//}
//
//func checkMove03(gPara *GPara, r1 int, n1 []int, n2 []int, d []float64) (delta float64, ok bool, newD []float64) {
//	newD = make([]float64, len(d))
//	var dtD float64 = 0.0
//	var dtDrt float64 = 0.0
//	//distance & duration
//	dtD = -gPara.Cost[r1][0][n1[0]][n1[1]] - gPara.Cost[r1][0][n1[1]][n1[2]] - gPara.Cost[r1][0][n2[0]][n2[1]] +
//		gPara.Cost[r1][0][n1[0]][n1[2]] + gPara.Cost[r1][0][n2[0]][n1[1]] + gPara.Cost[r1][0][n1[1]][n2[1]]
//
//	dtDrt = -gPara.Cost[r1][1][n1[0]][n1[1]] - gPara.Cost[r1][1][n1[1]][n1[2]] - gPara.Cost[r1][1][n2[0]][n2[1]] +
//		gPara.Cost[r1][1][n1[0]][n1[2]] + gPara.Cost[r1][1][n2[0]][n1[1]] + gPara.Cost[r1][1][n1[1]][n2[1]]
//
//	newD[0] = d[0] + dtD
//	//if (d[0]-newD[0])/d[0]*100 >= 30 {
//	//	fmt.Println("stop")
//	//}
//	newD[3] = d[3] + dtDrt
//	newD[1] = d[1]
//	newD[2] = d[2]
//	newD[4], newD[5] = GetSeqMFCostByDist(gPara, r1, newD[0])
//
//	delta = GetSingleDelta(gPara, d, newD)
//
//	//if (d[0]-newD[0])/d[0]*100 > 30 {
//	//	fmt.Println("大于30%")
//	//}
//
//	dD := (newD[0] <= gPara.CapRes[r1][1]) && (newD[0] >= gPara.CapRes[r1][0])
//	dDrt := (newD[3] <= gPara.CapRes[r1][7]) && (newD[3] >= gPara.CapRes[r1][6])
//
//	ok = dD && dDrt
//	return
//}
//
////d[0] -- distance
////d[1] -- parcel
////d[2] -- weight
////d[3] -- duration
//
//func checkMove02(gPara *GPara, s1, s2 []int, i1, i2 int, r1, r2 int, d1, d2 []float64, n1 []int, n2 []int) (delta float64, ok bool, newd1 []float64, newd2 []float64) {
//	newd1 = make([]float64, len(d1))
//	newd2 = make([]float64, len(d2))
//	//dt1 Parcel
//	dtP1 := -gPara.CapTask[s1[i1]][0] + gPara.CapTask[s2[i2]][0]
//	//dt1 Weight
//	dtW1 := -gPara.CapTask[s1[i1]][1] + gPara.CapTask[s2[i2]][1]
//	//dt1 Distance
//	dtD1 := -gPara.Cost[r1][0][n1[0]][n1[1]] - gPara.Cost[r1][0][n1[1]][n1[2]] + gPara.Cost[r1][0][n1[0]][n2[1]] + gPara.Cost[r1][0][n2[1]][n1[2]]
//	//dt1 Duration
//	dtDrt1 := -gPara.Cost[r1][1][n1[0]][n1[1]] - gPara.Cost[r1][1][n1[1]][n1[2]] - gPara.CapTask[s1[i1]][2] + gPara.Cost[r1][1][n1[0]][n2[1]] + gPara.Cost[r1][1][n2[1]][n1[2]] + gPara.CapTask[s2[i2]][2]
//
//	// check res cont
//	newd1[1] = d1[1] + dtP1
//	newd1[2] = d1[2] + dtW1
//	newd1[0] = d1[0] + dtD1
//	newd1[3] = d1[3] + dtDrt1
//	newd1[4], newd1[5] = GetSeqMFCostByDist(gPara, r1, newd1[0])
//
//	//dt2 Parcel
//	dtP2 := -gPara.CapTask[s2[i2]][0] + gPara.CapTask[s1[i1]][0]
//	//dt2 Weight
//	dtW2 := -gPara.CapTask[s2[i2]][1] + gPara.CapTask[s1[i1]][1]
//	//dt2 Distance
//	dtD2 := -gPara.Cost[r2][0][n2[0]][n2[1]] - gPara.Cost[r2][0][n2[1]][n2[2]] + gPara.Cost[r2][0][n2[0]][n1[1]] + gPara.Cost[r2][0][n1[1]][n2[2]]
//	//dt2 Duration
//	dtDrt2 := -gPara.Cost[r2][1][n2[0]][n2[1]] - gPara.Cost[r2][1][n2[1]][n2[2]] - gPara.CapTask[s2[i2]][2] + gPara.Cost[r2][1][n2[0]][n1[1]] + gPara.Cost[r2][1][n1[1]][n2[2]] + gPara.CapTask[s1[i1]][2]
//
//	// check res cont
//	newd2[1] = d2[1] + dtP2
//	newd2[2] = d2[2] + dtW2
//	newd2[0] = d2[0] + dtD2
//	newd2[3] = d2[3] + dtDrt2
//	newd2[4], newd2[5] = GetSeqMFCostByDist(gPara, r2, newd2[0])
//
//	d1P := (newd1[1] <= gPara.CapRes[r1][3]) && (newd1[1] >= gPara.CapRes[r1][2])
//	d1W := (newd1[2] <= gPara.CapRes[r1][5]) && (newd1[2] >= gPara.CapRes[r1][4])
//	d1D := (newd1[0] <= gPara.CapRes[r1][1]) && (newd1[0] >= gPara.CapRes[r1][0])
//	d1Drt := (newd1[3] <= gPara.CapRes[r1][7]) && (newd1[3] >= gPara.CapRes[r1][6])
//	d2P := (newd2[1] <= gPara.CapRes[r2][3]) && (newd2[1] >= gPara.CapRes[r2][2])
//	d2W := (newd2[2] <= gPara.CapRes[r2][5]) && (newd2[2] >= gPara.CapRes[r2][4])
//	d2D := (newd2[0] <= gPara.CapRes[r2][1]) && (newd2[0] >= gPara.CapRes[r2][0])
//	d2Drt := (newd2[3] <= gPara.CapRes[r2][7]) && (newd2[3] >= gPara.CapRes[r2][6])
//
//	ok = d1P && d1W && d1D && d1Drt && d2P && d2W && d2D && d2Drt
//	delta = GetGroupDelta(gPara, d1, d2, newd1, newd2)
//	return
//}
//
//func dMove02(gPara *GPara, s1, s2 []int, i1, i2 int, r1, r2 int, d1, d2 []float64, feat1, feat2 []float64) (delta float64, ok bool, newD1 []float64, newD2 []float64) {
//	n1 := make([]int, 3)
//	n2 := make([]int, 3)
//
//	if i1 > 0 {
//		n1[0] = gPara.TaskLoc[s1[i1-1]]
//	}
//	n1[1] = gPara.TaskLoc[s1[i1]]
//	if i1 < len(s1)-1 {
//		n1[2] = gPara.TaskLoc[s1[i1+1]]
//	}
//
//	if i2 > 0 {
//		n2[0] = gPara.TaskLoc[s2[i2-1]]
//	}
//	n2[1] = gPara.TaskLoc[s2[i2]]
//	if i2 < len(s2)-1 {
//		n2[2] = gPara.TaskLoc[s2[i2+1]]
//	}
//
//	// acceptance
//	//校验 parcel 1  weight  2
//	delta, ok, newD1, newD2 = checkMove02(gPara, s1, s2, i1, i2, r1, r2, d1, d2, n1, n2)
//
//	return
//}
//
//// move i1 after i2 in s1
//func dMove03(gPara *GPara, s1 []int, i1, i2 int, r1 int, d1 []float64) (delta float64, ok bool, newD []float64) {
//	if i1 == i2+1 {
//		return
//	}
//
//	n1 := make([]int, 3)
//	n2 := make([]int, 2)
//	if i1 > 0 {
//		n1[0] = gPara.TaskLoc[s1[i1-1]]
//	}
//	n1[1] = gPara.TaskLoc[s1[i1]]
//	if i1 < len(s1)-1 {
//		n1[2] = gPara.TaskLoc[s1[i1+1]]
//	}
//	n2[0] = gPara.TaskLoc[s1[i2]]
//	if i2 < len(s1)-1 {
//		n2[1] = gPara.TaskLoc[s1[i2+1]]
//	}
//
//	// acceptance
//	delta, ok, newD = checkMove03(gPara, r1, n1, n2, d1)
//	return
//}
//
//// apply 2-opt in s for (i1,i1+1) (i2,i2+1)
//func dMove04(gPara *GPara, s1 []int, i1 int, i2 int, r1 int, d1 []float64, feat1 []float64) (delta float64, ok bool, newD []float64) {
//
//	n1 := make([]int, 2)
//	n2 := make([]int, 2)
//	if i1 == 0 {
//		n1[0] = 0
//	} else {
//		n1[0] = gPara.TaskLoc[s1[i1-1]]
//	}
//	n1[1] = gPara.TaskLoc[s1[i1]]
//	n2[0] = gPara.TaskLoc[s1[i2]]
//	if i2 == len(s1)-1 {
//		n2[1] = 0
//	} else {
//		n2[1] = gPara.TaskLoc[s1[i2+1]]
//	}
//
//	// acceptance
//	delta, ok, newD = checkMove04(gPara, s1, i1, i2, r1, n1, n2, d1)
//
//	return
//}
//
//func checkMove04(gPara *GPara, s1 []int, i1 int, i2 int, r1 int, n1 []int, n2 []int, d []float64) (delta float64, ok bool, newD []float64) {
//	newD = make([]float64, len(d))
//	var dtD float64 = 0.0
//	var dtDrt float64 = 0.0
//
//	// first assume d12 = d21 for simplicity, then check full path
//	dtD = -gPara.Cost[r1][0][n1[0]][n1[1]] - gPara.Cost[r1][0][n2[0]][n2[1]] + gPara.Cost[r1][0][n1[0]][n2[0]] + gPara.Cost[r1][0][n1[1]][n2[1]]
//	dtDrt = -gPara.Cost[r1][1][n1[0]][n1[1]] - gPara.Cost[r1][1][n2[0]][n2[1]] + gPara.Cost[r1][1][n1[0]][n2[0]] + gPara.Cost[r1][1][n1[1]][n2[1]]
//	//newD[0] = d[0] + dtD
//	//newD[3] = d[3] + dtDrt
//	// (i1, i2) reverse diff
//	//if dtD < 0 {
//	for i := i1; i < i2; i++ {
//		dtD += -gPara.Cost[r1][0][gPara.TaskLoc[s1[i]]][gPara.TaskLoc[s1[i+1]]] + gPara.Cost[r1][0][gPara.TaskLoc[s1[i+1]]][gPara.TaskLoc[s1[i]]]
//		dtDrt += -gPara.Cost[r1][1][gPara.TaskLoc[s1[i]]][gPara.TaskLoc[s1[i+1]]] + gPara.Cost[r1][1][gPara.TaskLoc[s1[i+1]]][gPara.TaskLoc[s1[i]]]
//	}
//	newD[0] = d[0] + dtD
//
//	//if (d[0]-newD[0])/d[0]*100 >= 30 {
//	//	fmt.Println("stop")
//	//}
//	newD[3] = d[3] + dtDrt
//	//}
//
//	newD[1] = d[1]
//	newD[2] = d[2]
//	newD[4], newD[5] = GetSeqMFCostByDist(gPara, r1, newD[0])
//
//	delta = GetSingleDelta(gPara, d, newD)
//
//	dD := (newD[0] <= gPara.CapRes[r1][1]) && (newD[0] >= gPara.CapRes[r1][0])
//	dDrt := (newD[3] <= gPara.CapRes[r1][7]) && (newD[3] >= gPara.CapRes[r1][6])
//
//	ok = dD && dDrt
//	return
//}
//
//// *********************************** greedy swap **********************************************
//func Method_tmp(optTime int, seqs [][]int, seqDtls [][]float64, gPara *GPara, gState *GState) {
//	var ok bool = true
//	var eff1, eff2 bool
//	for gIter := 0; gIter < 1 && ok; gIter++ {
//		for iter := 0; iter < 1; iter++ {
//			eff1 = gSearch1_greedy(gPara, seqs, seqDtls, []int{}) || eff1
//			if !eff1 {
//				break
//			}
//		}
//		for iter := 0; iter < 1; iter++ {
//			eff2 = gSearch2_greedy(gPara, gState, seqs, seqDtls) || eff2
//			if !eff2 {
//				break
//			}
//		}
//		ok = eff1 || eff2
//	}
//}
//
//func gSearch1_greedy(gPara *GPara, seqs [][]int, seqDtls [][]float64, seqIdx []int) (ok bool) {
//	if len(seqIdx) > 0 {
//		for _, i := range seqIdx {
//			for m := 0; m < len(seqs[i])-1; m++ {
//				for n := m + 1; n < len(seqs[i]); n++ {
//					ok = ok || search1(gPara, seqs[i], 0, m, n, seqDtls[i])
//				}
//			}
//		}
//	} else {
//		for i := 0; i < len(seqs); i++ {
//			for m := 0; m < len(seqs[i])-1; m++ {
//				for n := m + 1; n < len(seqs[i]); n++ {
//					ok = ok || search1(gPara, seqs[i], 0, m, n, seqDtls[i])
//				}
//			}
//		}
//	}
//	return
//}
//
//func gSearch2_greedy(gPara *GPara, gState *GState, seqs [][]int, seqDtls [][]float64) (ok bool) {
//	for i := 0; i < len(seqs); i++ {
//		for j := i + 1; j < len(seqs); j++ {
//			for m := 0; m < len(seqs[i]); m++ {
//				for n := 0; n < len(seqs[j]); n++ {
//					if search2(gPara, seqs[i], seqs[j], 0, 0, m, n, seqDtls[i], seqDtls[j], gState.InnerFeats[i], gState.InnerFeats[j]) {
//						ok = true
//						gSearch1_greedy(gPara, seqs, seqDtls, []int{i, j})
//					}
//				}
//			}
//		}
//	}
//	return
//}
//
//func gSearch3_greedy(gPara *GPara, gState *GState, seqs [][]int, seqDtls [][]float64) (ok bool) {
//	for i := 0; i < len(seqs); i++ {
//		for j := i + 1; j < len(seqs); j++ {
//			for m := 0; m < len(seqs[i]); m++ {
//				for n := 0; n < len(seqs[j]); n++ {
//					ok = ok || search2(gPara, seqs[i], seqs[j], 0, 0, m, n, seqDtls[i], seqDtls[j], gState.InnerFeats[i], gState.InnerFeats[j])
//				}
//			}
//		}
//	}
//	return
//}

// **************************************** old version **********************************************************

//func SelectSeqs(k1 int, feats [][]float64) (k2 int) {
//	kCent := feats[k1]
//	minD := math.MaxFloat64
//	k2 = k1
//	for i := 0; i < len(feats); i++ {
//		if i == k1 {
//			continue
//		}
//		dist := GreatCircleDistance(kCent, feats[i])
//		if dist < minD {
//			minD = dist
//			k2 = i
//		}
//	}
//	return
//}

// apply ops
func Method01(optEndTime int64, gPara *GPara, gState *GState) {
	//cntIter := []int{2e4, 2e4}
	if time.Now().Unix() < optEndTime-5 {
		search1Weighted(gPara, gState)
		search3Weighted(gPara, gState)
		//search4Weighted(gPara, gState)
	}

	if time.Now().Unix() < optEndTime-5 {
		search1Greedy(gPara, gState)
		search3Greedy(gPara, gState)
		//search4Greedy(gPara, gState)
	}

	for s := 0; s < len(gState.InnerSeqDtls); s++ {
		gState.InnerSeqDtls[s][4], gState.InnerSeqDtls[s][5] = GetSeqMFCostByDist(gPara, gState.InnerAsgmts[s], gState.InnerSeqDtls[s][0])
	}
}

func search1Greedy(gPara *GPara, gState *GState) {
	for i := 0; i < len(gState.InnerSeqs); i++ {
		//preDist := gState.InnerSeqDtls[i][0]
		//fmt.Println("search4-", i, ":", gState.InnerSeqDtls[i])
		//fmt.Println(i, ":", gState.InnerSeqDtls[i])

		for i1 := 0; i1 < len(gState.InnerSeqs[i])-1; i1++ {
			for i2 := i1 + 1; i2 < len(gState.InnerSeqs[i]); i2++ {
				search1(gPara, gState.InnerSeqs[i], gState.InnerAsgmts[i], i1, i2, gState.InnerSeqDtls[i])
				//if ok1 {
				//	fmt.Println("search1-", i, "success", i1, " ", i2, ":", gState.InnerSeqDtls[i])
				//	postDist := gState.InnerSeqDtls[i][0]
				//	improvePerc := (preDist - postDist) / preDist * 100
				//	fmt.Println("search1-", i, "improvePerc:", improvePerc)
				//	//if improvePerc == 0 {
				//	//	fmt.Println("stop")
				//	//}
				//}
			}
		}
		//fmt.Println("search3-", i, ":", gState.InnerSeqDtls[i])
		//postDist := getObj()
		//improvePerc := (preDist - postDist) / preDist * 100
		//fmt.Println("search3-", i, "improvePerc:", improvePerc)
	}
}

func search3Greedy(gPara *GPara, gState *GState) {
	for i := 0; i < len(gState.InnerSeqs); i++ {
		//preDist := gState.InnerSeqDtls[i][0]
		//fmt.Println("search4-", i, ":", gState.InnerSeqDtls[i])
		//fmt.Println(i, ":", gState.InnerSeqDtls[i])

		for i1 := 0; i1 < len(gState.InnerSeqs[i])-1; i1++ {
			for i2 := i1 + 1; i2 < len(gState.InnerSeqs[i]); i2++ {
				search3(gPara, gState.InnerSeqs[i], gState.InnerAsgmts[i], i1, i2, gState.InnerSeqDtls[i])
				//if ok1 {
				//	fmt.Println("search3-", i, "success", i1, " ", i2, ":", gState.InnerSeqDtls[i])
				//	postDist := gState.InnerSeqDtls[i][0]
				//	improvePerc := (preDist - postDist) / preDist * 100
				//	fmt.Println("search3-", i, "improvePerc:", improvePerc)
				//	//if improvePerc == 0 {
				//	//	fmt.Println("stop")
				//	//}
				//}
			}
		}
		//fmt.Println("search3-", i, ":", gState.InnerSeqDtls[i])
		//postDist := getObj()
		//improvePerc := (preDist - postDist) / preDist * 100
		//fmt.Println("search3-", i, "improvePerc:", improvePerc)
	}
}

func search4Greedy(gPara *GPara, gState *GState) {
	for i := 0; i < len(gState.InnerSeqs); i++ {
		//preDist := gState.InnerSeqDtls[i][0]
		//fmt.Println("search4-", i, ":", gState.InnerSeqDtls[i])
		//fmt.Println(i, ":", gState.InnerSeqDtls[i])

		for i1 := 0; i1 < len(gState.InnerSeqs[i])-1; i1++ {
			for i2 := i1 + 1; i2 < len(gState.InnerSeqs[i]); i2++ {
				search4(gPara, gState.InnerSeqs[i], gState.InnerAsgmts[i], i1, i2, gState.InnerSeqDtls[i], gState.InnerFeats[i])
				//if ok4 {
				//	fmt.Println("search4-", i, "success", i1, " ", i2, ":", gState.InnerSeqDtls[i])
				//	postDist := gState.InnerSeqDtls[i][0]
				//	improvePerc := (preDist - postDist) / preDist * 100
				//	fmt.Println("search4-", i, "improvePerc:", improvePerc)
				//	//if improvePerc == 0 {
				//	//	fmt.Println("stop")
				//	//}
				//}
			}
		}
		//fmt.Println("search3-", i, ":", gState.InnerSeqDtls[i])
		//postDist := getObj()
		//improvePerc := (preDist - postDist) / preDist * 100
		//fmt.Println("search3-", i, "improvePerc:", improvePerc)
	}
}

// **************************************** search **********************************************************
// swap two gPara.Nodes in a seq
func search1(gPara *GPara, s []int, r int, i1, i2 int, d []float64) bool {
	if i1 < 0 {
		i1 = rand.Intn(len(s) - 1)
		i2 = i1 + 1 + rand.Intn(len(s)-1-i1)
	}

	delta, newD, ok := dMove01(gPara, s, i1, i2, r, d, nil)

	if ok && delta < 0 {
		s[i1], s[i2] = s[i2], s[i1]
		for i := 0; i < len(d); i++ {
			d[i] = newD[i]
		}
	}
	return ok && delta < 0
}

// swap two gPara.Nodes between two different seqs
func search2(gPara *GPara, s1 []int, s2 []int, r1 int, r2 int, i1, i2 int, d1 []float64, d2 []float64, f1, f2 []float64) bool {
	if i1 < 0 {
		i1 = rand.Intn(len(s1) - 1)
		i2 = rand.Intn(len(s2) - 1)
	}

	delta, ok, newD1, newD2 := dMove02(gPara, s1, s2, i1, i2, r1, r2, d1, d2, f1, f2)
	if ok && delta < 0 {
		//update cent
		lat1 := (f1[0]*float64(len(s1)) - gPara.Nodes[gPara.TaskLoc[s1[i1]]][0] + gPara.Nodes[gPara.TaskLoc[s2[i2]]][0]) / float64(len(s1))
		lng1 := (f1[1]*float64(len(s1)) - gPara.Nodes[gPara.TaskLoc[s1[i1]]][1] + gPara.Nodes[gPara.TaskLoc[s2[i2]]][1]) / float64(len(s1))

		lat2 := (f2[0]*float64(len(s2)) - gPara.Nodes[gPara.TaskLoc[s2[i2]]][0] + gPara.Nodes[gPara.TaskLoc[s1[i1]]][0]) / float64(len(s2))
		lng2 := (f2[1]*float64(len(s2)) - gPara.Nodes[gPara.TaskLoc[s2[i2]]][1] + gPara.Nodes[gPara.TaskLoc[s1[i1]]][1]) / float64(len(s2))

		f1[0] = lat1
		f1[1] = lng1
		f2[0] = lat2
		f2[1] = lng2

		s1[i1], s2[i2] = s2[i2], s1[i1]
		for i := 0; i < len(d1); i++ {
			d1[i] = newD1[i]
			d2[i] = newD2[i]
		}

	}
	return ok && delta < 0
}

// move i1 to after i2 in s
func search3(gPara *GPara, s []int, r int, i1, i2 int, d []float64) bool {
	if len(s) < 3 {
		return false
	}
	if i1 < 0 {
		i1 = rand.Intn(len(s) - 1)
		i2 = i1 + 1 + rand.Intn(len(s)-1-i1)
		if rand.Intn(2) > 0 {
			i1, i2 = i2, i1
		}
	}

	delta, ok, newD := dMove03(gPara, s, i1, i2, r, d)

	if ok && delta < 0 {
		if i1 < i2 {
			//fmt.Println(s[i1], s[i2])
			tmp := s[i1]
			for i := i1; i < i2; i++ {
				s[i] = s[i+1]
			}
			s[i2] = tmp
			//fmt.Println(s[i2], s[i2-1])
		} else {
			//fmt.Println(s[i1], s[i2])
			//fmt.Println("i1:", i1, "i2:", i2)
			tmp := s[i1]
			for i := i1; i > i2+1; i-- {
				s[i] = s[i-1]
			}
			s[i2+1] = tmp
			//fmt.Println(s[i2+1], s[i2])
		}
		for i := 0; i < len(d); i++ {
			d[i] = newD[i]
		}
	}
	return ok && delta < 0
}

// apply 2-opt in s
func search4(gPara *GPara, s []int, r int, i1, i2 int, d []float64, f []float64) bool {
	if i1 < 0 {
		i1 = rand.Intn(len(s) - 3)
		i2 = i1 + 2 + rand.Intn(len(s)-1-i1-2)
	}

	delta, ok, newD := dMove04(gPara, s, i1, i2, r, d, f)

	if ok && delta < 0 {
		//fmt.Println("before search4 opt:", getSeqObj(s, r), "detail:", d, "seq:", s)
		reverse(s, i1, i2)
		//fmt.Println("reverse:", i1, "--", i2)
		for i := 0; i < len(d); i++ {
			d[i] = newD[i]
		}
		//dist := getSeqDist(s, r)
		//if d[0] != dist {
		//	fmt.Println("距离不一致:", "dist:", dist, "  d[0]:", d[0])
		//}
		//fmt.Println("after search4 opt:", getSeqObj(s, r), "detail:", d, "seq:", s)
	}
	return ok && delta < 0
}

// **************************************** check delta **********************************************************
// swap two gPara.Nodes s1[i1] and s1[i2]
func dMove01(gPara *GPara, s1 []int, i1, i2 int, r1 int, d1 []float64, feat1 []float64) (delta float64, newD []float64, ok bool) {
	if i1 > i2 {
		i1, i2 = i2, i1
	}

	n1 := make([]int, 3)
	n2 := make([]int, 3)
	if i1 > 0 {
		n1[0] = gPara.TaskLoc[s1[i1-1]]
	}
	n1[1] = gPara.TaskLoc[s1[i1]]
	n1[2] = gPara.TaskLoc[s1[i1+1]]
	n2[0] = gPara.TaskLoc[s1[i2-1]]
	n2[1] = gPara.TaskLoc[s1[i2]]
	if i2 < len(s1)-1 {
		n2[2] = gPara.TaskLoc[s1[i2+1]]
	}

	// acceptance
	ok, newD, delta = checkMove01(gPara, i1, i2, r1, n1, n2, d1)

	return
}

func checkMove01(gPara *GPara, i1, i2 int, r1 int, n1 []int, n2 []int, d []float64) (ok bool, newD []float64, dtD float64) {
	newD = make([]float64, len(d))
	dtD = 0.0
	var dtDrt float64 = 0.0
	//distance & duration
	if i1+1 < i2 {
		dtD = -gPara.Cost[r1][0][n1[0]][n1[1]] - gPara.Cost[r1][0][n1[1]][n1[2]] - gPara.Cost[r1][0][n2[0]][n2[1]] - gPara.Cost[r1][0][n2[1]][n2[2]] +
			gPara.Cost[r1][0][n1[0]][n2[1]] + gPara.Cost[r1][0][n2[1]][n1[2]] + gPara.Cost[r1][0][n2[0]][n1[1]] + gPara.Cost[r1][0][n1[1]][n2[2]]

		dtDrt = -gPara.Cost[r1][1][n1[0]][n1[1]] - gPara.Cost[r1][1][n1[1]][n1[2]] - gPara.Cost[r1][1][n2[0]][n2[1]] - gPara.Cost[r1][1][n2[1]][n2[2]] +
			gPara.Cost[r1][1][n1[0]][n2[1]] + gPara.Cost[r1][1][n2[1]][n1[2]] + gPara.Cost[r1][1][n2[0]][n1[1]] + gPara.Cost[r1][1][n1[1]][n2[2]]
	} else if i1 < i2 {
		dtD = -gPara.Cost[r1][0][n1[0]][n1[1]] - gPara.Cost[r1][0][n1[1]][n1[2]] - gPara.Cost[r1][0][n2[1]][n2[2]] +
			gPara.Cost[r1][0][n1[0]][n2[1]] + gPara.Cost[r1][0][n2[1]][n1[1]] + gPara.Cost[r1][0][n1[1]][n2[2]]

		dtDrt = -gPara.Cost[r1][1][n1[0]][n1[1]] - gPara.Cost[r1][1][n1[1]][n1[2]] - gPara.Cost[r1][1][n2[1]][n2[2]] +
			gPara.Cost[r1][1][n1[0]][n2[1]] + gPara.Cost[r1][1][n2[1]][n1[1]] + gPara.Cost[r1][1][n1[1]][n2[2]]
	}

	newD[0] = d[0] + dtD
	//if (d[0]-newD[0])/d[0]*100 >= 30 {
	//	fmt.Println("stop")
	//}
	newD[3] = d[3] + dtDrt
	newD[1] = d[1]
	newD[2] = d[2]

	dD := (newD[0] <= gPara.CapRes[r1][1]) && (newD[0] >= gPara.CapRes[r1][0])
	dDrt := (newD[3] <= gPara.CapRes[r1][7]) && (newD[3] >= gPara.CapRes[r1][6])

	ok = dD && dDrt
	return
}

func checkMove03(gPara *GPara, s1 []int, r1 int, i1 int, i2 int, n1 []int, n2 []int, d []float64) (dtD float64, ok bool, newD []float64) {
	newD = make([]float64, len(d))
	dtD = 0.0
	var dtDrt float64 = 0.0
	//distance & duration
	dtD = -gPara.Cost[r1][0][n1[0]][n1[1]] - gPara.Cost[r1][0][n1[1]][n1[2]] - gPara.Cost[r1][0][n2[0]][n2[1]] +
		gPara.Cost[r1][0][n1[0]][n1[2]] + gPara.Cost[r1][0][n2[0]][n1[1]] + gPara.Cost[r1][0][n1[1]][n2[1]]

	dtDrt = -gPara.Cost[r1][1][n1[0]][n1[1]] - gPara.Cost[r1][1][n1[1]][n1[2]] - gPara.Cost[r1][1][n2[0]][n2[1]] +
		gPara.Cost[r1][1][n1[0]][n1[2]] + gPara.Cost[r1][1][n2[0]][n1[1]] + gPara.Cost[r1][1][n1[1]][n2[1]]

	newD[0] = d[0] + dtD
	//if (d[0]-newD[0])/d[0]*100 >= 30 {
	//	fmt.Println("stop")
	//}
	newD[3] = d[3] + dtDrt
	newD[1] = d[1]
	newD[2] = d[2]

	//if (d[0]-newD[0])/d[0]*100 > 30 {
	//	fmt.Println("大于30%")
	//}

	dD := (newD[0] <= gPara.CapRes[r1][1]) && (newD[0] >= gPara.CapRes[r1][0])
	dDrt := (newD[3] <= gPara.CapRes[r1][7]) && (newD[3] >= gPara.CapRes[r1][6])

	ok = dD && dDrt
	return
}

//d[0] -- distance
//d[1] -- parcel
//d[2] -- weight
//d[3] -- duration

func checkMove02(gPara *GPara, s1, s2 []int, i1, i2 int, r1, r2 int, d1, d2 []float64, n1 []int, n2 []int) (delta float64, ok bool, newd1, newd2 []float64) {
	newd1 = make([]float64, len(d1))
	newd2 = make([]float64, len(d2))
	//dt1 Parcel
	dtP1 := -gPara.CapTask[s1[i1]][0] + gPara.CapTask[s2[i2]][0]
	//dt1 Weight
	dtW1 := -gPara.CapTask[s1[i1]][1] + gPara.CapTask[s2[i2]][1]
	//dt1 Distance
	dtD1 := -gPara.Cost[r1][0][n1[0]][n1[1]] - gPara.Cost[r1][0][n1[1]][n1[2]] + gPara.Cost[r1][0][n1[0]][n2[1]] + gPara.Cost[r1][0][n2[1]][n1[2]]
	//dt1 Duration
	dtDrt1 := -gPara.Cost[r1][1][n1[0]][n1[1]] - gPara.Cost[r1][1][n1[1]][n1[2]] - gPara.CapTask[s1[i1]][2] + gPara.Cost[r1][1][n1[0]][n2[1]] + gPara.Cost[r1][1][n2[1]][n1[2]] + gPara.CapTask[s2[i2]][2]

	// check res cont
	newd1[1] = d1[1] + dtP1
	newd1[2] = d1[2] + dtW1
	newd1[0] = d1[0] + dtD1
	newd1[3] = d1[3] + dtDrt1

	//dt2 Parcel
	dtP2 := -gPara.CapTask[s2[i2]][0] + gPara.CapTask[s1[i1]][0]
	//dt2 Weight
	dtW2 := -gPara.CapTask[s2[i2]][1] + gPara.CapTask[s1[i1]][1]
	//dt2 Distance
	dtD2 := -gPara.Cost[r2][0][n2[0]][n2[1]] - gPara.Cost[r2][0][n2[1]][n2[2]] + gPara.Cost[r2][0][n2[0]][n1[1]] + gPara.Cost[r2][0][n1[1]][n2[2]]
	//dt2 Duration
	dtDrt2 := -gPara.Cost[r2][1][n2[0]][n2[1]] - gPara.Cost[r2][1][n2[1]][n2[2]] - gPara.CapTask[s2[i2]][2] + gPara.Cost[r2][1][n2[0]][n1[1]] + gPara.Cost[r2][1][n1[1]][n2[2]] + gPara.CapTask[s1[i1]][2]

	// check res cont
	newd2[1] = d2[1] + dtP2
	newd2[2] = d2[2] + dtW2
	newd2[0] = d2[0] + dtD2
	newd2[3] = d2[3] + dtDrt2

	d1P := (newd1[1] <= gPara.CapRes[r1][3]) && (newd1[1] >= gPara.CapRes[r1][2])
	d1W := (newd1[2] <= gPara.CapRes[r1][5]) && (newd1[2] >= gPara.CapRes[r1][4])
	d1D := (newd1[0] <= gPara.CapRes[r1][1]) && (newd1[0] >= gPara.CapRes[r1][0])
	d1Drt := (newd1[3] <= gPara.CapRes[r1][7]) && (newd1[3] >= gPara.CapRes[r1][6])
	d2P := (newd2[1] <= gPara.CapRes[r2][3]) && (newd2[1] >= gPara.CapRes[r2][2])
	d2W := (newd2[2] <= gPara.CapRes[r2][5]) && (newd2[2] >= gPara.CapRes[r2][4])
	d2D := (newd2[0] <= gPara.CapRes[r2][1]) && (newd2[0] >= gPara.CapRes[r2][0])
	d2Drt := (newd2[3] <= gPara.CapRes[r2][7]) && (newd2[3] >= gPara.CapRes[r2][6])

	ok = d1P && d1W && d1D && d1Drt && d2P && d2W && d2D && d2Drt
	delta = dtD1 + dtD2
	return
}

func dMove02(gPara *GPara, s1, s2 []int, i1, i2 int, r1, r2 int, d1, d2 []float64, feat1, feat2 []float64) (delta float64, ok bool, newD1, newD2 []float64) {
	n1 := make([]int, 3)
	n2 := make([]int, 3)

	if i1 > 0 {
		n1[0] = gPara.TaskLoc[s1[i1-1]]
	}
	n1[1] = gPara.TaskLoc[s1[i1]]
	if i1 < len(s1)-1 {
		n1[2] = gPara.TaskLoc[s1[i1+1]]
	}

	if i2 > 0 {
		n2[0] = gPara.TaskLoc[s2[i2-1]]
	}
	n2[1] = gPara.TaskLoc[s2[i2]]
	if i2 < len(s2)-1 {
		n2[2] = gPara.TaskLoc[s2[i2+1]]
	}

	// acceptance
	//校验 parcel 1  weight  2
	delta, ok, newD1, newD2 = checkMove02(gPara, s1, s2, i1, i2, r1, r2, d1, d2, n1, n2)

	return
}

// move i1 after i2 in s1
func dMove03(gPara *GPara, s1 []int, i1, i2 int, r1 int, d1 []float64) (delta float64, ok bool, newD []float64) {
	if i1 == i2+1 {
		return
	}

	n1 := make([]int, 3)
	n2 := make([]int, 2)
	if i1 > 0 {
		n1[0] = gPara.TaskLoc[s1[i1-1]]
	}
	n1[1] = gPara.TaskLoc[s1[i1]]
	if i1 < len(s1)-1 {
		n1[2] = gPara.TaskLoc[s1[i1+1]]
	}
	n2[0] = gPara.TaskLoc[s1[i2]]
	if i2 < len(s1)-1 {
		n2[1] = gPara.TaskLoc[s1[i2+1]]
	}

	// acceptance
	delta, ok, newD = checkMove03(gPara, s1, r1, i1, i2, n1, n2, d1)
	return
}

// apply 2-opt in s for (i1,i1+1) (i2,i2+1)
func dMove04(gPara *GPara, s1 []int, i1 int, i2 int, r1 int, d1 []float64, feat1 []float64) (delta float64, ok bool, newD []float64) {

	n1 := make([]int, 2)
	n2 := make([]int, 2)
	if i1 == 0 {
		n1[0] = 0
	} else {
		n1[0] = gPara.TaskLoc[s1[i1-1]]
	}
	n1[1] = gPara.TaskLoc[s1[i1]]
	n2[0] = gPara.TaskLoc[s1[i2]]
	if i2 == len(s1)-1 {
		n2[1] = 0
	} else {
		n2[1] = gPara.TaskLoc[s1[i2+1]]
	}

	// acceptance
	delta, ok, newD = checkMove04(gPara, s1, i1, i2, r1, n1, n2, d1)

	return
}

func checkMove04(gPara *GPara, s1 []int, i1 int, i2 int, r1 int, n1 []int, n2 []int, d []float64) (dtD float64, ok bool, newD []float64) {
	newD = make([]float64, len(d))
	dtD = 0.0
	var dtDrt float64 = 0.0

	// first assume d12 = d21 for simplicity, then check full path
	dtD = -gPara.Cost[r1][0][n1[0]][n1[1]] - gPara.Cost[r1][0][n2[0]][n2[1]] + gPara.Cost[r1][0][n1[0]][n2[0]] + gPara.Cost[r1][0][n1[1]][n2[1]]
	dtDrt = -gPara.Cost[r1][1][n1[0]][n1[1]] - gPara.Cost[r1][1][n2[0]][n2[1]] + gPara.Cost[r1][1][n1[0]][n2[0]] + gPara.Cost[r1][1][n1[1]][n2[1]]
	//newD[0] = d[0] + dtD
	//newD[3] = d[3] + dtDrt
	// (i1, i2) reverse diff
	//if dtD < 0 {
	for i := i1; i < i2; i++ {
		dtD += -gPara.Cost[r1][0][gPara.TaskLoc[s1[i]]][gPara.TaskLoc[s1[i+1]]] + gPara.Cost[r1][0][gPara.TaskLoc[s1[i+1]]][gPara.TaskLoc[s1[i]]]
		dtDrt += -gPara.Cost[r1][1][gPara.TaskLoc[s1[i]]][gPara.TaskLoc[s1[i+1]]] + gPara.Cost[r1][1][gPara.TaskLoc[s1[i+1]]][gPara.TaskLoc[s1[i]]]
	}
	newD[0] = d[0] + dtD

	//if (d[0]-newD[0])/d[0]*100 >= 30 {
	//	fmt.Println("stop")
	//}
	newD[3] = d[3] + dtDrt
	//}

	newD[1] = d[1]
	newD[2] = d[2]

	dD := (newD[0] <= gPara.CapRes[r1][1]) && (newD[0] >= gPara.CapRes[r1][0])
	dDrt := (newD[3] <= gPara.CapRes[r1][7]) && (newD[3] >= gPara.CapRes[r1][6])

	ok = dD && dDrt
	return
}

// *********************************** greedy swap **********************************************
func Method_tmp(optTime int, seqs [][]int, seqDtls [][]float64, gPara *GPara, gState *GState) {
	var ok bool = true
	var eff1, eff2 bool
	for gIter := 0; gIter < 1 && ok; gIter++ {
		for iter := 0; iter < 1; iter++ {
			eff1 = gSearch1_greedy(gPara, seqs, seqDtls, []int{}) || eff1
			if !eff1 {
				break
			}
		}
		for iter := 0; iter < 1; iter++ {
			eff2 = gSearch2_greedy(gPara, gState, seqs, seqDtls) || eff2
			if !eff2 {
				break
			}
		}
		ok = eff1 || eff2
	}
}

func gSearch1_greedy(gPara *GPara, seqs [][]int, seqDtls [][]float64, seqIdx []int) (ok bool) {
	if len(seqIdx) > 0 {
		for _, i := range seqIdx {
			for m := 0; m < len(seqs[i])-1; m++ {
				for n := m + 1; n < len(seqs[i]); n++ {
					ok = ok || search1(gPara, seqs[i], 0, m, n, seqDtls[i])
				}
			}
		}
	} else {
		for i := 0; i < len(seqs); i++ {
			for m := 0; m < len(seqs[i])-1; m++ {
				for n := m + 1; n < len(seqs[i]); n++ {
					ok = ok || search1(gPara, seqs[i], 0, m, n, seqDtls[i])
				}
			}
		}
	}
	return
}

func gSearch2_greedy(gPara *GPara, gState *GState, seqs [][]int, seqDtls [][]float64) (ok bool) {
	for i := 0; i < len(seqs); i++ {
		for j := i + 1; j < len(seqs); j++ {
			for m := 0; m < len(seqs[i]); m++ {
				for n := 0; n < len(seqs[j]); n++ {
					if search2(gPara, seqs[i], seqs[j], 0, 0, m, n, seqDtls[i], seqDtls[j], gState.InnerFeats[i], gState.InnerFeats[j]) {
						ok = true
						gSearch1_greedy(gPara, seqs, seqDtls, []int{i, j})
					}
				}
			}
		}
	}
	return
}

func gSearch3_greedy(gPara *GPara, gState *GState, seqs [][]int, seqDtls [][]float64) (ok bool) {
	for i := 0; i < len(seqs); i++ {
		for j := i + 1; j < len(seqs); j++ {
			for m := 0; m < len(seqs[i]); m++ {
				for n := 0; n < len(seqs[j]); n++ {
					ok = ok || search2(gPara, seqs[i], seqs[j], 0, 0, m, n, seqDtls[i], seqDtls[j], gState.InnerFeats[i], gState.InnerFeats[j])
				}
			}
		}
	}
	return
}
