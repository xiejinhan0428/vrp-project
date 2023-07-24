package solver

import (
	"errors"
	"fmt"
	"math"
	"sort"
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

func SolveTspAsymmetrical(names []string, distanceMatrix [][]float64, fixedEdges [][2]int, endTime time.Time) (sequence []int, distance float64, retcode RetCode, err error) {
	G := make([][]float64, len(distanceMatrix)*2)
	minFl := -1.0
	doubleNames := make([]string, len(distanceMatrix)*2)
	for i := 0; i < len(G); i++ {
		G[i] = make([]float64, len(distanceMatrix)*2)
	}
	lenth := len(distanceMatrix)
	for i := 0; i < lenth; i++ {
		for j := 0; j < lenth; j++ {
			G[i][j] = MaxFl
			G[i+lenth][j+lenth] = MaxFl
			G[i+lenth][j] = distanceMatrix[i][j]
			G[j][i+lenth] = distanceMatrix[i][j]
			minFl = math.Max(distanceMatrix[i][j], minFl)
		}
	}
	minFl = -minFl * 50
	for i := 0; i < lenth; i++ {
		G[i][i+lenth] = minFl
		G[i+lenth][i] = minFl
		doubleNames[i] = names[i]
		doubleNames[i+lenth] = names[i] + "_2"
	}
	for i := 0; i < lenth; i++ {
		fixedEdges = append(fixedEdges, [2]int{i, i + lenth})
	}
	seq, dis, rc, err := Solve(doubleNames, G, fixedEdges, endTime)
	check := make([]bool, lenth)
	newSeq := make([]int, len(distanceMatrix))
	for i := 0; i < len(seq); i += 2 {
		newSeq[i/2] = seq[i]
		if seq[i] >= lenth {
			err = errors.New("TspAsymmetrical: error newSeq :" + strconv.Itoa(seq[i]))
			rc = InvalidTour
			return newSeq, dis, rc, err
		}
		check[seq[i]] = true
	}
	//dis1 := dis - float64(lenth)*minFl
	//fmt.Printf("dis: %f   dis1: %f\n", dis, dis1)
	for i := 0; i < len(check); i++ {
		if !check[i] {
			err = errors.New("TspAsymmetrical: loss point :" + strconv.Itoa(i))
			rc = LossPoint
			break
		}
	}
	disa, disb := CalculateDis(distanceMatrix, newSeq)
	if disb < disa {
		sort.SliceStable(newSeq, func(i, j int) bool { return true })
		dis = disb
	} else {
		dis = disa
	}
	fmt.Printf("disa:%f     disb:%f\nseq:%v\n", disa, disb, newSeq)
	return newSeq, dis, rc, err
}
func CalculateDis(distanceMatrix [][]float64, seq []int) (float64, float64) {
	lenth := len(seq)
	dis1 := distanceMatrix[seq[lenth-1]][seq[0]]
	dis2 := distanceMatrix[seq[0]][seq[lenth-1]]
	for i := 0; i < lenth-1; i++ {
		dis1 += distanceMatrix[seq[i]][seq[i+1]]
		dis2 += distanceMatrix[seq[i+1]][seq[i]]
	}
	return dis1, dis2
}
