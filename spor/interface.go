package solver

type Node struct {
	Lat  float64
	Lng  float64
	Zone []int
}

func NewNode() *Node {
	return &Node{}
}

type EdgeInfo struct {
	Distance [][]float64
	Duration [][]float64
}

func NewEdgeInfo() *EdgeInfo {
	return &EdgeInfo{}
}

type Task struct {
	NodeIdx int
	ReqCap  []float64
}

func NewTask() *Task {
	return &Task{}
}

type Resource struct {
	Quantity int
	Capacity []float64
	CostTier [][]float64
}

func NewResource() *Resource {
	return &Resource{}
}

type SolverInput struct {
	Tasks         []*Task
	Resources     []*Resource
	Nodes         []*Node
	EdgeInfos     []*EdgeInfo
	Obj           int
	CalTime       int64 //第一阶段结束时间戳
	SolverEndTime int64 //第二阶段结束时间戳
	TaskId        string
}

func NewSolverInput() *SolverInput {
	return &SolverInput{}
}

type SolverOutput struct {
	Asgmts       [][]int
	Seqs         [][][]int
	InfeaTasks   [][]int
	UnasgTasks   [][]int
	Flag         int
	PanicMessage string
}
