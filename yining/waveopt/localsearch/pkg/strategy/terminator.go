package strategy

import (
	"fmt"
	"math"
	"time"

	merror "git.garena.com/shopee/bg-logistics/algo/wms/waveopt/localsearch/pkg/error"
	"git.garena.com/shopee/bg-logistics/algo/wms/waveopt/localsearch/pkg/solver"
)

type TimeUnit int64

const (
	Hour        TimeUnit = 60 * Minute
	Minute               = 60 * Second
	Second               = 1000 * MilliSecond
	MilliSecond          = 1000000 * NanoSecond
	NanoSecond           = 1
)

type BaseTerminator struct{}

func (t *BaseTerminator) UpdateAtSolverStart(solverContext *solver.SolverContext) error { return nil }
func (t *BaseTerminator) UpdateAtSolverEnd(solverContext *solver.SolverContext) error   { return nil }
func (t *BaseTerminator) UpdateAtStepStart(stepContext *solver.StepContext) error       { return nil }
func (t *BaseTerminator) UpdateAtStepEnd(stepContext *solver.StepContext) error         { return nil }
func (t *BaseTerminator) UpdateAtMoveStart(moveContext *solver.MoveContext) error       { return nil }
func (t *BaseTerminator) UpdateAtMoveEnd(moveContext *solver.MoveContext) error         { return nil }

type StepCountLimitTerminator struct {
	*BaseTerminator
	stepCountLimit int

	currentStepCount int
}

func (t *StepCountLimitTerminator) UpdateAtStepStart(stepContext *solver.StepContext) error {
	t.currentStepCount = stepContext.StepCount
	return nil
}

func (t *StepCountLimitTerminator) IsTerminated() (bool, error) {
	return t.currentStepCount >= t.stepCountLimit, nil
}

func (t *StepCountLimitTerminator) SolvingProgress(stepContext *solver.StepContext) (float64, error) {
	return math.Min(1.0, float64(stepContext.StepCount)/float64(t.stepCountLimit)), nil
}

type StepCountLimitTerminatorBuilder struct {
	stepCountLimit int
}

func NewStepCountLimitTerminatorBuilder() *StepCountLimitTerminatorBuilder {
	return &StepCountLimitTerminatorBuilder{}
}

func (b *StepCountLimitTerminatorBuilder) WithStepCountLimit(stepCountLimit int) *StepCountLimitTerminatorBuilder {
	b.stepCountLimit = stepCountLimit
	return b
}

func (b *StepCountLimitTerminatorBuilder) Build() (*StepCountLimitTerminator, error) {
	return &StepCountLimitTerminator{
		stepCountLimit: b.stepCountLimit,
	}, nil
}

type TimeLimitTerminator struct {
	*BaseTerminator
	nanoSecondsLimit int64

	startTime   time.Time
	currentTime time.Time
}

func (t *TimeLimitTerminator) UpdateAtSolverStart(solverContext *solver.SolverContext) error {
	t.startTime = solverContext.StartTime
	t.currentTime = t.startTime
	return nil
}

func (t *TimeLimitTerminator) UpdateAtStepStart(stepContext *solver.StepContext) error {
	t.currentTime = stepContext.StartTime
	return nil
}

func (t *TimeLimitTerminator) UpdateAtMoveStart(moveContext *solver.MoveContext) error {
	t.currentTime = moveContext.StartTime
	return nil
}

func (t *TimeLimitTerminator) IsTerminated() (bool, error) {
	return t.currentTime.Sub(t.startTime).Nanoseconds() > t.nanoSecondsLimit, nil
}

func (t *TimeLimitTerminator) SolvingProgress(stepContext *solver.StepContext) (float64, error) {
	duration := stepContext.StartTime.Sub(t.startTime).Nanoseconds()
	return math.Min(1.0, float64(duration)/float64(t.nanoSecondsLimit)), nil
}

type TimeLimitTerminatorBuilder struct {
	timeLimitMap map[TimeUnit]int64
}

func NewTimeLimitTerminatorBuilder() *TimeLimitTerminatorBuilder {
	builder := &TimeLimitTerminatorBuilder{
		timeLimitMap: make(map[TimeUnit]int64),
	}

	return builder
}

func (b *TimeLimitTerminatorBuilder) WithTimeLimit(time int64, unit TimeUnit) *TimeLimitTerminatorBuilder {
	if time <= 0 {
		return b
	}

	b.timeLimitMap[unit] = time

	return b
}

func (b *TimeLimitTerminatorBuilder) Build() (*TimeLimitTerminator, error) {
	nanos := int64(1) // TimeLimitTerminator's nanoSecondsLimit is greater or equal to 1 to avoid dividing zero error when calculating SolvingProgress in StepContext and MoveContext
	for time_, unit := range b.timeLimitMap {
		nanos += int64(time_) * unit
	}

	return &TimeLimitTerminator{
		nanoSecondsLimit: nanos,
	}, nil
}

type UnimprovedStepCountLimitTerminator struct {
	*BaseTerminator
	unimprovedStepCountLimit int

	bestStepCount    int
	currentStepCount int
}

func (t *UnimprovedStepCountLimitTerminator) UpdateAtSolverStart(solverContext *solver.SolverContext) error {
	t.bestStepCount = 0
	t.currentStepCount = 0
	return nil
}

func (t *UnimprovedStepCountLimitTerminator) UpdateAtStepStart(stepContext *solver.StepContext) error {
	t.bestStepCount = stepContext.SolverContext.BestSolutionStepCount
	t.currentStepCount = stepContext.StepCount
	return nil
}

func (t *UnimprovedStepCountLimitTerminator) IsTerminated() (bool, error) {
	return t.currentStepCount-t.bestStepCount-1 >= t.unimprovedStepCountLimit, nil
}

func (t *UnimprovedStepCountLimitTerminator) SolvingProgress(stepContext *solver.StepContext) (float64, error) {
	return math.Min(1.0, float64(stepContext.StepCount-stepContext.SolverContext.BestSolutionStepCount-1)/float64(t.unimprovedStepCountLimit)), nil
}

type UnimprovedStepCountLimitTerminatorBuilder struct {
	stepCountLimit int
}

func NewUnimprovedStepCountLimitTerminatorBuilder() *UnimprovedStepCountLimitTerminatorBuilder {
	return &UnimprovedStepCountLimitTerminatorBuilder{}
}

func (b *UnimprovedStepCountLimitTerminatorBuilder) WithStepCountLimit(unimprovedStepCountLimit int) *UnimprovedStepCountLimitTerminatorBuilder {
	b.stepCountLimit = unimprovedStepCountLimit

	return b
}

func (b *UnimprovedStepCountLimitTerminatorBuilder) Build() (*UnimprovedStepCountLimitTerminator, error) {
	return &UnimprovedStepCountLimitTerminator{
		unimprovedStepCountLimit: b.stepCountLimit,
	}, nil
}

type UnimprovedTimeLimitTerminator struct {
	*BaseTerminator
	unimprovedNanoSecondsLimit int64
	bestSolutionTime           time.Time
	currentTime                time.Time
}

func (t *UnimprovedTimeLimitTerminator) UpdateAtSolverStart(solverContext *solver.SolverContext) error {
	t.bestSolutionTime = solverContext.StartTime
	t.currentTime = t.bestSolutionTime
	return nil
}

func (t *UnimprovedTimeLimitTerminator) UpdateAtStepStart(stepContext *solver.StepContext) error {
	t.bestSolutionTime = stepContext.SolverContext.BestSolutionTime
	t.currentTime = stepContext.StartTime
	return nil
}

func (t *UnimprovedTimeLimitTerminator) IsTerminated() (bool, error) {
	return t.currentTime.Sub(t.bestSolutionTime).Nanoseconds() >= t.unimprovedNanoSecondsLimit, nil
}

func (t *UnimprovedTimeLimitTerminator) SolvingProgress(stepContext *solver.StepContext) (float64, error) {
	progress := float64(t.currentTime.Sub(t.bestSolutionTime).Nanoseconds()) / float64(t.unimprovedNanoSecondsLimit)
	if progress < 0.0 {
		return 0.0, merror.New("UnimprovedTimeLimitTerminator: Weird solving progress", fmt.Sprintf("%.2f%%", progress*100))
	}

	return math.Min(1.0, progress), nil
}

type UnimprovedTimeLimitTerminatorBuilder struct {
	timeLimitMap map[TimeUnit]int64
}

func NewUnimprovedTimeLimitTerminatorBuilder() *UnimprovedTimeLimitTerminatorBuilder {
	return &UnimprovedTimeLimitTerminatorBuilder{
		timeLimitMap: make(map[TimeUnit]int64),
	}
}

func (b *UnimprovedTimeLimitTerminatorBuilder) WithUnimprovedTimeLimit(time int64, unit TimeUnit) *UnimprovedTimeLimitTerminatorBuilder {
	oldTime := b.timeLimitMap[unit]
	b.timeLimitMap[unit] = oldTime + time

	return b
}

func (b *UnimprovedTimeLimitTerminatorBuilder) Build() (*UnimprovedTimeLimitTerminator, error) {
	nanoSecondsLimit := int64(1)
	for time_, unit := range b.timeLimitMap {
		nanoSecondsLimit += int64(time_) * unit
	}

	return &UnimprovedTimeLimitTerminator{
		unimprovedNanoSecondsLimit: nanoSecondsLimit,
	}, nil
}
