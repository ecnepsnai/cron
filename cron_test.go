package cron_test

import (
	"testing"
	"time"

	"github.com/ecnepsnai/cron"
)

func TestCronStop(t *testing.T) {
	var tab *cron.Tab
	tab = cron.New([]cron.Job{
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
	didPanic := 0
	var tab *cron.Tab
	tab = cron.New([]cron.Job{
		{
			Name:    "PanicCron",
			Pattern: "* * * * *",
			Exec: func() {
				didPanic = 1
				panic("(intential panic)")
			},
		},
	})
	tab.Interval = 1 * time.Minute
	go tab.ForceStart()
	i := 0
	for {
		i++
		if i > 10 {
			t.Fatalf("Tabd job never ran?")
		}
		if didPanic == 1 {
			return
		}
		time.Sleep(1 * time.Millisecond)
	}
}
