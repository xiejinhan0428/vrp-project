package lkh

import (
	"fmt"
	"time"
)

// Solve finds the shortest tour of a graph.
//  names: the names of nodes.
//  distanceMatrix: the matrix of distances between nodes. It must be square and symmetric.
//  fixedEdges: elements of fixedEdges denote the fixed edges in the final tour.
//  sequence: sequence[i] == k means that the k-th node is the i-th visited node in the final tour.
//  retcode: return code, see lkh/config.go for details.
//  err: error that occurs in finding the shortest tour.
func Solve(names []string, distanceMatrix [][]float64, fixedEdges [][2]int, endTime time.Time) (sequence []int, distance float64, retcode RetCode, err error) {
	var start time.Time
	var elapsed time.Duration
	if len(names) != 0 && len(names) != len(distanceMatrix) {
		return nil, 0, InvalidDistanceMatrix, fmt.Errorf("the length of names is %d, the length of distance matrix is %d, they are not equal", len(names), len(distanceMatrix))
	}
	start = time.Now()
	start2 := time.Now()
	G, err := checkDistMat(distanceMatrix)
	if err != nil {
		return nil, 0, InvalidDistanceMatrix, err
	}
	nodes, err := genNodes(names, G, fixedEdges)
	if err != nil {
		return nil, 0, InvalidFixedEdges, err
	}
	lowerBound, err := ascent(nodes, G)
	//fmt.Printf("lower bound is %v\n", lowerBound/PRECISION)
	if err != nil {
		return nil, 0, NoLowerBound, err
	}
	elapsed = time.Since(start)
	fmt.Printf("ascent: %v\n", elapsed)
	maxNearness := float64(lowerBound) / float64(len(nodes))
	err = genCandidates(nodes, G, getDefaultMaxCandidates(len(distanceMatrix)), maxNearness) //maxNearness和lowerBound含义
	if err != nil {
		return nil, 0, NoCandidates, err
	}
	for _, n := range nodes {
		n.pi = 0
	}
	//根据Candidates产生初始解
	//initTour := randomInitTour(nodes)
	initTour, err := genInitTour(nodes, G)
	if err != nil {
		return nil, 0, InvalidTour, err
	}
	initTourLen := 0
	for i := 0; i < len(initTour); i++ {
		cur := initTour[i]
		j := i + 1
		if j >= len(initTour) {
			j = 0
		}
		next := initTour[j]
		initTourLen += calDis(cur, next, G)
	}

	//fmt.Printf("Length of initial tour is %.2f\n", float64(initTourLen)/float64(PRECISION))

	tour := &Tour{
		path:        initTour,
		size:        len(nodes),
		graph:       G,
		edges:       make(map[Edge]bool),
		dis:         float64(initTourLen) / float64(PRECISION),
		symmetrical: CheckSymmetrical(G),
	}
	better := true
	//fmt.Println(SkipKopt)
	start = time.Now()
	for better && !SkipKopt {
		better = optimize(tour, tour.graph, endTime)
	}
	elapsed = time.Since(start)
	elapsed2 := time.Since(start2)
	fmt.Printf("optimize: %v\n", elapsed)
	fmt.Printf("total time: %v\n", elapsed2)
	fmt.Printf("config: %v  %v   %v\n", getK(tour.symmetrical), getKT(len(distanceMatrix)), getDefaultMaxCandidates(len(distanceMatrix)))
	sequence = make([]int, 0)
	for _, n := range tour.path {
		sequence = append(sequence, n.idx)
	}
	//disa, disb := CalculateDis(distanceMatrix, sequence)
	//fmt.Printf("disa:%f     disb:%f\nseq:%v\n", disa, disb, sequence)
	//if distanceMatrix[sequence[len(sequence)-1]][sequence[0]] > MaxFl/100 {
	//	fmt.Printf("more than 1000000:%v,%v  %v\n", len(sequence)-1, sequence[0], distanceMatrix[sequence[len(sequence)-1]][sequence[0]])
	//}
	//for i := 0; i < len(sequence)-1; i++ {
	//	if distanceMatrix[sequence[i]][sequence[i+1]] > MaxFl/100 {
	//		fmt.Printf("more than 1000000:%v,%v  %v\n", sequence[i], sequence[i+1], distanceMatrix[sequence[i]][sequence[i+1]])
	//	}
	//}
	return sequence, tour.dis, Success, nil
}
