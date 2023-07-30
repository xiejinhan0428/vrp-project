package waveoptsolver

import (
	"errors"
	"fmt"
	"git.garena.com/shopee/bg-logistics/go/gocommon/logger"
	"math"
	"sort"
	"strconv"
	"strings"
)

// validate the *Wave while converting it to a *SolverWave
func validateAndConvertWaveToSolverWave(wave *Wave) (*SolverWave, []*SolverWaveOrder, error) {
	var err error

	// validate the wave
	err = validateWave(wave)
	if err != nil {
		return nil, nil, err
	}

	// retrieve the wave rule
	waveRule := wave.WaveRule

	// new a *SolverWave
	solverWave := &SolverWave{
		Id:     wave.WaveSn,
		Groups: make([]*SolverWaveGroup, 0),
	}

	orphanOrders := make([]*SolverWaveOrder, 0)

	// validate and hash and convert locations
	// key: location id; value: *Location
	locationMap := make(map[string]*SolverSkuLocation)
	for _, location := range wave.Locations {
		err = validateLocation(location)
		if err != nil {
			return nil, nil, err
		}

		locationMap[location.LocationId] = &SolverSkuLocation{
			LocationId:    location.LocationId,
			SegmentId:     strings.Join([]string{location.ZoneId, location.PathwayId, location.SegmentId}, "_"),
			PathwayId:     strings.Join([]string{location.ZoneId, location.PathwayId}, "_"),
			ZoneId:        location.ZoneId,
			ZoneClusterId: strconv.FormatInt(location.ZoneClusterId, 10),
			ZoneSectorId:  strconv.FormatInt(location.ZoneSectorId, 10),
		}
	}

	// validate and hash SKUs
	// key: SKU id; value: *SolverWaveSku
	skuMap := make(map[string]*WaveSku)
	for _, sku := range wave.Skus {
		err = validateWaveSku(sku)
		if err != nil {
			return nil, nil, err
		}

		skuMap[sku.SkuId] = sku
	}

	convertedOrderNum := 0
	maxSolveOrderNum := math.MaxInt64
	if wave.SolverConfig.MaxWaveSnSolveOrder > 0 {
		maxSolveOrderNum = int(wave.SolverConfig.MaxWaveSnSolveOrder)
	}
	// validate and convert groups, add to solverWave

	// sort group by priority, asc
	sort.SliceStable(wave.Groups, func(i, j int) bool {
		left := wave.Groups[i]
		right := wave.Groups[j]

		return left.Priority < right.Priority
	})

	// a map to merge groups with the same sn. key: group sn; value: *SolverWaveGroup
	groupMap := make(map[string]*SolverWaveGroup)
	for _, group := range wave.Groups {
		err = validateWaveGroup(group)
		if err != nil {
			return nil, nil, err
		}

		// skip groups without orders
		if len(group.Orders) == 0 {
			continue
		}

		// new or get a group
		var solverGroup *SolverWaveGroup
		if theGroup, ok := groupMap[group.GroupSn]; ok {
			solverGroup = theGroup
		} else {
			solverGroup = &SolverWaveGroup{
				Id:        group.GroupSn,
				Wave:      solverWave,
				GroupType: group.GroupType,
				Priority:  group.Priority,
				Orders:    make([]*SolverWaveOrder, 0),
			}
			groupMap[group.GroupSn] = solverGroup
		}

		// sort order by cut off time, asc
		sort.SliceStable(group.Orders, func(i, j int) bool {
			return group.Orders[i].CutOffTime < group.Orders[j].CutOffTime
		})

		// convert orders, add to solverGroup
		for _, order := range group.Orders {
			err = validateWaveOrder(order)
			if err != nil {
				return nil, nil, err
			}

			// new a *SolverWaveOrder
			solverOrder := &SolverWaveOrder{
				Id:                 order.OrderNo,
				Group:              solverGroup,
				Skus:               make([]*SolverWaveSku, 0),
				PickingTask:        nil,
				isUrgent:           order.IsUrgent,
				isMultiPickerOrder: false,
			}

			// convert SKU
			cross := make(map[string]bool)
			orderTotalVolume := int64(0)
			orderTotalSkuQty := int64(0)
			for _, orderSku := range order.Skus {
				err = validateOrderSku(orderSku, order.OrderNo, skuMap, locationMap)
				if err != nil {
					return nil, nil, err
				}

				waveSku := skuMap[orderSku.SkuId]
				solverSku := &SolverWaveSku{
					Id:       orderSku.SkuId,
					Order:    solverOrder,
					Location: locationMap[orderSku.LocationId],
					Length:   waveSku.Length,
					Width:    waveSku.Width,
					Height:   waveSku.Height,
					Weight:   waveSku.Weight,
					Qty:      orderSku.Qty,
					SkuSize:  waveSku.SkuSize,
				}

				// calculate sku volume
				solverSku.Volume = solverSku.SingleVolume()
				solverSku.totalVolume = solverSku.TotalVolume()

				orderTotalSkuQty += solverSku.Qty
				orderTotalVolume += solverSku.totalVolume

				if waveRule.WavePickerMode == SinglePickerOnly {
					solverOrder.Skus = append(solverOrder.Skus, solverSku)
					continue
				}

				// below are peculiar processing in multi-picker modes
				level, err := waveRule.splitLevel()
				if err != nil {
					return nil, nil, err
				}
				crossLocation, err := extractSplitLevel(solverSku.Location, level)
				if err != nil {
					return nil, nil, err
				}

				cross[crossLocation] = true

				// if the quantity (total volume) of an SKU exceeds the limit of a sub picking task,
				// split this SKU into multiple SKUs
				if solverSku.Qty <= waveRule.CommonRule.MaxItemQtyPerSubPickingTask && solverSku.TotalVolume() <= waveRule.CommonRule.MaxItemVolumePerSubPickingTask {
					solverOrder.Skus = append(solverOrder.Skus, solverSku)
				} else if solverSku.Qty > waveRule.CommonRule.MaxItemQtyPerSubPickingTask {
					solverOrder.isMultiPickerOrder = true
					skuQty := solverSku.Qty
					for skuQty > 0 {
						newSku := *solverSku
						if skuQty > waveRule.CommonRule.MaxItemQtyPerSubPickingTask {
							newSku.Qty = waveRule.CommonRule.MaxItemQtyPerSubPickingTask
						} else {
							newSku.Qty = skuQty
						}
						if newSku.TotalVolume() > waveRule.CommonRule.MaxItemVolumePerSubPickingTask {
							if newSku.SingleVolume() > waveRule.CommonRule.MaxItemVolumePerSubPickingTask {
								solverOrder.Skus = append(solverOrder.Skus, &newSku)
							} else {
								newMaxSkuQty := waveRule.CommonRule.MaxItemVolumePerSubPickingTask / newSku.SingleVolume()
								newSkuQty := newSku.Qty
								for newSkuQty > 0 {
									newNewSku := *solverSku
									if newSkuQty > newMaxSkuQty {
										newNewSku.Qty = newMaxSkuQty
									} else {
										newNewSku.Qty = newSkuQty
									}
									solverOrder.Skus = append(solverOrder.Skus, &newNewSku)
									newSkuQty -= newMaxSkuQty
								}
							}
						} else {
							solverOrder.Skus = append(solverOrder.Skus, &newSku)
						}
						skuQty -= waveRule.CommonRule.MaxItemQtyPerSubPickingTask
					}
				} else if solverSku.TotalVolume() > waveRule.CommonRule.MaxItemVolumePerSubPickingTask {
					solverOrder.isMultiPickerOrder = true
					if solverSku.SingleVolume() > waveRule.CommonRule.MaxItemVolumePerSubPickingTask {
						solverOrder.Skus = append(solverOrder.Skus, solverSku)
					} else {
						maxSkuQty := waveRule.CommonRule.MaxItemVolumePerSubPickingTask / solverSku.SingleVolume()
						skuQty := solverSku.Qty
						for skuQty > 0 {
							newSku := *solverSku
							if skuQty > maxSkuQty {
								newSku.Qty = maxSkuQty
							} else {
								newSku.Qty = skuQty
							}
							solverOrder.Skus = append(solverOrder.Skus, &newSku)
							skuQty -= maxSkuQty
						}
					}
				}
			}

			// mark the picker mode of an order
			if waveRule.WavePickerMode != SinglePickerOnly && !solverOrder.isMultiPickerOrder {
				if len(cross) > 1 || orderTotalVolume > waveRule.CommonRule.MaxItemVolumePerSubPickingTask || orderTotalSkuQty > waveRule.CommonRule.MaxItemQtyPerSubPickingTask {
					solverOrder.isMultiPickerOrder = true
				}
			}

			if convertedOrderNum < maxSolveOrderNum {
				solverGroup.Orders = append(solverGroup.Orders, solverOrder)
			} else {
				orphanOrders = append(orphanOrders, solverOrder)
			}
			convertedOrderNum += 1
		}

		if len(solverGroup.Orders) > 0 {
			solverWave.Groups = append(solverWave.Groups, solverGroup)
		}
	}

	logger.LogInfof("WaveOptAlgo - %s: converted %d orders of %d groups, in which %d are orphans", wave.WaveSn, convertedOrderNum, len(solverWave.Groups), len(orphanOrders))
	return solverWave, orphanOrders, nil
}

// extract splitting level of an order
func extractSplitLevel(location *SolverSkuLocation, level SplitLevel) (string, error) {
	switch level {
	case ByZone:
		return location.ZoneId, nil
	case ByZoneCluster:
		return location.ZoneClusterId, nil
	case ByZoneSector:
		return location.ZoneSectorId, nil
	default:
		return "", fmt.Errorf("unknown split level: %v", level)
	}
}

// validate wave
// non-blank id, non-nil rule and solver config, non-empty groups, SKUs, and locations
func validateWave(wave *Wave) error {
	if len(wave.WaveSn) == 0 {
		return errors.New("blank wave sn")
	}
	if wave.WaveRule == nil {
		return fmt.Errorf("no rule in wave %v", wave.WaveSn)
	}
	if wave.SolverConfig == nil {
		return fmt.Errorf("no solver config in wave %v", wave.SolverConfig)
	}
	if len(wave.Groups) == 0 {
		return fmt.Errorf("empty groups in wave %v", wave.Groups)
	}
	if len(wave.Skus) == 0 {
		return fmt.Errorf("empty SKUs in wave %v", wave.Skus)
	}
	if len(wave.Locations) == 0 {
		return fmt.Errorf("empty locations in wave %v", wave.Locations)
	}

	return nil
}

// validate location
// all fields should be non-blank
func validateLocation(location *Location) error {
	if len(location.LocationId) == 0 {
		return errors.New("blank location id")
	}
	if len(location.ZoneId) == 0 {
		return fmt.Errorf("blank zone id of location %v", location.LocationId)
	}
	if len(location.PathwayId) == 0 {
		return fmt.Errorf("blank pathway id of location %v", location.LocationId)
	}
	if len(location.SegmentId) == 0 {
		return fmt.Errorf("blank segment id of location %v", location.LocationId)
	}

	return nil
}

// validate SKU
// non-blank id, non-negative size and weight
func validateWaveSku(sku *WaveSku) error {
	if len(sku.SkuId) == 0 {
		return errors.New("blank SKU id")
	}
	if sku.Length < 0 {
		return fmt.Errorf("negative length of SKU %v", sku.SkuId)
	}
	if sku.Width < 0 {
		return fmt.Errorf("negative width of SKU %v", sku.SkuId)
	}
	if sku.Height < 0 {
		return fmt.Errorf("negative height of SKU %v", sku.SkuId)
	}
	if sku.Weight < 0 {
		return fmt.Errorf("negative weight of SKU %v", sku.SkuId)
	}

	return nil
}

// validate group
// non-blank id
func validateWaveGroup(group *WaveGroup) error {
	if len(group.GroupSn) == 0 {
		return errors.New("blank group sn")
	}

	return nil
}

// validate order
// non-blank order no
func validateWaveOrder(order *WaveOrder) error {
	if len(order.OrderNo) == 0 {
		return errors.New("blank order no")
	}

	return nil
}

// validate order sku
// non-blank sku id, non-negative qty, non-blank location id
// sku and location ids must be in the provided location and sku maps
func validateOrderSku(sku *OrderSku, orderNo string, skuMap map[string]*WaveSku, locationMap map[string]*SolverSkuLocation) error {
	if len(sku.SkuId) == 0 {
		return fmt.Errorf("blank SKU id in order %v", orderNo)
	}
	if sku.Qty < 0 {
		return fmt.Errorf("negative quantity %v of SKU %v in order %v", sku.Qty, sku.SkuId, orderNo)
	}
	if len(sku.LocationId) == 0 {
		return fmt.Errorf("blank location id of SKU %v in order %v", sku.SkuId, orderNo)
	}
	if _, ok := skuMap[sku.SkuId]; !ok {
		return fmt.Errorf("SKU %v in order %v is now provided in the wave", sku.SkuId, orderNo)
	}
	if _, ok := locationMap[sku.LocationId]; !ok {
		return fmt.Errorf("location %v of SKU %v in order %v is not provided in the wave", sku.LocationId, sku.SkuId, orderNo)
	}

	return nil
}

func detectConflictsWithinRules(rule *WaveRule) error {
	if rule == nil {
		return fmt.Errorf("wave has no rules")
	}

	var err error

	// validate the common rules
	commonRule := rule.CommonRule
	err = validateBoundConstraint(float64(commonRule.MaxPickingTaskQtyPerWave), float64(commonRule.MinPickingTaskQtyPerWave), "PickingTaskQtyPerWave")
	if err != nil {
		return err
	}
	err = validateBoundConstraint(float64(commonRule.MaxOrderQtyPerPickingTask), float64(commonRule.MinOrderQtyPerPickingTask), "OrderQtyPerPickingTask")
	if err != nil {
		return err
	}
	err = validateBoundConstraint(float64(commonRule.MaxItemQtyPerSubPickingTask), float64(commonRule.MinItemQtyPerSubPickingTask), "ItemQtyPerSubPickingTask")
	if err != nil {
		return err
	}
	err = validateBoundConstraint(float64(commonRule.MaxItemVolumePerSubPickingTask), float64(commonRule.MinItemVolumePerSubPickingTask), "ItemVolumePerSubPickingTask")
	if err != nil {
		return err
	}
	err = validateBoundConstraint(float64(commonRule.MaxBulkyTaskLoad), float64(commonRule.MaxNonBulkyTaskLoad), "BulkyTaskLoadThreshold")
	if err != nil {
		return err
	}
	err = validateBoundConstraint(float64(commonRule.MaxBulkyTaskVolume), float64(commonRule.MaxNonBulkyTaskVolume), "BulkyTaskVolumeThreshold")
	if err != nil {
		return err
	}

	// validate mode rules
	if rule.WavePickerMode == SinglePickerOnly {
		// single picker mode does not have mode rules
	} else if rule.WavePickerMode == MultiPickerAtMWSWithTotalOrderQty {
		modeRule, ok := rule.ModeRule.(*MultiPickerAtMWSTotalQtyRule)
		if !ok {
			return fmt.Errorf("picker mode %v has no detailed rules", MultiPickerAtMWSWithTotalOrderQty)
		}

		if modeRule.MaxBacklogAtMWSPerWave < 0 {
			return fmt.Errorf("max backlog at MWS per wave is %v, should be positive", modeRule.MaxBacklogAtMWSPerWave)
		}
	} else if rule.WavePickerMode == MultiPickerAtMWSWithRespectiveOrderQty {
		modeRule, ok := rule.ModeRule.(*MultiPickerAtMWSRespectiveQtyRule)
		if !ok {
			return fmt.Errorf("picker mode %v has no detailed rules", MultiPickerAtMWSWithRespectiveOrderQty)
		}

		if modeRule.MaxBulkyBacklogAtMWSPerWave < 0 {
			return fmt.Errorf("max bulky backlog at MWS per wave is %v, should be positive", modeRule.MaxBulkyBacklogAtMWSPerWave)
		}
		if modeRule.MaxNonBulkyBacklogAtMWSPerWave < 0 {
			return fmt.Errorf("max non-bulky backlog at MWS per wave is %v, should be positive", modeRule.MaxNonBulkyBacklogAtMWSPerWave)
		}
	} else if rule.WavePickerMode == MultiPickerAtMLWithTotalPickingTaskQty {
		modeRule, ok := rule.ModeRule.(*MultiPickerAtMLTotalQtyRule)
		if !ok {
			return fmt.Errorf("picker mode %v has no detailed rules", MultiPickerAtMLWithTotalPickingTaskQty)
		}

		if modeRule.MaxBacklogAtMLPerWave < 0 {
			return fmt.Errorf("max backlog at ML per wave is %v, should be positive", modeRule.MaxBacklogAtMLPerWave)
		}
	} else if rule.WavePickerMode == MultiPickerAtMLWithRespectivePickingTaskQty {
		modeRule, ok := rule.ModeRule.(*MultiPickerAtMLRespectiveQtyRule)
		if !ok {
			return fmt.Errorf("picker mode %v has no detailed rules", MultiPickerAtMLWithRespectivePickingTaskQty)
		}

		if modeRule.MaxBulkyBacklogAtMLPerWave < 0 {
			return fmt.Errorf("max bulky backlog at ML per wave is %v, should be positive", modeRule.MaxBulkyBacklogAtMLPerWave)
		}
		if modeRule.MaxNonBulkyBacklogAtMLPerWave < 0 {
			return fmt.Errorf("max non-bulky backlog at ML per wave is %v, should be positive", modeRule.MaxNonBulkyBacklogAtMLPerWave)
		}
	}

	return nil
}

func detectConflictsBetweenWaveAndRules(solverWave *SolverWave, rule *WaveRule) error {
	commonRule := rule.CommonRule

	// total number of orders provided should be greater or equal to the product of minimal number of tasks and minimal order number of a task
	totalOrderQty := int64(0)
	for _, group := range solverWave.Groups {
		totalOrderQty += int64(len(group.Orders))
	}

	minRequiredOrderQty := commonRule.MinOrderQtyPerPickingTask * commonRule.MinPickingTaskQtyPerWave

	if totalOrderQty < minRequiredOrderQty {
		return fmt.Errorf("wave %v has %v orders, but at least %v is required to meet the rule", solverWave.Id, totalOrderQty, minRequiredOrderQty)
	}

	// total number (volume) of items should be greater or equal to the minimal item number (volume) of a subtask
	totalItemQty := int64(0)
	totalItemVolume := int64(0)
	for _, group := range solverWave.Groups {
		for _, order := range group.Orders {
			for _, sku := range order.Skus {
				totalItemQty += sku.Qty
				totalItemVolume += sku.TotalVolume()
			}
		}
	}

	minRequiredItemQty := commonRule.MinItemQtyPerSubPickingTask
	minRequiredItemVolume := commonRule.MinItemVolumePerSubPickingTask

	if totalItemQty < minRequiredItemQty {
		return fmt.Errorf("wave %v has %v items, but at least %v is required to meet the rule", solverWave.Id, totalItemQty, minRequiredItemQty)
	}

	if totalItemVolume < minRequiredItemVolume {
		return fmt.Errorf("wave %v has item volume of %v, but at least %v is required to meet the rule", solverWave.Id, totalItemVolume, minRequiredItemVolume)
	}

	return nil
}

func validateBoundConstraint(max float64, min float64, name string) error {
	if min < 0 {
		return fmt.Errorf("negative lower bound of %v", name)
	}
	if max < min {
		return fmt.Errorf("lower bound of %v is higher then the upper bound", name)
	}
	if max <= 0 {
		return fmt.Errorf("non-positive upper bound of %v", name)
	}

	return nil
}
