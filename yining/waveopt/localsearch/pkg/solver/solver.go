package solver

type Solver interface {
	Solve() (Solution, Score, error)
	AsyncTerminate()
}

type Evaluator interface {
	SolverLifeCycleListener
	Evaluate(Solution) (Score, error)
}

type Move interface {
	Do(Solution) (Move, error)
	MovedVariables() ([]Variable, error)
	ToValues() ([]Value, error)
}

type MoveFactory interface {
	SolverLifeCycleListener
	CreateMove(Solution) (Move, error)
}

type Params struct {
	MovesPerStep int
}

type ScoreTrend int

const (
	UpScore   ScoreTrend = 1
	DownScore ScoreTrend = -1
)

type Score interface {
	Trend() (ScoreTrend, error)
	IsFeasible() (bool, error)
	CompareToScore(Score) (int, error)
	Sub(Score) ([]float64, error)
}

type Selector interface {
	SolverLifeCycleListener
	AddCandidateMove(*MoveContext) error
	Select() (*MoveContext, error)
}

type Acceptor interface {
	SolverLifeCycleListener
	IsAccepted(*MoveContext) (bool, error)
}

type Terminator interface {
	SolverLifeCycleListener
	IsTerminated() (bool, error)
	SolvingProgress(*StepContext) (float64, error)
}

type Identifiable interface {
	Identifier() interface{}
}

type Solution interface {
	Copy() (Solution, error)
}

type Value interface {
	Identifiable
	Variables() ([]Variable, error)
}

type Variable interface {
	Identifiable
	Value() (Value, error)
}

type MoveFactoryScheduler interface {
	SolverLifeCycleListener
	MoveFactories() ([]MoveFactory, error)
	SelectMoveFactory() (MoveFactory, error)
}

type SolverLifeCycleListener interface {
	UpdateAtSolverStart(*SolverContext) error
	UpdateAtStepStart(*StepContext) error
	UpdateAtMoveStart(*MoveContext) error

	UpdateAtSolverEnd(*SolverContext) error
	UpdateAtStepEnd(*StepContext) error
	UpdateAtMoveEnd(*MoveContext) error
}
