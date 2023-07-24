package solver

type GPara struct {
	Obj     int
	NJob    int     // jobs = [0,...,nJob-1] routing 不需要
	NTask   int     // tasks = [0,...,nTask-1]
	JobTask [][]int // task list for each job routing 不需要
	TaskLoc []int   // loc idx for each task idx

	//CapTask[t][0]: Parcel
	//CapTask[t][1]: Weight
	//CapTask[t][2]: Duration
	CapTask   [][]float64
	CapTaskI  [][]int
	CapResMap []int       // 入参resource idx 与CapRes idx 映射关系
	CapRes    [][]float64 //车辆基本信息，入参信息

	//CapResCost[resIdx][t][0]:LowerBound
	//CapResCost[resIdx][t][1]:UpperBound
	//CapResCost[resIdx][t][2]:outerK
	//CapResCost[resIdx][t][3]:LowerCost
	//CapResCost[resIdx][t][4]:UpperCost
	//CapResCost[resIdx][t][5]:innerK
	CapResCost [][][]float64
	CapResI    [][]int

	CapResGrp  []float64
	CapResGrpI []int
	CapSeq     []int // max len of seq

	ProxTask [][][]int // neighbors of each task 改变
	PosTask  []int     // default position for each task
	PosJob   []int     // default position for each job

	Cost [][][][]float64 // cost between loc idx [resIdx][type(0:dist,1:dur)][][]

	Nodes [][]float64 // nodes[][0]Lat  nodes[][1]Lng

	CntResGrp0    int // initial guess for resGrp cnt
	CntRes0       int // initial guess for seq cnt in each ResGroup
	CntTaskGroup0 int // initial guess for task cnt in each ResGroup routing 不需要
}
