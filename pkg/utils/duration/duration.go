package duration

func GetNextMonthOffset(value string, offset int) string {
	var result string
	for i := 0; i < offset; i++ {
		result = GetNext(value)
		value = result
	}

	return result
}

func GetNext(value string) string {
	switch value {
	case "JAN":
		return "FEB"
	case "FEB":
		return "MAR"
	case "MAR":
		return "APR"
	case "APR":
		return "MAY"
	case "MAY":
		return "JUN"
	case "JUN":
		return "JUL"
	case "JUL":
		return "AUG"
	case "AUG":
		return "SEP"
	case "SEP":
		return "OCT"
	case "OCT":
		return "NOV"
	case "NOV":
		return "DEC"
	case "DEC":
		return "JAN"
	}

	return ""
}
