package score

import (
	"fmt"
	"strings"

	"git.garena.com/shopee/bg-logistics/algo/wms/waveopt/localsearch/internal/arithutil"
	merror "git.garena.com/shopee/bg-logistics/algo/wms/waveopt/localsearch/pkg/error"
	"git.garena.com/shopee/bg-logistics/algo/wms/waveopt/localsearch/pkg/solver"
)

// a float implementation of the Score interface

type FloatScore struct {
	ConstraintScores []float64
	ObjectiveScores  []float64
	ScoreTrend       solver.ScoreTrend
}

// implements the Score interface

func (fs *FloatScore) Trend() (solver.ScoreTrend, error) {
	err := checkFloatScore(*fs)
	if err != nil {
		return solver.DownScore, err
	}
	return fs.ScoreTrend, nil
}

func (fs *FloatScore) IsFeasible() (bool, error) {
	err := checkFloatScore(*fs)
	if err != nil {
		return false, err
	}

	trend := float64(fs.ScoreTrend)
	for _, s := range fs.ConstraintScores {
		if s*trend < 0.0 {
			return false, nil
		}
	}
	return true, nil
}

func (fs *FloatScore) CompareToScore(s solver.Score) (int, error) {
	that, ok := s.(*FloatScore)
	if !ok {
		return 0, merror.New("FloatScore: Comparing FloatScore with unknown type Score.")
	}

	err := checkFloatScore(*fs)
	if err != nil {
		return 0, err
	}

	if fs.ScoreTrend != that.ScoreTrend {
		thisArticle := "an"
		if fs.ScoreTrend == solver.DownScore {
			thisArticle = "a"
		}

		thatArticle := "an"
		if that.ScoreTrend == solver.DownScore {
			thatArticle = "solver.DownScore"
		}

		thisTrendName := "UP"
		if fs.ScoreTrend == solver.DownScore {
			thisTrendName = "solver.DownScore"
		}
		thatTrendName := "UP"
		if that.ScoreTrend == solver.DownScore {
			thatTrendName = "solver.DownScore"
		}

		return 0, merror.New("Comparing ", thisArticle, thisTrendName, "FloatScore with", thatArticle, thatTrendName, "score.")
	}

	result := 0

	thisCons := fs.ConstraintScores
	thatCons := that.ConstraintScores
	size, err := arithutil.MinInt(len(thisCons), len(thatCons))
	if err != nil {
		return 0, err
	}
	for i := 0; i < size; i++ {
		if thisCons[i] == thatCons[i] {
			continue
		} else if thisCons[i] > thatCons[i] {
			result = 1
			break
		} else {
			result = -1
			break
		}
	}

	if result != 0 {
		return int(fs.ScoreTrend) * result, nil
	}

	thisObjs := fs.ObjectiveScores
	thatObjs := that.ObjectiveScores
	size, err = arithutil.MinInt(len(thisObjs), len(thatObjs))
	if err != nil {
		return 0, err
	}
	for i := 0; i < size; i++ {
		if thisObjs[i] == thatObjs[i] {
			continue
		} else if thisObjs[i] > thatObjs[i] {
			result = 1
			break
		} else {
			result = -1
			break
		}
	}

	return result * int(fs.ScoreTrend), nil
}

func (fs *FloatScore) Sub(score solver.Score) ([]float64, error) {
	scr, isFloatScore := score.(*FloatScore)
	if !isFloatScore {
		return nil, merror.New("FloatScore cannot Sub non-FloatScore.")
	}

	err := checkFloatScore(*fs)
	if err != nil {
		return nil, err
	}
	err = checkFloatScore(*scr)
	if err != nil {
		return nil, err
	}
	if len(fs.ConstraintScores) != len(scr.ConstraintScores) {
		return nil, merror.New("FloatScore: Cannot perform substraction on scores with different constraint level:", fs.String(), scr.String())
	}
	if len(fs.ObjectiveScores) != len(scr.ObjectiveScores) {
		return nil, merror.New("FloatScore: Cannot perform substraction on scores with different objective level:", fs.String(), scr.String())
	}

	diff := make([]float64, 0)
	for i, cons := range fs.ConstraintScores {
		diff = append(diff, cons-scr.ConstraintScores[i])
	}
	for i, obj := range fs.ObjectiveScores {
		diff = append(diff, obj-scr.ObjectiveScores[i])
	}

	return diff, nil
}

// implements the fmt.Stringer interface
func (fs *FloatScore) String() string {
	strCons := make([]string, len(fs.ConstraintScores))
	for i, v := range fs.ConstraintScores {
		strCons[i] = fmt.Sprint(v)
	}
	consStr := strings.Join(strCons, "/")

	strObjs := make([]string, len(fs.ObjectiveScores))
	for i, v := range fs.ObjectiveScores {
		strObjs[i] = fmt.Sprint(v)
	}
	objsStr := strings.Join(strObjs, "/")

	return fmt.Sprintf("FloatScore{Cons:[%v], Objs:[%v]}", consStr, objsStr)
}

// NewFloatScore create a FloatScore by specifying the trend and the levels of constraints and objectives
func NewFloatScore(consLevel int, objLevel int, trend solver.ScoreTrend) (FloatScore, error) {
	if consLevel < 1 {
		return FloatScore{}, merror.New("FloatScore: Cannot initialize a FloatScore with non-positive constraint level")
	}

	if objLevel < 1 {
		return FloatScore{}, merror.New("FloatScore: Cannot initialize a FloatScore with non-positive objective level")
	}

	cons := make([]float64, consLevel)
	obj := make([]float64, objLevel)
	fs := FloatScore{
		ConstraintScores: cons,
		ObjectiveScores:  obj,
		ScoreTrend:       trend,
	}
	return fs, nil
}

// FloatScoreFrom create a FloatScore by values of constraints and objectives and trend. levels of constraints and objectives are lengths of the corresponding slices.
func FloatScoreFrom(cons []float64, obj []float64, trend solver.ScoreTrend) (FloatScore, error) {
	consLevelNum := len(cons)
	objLevelNum := len(obj)

	if consLevelNum < 1 {
		return FloatScore{}, merror.New("FloatScore: Cannot initialize a FloatScore with non-positive constraint level")
	}

	if objLevelNum < 1 {
		return FloatScore{}, merror.New("FloatScore: Cannot initialize a FloatScore with non-positive objective level")
	}

	fs := FloatScore{
		ConstraintScores: cons,
		ObjectiveScores:  obj,
		ScoreTrend:       trend,
	}

	return fs, nil
}

func checkFloatScore(score FloatScore) error {
	if score.ConstraintScores == nil || len(score.ConstraintScores) == 0 {
		return merror.New("FloatScore: Zero length constraint part of a FloatScore.")
	}
	if score.ObjectiveScores == nil || len(score.ObjectiveScores) == 0 {
		return merror.New("FloatScore: Zero length objective part of a FloatScore.")
	}
	if score.ScoreTrend != solver.UpScore && score.ScoreTrend != solver.DownScore {
		return merror.New("FloatScore: Unknown trend type of a FloatScore ", fmt.Sprint(score.Trend()))
	}

	return nil
}
