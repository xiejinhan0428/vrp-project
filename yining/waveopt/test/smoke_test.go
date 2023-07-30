package test

import (
	"encoding/json"
	"git.garena.com/shopee/bg-logistics/algo/wms/waveopt/waveoptsolver"
	"io/ioutil"
	"strconv"
	"testing"
)

//用例名称
//不集货：
//wave 约束（picking task个数&&picking task 的order数，上限约束）
//picking task 不约束，
//sub-picking task 约束（volume 上限约束）
//订单约束
//
//前置条件
//1、admin protal 配置wave_algorithm_switc = 1
//2、Multi-Picker  关闭
//3、准备订单，为非特殊单：
//a、
//准备order 1，order2，order3，order 4中均为sku 1  qty =1
//准备order5，6，7，8  sku2   qty = 1
//b、配置一个波次规则W1，sssq，
//Total Order Per Picking List最小为1，最大为2，（约束条件）
//Total Picking Task Qty Per Wave最小为1，最大为4（约束条件）
//允许cross zone cluster
//Min Items Per sub-task     1
//Max Items Per sub-task    10
//Min Task Size Per sub-task 体积  1
//Max Task Size Per sub-task 体积 11ml（约束条件）
//c、sku1 的库存均在L1
//d、sku1体积为2ml，sku2体积为10ml
//
//波次算法运算，
//一共生成4个拣货任务：
//a、order 1，order2，order3，order 4，分别生成两个拣货任务picking task1 和pt2，一个拣货任务两个order
//b、order5-8随机生成两个拣货任务，一个拣货任务一个order
func TestSmoke1(t *testing.T) {
	rawRule, rawWave, _ := readOneWave("test_demo_idl.csv", "", 1)
	wave, _ := convertRawWaveToWave(rawWave, rawRule, make(map[string]int64), make(map[string][]float64))
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

	wave.WaveRule.WavePickerMode = waveoptsolver.SinglePickerOnly
	wave.WaveRule.ModeRule = nil
	wave.WaveRule.CommonRule.MaxOrderQtyPerPickingTask = 2
	wave.WaveRule.CommonRule.MinOrderQtyPerPickingTask = 1
	wave.WaveRule.CommonRule.MaxItemQtyPerSubPickingTask = 10
	wave.WaveRule.CommonRule.MinItemQtyPerSubPickingTask = 1
	wave.WaveRule.CommonRule.MaxItemVolumePerSubPickingTask = 11
	wave.WaveRule.CommonRule.MinItemVolumePerSubPickingTask = 1
	wave.WaveRule.CommonRule.MaxPickingTaskQtyPerWave = 4
	wave.WaveRule.CommonRule.MinPickingTaskQtyPerWave = 1

	wave.Locations = wave.Locations[:1]
	wave.Skus = wave.Skus[:2]
	wave.Groups = wave.Groups[:1]
	wave.Groups[0].Orders = wave.Groups[0].Orders[:8]

	sku1 := wave.Skus[0]
	sku2 := wave.Skus[1]

	sku1.SkuId = "sku1"
	sku1.Height = 1
	sku1.Width = 1
	sku1.Length = 2

	sku2.SkuId = "sku2"
	sku2.Height = 1
	sku2.Width = 1
	sku2.Length = 11

	for i, order := range wave.Groups[0].Orders {
		order.OrderNo = "order" + strconv.FormatInt(int64(i+1), 10)
		if i < 4 {
			order.Skus = append(make([]*waveoptsolver.OrderSku, 0), &waveoptsolver.OrderSku{
				OrderNo:    order.OrderNo,
				SkuId:      sku1.SkuId,
				Qty:        1,
				LocationId: wave.Locations[0].LocationId,
			})
		} else {
			order.Skus = append(make([]*waveoptsolver.OrderSku, 0), &waveoptsolver.OrderSku{
				OrderNo:    order.OrderNo,
				SkuId:      sku2.SkuId,
				Qty:        1,
				LocationId: wave.Locations[0].LocationId,
			})
		}
	}

	inputJson, _ := json.MarshalIndent(wave, "", "    ")
	_ = ioutil.WriteFile("./smoke_test_1_input.json", inputJson, 0644)

	waveOptSolver := &waveoptsolver.WavePickingTaskOptSolver{Wave: wave}
	result := waveOptSolver.GeneratePickingTasks()

	outputJson, _ := json.MarshalIndent(result, "", "    ")
	_ = ioutil.WriteFile("./smoke_test_1_output.json", outputJson, 0644)
}

func TestSmoke2(t *testing.T) {
	inputJson, _ := ioutil.ReadFile("./smoke_test_2_input.json")

	wave := new(waveoptsolver.Wave)
	_ = json.Unmarshal(inputJson, wave)

	waveOptSolver := &waveoptsolver.WavePickingTaskOptSolver{Wave: wave}
	result := waveOptSolver.GeneratePickingTasks()

	outputJson, _ := json.MarshalIndent(result, "", "    ")
	_ = ioutil.WriteFile("./smoke_test_2_output.json", outputJson, 0644)
}

// 用例名称
// 集货MWS：
//wave 不约束
//picking task 不约束，
//sub-picking task 约束（volume 上下限约束，item 上下限约束）
//订单约束
//
// 前置条件
//1、admin protal 配置wave_algorithm_switc = 1
//2、Multi-Picker  打开选择MWS，
//Split Picking Task By  = zone
//Backlog Threshold Setting   = total
//3、准备订单，为非特殊单：
//a、order1，order2 ，order3，order4，order5， 分别有sku1 ，qty = 2  sku2   qty = 2
//order6-10   分别有sku3   qty = 2，sku4    qty = 2
//b、配置一个波次规则W1，mssq，允许跨zone cluste
//Total Order Per Picking List最小为1，最大为10，
//Total Picking Task Qty Per Wave最小为1，最大为10
//允许cross zone cluster
//Min Items Per sub-task     5
//Max Items Per sub-task    6
//Min Task Size Per sub-task 体积  1ml
//Max Task Size Per sub-task 体积 9ml
//c、sku1 的库存均在L1， sku2 库存均在L2，（L1和L2跨zone ）
//d、sku1体积2ml，sku2体积为2ml
//sku3体积为1ml，sku4体积为1ml
//4、WMS可用格口数为：MaxBacklogAtMWSPerWave=10
//
//1、波次算法运算，order1-order5的单无法生成拣货任务（最小Min item sub = 5，即一个sub中sku1最少有5个，5个sku1体积为5ml，超过了Max Volume 4ml ，所以无法生成）
//2、order6-10可以生成一个拣货任务拣货任务，为sku3有两个sub，qty为5，sku4有两个sub-qty为5（最小Min item sub = 5，即一个sub中sku3最少有5个，5个sku3体积为1ml，满足体积约束 ）
func TestSmoke3(t *testing.T) {
	rawRule, rawWave, _ := readOneWave("test_demo_idl.csv", "", 1)
	wave, _ := convertRawWaveToWave(rawWave, rawRule, make(map[string]int64), make(map[string][]float64))
	solverConfig := &waveoptsolver.WaveSolverConfig{
		MaxSecondsSpent:    30,
		Parallelism:        3,
		VariableTabuTenure: 1,
		ValueTabuTenure:    1,
		ZoneCoeff:          100,
		PathwayCoeff:       10,
		SegmentCoeff:       1,
	}
	wave.SolverConfig = solverConfig

	wave.WaveRule.WavePickerMode = waveoptsolver.MultiPickerAtMWSWithTotalOrderQty
	wave.WaveRule.ModeRule = &waveoptsolver.MultiPickerAtMWSTotalQtyRule{
		MaxBacklogAtMWSPerWave: 10,
		PickingTaskSplitLevel:  waveoptsolver.ByZone,
	}
	wave.WaveRule.CommonRule.IsCrossZoneCluster = true
	wave.WaveRule.CommonRule.MaxOrderQtyPerPickingTask = 10
	wave.WaveRule.CommonRule.MinOrderQtyPerPickingTask = 1
	wave.WaveRule.CommonRule.MaxItemQtyPerSubPickingTask = 6
	wave.WaveRule.CommonRule.MinItemQtyPerSubPickingTask = 5
	wave.WaveRule.CommonRule.MaxItemVolumePerSubPickingTask = 7
	wave.WaveRule.CommonRule.MinItemVolumePerSubPickingTask = 1
	wave.WaveRule.CommonRule.MaxPickingTaskQtyPerWave = 10
	wave.WaveRule.CommonRule.MinPickingTaskQtyPerWave = 1

	wave.Locations = wave.Locations[:4]
	wave.Skus = wave.Skus[:4]
	wave.Groups = wave.Groups[:1]
	wave.Groups[0].Orders = wave.Groups[0].Orders[:10]

	sku1 := wave.Skus[0]
	sku2 := wave.Skus[1]
	sku3 := wave.Skus[2]
	sku4 := wave.Skus[3]

	sku1.SkuId = "sku1"
	sku1.Height = 1
	sku1.Width = 1
	sku1.Length = 2

	sku2.SkuId = "sku2"
	sku2.Height = 1
	sku2.Width = 1
	sku2.Length = 2

	sku3.SkuId = "sku3"
	sku3.Height = 1
	sku3.Width = 1
	sku3.Length = 1

	sku4.SkuId = "sku4"
	sku4.Height = 1
	sku4.Width = 1
	sku4.Length = 1

	wave.Locations[0].ZoneId = "GA"
	wave.Locations[1].ZoneId = "GB"
	wave.Locations[2].ZoneId = "GC"
	wave.Locations[3].ZoneId = "GD"

	for i, order := range wave.Groups[0].Orders {
		order.OrderNo = "order" + strconv.FormatInt(int64(i+1), 10)
		if i < 5 {
			order.Skus = append(make([]*waveoptsolver.OrderSku, 0), &waveoptsolver.OrderSku{
				OrderNo:    order.OrderNo,
				SkuId:      sku1.SkuId,
				Qty:        2,
				LocationId: wave.Locations[0].LocationId,
			}, &waveoptsolver.OrderSku{
				OrderNo:    order.OrderNo,
				SkuId:      sku2.SkuId,
				Qty:        2,
				LocationId: wave.Locations[1].LocationId,
			})
		} else {
			order.Skus = append(make([]*waveoptsolver.OrderSku, 0), &waveoptsolver.OrderSku{
				OrderNo:    order.OrderNo,
				SkuId:      sku3.SkuId,
				Qty:        2,
				LocationId: wave.Locations[2].LocationId,
			}, &waveoptsolver.OrderSku{
				OrderNo:    order.OrderNo,
				SkuId:      sku4.SkuId,
				Qty:        2,
				LocationId: wave.Locations[3].LocationId,
			})
		}
	}

	//wave.Groups[0].Orders = wave.Groups[0].Orders[:5]

	inputJson, _ := json.MarshalIndent(wave, "", "    ")
	_ = ioutil.WriteFile("./smoke_test_3_input.json", inputJson, 0644)

	waveOptSolver := &waveoptsolver.WavePickingTaskOptSolver{Wave: wave}
	result := waveOptSolver.GeneratePickingTasks()

	outputJson, _ := json.MarshalIndent(result, "", "    ")
	_ = ioutil.WriteFile("./smoke_test_3_output.json", outputJson, 0644)
}

func TestSmoke4(t *testing.T) {
	inputJson, _ := ioutil.ReadFile("./smoke_test_4_input.json")

	wave := new(waveoptsolver.Wave)
	_ = json.Unmarshal(inputJson, wave)

	rawModeRule := wave.WaveRule.ModeRule.(map[string]interface{})
	maxBacklogAtMLPerWave := int64(rawModeRule["max_backlog_at_ml_per_wave"].(float64))
	splitLeve := rawModeRule["picking_task_split_level"].(string)

	wave.WaveRule.ModeRule = &waveoptsolver.MultiPickerAtMLTotalQtyRule{
		MaxBacklogAtMLPerWave: maxBacklogAtMLPerWave,
		PickingTaskSplitLevel: waveoptsolver.SplitLevel(splitLeve),
	}

	waveOptSolver := &waveoptsolver.WavePickingTaskOptSolver{Wave: wave}
	result := waveOptSolver.GeneratePickingTasks()

	outputJson, _ := json.MarshalIndent(result, "", "    ")
	_ = ioutil.WriteFile("./smoke_test_4_output.json", outputJson, 0644)
}

func TestSmoke5(t *testing.T) {
	inputJson, _ := ioutil.ReadFile("./smoke_test_5_input.json")

	wave := new(waveoptsolver.Wave)
	_ = json.Unmarshal(inputJson, wave)

	rawModeRule := wave.WaveRule.ModeRule.(map[string]interface{})
	maxBacklogAtMLPerWave := int64(rawModeRule["max_backlog_at_mws_per_wave"].(float64))
	splitLeve := rawModeRule["picking_task_split_level"].(string)

	wave.WaveRule.ModeRule = &waveoptsolver.MultiPickerAtMWSTotalQtyRule{
		MaxBacklogAtMWSPerWave: maxBacklogAtMLPerWave,
		PickingTaskSplitLevel:  waveoptsolver.SplitLevel(splitLeve),
	}

	waveOptSolver := &waveoptsolver.WavePickingTaskOptSolver{Wave: wave}
	result := waveOptSolver.GeneratePickingTasks()

	outputJson, _ := json.MarshalIndent(result, "", "    ")
	_ = ioutil.WriteFile("./smoke_test_5_output.json", outputJson, 0644)
}

func TestSmoke6(t *testing.T) {
	inputJson, _ := ioutil.ReadFile("./smoke_test_6_input.json")

	wave := new(waveoptsolver.Wave)
	_ = json.Unmarshal(inputJson, wave)

	rawModeRule := wave.WaveRule.ModeRule.(map[string]interface{})
	maxBacklogAtMLPerWave := int64(rawModeRule["max_backlog_at_ml_per_wave"].(float64))
	splitLeve := rawModeRule["picking_task_split_level"].(string)

	wave.WaveRule.ModeRule = &waveoptsolver.MultiPickerAtMLTotalQtyRule{
		MaxBacklogAtMLPerWave: maxBacklogAtMLPerWave,
		PickingTaskSplitLevel: waveoptsolver.SplitLevel(splitLeve),
	}

	waveOptSolver := &waveoptsolver.WavePickingTaskOptSolver{Wave: wave}
	result := waveOptSolver.GeneratePickingTasks()

	outputJson, _ := json.MarshalIndent(result, "", "    ")
	_ = ioutil.WriteFile("./smoke_test_6_output.json", outputJson, 0644)
}

func Test1(t *testing.T) {
	inputJson, _ := ioutil.ReadFile("./awave.json")
	wave := new(waveoptsolver.Wave)
	_ = json.Unmarshal(inputJson, wave)

	rawModeRule := wave.WaveRule.ModeRule.(map[string]interface{})
	var splitLevel waveoptsolver.SplitLevel
	if rawModeRule != nil && len(rawModeRule) > 0 {
		splitLevel = waveoptsolver.SplitLevel(rawModeRule["picking_task_split_level"].(string))
	}
	switch wave.WaveRule.WavePickerMode {
	case waveoptsolver.SinglePickerOnly:
		wave.WaveRule.ModeRule = nil
	case waveoptsolver.MultiPickerAtMWSWithTotalOrderQty:
		maxBacklogAtMWSPerWave := int64(rawModeRule["max_backlog_at_mws_per_wave"].(float64))

		wave.WaveRule.ModeRule = &waveoptsolver.MultiPickerAtMWSTotalQtyRule{
			MaxBacklogAtMWSPerWave: maxBacklogAtMWSPerWave,
			PickingTaskSplitLevel:  splitLevel,
		}
	case waveoptsolver.MultiPickerAtMWSWithRespectiveOrderQty:
		maxBulkyBacklogAtMWSPerWave := int64(rawModeRule["max_bulky_backlog_at_mws_per_wave"].(float64))
		maxNonBulkyBacklogAtMWSPerWave := int64(rawModeRule["max_non_bulky_backlog_at_mws_per_wave"].(float64))

		wave.WaveRule.ModeRule = &waveoptsolver.MultiPickerAtMWSRespectiveQtyRule{
			MaxBulkyBacklogAtMWSPerWave:    maxBulkyBacklogAtMWSPerWave,
			MaxNonBulkyBacklogAtMWSPerWave: maxNonBulkyBacklogAtMWSPerWave,
			PickingTaskSplitLevel:          splitLevel,
		}
	case waveoptsolver.MultiPickerAtMLWithTotalPickingTaskQty:
		maxBacklogAtMlPerWave := int64(rawModeRule["max_backlog_at_ml_per_wave"].(float64))

		wave.WaveRule.ModeRule = &waveoptsolver.MultiPickerAtMLTotalQtyRule{
			MaxBacklogAtMLPerWave: maxBacklogAtMlPerWave,
			PickingTaskSplitLevel: splitLevel,
		}
	case waveoptsolver.MultiPickerAtMLWithRespectivePickingTaskQty:
		maxBulkyBacklogAtMLPerWave := int64(rawModeRule["max_bulky_backlog_at_ml_per_wave"].(float64))
		maxNonBulkyBacklogAtMLPerWave := int64(rawModeRule["max_non_bulky_backlog_at_ml_per_wave"].(float64))

		wave.WaveRule.ModeRule = &waveoptsolver.MultiPickerAtMLRespectiveQtyRule{
			MaxBulkyBacklogAtMLPerWave:    maxBulkyBacklogAtMLPerWave,
			MaxNonBulkyBacklogAtMLPerWave: maxNonBulkyBacklogAtMLPerWave,
			PickingTaskSplitLevel:         splitLevel,
		}
	}

	waveOptSolver := &waveoptsolver.WavePickingTaskOptSolver{Wave: wave}

	result := waveOptSolver.GeneratePickingTasks()

	t.Log(result)
}
