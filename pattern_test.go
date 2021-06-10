package cron

import (
	"fmt"
	"testing"
	"time"
)

func TestPatternDoesMatch(t *testing.T) {
	t.Parallel()

	expect := func(expected bool, pattern string, clock time.Time) {
		result := patternDoesMatch(getRealPattern(pattern), clock)
		if result != expected {
			t.Errorf("Incorrect run now result for pattern '%s' at time '%s'. Got %v expected %v", pattern, clock, result, expected)
		}
	}

	// Every minute
	expect(true, "* * * * *", time.Now())

	// At 00:00 every day
	expect(true, "0 0 * * *", time.Date(2021, time.January, 1, 0, 0, 0, 0, time.UTC))

	// At 00:00 every day
	expect(false, "0 0 * * *", time.Date(2021, time.January, 1, 15, 26, 0, 0, time.UTC))

	// At the start of every 4th hour every day
	expect(true, "0 */4 * * *", time.Date(2021, time.January, 1, 16, 0, 0, 0, time.UTC))
	expect(false, "0 */4 * * *", time.Date(2021, time.January, 1, 17, 0, 0, 0, time.UTC))

	// At 00:00 January 1st
	expect(true, "0 0 1 JAN *", time.Date(2021, time.January, 1, 0, 0, 0, 0, time.UTC))
	expect(false, "0 0 1 JAN *", time.Date(2021, time.January, 5, 0, 0, 0, 0, time.UTC))

	// The SysV/BSD cron specification states that the day-of-month and day-of-week fields are OR-d when both are specified
	// Therefor, the pattern "* * 13 * 5" would run every Friday and on the 13th day of each month, where you may expect
	// it to run only on Friday the 13th.
	expect(true, "* * 13 * 5", time.Date(2021, time.January, 1, 0, 0, 0, 0, time.UTC))   // A friday
	expect(true, "* * 13 * 5", time.Date(2021, time.January, 13, 0, 0, 0, 0, time.UTC))  // The 13th
	expect(true, "* * 13 * 5", time.Date(2020, time.November, 13, 0, 0, 0, 0, time.UTC)) // Firday the 13th
	expect(false, "* * 13 * FRI", time.Date(2020, time.November, 1, 0, 0, 0, 0, time.UTC))

	// At the start of each hour between 9AM to 5PM
	expect(true, "0 9-17 * * *", time.Date(2021, time.January, 1, 12, 0, 0, 0, time.UTC))
	expect(false, "0 9-17 * * *", time.Date(2021, time.January, 1, 8, 0, 0, 0, time.UTC))
	expect(false, "0 9-17 * * *", time.Date(2021, time.January, 1, 19, 0, 0, 0, time.UTC))

	// At 03:00, 05:00, 07:00 every day
	expect(true, "0 3,5,7 * * *", time.Date(2021, time.January, 1, 3, 0, 0, 0, time.UTC))
	expect(true, "0 3,5,7 * * *", time.Date(2021, time.January, 1, 5, 0, 0, 0, time.UTC))
	expect(true, "0 3,5,7 * * *", time.Date(2021, time.January, 1, 7, 0, 0, 0, time.UTC))
	expect(false, "0 3,5,7 * * *", time.Date(2021, time.January, 1, 3, 1, 0, 0, time.UTC))
	expect(false, "0 3,5,7 * * *", time.Date(2021, time.January, 1, 12, 0, 0, 0, time.UTC))
}

func TestJobWouldRunNow(t *testing.T) {
	t.Parallel()

	job := Job{Pattern: "* * * * *"}
	if !job.WouldRunNow() {
		t.Errorf("Incorrect WouldRunNow result for all wildcard pattern")
	}

	job = Job{Pattern: fmt.Sprintf("* * * %d *", time.Now().Month())}
	if !job.WouldRunNow() {
		t.Errorf("Incorrect WouldRunNow result for all wildcard pattern")
	}
}

func TestWouldRunNowInTZ(t *testing.T) {
	t.Parallel()

	job := Job{Pattern: "* * * * *"}
	if !job.WouldRunNowInTZ(time.UTC) {
		t.Errorf("Incorrect WouldRunNowInTZ result for all wildcard pattern")
	}

	hour := time.Now().Hour()
	_, offset := time.Now().Zone()
	hour += offset

	job = Job{Pattern: fmt.Sprintf("* %d * * *", hour)}
	if !job.WouldRunNowInTZ(time.UTC) {
		t.Errorf("Incorrect WouldRunNowInTZ result for all wildcard pattern")
	}
}
