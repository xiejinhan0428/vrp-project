package waveoptsolver

import (
	"fmt"
	"git.garena.com/shopee/bg-logistics/algo/wms/waveopt/localsearch/pkg/solver"
	"sort"
)

type SolverWave struct {
	Id string

	// groups
	Groups []*SolverWaveGroup
}

type SolverWaveGroup struct {
	Id string

	// the wave that this group belongs to
	Wave *SolverWave

	// group type
	GroupType GroupType

	// order group priority
	Priority int64

	// orders
	Orders []*SolverWaveOrder
}

type SolverWaveOrder struct {
	Id string

	// the group that this order belongs to
	Group *SolverWaveGroup

	// sku list
	Skus []*SolverWaveSku

	// the task this order belongs to
	PickingTask *SolverWavePickingTask

	// is this order a multi-picker order, true for yes
	isMultiPickerOrder bool

	// is this order urgent, true for yes
	isUrgent bool

	maxPickingSequence int64
}

func (o *SolverWaveOrder) Identifier() interface{} {
	return o.Id
}

func (o *SolverWaveOrder) Value() (solver.Value, error) {
	return o.PickingTask, nil
}

func (o *SolverWaveOrder) String() string {
	return fmt.Sprintf("Order{Id: %v,sku:{%v}}", o.Id, o.Skus)
}

type SolverWaveSku struct {
	Id string

	// the order that this sku belongs to
	Order *SolverWaveOrder

	// sku location info
	Location *SolverSkuLocation

	// measurable info
	Length int64
	Width  int64
	Height int64
	Weight int64

	Qty int64

	Volume      int64
	totalVolume int64

	// sku size
	SkuSize SkuSizeType

	//picking sequence
	PickingSequence int64

	// coordinate
	Coordinate []float64
}

func (sku *SolverWaveSku) SingleVolume() int64 {
	return sku.Length * sku.Width * sku.Height
}

func (sku *SolverWaveSku) TotalVolume() int64 {
	return sku.SingleVolume() * sku.Qty
}

func (sku *SolverWaveSku) SingleWeight() int64 {
	return sku.Weight
}

func (sku *SolverWaveSku) TotalWeight() int64 {
	return sku.Weight * sku.Qty
}

func (sku *SolverWaveSku) String() string {
	return fmt.Sprintf("SKU{Id: %v, Qty: %v}", sku.Id, sku.Qty)
}

type SolverSkuLocation struct {
	LocationId      string
	SegmentId       string
	PathwayId       string
	ZoneId          string
	ZoneClusterId   string
	ZoneSectorId    string
	PickingSequence int64
	CoordinateX     float64
	CoordinateY     float64
}

type SolverWavePickingTask struct {
	Id string

	// if IsReserved == true, this task only holds orders that cannot form a valid task
	IsReserved bool

	// the wave this picking task belongs to
	Group *SolverWaveGroup

	Orders []*SolverWaveOrder

	// sub picking tasks of this task
	// after each move, this field is updated
	// evaluator should directly access this field to get the sub picking tasks to avoid repeatly generating them
	SubPickingTasks []*SolverWaveSubPickingTask

	// the divider is for sub picking task splitting
	// based on the sub picking task dividing granula and picker mode, picking tasks may have different divider instances
	Divider Divider

	isMultiPickerTask bool
}

func (t *SolverWavePickingTask) String() string {
	return fmt.Sprintf("PickingTask{Id: %v}", t.Id)
}

func (t *SolverWavePickingTask) Identifier() interface{} {
	return t.Id
}

func (t *SolverWavePickingTask) Variables() ([]solver.Variable, error) {
	vars := make([]solver.Variable, 0)
	for _, order := range t.Orders {
		vars = append(vars, order)
	}

	return vars, nil
}

func (t *SolverWavePickingTask) removeOrder(order *SolverWaveOrder) error {
	if order.PickingTask.Id != t.Id {
		return fmt.Errorf("no such order ID %v in picking task %v", order.Id, t.Id)
	}

	idx := -1
	for i, odr := range t.Orders {
		if order.Id == odr.Id {
			idx = i
			break
		}
	}

	if idx < 0 {
		return fmt.Errorf("no such order ID %v in picking task %v", order.Id, t.Id)
	}

	newOrders := append(make([]*SolverWaveOrder, 0), t.Orders[0:idx]...)
	newOrders = append(newOrders, t.Orders[idx+1:]...)
	t.Orders = newOrders

	order.PickingTask = nil

	return nil
}

func (t *SolverWavePickingTask) addOrder(order *SolverWaveOrder) {
	order.PickingTask = t
	t.Orders = append(t.Orders, order)
}

type SolverWaveSubPickingTask struct {
	Id string

	PickingTask *SolverWavePickingTask

	Skus []*SolverWaveSku
}

func (st *SolverWaveSubPickingTask) TotalSkuQty() int64 {
	qty := int64(0)
	for _, sku := range st.Skus {
		qty += sku.Qty
	}

	return qty
}

func (st *SolverWaveSubPickingTask) TotalVolume() int64 {
	volume := int64(0)
	for _, sku := range st.Skus {
		volume += sku.TotalVolume()
	}

	return volume
}

// WaveOptSolution is the picking task arrangement of a wave group
type WaveOptSolution struct {
	Group *SolverWaveGroup

	Orders []*SolverWaveOrder
	Tasks  []*SolverWavePickingTask

	NormalTasks []*SolverWavePickingTask
	OrphanTasks []*SolverWavePickingTask
}

func (s *WaveOptSolution) Copy() (solver.Solution, error) {
	orderIdToIdxMap := make(map[string]int)
	taskIdToIdxMap := make(map[string]int)

	for i, order := range s.Orders {
		orderIdToIdxMap[order.Id] = i
	}
	for i, task := range s.Tasks {
		taskIdToIdxMap[task.Id] = i
	}

	newSolution := &WaveOptSolution{
		Group:  s.Group,
		Orders: make([]*SolverWaveOrder, 0),
		Tasks:  make([]*SolverWavePickingTask, 0),
	}

	orderIdCopiedSet := make(map[string]bool)

	for _, oldTask := range s.Tasks {
		newTask := &SolverWavePickingTask{
			Id:                oldTask.Id,
			IsReserved:        oldTask.IsReserved,
			Group:             oldTask.Group,
			Orders:            make([]*SolverWaveOrder, 0),
			SubPickingTasks:   make([]*SolverWaveSubPickingTask, 0),
			Divider:           oldTask.Divider,
			isMultiPickerTask: oldTask.isMultiPickerTask,
		}

		newSolution.Tasks = append(newSolution.Tasks, newTask)

		for _, oldOrder := range oldTask.Orders {
			newOrder := &SolverWaveOrder{
				Id:          oldOrder.Id,
				Group:       oldOrder.Group,
				Skus:        oldOrder.Skus,
				PickingTask: newTask,
			}
			newTask.Orders = append(newTask.Orders, newOrder)
			newSolution.Orders = append(newSolution.Orders, newOrder)
			orderIdCopiedSet[newOrder.Id] = true
		}
	}

	for _, task := range newSolution.Tasks {
		if task.IsReserved {
			newSolution.OrphanTasks = append(newSolution.OrphanTasks, task)
		} else {
			newSolution.NormalTasks = append(newSolution.NormalTasks, task)
		}
		task.SubPickingTasks = task.Divider.Divide(task)
	}

	for _, oldOrder := range s.Orders {
		if !orderIdCopiedSet[oldOrder.Id] {
			newOrder := &SolverWaveOrder{
				Id:          oldOrder.Id,
				Group:       oldOrder.Group,
				Skus:        oldOrder.Skus,
				PickingTask: nil,
			}
			newSolution.Orders = append(newSolution.Orders, newOrder)
		}
	}

	// keep the orders and tasks in the same order across multiple solutions
	sort.Slice(newSolution.Orders, func(i, j int) bool {
		left := newSolution.Orders[i].Id
		right := newSolution.Orders[j].Id

		return orderIdToIdxMap[left] < orderIdToIdxMap[right]
	})

	sort.Slice(newSolution.Tasks, func(i, j int) bool {
		left := newSolution.Tasks[i].Id
		right := newSolution.Tasks[j].Id

		return taskIdToIdxMap[left] < taskIdToIdxMap[right]
	})

	return newSolution, nil
}
