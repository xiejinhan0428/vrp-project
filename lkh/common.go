package lkh

import (
	"fmt"
)

// isSpecialNode returns true if the node n is the special/fixed node in a 1-tree
func isSpecialNode(n *Node) bool {
	return n.oneTreeSucc != nil
}

// isEdgeFixed checks if node a and be are fixed.
func isEdgeFixed(a, b *Node) bool {
	return a.fixNode1 != nil && a.fixNode1.idx == b.idx ||
		a.fixNode2 != nil && a.fixNode2.idx == b.idx ||
		b.fixNode1 != nil && a.idx == b.fixNode1.idx ||
		b.fixNode2 != nil && a.idx == b.fixNode2.idx
}

// fixEdge makes a permanent connection between node a and b. These connections will be preserved in generating
//candidates and improving the tour. Each node can be fixed to at most two other nodes. If more than two nodes are
//attempted to be fixed to one node, an error is returned.
func fixEdge(a, b *Node) (err error) {
	if a.fixNode1 == nil {
		a.fixNode1 = b
	} else if a.fixNode2 == nil {
		a.fixNode2 = b
	} else {
		return fmt.Errorf("fixEdge: Node %d has been fixed to Node %d and Node %d, cannot be fixed to Node %d", a.idx, a.fixNode1.idx, a.fixNode2.idx, b.idx)
	}

	if b.fixNode1 == nil {
		b.fixNode1 = a
	} else if b.fixNode2 == nil {
		b.fixNode2 = a
	} else {
		return fmt.Errorf("fixEdge: Node %d has been fixed to Node %d and Node %d, cannot be fixed to Node %d", b.idx, b.fixNode1.idx, b.fixNode2.idx, a.idx)
	}

	return nil
}

// calDis returns the distance between node a and b, including the pi-value
func calDis(a *Node, b *Node, G [][]int) int {
	return G[a.idx][b.idx] + a.pi + b.pi
}
