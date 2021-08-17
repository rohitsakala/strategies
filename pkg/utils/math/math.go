package math

import "math"

// GetFloorAfterPercentage will return the floor value nearest
// to multiple after subtracting percentage value
func GetFloorAfterPercentage(value float64, percentage, multiple int) float64 {
	percentageValue := int(value * float64(float64(percentage)/100))
	afterPercentageValue := value - float64(percentageValue)

	if multiple == 0 {
		return afterPercentageValue
	}

	remainder := int(afterPercentageValue) % multiple
	if remainder == 0 {
		return afterPercentageValue
	}

	return float64(int(afterPercentageValue) - remainder)
}

// GetNearestMultiple returns the nearest multiple
// example nearest 50th, 100th (multiple)
func GetNearestMultiple(value float64, multiple float64) float64 {
	return math.Round(value/multiple) * multiple
}
