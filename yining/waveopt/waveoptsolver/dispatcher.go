package waveoptsolver

import (
	"git.garena.com/shopee/bg-logistics/go/gocommon/logger"
	"sync"
)

type groupJob struct {
	group          *SolverWaveGroup
	rule           *WaveRule
	solverConfig   *WaveSolverConfig
	solvingSeconds int64
}

type groupWorker struct {
	id           int64
	groupJobChan chan *groupJob
	dispatcher   *groupDispatcher
	quitSolving  chan bool
	quitWorking  chan bool
}

func newGroupWorker(dispatcher *groupDispatcher, idx int64) *groupWorker {
	worker := &groupWorker{
		id:           idx,
		groupJobChan: make(chan *groupJob, 1),
		dispatcher:   dispatcher,
		quitSolving:  make(chan bool, 1),
		quitWorking:  make(chan bool, 1),
	}
	dispatcher.workers = append(dispatcher.workers, worker)
	return worker
}

func (w *groupWorker) run() {
	logger.LogInfof("WaveOptAlgo - Worker %d starts.", w.id)
	go func() {
		for {
			w.dispatcher.workerPool <- w.groupJobChan

			select {
			case grpJob := <-w.groupJobChan:
				var grpResult *groupResult
				var err error

				if grpJob.rule.WavePickerMode == SinglePickerOnly || grpJob.rule.isZeroMulti() {
					grpResult, err = solveSinglePickerGroup(grpJob.group, grpJob.rule, grpJob.solverConfig, grpJob.solvingSeconds, w.quitSolving)
				} else {
					grpResult, err = solveMultiPickerGroup(grpJob.group, grpJob.rule, grpJob.solverConfig, grpJob.solvingSeconds, w.quitSolving)
				}
				if grpResult == nil || err != nil {
					grpResult = &groupResult{
						groupId:       grpJob.group.Id,
						groupPriority: grpJob.group.Priority,
						groupType:     grpJob.group.GroupType,
						tasks:         nil,
						orphanOrders:  append(make([]*SolverWaveOrder, 0), grpJob.group.Orders...),
					}
				}
				w.dispatcher.resultLock.Lock()
				w.dispatcher.resultMap[grpResult] = err
				w.dispatcher.resultLock.Unlock()
				w.dispatcher.jobCompletionCountDownWg.Done()
			case <-w.quitWorking:
				w.dispatcher.workerShutDownWg.Done()
				logger.LogInfof("WaveOptAlgo - Worker %d quits.", w.id)
				return
			}
		}
	}()
}

type groupDispatcher struct {
	jobQueue                 chan *groupJob
	workerPool               chan chan *groupJob
	workers                  []*groupWorker
	maxWorkers               int
	resultMap                map[*groupResult]error
	resultLock               *sync.Mutex
	stopDispatcher           chan bool
	isStopped                bool
	workerShutDownWg         *sync.WaitGroup
	jobCompletionCountDownWg *sync.WaitGroup
}

func newGroupDispatcher(maxWorkers int) *groupDispatcher {
	return &groupDispatcher{
		workerPool:               make(chan chan *groupJob, maxWorkers),
		jobQueue:                 make(chan *groupJob, 1e4),
		maxWorkers:               maxWorkers,
		workers:                  make([]*groupWorker, 0),
		resultMap:                make(map[*groupResult]error),
		resultLock:               &sync.Mutex{},
		stopDispatcher:           make(chan bool, 1),
		isStopped:                false,
		workerShutDownWg:         &sync.WaitGroup{},
		jobCompletionCountDownWg: &sync.WaitGroup{},
	}
}

func (d *groupDispatcher) run() {
	d.workerShutDownWg.Add(1)
	for i := 0; i < d.maxWorkers; i++ {
		worker := newGroupWorker(d, int64(i))
		d.workerShutDownWg.Add(1)
		worker.run()
	}

	go d.dispatch()
}

func (d *groupDispatcher) addJob(job *groupJob) {
	d.jobCompletionCountDownWg.Add(1)
	d.jobQueue <- job
}

func (d *groupDispatcher) dispatch() {
	for {
		select {
		case job := <-d.jobQueue:
			go func(job *groupJob) {
				workerJobQueue := <-d.workerPool
				workerJobQueue <- job
			}(job)
		case <-d.stopDispatcher:
			d.workerShutDownWg.Done()
			return
		}
	}
}

func (d *groupDispatcher) stop() {
	// if a dispatcher has been stopped, it needs not to be stopped again
	if d.isStopped {
		return
	}

	for _, worker := range d.workers {
		worker.quitSolving <- true
	}
	for _, worker := range d.workers {
		worker.quitWorking <- true
	}

	d.stopDispatcher <- true
	d.isStopped = true
}
