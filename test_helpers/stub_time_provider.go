package test_helpers

import "time"

var FixedTime = time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC)

type StubTimeProvider struct {
	FixedTime time.Time
}

func (stp *StubTimeProvider) Now() time.Time {
	return stp.FixedTime
}
