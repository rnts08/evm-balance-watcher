package utils

import (
	"fmt"
	"math/big"
	"strings"
)

func TruncateString(str string, num int) string {
	if len(str) <= num {
		return str
	}
	if num <= 3 {
		return str[:num]
	}
	return str[0:num-3] + "..."
}

func AddCommas(s string) string {
	if len(s) == 0 {
		return s
	}
	parts := strings.Split(s, ".")
	integerPart := parts[0]
	sign := ""
	if strings.HasPrefix(integerPart, "-") {
		sign = "-"
		integerPart = integerPart[1:]
	}

	n := len(integerPart)
	if n <= 3 {
		return s
	}

	var result strings.Builder
	result.WriteString(sign)
	remainder := n % 3
	if remainder > 0 {
		result.WriteString(integerPart[:remainder])
		result.WriteString(",")
	}
	for i := remainder; i < n; i += 3 {
		if i > remainder {
			result.WriteString(",")
		}
		result.WriteString(integerPart[i : i+3])
	}

	if len(parts) > 1 {
		result.WriteString(".")
		result.WriteString(parts[1])
	}
	return result.String()
}

func FormatFloat(f float64, decimals int) string {
	return AddCommas(fmt.Sprintf("%.*f", decimals, f))
}

func FormatBigFloat(f *big.Float, decimals int) string {
	if f == nil {
		return "0"
	}
	return AddCommas(f.Text('f', decimals))
}

func BigFloatToFloat64(f *big.Float) float64 {
	if f == nil {
		return 0
	}
	val, _ := f.Float64()
	return val
}
