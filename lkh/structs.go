package lkh

import (
	"fmt"
)

type TSPSolverInput struct {
	Names          []string
	DistanceMatrix [][]float64
	FixedEdges     [][2]int
	StartIdx       int
	EndIdx         int
	LeftTime       int
	ISSymmetrical  bool
}
type TSPSolverOutput struct {
	Sequence []int
	Distance float64
	Retcode  RetCode
	Err      error
}

// Node is a "city" in the travelling salesman problem.
type Node struct {
	// idx is the index of this node of a global node list
	idx int
	// name of a node
	name string
	// mstDad is the dad node of this node in the minimum spanning tree
	mstDad *Node
	// mstRank represents the topological order
	mstRank int
	// degree is the degree of this node in the minimum 1-tree
	degree int
	// candidates are the promising edges incident to this node
	candidates []*Node
	// oneTreeSucc is the "successor" of this node in the 1-tree, if this node is the special one
	oneTreeSucc *Node
	// pi is the dual variable of the relaxed TSP
	pi int
	// The edge from this node to fixNode1 must be in the final tour
	fixNode1 *Node
	// The edge from node to fixNode2 must be in the final tour
	fixNode2 *Node
}

// isIn returns true if this node is an element of a given list of nodes.
func (n *Node) isIn(nodeList []*Node) bool {
	for _, node := range nodeList {
		if n.idx == node.idx {
			return true
		}
	}
	return false
}

func (n *Node) String() string {
	return fmt.Sprintf("%d", n.idx)
}

// A candidate is a promising subset of nodes connected to this node.
type candidate struct {
	to    *Node
	alpha int
}

// An Edge is the connection between two nodes.
type Edge struct {
	first  *Node
	second *Node
}

func (e Edge) String() string {
	return fmt.Sprintf("%d-%d", e.first.idx, e.second.idx)
}

type Tour struct {
	graph       [][]int
	path        []*Node
	size        int
	edges       map[Edge]bool
	idxMap      map[*Node]int
	dis         float64
	symmetrical bool
}

// find nodes that nearby the input node
func (t *Tour) around(node *Node) []*Node {
	idx := t.idxMap[node]
	pred := idx - 1
	succ := idx + 1
	if succ == t.size {
		succ = 0
	}
	if pred < 0 {
		pred += t.size
	}
	return []*Node{t.path[pred], t.path[succ]}
}

// init the edges from a path slice
func (t *Tour) basicInit() {
	idxMap := make(map[*Node]int)
	for idx, node := range t.path {
		idxMap[node] = idx
	}
	t.idxMap = idxMap
	edges := make(map[Edge]bool)
	for idx, node := range t.path {
		if idx < t.size-1 {
			tmpEdge := makeEdge(node, t.path[idx+1], t)
			edges[tmpEdge] = true
		} else {
			//makeEdge(node, t.path[0], t)
			tmpEdge := makeEdge(node, t.path[0], t)
			edges[tmpEdge] = true
		}
	}
	t.edges = edges
}

// build a new tour by destroy & repair edges
func (t *Tour) buildNewTour(destroy map[Edge]bool, repair map[Edge]bool) (bool, []*Node) {
	tmpEdges := make(map[Edge]bool, 0)
	for e := range t.edges {
		tmpEdges[e] = true
	}

	for dE := range destroy {
		delete(tmpEdges, dE)
	}
	for rE := range repair {
		tmpEdges[rE] = true
	}
	if len(tmpEdges) < t.size { //判断不成一个环的标准1：边数小于点数
		return false, nil
	}

	successors := make(map[*Node]*Node)
	node := t.path[0]
	edgeSilces := make([][]*Node, t.size)
	for edge := range tmpEdges {
		if len(edgeSilces[edge.first.idx]) == 0 {
			edgeSilces[edge.first.idx] = append(edgeSilces[edge.first.idx], edge.first)
		}
		edgeSilces[edge.first.idx] = append(edgeSilces[edge.first.idx], edge.second)
		if len(edgeSilces[edge.second.idx]) == 0 {
			edgeSilces[edge.second.idx] = append(edgeSilces[edge.second.idx], edge.second)
		}
		edgeSilces[edge.second.idx] = append(edgeSilces[edge.second.idx], edge.first)
	}
	preAppear := make([]bool, t.size)
	postAppear := make([]bool, t.size)
	preNode := t.path[0]
	postNode := edgeSilces[0][1]
	for i := 0; i < t.size; i++ {
		successors[preNode] = postNode
		preAppear[preNode.idx] = true
		postAppear[postNode.idx] = true
		for j := 1; j < len(edgeSilces[postNode.idx]); j++ { //edgeSilces必定为n*3规模，因为上一解保证所有点度为2
			if edgeSilces[postNode.idx][j].idx != preNode.idx {
				preNode = postNode
				postNode = edgeSilces[postNode.idx][j]
				break
			}
		}
	}
	for i := 0; i < t.size; i++ {
		if !preAppear[i] || !postAppear[i] { //有不在第一个环内的点
			return false, nil
		}
	}

	//fmt.Print(successors)
	succ := successors[node]
	newPath := make([]*Node, 0)
	//fmt.Println("\n*****")
	newPath = append(newPath, node)
	visited := make(map[*Node]bool)
	visited[node] = true

	for !visited[succ] {
		newPath = append(newPath, succ)
		visited[succ] = true
		succ = successors[succ]
	}
	return true, newPath
}
