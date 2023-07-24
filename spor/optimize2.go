package solver

func GetWeight1(gPara *GPara, s []int, r int) (weights []float64) {
	weights = make([]float64, len(s))
	if len(s) < 2 {
		return
	}

	for i := 0; i < len(s); i++ {
		var n0, n1, n2 int = 0, 0, 0

		if i == 0 {
			n0 = 0
		} else {
			n0 = gPara.TaskLoc[s[i-1]]
		}
		n1 = gPara.TaskLoc[s[i]]
		if i == len(s)-1 {
			n2 = 0
		} else {
			n2 = gPara.TaskLoc[s[i+1]]
		}

		delta := gPara.Cost[r][0][n0][n1] + gPara.Cost[r][0][n1][n2] - gPara.Cost[r][0][n0][n2]
		if delta < 0 {
			weights[i] = 0
		} else {
			weights[i] = delta
		}
	}
	return
}

func GetWeightedRand(gPara *GPara, s []int, r int, num int) (li []int) {
	weights := GetWeight1(gPara, s, r)
	li = RouletteMultiSelectForFloat(weights, num)
	return
}

func search1Weighted(gPara *GPara, gState *GState) {
	for i := 0; i < len(gState.InnerSeqs); i++ {
		if len(gState.InnerSeqs[i]) < 2 {
			return
		}
		li := GetWeightedRand(gPara, gState.InnerSeqs[i], gState.InnerAsgmts[i], 2)
		if len(li) != 2 {
			return
		}
		var i1, i2 int
		if li[0] < li[1] {
			i1 = li[0]
			i2 = li[1]
		} else {
			i1 = li[1]
			i2 = li[0]
		}
		search1(gPara, gState.InnerSeqs[i], gState.InnerAsgmts[i], i1, i2, gState.InnerSeqDtls[i])
	}
	return
}

func search3Weighted(gPara *GPara, gState *GState) {
	for i := 0; i < len(gState.InnerSeqs); i++ {
		if len(gState.InnerSeqs[i]) < 2 {
			return
		}
		li := GetWeightedRand(gPara, gState.InnerSeqs[i], gState.InnerAsgmts[i], 2)
		if len(li) != 2 {
			return
		}

		var i1, i2 int
		if li[0] < li[1] {
			i1 = li[0]
			i2 = li[1]
		} else {
			i1 = li[1]
			i2 = li[0]
		}
		search3(gPara, gState.InnerSeqs[i], gState.InnerAsgmts[i], i1, i2, gState.InnerSeqDtls[i])
	}
	return
}

func search4Weighted(gPara *GPara, gState *GState) {
	for i := 0; i < len(gState.InnerSeqs); i++ {
		if len(gState.InnerSeqs[i]) < 2 {
			return
		}
		li := GetWeightedRand(gPara, gState.InnerSeqs[i], gState.InnerAsgmts[i], 2)
		if len(li) != 2 {
			return
		}

		var i1, i2 int
		if li[0] < li[1] {
			i1 = li[0]
			i2 = li[1]
		} else {
			i1 = li[1]
			i2 = li[0]
		}
		search4(gPara, gState.InnerSeqs[i], gState.InnerAsgmts[i], i1, i2, gState.InnerSeqDtls[i], gState.InnerFeats[i])
	}
	return
}
