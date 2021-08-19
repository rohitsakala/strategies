package duration

import "time"

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

func ValidateTime(start time.Time, end time.Time, timeZone time.Location) bool {
	now := time.Now().In(&timeZone)

	if (start != time.Time{}) && (end != time.Time{}) {
		// compare hours
		if end.Sub(now).Hours() >= 0 {
			// compare minutes
			if end.Sub(now).Minutes() >= 0 {
				// compare hours
				if now.Sub(start).Hours() >= 0 {
					// compare minutes
					if now.Sub(start).Minutes() >= 0 {
						return true
					}
				}
			}
		}
	}

	return false
}
