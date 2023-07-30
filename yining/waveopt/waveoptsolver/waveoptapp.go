package waveoptsolver

import (
	"fmt"
	"git.garena.com/shopee/bg-logistics/go/gocommon/logger"
	"math"
	"sort"
	"sync"
)

type WavePickingTaskOptSolver struct {
	Wave       *Wave
	dispatcher *groupDispatcher
	terminate  chan bool
}

func (s *WavePickingTaskOptSolver) GeneratePickingTasks() *WaveResult {
	var err error
	var errMsg string

	s.terminate = make(chan bool, 1)

	// detect conflicts within WaveRule
	logger.LogInfof("WaveOptAlgo - %s: detecting conflict within rules", s.Wave.WaveSn)
	err = detectConflictsWithinRules(s.Wave.WaveRule)
	if err != nil {
		logger.LogErrorf("WaveOptAlgo - %s: conflict detected within rules: %s", s.Wave.WaveSn, err.Error())
		return errorWaveResult(s.Wave.WaveSn, UnreasonableConstraintsResult, err.Error())
	}
	logger.LogInfof("WaveOptAlgo - %s: no conflict within rules", s.Wave.WaveSn)

	// validate the wave
	logger.LogInfof("WaveOptAlgo - %s: validating wave", s.Wave.WaveSn)
	solverWave, orphanOrders, err := validateAndConvertWaveToSolverWave(s.Wave)
	if err != nil {
		logger.LogErrorf("WaveOptAlgo - %s: invalid wave data: %s", s.Wave.WaveSn, err.Error())
		return errorWaveResult(s.Wave.WaveSn, InvalidInputDateResult, err.Error())
	}
	logger.LogInfof("WaveOptAlgo - %s: wave data is valid", s.Wave.WaveSn)

	// detect conflicts between the wave and rules
	logger.LogInfof("WaveOptAlgo - %s: detecting conflict between wave and rules", s.Wave.WaveSn)
	err = detectConflictsBetweenWaveAndRules(solverWave, s.Wave.WaveRule)
	if err != nil {
		logger.LogErrorf("WaveOptAlgo - %s: conflict detected between wave and rules: %s", s.Wave.WaveSn, err.Error())
		return errorWaveResult(s.Wave.WaveSn, UnreasonableConstraintsResult, err.Error())
	}
	logger.LogInfof("WaveOptAlgo - %s: no conflict between wave and rules", s.Wave.WaveSn)

	// start the dispatcher
	groupNum := float64(len(solverWave.Groups))
	maxWorkers := int(math.Min(float64(s.Wave.SolverConfig.Parallelism), groupNum))
	dp := newGroupDispatcher(maxWorkers)
	s.dispatcher = dp

	totalRunningSeconds := s.Wave.SolverConfig.MaxSecondsSpent
	if totalRunningSeconds < fixedProcessingSeconds {
		return errorWaveResult(s.Wave.WaveSn, UnknownSolverErrorResult, fmt.Sprintf("at least %v seconds are required to solve a wave, but only %v allowed", fixedProcessingSeconds, totalRunningSeconds))
	}
	sort.Slice(solverWave.Groups, func(i, j int) bool {
		left := solverWave.Groups[i].Priority
		right := solverWave.Groups[j].Priority

		return left < right
	})

	// arrange the groups into batches in a round-robin way
	groupBatches := make([][]*SolverWaveGroup, 0)
	for i := 0; i < maxWorkers; i++ {
		groupBatches = append(groupBatches, make([]*SolverWaveGroup, 0))
	}
	batchIdx := 0
	for _, group := range solverWave.Groups {
		batch := groupBatches[batchIdx]
		batch = append(batch, group)
		groupBatches[batchIdx] = batch

		batchIdx++
		if batchIdx%maxWorkers == 0 {
			batchIdx = 0
		}
	}

	// count orders in each batch
	batchOrderCounts := make([]int, 0)
	for _, batch := range groupBatches {
		orderCount := 0
		for _, group := range batch {
			orderCount += len(group.Orders)
		}
		batchOrderCounts = append(batchOrderCounts, orderCount)
	}

	// allocate running time to groups. the groups in the same batch, i.e., a worker, share the running time
	// of (totalRunningSeconds-fixedProcessingSeconds). each group's running time is proportional to its number of orders
	groupRunTimes := make([]int64, 0)
	for i, group := range solverWave.Groups {
		batchOrderNum := batchOrderCounts[i%maxWorkers]
		groupRunTime := int64(float64(len(group.Orders)) / float64(batchOrderNum) * float64(totalRunningSeconds-fixedProcessingSeconds))
		groupRunTimes = append(groupRunTimes, groupRunTime)
	}

	for i, group := range solverWave.Groups {
		groupRunTime := groupRunTimes[i]
		grpJob := &groupJob{
			group:          group,
			rule:           s.Wave.WaveRule,
			solverConfig:   s.Wave.SolverConfig,
			solvingSeconds: groupRunTime,
		}
		dp.addJob(grpJob)
	}

	// clean dispatcher resources after solving is over
	defer dp.stop()
	dp.run()

	// wait for jobs completion
	jobDone := make(chan bool)
	go func(wg *sync.WaitGroup, jobDone chan bool) {
		wg.Wait()
		logger.LogInfof("WaveOptAlgo - %s: all groups have been computed", s.Wave.WaveSn)
		jobDone <- true
	}(dp.jobCompletionCountDownWg, jobDone)

	select {
	case <-jobDone:
		// combine & sort
		tasks := make([]*SolverWavePickingTask, 0)
		for result, e := range dp.resultMap {
			// combine error messages
			if e != nil {
				errMsg = fmt.Sprintf("%v;  %v: %v", errMsg, result.groupId, e.Error())
			}
			tasks = append(tasks, result.tasks...)
			orphanOrders = append(orphanOrders, result.orphanOrders...)
		}

		waveResult := combineAndSort(s.Wave.WaveSn, tasks, orphanOrders, s.Wave.WaveRule, s.Wave.SolverConfig, errMsg)

		return waveResult
	case <-s.terminate:
		return errorWaveResult(s.Wave.WaveSn, ExternalTermination, "")
	}
}

func (s *WavePickingTaskOptSolver) Terminate() error {
	for {
		select {
		case s.terminate <- true:
			logger.LogInfof("WaveOptAlgo - %s: async terminating", s.Wave.WaveSn)
			s.dispatcher.stop()
			s.dispatcher.workerShutDownWg.Wait()
			if s.dispatcher.isStopped {
				logger.LogInfof("WaveOptAlgo - %s: async terminated", s.Wave.WaveSn)
				return nil
			}
			return fmt.Errorf("fail to release the solver resource for %s", s.Wave.WaveSn)
		default:
			if s.dispatcher == nil || s.dispatcher.isStopped {
				return nil
			}
		}
	}
}

func errorWaveResult(waveSn string, retCode int64, errMsg string) *WaveResult {
	return &WaveResult{
		WaveSn:           waveSn,
		PickingTasks:     nil,
		UnsolvedOrderNos: nil,
		RetCode:          retCode,
		Msg:              errMsg,
	}
}
