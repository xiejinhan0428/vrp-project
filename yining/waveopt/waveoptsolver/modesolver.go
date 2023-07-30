package waveoptsolver

import (
	"fmt"
	"git.garena.com/shopee/bg-logistics/algo/wms/waveopt/localsearch/pkg/score"
	"git.garena.com/shopee/bg-logistics/algo/wms/waveopt/localsearch/pkg/solver"
	"git.garena.com/shopee/bg-logistics/algo/wms/waveopt/localsearch/pkg/strategy"
	"git.garena.com/shopee/bg-logistics/go/gocommon/logger"
	"math"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type groupResult struct {
	groupId       string
	groupPriority int64
	groupType     GroupType
	tasks         []*SolverWavePickingTask
	orphanOrders  []*SolverWaveOrder
}

func solveMultiPickerGroup(group *SolverWaveGroup, rule *WaveRule, solverConfig *WaveSolverConfig, solvingSeconds int64, terminateChan chan bool) (*groupResult, error) {
	logger.LogInfof("WaveOptAlgo - %s: solving multi-picker group %s", group.Wave.Id, group.Id)

	logger.LogInfof("WaveOptAlgo - %s: solving multi-picker orders in group %s for %d seconds", group.Wave.Id, group.Id, solvingSeconds)
	multiPickerOrdersGroupResult, _, err := solveMultiPickerOrders(group, rule, solverConfig, solvingSeconds, terminateChan)
	if err != nil {
		logger.LogErrorf("WaveOptAlgo - %s: error in solving group %s: %s", group.Wave.Id, group.Id, err.Error())
		return nil, err
	}

	grpResult := &groupResult{
		groupId:       group.Id,
		groupPriority: group.Priority,
		groupType:     group.GroupType,
		tasks:         make([]*SolverWavePickingTask, 0),
		orphanOrders:  make([]*SolverWaveOrder, 0),
	}

	// merge tasks and orphan orders
	if multiPickerOrdersGroupResult != nil {
		grpResult.tasks = append(grpResult.tasks, multiPickerOrdersGroupResult.tasks...)
		grpResult.orphanOrders = append(grpResult.orphanOrders, multiPickerOrdersGroupResult.orphanOrders...)
	} else {
		grpResult.orphanOrders = append(grpResult.orphanOrders, group.Orders...)
	}

	logger.LogInfof("WaveOptAlgo - %s: multi-picker group %s solved, %d picking tasks and %d orphan orders are found.", group.Wave.Id, group.Id, len(grpResult.tasks), len(grpResult.orphanOrders))
	return grpResult, nil
}

func solveMultiPickerOrders(group *SolverWaveGroup, rule *WaveRule, solverConfig *WaveSolverConfig, solvingSeconds int64, terminateChan chan bool) (*groupResult, time.Duration, error) {
	var err error

	startTime := time.Now()

	// filter out the multi picker orders and remove orphan orders
	isCrossCluster := rule.CommonRule.IsCrossZoneCluster
	multiPickerOrders, orphanOrders := separateNormalAndOrphanOrders(rule, group.Orders, isCrossCluster, rule.CommonRule.MaxItemVolumePerSubPickingTask)

	// skip solving if no multi-picker orders
	if len(multiPickerOrders) < int(rule.CommonRule.MinOrderQtyPerPickingTask) {
		result := &groupResult{
			groupId:       group.Id,
			groupPriority: group.Priority,
			groupType:     group.GroupType,
			tasks:         make([]*SolverWavePickingTask, 0),
			orphanOrders:  orphanOrders,
		}
		return result, getRemainSeconds(startTime, time.Now(), solvingSeconds), nil
	}

	// calculate task #
	// valid tasks # = order # / min order # per task
	// reserved task # = order # / max order # per task
	orderNum := len(multiPickerOrders)
	normalTaskNum := int(math.Ceil(float64(orderNum) / float64(rule.CommonRule.MinOrderQtyPerPickingTask)))
	orphanTaskNum := int(math.Ceil(float64(orderNum) / float64(rule.CommonRule.MaxOrderQtyPerPickingTask)))

	normalTasks, orphanTasks, err := getEmptyTaskForSolvingMultiPickerOrders(group, normalTaskNum, orphanTaskNum, rule)
	if err != nil {
		return nil, getRemainSeconds(startTime, time.Now(), solvingSeconds), err
	}

	//initSolution, initScore, remainTime, err := constructInitialSolutionForMultiPickerOrders(
	//initSolution, initScore, remainTime, err := randomInitialSolutionForMultiPickerOrders(
	initSolution, remainTime, err := bestFit(
		group, rule, solverConfig, multiPickerOrders, normalTasks, orphanTasks, solvingSeconds, terminateChan)
	if err != nil {
		return nil, getRemainSeconds(startTime, time.Now(), solvingSeconds), err
	}

	optSolvingSeconds := int64(remainTime.Seconds())
	bestSolution, bestScore, err := optMultiPickerOrders(group, rule, solverConfig, initSolution, optSolvingSeconds, terminateChan)
	if err != nil {
		return nil, getRemainSeconds(startTime, time.Now(), solvingSeconds), err
	}

	if isFeasible, _ := bestScore.IsFeasible(); !isFeasible {
		resultNormalTasks := make([]*SolverWavePickingTask, 0)
		resultOrphanTasks := make([]*SolverWavePickingTask, 0)
		resultOrphanTasks = append(resultOrphanTasks, bestSolution.OrphanTasks...)
		for _, task := range bestSolution.NormalTasks {
			if len(task.Orders) <= 0 {
				continue
			}

			isCrossZoneClusterPenalty, _ := penalizeCrossCluster(task, rule.CommonRule.IsCrossZoneCluster)
			if isCrossZoneClusterPenalty > 0.0 {
				resultOrphanTasks = append(resultOrphanTasks, task)
				continue
			}

			orderNumPenalty, _ := penalizeOrderQtyPerTask(task, rule.CommonRule.MaxOrderQtyPerPickingTask, rule.CommonRule.MinOrderQtyPerPickingTask)
			if orderNumPenalty > 0.0 {
				resultOrphanTasks = append(resultOrphanTasks, task)
				continue
			}

			itemQtyPenalty, _ := penalizeItemQtyPerSubPickingTask(task, rule.CommonRule.MaxItemQtyPerSubPickingTask, rule.CommonRule.MinItemQtyPerSubPickingTask)
			if itemQtyPenalty > 0.0 {
				resultOrphanTasks = append(resultOrphanTasks, task)
				continue
			}

			itemVolumePenalty, _ := penalizeItemVolumePerSubPickingTask(task, rule.CommonRule.MaxItemVolumePerSubPickingTask, rule.CommonRule.MinItemVolumePerSubPickingTask)
			if itemVolumePenalty > 0.0 {
				resultOrphanTasks = append(resultOrphanTasks, task)
				continue
			}

			resultNormalTasks = append(resultNormalTasks, task)
		}

		bestSolution.NormalTasks = resultNormalTasks
		bestSolution.OrphanTasks = resultOrphanTasks
	}

	if len(bestSolution.NormalTasks) <= 0 {
		return nil, getRemainSeconds(startTime, time.Now(), solvingSeconds), fmt.Errorf("cannot find a feasible solution of wave %v group %v in multi-picker mode", group.Wave.Id, group.Id)
	}

	// move orders in reserved tasks to orphanOrders
	for _, orphanTask := range bestSolution.OrphanTasks {
		orphanOrders = append(orphanOrders, orphanTask.Orders...)
	}

	// mark normal tasks as multi-picker tasks and exclude empty tasks
	nonEmptyNormalTasks := make([]*SolverWavePickingTask, 0)
	for _, normalTask := range bestSolution.NormalTasks {
		if len(normalTask.Orders) <= 0 {
			continue
		}

		if len(normalTask.SubPickingTasks) <= 1 {
			normalTask.isMultiPickerTask = false
		} else {
			normalTask.isMultiPickerTask = true
		}

		nonEmptyNormalTasks = append(nonEmptyNormalTasks, normalTask)
	}

	// new a group result
	grpResult := &groupResult{
		groupId:       group.Id,
		groupPriority: group.Priority,
		groupType:     group.GroupType,
		tasks:         nonEmptyNormalTasks,
		orphanOrders:  orphanOrders,
	}

	return grpResult, getRemainSeconds(startTime, time.Now(), solvingSeconds), nil
}

func getRemainSeconds(startTime, currentTime time.Time, limit int64) time.Duration {
	d := int64(currentTime.Sub(startTime).Seconds())
	if d < 0 {
		return 0
	}

	if d < limit {
		return time.Duration(limit-d) * time.Second
	} else {
		return 0
	}
}

func separateNormalAndOrphanOrders(rule *WaveRule, orders []*SolverWaveOrder, isCrossCluster bool, maxItemVolumePerSubPickingTask int64) ([]*SolverWaveOrder, []*SolverWaveOrder) {
	multiPickerOrders := make([]*SolverWaveOrder, 0)
	orphanOrders := make([]*SolverWaveOrder, 0)
	isOrphanOrder := false
	isZeroMulti := rule.isZeroMulti()
	for _, order := range orders {
		// an order is an orphan if:
		// 1. any SKU in it has volume > rule.CommonRule.MaxItemVolumePerSubPickingTask
		// 2. locations of SKUs violate the isCrossZoneCluster rule
		isOrphanOrder = false
		clusterSet := make(map[string]bool)
		for _, sku := range order.Skus {
			if sku.TotalVolume() > maxItemVolumePerSubPickingTask {
				isOrphanOrder = true
				break
			}
			clusterSet[sku.Location.ZoneClusterId] = true
			if len(clusterSet) > 1 && !isCrossCluster {
				isOrphanOrder = true
				break
			}
		}

		if isZeroMulti && order.isMultiPickerOrder {
			isOrphanOrder = true
		}

		if isOrphanOrder {
			orphanOrders = append(orphanOrders, order)
			continue
		}

		multiPickerOrders = append(multiPickerOrders, order)
	}

	return multiPickerOrders, orphanOrders
}

func getEmptyTaskForSolvingMultiPickerOrders(group *SolverWaveGroup, normalTaskNum, orphanTaskNum int, rule *WaveRule) ([]*SolverWavePickingTask, []*SolverWavePickingTask, error) {
	// decide divider granularity
	splitLevel, _ := rule.splitLevel()
	var granularity DivideGranularity
	granularity, _ = fromSplitLevelToDivideGranularity(splitLevel)

	granularDivider, err := NewLocationBasedDivider(granularity, rule.CommonRule.MaxItemQtyPerSubPickingTask, rule.CommonRule.MinItemQtyPerSubPickingTask, rule.CommonRule.MaxItemVolumePerSubPickingTask, rule.CommonRule.MinItemVolumePerSubPickingTask)
	if err != nil {
		return nil, nil, err
	}
	noneDivider, err := NewDivideByNoneDivider()
	if err != nil {
		return nil, nil, err
	}

	// generate tasks
	normalTasks := make([]*SolverWavePickingTask, 0)
	for i := 0; i < normalTaskNum; i++ {
		taskId := fmt.Sprintf("%v_NormalPickingTask_MultiPicker_%v", group.Id, strconv.Itoa(i))
		task := &SolverWavePickingTask{
			Id:              taskId,
			IsReserved:      false,
			Group:           group,
			Orders:          make([]*SolverWaveOrder, 0),
			SubPickingTasks: make([]*SolverWaveSubPickingTask, 0),
			Divider:         granularDivider,
		}
		normalTasks = append(normalTasks, task)
	}

	orphanTasks := make([]*SolverWavePickingTask, 0)
	for i := 0; i < orphanTaskNum; i++ {
		taskId := fmt.Sprintf("%v_ReservedPickingTask_MultiPicker_%v", group.Id, strconv.Itoa(i))
		task := &SolverWavePickingTask{
			Id:              taskId,
			IsReserved:      true,
			Group:           group,
			Orders:          make([]*SolverWaveOrder, 0),
			SubPickingTasks: make([]*SolverWaveSubPickingTask, 0),
			Divider:         noneDivider,
		}
		orphanTasks = append(orphanTasks, task)
	}

	return normalTasks, orphanTasks, nil
}

func bestFit(group *SolverWaveGroup, rule *WaveRule, solverConfig *WaveSolverConfig, orders []*SolverWaveOrder, normalTasks, orphanTasks []*SolverWavePickingTask, solvingSeconds int64, terminateChan chan bool) (*WaveOptSolution, time.Duration, error) {
	startTime := time.Now()

	tasks := make([]*SolverWavePickingTask, 0)
	tasks = append(tasks, normalTasks...)
	tasks = append(tasks, orphanTasks...)

	// new an init solution
	initSolution := &WaveOptSolution{
		Group:       group,
		Orders:      orders,
		Tasks:       tasks,
		NormalTasks: normalTasks,
		OrphanTasks: orphanTasks,
	}

	locationKeyMap := make(map[string]string)
	for _, order := range initSolution.Orders {
		keys := make([]string, 0)
		for _, sku := range order.Skus {
			key := strings.Join([]string{sku.Location.ZoneSectorId, sku.Location.ZoneClusterId, sku.Location.ZoneId, sku.Location.PathwayId, sku.Location.SegmentId}, "_")
			keys = append(keys, key)
		}
		sort.Slice(keys, func(i, j int) bool {
			return keys[i] < keys[j]
		})
		locationKeyMap[order.Id] = keys[0]
	}

	sort.Slice(initSolution.Orders, func(i, j int) bool {
		left := locationKeyMap[initSolution.Orders[i].Id]
		right := locationKeyMap[initSolution.Orders[j].Id]
		return left < right
	})

	distanceCalculator, err := NewLocationBasedDistanceCalculator(solverConfig.ZoneCoeff, solverConfig.PathwayCoeff, solverConfig.SegmentCoeff)
	if err != nil {
		return nil, 0, err
	}
	oldTaskScoreMap := make(map[string]score.FloatScore)
	for _, task := range initSolution.Tasks {
		scr, _ := score.FloatScoreFrom([]float64{0, 0, 0, 0}, []float64{0}, solver.DownScore)
		oldTaskScoreMap[task.Id] = scr
	}

	overtimeTimer := time.NewTimer(time.Duration(solvingSeconds) * time.Second)
BestFitInit:
	for i, order := range initSolution.Orders {
		bestImprovementScore := []float64{1e10, 1e10, 1e10, 1e10, 1e10}
		var bestScore score.FloatScore
		var bestTask *SolverWavePickingTask

		isEmptyNormalTaskChecked := false
		isEmptyReservedTaskChecked := false
		for _, task := range initSolution.Tasks {

			select {
			case <-terminateChan:
				return nil, 0, fmt.Errorf("init solution construction of multi-picker group %v is async terminated", group.Id)
			case <-overtimeTimer.C:
				orphanTask := initSolution.OrphanTasks[0]
				orphanTask.Orders = append(orphanTask.Orders, initSolution.Orders[i:]...)
				orphanTask.SubPickingTasks = orphanTask.Divider.Divide(orphanTask)
				break BestFitInit
			default:
			}

			if isEmptyNormalTaskChecked && !task.IsReserved && len(task.Orders) <= 0 {
				continue
			}

			if isEmptyReservedTaskChecked && task.IsReserved && len(task.Orders) <= 0 {
				continue
			}

			if !task.IsReserved && len(task.Orders) <= 0 {
				isEmptyNormalTaskChecked = true
			}

			if task.IsReserved && len(task.Orders) <= 0 {
				isEmptyReservedTaskChecked = true
			}

			currentScore, err := score.NewFloatScore(4, 1, solver.DownScore)
			if err != nil {
				return nil, 0, err
			}

			oldTaskScore := oldTaskScoreMap[task.Id]

			task.addOrder(order)
			task.SubPickingTasks = task.Divider.Divide(task)

			crossClusterPenalty, err := penalizeCrossCluster(task, rule.CommonRule.IsCrossZoneCluster)
			if err != nil {
				return nil, 0, err
			}
			currentScore.ConstraintScores[0] = crossClusterPenalty

			taskOrderQtyPenalty, err := penalizeOrderQtyPerTask(task, rule.CommonRule.MaxOrderQtyPerPickingTask, 0)
			if err != nil {
				return nil, 0, err
			}
			currentScore.ConstraintScores[1] = taskOrderQtyPenalty

			itemQtyPenalty, err := penalizeItemQtyPerSubPickingTask(task, rule.CommonRule.MaxItemQtyPerSubPickingTask, 0)
			if err != nil {
				return nil, 0, err
			}
			if task.IsReserved {
				currentScore.ConstraintScores[2] = 0
			} else {
				currentScore.ConstraintScores[2] = itemQtyPenalty
			}

			itemVolumePenalty, err := penalizeItemVolumePerSubPickingTask(task, rule.CommonRule.MaxItemVolumePerSubPickingTask, 0)
			if err != nil {
				return nil, 0, err
			}
			if task.IsReserved {
				currentScore.ConstraintScores[3] = 0
			} else {
				currentScore.ConstraintScores[3] = itemVolumePenalty
			}

			distance := 0.0
			for _, subTask := range task.SubPickingTasks {
				distance += distanceCalculator.CalculateDistance(subTask)
			}

			if task.IsReserved {
				currentScore.ObjectiveScores[0] = 5.0 * distance
			} else {
				currentScore.ObjectiveScores[0] = distance
			}

			scoreDiff, err := currentScore.Sub(&oldTaskScore)
			if err != nil {
				return nil, 0, err
			}

			for i := 0; i < len(scoreDiff); i++ {
				if scoreDiff[i] < bestImprovementScore[i] {
					// a better score is found
					bestTask = task
					bestImprovementScore = scoreDiff
					bestScore = currentScore
					break
				}
			}

			err = task.removeOrder(order)
			task.SubPickingTasks = task.Divider.Divide(task)
			if err != nil {
				return nil, 0, err
			}
		}
		if bestTask == nil {
			return nil, 0, fmt.Errorf("cannot init order %v, no task fits", order.Id)
		}
		bestTask.addOrder(order)
		subTasks := bestTask.Divider.Divide(bestTask)
		bestTask.SubPickingTasks = subTasks
		oldTaskScoreMap[bestTask.Id] = bestScore
	}

	for _, task := range initSolution.Tasks {
		subTasks := task.Divider.Divide(task)
		task.SubPickingTasks = subTasks
	}

	endTime := time.Now()

	duration := int64(endTime.Sub(startTime).Seconds())

	var remainSeconds int64
	if duration >= solvingSeconds {
		remainSeconds = 0
	} else {
		remainSeconds = solvingSeconds - duration
	}

	return initSolution, time.Duration(remainSeconds) * time.Second, nil
}

func optMultiPickerOrders(group *SolverWaveGroup, rule *WaveRule, solverConfig *WaveSolverConfig, initSolution *WaveOptSolution, solvingSeconds int64, terminateChan chan bool) (*WaveOptSolution, solver.Score, error) {
	// optimization solver
	multiPickerOrderMoveFactoryMap := make(map[solver.MoveFactory]float64)
	randomOrderChangeMoveFactory := NewRandomOrderChangeMoveFactory()
	randomOrderSwapMoveFactory := NewRandomOrderSwapMoveFactory()
	multiPickerOrderMoveFactoryMap[randomOrderChangeMoveFactory] = RandomOrderChangeMoveFactoryWeight
	multiPickerOrderMoveFactoryMap[randomOrderSwapMoveFactory] = RandomOrderSwapMoveFactoryWeight
	runningTimeTerminator, err := strategy.NewTimeLimitTerminatorBuilder().WithTimeLimit(solvingSeconds, strategy.Second).Build()
	if err != nil {
		return nil, nil, err
	}
	convergenceTimeTerminator, err := strategy.NewUnimprovedTimeLimitTerminatorBuilder().WithUnimprovedTimeLimit(int64(float64(solvingSeconds)*ConvergenceTimeRatio), strategy.Second).Build()
	if err != nil {
		return nil, nil, err
	}
	convergenceStepTerminator, err := strategy.NewUnimprovedStepCountLimitTerminatorBuilder().WithStepCountLimit(300).Build()
	if err != nil {
		return nil, nil, err
	}
	optAcceptor, err := strategy.NewTabuSearchAcceptorBuilder().WithVariableTabuSize(int(solverConfig.VariableTabuTenure)).WithValueTabuSize(int(solverConfig.ValueTabuTenure)).Build()
	if err != nil {
		return nil, nil, err
	}

	optSelector, _ := strategy.NewEpsilonGreedSelector(0.2)

	evaluatorFactory := &EvaluatorFactory{}
	var evaluator solver.Evaluator
	switch rule.WavePickerMode {
	case MultiPickerAtMWSWithTotalOrderQty:
		fallthrough
	case MultiPickerAtMWSWithRespectiveOrderQty:
		evaluator, err = evaluatorFactory.CreateMultiPickerWithMergingWhileSortingStationEvaluator(
			rule.CommonRule.IsCrossZoneCluster,
			rule.CommonRule.MinOrderQtyPerPickingTask,
			rule.CommonRule.MaxOrderQtyPerPickingTask,
			rule.CommonRule.MinItemQtyPerSubPickingTask,
			rule.CommonRule.MaxItemQtyPerSubPickingTask,
			rule.CommonRule.MinItemVolumePerSubPickingTask,
			rule.CommonRule.MaxItemVolumePerSubPickingTask,
			solverConfig.ZoneCoeff,
			solverConfig.PathwayCoeff,
			solverConfig.SegmentCoeff,
		)
		if err != nil {
			return nil, nil, err
		}
	case MultiPickerAtMLWithTotalPickingTaskQty:
		fallthrough
	case MultiPickerAtMLWithRespectivePickingTaskQty:
		evaluator, err = evaluatorFactory.CreateMultiPickerWithMergingLaneEvaluator(
			rule.CommonRule.IsCrossZoneCluster,
			rule.CommonRule.MinOrderQtyPerPickingTask,
			rule.CommonRule.MaxOrderQtyPerPickingTask,
			rule.CommonRule.MinItemQtyPerSubPickingTask,
			rule.CommonRule.MaxItemQtyPerSubPickingTask,
			rule.CommonRule.MinItemVolumePerSubPickingTask,
			rule.CommonRule.MaxItemVolumePerSubPickingTask,
			solverConfig.ZoneCoeff,
			solverConfig.PathwayCoeff,
			solverConfig.SegmentCoeff,
		)
		if err != nil {
			return nil, nil, err
		}
	default:
		return nil, nil, fmt.Errorf("fail to initialize the evaluator when solving multi-picker orders of group %v, unknown wave picker mode", group.Id)
	}

	multiPickerOrderOptSolver, err := solver.NewSerialSolverBuilder().
		WithName(fmt.Sprintf("Opt Solver for Multi-Picker Orders of Group %v In Wave %v", group.Id, group.Wave.Id)).
		WithSolution(initSolution).
		WithEvaluator(evaluator).
		WithRouletteMoveFactories(multiPickerOrderMoveFactoryMap).
		WithTerminator(runningTimeTerminator).
		WithTerminator(convergenceTimeTerminator).
		WithTerminator(convergenceStepTerminator).
		WithAcceptor(optAcceptor).
		WithSelector(optSelector).
		WithParams(&solver.Params{MovesPerStep: int(math.Min(float64(len(initSolution.Orders)), 500))}).
		WithStableRandSeed(int64(0)).
		IsLoggingToConsole(isLoggingToConsole).
		Build()

	wg := &sync.WaitGroup{}
	wg.Add(1)
	var bestSolution solver.Solution
	var bestScore solver.Score
	go func() {
		defer wg.Done()
		bestSolution, bestScore, err = multiPickerOrderOptSolver.Solve()
	}()

	isAsyncTerminated := false
	quit := make(chan bool, 1)
	go func() {
		select {
		case <-terminateChan:
			multiPickerOrderOptSolver.AsyncTerminate()
			isAsyncTerminated = true
		case <-quit:
			return
		}
	}()

	wg.Wait()
	if isAsyncTerminated {
		return nil, nil, fmt.Errorf("solver is terminated asynchronously when solving multi-picker orders of group %v", group.Id)
	} else {
		quit <- true
	}
	if err != nil {
		return nil, nil, err
	}

	bestSol := bestSolution.(*WaveOptSolution)
	return bestSol, bestScore, nil
}

func solveSinglePickerOrders(group *SolverWaveGroup, rule *WaveRule, solverConfig *WaveSolverConfig, solvingSeconds int64, terminateChan chan bool) (*groupResult, error) {

	//decide divider granularity
	splitLevel, _ := rule.splitLevel()
	if !rule.CommonRule.IsCrossZoneCluster {
		if splitLevel == ByZoneSector {
			splitLevel = ByZoneCluster
		}
	}

	//var granularity DivideGranularity
	//granularity, _ = fromSplitLevelToDivideGranularity(splitLevel)

	var err error

	emptyGroupResult := &groupResult{
		groupId:       group.Id,
		groupPriority: group.Priority,
		groupType:     group.GroupType,
		tasks:         make([]*SolverWavePickingTask, 0),
		orphanOrders:  make([]*SolverWaveOrder, 0),
	}

	// filter the orders which violate constraints before solve
	normalOrders, violatedOrders := singlePickerProcessing(group, rule)
	if len(normalOrders) <= 0 {
		return emptyGroupResult, nil
	}

	//startTime := time.Now()

	// create an initial solution for solver, including random init & greedy init.
	initialSolution_ := new(WaveOptSolutionSingle)
	initialSolution, remainSeconds, err := initialSolution_.InitSolution(group, normalOrders, rule, solverConfig, solvingSeconds, terminateChan, splitLevel)
	//endTime := time.Now()
	if err != nil {
		return nil, err
	}

	dropTaskIdSet := make(map[int]bool)
	//orderNums := 0
	for _, task := range initialSolution.PickingTasks {
		task.PickingTaskEvaluate(rule, solverConfig, splitLevel)
		//orderNums += len(task.Orders)
		if task.ConstraintViolation > math.MaxInt32 {
			dropTaskIdSet[task.id] = true
		}
	}

	//keepTasks := make([]*PickingTaskSingle, 0)
	//dropOrders := make([]*SolverWaveOrder, 0)
	//keepOrderNum := 0
	//for _, task := range initialSolution.PickingTasks {
	//	if hit, ok := dropTaskIdSet[task.id]; hit && ok {
	//		dropOrders = append(dropOrders, task.Orders...)
	//	} else {
	//		keepTasks = append(keepTasks, task)
	//		keepOrderNum += len(task.Orders)
	//	}
	//}
	//
	//if keepOrderNum <= 0 {
	//	return emptyGroupResult, nil
	//}
	//
	//initialSolution.PickingTasks = keepTasks

	// in one step, solver do 500 moves
	params := &solver.Params{
		MovesPerStep: 500,
	}

	// annealing acceptor params
	initialTemperature := []float64{3, 3, 1000000}

	acceptor1, err := strategy.NewSimulatedAnnealingAcceptorBuilder().WithInitialTemperatures(initialTemperature).Build()
	if err != nil {
		return nil, err
	}
	acceptor2, _ := strategy.NewTabuSearchAcceptorBuilder().WithValueTabuSize(2).WithVariableTabuSize(7).Build()
	selector := strategy.NewGreedySelector()
	evaluator := &WavePickingSingleEvaluator{
		rule:   rule,
		config: solverConfig,
	}
	_, err = evaluator.Evaluate(initialSolution)
	if err != nil {
		return nil, err
	}

	terminator1, _ := strategy.NewStepCountLimitTerminatorBuilder().WithStepCountLimit(10000).Build()
	terminator2, _ := strategy.NewTimeLimitTerminatorBuilder().WithTimeLimit(int64(remainSeconds.Seconds()), strategy.Second).Build()
	terminator3, _ := strategy.NewUnimprovedStepCountLimitTerminatorBuilder().WithStepCountLimit(300).Build()
	terminator4, _ := strategy.NewUnimprovedTimeLimitTerminatorBuilder().WithUnimprovedTimeLimit(20, strategy.Second).Build()

	//moveFactory := &WavePickingSingleChangeMoveFactory{}
	//moveFactoryToWeightMap := make(map[solver.MoveFactory]float64)
	//moveFactoryToWeightMap[moveFactory] = 1.0

	moveFactories := make([]solver.MoveFactory, 0)
	for i := 0; i <= 10; i++ {
		if i <= 0 {
			tmp1 := &WavePickingSingleDnrChangeMoveFactory{
				rule:   rule,
				config: solverConfig,
			}
			moveFactories = append(moveFactories, tmp1)
		} else {
			tmp2 := &WavePickingSingleChangeMoveFactory{
				rule:   rule,
				config: solverConfig,
			}
			moveFactories = append(moveFactories, tmp2)
		}
	}
	bpSolver, err := solver.NewSerialSolverBuilder().
		WithName(fmt.Sprintf("Wave Opt Solver of Group %v in Wave %v", group.Id, group.Wave.Id)).
		WithSolution(initialSolution).
		WithParams(params).
		WithAcceptor(acceptor1).
		WithAcceptor(acceptor2).
		WithSelector(selector).
		WithTerminator(terminator1).
		WithTerminator(terminator2).
		WithTerminator(terminator3).
		WithTerminator(terminator4).
		WithEvaluator(evaluator).
		//WithRouletteMoveFactories(moveFactoryToWeightMap).
		WithSequentialMoveFactories(moveFactories).
		IsLoggingToConsole(true).
		Build()
	if err != nil {
		return nil, err
	}

	var solution_ solver.Solution
	wg := &sync.WaitGroup{}
	wg.Add(1)
	quit := make(chan bool, 1)
	go func() {
		defer wg.Done()
		solution_, _, err = bpSolver.Solve()
	}()

	isAsyncTerminated := false
	go func() {
		select {
		case <-terminateChan:
			bpSolver.AsyncTerminate()
			isAsyncTerminated = true
		case <-quit:
			return
		}
	}()

	wg.Wait()

	if isAsyncTerminated {
		return nil, fmt.Errorf("solver is terminated asynchronously when solving single-picker orders of group %v", group.Id)
	} else {
		quit <- true
	}

	if err != nil {
		return nil, err
	}

	// add the violated orders into orphan orders list
	solution, _ := solution_.(*WaveOptSolutionSingle)
	solution.orphanOrders = append(solution.orphanOrders, violatedOrders...)
	//solution.orphanOrders = append(solution.orphanOrders, dropOrders...)

	// convert to waveOptSolutionTasks
	tmpTasks := make([]*SolverWavePickingTask, 0)
	noneDivider, err := NewDivideByNoneDivider()
	if err != nil {
		return nil, err
	}
	for idx, task := range solution.normalTasks {
		if len(task.Orders) > 0 {
			tmpTask := &SolverWavePickingTask{
				Id:                fmt.Sprintf("%v_NormalPickingTask_SinglePicker_%v", group.Id, strconv.Itoa(idx)),
				IsReserved:        false,
				Group:             group,
				Orders:            task.Orders,
				SubPickingTasks:   nil,
				Divider:           nil,
				isMultiPickerTask: false}
			tmpSubTasks := noneDivider.Divide(tmpTask)
			tmpTask.SubPickingTasks = tmpSubTasks

			tmpTasks = append(tmpTasks, tmpTask)
		}
	}
	// new a group result
	grpResult := &groupResult{
		groupId:       group.Id,
		groupPriority: group.Priority,
		groupType:     group.GroupType,
		tasks:         tmpTasks,
		orphanOrders:  solution.orphanOrders,
	}
	return grpResult, nil
}

// in single picker mode, filter the orders that exceed MaxItemQty or MaxVolume
func singlePickerProcessing(group *SolverWaveGroup, rule *WaveRule) ([]*SolverWaveOrder, []*SolverWaveOrder) {
	violatedOrders := make([]*SolverWaveOrder, 0)
	normalOrders := make([]*SolverWaveOrder, 0)
	for _, order := range group.Orders {
		if order.isMultiPickerOrder {
			violatedOrders = append(violatedOrders, order)
			continue
		}
		itemQty := int64(0)
		clusterSet := make(map[string]bool)
		orderTotalVolume := int64(0)
		for _, sku := range order.Skus {
			itemQty += sku.Qty
			clusterSet[sku.Location.ZoneClusterId] = true
			orderTotalVolume += sku.TotalVolume()
		}
		if itemQty > rule.CommonRule.MaxItemQtyPerSubPickingTask {
			violatedOrders = append(violatedOrders, order)
		} else if !rule.CommonRule.IsCrossZoneCluster && len(clusterSet) > 1 {
			violatedOrders = append(violatedOrders, order)
		} else if orderTotalVolume > rule.CommonRule.MaxItemVolumePerSubPickingTask {
			violatedOrders = append(violatedOrders, order)
		} else {
			normalOrders = append(normalOrders, order)
		}
	}
	return normalOrders, violatedOrders
}

func solveSinglePickerGroup(group *SolverWaveGroup, rule *WaveRule, solverConfig *WaveSolverConfig, solvingSeconds int64, terminateChan chan bool) (*groupResult, error) {
	return solveSinglePickerOrders(group, rule, solverConfig, solvingSeconds, terminateChan)
}

func fromSplitLevelToDivideGranularity(splitLevel SplitLevel) (DivideGranularity, error) {
	switch splitLevel {
	case ByZone:
		return DivideByZone, nil
	case ByZoneCluster:
		return DivideByZoneCluster, nil
	case ByZoneSector:
		return DivideByZoneSector, nil
	default:
		return DivideByNone, fmt.Errorf("unknown type of SplitLevel: %v", splitLevel)
	}
}
