package cron_test

import "github.com/ecnepsnai/cron"

func ExampleNew() {
	schedule := cron.New([]cron.Job{
		{
			Pattern: "* * * * *",
			Name:    "RunsEveryMinute",
			Exec: func() {
				// This would run every minute
			},
		},
		{
			Pattern: "0 * * * *",
			Name:    "OnTheHour",
			Exec: func() {
				// This would run at the start of every hour
			},
		},
		{
			Pattern: "*/5 * * * *",
			Name:    "Every5Minutes",
			Exec: func() {
				// This would run every 5 minutes
			},
		},
	})
	// This will start the cron at (or as close to as possible) 0 seconds of the next minute
	go schedule.Start()
}
