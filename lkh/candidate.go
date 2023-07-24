package lkh

import (
	"fmt"
	"math"
	"sort"
)

// checkDistMat checks the regularity of the provided distance matrix, which should be square, symmetric, and has zero
// diagonal elements. If the input distance matrix is valid, it will be converted to an integer matrix G according to
// a default PRECISION.
func checkDistMat(distMat [][]float64) (G [][]int, err error) {
	N := len(distMat)
	if N <= 0 {
		return nil, fmt.Errorf("checkDistMat: empty distance matrix")
	}
	if N == 1 {
		return nil, fmt.Errorf("checkDistMat: no need to find the shortest tour of ONE node")
	}

	for i, row := range distMat {
		if len(row) != N {
			return nil, fmt.Errorf("checkDistMat: invalid distance matrix: row %d has %d elements, but %d is required", i, len(row), N)
		}
	}

	G = make([][]int, N)
	for i := range G {
		G[i] = make([]int, N)
	}

	for i := 0; i < N; i++ {
		for j := i; j < N; j++ {
			if i == j {
				//if distMat[i][j] != 0 {
				//	return nil, fmt.Errorf("checkDistMat: invalid matrix: the distance from %d to %d should be 0, but %.2f provided", i, j, distMat[i][j])
				//}
				//G[i][j] = 0
				G[i][j] = int(distMat[i][j] * PRECISION)
			} else {
				if distMat[i][j] != distMat[j][i] {
					return nil, fmt.Errorf("checkDistMat: invalid matrix: asymmetric distance matrix is no supported now")
				}
				G[i][j] = int(distMat[i][j] * PRECISION)
				G[j][i] = int(distMat[i][j] * PRECISION)
			}
		}
	}

	return G, nil
}

// genNodes generates the nodes according to the distance matrix and associates the nodes in the fixed edges.
func genNodes(names []string, G [][]int, fixedEdges [][2]int) (nodes []*Node, err error) {
	N := len(G)
	nodes = make([]*Node, N)
	for i := 0; i < N; i++ {
		var name string
		if len(names) > 0 {
			name = names[i]
		}
		nodes[i] = &Node{
			idx:         i,
			name:        name,
			mstDad:      nil,
			mstRank:     0,
			degree:      0,
			candidates:  make([]*Node, 0),
			oneTreeSucc: nil,
			pi:          0,
			fixNode1:    nil,
			fixNode2:    nil,
		}
	}

	for _, edge := range fixedEdges {
		a := edge[0]
		b := edge[1]
		err = fixEdge(nodes[a], nodes[b])
		if err != nil {
			return nil, err
		}
	}

	return nodes, nil
}

// prim generates the minimum spanning tree (MST) of the graph
func prim(nodes []*Node, G [][]int) (err error) {
	N := len(nodes)
	isInMST := make([]bool, N) // the array to record whether a node has been added to the MST
	dist := make([]int, N)     // the array to record the distance from a node to the MST
	// put the first node in the MST
	dist[0] = 0
	isInMST[0] = true
	nodes[0].mstRank = 0
	nodes[0].degree = 0
	// update the distances between nodes[1:N-1] and nodes[0]
	for i := 1; i < N; i++ {
		dist[i] = calDis(nodes[0], nodes[i], G)
		nodes[i].mstDad = nodes[0]
		nodes[i].degree = 0
	}
	for i := 1; i < N; i++ {
		minDist := math.MaxInt64
		t := -1 // the node added to the MST in the current iteration
		for j := 1; j < N; j++ {
			if !isInMST[j] {
				//if nodes[j].fixNode1 != nil && isInMST[nodes[j].fixNode1.idx] {
				//	t = nodes[j].idx
				//	nodes[t].mstDad = nodes[j].fixNode1
				//	break
				//} else if nodes[j].fixNode2 != nil && isInMST[nodes[j].fixNode2.idx] {
				//	t = nodes[j].idx
				//	nodes[t].mstDad = nodes[j].fixNode2
				//	break
				//} else

				if dist[j] < minDist {
					minDist = dist[j]
					t = j
				}
			}
		}
		if t < 0 {
			return fmt.Errorf("prim: faild to generate the MST, maybe there is a unconnected node")
		}
		isInMST[t] = true
		nodes[t].mstRank = nodes[t].mstDad.mstRank + 1
		nodes[t].degree = 1
		nodes[t].mstDad.degree += 1
		for k := 1; k < N; k++ {
			if !isInMST[k] && dist[k] > calDis(nodes[t], nodes[k], G) {
				dist[k] = calDis(nodes[t], nodes[k], G) // update the distance from nodes[k] to the MST
				nodes[k].mstDad = nodes[t]              // update the temp dad of nodes[k]
			}
		}
	}

	return nil
}

// min1Tree finds the minimum 1-tree of the graph
func min1Tree(nodes []*Node, G [][]int) (cost int, err error) {
	if err = prim(nodes, G); err != nil {
		return 0, fmt.Errorf("min1Tree: failed to generate the MST: %w", err)
	}
	mstCost := 0

	var chosenLeafNode *Node
	var secondNearestNeighbor *Node
	maxSecondMinWeight := -math.MaxInt64
	sumPi := 0
	for _, n := range nodes {
		sumPi += n.pi
		if n.mstDad == nil {
			continue
		}
		mstCost += calDis(n, n.mstDad, G)
		if n.degree == 1 {
			secondMinWeightNow := math.MaxInt64
			var secondNearestNeighborNow *Node
			for _, nn := range nodes {
				//if nn == n || nn == n.mstDad || n == nn.mstDad {
				if nn.degree != 1 {
					continue
				}
				//对每个叶子结点(度为1)，找其他（叶子）点相连，生成树形成环变成1-tree，寻找目标为成环最短，为目标1-边
				if calDis(nn, n, G) < secondMinWeightNow {
					secondMinWeightNow = calDis(nn, n, G)
					secondNearestNeighborNow = nn
				}
			}
			//寻找1-边最大的叶子节点，作为最终1-tree增加边
			if secondMinWeightNow > maxSecondMinWeight {
				maxSecondMinWeight = secondMinWeightNow
				secondNearestNeighbor = secondNearestNeighborNow
				chosenLeafNode = n
			}
		}
		n.oneTreeSucc = nil
	}

	chosenLeafNode.oneTreeSucc = secondNearestNeighbor
	chosenLeafNode.degree += 1
	secondNearestNeighbor.degree += 1
	return mstCost + maxSecondMinWeight - 2*sumPi, nil
}

// ascent finds the maximum 1-tree by solving the relaxed dual of TSP by the sub-gradient method.
func ascent(nodes []*Node, G [][]int) (cost int, err error) {
	bestCost, err := min1Tree(nodes, G)
	if err != nil {
		return 0, fmt.Errorf("ascent: error in calculating the 1-tree cost: %w", err)
	}
	N := len(nodes)
	V := make([]int, N)
	lastV := make([]int, N)
	bestPi := make([]int, N)
	minDistance := math.MaxInt64
	for i := 0; i < N; i++ {
		V[i] = nodes[i].degree - 2
		lastV[i] = V[i]
		nodes[i].pi = 0
		bestPi[i] = 0
		for _, d := range G[i] {
			if d > 0 && d < minDistance {
				minDistance = d
			}
		}
	}
	t := 100
	periodNum := int(math.Max(float64(N/2), MLoop))
	//firstPeriod := true
	bestVNorm2 := math.MaxInt64
	for round := 0; round < len(nodes) && t > 0; round++ {
		//每一步，node的pi值根据上一步和上上步的V值调整，V值为每一步结束后1-tree的度距2的绝对值。比2大，pi值增加，增加1-tree总值成为负担，不再偏向经过这个node，反之亦然。
		for step := 0; step < periodNum; step++ {
			VNorm2 := 0
			for _, v := range V {
				VNorm2 += v * v
			}
			if VNorm2 <= 0.0 {
				break
			}
			for i := 0; i < N; i++ {
				if V[i] != 0 {
					nodes[i].pi += t * (7*V[i] + 3*lastV[i]) / 10
				}
			}
			Pi := make([]int, N, N)
			for i := 0; i < N; i++ {
				Pi[i] = nodes[i].pi
			}
			newCost, err := min1Tree(nodes, G)
			if err != nil {
				return 0, fmt.Errorf("ascent: error in calculating the 1-tree cost: %w", err)
			}
			//degree := make([]int, N)
			//for i, n := range nodes {
			//	degree[i] = n.degree
			//}
			//fmt.Printf("round: %d; step: %d; t: %d; newCost: %.2f; bestCost: %.2f; VNorm2: %d, V: %v, Pi: %v\n", round, step, t, float64(newCost)/float64(PRECISION), float64(bestCost)/float64(PRECISION), VNorm2, V, Pi)
			//求最大的，各node都尽量接近度为2的最小1-tree。因为1-tree是最优解的下界，最大下界即逼近最优解。
			if (newCost > bestCost) || (newCost == bestCost && VNorm2 < bestVNorm2) {
				bestCost = newCost
				if VNorm2 < bestVNorm2 {
					bestVNorm2 = VNorm2
				}
				for i := 0; i < N; i++ {
					bestPi[i] = nodes[i].pi
				}
				//if firstPeriod {
				//	t += t / 10
				//}
				//if step == period-1 {
				//	period *= 2
				//	if period > periodNum {
				//		period = periodNum
				//	}
				//}
			}

			//else if firstPeriod && step > period/2 {
			//	firstPeriod = false
			//	step = -1
			//	t = 3 * t / 4
			//}
			lastV = append(make([]int, 0), V...)
			for i := 0; i < N; i++ {
				V[i] = nodes[i].degree - 2
			}
		}
		t = t * getKT(len(nodes)) / 100
	}
	for i := 0; i < N; i++ {
		nodes[i].pi = bestPi[i]
	}
	cost, err = min1Tree(nodes, G)
	if err != nil {
		return 0, fmt.Errorf("ascent: error in calculating the 1-tree cost: %w", err)
	}
	//log.Println(bestPi)

	return cost, nil
}

// genCandidates decides the candidate neighbors of each node.
func genCandidates(nodes []*Node, G [][]int, maxCandidates int, maxNearness float64) (err error) {
	// make sure the special node exists
	hasSpecialNode := false
	for _, n := range nodes {
		if isSpecialNode(n) {
			hasSpecialNode = true
			break
		}
	}
	if !hasSpecialNode {
		return fmt.Errorf("genCandidates: the 1-tree does not have a special node")
	}
	// sort the nodes by topological order
	sort.SliceStable(nodes, func(i, j int) bool {
		return nodes[i].mstRank < nodes[j].mstRank
	})
	N := len(nodes)
	candidates := make([][]candidate, N)
	for i := 0; i < N; i++ {
		candidates[i] = make([]candidate, 0)
	}
	for _, from := range nodes {
		mark := make([]bool, N) // the array indicating whether a node is in the dad-path
		beta := make([]int, N)
		for i := range beta {
			beta[i] = -math.MaxInt64
		}
		if !isSpecialNode(from) {
			to := from
			for to.mstDad != nil { //对每个from，重计算其父系节点的beta，更新为某父系节点到该节点路线的最大边
				beta[to.mstDad.idx] = int(math.Max(float64(beta[to.idx]), float64(calDis(to, to.mstDad, G))))
				mark[to.mstDad.idx] = true
				to = to.mstDad
			}
		}
		for _, to := range nodes {
			if to == from {
				continue
			}
			alpha := 0
			if isSpecialNode(from) {
				if from.mstDad != to { //如果from点是原先虚拟边的叶子点，则删虚拟边，加from到to边
					alpha = calDis(from, to, G) - calDis(from, from.oneTreeSucc, G)
				}
			} else if isSpecialNode(to) { //如果to点是原先虚拟边的叶子点，则删虚拟边，加from到to边
				if to.mstDad != from {
					alpha = calDis(from, to, G) - calDis(to, to.oneTreeSucc, G)
				}
			} else { //如果都不是，则减to点的beta，加from到to边
				if !mark[to.idx] { //如果to不在from的父系路径上
					beta[to.idx] = int(math.Max(float64(beta[to.mstDad.idx]), float64(calDis(to, to.mstDad, G))))
				}
				alpha = calDis(from, to, G) - beta[to.idx] //加from到to边，如果to在from的父系路径上，删to到from节点(不经过根结点)最大的实边；如果to不在from的父系路径上，删from到to(经过根结点)的route的最大实边
			}
			//if alpha < 0 {
			//	return fmt.Errorf("genCandidates: negative alpha value of edge (%d, %d)", from.idx, to.idx)
			//}
			//if float64(alpha) < maxNearness {
			candidates[from.idx] = append(candidates[from.idx], candidate{
				to:    to,
				alpha: alpha,
			})
			//}
		}
	}

	// truncate the node.candidates according to maxCandidates
	//log.Printf("genCandidates: sort and truncate the candidates")
	for _, n := range nodes {
		nodeCandidates := candidates[n.idx]
		candidates[n.idx] = nil
		// sort the candidates by alpha values in ascending order
		sort.SliceStable(nodeCandidates, func(i, j int) bool {
			return nodeCandidates[i].alpha < nodeCandidates[j].alpha
		})
		candidateNum := len(nodeCandidates)
		if candidateNum > maxCandidates {
			candidateNum = maxCandidates
		}
		if n.fixNode1 != nil {
			n.candidates = append(n.candidates, n.fixNode1)
			candidateNum -= 1
		}
		if n.fixNode2 != nil {
			n.candidates = append(n.candidates, n.fixNode2)
		}
		for i := 0; i < candidateNum; i++ {
			n.candidates = append(n.candidates, nodeCandidates[i].to)
		}
	}

	return nil
}
