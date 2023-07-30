package waveoptsolver

import (
	"fmt"
	"git.garena.com/shopee/bg-logistics/go/gocommon/logger"
	"math"
	"math/rand"
	"sort"
	"time"
)

// 初始解指标排序
type customerSortPickingTaskEvaluation struct {
	t    []*TmpPickingTaskEvaluation
	less func(x, y *TmpPickingTaskEvaluation) bool
}

func (x customerSortPickingTaskEvaluation) Len() int {
	return len(x.t)
}

func (x customerSortPickingTaskEvaluation) Less(i, j int) bool {
	return x.less(x.t[i], x.t[j])
}

func (x customerSortPickingTaskEvaluation) Swap(i, j int) {
	x.t[i], x.t[j] = x.t[j], x.t[i]
}

func EvaluationSort(pickingTaskEvaluator []*TmpPickingTaskEvaluation) []*TmpPickingTaskEvaluation {
	sort.Sort(customerSortPickingTaskEvaluation{pickingTaskEvaluator, func(x, y *TmpPickingTaskEvaluation) bool {
		if x.Distance != y.Distance {
			return x.Distance > y.Distance
		}
		if x.IsUrgent != y.IsUrgent {
			return x.IsUrgent
		}
		if x.OrderQty != y.OrderQty {
			return x.OrderQty > y.OrderQty
		}
		if x.SkuQty != y.SkuQty {
			return x.SkuQty > y.SkuQty
		}

		return false

	}})
	return pickingTaskEvaluator
}

type TmpPickingTaskEvaluation struct {
	// 构造初始解时，临时组建的pickingtask信息
	Orders        []*SolverWaveOrder
	OrderIndex    int
	IsUrgent      bool
	OrderQty      int64
	SkuQty        int64
	Distance      int64
	Volume        int64
	SplitLevelNum int64
	config        *WaveSolverConfig
}

func (tmp *TmpPickingTaskEvaluation) TmpPickingTaskEvaluator(level SplitLevel) {
	tmpSkus := make([]*SolverWaveSku, 0)
	tmpPosition := make([]*SolverSkuLocation, 0)
	tmpSplitSet := make(map[string]bool)
	for _, order := range tmp.Orders {
		tmpSkus = append(tmpSkus, order.Skus...)
		if order.isUrgent == true {
			tmp.IsUrgent = true
		}
		tmp.OrderQty += 1
	}
	for _, sku := range tmpSkus {
		tmpSplitLevel, _ := extractSplitLevel(sku.Location, level)
		tmpSplitSet[tmpSplitLevel] = true
		tmp.SkuQty += sku.Qty
		tmp.Volume += sku.totalVolume
		tmpPosition = append(tmpPosition, sku.Location)
	}
	tmp.SplitLevelNum = int64(len(tmpSplitSet))
	tmp.Distance = CalculateDistance(tmpPosition, tmp.config)

}

func (sol *WaveOptSolutionSingle) InitSolutionRandom(group *SolverWaveGroup, orders []*SolverWaveOrder, rule *WaveRule, config *WaveSolverConfig, solvingSeconds int64, terminateChan chan bool) (*WaveOptSolutionSingle, time.Duration, error) {
	// 遍历order点，一个一个加进去，加进去的规则：先判断距离，其次判断isurgent，再判断订单个数
	unselectedOrder := make([]*SolverWaveOrder, 0)
	unselectedOrder = append(unselectedOrder, orders...)

	startTime := time.Now()
	overdueTimer := time.NewTimer(time.Duration(solvingSeconds) * time.Second)

	for len(unselectedOrder) > 0 {
		pickingtask := make([]*SolverWaveOrder, 0)
		for {
			if int64(len(pickingtask)) == 1 || len(unselectedOrder) == 0 {
				tmpPickingTask := new(PickingTaskSingle)
				tmpPickingTask.Orders = pickingtask
				sol.PickingTasks = append(sol.PickingTasks, tmpPickingTask)
				nilTask := new(PickingTaskSingle)
				sol.PickingTasks = append(sol.PickingTasks, nilTask)
				break
			}
			select {
			case <-terminateChan:
				return nil, 0, fmt.Errorf("init solution construction of group %v single picker orders is terminated", group.Id)
			case <-overdueTimer.C:
				return nil, 0, fmt.Errorf("no enough time to init multi-picker order of group %v", group.Id)
			default:
			}

			randomIndex := rand.Intn(len(unselectedOrder))
			pickingtask = append(pickingtask, unselectedOrder[randomIndex])
			unselectedOrder = append(unselectedOrder[:randomIndex], unselectedOrder[randomIndex+1:]...)
		}

	}
	for i, task := range sol.PickingTasks {
		task.id = i
	}
	endTime := time.Now()
	duration := int64(endTime.Sub(startTime).Seconds())
	var remainSeconds int64
	if duration >= solvingSeconds {
		remainSeconds = 0
	} else {
		remainSeconds = solvingSeconds - duration
	}

	return sol, time.Duration(remainSeconds) * time.Second, nil
}

//func (sol *WaveOptSolutionSingle) V1InitSolution(wave *Wave)

func (sol *WaveOptSolutionSingle) InitSolution(group *SolverWaveGroup, orders []*SolverWaveOrder, rule *WaveRule, config *WaveSolverConfig, solvingSeconds int64, terminateChan chan bool, splitLevel SplitLevel) (*WaveOptSolutionSingle, time.Duration, error) {
	// decide divider granularity
	//splitLevel, _ := rule.splitLevel()
	// 遍历order点，一个一个加进去，加进去的规则：先判断距离，其次判断isurgent，再判断订单个数
	selectedOrder := make([]*SolverWaveOrder, 0)
	unselectedOrder := make([]*SolverWaveOrder, 0)

	// 遍历订单列表生成未被选择订单切片
	unselectedOrder = append(unselectedOrder, orders...)
	//for _, order := range unselectedOrder {
	//	order.maxPickingSequence = int64(-1e10)
	//	for _, sku := range order.Skus {
	//		if sku.PickingSequence > order.maxPickingSequence {
	//			order.maxPickingSequence = sku.PickingSequence
	//		}
	//	}
	//}
	//sort.Slice(unselectedOrder, func(i, j int) bool {
	//	return unselectedOrder[i].maxPickingSequence > unselectedOrder[j].maxPickingSequence
	//})

	startTime := time.Now()
	logger.LogInfof("WaveOptAlgo - %s: initialization is running!", startTime)

	overdueTimer := time.NewTimer(time.Duration(solvingSeconds) * time.Second)

	// 随机选取一个订单作为初始解第一位
	for len(unselectedOrder) > 0 {
		pickingtask := make([]*SolverWaveOrder, 0)
		randomIndex := rand.Intn(len(unselectedOrder))
		pickingtask = append(pickingtask, unselectedOrder[randomIndex])
		unselectedOrder = append(unselectedOrder[:randomIndex], unselectedOrder[randomIndex+1:]...)
		for {
			tmpPickingTaskEvaluations := make([]*TmpPickingTaskEvaluation, 0)
			bestTmpPickingTaskEvaluation := &TmpPickingTaskEvaluation{
				Distance: math.MaxInt64,
			}
			for ind, order := range unselectedOrder {
				select {
				case <-terminateChan:
					return nil, 0, fmt.Errorf("init solution construction of group %v single picker orders is terminated", group.Id)
				case <-overdueTimer.C:
					sol.normalTasks = make([]*PickingTaskSingle, 0)
					sol.orphanOrders = make([]*SolverWaveOrder, 0)
					for _, pt := range sol.PickingTasks {
						if len(pt.Orders) <= 0 {
							continue
						}

						pt.PickingTaskEvaluate(rule, config, splitLevel)

						if pt.ConstraintViolation > 0 {
							//fmt.Println(pt)
							sol.orphanOrders = append(sol.orphanOrders, pt.Orders...)
						} else {
							sol.normalTasks = append(sol.normalTasks, pt)
						}
					}
					sol.orphanOrders = append(sol.orphanOrders, unselectedOrder...)
					return sol, 0, nil
				default:
				}

				// 选择未被选择的点加入pickingtask
				tmpPickingTaskEvaluation := &TmpPickingTaskEvaluation{
					Orders:     append(pickingtask, order),
					OrderIndex: ind,
					config:     config,
				}
				// 对每个加入的pickingtask进行指标评估
				tmpPickingTaskEvaluation.TmpPickingTaskEvaluator(splitLevel)
				// 没有超过的约束进入评估列表
				if tmpPickingTaskEvaluation.OrderQty <= rule.CommonRule.MaxOrderQtyPerPickingTask &&
					//tmpPickingTaskEvaluation.OrderQty >= rule.CommonRule.MinOrderQtyPerPickingTask+1 &&
					tmpPickingTaskEvaluation.SkuQty <= rule.CommonRule.MaxItemQtyPerSubPickingTask &&
					//tmpPickingTaskEvaluation.SkuQty >= rule.CommonRule.MinItemQtyPerSubPickingTask &&
					tmpPickingTaskEvaluation.Volume <= rule.CommonRule.MaxItemVolumePerSubPickingTask {
					//tmpPickingTaskEvaluation.Volume >= rule.CommonRule.MinItemVolumePerSubPickingTask
					if splitLevel == "" || tmpPickingTaskEvaluation.SplitLevelNum < 2 {
						if tmpPickingTaskEvaluation.Distance < bestTmpPickingTaskEvaluation.Distance {
							bestTmpPickingTaskEvaluation = tmpPickingTaskEvaluation
						}
						tmpPickingTaskEvaluations = append(tmpPickingTaskEvaluations, tmpPickingTaskEvaluation)
					}
				}
			}
			if len(tmpPickingTaskEvaluations) == 0 {
				tmpPickingTask := new(PickingTaskSingle)
				tmpPickingTask.Orders = pickingtask
				sol.PickingTasks = append(sol.PickingTasks, tmpPickingTask)
				nilTask := new(PickingTaskSingle)
				sol.PickingTasks = append(sol.PickingTasks, nilTask)
				break
			}
			// 对评估列表进行排序，找出插入使得初识解最好的订单
			//tmpPickingTaskEvaluations = EvaluationSort(tmpPickingTaskEvaluations)
			orderIndex := bestTmpPickingTaskEvaluation.OrderIndex
			//fmt.Println(orderIndex, selectedOrder, unselectedOrder)
			pickingtask = append(pickingtask, unselectedOrder[orderIndex])
			selectedOrder = append(selectedOrder, unselectedOrder[orderIndex])
			// 从未被选择的订单列表剔除该点
			unselectedOrder = append(unselectedOrder[:orderIndex], unselectedOrder[orderIndex+1:]...)
		}

	}

	for i, task := range sol.PickingTasks {
		task.id = i
	}

	endTime := time.Now()
	duration := int64(endTime.Sub(startTime).Seconds())
	var remainSeconds int64
	if duration >= solvingSeconds {
		remainSeconds = 0
	} else {
		remainSeconds = solvingSeconds - duration
	}
	logger.LogInfof("WaveOptAlgo - %s: initialization is finished!", endTime)

	//sol.GroupSolutionEvaluator()
	return sol, time.Duration(remainSeconds) * time.Second, nil
}
