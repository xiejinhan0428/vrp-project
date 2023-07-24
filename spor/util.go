package solver

import (
	"math"
	"math/rand"
	"time"
)

const (
	// Radius 地球半径，米
	Radius = 6371 * 1000
	// Diameter 地球直径，米
	Diameter = Radius * 2
)

// Radians 将角度转为弧度
func Radians(degree float64) float64 {
	return degree * math.Pi / 180
}

// GreatCircle great circle distance calculation
func GreatCircleDistance(first []float64, second []float64) float64 {
	leftLat := Radians(first[0])
	rightLat := Radians(second[0])
	diffLat := leftLat - rightLat
	diffLng := Radians(first[1]) - Radians(second[1])
	product := math.Cos(leftLat) * math.Cos(rightLat) * math.Pow(math.Sin(diffLng/2), 2)
	square := math.Sqrt(math.Pow(math.Sin(diffLat/2), 2) + product)
	return float64(math.Asin(square) * Diameter)
}

type SeqDtl struct {
	Distance float64
	Parcels  float64
	Weight   float64
	Duration float64
	FitCost  float64
	MapCost  float64
}

func GenerateSeqDtls(seq [][]int, asgmts []int, gPara *GPara) [][]float64 {
	res := [][]float64{}
	for i := 0; i < len(seq); i++ {
		var totalParcel float64 = 0.0
		var totalDistance float64 = GetSeqDistance(seq[i], i, asgmts, gPara)
		var totalDuration float64 = 0.0
		var totalWeight float64 = 0.0
		var totalFitCost float64 = 0.0
		var totalMapCost float64 = 0.0
		for j := 0; j < len(seq[i]); j++ {
			totalParcel += gPara.CapTask[seq[i][j]][0]
			totalWeight += gPara.CapTask[seq[i][j]][1]
			totalDuration += gPara.CapTask[seq[i][j]][2]
		}
		totalDuration += GetSeqDuration(seq[i], i, asgmts, gPara)
		totalFitCost = GetSeqFitCost(gPara, asgmts[i], totalDistance)
		totalMapCost = GetSeqMapCost(gPara, asgmts[i], totalDistance)
		res = append(res, []float64{totalDistance, totalParcel, totalWeight, totalDuration, totalFitCost, totalMapCost})
	}
	return res
}

func GetSeqDtl(seq []int, sIdx int, asgmts []int, gPara *GPara) []float64 {
	var seqDtl = make([]float64, 0)
	var totalParcel float64 = 0.0
	var totalDistance float64 = GetSeqDistance(seq, sIdx, asgmts, gPara)
	var totalDuration float64 = 0.0
	var totalWeight float64 = 0.0
	var totalFitCost float64 = 0.0
	var totalMapCost float64 = 0.0
	for j := 0; j < len(seq); j++ {
		totalParcel += gPara.CapTask[seq[j]][0]
		totalWeight += gPara.CapTask[seq[j]][1]
		totalDuration += gPara.CapTask[seq[j]][2]
	}
	totalDuration += GetSeqDuration(seq, sIdx, asgmts, gPara)
	totalFitCost, totalMapCost = GetSeqMFCostByDist(gPara, asgmts[sIdx], totalDistance)
	seqDtl = append(seqDtl, []float64{totalDistance, totalParcel, totalWeight, totalDuration, totalFitCost, totalMapCost}...)
	return seqDtl
}

func GetSeqConsDtl(gPara *GPara, seq []int, resIdx int) SeqDtl {
	var tmpDist, tmpPar, tmpWeight, tmpDur, tmpFCst, tmpMCst float64 = 0.0, 0.0, 0.0, 0.0, 0.0, 0.0
	for i, taskIdx := range seq {
		if i == 0 {
			tmpDist += gPara.Cost[resIdx][0][0][gPara.TaskLoc[taskIdx]]
			tmpDur += gPara.Cost[resIdx][1][0][gPara.TaskLoc[taskIdx]] + gPara.CapTask[taskIdx][2]
		} else {
			tmpDist += gPara.Cost[resIdx][0][gPara.TaskLoc[seq[i-1]]][gPara.TaskLoc[taskIdx]]
			tmpDur += gPara.Cost[resIdx][1][gPara.TaskLoc[seq[i-1]]][gPara.TaskLoc[taskIdx]] + gPara.CapTask[taskIdx][2]
		}
		tmpPar += gPara.CapTask[taskIdx][0]
		tmpWeight += gPara.CapTask[taskIdx][1]
	}
	tmpDist += gPara.Cost[resIdx][0][gPara.TaskLoc[seq[len(seq)-1]]][0]
	tmpDur += gPara.Cost[resIdx][1][gPara.TaskLoc[seq[len(seq)-1]]][0]

	tmpFCst, tmpMCst = GetSeqMFCostByDist(gPara, resIdx, tmpDist)
	seqDtl := SeqDtl{
		Distance: tmpDist,
		Parcels:  tmpPar,
		Weight:   tmpWeight,
		Duration: tmpDur,
		FitCost:  tmpFCst,
		MapCost:  tmpMCst,
	}
	return seqDtl
}

func CheckSeqsDtl(gPara *GPara, state *GState) []SeqDtl {
	var seqDtls = make([]SeqDtl, len(state.InnerSeqs))
	for i := 0; i < len(state.InnerSeqs); i++ {
		resIdx := state.InnerAsgmts[i]
		seqDtls[i] = GetSeqConsDtl(gPara, state.InnerSeqs[i], resIdx)
	}
	return seqDtls
}

// *********************************************
// reverse slice s[i1:i2+1]
func reverse(s []int, i1 int, i2 int) {
	for p, q := i1, i2; p < q; p, q = p+1, q-1 {
		s[p], s[q] = s[q], s[p]
	}
}

func RouletteMultiSelectForFloat(weights []float64, number int) []int {
	rand.Seed(time.Now().UnixNano())

	var valueSum float64
	weightLen := len(weights)
	var selectedList []int
	weightsBool := make([]bool, weightLen)
	for ind, _ := range weights {
		weightsBool[ind] = true
	}
	for i := 0; i < number; i++ {
		valueSum = 0
		for ind, value := range weights {
			if weightsBool[ind] {
				valueSum += value
			}
		}
		if valueSum <= 0 {
			valueSum = 1
		}
		//rand.Seed(time.Now().UnixNano())
		randValue := float64(rand.Intn(int(valueSum) + 1))
		for indWeight, weight := range weights {
			if weightsBool[indWeight] {
				randValue -= weight
				if randValue <= 0 {
					weightsBool[indWeight] = false
					selectedList = append(selectedList, indWeight)
					break
				}
			}
		}
	}
	return selectedList
}

func Float64Tofloat64(slice64 []float64) []float64 {
	var slice32 = make([]float64, len(slice64))
	for i := 0; i < len(slice64); i++ {
		slice32[i] = float64(slice64[i])
	}
	return slice32
}

func float64ToFloat64(slice32 []float64) []float64 {
	var slice64 = make([]float64, len(slice32))
	for i := 0; i < len(slice32); i++ {
		slice64[i] = float64(slice32[i])
	}
	return slice64
}

func Maxfloat64(v1 float64, v2 float64) float64 {
	if v1 <= v2 {
		return v2
	} else {
		return v1
	}
}

func Minfloat64(v1 float64, v2 float64) float64 {
	if v1 <= v2 {
		return v1
	} else {
		return v2
	}
}

func CopyMatrixI(m [][]int) (mCp [][]int) {
	mCp = make([][]int, len(m))
	for i := 0; i < len(m); i++ {
		mCp[i] = CopySliceInt(m[i])
	}
	return
}

func CopyMatrixF(m [][]float64) (mCp [][]float64) {
	mCp = make([][]float64, len(m))
	for i := 0; i < len(m); i++ {
		mCp[i] = CopySliceFloat(m[i])
	}
	return
}

func CopySliceInt(s []int) (sDup []int) {
	sDup = make([]int, len(s))
	for i := 0; i < len(s); i++ {
		sDup[i] = s[i]
	}
	return
}

func copyPartBack(ssnew [][]int, ss [][]int) {
	for i := 0; i < len(ssnew); i++ {
		ss[i] = ss[i][:0]
		for j := 1; j < len(ssnew[i])-1; j++ {
			ss[i] = append(ss[i], ssnew[i][j]-1) // 默认taskId = nodeId - 1
		}
	}
}

func CopySlicefloat64(s []float64) (sDup []float64) {
	sDup = make([]float64, len(s))
	for i := 0; i < len(s); i++ {
		sDup[i] = s[i]
	}
	return
}

func findMin(s []float64) (idx int, v float64) {
	v = 1e8
	idx = -1
	for i := 0; i < len(s); i++ {
		if s[i] < v {
			v = s[i]
			idx = i
		}
	}
	return
}

func maxSlice32(s []float64) float64 {
	var smax float64
	for i := 0; i < len(s); i++ {
		if smax < s[i] {
			smax = s[i]
		}
	}
	return smax
}
func GetInnerDistance(seqs [][]int, asgmets []int, gPara *GPara) float64 {
	innerdis := 0.0
	for i := 0; i < len(seqs); i++ {
		for j := 0; j < len(seqs[i]); j++ {
			for k := 0; k < len(seqs[i]); k++ {
				innerdis += gPara.Cost[asgmets[i]][0][gPara.TaskLoc[seqs[i][j]]][gPara.TaskLoc[seqs[i][k]]]
			}
		}
	}
	return innerdis
}
func GetInnerDistance2(seqs [][]int, asgmets []int, gPara *GPara) float64 {
	average := 0.0
	cnt := 0
	for i := 0; i < len(seqs); i++ {
		innerdis := 0.0
		for j := 0; j < len(seqs[i]); j++ {
			mina := 999999999999.0
			cnt += 1
			for k := 0; k < len(seqs[i]); k++ {
				innerdis += gPara.Cost[asgmets[i]][0][gPara.TaskLoc[seqs[i][j]]][gPara.TaskLoc[seqs[i][k]]]
			}
			if len(seqs[i]) == 1 {
				innerdis = 0
			} else {
				innerdis /= float64(len(seqs[i]) - 1)
			}
			for k := 0; k < len(seqs); k++ {
				if k == i {
					continue
				}
				tmp := 0.0
				for x := 0; x < len(seqs[k]); x++ {
					tmp += gPara.Cost[asgmets[i]][0][gPara.TaskLoc[seqs[i][j]]][gPara.TaskLoc[seqs[k][x]]]
				}
				tmp /= float64(len(seqs[k]))
				mina = math.Min(mina, tmp)
			}
			sss := (mina - innerdis) / math.Max(mina, innerdis)
			average += sss
		}

	}
	average /= float64(cnt)
	return (1 - average) * 1000000
}
