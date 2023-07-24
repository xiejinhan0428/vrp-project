package solver

func search4GreedyForInit(gPara *GPara, s []int, capUse *CapUse, r int, avgCent []float64) {
	//preDist := innerSeqDtls[i][0]
	//fmt.Println("search4-", i, ":", innerSeqDtls[i])
	//fmt.Println(i, ":", innerSeqDtls[i])

	for i1 := 0; i1 < len(s)-1; i1++ {
		for i2 := i1 + 1; i2 < len(s); i2++ {
			search4(gPara, s, r, i1, i2, capUse.F, avgCent)
			//if ok4 {
			//	fmt.Println("search4-", i, "success", i1, " ", i2, ":", innerSeqDtls[i])
			//	postDist := innerSeqDtls[i][0]
			//	improvePerc := (preDist - postDist) / preDist * 100
			//	fmt.Println("search4-", i, "improvePerc:", improvePerc)
			//	//if improvePerc == 0 {
			//	//	fmt.Println("stop")
			//	//}
			//}
		}
	}
	//fmt.Println("search3-", i, ":", innerSeqDtls[i])
	//postDist := getObj()
	//improvePerc := (preDist - postDist) / preDist * 100
	//fmt.Println("search3-", i, "improvePerc:", improvePerc)
}

func search3GreedyForInit(gPara *GPara, s []int, capUse *CapUse, r int) {
	//preDist := innerSeqDtls[i][0]
	//fmt.Println("search4-", i, ":", innerSeqDtls[i])
	//fmt.Println(i, ":", innerSeqDtls[i])

	for i1 := 0; i1 < len(s)-1; i1++ {
		for i2 := i1 + 1; i2 < len(s); i2++ {
			search3(gPara, s, r, i1, i2, capUse.F)
			//if ok1 {
			//	fmt.Println("search3-", i, "success", i1, " ", i2, ":", innerSeqDtls[i])
			//	postDist := innerSeqDtls[i][0]
			//	improvePerc := (preDist - postDist) / preDist * 100
			//	fmt.Println("search3-", i, "improvePerc:", improvePerc)
			//	//if improvePerc == 0 {
			//	//	fmt.Println("stop")
			//	//}
			//}
		}
	}
	//fmt.Println("search3-", i, ":", innerSeqDtls[i])
	//postDist := getObj()
	//improvePerc := (preDist - postDist) / preDist * 100
	//fmt.Println("search3-", i, "improvePerc:", improvePerc)

}

func search1GreedyForInit(gPara *GPara, s []int, capUse *CapUse, r int) {
	//preDist := innerSeqDtls[i][0]
	//fmt.Println("search4-", i, ":", innerSeqDtls[i])
	//fmt.Println(i, ":", innerSeqDtls[i])

	for i1 := 0; i1 < len(s)-1; i1++ {
		for i2 := i1 + 1; i2 < len(s); i2++ {
			search1(gPara, s, r, i1, i2, capUse.F)
			//if ok1 {
			//	fmt.Println("search3-", i, "success", i1, " ", i2, ":", innerSeqDtls[i])
			//	postDist := innerSeqDtls[i][0]
			//	improvePerc := (preDist - postDist) / preDist * 100
			//	fmt.Println("search3-", i, "improvePerc:", improvePerc)
			//	//if improvePerc == 0 {
			//	//	fmt.Println("stop")
			//	//}
			//}
		}
	}
	//fmt.Println("search3-", i, ":", innerSeqDtls[i])
	//postDist := getObj()
	//improvePerc := (preDist - postDist) / preDist * 100
	//fmt.Println("search3-", i, "improvePerc:", improvePerc)

}
