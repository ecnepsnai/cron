package cron

import (
	"testing"
	"time"
)

func BenchmarkWouldRunAtTime(b *testing.B) {
	for n := 0; n < b.N; n++ {
		patternDoesMatch(getRealPattern("*/5 0 1 JAN *"), time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC))
	}
}
