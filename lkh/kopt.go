package lkh

import (
	"math"
	"math/rand"
	"sort"
	"time"
)

// make edge by two nodes
func makeEdge(n1 *Node, n2 *Node, tour *Tour) Edge {

	if n1.idx > n2.idx {
		return Edge{
			first:  n2,
			second: n1,
		}
	} else {
		return Edge{
			first:  n1,
			second: n2,
		}
	}

}

type closestInfo struct {
	node *Node
	gi   int
}

type closestList []*closestInfo

func (s closestList) Len() int           { return len(s) }
func (s closestList) Less(i, j int) bool { return s[i].gi > s[j].gi }
func (s closestList) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

func (t *Tour) closest(ni *Node, destroyed map[Edge]bool, repaired map[Edge]bool, gain int) closestList {
	cL := make([]*closestInfo, 0)
	rand.Seed(time.Now().UnixNano())
	for _, node := range ni.candidates {
		yi := makeEdge(ni, node, t)
		if t.graph[ni.idx][node.idx] <= 0 || t.edges[yi] { //路网小于零或已存在此边，放弃
			continue
		}
		gi := gain - calDis(ni, node, t.graph)
		coe := SolCoefficient
		if t.symmetrical {
			coe = coe * 1000
		}
		if destroyed[yi] || repaired[yi] || (gi <= 0 && rand.Float64() > math.Exp(-coe*float64(len(destroyed)))) { //add集和destory集中存在或收益小于零，放弃
			//if destroyed[yi] || repaired[yi] || gi <= 0 { //add集和destory集中存在或收益小于零，放弃
			continue
		}
		tmp := &closestInfo{
			node: node,
			gi:   gi,
		}
		cL = append(cL, tmp)
	}
	sort.Sort(closestList(cL))
	return cL
}

// choose destroyed edges in LKH
//对last点的存在实边，作为新增加的一条摧毁边，并根据这条边终点回连第一个点加入修复边。若成功，返回，若失败，删除回连的修复边，继续寻找修复边，若修复边寻找也失败，删除新增的摧毁边。直到last点所有存在边均试过，若失败，返回失败
func chooseX(tour *Tour, n1 *Node, last *Node, destroyed map[Edge]bool, repaired map[Edge]bool, gain, k int, endTime time.Time) bool {
	if k > getK(tour.symmetrical) || time.Since(endTime) > 0 { //超迭代深度直接返回
		return false
	}
	around := tour.around(last)
	for _, n2i := range around {
		if isEdgeFixed(last, n2i) {
			continue
		}
		xi := makeEdge(last, n2i, tour)
		gi := gain + calDis(last, n2i, tour.graph)
		if !destroyed[xi] && !repaired[xi] {
			destroyed[xi] = true
			relinkEdge := makeEdge(n2i, n1, tour)
			if repaired[relinkEdge] || tour.edges[relinkEdge] { //若此边在repair集中已存在或真实存在
				delete(destroyed, xi)
				continue
			}
			repaired[relinkEdge] = true
			relinkGi := gi - calDis(n2i, n1, tour.graph)
			isTour, newPath := tour.buildNewTour(destroyed, repaired)

			// if 2opt already and it can't be a whole tour, quit the node
			rand.Seed(time.Now().UnixNano())
			coe := DeepCoefficient
			if tour.symmetrical {
				coe = coe * 1000
			}
			if !isTour && len(destroyed) > 2 && rand.Float64() > math.Exp(coe*float64(1-len(destroyed))) {
				//if !isTour && len(destroyed) > 2 {
				delete(destroyed, xi)
				delete(repaired, relinkEdge)
				continue
			}

			// Todo: 判断solution是否在
			if isTour && relinkGi > 0 {
				tour.path = newPath
				//fmt.Println(newPath)
				//fmt.Printf("%d-opt with gain %d\n", len(destroyed), relinkGi)
				var distance int
				distance += calDis(tour.path[0], tour.path[len(tour.path)-1], tour.graph)
				for ind, node := range tour.path {
					if ind == len(tour.path)-1 {
						break
					} else {
						distance += calDis(node, tour.path[ind+1], tour.graph)
					}
				}
				tour.dis = float64(distance) / float64(PRECISION)
				//fmt.Printf("tour distance is %.2f\n", float64(distance)/float64(PRECISION))
				return true
			} else {
				delete(repaired, relinkEdge)
				status := chooseY(tour, n1, n2i, destroyed, repaired, gi, k, endTime)
				if status {
					return true
				} else {
					delete(destroyed, xi)
				}
				//return status
			}
		}
	}
	return false
}

// choose repaired edges in LKH
//对n2i点的candidate集合，选择一条成功的修复边，返回寻找新的摧毁边对(k值+1，即迭代深度+1)。若寻找新摧毁边对失败 或 没有合格的修复边，则删选择的修复边，返回（相当于不做任何变化）
func chooseY(tour *Tour, n1 *Node, n2i *Node, destroyed map[Edge]bool, repaired map[Edge]bool, gain, k int, endTime time.Time) bool {
	cL := tour.closest(n2i, destroyed, repaired, gain)
	for _, cN := range cL {
		yi := makeEdge(n2i, cN.node, tour)
		repaired[yi] = true
		gi := gain - calDis(n2i, cN.node, tour.graph)
		if chooseX(tour, n1, cN.node, destroyed, repaired, gi, k+1, endTime) {
			return true
		}
		delete(repaired, yi)
		//if ind >= 5 {
		//	return false
		//}
	}
	return false
}

// main process in LKH
func optimize(tour *Tour, G [][]int, endTime time.Time) bool {
	tour.basicInit()
	for _, n1 := range tour.path {
		around := tour.around(n1)
		for _, n2 := range around {
			if isEdgeFixed(n1, n2) {
				continue
			}
			destroyed := map[Edge]bool{makeEdge(n1, n2, tour): true}
			gain := calDis(n1, n2, G)
			//cL := tour.closest(n2, destroyed, make(map[Edge]bool), gain)
			for _, n3 := range n2.candidates {
				if n3.isIn(tour.around(n2)) {
					continue
				}
				gain1 := gain - calDis(n2, n3, G)
				repaired := map[Edge]bool{makeEdge(n2, n3, tour): true}
				//stat, newT := tour.buildNewTour(destroyed, repaired)
				if chooseX(tour, n1, n3, destroyed, repaired, gain1, 2, endTime) {
					return true
				}
				//if ind >= 4 {
				//	break
				//}
			}
		}
	}
	return false
}
