package utils

import (
	"fmt"
	"math/big"
	"strings"
)

// FormatEther converts wei value into a decimal string using 18 decimal places.
func FormatEther(wei *big.Int) string {
	return FormatUnits(wei, 18)
}

// FormatUnits converts a value into a decimal string, assuming the specified number of decimal places.
func FormatUnits(value *big.Int, decimals int) string {
	if value == nil {
		return "0"
	}

	// Convert value to a decimal with the specified number of decimal places
	divisor := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)
	quotient, remainder := new(big.Int).QuoRem(value, divisor, new(big.Int))

	// Format the integer part
	integerPart := quotient.String()

	// Format the fractional part
	fractionalPart := fmt.Sprintf("%0*s", decimals, remainder.Abs(remainder).String())
	fractionalPart = strings.TrimRight(fractionalPart, "0")

	// Combine integer and fractional parts
	if fractionalPart == "" {
		return integerPart
	}
	return fmt.Sprintf("%s.%s", integerPart, fractionalPart)
}

// FormatUnitsString is a helper function that accepts a string input
func FormatUnitsString(value string, decimals int) (string, error) {
	// Remove "0x" prefix if present
	value = strings.TrimPrefix(value, "0x")

	// Parse the hex string to a big.Int
	bigIntValue, success := new(big.Int).SetString(value, 16)
	if !success {
		return "", fmt.Errorf("invalid hex string: %s", value)
	}

	return FormatUnits(bigIntValue, decimals), nil
}
