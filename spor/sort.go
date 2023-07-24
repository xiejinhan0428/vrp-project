package solver

import "sort"

// IndexSorter 排序并返回原数组下标
type IndexSorter struct {
	sort.Interface
	index []int
}

// Swap 覆写原Swap，增加下标的位置交换
func (s *IndexSorter) Swap(i, j int) {
	s.Interface.Swap(i, j)
	s.index[i], s.index[j] = s.index[j], s.index[i]
}

// GetIndex 获得原数组的下标
func (s *IndexSorter) GetIndex() []int {
	return s.index
}

// ResetIndex 重设下标
func (s *IndexSorter) ResetIndex() {
	for i := range s.index {
		s.index[i] = i
	}
}

// NewIndexSorter 构造附带下标的排序器
func NewIndexSorter(sli sort.Interface, idx []int) *IndexSorter {
	//idx := make([]int, sli.Len())
	for i := range idx {
		idx[i] = i
	}
	return &IndexSorter{sli, idx}
}
