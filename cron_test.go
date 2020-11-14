package cron_test

import (
	"testing"
	"time"

	"github.com/ecnepsnai/cron"
)

func TestCronStop(t *testing.T) {
	t.Parallel()

	var tab *cron.Tab
	tab, _ = cron.New([]cron.Job{
		{
			Name:    "StopCron",
			Pattern: "* * * * *",
			Exec: func() {
				tab.StopSoon()
			},
		},
	})
	tab.Interval = 1 * time.Millisecond
	tab.ForceStart()
}

func TestCronPanic(t *testing.T) {
	t.Parallel()

	didPanic := 0
	var tab *cron.Tab
	tab, _ = cron.New([]cron.Job{
		{
			Name:    "PanicCron",
			Pattern: "* * * * *",
			Exec: func() {
				didPanic = 1
				panic("(intentional panic)")
			},
		},
	})
	tab.Interval = 1 * time.Minute
	go tab.ForceStart()
	i := 0
	for {
		i++
		if i > 10 {
			t.Fatalf("Scheduled job never ran?")
		}
		if didPanic == 1 {
			return
		}
		time.Sleep(1 * time.Millisecond)
	}
}
