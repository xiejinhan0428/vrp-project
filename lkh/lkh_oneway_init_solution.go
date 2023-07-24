package lkh

import "sort"

func SolveTspOnewayInit(input TSPSolverInput) (output TSPSolverOutput) {
	//兜底解不支持fixedEdge，因为有可能是fixedEdge导致的错误
	distanceMatrix, startIdx, endIdx := input.DistanceMatrix, input.StartIdx, input.EndIdx
	sequence, distance := output.Sequence, output.Distance
	rawLen := len(distanceMatrix)
	sequence = make([]int, rawLen)
	sequence[0] = startIdx
	tmpDisList := make([]MySort, rawLen)
	usedList := make([]bool, rawLen)
	usedList[startIdx] = true
	free := 0
	if endIdx >= 0 && endIdx != startIdx {
		usedList[endIdx] = true
		sequence[rawLen-1] = endIdx
		free = 1
	}
	for i := 1; i < rawLen-free; i++ {
		for j := 0; j < rawLen; j++ {
			tmpDisList[j].Distance = distanceMatrix[startIdx][j]
			tmpDisList[j].Index = j
		}
		sort.SliceStable(tmpDisList, func(k, l int) bool {
			return tmpDisList[k].Distance < tmpDisList[l].Distance
		})
		for k := 0; k < rawLen; k++ {
			idx := tmpDisList[k].Index
			if !usedList[idx] {
				sequence[i] = idx
				usedList[idx] = true
				startIdx = idx
				break
			}
		}
	}
	distance = 0
	for i := 0; i < rawLen-1; i++ {
		distance += distanceMatrix[sequence[i]][sequence[i+1]]
	}
	if startIdx == endIdx {
		distance += distanceMatrix[sequence[rawLen-1]][sequence[0]]
	}
	return TSPSolverOutput{sequence, distance, Success, nil}
}

type MySort struct {
	Index    int
	Distance float64
}
