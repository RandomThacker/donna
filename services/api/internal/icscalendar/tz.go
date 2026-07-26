package icscalendar

import (
	"strings"
	"time"
)

// windowsTZ maps common Microsoft Outlook TZIDs to IANA names.
var windowsTZ = map[string]string{
	"India Standard Time":          "Asia/Kolkata",
	"Sri Lanka Standard Time":      "Asia/Colombo",
	"UTC":                          "UTC",
	"GMT Standard Time":            "Europe/London",
	"Greenwich Standard Time":      "Atlantic/Reykjavik",
	"W. Europe Standard Time":      "Europe/Berlin",
	"Central Europe Standard Time": "Europe/Budapest",
	"Romance Standard Time":        "Europe/Paris",
	"Central European Standard Time": "Europe/Warsaw",
	"GTB Standard Time":            "Europe/Bucharest",
	"E. Europe Standard Time":      "Europe/Chisinau",
	"Egypt Standard Time":          "Africa/Cairo",
	"South Africa Standard Time":   "Africa/Johannesburg",
	"FLE Standard Time":            "Europe/Kiev",
	"Israel Standard Time":         "Asia/Jerusalem",
	"Arabic Standard Time":         "Asia/Baghdad",
	"Arab Standard Time":           "Asia/Riyadh",
	"Russian Standard Time":        "Europe/Moscow",
	"SA Western Standard Time":     "America/La_Paz",
	"Atlantic Standard Time":       "America/Halifax",
	"SA Pacific Standard Time":     "America/Bogota",
	"US Eastern Standard Time":     "America/Indianapolis",
	"Eastern Standard Time":        "America/New_York",
	"US Mountain Standard Time":    "America/Phoenix",
	"Mountain Standard Time":       "America/Denver",
	"Central Standard Time":        "America/Chicago",
	"Pacific Standard Time":        "America/Los_Angeles",
	"Alaskan Standard Time":        "America/Anchorage",
	"Hawaiian Standard Time":       "Pacific/Honolulu",
	"Tokyo Standard Time":          "Asia/Tokyo",
	"Korea Standard Time":          "Asia/Seoul",
	"China Standard Time":          "Asia/Shanghai",
	"Singapore Standard Time":      "Asia/Singapore",
	"W. Australia Standard Time":   "Australia/Perth",
	"AUS Eastern Standard Time":    "Australia/Sydney",
	"E. Australia Standard Time":   "Australia/Brisbane",
	"New Zealand Standard Time":    "Pacific/Auckland",
	"Nepal Standard Time":          "Asia/Kathmandu",
	"Bangladesh Standard Time":     "Asia/Dhaka",
	"Pakistan Standard Time":       "Asia/Karachi",
	"Afghanistan Standard Time":    "Asia/Kabul",
	"Iran Standard Time":           "Asia/Tehran",
	"Arabian Standard Time":        "Asia/Dubai",
	"Azores Standard Time":         "Atlantic/Azores",
	"Cape Verde Standard Time":     "Atlantic/Cape_Verde",
	"Morocco Standard Time":        "Africa/Casablanca",
}

func loadLocation(tzid string) *time.Location {
	tzid = strings.TrimSpace(tzid)
	if tzid == "" {
		return time.UTC
	}
	if loc, err := time.LoadLocation(tzid); err == nil {
		return loc
	}
	if mapped, ok := windowsTZ[tzid]; ok {
		if loc, err := time.LoadLocation(mapped); err == nil {
			return loc
		}
	}
	// Case-insensitive Windows lookup.
	lower := strings.ToLower(tzid)
	for win, iana := range windowsTZ {
		if strings.ToLower(win) == lower {
			if loc, err := time.LoadLocation(iana); err == nil {
				return loc
			}
		}
	}
	return time.UTC
}
