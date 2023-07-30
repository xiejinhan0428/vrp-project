package waveoptsolver

import (
	"fmt"
)

type Wave struct {
	WaveSn       string              `json:"wave_sn"`
	WaveType     WaveTypeEnum        `json:"wave_type"`
	WaveRule     *WaveRule           `json:"wave_rule"`
	SolverConfig *WaveSolverConfig   `json:"solver_config"`
	Groups       []*WaveGroup        `json:"groups"`
	Skus         []*WaveSku          `json:"skus"`
	Locations    []*Location         `json:"locations"`
	WhsMapHelper *WarehouseMapHelper `json:"-"`
}

type WaveGroup struct {
	GroupSn   string       `json:"group_sn"`
	GroupType GroupType    `json:"group_type"`
	Priority  int64        `json:"priority"`
	Orders    []*WaveOrder `json:"orders"`
}

type WaveOrder struct {
	OrderNo    string      `json:"order_no"`
	IsUrgent   bool        `json:"is_urgent"`
	CutOffTime int64       `json:"cut_off_time"`
	Skus       []*OrderSku `json:"skus"`
}

type OrderSku struct {
	OrderNo    string `json:"order_no"`
	SkuId      string `json:"sku_id"`
	Qty        int64  `json:"qty"`
	LocationId string `json:"location_id"`
}

type WaveSku struct {
	SkuId string `json:"sku_id"`

	Length int64 `json:"length"`
	Width  int64 `json:"width"`
	Height int64 `json:"height"`

	Weight int64 `json:"weight"`

	SkuSize SkuSizeType `json:"sku_size"`
}

type Location struct {
	LocationId      string `json:"location_id"`
	ZoneSectorId    int64  `json:"zone_sector_id"`
	ZoneClusterId   int64  `json:"zone_cluster_id"`
	ZoneId          string `json:"zone_id"`
	PathwayId       string `json:"pathway_id"`
	SegmentId       string `json:"segment_id"`
	PickingSequence int64  `json:"picking_sequence"`
	MapIndex        int64  `json:"map_index"`
}

type WaveResult struct {
	WaveSn           string             `json:"wave_sn"`
	PickingTasks     []*WavePickingTask `json:"picking_tasks"`
	UnsolvedOrderNos []string           `json:"unsolved_order_nos"`
	RetCode          int64              `json:"ret_code"`
	Msg              string             `json:"msg"`
}

type WavePickingTask struct {
	PickingTaskId string `json:"picking_task_id"`

	GroupSn   string    `json:"group_sn"`
	GroupType GroupType `json:"group_type"`

	OrderNos        []string              `json:"order_nos"`
	SubPickingTasks []*WaveSubPickingTask `json:"sub_picking_tasks"`

	PickerMode PickingTaskPickerModeType `json:"picker_mode"`

	PickingTaskSize TaskSizeType `json:"picking_task_size"`
}

type WaveSubPickingTask struct {
	Id   string      `json:"id"`
	Skus []*OrderSku `json:"skus"`

	SubPickingTaskSize TaskSizeType `json:"sub_picking_task_size"`
}

type WaveRule struct {
	CommonRule     *CommonRule        `json:"common_rule"`
	WavePickerMode WavePickerMode     `json:"wave_picker_mode"`
	ModeRule       SplitLevelProvider `json:"mode_rule"`
}

type SplitLevelProvider interface {
	SplitLevel() SplitLevel
}

type CommonRule struct {
	MaxPickingTaskQtyPerWave int64 `json:"max_picking_task_qty_per_wave"`
	MinPickingTaskQtyPerWave int64 `json:"min_picking_task_qty_per_wave"`

	IsCrossZoneCluster        bool  `json:"is_cross_zone_cluster"`
	MaxOrderQtyPerPickingTask int64 `json:"max_order_qty_per_picking_task"`
	MinOrderQtyPerPickingTask int64 `json:"min_order_qty_per_picking_task"`

	MaxItemQtyPerSubPickingTask    int64 `json:"max_item_qty_per_sub_picking_task"`
	MinItemQtyPerSubPickingTask    int64 `json:"min_item_qty_per_sub_picking_task"`
	MaxItemVolumePerSubPickingTask int64 `json:"max_item_volume_per_sub_picking_task"`
	MinItemVolumePerSubPickingTask int64 `json:"min_item_volume_per_sub_picking_task"`

	MaxNonBulkyTaskVolume int64 `json:"max_non_bulky_task_volume"`
	MaxNonBulkyTaskLoad   int64 `json:"max_non_bulky_task_load"`
	MaxBulkyTaskVolume    int64 `json:"max_bulky_task_volume"`
	MaxBulkyTaskLoad      int64 `json:"max_bulky_task_load"`

	TaskSortConds []TaskSortCondType `json:"task_sort_conds"`
}

type MultiPickerAtMWSTotalQtyRule struct {
	MaxBacklogAtMWSPerWave int64      `json:"max_backlog_at_mws_per_wave"`
	PickingTaskSplitLevel  SplitLevel `json:"picking_task_split_level"`
}

func (m *MultiPickerAtMWSTotalQtyRule) SplitLevel() SplitLevel {
	return m.PickingTaskSplitLevel
}

type MultiPickerAtMWSRespectiveQtyRule struct {
	MaxBulkyBacklogAtMWSPerWave    int64      `json:"max_bulky_backlog_at_mws_per_wave"`
	MaxNonBulkyBacklogAtMWSPerWave int64      `json:"max_non_bulky_backlog_at_mws_per_wave"`
	PickingTaskSplitLevel          SplitLevel `json:"picking_task_split_level"`
}

func (m *MultiPickerAtMWSRespectiveQtyRule) SplitLevel() SplitLevel {
	return m.PickingTaskSplitLevel
}

type MultiPickerAtMLTotalQtyRule struct {
	MaxBacklogAtMLPerWave int64      `json:"max_backlog_at_ml_per_wave"`
	PickingTaskSplitLevel SplitLevel `json:"picking_task_split_level"`
}

func (m *MultiPickerAtMLTotalQtyRule) SplitLevel() SplitLevel {
	return m.PickingTaskSplitLevel
}

type MultiPickerAtMLRespectiveQtyRule struct {
	MaxBulkyBacklogAtMLPerWave    int64      `json:"max_bulky_backlog_at_ml_per_wave"`
	MaxNonBulkyBacklogAtMLPerWave int64      `json:"max_non_bulky_backlog_at_ml_per_wave"`
	PickingTaskSplitLevel         SplitLevel `json:"picking_task_split_level"`
}

func (m *MultiPickerAtMLRespectiveQtyRule) SplitLevel() SplitLevel {
	return m.PickingTaskSplitLevel
}

func (r *WaveRule) splitLevel() (SplitLevel, error) {
	if r.WavePickerMode == SinglePickerOnly {
		return "", fmt.Errorf("no split level in %v mode", SinglePickerOnly)
	} else {
		return r.ModeRule.SplitLevel(), nil
	}
}

func (r *WaveRule) isZeroMulti() bool {
	switch r.WavePickerMode {
	case SinglePickerOnly:
		return false
	case MultiPickerAtMWSWithTotalOrderQty:
		modeRule, _ := r.ModeRule.(*MultiPickerAtMWSTotalQtyRule)
		return modeRule.MaxBacklogAtMWSPerWave == 0
	case MultiPickerAtMWSWithRespectiveOrderQty:
		modeRule, _ := r.ModeRule.(*MultiPickerAtMWSRespectiveQtyRule)
		return modeRule.MaxBulkyBacklogAtMWSPerWave == 0 && modeRule.MaxNonBulkyBacklogAtMWSPerWave == 0
	case MultiPickerAtMLWithTotalPickingTaskQty:
		modeRule, _ := r.ModeRule.(*MultiPickerAtMLTotalQtyRule)
		return modeRule.MaxBacklogAtMLPerWave == 0
	case MultiPickerAtMLWithRespectivePickingTaskQty:
		modeRule, _ := r.ModeRule.(*MultiPickerAtMLRespectiveQtyRule)
		return modeRule.MaxBulkyBacklogAtMLPerWave == 0 && modeRule.MaxNonBulkyBacklogAtMLPerWave == 0
	default:
		return false
	}
}

type WaveSolverConfig struct {
	MaxWaveSnSolveOrder int64       `json:"max_wave_sn_solve_order"`
	MaxSecondsSpent     int64       `json:"max_seconds_spent"`
	Parallelism         int64       `json:"parallelism"`
	VariableTabuTenure  int64       `json:"variable_tabu_tenure"`
	ValueTabuTenure     int64       `json:"value_tabu_tenure"`
	ZoneCoeff           float64     `json:"zone_coeff"`
	PathwayCoeff        float64     `json:"pathway_coeff"`
	SegmentCoeff        float64     `json:"segment_coeff"`
	ObjFunc             ObjFuncType `json:"obj_func"`
}

type WaveResultStat struct {
	PickingTaskNum         int64
	SubPickingTaskNum      int64
	MultiPickingTaskNum    int64
	SinglePickingTaskNum   int64
	MultiSubPickingTaskNum int64

	OrderNum                      int64
	UnsolvedOrderNum              int64
	MultiPickingTaskOrderNum      int64
	MultiPickingTaskMultiOrderNum int64
	SingleOrderNum                int64

	ItemNum                           int64
	MultiPickingTaskItemNum           int64
	MultiPickingTaskMultiOrderItemNum int64
	SingleOrderItem                   int64

	SkuNum                           int64
	MultiPickingTaskSkuNum           int64
	MultiPickingTaskMultiOrderSkuNum int64
	SingleOrderSkuNum                int64

	MaxOrderNumOfPickingTask int64
	AvgOrderNumOfPickingTask int64
	MinOrderNumOfPickingTask int64

	MaxItemNumOfSubPickingTask int64
	AvgItemNumOfSubPickingTask int64
	MinItemNumOfSubPickingTask int64

	MaxSkuNumOfSubPickingTask int64
	AvgSkuNumOfSubPickingTask int64
	MinSkuNumOfSubPickingTask int64

	MaxVolumeOfSubPickingTask int64
	AvgVolumeOfSubPickingTask int64
	MinVolumeOfSubPickingTask int64

	Distance                             int64
	DistancePerItem                      int64
	DistancePerSubPickingTask            int64
	TotalMultiPickingTaskWalkingDistance int64
}

type WarehouseMapHelper struct {
	Origins      []string    // 起点locationId集合
	Destinations []string    // 终点locationId集合
	DistanceInfo interface{} // 距离信息
}

func (h *WarehouseMapHelper) DistanceBetween(x, y int64) (float64, bool) {
	return 0, false
}
