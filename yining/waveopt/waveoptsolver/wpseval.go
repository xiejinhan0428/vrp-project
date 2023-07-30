package waveoptsolver

import (
	"errors"
	"git.garena.com/shopee/bg-logistics/algo/wms/waveopt/localsearch/pkg/score"
	"git.garena.com/shopee/bg-logistics/algo/wms/waveopt/localsearch/pkg/solver"
	"math"
	"sort"
)

//
type WavePickingSingleEvaluator struct {
	*score.BaseEvaluator
	rule   *WaveRule
	config *WaveSolverConfig
}

// 解指标计算

func (pt *PickingTaskSingle) PickingTaskEvaluate(rule *WaveRule, config *WaveSolverConfig, level SplitLevel) {
	// 计算pickingtask指标
	pt.OrderQty = 0
	pt.Distance = 0
	pt.isUrgent = false
	pt.ConstraintViolation = 0
	pt.SkuQty = 0
	pt.Volume = 0
	pt.SplitLevelNum = 0
	tmpLocation := make([]*SolverSkuLocation, 0)
	tmpSplitSet := make(map[string]bool)
	for _, order := range pt.Orders {
		// 存储是否紧急单信息
		if order.isUrgent == true {
			pt.isUrgent = true
		}

		// 计算订单数量
		pt.OrderQty += 1

		for _, sku := range order.Skus {
			tmpSplitLevel, _ := extractSplitLevel(sku.Location, level)
			tmpSplitSet[tmpSplitLevel] = true
			tmpLocation = append(tmpLocation, sku.Location)
			pt.SkuQty += sku.Qty
			pt.Volume += sku.totalVolume
		}
	}
	pt.Distance = CalculateDistance(tmpLocation, config)
	pt.SplitLevelNum = int64(len(tmpSplitSet))
	if len(pt.Orders) > 0 {
		if pt.SkuQty > rule.CommonRule.MaxItemQtyPerSubPickingTask {
			pt.ConstraintViolation += pt.SkuQty - rule.CommonRule.MaxItemQtyPerSubPickingTask
		}
		if pt.SkuQty < rule.CommonRule.MinItemQtyPerSubPickingTask {
			pt.ConstraintViolation += rule.CommonRule.MinItemQtyPerSubPickingTask - pt.SkuQty
		}
		if pt.OrderQty > rule.CommonRule.MaxOrderQtyPerPickingTask {
			pt.ConstraintViolation += pt.OrderQty - rule.CommonRule.MaxOrderQtyPerPickingTask
		}
		if pt.OrderQty < rule.CommonRule.MinOrderQtyPerPickingTask {
			pt.ConstraintViolation += rule.CommonRule.MinOrderQtyPerPickingTask - pt.OrderQty
		}
		if pt.Volume > rule.CommonRule.MaxItemVolumePerSubPickingTask {
			pt.ConstraintViolation += pt.Volume - rule.CommonRule.MaxItemVolumePerSubPickingTask
		}
		if pt.Volume < rule.CommonRule.MinItemVolumePerSubPickingTask {
			pt.ConstraintViolation += rule.CommonRule.MinItemVolumePerSubPickingTask - pt.Volume
		}
		if level != "" && pt.SplitLevelNum > 1 {
			pt.ConstraintViolation += pt.SplitLevelNum - 1
		}
	}
}

func (wps *WavePickingSingleEvaluator) Evaluate(solution solver.Solution) (solver.Score, error) {
	sol, ok := solution.(*WaveOptSolutionSingle)
	if !ok {
		return nil, errors.New("BinPackingEvaluator: The solution to be evaluated must be a instance of BinPackingSolution")
	}

	splitLevel, _ := wps.rule.splitLevel()
	if !wps.rule.CommonRule.IsCrossZoneCluster {
		if splitLevel == ByZoneSector {
			splitLevel = ByZoneCluster
		}
	}

	var violationDegree int64
	distance := int64(0)
	sol.normalTasks = make([]*PickingTaskSingle, 0)
	sol.orphanOrders = make([]*SolverWaveOrder, 0)
	for _, pt := range sol.PickingTasks {
		if len(pt.Orders) <= 0 {
			continue
		}

		pt.PickingTaskEvaluate(wps.rule, wps.config, splitLevel)

		if pt.ConstraintViolation > 0 {
			//fmt.Println(pt)
			violationDegree += pt.ConstraintViolation
			sol.orphanOrders = append(sol.orphanOrders, pt.Orders...)
		} else {
			sol.normalTasks = append(sol.normalTasks, pt)
		}
		distance += pt.Distance
	}

	newScore, _ := wps.NewScore()
	score, _ := newScore.(*score.FloatScore)
	score.ConstraintScores[0] = float64(len(sol.orphanOrders))
	score.ConstraintScores[1] = float64(violationDegree)
	score.ObjectiveScores[0] = float64(distance)

	return score, nil

}

func (wps *WavePickingSingleEvaluator) NewScore() (solver.Score, error) {
	score, _ := score.NewFloatScore(2, 1, solver.DownScore)
	return &score, nil
}

func CalculateDistance(locations []*SolverSkuLocation, config *WaveSolverConfig) int64 {
	// 计算距离损失函数
	ZoneSet := make(map[string]bool)
	SegmentSet := make(map[string]bool)
	PathwaySet := make(map[string]bool)
	for _, i := range locations {
		ZoneSet[i.ZoneId] = true
		SegmentSet[i.SegmentId] = true
		PathwaySet[i.PathwayId] = true
	}
	distance := int64(config.ZoneCoeff)*int64(len(ZoneSet)) +
		int64(config.SegmentCoeff)*int64(len(SegmentSet)) +
		int64(config.PathwayCoeff)*int64(len(PathwaySet))
	return distance
}

func CalculatePickingSequence(locations []*SolverSkuLocation) int64 {
	minSequence := int64(1e10)
	maxSequence := int64(-1e10)
	for _, i := range locations {
		if i.PickingSequence > maxSequence {
			maxSequence = i.PickingSequence
		}
		if i.PickingSequence < minSequence {
			minSequence = i.PickingSequence
		}
	}
	distance := maxSequence - minSequence
	return distance
}

func CalculateManhattanDistance(locations []*SolverSkuLocation) int64 {
	sort.Slice(locations, func(i, j int) bool {
		return locations[i].PickingSequence > locations[j].PickingSequence
	})
	tmpDistance := float64(0)
	if len(locations) > 0 {
		tmpDistance += math.Abs(87.1-locations[0].CoordinateX) + math.Abs(-36.34-locations[0].CoordinateY)
		tmpDistance += math.Abs(locations[len(locations)-1].CoordinateX-87.1) + math.Abs(locations[len(locations)-1].CoordinateY+36.34)
	}
	for ind, loc := range locations {
		if ind == len(locations)-1 {
			break
		}
		tmpDistance += (math.Abs(loc.CoordinateY-locations[ind+1].CoordinateY) + math.Abs(loc.CoordinateX-locations[ind+1].CoordinateX))
	}
	distance := int64(tmpDistance)
	return distance
}
