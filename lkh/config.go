package lkh

type RetCode int

const (
	Success RetCode = iota
	InvalidDistanceMatrix
	InvalidFixedEdges
	NoLowerBound
	NoCandidates
	InvalidTour
	LossPoint
	MaxFl           = 100000000000
	MinFl           = -1000000000
	DeepCoefficient = 0.3
	SolCoefficient  = 0.3
)

const PRECISION = 100

var DefaultMaxCandidates = 7
var K = 8
var KT = 90
var MLoop float64 = 128
var SkipKopt bool

func getDefaultMaxCandidates(orderNum int) int {
	return 8
	num := orderNum / 2
	if num <= 70 {
		return 7
	}
	if num <= 74 {
		return 6
	}
	if num <= 80 {
		return 5
	}
	if num <= 85 {
		return 4
	}
	if num <= 100 {
		return 3
	}
	if num <= 110 {
		return 3
	}
	if num <= 150 {
		return 2
	}
	if num <= 200 {
		return 2
	}
	return 2
}
func getK(symmetrical bool) int {
	if symmetrical {
		return 5
	}
	return 7
}
func getKT(orderNum int) int {
	return 90
	num := orderNum / 2
	if num <= 100 {
		return 85
	}
	return 90
	if num <= 120 {
		return 70
	}
	if num <= 150 {
		return 50
	}
	if num <= 200 {
		return 30
	}
	return 20
}
