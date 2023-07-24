package solver

import (
	"fmt"
	"math"
	"math/rand"
	"time"
)

func genInitTour(nodes []*TSPNode, G [][]int) (initTour []*TSPNode, err error) {
	N := len(nodes)
	isInTour := make([]bool, N)
	initTour = make([]*TSPNode, N)
AddNode:
	for i := 0; i < N; i++ {
		if i == 0 {
			initTour[0] = nodes[0]
			isInTour[nodes[0].idx] = true
			continue
		}
		// add the best candidate of prevNode to the tour
		prevNode := initTour[i-1]
		if prevNode.fixNode1 != nil && !isInTour[prevNode.fixNode1.idx] {
			initTour[i] = prevNode.fixNode1
			isInTour[prevNode.fixNode1.idx] = true
			continue
		} else if prevNode.fixNode2 != nil && !isInTour[prevNode.fixNode2.idx] {
			initTour[i] = prevNode.fixNode2
			isInTour[prevNode.fixNode2.idx] = true
			continue
		}
		for _, c := range prevNode.candidates {
			if !isInTour[c.idx] {
				initTour[i] = c
				isInTour[c.idx] = true
				continue AddNode
			}
		}
		// if all of prevNode's candidates have been added to the tour, pick a node not chosen
		minDis := math.MaxInt64
		t := -1
		for idx, node := range nodes {
			if !isInTour[node.idx] {
				if calDis(node, prevNode, G) < minDis {
					minDis = calDis(node, prevNode, G)
					t = idx
				}
			}
		}
		if t < 0 {
			return nil, fmt.Errorf("genInitTour: cannot find a node not in the initial tour")
		}
		initTour[i] = nodes[t]
		isInTour[nodes[t].idx] = true
	}

	return initTour, nil
}

func randomInitTour(nodes []*TSPNode) (initTour []*TSPNode) {
	N := len(nodes)
	initTour = make([]*TSPNode, N)
	for i := 0; i < N; i++ {
		initTour[i] = nodes[i]
	}
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(initTour), func(i, j int) {
		initTour[i], initTour[j] = initTour[j], initTour[i]
	})
	return initTour
}
func CheckSymmetrical(G [][]int) bool {
	lenth := len(G)
	for i := 0; i < lenth/2; i++ {
		for j := 0; j < lenth/2; j++ {
			if G[i][j] != MaxFl*PRECISION {
				return true
			}
		}
	}
	return false
}
