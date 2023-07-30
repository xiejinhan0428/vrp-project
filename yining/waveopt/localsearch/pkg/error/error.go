package error

import (
	"fmt"
	"strings"
)

// customized error type used in metaheuristcs
// so far (2021-12, alpha version), runtime/config/data errors share this error type
type MetaHeurError struct {
	msg string
}

// implementing the "error" interface
func (err *MetaHeurError) Error() string {
	return err.msg
}

// create a new MetaHeurError
func New(msg ...string) *MetaHeurError {
	err := MetaHeurError{
		msg: fmt.Sprintf("MetaHeurSolver: %v", strings.Join(msg, " ")),
	}

	return &err
}

func (err *MetaHeurError) CausedBy(errs ...error) *MetaHeurError {
	errMsgs := make([]string, 0)
	for _, err := range errs {
		errMsgs = append(errMsgs, err.Error())
	}

	err.msg = fmt.Sprintf("\nCaused by %v", strings.Join(errMsgs, "\nCaused by "))
	return err
}
