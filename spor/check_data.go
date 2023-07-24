package solver

const (
	contraintNum = 8
)

func CheckDataPro(solveInput *SolverInput) (flagSolver int) {
	if (solveInput.Nodes == nil) || (solveInput.Tasks == nil) || (solveInput.Resources == nil) || (solveInput.EdgeInfos == nil) {
		flagSolver = -101
		return flagSolver
	}
	if solveInput.Obj < 0 {
		solveInput.Obj = 0
	}

	if len(solveInput.EdgeInfos) != len(solveInput.Resources) {
		flagSolver = -102
		return flagSolver
	}
	flagSolver = CheckEdgeInfos(len(solveInput.Nodes), solveInput.EdgeInfos, flagSolver)
	flagSolver = CheckTasks(len(solveInput.Nodes), solveInput.Tasks, flagSolver)
	flagSolver = CheckResources(solveInput.Resources, flagSolver)
	return flagSolver
}

func CheckResources(resources []*Resource, flagSolver int) int {
	for i := 0; i < len(resources); i++ {
		if len(resources[i].Capacity) != contraintNum {
			flagSolver = -103
			return flagSolver
		}
		if resources[i].Quantity < 0 {
			resources[i].Quantity = 0
		}
		//check cont
		for j := 0; j < len(resources[i].Capacity); j++ {
			ok := true
			if j%2 == 0 {
				if ok = resources[i].Capacity[j] <= resources[i].Capacity[j+1]; !ok {
					flagSolver = -104
					return flagSolver
				}
			}
		}
	}
	return flagSolver
}

func CheckEdgeInfos(nodeLength int, edgeInfos []*EdgeInfo, flagSolver int) int {
	for i := 0; i < len(edgeInfos); i++ {
		//check distance
		if len(edgeInfos[i].Distance) != nodeLength {
			flagSolver = -105
			return flagSolver
		}
		for j := 0; j < len(edgeInfos[i].Distance); j++ {
			if len(edgeInfos[i].Distance[j]) != nodeLength {
				flagSolver = -106
				return flagSolver
			}
		}
	}
	return flagSolver
}

func CheckTasks(nodeLength int, tasks []*Task, flagSolver int) int {
	var maxNodeIdx int = 0
	for i := 0; i < len(tasks); i++ {
		if len(tasks[i].ReqCap) != 3 {
			flagSolver = -107
			return flagSolver
		}
		if tasks[i].NodeIdx < 0 {
			flagSolver = -108
			return flagSolver
		}

		if maxNodeIdx < tasks[i].NodeIdx {
			maxNodeIdx = tasks[i].NodeIdx
		}
	}
	if maxNodeIdx >= nodeLength {
		flagSolver = -109
		return flagSolver
	}
	return flagSolver
}

func CheckSeqTasks(state *GState, para *GPara) {
	var taskMap = make(map[int]int)
	for i := 0; i < len(state.BestInnerSeqs); i++ {
		for j := 0; j < len(state.BestInnerSeqs[i]); j++ {
			if _, ok := taskMap[state.BestInnerSeqs[i][j]]; !ok {
				taskMap[state.BestInnerSeqs[i][j]] = 1
			} else {
				taskMap[state.BestInnerSeqs[i][j]] += 1
			}
		}
	}

	for i := 0; i < len(state.BestInnerUnasgTasks); i++ {
		if _, ok := taskMap[state.BestInnerUnasgTasks[i]]; !ok {
			taskMap[state.BestInnerUnasgTasks[i]] = 1
		} else {
			taskMap[state.BestInnerUnasgTasks[i]] += 1
		}
	}

	for i := 0; i < len(state.BestInnerInfeaTasks); i++ {
		if _, ok := taskMap[state.BestInnerInfeaTasks[i]]; !ok {
			taskMap[state.BestInnerInfeaTasks[i]] = 1
		} else {
			taskMap[state.BestInnerInfeaTasks[i]] += 1
		}
	}

	if len(taskMap) != para.NTask {
		state.Flag = -110
		//fmt.Println("结果task长度异常")
	}

	for _, v := range taskMap {
		if v > 1 {
			state.Flag = -111
			//fmt.Println("出现重复task id")
		}
	}
}

func CheckSingleSeqCont(sIdx int, para *GPara, state *GState) (ok bool) {
	ok = true
	cont0 := state.InnerSeqDtls[sIdx][0] >= para.CapRes[state.InnerAsgmts[sIdx]][0]
	cont1 := state.InnerSeqDtls[sIdx][0] <= para.CapRes[state.InnerAsgmts[sIdx]][1]
	cont2 := state.InnerSeqDtls[sIdx][1] >= para.CapRes[state.InnerAsgmts[sIdx]][2]
	cont3 := state.InnerSeqDtls[sIdx][1] <= para.CapRes[state.InnerAsgmts[sIdx]][3]
	cont4 := state.InnerSeqDtls[sIdx][2] >= para.CapRes[state.InnerAsgmts[sIdx]][4]
	cont5 := state.InnerSeqDtls[sIdx][2] <= para.CapRes[state.InnerAsgmts[sIdx]][5]
	cont6 := state.InnerSeqDtls[sIdx][3] >= para.CapRes[state.InnerAsgmts[sIdx]][6]
	cont7 := state.InnerSeqDtls[sIdx][3] <= para.CapRes[state.InnerAsgmts[sIdx]][7]

	ok = cont0 && cont1 && cont2 && cont3 && cont4 && cont5 && cont6 && cont7

	return
}

//capUse.F[0] -- distance
//capUse.F[1] -- parcel
//capUse.F[2] -- weight
//capUse.F[3] -- duration
func CheckSeqCont2(para *GPara, state *GState) (ok bool) {
	ok = true
	//if len(state.BestInnerAsgmts) == 0 {
	//	fmt.Println("车辆长度为0！！")
	//}
	bestInnerSeqDtls := GenerateSeqDtls(state.BestInnerSeqs, state.BestInnerAsgmts, para)
	for i := 0; i < len(state.BestInnerSeqDtls); i++ {
		cont0 := bestInnerSeqDtls[i][0] >= para.CapRes[state.BestInnerAsgmts[i]][0]
		cont1 := bestInnerSeqDtls[i][0] <= para.CapRes[state.BestInnerAsgmts[i]][1]
		cont2 := bestInnerSeqDtls[i][1] >= para.CapRes[state.BestInnerAsgmts[i]][2]
		cont3 := bestInnerSeqDtls[i][1] <= para.CapRes[state.BestInnerAsgmts[i]][3]
		cont4 := bestInnerSeqDtls[i][2] >= para.CapRes[state.BestInnerAsgmts[i]][4]
		cont5 := bestInnerSeqDtls[i][2] <= para.CapRes[state.BestInnerAsgmts[i]][5]
		cont6 := bestInnerSeqDtls[i][3] >= para.CapRes[state.BestInnerAsgmts[i]][6]
		cont7 := bestInnerSeqDtls[i][3] <= para.CapRes[state.BestInnerAsgmts[i]][7]

		ok = cont0 && cont1 && cont2 && cont3 && cont4 && cont5 && cont6 && cont7
		if !ok {
			state.Flag = -3
			return
		}
	}
	return
}

func CheckAsgts(state *GState) {
	if len(state.BestInnerSeqs) != len(state.BestInnerAsgmts) {
		state.Flag = -112
		//fmt.Println("资源数量不一致！！")
	}
	for i := 0; i < len(state.BestInnerSeqs); i++ {
		if len(state.BestInnerSeqs[i]) == 0 {
			state.Flag = -113
			//fmt.Println("出现空线路！！")
		}
	}
}

func CheckDataPost(state *GState, para *GPara) {
	//验证 1、未丢task 2、无重复task 3、约束 4、资源使用情况
	CheckAsgts(state)
	if state.Flag < 0 {
		return
	}
	CheckSeqTasks(state, para)
	if state.Flag < 0 {
		return
	}
	//CheckSeqCont(seqDtl)
	CheckSeqCont2(para, state)
}

func CheckSeqTasksOpt(state *GState, para *GPara) (checkOK bool) {
	checkOK = true
	var taskMap = make(map[int]int)
	for i := 0; i < len(state.InnerSeqs); i++ {
		for j := 0; j < len(state.InnerSeqs[i]); j++ {
			if _, ok := taskMap[state.InnerSeqs[i][j]]; !ok {
				taskMap[state.InnerSeqs[i][j]] = 1
			} else {
				taskMap[state.InnerSeqs[i][j]] += 1
			}
		}
	}

	for i := 0; i < len(state.InnerUnasgTasks); i++ {
		if _, ok := taskMap[state.InnerUnasgTasks[i]]; !ok {
			taskMap[state.InnerUnasgTasks[i]] = 1
		} else {
			taskMap[state.InnerUnasgTasks[i]] += 1
		}
	}

	for i := 0; i < len(state.InnerInfeaTasks); i++ {
		if _, ok := taskMap[state.InnerInfeaTasks[i]]; !ok {
			taskMap[state.InnerInfeaTasks[i]] = 1
		} else {
			taskMap[state.InnerInfeaTasks[i]] += 1
		}
	}

	if len(taskMap) != para.NTask {
		checkOK = false
		//fmt.Println("结果task长度异常")
	}

	for _, v := range taskMap {
		if v > 1 {
			checkOK = false
			//fmt.Println("出现重复task id")
		}
	}
	return
}

func CheckAsgtsOpt(state *GState) (checkOK bool) {
	checkOK = true
	if len(state.InnerSeqs) != len(state.InnerAsgmts) {
		checkOK = false
		//fmt.Println("资源数量不一致！！")
	}
	for i := 0; i < len(state.InnerSeqs); i++ {
		if len(state.InnerSeqs[i]) == 0 {
			checkOK = false
			//fmt.Println("出现空线路！！")
		}
	}
	return
}

func CheckSeqCont2Opt(state *GState, para *GPara) (checkOK bool) {
	checkOK = true
	bestInnerSeqDtls := GenerateSeqDtls(state.BestInnerSeqs, state.BestInnerAsgmts, para)
	for i := 0; i < len(state.SeqDtl); i++ {
		cont0 := bestInnerSeqDtls[i][0] >= para.CapRes[state.BestInnerAsgmts[i]][0]
		cont1 := bestInnerSeqDtls[i][0] <= para.CapRes[state.BestInnerAsgmts[i]][1]
		cont2 := bestInnerSeqDtls[i][1] >= para.CapRes[state.BestInnerAsgmts[i]][2]
		cont3 := bestInnerSeqDtls[i][1] <= para.CapRes[state.BestInnerAsgmts[i]][3]
		cont4 := bestInnerSeqDtls[i][2] >= para.CapRes[state.BestInnerAsgmts[i]][4]
		cont5 := bestInnerSeqDtls[i][2] <= para.CapRes[state.BestInnerAsgmts[i]][5]
		cont6 := bestInnerSeqDtls[i][3] >= para.CapRes[state.BestInnerAsgmts[i]][6]
		cont7 := bestInnerSeqDtls[i][3] <= para.CapRes[state.BestInnerAsgmts[i]][7]

		checkOK = cont0 && cont1 && cont2 && cont3 && cont4 && cont5 && cont6 && cont7
	}
	return
}

func CheckDataOpt(state *GState, para *GPara) (checkOK bool) {
	//验证 1、未丢task 2、无重复task 4、资源使用情况
	checkOK = true
	checkOK1 := CheckSeqTasksOpt(state, para)
	checkOK2 := CheckAsgtsOpt(state)
	checkOK3 := CheckSeqCont2Opt(state, para)
	checkOK = checkOK1 && checkOK2 && checkOK3
	return
}
