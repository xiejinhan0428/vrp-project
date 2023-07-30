package test

import (
	"encoding/json"
	"fmt"
	"git.garena.com/shopee/bg-logistics/algo/wms/waveopt/waveoptsolver"
	"io/ioutil"
	"testing"
)

func TestConvData(t *testing.T) {
	for i := 1; i < 6; i++ {
		rawRule, rawWave, _ := readOneWave("test_demo_idl.csv", "", i)
		wave, _ := convertRawWaveToWave(rawWave, rawRule, make(map[string]int64), make(map[string][]float64))
		solverConfig := &waveoptsolver.WaveSolverConfig{
			MaxSecondsSpent:    1200,
			Parallelism:        3,
			VariableTabuTenure: 5,
			ValueTabuTenure:    5,
			ZoneCoeff:          100,
			PathwayCoeff:       10,
			SegmentCoeff:       1,
		}
		wave.SolverConfig = solverConfig

		inputJson, _ := json.MarshalIndent(wave, "", "    ")
		_ = ioutil.WriteFile(fmt.Sprintf("data_sample_%d.json", i), inputJson, 0644)
	}
}
