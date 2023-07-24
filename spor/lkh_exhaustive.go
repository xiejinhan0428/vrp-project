package solver

func SolveTspOnewayExhaustive(input TSPSolverInput) (output TSPSolverOutput) {
	//暴力解不支持fixedEdge，因为有可能是fixedEdge导致的错误
	//暴力解默认起始点是0号
	distanceMatrix, startIdx, endIdx := input.DistanceMatrix, input.StartIdx, input.EndIdx
	//sequence, distance := output.Sequence, output.Distance
	rawLen := len(distanceMatrix)
	//sequence = make([]int, rawLen)
	//sequence[0] = startIdx
	oriList := make([]int, rawLen-1)
	for i := 0; i < rawLen-1; i++ {
		oriList[i] = i + 1
	}
	allList := GenIntSlicePermutation(oriList)
	for i := 0; i < len(allList); i++ {
		tmp := []int{0}
		tmp = append(tmp, allList[i]...)
		allList[i] = tmp
	}
	minDis := 999999999.0
	minSeq := allList[0]
	isRoundway := true
	if endIdx != startIdx {
		isRoundway = false
	}
	for i := 0; i < len(allList); i++ {
		if endIdx >= 0 && !isRoundway && allList[i][len(allList[i])-1] != endIdx {
			continue
		}
		tmpDis := CalculateDis2(distanceMatrix, allList[i], isRoundway)
		if tmpDis < minDis {
			minDis = tmpDis
			minSeq = allList[i]
		}
	}
	return TSPSolverOutput{minSeq, minDis, Success, nil}
}
func sliceArrayCount(count int) int {
	if count == 0 {
		return 1
	}
	var result = 1
	for ; count > 0; count-- {
		result *= count
	}
	return result
}

// 生成全排列
func GenIntSlicePermutation(ori []int) [][]int {
	resSlice := make([][]int, 0, sliceArrayCount(len(ori)))
	permutation(&resSlice, ori, 0)
	return resSlice
}
func permutation(permutationSlice *[][]int, seq []int, index int) {
	if seq == nil || len(seq) == 0 {
		return
	}
	// 判断递归出口
	if index == len(seq) {
		res := make([]int, len(seq))
		copy(res, seq)
		*permutationSlice = append(*permutationSlice, res)
		return
	}

	// 循环过程
	for i := index; i < len(seq); i++ {
		seq[i], seq[index] = seq[index], seq[i] // 选定元素放在当次递归的第一个位置
		// 递归过程
		permutation(permutationSlice, seq, index+1)
		seq[i], seq[index] = seq[index], seq[i] // 使用的原数组进行的交换，保证在循环过程中元素的顺序不变
	}
}
func CalculateDis2(distanceMatrix [][]float64, seq []int, isRoundWay bool) float64 {
	lenth := len(seq)
	dis1 := 0.0
	if isRoundWay {
		dis1 = distanceMatrix[seq[lenth-1]][seq[0]]
	}
	for i := 0; i < lenth-1; i++ {
		dis1 += distanceMatrix[seq[i]][seq[i+1]]
	}
	return dis1
}
