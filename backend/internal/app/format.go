package app

import (
	"net/http"
	"strconv"
	"strings"
	"time"
)

var vietnam = time.FixedZone("ICT", 7*60*60)

func money(n int64) string {
	s := strconv.FormatInt(n, 10)
	sign := ""
	if magnitude, negative := strings.CutPrefix(s, "-"); negative {
		sign, s = "-", magnitude
	}
	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + "." + s[i:]
	}
	return sign + s + " ₫"
}
func currentPeriod() string { return time.Now().In(vietnam).Format("2006-01") }
func periodOf(s string) string {
	if _, err := time.Parse("2006-01", s); err != nil {
		return currentPeriod()
	}
	return s
}
func number(r *http.Request, key string, low, high int64) (int64, error) {
	n, err := strconv.ParseInt(strings.TrimSpace(r.Form.Get(key)), 10, 64)
	if err != nil || n < low || n > high {
		return 0, errInvalid
	}
	return n, nil
}

func plainCSV(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "=") || strings.HasPrefix(s, "+") || strings.HasPrefix(s, "-") || strings.HasPrefix(s, "@") {
		return "'" + s
	}
	return s
}
