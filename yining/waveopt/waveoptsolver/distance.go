package waveoptsolver

import (
	"fmt"
)

type DistanceCalculator interface {
	CalculateDistance(*SolverWaveSubPickingTask) float64
}

// in a sub picking task, sku possess differenct zones/pathways/segments. denote them as z_i, p_i and s_i
// the picking distance d_t of a sub tasks t is a weighted sum of z_i, p_i and s_i of the skus, respectively.
type LocationBasedDistanceCalculator struct {
	ZoneWeight    float64
	PathwayWeight float64
	SegmentWeight float64
}

func (c *LocationBasedDistanceCalculator) CalculateDistance(subPickingTask *SolverWaveSubPickingTask) float64 {
	if subPickingTask == nil || subPickingTask.Skus == nil || len(subPickingTask.Skus) == 0 {
		return 0.0
	}

	zoneMap := make(map[string]bool)
	pathwayMap := make(map[string]bool)
	segmentMap := make(map[string]bool)

	for _, sku := range subPickingTask.Skus {
		zoneMap[sku.Location.ZoneId] = true
		pathwayMap[sku.Location.PathwayId] = true
		segmentMap[sku.Location.SegmentId] = true
	}

	zoneNum := len(zoneMap)
	pathwayNum := len(pathwayMap)
	segmentNum := len(segmentMap)

	return float64(zoneNum)*c.ZoneWeight + float64(pathwayNum)*c.PathwayWeight + float64(segmentNum)*c.SegmentWeight
}

func NewLocationBasedDistanceCalculator(zoneWeight, pathwayWeight, segmentWeight float64) (*LocationBasedDistanceCalculator, error) {
	if zoneWeight <= 0.0 {
		return nil, fmt.Errorf("weight of zone in distance calculation must be positive, but %v found", zoneWeight)
	}
	if pathwayWeight <= 0.0 {
		return nil, fmt.Errorf("weight of pathway in distance calculation must be positive, but %v found", pathwayWeight)
	}
	if segmentWeight <= 0.0 {
		return nil, fmt.Errorf("weight of segment in distance calculation must be positive, but %v found", segmentWeight)
	}

	return &LocationBasedDistanceCalculator{
		ZoneWeight:    zoneWeight,
		PathwayWeight: pathwayWeight,
		SegmentWeight: segmentWeight,
	}, nil
}
