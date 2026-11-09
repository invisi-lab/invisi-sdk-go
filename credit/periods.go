package credit

type Period struct{ CurrentDay, CurrentMonth uint64 }

// Periods derives periods from an explicit chain block timestamp. It does not use local time.
func Periods(timestamp uint64) Period {
	days := timestamp / 86400
	z := days + 719468
	era := z / 146097
	doe := z - era*146097
	yoe := (doe - doe/1460 + doe/36524 - doe/146096) / 365
	year := yoe + era*400
	doy := doe - (365*yoe + yoe/4 - yoe/100)
	mp := (5*doy + 2) / 153
	var month uint64
	if mp < 10 {
		month = mp + 3
	} else {
		month = mp - 9
	}
	if month <= 2 {
		year++
	}
	return Period{days + 1, (year-1970)*12 + month}
}
