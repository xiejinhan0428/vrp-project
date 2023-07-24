package solver

//func getSeqDist(seqs []int) (d float64) {
//	for i := 0; i < len(seqs)-1; i++ {
//		d += cost[0][0][taskLoc[seqs[i]]][taskLoc[seqs[i+1]]]
//	}
//	return
//}
//func getSeqDist(s []int, rIdx int) (d float64) {
//	d = 0.0
//	for i := 0; i < len(s); i++ {
//		if i == 0 {
//			d += cost[rIdx][0][0][taskLoc[s[i]]]
//		} else {
//			d += cost[rIdx][0][taskLoc[s[i-1]]][taskLoc[s[i]]]
//		}
//	}
//	d += cost[rIdx][0][taskLoc[s[len(s)-1]]][0]
//	return
//}

//CapResCost[resIdx][t][0]:LowerBound
//CapResCost[resIdx][t][1]:UpperBound
//CapResCost[resIdx][t][2]:outerK
//CapResCost[resIdx][t][3]:LowerCost
//CapResCost[resIdx][t][4]:UpperCost
//CapResCost[resIdx][t][5]:innerK
func GetSeqsFitCost(seqs [][]int, asgmets []int, gPara *GPara) float64 {
	var totalCost float64 = 0.0
	for i := 0; i < len(seqs); i++ {
		resIdx := asgmets[i]
		totalCost += GetSeqFitCost(gPara, resIdx, GetSeqDistance(seqs[i], i, asgmets, gPara))
	}
	return totalCost
}

func GetSeqsMapCost(seqs [][]int, asgmets []int, gPara *GPara) float64 {
	var totalCost float64 = 0.0
	for i := 0; i < len(seqs); i++ {
		resIdx := asgmets[i]
		totalCost += GetSeqMapCost(gPara, resIdx, GetSeqDistance(seqs[i], i, asgmets, gPara))
	}
	return totalCost
}

func GetFitCost(state *GState, gPara *GPara) float64 {
	var totalCost float64 = 0.0
	for i := 0; i < len(state.InnerSeqs); i++ {
		resIdx := state.InnerAsgmts[i]
		totalCost += GetSeqFitCost(gPara, resIdx, state.InnerSeqDtls[i][0])
		//totalCost += GetSeqFitCost(gPara, resIdx, GetSeqDistance(seqs[i], resIdx, asgmets, gPara))
	}
	return totalCost
}

func GetSeqFitCost(gPara *GPara, resIdx int, dist float64) float64 {
	var seqCost float64 = 0.0
	for t := 0; t < len(gPara.CapResCost[resIdx]); t++ {
		if dist >= gPara.CapResCost[resIdx][t][0] && dist < gPara.CapResCost[resIdx][t][1] {
			seqCost = (dist-gPara.CapResCost[resIdx][t][0])*gPara.CapResCost[resIdx][t][5] + gPara.CapResCost[resIdx][t][3]
			break
		}
	}
	if dist >= gPara.CapResCost[resIdx][len(gPara.CapResCost[resIdx])-1][1] {
		seqCost = gPara.CapResCost[resIdx][len(gPara.CapResCost[resIdx])-1][4]
	}
	return seqCost
}

func GetSeqMFCostByDist(gPara *GPara, resIdx int, dist float64) (float64, float64) {
	var seqFCost, seqMCost float64 = 0.0, 0.0
	for t := 0; t < len(gPara.CapResCost[resIdx]); t++ {
		if dist >= gPara.CapResCost[resIdx][t][0] && dist < gPara.CapResCost[resIdx][t][1] {
			seqFCost = (dist-gPara.CapResCost[resIdx][t][0])*gPara.CapResCost[resIdx][t][5] + gPara.CapResCost[resIdx][t][3]
			seqMCost = gPara.CapResCost[resIdx][t][4]
			break
		}
	}
	if dist >= gPara.CapResCost[resIdx][len(gPara.CapResCost[resIdx])-1][1] {
		seqFCost = gPara.CapResCost[resIdx][len(gPara.CapResCost[resIdx])-1][4]
		seqMCost = gPara.CapResCost[resIdx][len(gPara.CapResCost[resIdx])-1][4]
	}
	return seqFCost, seqMCost
}

func GetSeqMapCost(gPara *GPara, resIdx int, dist float64) float64 {
	var seqCost float64 = 0.0
	for t := 0; t < len(gPara.CapResCost[resIdx]); t++ {
		if dist >= gPara.CapResCost[resIdx][t][0] && dist < gPara.CapResCost[resIdx][t][1] {
			seqCost = gPara.CapResCost[resIdx][t][4]
			break
		}
	}
	if dist >= gPara.CapResCost[resIdx][len(gPara.CapResCost[resIdx])-1][1] {
		seqCost = gPara.CapResCost[resIdx][len(gPara.CapResCost[resIdx])-1][4]
	}
	return seqCost
}

func GetDistObj(state *GState, gPara *GPara) (d float64) {
	d = 0.0
	for i := 0; i < len(state.InnerSeqDtls); i++ {
		d += state.InnerSeqDtls[i][0]
	}
	return
}

func GetMapCostObj(state *GState, gPara *GPara) (c float64) {
	c = 0.0
	for i := 0; i < len(state.InnerSeqDtls); i++ {
		resIdx := state.InnerAsgmts[i]
		for j := 0; j < len(gPara.CapResCost[resIdx]); j++ {
			if state.InnerSeqDtls[i][0] >= gPara.CapResCost[resIdx][j][0] && state.InnerSeqDtls[i][0] < gPara.CapResCost[resIdx][j][1] {
				c += gPara.CapResCost[resIdx][j][4]
				break
			}
		}
		if state.InnerSeqDtls[i][0] >= gPara.CapResCost[resIdx][len(gPara.CapResCost[resIdx])-1][1] {
			c += gPara.CapResCost[resIdx][len(gPara.CapResCost[resIdx])-1][4]
		}
	}
	return
}
func GenerateFeats(gPara *GPara, seq [][]int) [][]float64 {
	// lat & lng float64
	res := [][]float64{}
	for i := 0; i < len(seq); i++ {
		latAll := 0.0
		lngAll := 0.0
		for j := 0; j < len(seq[i]); j++ {
			latAll += gPara.Nodes[gPara.TaskLoc[seq[i][j]]][0]
			lngAll += gPara.Nodes[gPara.TaskLoc[seq[i][j]]][1]
		}
		res = append(res, []float64{latAll / float64(len(seq[i])), lngAll / float64(len(seq[i]))})
	}
	return res
}
func GetTotalDelta(gPara *GPara, oldD, newD, oldFC, newFC, oldMC, newMC float64) float64 {
	var delta float64 = 0.0
	if gPara.Obj == 0 {
		delta = newD - oldD
	} else {
		if newMC-oldMC < 0 {
			delta = newMC - oldMC
		} else if newMC-oldMC == 0 {
			delta = newFC - oldFC
		} else {
			delta = 10000
		}
		if delta == 0 {
			delta = newD - oldD
		}
	}
	return delta
}
