package cron_test

import (
	"testing"

	"github.com/ecnepsnai/cron"
)

func TestValidate(t *testing.T) {
	t.Parallel()

	expect := func(e bool, p string) {
		j := cron.Job{Pattern: p}
		r := j.Validate()
		if (r == nil) != e {
			t.Errorf("Incorrect validation result for pattern '%s'. Error: %s", p, r.Error())
		}
	}

	expect(true, "* * * * *")
	expect(true, "0 0 1 1 0")
	expect(false, "0 f 1 1 0")
	expect(false, "0 0 0 0 0")
	expect(false, "foo")
	expect(false, "0 */-1 * * *")
	expect(false, "0 */f * * *")
	expect(false, "61 0 0 0 0")
	expect(false, "* * * * * *")
	expect(false, "0 */1/1 * * *")
	expect(false, "0 f-1 * * *")
	expect(false, "0 1-f * * *")
	expect(false, "0 1-2-3 * * *")
	expect(false, "0 5-2 * * *")
	expect(false, "0 12-26 * * *")
	expect(true, "0 12-17 * * *")
	expect(true, "0 12,17 * * *")
	expect(false, "0 12,f * * *")
	expect(false, "0 12, * * *")
	expect(false, "0 12,25 * * *")
}
