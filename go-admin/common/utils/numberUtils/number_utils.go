package numberUtils

import (
	"errors"
	"fmt"
	"math"
)

// Abs 辅助函数：绝对值
func Abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

// Min 辅助函数：最小值
func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Max 辅助函数：最大值
func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Percentage 计算百分比
func Percentage[N, D int | int64 | float32 | float64](numerator N, denominator D, isOverDenominator bool) (string, error) {
	if denominator == 0 {
		return "0.00%", errors.New("分母不能为零")
	}
	percentage := (float64(numerator) / float64(denominator)) * 100
	if isOverDenominator && percentage > 100 {
		percentage = 100
	}
	result := fmt.Sprintf("%.2f%%", percentage)
	return result, nil
}

// PercentageNum 计算百分比
func PercentageNum[N, D int | int64 | float32 | float64](numerator N, denominator D, isOverDenominator bool) (float64, error) {
	if denominator == 0 {
		return 0.00, errors.New("分母不能为零")
	}
	percentage := (float64(numerator) / float64(denominator)) * 100
	percentage = math.Round(percentage*100) / 100
	if isOverDenominator && percentage > 100 {
		percentage = 100
	}
	return percentage, nil
}
