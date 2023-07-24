package solver

import (
	"errors"
	"strconv"
	"time"
)

// Solve finds the shortest tour of a graph.
//  names: the names of nodes.
//  distanceMatrix: the matrix of distances between nodes. It must be square and symmetric.
//  fixedEdges: elements of fixedEdges denote the fixed edges in the final tour.
//  sequence: sequence[i] == k means that the k-th node is the i-th visited node in the final tour.
//  retcode: return code, see lkh/config.go for details.
//  err: error that occurs in finding the shortest tour.

func SolveTspOneway(input TSPSolverInput) (output TSPSolverOutput) {
	names, distanceMatrix, fixedEdges, startIdx, endIdx, leftTime, iSSymmetrical := input.Names, input.DistanceMatrix, input.FixedEdges, input.StartIdx, input.EndIdx, input.LeftTime, input.ISSymmetrical
	sequence, distance, retcode, err := output.Sequence, output.Distance, output.Retcode, output.Err
	rawLen := len(distanceMatrix)
	if rawLen < 12 {
		outputExhaustive := SolveTspOnewayExhaustive(input)
		return outputExhaustive
	}
	endTime := time.Now().Add(time.Duration(leftTime) * time.Millisecond)
	if fixedEdges != nil && !iSSymmetrical {
		lenthFixed := len(fixedEdges)
		for i := 0; i < lenthFixed; i++ {
			fixedEdges = append(fixedEdges, [2]int{fixedEdges[i][1], fixedEdges[i][1] + rawLen})
			fixedEdges[i][1] += rawLen
		}
	}
	//roundtrip模式
	if endIdx == startIdx {
		if iSSymmetrical {
			sequence, distance, retcode, err = Solve(names, distanceMatrix, fixedEdges, endTime)
		} else {
			sequence, distance, retcode, err = SolveTspAsymmetrical(names, distanceMatrix, fixedEdges, endTime)
		}
		newSeq := []int{}
		for i := 0; i < len(sequence); i++ {
			if sequence[i] == startIdx {
				newSeq = append(newSeq, sequence[i:]...)
				newSeq = append(newSeq, sequence[:i]...)
				break
			}
		}
		return TSPSolverOutput{newSeq, distance, retcode, err}
	}
	//oneway 模式增加虚拟点
	names = append(names, "zeroPoint")
	for i := 0; i < rawLen; i++ {
		distanceMatrix[i] = append(distanceMatrix[i], 0)
	}
	distanceMatrix = append(distanceMatrix, make([]float64, rawLen+1))
	if fixedEdges != nil {
		if iSSymmetrical {
			fixedEdges = append(fixedEdges, [2]int{startIdx, rawLen})
		} else {
			fixedEdges = append(fixedEdges, [2]int{startIdx, rawLen + rawLen})
			fixedEdges = append(fixedEdges, [2]int{rawLen, rawLen + rawLen})
		}
	} else {
		if iSSymmetrical {
			fixedEdges = [][2]int{{startIdx, rawLen}}
		} else {
			fixedEdges = [][2]int{{startIdx, rawLen + len(distanceMatrix)}, {rawLen, rawLen + len(distanceMatrix)}}
		}
	}
	//oneway_free模式
	if endIdx >= 0 {
		if !iSSymmetrical {
			for i := 0; i < rawLen+1; i++ {
				distanceMatrix[rawLen][i] = MaxFl / 5
				distanceMatrix[i][rawLen] = MaxFl / 5
			}
			distanceMatrix[rawLen][startIdx] = 0
			distanceMatrix[endIdx][rawLen] = 0
			distanceMatrix[rawLen][rawLen] = 0
			fixedEdges = append(fixedEdges, [2]int{rawLen, endIdx + len(distanceMatrix)})
			fixedEdges = append(fixedEdges, [2]int{endIdx, endIdx + len(distanceMatrix)})
		} else {
			fixedEdges = append(fixedEdges, [2]int{rawLen, endIdx})
		}
	}
	if iSSymmetrical {
		sequence, distance, retcode, err = Solve(names, distanceMatrix, fixedEdges, endTime)
	} else {
		sequence, distance, retcode, err = SolveTspAsymmetrical(names, distanceMatrix, fixedEdges, endTime)
	}
	if retcode != Success || err != nil {
		return TSPSolverOutput{sequence, distance, retcode, err}
	}
	tmpSeq := []int{}
	newSeq := make([]int, len(sequence)-1)
	tmpSeq = append(tmpSeq, sequence...)
	tmpSeq = append(tmpSeq, sequence...)
	itorStart := 0
	itorVirtua := 0
	for i := 0; i < len(sequence); i++ {
		if sequence[i] == startIdx {
			itorStart = i
		}
		if sequence[i] == rawLen {
			itorVirtua = i
		}
	}
	check := make([]bool, rawLen)
	//if endIdx < 0 {
	//	j := itorStart
	//	for i := 0; i < len(sequence)-1; i++ {
	//		if j == itorVirtua {
	//			j++
	//		}
	//		newSeq[i] = tmpSeq[j]
	//		check[newSeq[i]] = true
	//		j++
	//	}
	//	if (itorStart+1)%len(sequence) == itorVirtua%len(sequence) {
	//		fmt.Printf("%v\n", sequence)
	//	}
	//} else
	if (itorStart+1)%len(sequence) == itorVirtua%len(sequence) {
		j := itorStart + len(sequence)
		for i := 0; i < len(sequence)-1; i++ {
			newSeq[i] = tmpSeq[j]
			check[newSeq[i]] = true
			j--
		}
	} else if (itorVirtua+1)%len(sequence) == itorStart%len(sequence) {
		for i := 0; i < len(sequence)-1; i++ {
			newSeq[i] = tmpSeq[itorStart+i]
			check[newSeq[i]] = true
		}
	} else {
		err = errors.New("Oneway: The virtual point is not adjacent to the start point")
		retcode = InvalidTour
	}
	for i := 0; i < len(check); i++ {
		if !check[i] {
			err = errors.New("Oneway: loss point :" + strconv.Itoa(i))
			retcode = LossPoint
			break
		}
	}
	lenth := len(newSeq)
	distance = 0
	for i := 0; i < lenth-1; i++ {
		distance += distanceMatrix[newSeq[i]][newSeq[i+1]]
	}
	return TSPSolverOutput{newSeq, distance, retcode, err}
}
