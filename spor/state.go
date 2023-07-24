package solver

type GFun struct {
	UpdateFeat          func([]float64, FeatParam)
	InitCapUseJob       func(*CapUse)
	InitCapUseTask      func(*CapUse, *GPara)
	InitResUse          func(*GPara, *GState) []int
	UpdateResUse        func([]int, int) []int
	ResetCapUse         func(*CapUse)
	UpdateCapUse        func(*GPara, *CapUse, int, int, int, int)
	CheckCap            func(*GPara, *CapUse, int) bool
	CheckMaxCap         func(*GPara, *CapUse, int, int, int, int) bool
	CheckResCap         func(*GPara, *GState) (bool, int)
	GetSolFeat          func() [][][]float64
	GetProxTask         func(*GPara) [][][]int
	GetDefaultIdxSorted func(*GPara) []int
	GetObj              func(*GState, *GPara) float64
}

type GState struct {
	// state
	Flag int // default is 0

	DimFeat int
	Feats   [][][]float64 // 每条route order的质心的经纬度

	ResUse []int //每种车型使用了多少辆

	InnerSeqs       [][]int
	InnerAsgmts     []int
	InnerInfeaTasks []int
	InnerUnasgTasks []int
	InnerFeats      [][]float64
	InnerSeqDtls    [][]float64
	//InnerSeqDtls[0] -- distance
	//InnerSeqDtls[1] -- parcel
	//InnerSeqDtls[2] -- weight
	//InnerSeqDtls[3] -- duration
	//InnerSeqDtls[4] -- FitCost
	//InnerSeqDtls[5] -- MapCost

	BestInnerSeqs       [][]int
	BestInnerAsgmts     []int
	BestInnerUnasgTasks []int
	BestInnerInfeaTasks []int
	BestInnerSeqDtls    [][]float64

	// decision
	Asgmts     [][]int   // resGroup jobs
	TaskGroups [][]int   // resGroup tasks
	Seqs       [][][]int // task sequence (including res assign)
	InfeaTasks [][]int
	UnasgTasks [][]int
	SeqDtl     [][][]float64

	ColdStart    bool
	PanicMessage string
	LSLoopTime   int64
	IsTimeEnough bool
}

type FeatParam struct {
	Length int
	Node   int
}

type CapUse struct {
	I []int
	F []float64
}

// CapUse.F[0] distance
// CapUse.F[1] parcel
// CapUse.F[2] weight
// CapUse.F[3] duration
// CapUse.F[4] cost
