package cron

import (
	"testing"
	"time"
)

func TestWouldRunNow(t *testing.T) {
	t.Parallel()

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
	expect(false, Job{Pattern: "0 */4 * * *"}, time.Date(2020, time.January, 1, 17, 0, 0, 0, time.UTC))

	// At the start of every 4th hour on Monday January 1st
	expect(true, Job{Pattern: "0 */4 1 1 1"}, time.Date(2029, time.January, 1, 16, 0, 0, 0, time.UTC))

	// At the start of each hour between 9AM to 5PM
	expect(true, Job{Pattern: "0 9-17 * * *"}, time.Date(2020, time.January, 1, 12, 0, 0, 0, time.UTC))
	expect(false, Job{Pattern: "0 9-17 * * *"}, time.Date(2020, time.January, 1, 8, 0, 0, 0, time.UTC))
	expect(false, Job{Pattern: "0 9-17 * * *"}, time.Date(2020, time.January, 1, 19, 0, 0, 0, time.UTC))

	// At 03:00, 05:00, 07:00 every day
	expect(true, Job{Pattern: "0 3,5,7 * * *"}, time.Date(2020, time.January, 1, 3, 0, 0, 0, time.UTC))
	expect(true, Job{Pattern: "0 3,5,7 * * *"}, time.Date(2020, time.January, 1, 5, 0, 0, 0, time.UTC))
	expect(true, Job{Pattern: "0 3,5,7 * * *"}, time.Date(2020, time.January, 1, 7, 0, 0, 0, time.UTC))
	expect(false, Job{Pattern: "0 3,5,7 * * *"}, time.Date(2020, time.January, 1, 3, 1, 0, 0, time.UTC))
	expect(false, Job{Pattern: "0 3,5,7 * * *"}, time.Date(2020, time.January, 1, 12, 0, 0, 0, time.UTC))
}
