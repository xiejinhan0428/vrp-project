package solver

type NoChangeMove struct{}

func (m *NoChangeMove) Do(solution Solution) (Move, error) {
	return &NoChangeMove{}, nil
}

func (m *NoChangeMove) MovedVariables() ([]Variable, error) {
	return make([]Variable, 0), nil
}

func (m *NoChangeMove) ToValues() ([]Value, error) {
	return make([]Value, 0), nil
}

func (m *NoChangeMove) String() string {
	return "NoChangeMove{}"
}
