package libdate

import "google.golang.org/genproto/googleapis/type/date"

func Equal(d1, d2 *date.Date) bool {
	if d1 == nil {
		d1 = &date.Date{}
	}
	if d2 == nil {
		d2 = &date.Date{}
	}
	return d1.Day == d2.Day && d1.Month == d2.Month && d1.Year == d2.Year
}
