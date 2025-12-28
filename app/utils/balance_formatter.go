package utils

import (
	"fmt"
	"math"

	"goravel/app/models"
)

// FormatBalance 根据货币的小数位数格式化余额
// 如果货币没有小数（decimalPlaces = 0），则不返回小数部分
func FormatBalance(balance float64, currency *models.Currency) float64 {
	if currency == nil {
		// 如果没有货币信息，默认使用2位小数
		return math.Round(balance*100) / 100
	}

	decimalPlaces := currency.DecimalPlaces
	if decimalPlaces < 0 {
		decimalPlaces = 0
	}
	if decimalPlaces > 8 {
		decimalPlaces = 8
	}

	// 计算精度倍数
	multiplier := math.Pow(10, float64(decimalPlaces))

	// 四舍五入到指定小数位数
	formatted := math.Round(balance*multiplier) / multiplier

	return formatted
}

// FormatBalanceString 格式化余额为字符串（去除末尾多余的0）
func FormatBalanceString(balance float64, currency *models.Currency) string {
	formatted := FormatBalance(balance, currency)

	if currency == nil {
		return fmt.Sprintf("%.2f", formatted)
	}

	decimalPlaces := currency.DecimalPlaces
	if decimalPlaces < 0 {
		decimalPlaces = 0
	}
	if decimalPlaces > 8 {
		decimalPlaces = 8
	}

	// 如果小数位数为0，使用整数格式
	if decimalPlaces == 0 {
		return fmt.Sprintf("%.0f", formatted)
	}

	// 格式化字符串，去除末尾的0
	format := fmt.Sprintf("%%.%df", decimalPlaces)
	result := fmt.Sprintf(format, formatted)

	// 去除末尾的0和小数点（如果小数部分全为0）
	if decimalPlaces > 0 {
		// 去除末尾的0
		for len(result) > 0 && result[len(result)-1] == '0' {
			result = result[:len(result)-1]
		}
		// 如果最后是小数点，也去除
		if len(result) > 0 && result[len(result)-1] == '.' {
			result = result[:len(result)-1]
		}
	}

	return result
}
