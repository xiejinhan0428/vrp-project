package arithutil

import merror "git.garena.com/shopee/bg-logistics/algo/wms/waveopt/localsearch/pkg/error"

func MinInt(nums ...int) (int, error) {
	if len(nums) == 0 {
		return 0, merror.New("MinInt: Cannot find the miminal elemets in an empty slice")
	}

	result := nums[0]
	size := len(nums)
	for i := 1; i < size; i++ {
		if nums[i] < result {
			result = nums[i]
		}
	}

	return result, nil
}
