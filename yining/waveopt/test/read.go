package test

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"git.garena.com/shopee/bg-logistics/algo/wms/waveopt/waveoptsolver"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
)

type RawRule struct {
	WaveType                   int64 `json:"wave_type"`
	MinSkuQty                  int64 `json:"min_sku_qty"`
	MinItemQty                 int64 `json:"min_item_qty"`
	MinTaskQty                 int64 `json:"min_wave_pick_quantity"`
	MaxTaskQty                 int64 `json:"max_wave_pick_quantity"`
	MaxOrderQtyPerTask         int64 `json:"max_per_pick_order_quantity"`
	MinOrderQtyPerTask         int64 `json:"min_per_pick_order_quantity"`
	MaxPerPickSkuQuantity      int64 `json:"max_per_pick_sku_quantity"`
	MaxPerBulkyPickItemQty     int64 `json:"max_per_bulky_pick_item_quantity"`
	MaxPerNonBulkyItemPickList int64 `json:"max_per_non_bulky_item_pick_list"`
	CrossClusterFlag           int64 `json:"cross_zone_group_flag"`
	MultiPickerFlag            int64 `json:"allow_multi_picker"`
	SplitLevelFlag             int64 `json:"split_type"`
	MergeAtFlag                int64 `json:"merge_at"`
	TotalBacklogTaskQty        int64 `json:"max_generate_picking_task_num"`
	BulkyBacklogTaskQty        int64 `json:"max_generate_bulky_picking_task_num"`
	NonBulkyBacklogTaskQty     int64 `json:"max_generate_non_bulky_picking_task_num"`
	TotalBacklogOrderQty       int64 `json:"max_generate_wave_order_num"`
	BulkyBacklogOrderQty       int64 `json:"max_generate_wave_bulky_order_num"`
	NonBulkyBacklogOrderQty    int64 `json:"max_generate_wave_non_bulky_order_num"`
	MaxNonBulkyVolume          int64 `json:"non_bulk_max_volume"`
	MaxBulkyVolume             int64 `json:"bulk_max_volume"`
	MaxNonBulkyLoad            int64 `json:"non_bulk_max_load"`
	MaxBulkyLoad               int64 `json:"bulk_max_load"`
	BacklogThresholdType       int64 `json:"backlog_threshold_type"`
	BulkyBacklogThreshold      int64 `json:"bulky_backlog_threshold"`
	NonBulkyBacklogThreshold   int64 `json:"non_bulky_backlog_threshold"`
	TotalBacklogThreshold      int64 `json:"total_backlog_threshold"`

	MaxGridNo                     int64 `json:"max_grid_no"`
	HoldingBulkyPickingTaskNum    int64 `json:"holding_bulky_picking_task_num"`
	HoldingNonBulkyPickingTaskNum int64 `json:"holding_non_bulky_picking_task_num"`
	HoldingBulkyOrder             int64 `json:"holding_bulky_order "`
	HoldingNonBulkyOrder          int64 `json:"holding_non_bulky_order"`
}

type RawWave struct {
	WaveSn    string
	OrderSkus []*RawOrderSku
}

type RawOrderSku struct {
	OrderNo       string `json:"order_no"`
	SortFactor    int64  `json:"sort_factor"`
	SkuId         string `json:"sku_id"`
	SkuQty        int64  `json:"qty"`
	LocationId    string `json:"location_id"`
	GroupTag      int64  `json:"group_tag"`
	CutOffTime    int64  `json:"cut_off_time"`
	ZoneId        string `json:"zoneid"`
	PathwayId     string `json:"pathwayid"`
	SegmentId     string `json:"segmentid"`
	ZoneClusterId int64  `json:"zone_group_id"`
	ZoneSectorId  int64  `json:"zone_sector_id"`
	Length        int64  `json:"length"`
	Width         int64  `json:"width"`
	Height        int64  `json:"height"`
	Weight        int64  `json:"weight"`
	IsBulky       int64  `json:"is_bulk"`
	IsExtraBulky  int64  `json:"is_extra_bulk"`
}

func readOneWave(fileName string, selectedWaveSn string, line int) (*RawRule, *RawWave, error) {
	var err error

	f, err := os.Open(fileName)
	if err != nil {
		log.Panicf("cannot open file %v", fileName)
	}
	defer f.Close()

	reader := csv.NewReader(f)
	// skip the header
	_, err = reader.Read()
	if err != nil {
		return nil, nil, err
	}

	var data []string
	if len(selectedWaveSn) == 0 {
		for i := 0; i < line-1; i++ {
			_, err = reader.Read()
			//if data[0] == selectedWaveSn {
			//	break
			//}
			if err != nil {
				return nil, nil, err
			}
		}
		data, err = reader.Read()
		if err != nil {
			return nil, nil, err
		}
	} else {
		for i := 0; i < math.MaxInt64; i++ {
			data, err = reader.Read()
			if data[0] == selectedWaveSn {
				break
			}
			if err != nil {
				return nil, nil, err
			}
		}
		if err != nil {
			return nil, nil, err
		}
	}

	waveSn := data[0]
	log.Printf("read wave: %v", waveSn)

	ruleJson := []byte(data[1])
	var rawRules []*RawRule
	err = json.Unmarshal(ruleJson, &rawRules)
	if err != nil {
		return nil, nil, err
	}

	waveJson := data[2]
	var rawOrders []*RawOrderSku
	err = json.Unmarshal([]byte(waveJson), &rawOrders)
	if err != nil {
		return nil, nil, err
	}

	rawWave := &RawWave{
		WaveSn:    waveSn,
		OrderSkus: rawOrders,
	}

	return rawRules[0], rawWave, nil
}

func convertRawWaveToWave(rawWave *RawWave, rawRule *RawRule, sequenceMap map[string]int64, coordinateMap map[string][]float64) (*waveoptsolver.Wave, error) {
	// convert wave rule
	commonRule := &waveoptsolver.CommonRule{
		MaxPickingTaskQtyPerWave:  rawRule.MaxTaskQty,
		MinPickingTaskQtyPerWave:  rawRule.MinTaskQty,
		IsCrossZoneCluster:        rawRule.CrossClusterFlag == 1,
		MaxOrderQtyPerPickingTask: rawRule.MaxOrderQtyPerTask,
		MinOrderQtyPerPickingTask: 1,
		//MinOrderQtyPerPickingTask:      rawRule.MinOrderQtyPerTask,
		MaxItemQtyPerSubPickingTask:    10000,
		MinItemQtyPerSubPickingTask:    1,
		MaxItemVolumePerSubPickingTask: 1e10,
		MinItemVolumePerSubPickingTask: 1,
		MaxNonBulkyTaskVolume:          1e6,
		MaxNonBulkyTaskLoad:            1e3,
		MaxBulkyTaskVolume:             1e8,
		MaxBulkyTaskLoad:               1e5,
	}

	var pickerMode waveoptsolver.WavePickerMode
	if rawRule.MultiPickerFlag == 1 {
		isAtMWS := false
		if rawRule.MergeAtFlag == 2 {
			isAtMWS = true
		}

		isTotal := false
		if rawRule.TotalBacklogTaskQty > 0 || rawRule.TotalBacklogOrderQty > 0 {
			isTotal = true
		}

		if isAtMWS && isTotal {
			pickerMode = waveoptsolver.MultiPickerAtMWSWithTotalOrderQty
		} else if isAtMWS && !isTotal {
			pickerMode = waveoptsolver.MultiPickerAtMWSWithRespectiveOrderQty
		} else if !isAtMWS && isTotal {
			pickerMode = waveoptsolver.MultiPickerAtMLWithTotalPickingTaskQty
		} else {
			pickerMode = waveoptsolver.MultiPickerAtMLWithRespectivePickingTaskQty
		}
	} else {
		pickerMode = waveoptsolver.SinglePickerOnly
	}

	var splitLevel waveoptsolver.SplitLevel
	switch rawRule.SplitLevelFlag {
	case 0:
		splitLevel = waveoptsolver.ByZone
	case 1:
		splitLevel = waveoptsolver.ByZoneCluster
	case 2:
		splitLevel = waveoptsolver.ByZoneSector
	}

	var modeRule waveoptsolver.SplitLevelProvider
	switch pickerMode {
	case waveoptsolver.SinglePickerOnly:
		modeRule = nil
	case waveoptsolver.MultiPickerAtMWSWithTotalOrderQty:
		modeRule = &waveoptsolver.MultiPickerAtMWSTotalQtyRule{
			MaxBacklogAtMWSPerWave: rawRule.TotalBacklogOrderQty,
			PickingTaskSplitLevel:  splitLevel,
		}
	case waveoptsolver.MultiPickerAtMWSWithRespectiveOrderQty:
		modeRule = &waveoptsolver.MultiPickerAtMWSRespectiveQtyRule{
			MaxBulkyBacklogAtMWSPerWave:    rawRule.BulkyBacklogOrderQty,
			MaxNonBulkyBacklogAtMWSPerWave: rawRule.NonBulkyBacklogOrderQty,
			PickingTaskSplitLevel:          splitLevel,
		}
	case waveoptsolver.MultiPickerAtMLWithTotalPickingTaskQty:
		var backlogTaskQty int64
		if rawRule.TotalBacklogTaskQty <= 0 {
			backlogTaskQty = 1
		} else {
			backlogTaskQty = rawRule.TotalBacklogTaskQty
		}
		modeRule = &waveoptsolver.MultiPickerAtMLTotalQtyRule{
			MaxBacklogAtMLPerWave: backlogTaskQty,
			PickingTaskSplitLevel: splitLevel,
		}
	case waveoptsolver.MultiPickerAtMLWithRespectivePickingTaskQty:
		modeRule = &waveoptsolver.MultiPickerAtMLRespectiveQtyRule{
			MaxBulkyBacklogAtMLPerWave:    rawRule.BulkyBacklogTaskQty,
			MaxNonBulkyBacklogAtMLPerWave: rawRule.NonBulkyBacklogTaskQty,
			PickingTaskSplitLevel:         splitLevel,
		}
	}

	rule := &waveoptsolver.WaveRule{
		CommonRule:     commonRule,
		WavePickerMode: pickerMode,
		ModeRule:       modeRule,
	}

	groupMap := make(map[waveoptsolver.GroupType]*waveoptsolver.WaveGroup)
	orderMap := make(map[string]*waveoptsolver.WaveOrder)
	waveSkuMap := make(map[string]*waveoptsolver.WaveSku)
	locationMap := make(map[string]*waveoptsolver.Location)
	for _, rawOrder := range rawWave.OrderSkus {
		groupType := getGroupType(rawOrder.GroupTag)
		var group *waveoptsolver.WaveGroup
		if group_, ok := groupMap[groupType]; ok {
			group = group_
		} else {
			grp := &waveoptsolver.WaveGroup{
				GroupSn:   "",
				GroupType: groupType,
				Priority:  0,
				Orders:    make([]*waveoptsolver.WaveOrder, 0),
			}
			group = grp
			groupMap[groupType] = group
		}

		orderSku := &waveoptsolver.OrderSku{
			OrderNo:    rawOrder.OrderNo,
			SkuId:      rawOrder.SkuId,
			Qty:        rawOrder.SkuQty,
			LocationId: rawOrder.LocationId,
			//PickingSequence :
		}

		if _, ok := waveSkuMap[rawOrder.SkuId]; !ok {
			waveSku := &waveoptsolver.WaveSku{
				SkuId:   rawOrder.SkuId,
				Length:  rawOrder.Length,
				Width:   rawOrder.Width,
				Height:  rawOrder.Height,
				Weight:  rawOrder.Weight,
				SkuSize: waveoptsolver.NonBulkySkuType,
			}
			waveSkuMap[waveSku.SkuId] = waveSku
		}

		if _, ok := locationMap[rawOrder.LocationId]; !ok {
			location := &waveoptsolver.Location{
				LocationId:      rawOrder.LocationId,
				ZoneSectorId:    rawOrder.ZoneSectorId,
				ZoneClusterId:   rawOrder.ZoneClusterId,
				ZoneId:          rawOrder.ZoneId,
				PathwayId:       rawOrder.PathwayId,
				SegmentId:       rawOrder.SegmentId,
				PickingSequence: extractPickingSequence(rawOrder.LocationId, sequenceMap),
			}
			locationMap[location.LocationId] = location
		}

		var order *waveoptsolver.WaveOrder
		if odr, ok := orderMap[rawOrder.OrderNo]; ok {
			order = odr
		} else {
			order = &waveoptsolver.WaveOrder{
				OrderNo:    rawOrder.OrderNo,
				IsUrgent:   false,
				CutOffTime: rawOrder.CutOffTime,
				Skus:       make([]*waveoptsolver.OrderSku, 0),
			}
			orderMap[order.OrderNo] = order
			group.Orders = append(group.Orders, order)
		}
		order.Skus = append(order.Skus, orderSku)
	}

	groups := getGroups(groupMap, rawRule.WaveType)

	skus := make([]*waveoptsolver.WaveSku, 0)
	for _, sku := range waveSkuMap {
		skus = append(skus, sku)
	}

	locations := make([]*waveoptsolver.Location, 0)
	for _, location := range locationMap {
		locations = append(locations, location)
	}

	wave := &waveoptsolver.Wave{
		WaveSn:       rawWave.WaveSn,
		WaveRule:     rule,
		SolverConfig: nil, // solver configuration is specified in the test method
		Groups:       groups,
		Skus:         skus,
		Locations:    locations,
	}

	return wave, nil
}

func getGroupType(groupTag int64) waveoptsolver.GroupType {
	switch groupTag {
	case 0:
		return waveoptsolver.ExtraBulkyPickingOrderType
	case 1:
		return waveoptsolver.HighValuePickingOrderType
	case 2:
		return waveoptsolver.StorageHighValuePickingOrderType
	case 3:
		return waveoptsolver.BulkyPickingOrderType
	case 4:
		return waveoptsolver.StorageBulkyPickingOrderType
	case 5:
		return waveoptsolver.NormalPickingOrderType
	case 6:
		return waveoptsolver.StorageNormalPickingOrderType
	case 7:
		return waveoptsolver.MedicalPickingOrderType
	}

	panic(fmt.Sprintf("illegal group tag: %v", groupTag))
}

// getGroups turns raw groups to real groups
// waveType: 	1 - single sku single qty
//				2 - same sku same qty
//				3 - mixed wave
//				4 - single sku any qty
//				5 - mixed sku single qty
func getGroups(groupMap map[waveoptsolver.GroupType]*waveoptsolver.WaveGroup, waveType int64) []*waveoptsolver.WaveGroup {
	groups := make([]*waveoptsolver.WaveGroup, 0)
	for _, group := range groupMap {
		var newGroups []*waveoptsolver.WaveGroup
		if waveType == 1 || waveType == 4 {
			newGroups = splitGroupBySku(group)
		} else if waveType == 2 {
			newGroups = splitGroupBySkuAndQty(group)
		} else {
			group.GroupSn = fmt.Sprintf("Group-%v-0", group.GroupType)
			newGroups = append(make([]*waveoptsolver.WaveGroup, 0), group)
		}
		groups = append(groups, newGroups...)
	}

	return groups
}

// splitGroupBySku clusters the same sku into the same group, and returns the split groups.
// apply this method to single sku single qty groups.
// name each group.
func splitGroupBySku(group *waveoptsolver.WaveGroup) []*waveoptsolver.WaveGroup {
	skuMap := make(map[string][]*waveoptsolver.WaveOrder)
	for _, order := range group.Orders {
		if len(order.Skus) != 1 {
			panic("order does not has exactly one SKU")
		}

		sku := order.Skus[0]
		if sku.Qty != 1 {
			panic("sku does not has exactly one quantity")
		}

		if _, ok := skuMap[sku.SkuId]; !ok {
			skuMap[sku.SkuId] = make([]*waveoptsolver.WaveOrder, 0)
		}
		orders := skuMap[sku.SkuId]
		orders = append(orders, order)
		skuMap[sku.SkuId] = orders
	}

	if len(skuMap) <= 0 {
		panic(fmt.Sprintf("cannot split group %v", group.GroupSn))
	}

	groups := make([]*waveoptsolver.WaveGroup, 0)
	groupIdx := 0
	for _, orders := range skuMap {
		newGroup := &waveoptsolver.WaveGroup{
			GroupSn:   fmt.Sprintf("Group-%v-%v", group.GroupType, groupIdx),
			GroupType: group.GroupType,
			Priority:  group.Priority,
			Orders:    orders,
		}
		groups = append(groups, newGroup)
		groupIdx++
	}

	return groups
}

type skuKey string

func getSkuKey(sku *waveoptsolver.OrderSku) skuKey {
	return skuKey(sku.SkuId + "_" + strconv.FormatInt(sku.Qty, 10))
}

func getOrderSkuKey(order *waveoptsolver.WaveOrder) skuKey {
	skuKeys := make([]string, 0)
	for _, sku := range order.Skus {
		key := getSkuKey(sku)
		skuKeys = append(skuKeys, string(key))
	}

	return skuKey(strings.Join(skuKeys, "_"))
}

func splitGroupBySkuAndQty(group *waveoptsolver.WaveGroup) []*waveoptsolver.WaveGroup {
	skuMap := make(map[skuKey][]*waveoptsolver.WaveOrder)
	for _, order := range group.Orders {
		key := getOrderSkuKey(order)
		if _, ok := skuMap[key]; !ok {
			skuMap[key] = make([]*waveoptsolver.WaveOrder, 0)
		}
		orders := skuMap[key]
		orders = append(orders, order)
		skuMap[key] = orders
	}

	if len(skuMap) <= 0 {
		panic("cannot split group")
	}

	groups := make([]*waveoptsolver.WaveGroup, 0)
	groupIdx := 0
	for _, orders := range skuMap {
		newGroup := &waveoptsolver.WaveGroup{
			GroupSn:   fmt.Sprintf("Group-%v-%v", group.GroupType, groupIdx),
			GroupType: group.GroupType,
			Priority:  group.Priority,
			Orders:    orders,
		}
		groups = append(groups, newGroup)
		groupIdx++
	}

	return groups
}

func writeWaveResult(waveResult *waveoptsolver.WaveResult) [][]string {

	csvResult := make([][]string, 0)
	if waveResult.RetCode != 0 {
		rowResult := make([]string, 0)
		rowResult = append(rowResult,
			waveResult.WaveSn,
			strconv.FormatInt(waveResult.RetCode, 10),
			"",
			"",
			"",
			"",
			"",
			"",
			waveResult.Msg)
		csvResult = append(csvResult, rowResult)
	}

	for ptIdx, pt := range waveResult.PickingTasks {
		for subIdx, subPt := range pt.SubPickingTasks {
			for orderIdx, sku := range subPt.Skus {
				rowResult := make([]string, 0)
				rowResult = append(rowResult,
					waveResult.WaveSn,
					strconv.FormatInt(waveResult.RetCode, 10),
					strconv.FormatInt(int64(ptIdx), 10),
					strconv.FormatInt(int64(subIdx), 10),
					strconv.FormatInt(int64(orderIdx), 10),
					sku.OrderNo,
					sku.LocationId,
					strconv.FormatInt(sku.Qty, 10),
					waveResult.Msg,
				)
				csvResult = append(csvResult, rowResult)
			}
		}
	}

	return csvResult

}

func writeWaveResultV1(waveResult *waveoptsolver.WaveResult) [][]string {

	csvResult := make([][]string, 0)
	if waveResult.RetCode != 0 {
		rowResult := make([]string, 0)
		rowResult = append(rowResult,
			waveResult.WaveSn,
			strconv.FormatInt(waveResult.RetCode, 10),
			"",
			"",
			"",
			"",
			"",
			"",
			"")
		csvResult = append(csvResult, rowResult)
	}

	for ptIdx, pt := range waveResult.PickingTasks {
		for subIdx, subPt := range pt.SubPickingTasks {
			for orderIdx, sku := range subPt.Skus {
				rowResult := make([]string, 0)
				rowResult = append(rowResult,
					waveResult.WaveSn,
					strconv.FormatInt(waveResult.RetCode, 10),
					strconv.FormatInt(int64(ptIdx), 10),
					strconv.FormatInt(int64(subIdx), 10),
					strconv.FormatInt(int64(orderIdx), 10),
					sku.OrderNo,
					sku.LocationId,
					strconv.FormatInt(sku.Qty, 10),
					fmt.Sprint(pt.OrderNos),
				)
				csvResult = append(csvResult, rowResult)
			}
		}
	}

	return csvResult

}

func readBulkyMap(wave *RawWave) (map[string]int64, map[string]int64, error) {

	extraBulkyOrderMap := make(map[string]int64, 0)
	bulkyOrderMap := make(map[string]int64, 0)

	for _, order := range wave.OrderSkus {
		if order.IsExtraBulky == 1 {
			extraBulkyOrderMap[order.OrderNo] = 1
		}
		if order.IsBulky == 1 {
			bulkyOrderMap[order.OrderNo] = 1
		}
	}

	return extraBulkyOrderMap, bulkyOrderMap, nil
}

func readSortFactor(wave *RawWave) map[string]int64 {
	sortFactor := make(map[string]int64)
	for _, order := range wave.OrderSkus {
		sortFactor[order.OrderNo] = order.SortFactor
	}
	return sortFactor
}

func readSequence(fileName string) (map[string]int64, error) {
	var err error

	f, err := os.Open(fileName)
	if err != nil {
		log.Panicf("cannot open file %v", fileName)
	}
	defer f.Close()

	reader := csv.NewReader(f)
	// skip the header
	data, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	sequenceMap := make(map[string]int64)
	for ind, row := range data {
		if ind == 0 {
			continue
		} else {
			sequenceMap[row[0]], _ = strconv.ParseInt(row[1], 10, 64)
		}
	}
	return sequenceMap, err
}

func readLocation(fileName string) (map[string][]float64, error) {
	var err error

	f, err := os.Open(fileName)
	if err != nil {
		log.Panicf("cannot open file %v", fileName)
	}
	defer f.Close()

	reader := csv.NewReader(f)
	// skip the header
	data, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	locationMap := make(map[string][]float64)
	for ind, row := range data {
		if ind == 0 {
			continue
		} else {
			tmpX, _ := strconv.ParseFloat(row[3], 64)
			tmpY, _ := strconv.ParseFloat(row[4], 64)
			locationMap["IDL"+"-"+row[0]+"-"+row[1]+"-"+row[2]] = []float64{
				tmpX, tmpY,
			}
		}
	}
	return locationMap, err
}

func extractPickingSequence(locationId string, sequenceMap map[string]int64) int64 {
	splitLocation := strings.Split(locationId, "-")
	var newLocation string
	for i, tmp := range splitLocation {
		if i == 0 {
			newLocation += tmp
		}
		if i > 0 && i < 3 {
			newLocation += "-" + tmp
		}
		if i == 3 {
			newLocation += "-" + tmp
			break
		}
	}
	pickingSequence, ok := sequenceMap[newLocation]
	if !ok {
		pickingSequence = 1e10
	}
	return pickingSequence
}

func extractCoordinate(locationId string, coordinateMap map[string][]float64) []float64 {
	splitLocation := strings.Split(locationId, "-")
	var newLocation string
	for i, tmp := range splitLocation {
		if i == 0 {
			newLocation += tmp
		}
		if i > 0 && i < 3 {
			newLocation += "-" + tmp
		}
		if i == 3 {
			newLocation += "-" + tmp
			break
		}
	}
	pickingSequence := coordinateMap[newLocation]
	//if !ok {
	//	pickingSequence = [0.0,0.0]
	//}
	return pickingSequence
}
