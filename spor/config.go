package solver

func Config2(obj int) GFun {
	var gFun = GFun{
		InitCapUseTask:      InitCapUseTask2,
		InitResUse:          InitResUse2,
		UpdateResUse:        UpdateResUse2,
		ResetCapUse:         ResetCapUse2,
		UpdateCapUse:        UpdateCapUse2,
		CheckCap:            CheckCap2,
		CheckMaxCap:         CheckMaxCap2,
		CheckResCap:         CheckResCap2,
		GetProxTask:         GetProxTask2,
		GetDefaultIdxSorted: GetDefaultIdxSorted2,
	}
	if obj == 0 {
		gFun.GetObj = GetDistObj
	} else {
		gFun.GetObj = GetMapCostObj
	}
	return gFun
}
