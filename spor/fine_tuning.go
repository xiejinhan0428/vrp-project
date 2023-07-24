package solver

import (
	"fmt"
	"sort"
	"time"
)

type SCL struct {
	taskIdx    int
	routeIdx   int
	SCDistance []float64
}

type SortStruct struct {
	Index int
	Val   float64
}

func FineTuning(tEndUnix int64, gState *GState, gPara *GPara) {
	pointLen := 0
	for i := 0; i < len(gState.BestInnerSeqs); i++ {
		pointLen += len(gState.BestInnerSeqs[i])
	}
	silhouetteDisList := GenerSCL(gState, gPara)
	//silhouetteCoefficientList := make([]float64,pointLen)
	flag := true
	cnt := 0
	var loopTime float64 = 1
	for flag && cnt < len(silhouetteDisList) && time.Now().Unix() < tEndUnix-int64(loopTime*1.5) {
		flag = false
		cnt += 1
		start := time.Now().Unix()
		for i := 0; i < len(silhouetteDisList); i++ {
			sortList := make([]SortStruct, len(silhouetteDisList[i].SCDistance))
			for j := 0; j < len(silhouetteDisList[i].SCDistance); j++ {
				sortList[j].Index = j
				sortList[j].Val = silhouetteDisList[i].SCDistance[j]
			}
			sort.SliceStable(sortList, func(i, j int) bool {
				return sortList[i].Val < sortList[j].Val
			})
			for j := 0; j < len(sortList); j++ {
				if sortList[j].Index == silhouetteDisList[i].routeIdx {
					break
				}
				if Tuning(silhouetteDisList[i].taskIdx, silhouetteDisList[i].routeIdx, sortList[j].Index, silhouetteDisList, gState, gPara) {
					flag = true
					silhouetteDisList = GenerSCL(gState, gPara)
					break
				}
			}
		}
		loopTime = float64(time.Now().Unix() - start)
		fmt.Printf("t-loopTime:  %v\n", loopTime)
	}
	flag = false
}

func Tuning(taskID int, oriRouteIdx int, destRouteIdx int, silhouetteDisList []SCL, gState *GState, gPara *GPara) bool {
	index := FindTaskId(taskID, gState.BestInnerSeqs[oriRouteIdx])
	if index == -1 {
		return false
	}
	oriTmp := CopySliceInt(gState.BestInnerSeqs[oriRouteIdx])
	destTmp := CopySliceInt(gState.BestInnerSeqs[destRouteIdx])
	totalDistance1 := GetSeqDistance(gState.BestInnerSeqs[oriRouteIdx], oriRouteIdx, gState.BestInnerAsgmts, gPara)
	totalDistance2 := GetSeqDistance(gState.BestInnerSeqs[destRouteIdx], destRouteIdx, gState.BestInnerAsgmts, gPara)
	cost1 := GetSeqMapCost(gPara, gState.BestInnerAsgmts[oriRouteIdx], totalDistance1)
	cost2 := GetSeqMapCost(gPara, gState.BestInnerAsgmts[destRouteIdx], totalDistance2)

	gState.BestInnerSeqs[oriRouteIdx] = append(gState.BestInnerSeqs[oriRouteIdx][:index], gState.BestInnerSeqs[oriRouteIdx][index+1:]...)
	insertIdx := FindInsertIdx(gState.BestInnerSeqs[destRouteIdx], gPara, taskID, gState.BestInnerAsgmts[destRouteIdx])
	gState.BestInnerSeqs[destRouteIdx] = append(gState.BestInnerSeqs[destRouteIdx][:insertIdx], append([]int{taskID}, gState.BestInnerSeqs[destRouteIdx][insertIdx:]...)...)

	totalDistance3 := GetSeqDistance(gState.BestInnerSeqs[oriRouteIdx], oriRouteIdx, gState.BestInnerAsgmts, gPara)
	totalDistance4 := GetSeqDistance(gState.BestInnerSeqs[destRouteIdx], destRouteIdx, gState.BestInnerAsgmts, gPara)
	cost3 := GetSeqMapCost(gPara, gState.BestInnerAsgmts[oriRouteIdx], totalDistance3)
	cost4 := GetSeqMapCost(gPara, gState.BestInnerAsgmts[destRouteIdx], totalDistance4)

	if CheckSeqCont2(gPara, gState) && cost3+cost4 <= cost1+cost2 {
		return true
	}
	gState.BestInnerSeqs[oriRouteIdx] = CopySliceInt(oriTmp)
	gState.BestInnerSeqs[destRouteIdx] = CopySliceInt(destTmp)
	return false
}
func FindTaskId(taskID int, slist []int) int {
	for i := 0; i < len(slist); i++ {
		if taskID == slist[i] {
			return i
		}
	}
	return -1
}

func FindInsertIdx(slist []int, gPara *GPara, taskId int, carIdx int) (insertIdx int) {
	delta := 99999999.0
	for i := 1; i < len(slist); i++ {
		d := gPara.Cost[carIdx][0][gPara.TaskLoc[slist[i-1]]][gPara.TaskLoc[taskId]] + gPara.Cost[carIdx][0][gPara.TaskLoc[taskId]][gPara.TaskLoc[slist[i]]] - gPara.Cost[carIdx][0][gPara.TaskLoc[slist[i-1]]][gPara.TaskLoc[slist[i]]]
		if delta > d {
			delta = d
			insertIdx = i
		}
	}
	d := gPara.Cost[carIdx][0][0][gPara.TaskLoc[taskId]] + gPara.Cost[carIdx][0][gPara.TaskLoc[taskId]][gPara.TaskLoc[slist[0]]] - gPara.Cost[carIdx][0][0][gPara.TaskLoc[slist[0]]]
	if delta > d {
		delta = d
		insertIdx = 0
	}
	d = gPara.Cost[carIdx][0][gPara.TaskLoc[taskId]][0] + gPara.Cost[carIdx][0][gPara.TaskLoc[len(slist)-1]][taskId] - gPara.Cost[carIdx][0][gPara.TaskLoc[len(slist)-1]][0]
	if delta > d {
		delta = d
		insertIdx = len(slist)
	}
	return insertIdx
}

func GenerSCL(gState *GState, gPara *GPara) []SCL {
	routeLen := len(gState.BestInnerAsgmts)
	pointLen := 0
	for i := 0; i < len(gState.BestInnerSeqs); i++ {
		pointLen += len(gState.BestInnerSeqs[i])
	}
	silhouetteCoefficientList := make([]SCL, pointLen)
	for i := 0; i < len(silhouetteCoefficientList); i++ {
		silhouetteCoefficientList[i].taskIdx = i
		silhouetteCoefficientList[i].SCDistance = make([]float64, routeLen)
		for j := 0; j < routeLen; j++ {
			dis := 0.0
			resIdx := gState.BestInnerAsgmts[j]
			oriRoute := false
			for k := 0; k < len(gState.BestInnerSeqs[j]); k++ {
				dis += gPara.Cost[resIdx][0][gPara.TaskLoc[i]][gPara.TaskLoc[gState.BestInnerSeqs[j][k]]]
				if gState.BestInnerSeqs[j][k] == i {
					silhouetteCoefficientList[i].routeIdx = j
					oriRoute = true
				}
			}
			if oriRoute && len(gState.BestInnerSeqs[j]) == 1 {
				dis = 0
			} else if oriRoute {
				dis /= float64(len(gState.BestInnerSeqs[j]) - 1)
			} else {
				dis /= float64(len(gState.BestInnerSeqs[j]))
			}
			silhouetteCoefficientList[i].SCDistance[j] = dis
		}
	}
	return silhouetteCoefficientList
}
