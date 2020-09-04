// Package cron provides mechanism for executing scheduled tasks with cron-like expression.
//
// Cron expressions are a simple and flexible way to configure a schedule for which an automated task should run.
//
//    * * * * *
//    | | | | \
//    | | | \  The day of the week (0 = Sunday, 6 = Saturday)
//    | | \  The day of the month (1-31)
//    | \  Month (1-12)
//    \  Hour (0-23)
//     Minute (0-59)
//
// Each component of the expression can be a numerical value, an expression, or a wildcard. All components must match
// the current time for the job to run.
//
// If the component is a numerical value, then the same component (minute, hour, month, etc...) of the current time must
// match the exact value for the component.
//
// Components can also be an expression for a mod operation, such as */5 or */2. Where if the remainder from the
// current times component and the expression is zero, it matches.
//
// Lastly, components can be a wildcard *, which will match any time value.
//
// Some common expressions are:
//     "* * * * *" Run every minute
//     "0 * * * *" Run at the start of every hour
//     "0 0 * * *" Run every day at midnight
//     "*/5 * * *" Run every 5 minutes
//     "* */2 * *" Run every 2 hours
//
// Under normal circumstances cron is accurate up-to 1 second. Each job's method is called in a unique goroutine and
// will recover from any panics.
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
}

// New create a tab for the given slice of jobs. Does not start the tab.
func New(Jobs []Job) *Tab {
	return &Tab{
		Jobs:        Jobs,
		Interval:    60 * time.Second,
		ExpireAfter: nil,
	}
}

// Start will wait the next minute (up to 60 seconds) and then start the tab. This is the optimal way to start
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

// ForceStart will start the schedule immediately. This can have the undesired effect of jobs running at most
// 60 seconds later than they would if you used `Start`.
//
// This method blocks.
func (s *Tab) ForceStart() {
	log.Debug("Started tab")

	for {
		if s.ExpireAfter != nil {
			if time.Since(*s.ExpireAfter).Seconds() > 0 {
				log.Info("Tabr expired")
				return
			}
		}

		for _, job := range s.Jobs {
			if job.WouldRunNow() {
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

// WouldRunNow returns true if the pattern for the job matches the current time
func (job Job) WouldRunNow() bool {
	if job.Pattern == "* * * * *" {
		return true
	}

	components := strings.Split(job.Pattern, " ")
	clock := time.Now()

	return isItTime(components[0], clock.Minute()) &&
		isItTime(components[1], clock.Hour()) &&
		isItTime(components[2], clock.Day()) &&
		isItTime(components[3], int(clock.Month())) &&
		isItTime(components[4], int(clock.Weekday()))
}

func isItTime(dateComponent string, currentValue int) bool {
	if strings.Contains(dateComponent, "/") {
		divideBy, _ := strconv.Atoi(strings.Split(dateComponent, "/")[1])
		return currentValue%divideBy == 0
	}

	return dateComponent == toString(currentValue) || dateComponent == "*"
}

func (s *Tab) runJob(job Job) {
	start := time.Now()
	log.Debug("Starting scheduled job '%s'", job.Name)
	defer func() {
		if r := recover(); r != nil {
			log.Error("Tabd job '%s' panicked. Error: %v\n", job.Name, r)
		}
	}()
	job.Exec()
	elapsed := time.Since(start)
	log.Info("Tabd job '%s' finished in %s", job.Name, elapsed)
}

func toString(i int) string {
	return fmt.Sprintf("%d", i)
}