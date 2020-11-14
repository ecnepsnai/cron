package cron

import (
	"testing"
	"time"
)

func TestWouldRunNow(t *testing.T) {
	expect := func(e bool, j Job, c time.Time) {
		r := j.wouldRunAtTime(c)
		if r != e {
			t.Errorf("Incorrect run now result for pattern '%s' at time '%s'. Got %v expected %v", j.Pattern, c, r, e)
		}
	}

	// Every minute
	expect(true, Job{Pattern: "* * * * *"}, time.Now())
	// At 00:00 every day
	expect(true, Job{Pattern: "0 0 * * *"}, time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC))
	// At 00:00 every day
	expect(false, Job{Pattern: "0 0 * * *"}, time.Date(2020, time.January, 1, 15, 26, 0, 0, time.UTC))
	// At the start of every 4th hour every day
	expect(true, Job{Pattern: "0 */4 * * *"}, time.Date(2020, time.January, 1, 16, 0, 0, 0, time.UTC))
	// At the start of every 4th hour every day
	expect(false, Job{Pattern: "0 */4 * * *"}, time.Date(2020, time.January, 1, 17, 0, 0, 0, time.UTC))
	// At the start of every 4th hour on Monday January 1st
	expect(true, Job{Pattern: "0 */4 1 1 1"}, time.Date(2029, time.January, 1, 16, 0, 0, 0, time.UTC))
}
