package waveoptsolver

// retCodes
const (
	SuccessResult                 = int64(0)
	InvalidInputDateResult        = int64(1)
	UnreasonableConstraintsResult = int64(2)
	InfeasibleSolutionResult      = int64(3)
	ExternalTermination           = int64(4)
	UnknownSolverErrorResult      = int64(100)
)

// configs
const fixedProcessingSeconds = int64(5)
const isLoggingToConsole = true
const reservedOrderDistanceCoeff = 5.0

type WavePickerMode string

const (
	SinglePickerOnly                            WavePickerMode = "SinglePickerOnly"
	MultiPickerAtMWSWithTotalOrderQty           WavePickerMode = "MultiPickerAtMWSWithTotalQty"
	MultiPickerAtMWSWithRespectiveOrderQty      WavePickerMode = "MultiPickerAtMWSWithRespectiveQty"
	MultiPickerAtMLWithTotalPickingTaskQty      WavePickerMode = "MultiPickerAtMLWithTotalQty"
	MultiPickerAtMLWithRespectivePickingTaskQty WavePickerMode = "MultiPickerAtMLWithRespectiveQty"
)

type SplitLevel string

const (
	ByZone        SplitLevel = "ByZone"
	ByZoneCluster SplitLevel = "ByZoneCluster"
	ByZoneSector  SplitLevel = "ByZoneSector"
)

type GroupType string

const (
	ExtraBulkyPickingOrderType       GroupType = "ExtraBulkyPickingOrderType"
	HighValuePickingOrderType        GroupType = "HighValuePickingOrderType"
	StorageHighValuePickingOrderType GroupType = "StorageHighValuePickingOrderType"
	BulkyPickingOrderType            GroupType = "BulkyPickingOrderType"
	StorageBulkyPickingOrderType     GroupType = "StorageBulkyPickingOrderType"
	NormalPickingOrderType           GroupType = "NormalPickingOrderType"
	StorageNormalPickingOrderType    GroupType = "StorageNormalPickingOrderType"
	MedicalPickingOrderType          GroupType = "MedicalPickingOrderType"
	FreshOrderType                   GroupType = "FreshOrderType"
	StorageFreshOrderType            GroupType = "StorageFreshOrderType"
)

type SkuSizeType string

const (
	NonBulkySkuType   SkuSizeType = "NonBulkySku"
	BulkySkuType      SkuSizeType = "BulkySku"
	ExtraBulkySkuType SkuSizeType = "ExtraBulkySku"
)

type PickingTaskPickerModeType string

const (
	SinglePickerType PickingTaskPickerModeType = "SinglePicker"
	MultiPickerType  PickingTaskPickerModeType = "MultiPicker"
)

type TaskSizeType string

const (
	NonBulkyTaskType   TaskSizeType = "NonBulkyTask"
	BulkyTaskType      TaskSizeType = "BulkyTask"
	ExtraBulkyTaskType TaskSizeType = "ExtraBulkyTask"
)

var (
	RandomOrderChangeMoveFactoryWeight = 0.3
	RandomOrderSwapMoveFactoryWeight   = 0.7
	ConvergenceTimeRatio               = 0.2
)

type WaveTypeEnum string

const (
	SingleSkuSingleQty = "SingleSkuSingleQty"
	SameSkuSameQty     = "SameSkuSameQty"
	MixSkuAnyQty       = "MixSkuAnyQty"
	SingleSkuAnyQty    = "SingleSkuAnyQty"
	MixSkuSingleQty    = "MixSkuSingleQty"
)

type TaskSortCondType int64

const (
	MultiOrderNumInPickingTaskDesc TaskSortCondType = iota
	UrgentOrderNumInPickingTaskDesc
	TotalItemNumInPickingTaskDesc
	TotalOrderNumInPickingTaskDesc
	DistancePerOrderInPickingTaskDesc
	DistancePerItemPickingTaskDesc
)

var DefaultTaskSorCondList = []TaskSortCondType{
	MultiOrderNumInPickingTaskDesc,
	UrgentOrderNumInPickingTaskDesc,
	TotalItemNumInPickingTaskDesc,
	TotalOrderNumInPickingTaskDesc,
	DistancePerOrderInPickingTaskDesc,
	DistancePerItemPickingTaskDesc,
}

type ObjFuncType string

const (
	WeightedZoneObjFuncType ObjFuncType = "WeightedZoneObjFunc"
	MapDistanceObjFuncType  ObjFuncType = "MapDistanceObjFunc"
)
