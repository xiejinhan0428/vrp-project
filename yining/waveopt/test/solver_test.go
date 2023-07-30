package test

import (
	"encoding/csv"
	"git.garena.com/shopee/bg-logistics/algo/wms/waveopt/waveoptsolver"
	"os"
	"strconv"
	"testing"
)

func TestSolveAWave(t *testing.T) {
	runningTimes := []int64{
		120,
		//300,
	}

	for j := 0; j < 1; j++ {
		runningTime := runningTimes[j]

		// 创建文件
		f, err := os.OpenFile("result_v2_"+strconv.FormatInt(runningTime, 10)+"_0003_test.csv", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			panic(err)
		}
		defer f.Close()

		// 写入表头
		csvResult := make([][]string, 0)
		header := []string{
			"wave_sn", "return_code", "pt_idx", "sub_idx", "order_idx", "order_no", "location_id", "sku_qty", "message",
		}
		csvResult = append(csvResult, header)
		w := csv.NewWriter(f) //创建一个新的写入文件流
		w.WriteAll(csvResult) //写入数据
		w.Flush()

		for i := 3; i <= 3; i++ {
			rawRule, rawWave, _ := readOneWave("test_demo_idl.csv", "", i)
			rawRule.MaxTaskQty = 1000000
			rawRule.TotalBacklogOrderQty = 1e8
			rawRule.BulkyBacklogOrderQty = 1e8
			rawRule.NonBulkyBacklogOrderQty = 1e8
			rawRule.TotalBacklogTaskQty = 1e8
			rawRule.BulkyBacklogTaskQty = 1e8
			rawRule.NonBulkyBacklogTaskQty = 1e8

			wave, _ := convertRawWaveToWave(rawWave, rawRule, make(map[string]int64), make(map[string][]float64))

			wave.WaveRule.CommonRule.MaxItemQtyPerSubPickingTask = 100000000

			solverConfig := &waveoptsolver.WaveSolverConfig{
				MaxSecondsSpent:    runningTime,
				Parallelism:        3,
				VariableTabuTenure: 5,
				ValueTabuTenure:    5,
				ZoneCoeff:          100,
				PathwayCoeff:       10,
				SegmentCoeff:       1,
			}
			wave.SolverConfig = solverConfig
			//wave.Groups[0].Orders = wave.Groups[0].Orders[:100]

			waveOptSolver := &waveoptsolver.WavePickingTaskOptSolver{Wave: wave}
			//wave.WaveRule.WavePickerMode = waveoptsolver.SinglePickerOnly
			result := waveOptSolver.GeneratePickingTasks()

			w := csv.NewWriter(f) //创建一个新的写入文件流

			csvResult := writeWaveResult(result)

			w.WriteAll(csvResult) //写入数据
			w.Flush()

			if result.RetCode != waveoptsolver.SuccessResult {
				t.Errorf("RetCode: %d, Msg: %s", result.RetCode, result.Msg)
			}
		}
	}
}

//func TestSolveAWaveV1(t *testing.T) {
//	// 创建文件
//	f, err := os.Create("result_v1.csv")
//	if err != nil {
//		panic(err)
//	}
//	defer f.Close()
//
//	// 写入表头
//	csvResult := make([][]string, 0)
//	header := []string{
//		"wave_sn", "return_code", "pt_idx", "sub_idx", "order_idx", "order_no", "location_id", "sku_qty",
//	}
//	csvResult = append(csvResult, header)
//	w := csv.NewWriter(f) //创建一个新的写入文件流
//	w.WriteAll(csvResult) //写入数据
//	w.Flush()
//
//	for i := 1; i <= 1; i++ {
//
//		rawRule, rawWave, _ := readOneWave("test_demo_idl.csv", "", i)
//		rawRule.MaxTaskQty = 1000000
//		wave, _ := convertRawWaveToWave(rawWave, rawRule, make(map[string]int64), make(map[string][]float64))
//
//		sortFactor := readSortFactor(rawWave)
//		extraBulkyOrder, BulkyOrder, _ := readBulkyMap(rawWave)
//
//		waveCacheData := &wavev1script.WaveCacheData{
//			MaxPerPickSkuQuantity:       99,
//			MaxPerBulkyPickItemQuantity: rawRule.MaxPerBulkyPickItemQty,
//			MaxPerNonBulkyItemPickList:  rawRule.MaxPerNonBulkyItemPickList,
//			MultiPickerRule: &wavev1script.MultiPickerRuleTab{
//				ID:                            1,
//				WhsID:                         "whs001",
//				AllowMultiPicker:              rawRule.MultiPickerFlag,
//				SplitType:                     rawRule.SplitLevelFlag,
//				MinSkuQty:                     rawRule.MinSkuQty,
//				MinItemQty:                    rawRule.MinItemQty,
//				StagingType:                   0,
//				Sectors:                       "",
//				Clusters:                      "",
//				Zones:                         "",
//				Operator:                      "",
//				MergeAt:                       rawRule.MergeAtFlag,
//				BacklogThresholdType:          rawRule.BacklogThresholdType,
//				TotalBacklogThreshold:         rawRule.TotalBacklogThreshold,
//				BulkyBacklogThreshold:         rawRule.BulkyBacklogThreshold,
//				NonBulkyBacklogThreshold:      rawRule.NonBulkyBacklogThreshold,
//				AssignmentPriorityRestriction: 0,
//				Ctime:                         0,
//				Mtime:                         0,
//			},
//			HoldingTaskAttr: &wavev1script.HoldingTaskAttr{
//				HoldingTaskEnable:    true,
//				MaxStationGrid:       rawRule.MaxGridNo,
//				BulkyHoldingCount:    rawRule.HoldingBulkyPickingTaskNum,
//				NonBulkyHoldingCount: rawRule.HoldingNonBulkyPickingTaskNum,
//				BulkyOrderNum:        rawRule.HoldingBulkyOrder,
//				NonBulkyOrderNum:     rawRule.HoldingNonBulkyOrder,
//				FreeGridSum:          1000000,
//			},
//			OrderSortFactorMap: sortFactor,
//			ExtraBulkyOrderMap: extraBulkyOrder,
//			BulkyOrderMap:      BulkyOrder,
//		}
//
//		result := wavev1script.RunWaveV1Calculate(wave, waveCacheData)
//
//		locationMap := make(map[string]*waveoptsolver.Location)
//		for _, location := range wave.Locations {
//			locationMap[location.LocationId] = location
//		}
//
//		crossZoneNumMap := make(map[int]int)
//		for _, task := range result.PickingTasks {
//			for _, subTask := range task.SubPickingTasks {
//				clusterSet := make(map[int64]struct{})
//				for _, sku := range subTask.Skus {
//					clusterId := locationMap[sku.LocationId].ZoneClusterId
//					clusterSet[clusterId] = struct{}{}
//				}
//
//				clusterNum := len(clusterSet)
//				if _, ok := crossZoneNumMap[clusterNum]; !ok {
//					crossZoneNumMap[clusterNum] = 0
//				}
//				oldNum := crossZoneNumMap[clusterNum]
//				crossZoneNumMap[clusterNum] = oldNum + 1
//			}
//		}
//
//		scr := 0
//		for _, task := range result.PickingTasks {
//			for _, subTask := range task.SubPickingTasks {
//				zoneSet := make(map[string]struct{})
//				pathwaySet := make(map[string]struct{})
//				segmentSet := make(map[string]struct{})
//				for _, sku := range subTask.Skus {
//					skuLocation := locationMap[sku.LocationId]
//					zoneSet[skuLocation.ZoneId] = struct{}{}
//					pathwaySet[skuLocation.ZoneId] = struct{}{}
//					segmentSet[skuLocation.ZoneId] = struct{}{}
//				}
//				scr += 100*len(zoneSet) + 10*len(pathwaySet) + len(segmentSet)
//			}
//		}
//
//		w := csv.NewWriter(f) //创建一个新的写入文件流
//
//		csvResult := writeWaveResult(result)
//
//		w.WriteAll(csvResult) //写入数据
//		w.Flush()
//	}
//}

func TestOneWaveWithSn(t *testing.T) {
	rawRule, rawWave, _ := readOneWave("test_demo_idl.csv", "", 59)
	rawRule.MaxTaskQty = 1000000
	rawRule.TotalBacklogOrderQty = 1e8
	rawRule.BulkyBacklogOrderQty = 1e8
	rawRule.NonBulkyBacklogOrderQty = 1e8
	rawRule.TotalBacklogTaskQty = 1e8
	rawRule.BulkyBacklogTaskQty = 1e8
	rawRule.NonBulkyBacklogTaskQty = 1e8

	wave, _ := convertRawWaveToWave(rawWave, rawRule, make(map[string]int64), make(map[string][]float64))

	wave.WaveRule.CommonRule.MaxItemQtyPerSubPickingTask = 100000000

	solverConfig := &waveoptsolver.WaveSolverConfig{
		MaxSecondsSpent:    120,
		Parallelism:        3,
		VariableTabuTenure: 5,
		ValueTabuTenure:    5,
		ZoneCoeff:          100,
		PathwayCoeff:       10,
		SegmentCoeff:       1,
	}
	wave.SolverConfig = solverConfig
	//wave.Groups[0].Orders = wave.Groups[0].Orders[:100]

	waveOptSolver := &waveoptsolver.WavePickingTaskOptSolver{Wave: wave}
	//wave.WaveRule.WavePickerMode = waveoptsolver.SinglePickerOnly
	result := waveOptSolver.GeneratePickingTasks()

	if result.RetCode != waveoptsolver.SuccessResult {
		t.Errorf("RetCode: %d, Msg: %s", result.RetCode, result.Msg)
	}
}

func TestCheckWaveResult(t *testing.T) {
	rawRule, rawWave, _ := readOneWave("test_demo_idl.csv", "", 1)
	wave, _ := convertRawWaveToWave(rawWave, rawRule, make(map[string]int64), make(map[string][]float64))
	solverConfig := &waveoptsolver.WaveSolverConfig{
		MaxSecondsSpent:    300,
		Parallelism:        3,
		VariableTabuTenure: 5,
		ValueTabuTenure:    5,
		ZoneCoeff:          100,
		PathwayCoeff:       10,
		SegmentCoeff:       1,
	}
	wave.SolverConfig = solverConfig

	waveOptSolver := &waveoptsolver.WavePickingTaskOptSolver{Wave: wave}
	result := waveOptSolver.GeneratePickingTasks()

	// 总订单数相同
	expectedOrderNum := 0
	for _, group := range wave.Groups {
		expectedOrderNum += len(group.Orders)
	}

	actualOrderNum := 0
	for _, task := range result.PickingTasks {
		actualOrderNum += len(task.OrderNos)
	}
	actualOrderNum += len(result.UnsolvedOrderNos)

	if expectedOrderNum != actualOrderNum {
		t.Errorf("总订单数不同. expected: %d, actual: %d.", expectedOrderNum, actualOrderNum)
	}

	// 输出任务数符合上下限
	actualTaskNum := len(result.PickingTasks)
	expectedMinTaskNum := wave.WaveRule.CommonRule.MinPickingTaskQtyPerWave
	expectedMaxTaskNum := wave.WaveRule.CommonRule.MaxPickingTaskQtyPerWave
	if actualTaskNum < int(expectedMinTaskNum) || actualTaskNum > int(expectedMaxTaskNum) {
		t.Errorf("输出任务数有误. expected range: %d~%d, actual: %d.", expectedMinTaskNum, expectedMaxTaskNum, actualTaskNum)
	}

	// 每个任务的订单数符合上下限
	expectedMaxOrderQtyPerTask := wave.WaveRule.CommonRule.MaxOrderQtyPerPickingTask
	expectedMinOrderQtyPerTask := wave.WaveRule.CommonRule.MinOrderQtyPerPickingTask
	for _, task := range result.PickingTasks {
		actualOrderQtyPerTask := len(task.OrderNos)
		if actualOrderQtyPerTask < int(expectedMinOrderQtyPerTask) || actualOrderQtyPerTask > int(expectedMaxOrderQtyPerTask) {
			t.Errorf("拣货任务订单数有误. expected range: %d~%d, actual: %d.", expectedMinOrderQtyPerTask, expectedMaxOrderQtyPerTask, actualOrderQtyPerTask)
		}
	}

	// 每个子任务的item数、体积符合上下限，跨区、跨cluster符合约束
	expectedMinItemQtyPerSubTask := wave.WaveRule.CommonRule.MinItemQtyPerSubPickingTask
	expectedMaxItemQtyPerSubTask := wave.WaveRule.CommonRule.MaxItemQtyPerSubPickingTask
	expectedMinItemVolumePerSubTask := wave.WaveRule.CommonRule.MinItemVolumePerSubPickingTask
	expectedMaxItemVolumePerSubTask := wave.WaveRule.CommonRule.MaxItemVolumePerSubPickingTask

	skuMap := make(map[string]*waveoptsolver.WaveSku)
	for _, sku := range wave.Skus {
		skuMap[sku.SkuId] = sku
	}

	locationMap := make(map[string]*waveoptsolver.Location)
	for _, location := range wave.Locations {
		locationMap[location.LocationId] = location
	}

	for _, task := range result.PickingTasks {
		for _, subTask := range task.SubPickingTasks {
			actualItemQtyPerSubTask := len(subTask.Skus)
			if actualItemQtyPerSubTask < int(expectedMinItemQtyPerSubTask) || actualItemQtyPerSubTask > int(expectedMaxItemQtyPerSubTask) {
				t.Errorf("子任务商品数有误. expected range: %d~%d, actual: %d.", expectedMinItemQtyPerSubTask, expectedMaxItemQtyPerSubTask, actualItemQtyPerSubTask)
			}

			actualItemVolumePerSubTask := 0
			for _, sku := range subTask.Skus {
				skuDim := skuMap[sku.SkuId]
				skuVolume := skuDim.Height * skuDim.Length * skuDim.Width
				if skuVolume == 0 {
					actualItemVolumePerSubTask = 0
				}
			}

			if actualItemVolumePerSubTask > 0 && (actualItemVolumePerSubTask < int(expectedMinItemVolumePerSubTask) || actualItemVolumePerSubTask > int(expectedMaxItemVolumePerSubTask)) {
				t.Errorf("子任务商品体积有误. expected range: %d~%d, actual: %d.", expectedMinItemVolumePerSubTask, expectedMaxItemVolumePerSubTask, actualItemVolumePerSubTask)
			}
		}

		clusterIdSet := make(map[string]struct{})
		for _, subTask := range task.SubPickingTasks {
			for _, sku := range subTask.Skus {
				clusterId := strconv.FormatInt(locationMap[sku.LocationId].ZoneClusterId, 10)
				clusterIdSet[clusterId] = struct{}{}
			}
		}

		if !wave.WaveRule.CommonRule.IsCrossZoneCluster && len(clusterIdSet) > 1 {
			t.Errorf("任务%s违反跨Cluster约束", task.PickingTaskId)
		}

		if wave.WaveRule.WavePickerMode == waveoptsolver.SinglePickerOnly {
			continue
		}

		var splitLevel waveoptsolver.SplitLevel
		switch mr := wave.WaveRule.ModeRule.(type) {
		case *waveoptsolver.MultiPickerAtMWSTotalQtyRule:
			splitLevel = mr.PickingTaskSplitLevel
		case *waveoptsolver.MultiPickerAtMWSRespectiveQtyRule:
			splitLevel = mr.PickingTaskSplitLevel
		case *waveoptsolver.MultiPickerAtMLTotalQtyRule:
			splitLevel = mr.PickingTaskSplitLevel
		case waveoptsolver.MultiPickerAtMLRespectiveQtyRule:
			splitLevel = mr.PickingTaskSplitLevel
		default:
			t.Errorf("未知的ModeRule: %s", mr)
		}

		for _, subTask := range task.SubPickingTasks {
			locationSet := make(map[string]struct{})
			for _, sku := range subTask.Skus {
				skuLocationId := sku.LocationId
				location := locationMap[skuLocationId]
				switch splitLevel {
				case waveoptsolver.ByZone:
					locationSet[location.ZoneId] = struct{}{}
				case waveoptsolver.ByZoneCluster:
					locationSet[strconv.FormatInt(location.ZoneClusterId, 10)] = struct{}{}
				case waveoptsolver.ByZoneSector:
					locationSet[strconv.FormatInt(location.ZoneSectorId, 10)] = struct{}{}
				}
			}
			if len(locationSet) > 1 {
				t.Errorf("违反子任务拆分维度: %s", splitLevel)
			}
		}
	}

	// bulky/non-bulky 任务数/订单数 符合约束
	bulkyTaskNum := 0
	nonBulkyTaskNum := 0
	bulkyOrderNum := 0
	nonBulkyOrderNum := 0
	for _, task := range result.PickingTasks {
		if task.PickingTaskSize == waveoptsolver.BulkyTaskType || task.PickingTaskSize == waveoptsolver.ExtraBulkyTaskType {
			bulkyTaskNum += 1
			bulkyOrderNum += len(task.OrderNos)
		} else if task.PickingTaskSize == waveoptsolver.NonBulkyTaskType {
			nonBulkyTaskNum += 1
			nonBulkyOrderNum += len(task.OrderNos)
		} else {
			t.Errorf("未知的TaskSize: %s.", task.PickingTaskSize)
		}
	}

	switch wave.WaveRule.WavePickerMode {
	case waveoptsolver.SinglePickerOnly:
		break
	case waveoptsolver.MultiPickerAtMLWithTotalPickingTaskQty:
		modeRule := wave.WaveRule.ModeRule.(*waveoptsolver.MultiPickerAtMLTotalQtyRule)
		expectedTotalTaskNum := modeRule.MaxBacklogAtMLPerWave
		if bulkyTaskNum+nonBulkyTaskNum > int(expectedTotalTaskNum) {
			t.Errorf("输出Bulky + NonBulky任务数有误. expected: %d, actual: %d.", expectedTotalTaskNum, bulkyTaskNum+nonBulkyTaskNum)
		}
	case waveoptsolver.MultiPickerAtMLWithRespectivePickingTaskQty:
		modeRule := wave.WaveRule.ModeRule.(*waveoptsolver.MultiPickerAtMLRespectiveQtyRule)
		expectedBulkyTaskNum := modeRule.MaxBulkyBacklogAtMLPerWave
		if bulkyTaskNum > int(expectedBulkyTaskNum) {
			t.Errorf("输出Bulky任务数有误. expected: %d, actual: %d.", expectedBulkyTaskNum, bulkyTaskNum)
		}

		expectedNonBulkyTaskNUm := modeRule.MaxNonBulkyBacklogAtMLPerWave
		if nonBulkyTaskNum > int(expectedNonBulkyTaskNUm) {
			t.Errorf("输出NonBulky任务数有误. expected: %d, actual: %d.", expectedNonBulkyTaskNUm, nonBulkyTaskNum)
		}
	case waveoptsolver.MultiPickerAtMWSWithTotalOrderQty:
		modeRule := wave.WaveRule.ModeRule.(*waveoptsolver.MultiPickerAtMWSTotalQtyRule)
		expectedTotalOrderNum := modeRule.MaxBacklogAtMWSPerWave
		if bulkyOrderNum+nonBulkyOrderNum > int(expectedTotalOrderNum) {
			t.Errorf("输出Bulky + NonBulky订单数有误. expected: %d, actual: %d.", expectedTotalOrderNum, bulkyOrderNum+nonBulkyOrderNum)
		}
	case waveoptsolver.MultiPickerAtMWSWithRespectiveOrderQty:
		modeRule := wave.WaveRule.ModeRule.(*waveoptsolver.MultiPickerAtMWSRespectiveQtyRule)
		expectedBulkyOrderNum := modeRule.MaxBulkyBacklogAtMWSPerWave
		if bulkyOrderNum > int(expectedBulkyOrderNum) {
			t.Errorf("输出Bulky订单数有误. expected: %d, actual: %d.", expectedBulkyOrderNum, bulkyOrderNum)
		}

		expectedNonBulkyOrderNum := modeRule.MaxNonBulkyBacklogAtMWSPerWave
		if nonBulkyOrderNum > int(expectedNonBulkyOrderNum) {
			t.Errorf("输出NonBulky订单数有误. expected: %d, actual: %d.", expectedNonBulkyOrderNum, nonBulkyOrderNum)
		}
	}
}
