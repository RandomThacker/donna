package icscalendar

import "errors"

var (
	errEmptyURL          = errors.New("ics feed url is required")
	errUnsupportedScheme = errors.New("ics feed url must be http(s) or webcal")
	errEmptyCalendar     = errors.New("ics feed contained no calendar")
	errNotModified       = errors.New("ics feed not modified")
)
