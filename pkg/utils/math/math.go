package math

// GetFloorAfterPercentage will return the floor value nearest
// to multiple after subtracting percentage value
func GetFloorAfterPercentage(value float64, percentage, multiple int) int {
	percentageValue := int(value * float64(float64(percentage)/100))
	afterPercentageValue := int(value) - percentageValue

	if multiple == 0 {
		return afterPercentageValue
	}

	remainder := afterPercentageValue % multiple
	if remainder == 0 {
		return afterPercentageValue
	}

	return afterPercentageValue - remainder
}
