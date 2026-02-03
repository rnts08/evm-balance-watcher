package utils

import (
	"math/big"
	"testing"
)

func TestTruncateString(t *testing.T) {
	tests := []struct {
		input    string
		length   int
		expected string
	}{
		{"hello world", 5, "he..."},
		{"short", 10, "short"},
		{"exact", 5, "exact"},
		{"", 5, ""},
		{"abc", 2, "ab"},
		{"abc", 3, "abc"},
	}

	for _, tt := range tests {
		result := TruncateString(tt.input, tt.length)
		if result != tt.expected {
			t.Errorf("TruncateString(%q, %d) = %q; want %q", tt.input, tt.length, result, tt.expected)
		}
	}
}

func TestAddCommas(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"123", "123"},
		{"1234", "1,234"},
		{"123456", "123,456"},
		{"1234567", "1,234,567"},
		{"1234.56", "1,234.56"},
		{"-1234", "-1,234"},
		{"", ""},
	}

	for _, tt := range tests {
		result := AddCommas(tt.input)
		if result != tt.expected {
			t.Errorf("AddCommas(%q) = %q; want %q", tt.input, result, tt.expected)
		}
	}
}

func TestFormatFloat(t *testing.T) {
	tests := []struct {
		input    float64
		decimals int
		expected string
	}{
		{1234.5678, 2, "1,234.57"},
		{1234.5, 2, "1,234.50"},
		{0, 2, "0.00"},
	}

	for _, tt := range tests {
		result := FormatFloat(tt.input, tt.decimals)
		if result != tt.expected {
			t.Errorf("FormatFloat(%f, %d) = %q; want %q", tt.input, tt.decimals, result, tt.expected)
		}
	}
}

func TestFormatBigFloat(t *testing.T) {
	tests := []struct {
		input    *big.Float
		decimals int
		expected string
	}{
		{big.NewFloat(1234.5678), 2, "1,234.57"},
		{nil, 2, "0"},
	}

	for _, tt := range tests {
		result := FormatBigFloat(tt.input, tt.decimals)
		if result != tt.expected {
			t.Errorf("FormatBigFloat(%v, %d) = %q; want %q", tt.input, tt.decimals, result, tt.expected)
		}
	}
}
