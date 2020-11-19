// Package cron provides a pure-go mechanism for executing scheduled tasks with cron patterns.
//
// Cron patterns are a simple and flexible way to configure a schedule for which an automated task should run.
//
//    * * * * *
//    | | | | \
//    | | | \  The day of the week (0=Sunday - 6=Saturday or MON-FRI)
//    | | \  Month (1=January - 12=December or JAN-DEC)
//    | \  The day of the month (1-31)
//    \  Hour (0-23)
//     Minute (0-59)
//
// Each component of the pattern can be: a single numerical value, a range, a comma-separated list of numerical values,
// an pattern, or a wildcard. Typically all values must match the current time for the job to run, however, when a day
// of month or day of week is specified (not a wildcard) those two values are OR-d. This can be confusing to understand,
// so know that the only real gotcha with this quirk is that there is no way to have a job run on a schedule such as
// 'every Friday the 13th'. It would instead run on every Friday and the 13th of each month.
//
// If the component is a numerical value, then the same component (minute, hour, month, etc...) of the current time must
// match the exact value for the component. If the component is a range, the current time value must fall between that
// range. If the component is a comma-separated list of numerical values, the current time must match any one of the
// values.
//
// Month and Day of Week values can also be the first three letters of the english name of that unit. For example,
// JAN for January or THU for Thursday.
//
// Components can also be an pattern for a mod operation, such as */5 or */2. Where if the remainder from the
// current times component and the pattern is zero, it matches.
//
// Lastly, components can be a wildcard *, which will match any value.
//
// Some example patterns are:
//     "* * * * *" Run every minute
//     "0 * * * *" Run at the start of every hour
//     "0 0 * * *" Run every day at midnight
//     "*/5 * * * *" Run every 5 minutes
//     "* */2 * * *" Run every 2 hours
//     "0 9-17 * * *" Run every day at the start every hour between 9AM to 5PM
//     "0 3,5,7 * * *" Run every day at 3AM, 5AM, and 7AM
//
// Cron wakes up each minute to check for any jobs to run, then sleeps for the remainder of the minute. Under normal
// circumstances cron is accurate up-to 1 second. Each job's method is called in a unique goroutine and will recover
// from any panics.
//
// This package conforms to the POSIX crontab standard, which can be found here:
// https://pubs.opengroup.org/onlinepubs/9699919799/utilities/crontab.html
package cron

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/ecnepsnai/logtic"
)

var log = logtic.Connect("cron")

// Tab describes a group of jobs, known as a "Tab"
type Tab struct {
	// The jobs to run
	Jobs []Job
	// Optional time when the schedule should expire. Set to nil for no expiry date.
	ExpireAfter *time.Time
	// The frequency to check if the jobs should run. By default this is 60 seconds and should not be changed.
	Interval time.Duration
}

// Job describes a single job that will run based on the pattern
type Job struct {
	// Cron pattern describing the schedule of this job
	Pattern string
	// The name of this job, only used for logging
	Name string
	// The method to invoke when the job runs
	Exec func()

	pattern []string
}

// New create a new cron instance (known as a "tab") for the given slice of jobs but do not start it.
// Error is only populated if there is a validation error on any of the job patterns.
func New(Jobs []Job) (*Tab, error) {
	for _, job := range Jobs {
		if err := job.Validate(); err != nil {
			return nil, err
		}
		job.pattern = getRealPattern(job.Pattern)
	}

	return &Tab{
		Jobs:        Jobs,
		Interval:    60 * time.Second,
		ExpireAfter: nil,
	}, nil
}

// Start will wait until the next minute (up to 60 seconds) and then start the tab. This is the optimal way to start
// the tab since jobs will run at the start of the minute.
//
// This method blocks.
func (s *Tab) Start() {
	// Wait until the next minute to start the tab
	// This ensures that minute based jobs run at the top of the minute
	waitDur := time.Duration(int(s.Interval.Seconds()) - time.Now().Second())
	log.Debug("Starting tab in %d seconds", waitDur)
	time.Sleep(waitDur * time.Second)
	s.ForceStart()
}

// ForceStart will start the schedule immediately without waiting. This can have the undesired effect of jobs running
// at most 60 seconds later than they would if you used `Start`.
//
// This method blocks.
func (s *Tab) ForceStart() {
	log.Debug("Started tab")

	for {
		if s.ExpireAfter != nil {
			if time.Since(*s.ExpireAfter).Seconds() > 0 {
				log.Debug("Tab expired")
				return
			}
		}

		for _, job := range s.Jobs {
			if job.WouldRunNow() {
				log.Debug("Running job: %s", job.Name)
				go s.runJob(job)
			}
		}
		time.Sleep(s.Interval)
	}
}

// StopSoon will stop the tab in no more than 60 seconds
func (s *Tab) StopSoon() {
	e := time.Now().AddDate(-1, 0, 0)
	s.ExpireAfter = &e
}

// WouldRunNow returns true if this job would run right now
func (job Job) WouldRunNow() bool {
	log.Debug("Job pattern: %s = %s", job.Name, job.Pattern)

	if job.Pattern == "* * * * *" {
		return true
	}

	if job.pattern == nil {
		job.pattern = getRealPattern(job.Pattern)
	}

	return patternDoesMatch(job.pattern, time.Now())
}

// patternDoesMatch does the given pattern match the specified time
func patternDoesMatch(pattern []string, clock time.Time) bool {
	minute := pattern[0]
	hour := pattern[1]
	dayOfMonth := pattern[2]
	month := pattern[3]
	dayOfWeek := pattern[4]

	minuteMatch := isItTime(minute, clock.Minute())
	hourMatch := isItTime(hour, clock.Hour())
	dayOfMonthMatch := isItTime(dayOfMonth, clock.Day())
	monthMatch := isItTime(month, int(clock.Month()))
	dayOfWeekMatch := isItTime(dayOfWeek, int(clock.Weekday()))

	dowIsStar := dayOfWeek == "*"
	domIsStar := dayOfMonth == "*"

	var dateOfMatch bool
	// From the spec:
	//
	//   if either the month or day of month is specified as an element or list, and the day of week is also specified
	//   as an element or list, then any day matching either the month and day of month, or the day of week, shall be
	//   matched.
	//
	// 'element or list' means it is not a wildcard *
	// So, to put it in simpler terms, if the day-of-week and day-of-month are both not wildcards, those two values are
	// OR-d. If either or both the day-of-week or day-of-month are wildcards, those two values are AND-d.
	//
	// To quote the SysV cron source "this routine is hard to understand"
	if !dowIsStar && !domIsStar {
		dateOfMatch = dayOfMonthMatch || dayOfWeekMatch
	} else {
		dateOfMatch = dayOfMonthMatch && dayOfWeekMatch
	}

	return (minuteMatch && hourMatch && monthMatch) && dateOfMatch
}

func isItTime(dateComponent string, currentValue int) bool {
	// We don't validate any of the values here since we do that when the tab is created
	if strings.ContainsRune(dateComponent, '/') {
		divideBy, _ := strconv.Atoi(strings.Split(dateComponent, "/")[1])
		return currentValue%divideBy == 0
	} else if strings.ContainsRune(dateComponent, '-') {
		parts := strings.Split(dateComponent, "-")
		start, _ := strconv.Atoi(parts[0])
		end, _ := strconv.Atoi(parts[1])
		return currentValue >= start && currentValue <= end
	} else if strings.ContainsRune(dateComponent, ',') {
		parts := strings.Split(dateComponent, ",")
		for _, part := range parts {
			value, _ := strconv.Atoi(part)
			if currentValue == value {
				return true
			}
		}
		return false
	}

	return dateComponent == toString(currentValue) || dateComponent == "*"
}

func (s *Tab) runJob(job Job) {
	start := time.Now()
	log.Debug("Starting scheduled job '%s'", job.Name)
	defer func() {
		if r := recover(); r != nil {
			log.Error("Scheduled job '%s' panicked. Error: %v\n", job.Name, r)
		}
	}()
	job.Exec()
	elapsed := time.Since(start)
	log.Debug("Scheduled job '%s' finished in %s", job.Name, elapsed)
}

func toString(i int) string {
	return fmt.Sprintf("%d", i)
}
