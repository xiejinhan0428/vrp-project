package solver

import (
	"context"
	"fmt"
	"math/rand"
	"runtime/debug"
	"time"
)

const (
	centDisWeight  = 0.1
	graphDisWeight = 0.5
	GaAddNum       = 2  //遗传算法变异失败容忍次数
	InitGaAddNum   = 3  //遗传算法初始种群添加初始解序列轮次，不能超过5
	LoopGaAddNum   = 1  //遗传算法变异失败后重添加初始解序列轮次，不能超过5
	MaxMutationNum = 10 //最大遗传算法变异成功轮次
	CostCoef       = 10000
)

func Solve2(solverInput *SolverInput) (solverOutput *SolverOutput) {
	var flagSolver int = 666
	solverStartTime := time.Now().Unix()
	solverOutput = &SolverOutput{
		Asgmts:     nil,
		Seqs:       nil,
		InfeaTasks: nil,
		UnasgTasks: nil,
		Flag:       666,
	}
	defer func() {
		if e := recover(); e != nil {
			//fmt.Println("Task Id:", solverInput.TaskId, "-- Error:", e)
			solverOutput.Flag = -2
			solverOutput.Seqs = nil
			solverOutput.InfeaTasks = nil
			solverOutput.Asgmts = nil
			solverOutput.UnasgTasks = nil
			solverOutput.PanicMessage = string(debug.Stack())
			//err = errors.New("unexpected error!")
		}
	}()
	whetherOverTime := false
	flagSolver = CheckDataPro(solverInput)
	//flag = -1
	if flagSolver < 0 {
		solverOutput.Flag = flagSolver
		return solverOutput
	}
	gFun := Config2(solverInput.Obj)
	gPara := InitParaS2(solverInput) //数据转换
	gState := GState{}
	//InitTask2(0.3, 0.7)
	//sTime := time.Now().Unix()
	InitSols(solverStartTime+100, &gState, &gFun, &gPara)
	//eTime := time.Now().Unix()
	//initDur := eTime - sTime
	//fmt.Println("initDur:", initDur)
	//seqDtls := CheckSeqsDtl()
	//CheckDataPost(seqDtls)
	//initDist := GetDistObj(&gState, &gPara)
	//fmt.Printf("init dist:%f\n", initDist)
	tmpGState := GState{}
	if len(gState.InnerAsgmts) != 0 {
		initEndTime := time.Now().Unix()
		//beforeDur := int(initEndTime - solverStartTime)
		optTime := int(solverInput.CalTime-initEndTime) - 10
		//fmt.Println("opt time:", optTime)
		//if !ControlSolverTime(optTime-5, solverOutput, &gState, &gPara, &gFun) {
		//	whetherOverTime = true
		//}
		gState.InnerSeqs, gState.InnerAsgmts, gState.InnerUnasgTasks, gState.InnerInfeaTasks, gState.InnerSeqDtls = Optimize(optTime, &gState, &gPara, &gFun)
		if solverOutput.Flag == -2 {
			return solverOutput
		}
		//optEndTime := time.Now().Unix()
		//optCostTime := int(optEndTime - initEndTime)
		//fmt.Println("opt cost time:", optCostTime)
		tmpGState.BestInnerSeqs, tmpGState.BestInnerAsgmts, tmpGState.BestInnerUnasgTasks, tmpGState.BestInnerInfeaTasks = CopyResult(&gState)
		tmpGState.ResUse = CopySliceInt(gState.ResUse)
		CheckDataPost(&tmpGState, &gPara)
		if tmpGState.Flag >= 0 {
			tmpGState.BestInnerSeqDtls = GenerateSeqDtls(tmpGState.BestInnerSeqs, tmpGState.BestInnerAsgmts, &gPara)
			var valueA, valueB, delta float64
			if gPara.Obj == 0 {
				for i := 0; i < len(tmpGState.BestInnerSeqDtls); i++ {
					valueA += tmpGState.BestInnerSeqDtls[i][0]
				}
				for i := 0; i < len(gState.BestInnerSeqDtls); i++ {
					valueB += gState.BestInnerSeqDtls[i][0]
				}
				delta = valueA - valueB
			} else {
				var oldMC, newMC, oldFC, newFC float64 = 0.0, 0.0, 0.0, 0.0
				for i := 0; i < len(tmpGState.BestInnerSeqDtls); i++ {
					newMC += tmpGState.BestInnerSeqDtls[i][5]
					newFC += tmpGState.BestInnerSeqDtls[i][4]
				}
				for i := 0; i < len(gState.BestInnerSeqDtls); i++ {
					oldMC += gState.BestInnerSeqDtls[i][5]
					oldFC += gState.BestInnerSeqDtls[i][4]
				}
				deltaMC := newMC - oldMC
				if deltaMC < 0 {
					delta = deltaMC
				} else if deltaMC == 0 {
					delta = newFC - oldFC
				} else {
					delta = 10000
				}
			}
			if delta < 0 && len(tmpGState.InnerUnasgTasks) <= len(gState.BestInnerUnasgTasks) {
				gState.BestInnerSeqs, gState.BestInnerAsgmts, gState.BestInnerUnasgTasks, gState.BestInnerInfeaTasks, gState.BestInnerSeqDtls = tmpGState.BestInnerSeqs, tmpGState.BestInnerAsgmts, tmpGState.BestInnerUnasgTasks, tmpGState.BestInnerInfeaTasks, tmpGState.BestInnerSeqDtls
			}
		}
		//RenewResult(&gState.BestInnerSeqs, &gState.BestInnerAsgmts, &gState.BestInnerUnasgTasks, &gState.InnerInfeaTasks, &gState, &gPara)
		//fmt.Printf("solver totalDistance:%f\n", GetDistance(gState.BestInnerSeqs, gState.BestInnerAsgmts, &gPara))
		//fmt.Printf("solver totalFitCost:%f\n", GetSeqsFitCost(gState.BestInnerSeqs, gState.BestInnerAsgmts, &gPara)/CostCoef)
		//fmt.Printf("solver totalMapCost:%f\n", GetSeqsMapCost(gState.BestInnerSeqs, gState.BestInnerAsgmts, &gPara)/CostCoef)
		//var bestDtlMapCost float64 = 0.0
		//for i := 0; i < len(gState.BestInnerSeqDtls); i++ {
		//	tmpCost := gState.BestInnerSeqDtls[i][5]
		//	fmt.Println(i, "线路Cost:", tmpCost)
		//	bestDtlMapCost += tmpCost
		//}
		//fmt.Printf("solver totalMapCost in best dtl:%f\n", bestDtlMapCost/CostCoef)
	}
	// add TransProcess
	if gPara.Obj != 0 {
		Method04ForBest(&gPara, &gState)
		gState.InnerAsgmts = CopySliceInt(gState.BestInnerAsgmts)
		//if ifTrans {
		//	fmt.Println("换车成功")
		//}
	}
	CheckDataPost(&gState, &gPara)
	step1EndTime := time.Now().Unix()
	//end2Time := time.Unix(step1EndTime, 0).Add(time.Duration(solverInput.CalDur2) * time.Second).Unix()
	end2Time := solverInput.SolverEndTime - 160
	step1Elapsed := float64(step1EndTime-solverStartTime) / 60
	if gState.Flag >= 0 && gPara.Obj == 2 {
		//开始第二阶段优化 10min
		step2StartTime := time.Now().Unix()
		tmpGState.InnerSeqs, tmpGState.InnerAsgmts, tmpGState.InnerUnasgTasks, tmpGState.InnerInfeaTasks = CopyBestResult(&gState)
		tmpGState.BestInnerSeqs, tmpGState.BestInnerAsgmts, tmpGState.BestInnerUnasgTasks, tmpGState.BestInnerInfeaTasks = CopyBestResult(&gState)
		tmpGState.BestInnerSeqDtls = GenerateSeqDtls(tmpGState.BestInnerSeqs, tmpGState.BestInnerAsgmts, &gPara)
		tmpGState.InnerSeqDtls = GenerateSeqDtls(tmpGState.InnerSeqs, tmpGState.InnerAsgmts, &gPara)
		gState = GState{}
		gState = tmpGState
		total_dis1 := 0.0
		total_cost1 := 0.0
		total_dur1 := 0.0
		for i := 0; i < len(gState.InnerSeqDtls); i++ {
			total_dis1 += gState.InnerSeqDtls[i][0]
			total_cost1 += gState.InnerSeqDtls[i][5]
			total_dur1 += gState.InnerSeqDtls[i][3]
		}
		//inner_dis1 := GetInnerDistance(gState.InnerSeqs, gState.InnerAsgmts, &gPara)
		//routes1 := len(gState.InnerSeqs)
		//avgDur1 := total_dur1 / (float64(routes1) * 3600)
		//gState.InnerSeqs, gState.InnerAsgmts, gState.InnerUnasgTasks, gState.InnerInfeaTasks, gState.InnerSeqDtls = CopyMatrixI(gState.BestInnerSeqs), CopySliceInt(gState.BestInnerAsgmts), CopySliceInt(gState.BestInnerUnasgTasks), CopySliceInt(gState.BestInnerInfeaTasks), CopyMatrixF(gState.InnerSeqDtls)
		//fmt.Printf("routes1:%v, avgDur1:%v, total_cost1:%v, total_dis1:%v, inner_dis1:%v\n", routes1, avgDur1, total_cost1/CostCoef, total_dis1, inner_dis1)
		gState.IsTimeEnough = true
		if solverInput.Obj != 0 {
			ShapeAdjustLocalSearch(end2Time, &gPara, &gState)
		}
		step2EndTime := time.Now().Unix()
		step2Elapsed := float64(step2EndTime-step2StartTime) / 60
		total_dis2 := 0.0
		total_cost2 := 0.0
		total_dur2 := 0.0
		for i := 0; i < len(gState.InnerSeqDtls); i++ {
			total_dis2 += gState.InnerSeqDtls[i][0]
			total_cost2 += gState.InnerSeqDtls[i][5]
			total_dur2 += gState.InnerSeqDtls[i][3]
		}
		fmt.Printf("time1:%v  time2:%v\n", step1Elapsed, step2Elapsed)
	} else {
		fmt.Printf("error: 1 step gState.flag<0")
	}
	RenewResultPShape(&gState.BestInnerSeqs, &gState.BestInnerAsgmts, &gState.BestInnerUnasgTasks, &gState.InnerInfeaTasks, &gState, &gPara, solverInput.SolverEndTime)
	gState1 := GState{}
	gState1.BestInnerSeqs, gState1.BestInnerAsgmts, gState1.BestInnerUnasgTasks, gState1.BestInnerInfeaTasks = CopyBestResult(&gState)
	gState1.InnerSeqs, gState1.InnerAsgmts, gState1.InnerUnasgTasks, gState1.InnerInfeaTasks = CopyBestResult(&gState)
	gState1.InnerSeqDtls = GenerateSeqDtls(gState1.InnerSeqs, gState1.InnerAsgmts, &gPara)
	gState1.BestInnerSeqDtls = GenerateSeqDtls(gState1.BestInnerSeqs, gState1.BestInnerAsgmts, &gPara)
	//start := time.Now()
	//aa := start.Unix()
	FineTuning(solverInput.SolverEndTime, &gState1, &gPara)
	//elapsed := time.Since(start)
	gState1.BestInnerSeqDtls = GenerateSeqDtls(gState1.BestInnerSeqs, gState1.BestInnerAsgmts, &gPara)
	gState.BestInnerAsgmts, gState.BestInnerSeqs, gState.BestInnerUnasgTasks, gState.BestInnerInfeaTasks = gState1.BestInnerAsgmts, gState1.BestInnerSeqs, gState1.BestInnerUnasgTasks, gState1.BestInnerInfeaTasks
	gState.BestInnerSeqDtls = gState1.BestInnerSeqDtls
	CheckDataPost(&gState, &gPara)
	PostProcessMethod(&gState, &gPara)

	//fmt.Println(InnerSeqDtls)
	//seqDtls := CheckSeqsDtl()
	//fmt.Println(seqDtls)

	//solverEndTime := time.Now().Unix()
	//solverCostTime := int(solverEndTime - solverStartTime)
	//fmt.Println("solver total time:", solverCostTime)
	//fmt.Println("total resources:", len(gState.InnerSeqs))
	//fmt.Println("total InfeaTasks:", len(gState.InnerInfeaTasks))
	//fmt.Println("total UnasgTasks:", len(gState.InnerUnasgTasks))
	solverOutput.Asgmts = gState.Asgmts
	solverOutput.Seqs = gState.Seqs
	solverOutput.InfeaTasks = gState.InfeaTasks
	solverOutput.UnasgTasks = gState.UnasgTasks
	solverOutput.Flag = 0
	if whetherOverTime {
		solverOutput.Flag = 2
	}
	if gState.Flag < 0 {
		solverOutput.Flag = gState.Flag
	}
	return solverOutput
}

//var initInfea []int

func AddInt(ch chan int, optTime int, solverOutput *SolverOutput, gState *GState, gPara *GPara, gFun *GFun) {
	defer func() {
		if e := recover(); e != nil {
			//fmt.Println("something unexpected:", err)
			solverOutput.Flag = -2
			solverOutput.Seqs = nil
			solverOutput.InfeaTasks = nil
			solverOutput.Asgmts = nil
			solverOutput.UnasgTasks = nil
			//err = errors.New("unexpected error!")
		}
	}()
	gState.InnerSeqs, gState.InnerAsgmts, gState.InnerUnasgTasks, gState.InnerInfeaTasks, gState.InnerSeqDtls = Optimize(optTime, gState, gPara, gFun)
	ch <- 1
}
func ControlSolverTime(optTime int, solverOutput *SolverOutput, gState *GState, gPara *GPara, gFun *GFun) bool {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Duration(optTime)*time.Second)
	defer cancel()
	ch := make(chan int, 1)
	go AddInt(ch, optTime, solverOutput, gState, gPara, gFun)
	select {
	case <-ch:
		return true
	case <-ctx.Done():
		return false
	}
}

func Optimize(optTime int, gState *GState, gPara *GPara, gFun *GFun) ([][]int, []int, []int, []int, [][]float64) {
	optStartTime := time.Now().Unix()
	optEndTime := time.Unix(optStartTime, 0).Add(time.Duration(optTime) * time.Second).Unix()
	gState.BestInnerSeqs, gState.BestInnerAsgmts, gState.BestInnerUnasgTasks, gState.BestInnerInfeaTasks = CopyResult(gState)
	gState.IsTimeEnough = true
	initInfea := CopySliceInt(gState.InnerInfeaTasks)
	MutationNum := 0
	rand.Seed(time.Now().Unix())
	ethnicity := []Ethnicity{}
	AddResultToEthnicity(&ethnicity, InitGaAddNum, &gState.BestInnerSeqs, &gState.BestInnerAsgmts, &gState.BestInnerUnasgTasks, &gState.BestInnerInfeaTasks, optTime, optEndTime, gState, gPara, &initInfea, gFun)
	if len(gState.InnerAsgmts) != 0 {
		InitTask2(gState, gFun, gPara, centDisWeight, graphDisWeight)
	}
	GaAddIndex := 0
	for time.Now().Unix() < optEndTime {
		nowTime := time.Now().Unix()
		if nowTime >= optEndTime {
			break
		}
		for j := 0; j < 10; j++ {
			tmp := [][]int{}
			for i := 0; i < len(gState.InnerSeqs); i++ {
				tmp = append(tmp, CopySliceInt(gState.InnerSeqs[i]))
			}
			ethnicity = append(ethnicity, Ethnicity{tmp, CopySliceInt(gState.InnerAsgmts)})
			preObj := gFun.GetObj(gState, gPara)
			preCost := GetFitCost(gState, gPara)
			preDist := GetDistObj(gState, gPara)

			//fmt.Println("进入LS")
			LocalSearch(optEndTime, gPara, gState)
			//fmt.Println("结束LS")
			//if postObj > preObj {
			//	fmt.Println("输出更差的解！！")
			//}
			//if !checkInfea(initInfea, *gState) {
			//	fmt.Printf("===========wrong  LocalSearch!==============\n\n\n")
			//}
			//ifOK := CheckDataOpt(gState, gPara)
			//if !ifOK {
			//	fmt.Println("infeasible!")
			//}
			//fmt.Println("进入M01")
			Method01(optEndTime, gPara, gState)
			//fmt.Println("结束M01")
			//if postObj > preObj {
			//	fmt.Println("输出更差的解！！")
			//}
			//if !checkInfea(initInfea, *gState) {
			//	fmt.Printf("===========wrong  Method01!==============\n\n\n")
			//}
			//ifOK = CheckDataOpt(gState, gPara)
			//if !ifOK {
			//	fmt.Println("infeasible!")
			//}
			//fmt.Println("进入M03")
			if gPara.Obj != 0 {
				Method04ForBest(gPara, gState)
				gState.InnerAsgmts = CopySliceInt(gState.BestInnerAsgmts)
				//if ifTrans {
				//	fmt.Println("换车成功")
				//}
			}
			Method03(gPara, gState, gFun)
			//fmt.Println("结束M03")
			//if !checkInfea(initInfea, *gState) {
			//	fmt.Printf("===========wrong  Method03!==============\n\n\n")
			//}
			//ifOK = CheckDataOpt(gState, gPara)
			//if !ifOK {
			//	fmt.Println("infeasible!")
			//}
			//postObj := gFun.GetObj(gState, gPara)
			//postCost := GetFitCost(gState, gPara)
			//postDist := GetDistObj(gState, gPara)

			//if postObj > preObj {
			//	fmt.Println("输出更差的解！！")
			//}
			RenewResult(&gState.BestInnerSeqs, &gState.BestInnerAsgmts, &gState.BestInnerUnasgTasks, &gState.InnerInfeaTasks, gState, gPara, optEndTime)
			if gPara.Obj != 0 {
				//fmt.Println("进入M05")
				Method05(gPara, gState)
				//fmt.Println("结束M05")
				//if !checkInfea(initInfea, *gState) {
				//	fmt.Printf("===========wrong  Method01!==============\n\n\n")
				//}
				//ifOK := CheckDataOpt(gState, gPara)
				//if !ifOK {
				//	fmt.Println("infeasible!")
				//}
			}

			postObj := gFun.GetObj(gState, gPara)
			postCost := GetFitCost(gState, gPara)
			postDist := GetDistObj(gState, gPara)
			//if postObj > preObj {
			//	fmt.Println("输出更差的解！！")
			//}
			improveObjPerc := (preObj - postObj) / preObj * 100
			improveCostPerc := (preCost - postCost) / preCost * 100
			improveDistPerc := (preDist - postDist) / preDist * 100
			var improvePerc float64 = 0.0
			if gPara.Obj == 0 {
				improvePerc = improveDistPerc
			} else {
				if improveObjPerc > 0.0001 {
					improvePerc = improveObjPerc
				} else {
					improvePerc = improveCostPerc
				}
			}
			//fmt.Println("improveObjPerc:", improveObjPerc, " improveFitCostPerc:", improveCostPerc, " improveDistPerc:", improveDistPerc)
			nowTime = time.Now().Unix()
			RenewResult(&gState.BestInnerSeqs, &gState.BestInnerAsgmts, &gState.BestInnerUnasgTasks, &gState.InnerInfeaTasks, gState, gPara, optEndTime)
			if (improvePerc >= 0 && improvePerc < 0.00001) || nowTime >= optEndTime {
				break
			}
		}
		nowTime = time.Now().Unix()
		if nowTime >= optEndTime || MutationNum > MaxMutationNum {
			break
		}
		//fmt.Println("进入MGA")
		res, newChanges := MethodGA(&ethnicity, optEndTime, gPara)
		//fmt.Println("结束MGA")
		//if !checkInfea(initInfea, *gState) {
		//	fmt.Printf("===========wrong  MethodGA!==============\n\n\n")
		//}
		if newChanges {
			MutationNum++
			gState.InnerSeqs = res.Seq
			gState.InnerFeats = GenerateFeats(gPara, res.Seq)
			gState.InnerSeqDtls = GenerateSeqDtls(res.Seq, res.Asg, gPara)
			gState.InnerAsgmts = CopySliceInt(res.Asg)
			RenewResult(&gState.BestInnerSeqs, &gState.BestInnerAsgmts, &gState.BestInnerUnasgTasks, &gState.InnerInfeaTasks, gState, gPara, optEndTime)
			//fmt.Printf("遗传变异成功: 第%d次\n\n", MutationNum)
		} else if GaAddNum > GaAddIndex {
			GaAddIndex++
			AddResultToEthnicity(&ethnicity, LoopGaAddNum, &gState.BestInnerSeqs, &gState.BestInnerAsgmts, &gState.BestInnerUnasgTasks, &gState.BestInnerInfeaTasks, optTime, optEndTime, gState, gPara, &initInfea, gFun)
		} else {
			break
		}
	}
	return gState.BestInnerSeqs, gState.BestInnerAsgmts, gState.BestInnerUnasgTasks, gState.BestInnerInfeaTasks, GenerateSeqDtls(gState.BestInnerSeqs, gState.BestInnerAsgmts, gPara)
}
func checkInfea(initInnfea []int, gState GState) bool {
	return len(initInnfea) == len(gState.InnerInfeaTasks)
}

//func OneRouteTidy(useTime int, gState *GState, gPara *GPara, routeIdx, asgmentIdx int) (deltaDistance float64, err error) {
//	oriSeq := CopySliceInt(gState.InnerSeqs[routeIdx])
//	nodeName := make([]string, len(oriSeq)+1)
//	G := make([][]float64, len(oriSeq)+1)
//	for i := 0; i < len(G); i++ {
//		G[i] = make([]float64, len(oriSeq)+1)
//		if i == 0 {
//			nodeName[i] = "0"
//		} else {
//			nodeName[i] = strconv.Itoa(oriSeq[i-1])
//		}
//	}
//	oriDis := GetSeqDistance(oriSeq, asgmentIdx, gState.InnerAsgmts, gPara)
//	for i := 1; i < len(G); i++ {
//		for j := 1; j < len(G[i]); j++ {
//			G[i][j] = float64(gPara.Cost[gState.InnerAsgmts[asgmentIdx]][0][gPara.TaskLoc[oriSeq[i-1]]][gPara.TaskLoc[oriSeq[j-1]]])
//		}
//	}
//	for i := 1; i < len(G); i++ {
//		G[0][i] = float64(gPara.Cost[gState.InnerAsgmts[asgmentIdx]][0][0][gPara.TaskLoc[oriSeq[i-1]]])
//		G[i][0] = float64(gPara.Cost[gState.InnerAsgmts[asgmentIdx]][0][gPara.TaskLoc[oriSeq[i-1]]][0])
//		G[i][i] = 0
//	}
//	seq, newDis, rc, err := lkh.SolveTspAsymmetrical(nodeName, G, nil, time.Now().Add(time.Duration(useTime)*time.Millisecond))
//	if rc != lkh.Success {
//		return 0, err
//	}
//	//newDis := GetSeqDistance(seq, asgmentIdx, gState.InnerAsgmts, gPara)
//	fmt.Printf("newDis:%v   oriDis:%v\n", newDis, oriDis)
//	if newDis < oriDis && err == nil {
//		startIdx := 0
//		for i := 0; i < len(seq); i++ {
//			if seq[i] == 0 {
//				startIdx = i
//			}
//		}
//		for i := 0; i < startIdx; i++ {
//			oriSeq[len(oriSeq)-startIdx+i] = gState.InnerSeqs[routeIdx][seq[i]-1]
//		}
//		for i := startIdx + 1; i < len(seq); i++ {
//			oriSeq[i-startIdx-1] = gState.InnerSeqs[routeIdx][seq[i]-1]
//		}
//		gState.InnerSeqs[routeIdx] = oriSeq
//		return newDis - oriDis, nil
//	}
//	return newDis - oriDis, err
//}

func AddResultToEthnicity(ethnicity *[]Ethnicity, addNum int, BestInnerSeqs *[][]int, BestInnerAsgmts, BestInnerUnasgTasks, BestInnerInfeaTasks *[]int, optTime int, optEndTime int64, gState *GState, gPara *GPara, initInfea *[]int, gFun *GFun) {
	for j := 0; j < addNum; j++ {
		tmp := [][]int{}
		for i := 0; i < len(gState.InnerSeqs); i++ {
			tmp = append(tmp, CopySliceInt(gState.InnerSeqs[i]))
		}
		*ethnicity = append(*ethnicity, Ethnicity{tmp, CopySliceInt(gState.InnerAsgmts)})
		if len(gState.InnerAsgmts) != 0 {
			if gState.ColdStart {
				InitTask2(gState, gFun, gPara, rand.Float64()*0.03+0.05+0.03*float64(j), rand.Float64()*0.2+0.6+0.05*float64(j))
			} else {
				gState.ColdStart = true
			}
		}
		var improvePerc float64 = 0.1
		for improvePerc > 0.00001 {
			nowTime := time.Now().Unix()
			if nowTime >= optEndTime {
				break
			}

			preCost := GetFitCost(gState, gPara)
			preObj := gFun.GetObj(gState, gPara)
			preDist := GetDistObj(gState, gPara)

			//fmt.Println("进入LS")
			LocalSearch(optEndTime, gPara, gState)
			//fmt.Println("结束LS")
			RenewResult(&gState.BestInnerSeqs, &gState.BestInnerAsgmts, &gState.BestInnerUnasgTasks, &gState.InnerInfeaTasks, gState, gPara, optEndTime)
			//postObj := gFun.GetObj(gState, gPara)
			//postCost := GetFitCost(gState, gPara)
			//postDist := GetDistObj(gState, gPara)
			//if postObj > preObj {
			//	fmt.Println("LS输出更差的解！！")
			//}
			//if !checkInfea(*initInfea, *gState) {
			//	fmt.Printf("===========wrong  LocalSearch!==============\n\n\n")
			//}
			//ifOK = CheckDataOpt(gState, gPara)
			//if !ifOK {
			//	fmt.Println("infeasible!")
			//}
			//fmt.Println("进入M01")
			Method01(optEndTime, gPara, gState)
			//fmt.Println("结束M01")
			//postObj = gFun.GetObj(gState, gPara)
			//postCost = GetFitCost(gState, gPara)
			//postDist = GetDistObj(gState, gPara)
			//if postObj > preObj {
			//	fmt.Println("输出更差的解！！")
			//}
			//if !checkInfea(*initInfea, *gState) {
			//	fmt.Printf("===========wrong  Method01!==============\n\n\n")
			//}
			//ifOK = CheckDataOpt(gState, gPara)
			//if !ifOK {
			//	fmt.Println("infeasible!")
			//}
			//fmt.Println("进入M03")
			if gPara.Obj != 0 {
				Method04ForBest(gPara, gState)
				gState.InnerAsgmts = CopySliceInt(gState.BestInnerAsgmts)
				//if ifTrans {
				//	fmt.Println("换车成功")
				//}
			}
			Method03(gPara, gState, gFun)
			//fmt.Println("结束M03")
			//if !checkInfea(*initInfea, *gState) {
			//	fmt.Printf("===========wrong  Method03!==============\n\n\n")
			//}
			//ifOK = CheckDataOpt(gState, gPara)
			//if !ifOK {
			//	fmt.Println("infeasible!")
			//}
			//postObj := gFun.GetObj(gState, gPara)
			//postCost := GetFitCost(gState, gPara)
			//postDist := GetDistObj(gState, gPara)
			//if postObj > preObj {
			//	fmt.Println("输出更差的解！！")
			//}
			RenewResult(&gState.BestInnerSeqs, &gState.BestInnerAsgmts, &gState.BestInnerUnasgTasks, &gState.InnerInfeaTasks, gState, gPara, optEndTime)
			if gPara.Obj != 0 {
				//fmt.Println("进入M05")
				Method05(gPara, gState)
				//fmt.Println("结束M05")
				//if !checkInfea(*initInfea, *gState) {
				//	fmt.Printf("===========wrong  Method01!==============\n\n\n")
				//}
				//ifOK := CheckDataOpt(gState, gPara)
				//if !ifOK {
				//	fmt.Println("infeasible!")
				//}
			}

			postObj := gFun.GetObj(gState, gPara)
			postCost := GetFitCost(gState, gPara)
			postDist := GetDistObj(gState, gPara)

			improveCostPerc := (preCost - postCost) / preCost * 100
			improveObjPerc := (preObj - postObj) / preObj * 100
			improveDistPerc := (preDist - postDist) / preDist * 100
			if gPara.Obj == 0 {
				improvePerc = improveDistPerc
			} else {
				if improveObjPerc > 0.0001 {
					improvePerc = improveObjPerc
				} else {
					improvePerc = improveCostPerc
				}
			}
			//fmt.Println("add improveObjPerc:", improveObjPerc, " add improveCostPerc:", improveCostPerc, " add improveDistPerc:", improveDistPerc)
			tmp = [][]int{}
			for i := 0; i < len(gState.InnerSeqs); i++ {
				tmp = append(tmp, CopySliceInt(gState.InnerSeqs[i]))
			}
			*ethnicity = append(*ethnicity, Ethnicity{tmp, CopySliceInt(gState.InnerAsgmts)})
			RenewResult(BestInnerSeqs, BestInnerAsgmts, BestInnerUnasgTasks, BestInnerInfeaTasks, gState, gPara, optEndTime)
		}
	}
}

func RenewResult(BestInnerSeqs *[][]int, BestInnerAsgmts, BestInnerUnasgTasks, InnerInfeaTasks *[]int, gState *GState, gPara *GPara, allowTime int64) {
	if len(gState.InnerUnasgTasks) > len(*BestInnerUnasgTasks) || !CheckDataOpt(gState, gPara) || time.Now().Unix() >= allowTime {
		return
	}
	var valueA, valueB, delta float64
	if gPara.Obj == 0 {
		valueA = GetDistance(gState.InnerSeqs, gState.InnerAsgmts, gPara)
		valueB = GetDistance(*BestInnerSeqs, *BestInnerAsgmts, gPara)
		delta = valueA - valueB
	} else {
		newMC := GetSeqsMapCost(gState.InnerSeqs, gState.InnerAsgmts, gPara)
		oldMC := GetSeqsMapCost(*BestInnerSeqs, *BestInnerAsgmts, gPara)

		deltaMC := newMC - oldMC
		if deltaMC < 0 {
			delta = deltaMC
		} else if deltaMC == 0 {
			newFC := GetSeqsFitCost(gState.InnerSeqs, gState.InnerAsgmts, gPara)
			oldFC := GetSeqsFitCost(*BestInnerSeqs, *BestInnerAsgmts, gPara)
			delta = newFC - oldFC
		} else {
			delta = 10000
		}
		if delta == 0 {
			valueA = GetDistance(gState.InnerSeqs, gState.InnerAsgmts, gPara)
			valueB = GetDistance(*BestInnerSeqs, *BestInnerAsgmts, gPara)
			delta = valueA - valueB
		}
	}
	if delta < 0 || len(gState.InnerUnasgTasks) < len(*BestInnerUnasgTasks) {
		*BestInnerSeqs, *BestInnerAsgmts, *BestInnerUnasgTasks, *InnerInfeaTasks = CopyResult(gState)
		gState.BestInnerSeqDtls = GenerateSeqDtls(*BestInnerSeqs, *BestInnerAsgmts, gPara)
	}
}

func RenewResultPShape(BestInnerSeqs *[][]int, BestInnerAsgmts, BestInnerUnasgTasks, InnerInfeaTasks *[]int, gState *GState, gPara *GPara, allowTime int64) {
	//if len(gState.InnerUnasgTasks) > len(*BestInnerUnasgTasks) || !CheckDataOpt(gState, gPara) || time.Now().Unix() >= allowTime {
	//	return
	//}
	if len(gState.InnerUnasgTasks) > len(*BestInnerUnasgTasks) || !CheckDataOpt(gState, gPara) {
		return
	}
	var valueA, valueB, delta float64
	if gPara.Obj == 0 {
		valueA = GetDistance(gState.InnerSeqs, gState.InnerAsgmts, gPara)
		valueB = GetDistance(*BestInnerSeqs, *BestInnerAsgmts, gPara)
		delta = valueA - valueB
	} else {
		newMC := GetSeqsMapCost(gState.InnerSeqs, gState.InnerAsgmts, gPara)
		oldMC := GetSeqsMapCost(*BestInnerSeqs, *BestInnerAsgmts, gPara)

		newInnerDist := GetInnerDistance(gState.InnerSeqs, gState.InnerAsgmts, gPara)
		oldInnerDist := GetInnerDistance(*BestInnerSeqs, *BestInnerAsgmts, gPara)
		deltaMC := newMC - oldMC
		deltaInnerDist := newInnerDist - oldInnerDist
		if deltaMC <= 0 && deltaInnerDist <= 0 {
			delta = deltaInnerDist
		} else {
			delta = 10000
		}
	}
	if delta < 0 || len(gState.InnerUnasgTasks) < len(*BestInnerUnasgTasks) {
		*BestInnerSeqs, *BestInnerAsgmts, *BestInnerUnasgTasks, *InnerInfeaTasks = CopyResult(gState)
		gState.BestInnerSeqDtls = GenerateSeqDtls(*BestInnerSeqs, *BestInnerAsgmts, gPara)
	}
}

func CopyResult(gState *GState) ([][]int, []int, []int, []int) {
	BestInnerSeqs := make([][]int, len(gState.InnerSeqs))
	for i := 0; i < len(gState.InnerSeqs); i++ {
		BestInnerSeqs[i] = CopySliceInt(gState.InnerSeqs[i])
	}
	return BestInnerSeqs, CopySliceInt(gState.InnerAsgmts), CopySliceInt(gState.InnerUnasgTasks), CopySliceInt(gState.InnerInfeaTasks)
}
func CopyBestResult(gState *GState) ([][]int, []int, []int, []int) {
	BestInnerSeqs := make([][]int, len(gState.BestInnerSeqs))
	for i := 0; i < len(gState.BestInnerSeqs); i++ {
		BestInnerSeqs[i] = CopySliceInt(gState.BestInnerSeqs[i])
	}
	return BestInnerSeqs, CopySliceInt(gState.BestInnerAsgmts), CopySliceInt(gState.BestInnerUnasgTasks), CopySliceInt(gState.BestInnerInfeaTasks)
}

type VisInput struct {
	InnerSeqs       [][]int
	InnerInfeaTasks []int
	InnerUnasgTasks []int
	Tasks           []*Task
	Nodes           []*Node
	Title           string
	PlotPath        string
}

func GetLatLngTasksWithSta(visInput VisInput) ([][][]float64, [][]float64, [][]float64) {
	var seqLatLngTasks = make([][][]float64, 0)
	for i := 0; i < len(visInput.InnerSeqs); i++ {
		var sLoc = make([][]float64, 0)
		//开头加station
		sLoc = append(sLoc, []float64{visInput.Nodes[0].Lat, visInput.Nodes[0].Lng})
		for j := 0; j < len(visInput.InnerSeqs[i]); j++ {
			nodeLoc := []float64{visInput.Nodes[visInput.Tasks[visInput.InnerSeqs[i][j]].NodeIdx].Lat,
				visInput.Nodes[visInput.Tasks[visInput.InnerSeqs[i][j]].NodeIdx].Lng}
			sLoc = append(sLoc, nodeLoc)
		}
		//结束回station
		sLoc = append(sLoc, []float64{visInput.Nodes[0].Lat, visInput.Nodes[0].Lng})
		seqLatLngTasks = append(seqLatLngTasks, sLoc)
	}

	var unasgLatLngTasks = make([][]float64, 0)
	for i := 0; i < len(visInput.InnerUnasgTasks); i++ {
		nodeLoc := []float64{visInput.Nodes[visInput.Tasks[visInput.InnerUnasgTasks[i]].NodeIdx].Lat,
			visInput.Nodes[visInput.Tasks[visInput.InnerUnasgTasks[i]].NodeIdx].Lng}
		unasgLatLngTasks = append(unasgLatLngTasks, nodeLoc)
	}

	var infeaLatLngTasks = make([][]float64, 0)
	for i := 0; i < len(visInput.InnerInfeaTasks); i++ {
		nodeLoc := []float64{visInput.Nodes[visInput.Tasks[visInput.InnerInfeaTasks[i]].NodeIdx].Lat,
			visInput.Nodes[visInput.Tasks[visInput.InnerInfeaTasks[i]].NodeIdx].Lng}
		infeaLatLngTasks = append(infeaLatLngTasks, nodeLoc)
	}

	return seqLatLngTasks, unasgLatLngTasks, infeaLatLngTasks
}
