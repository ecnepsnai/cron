package cron

import (
	"fmt"
	"strconv"
	"strings"
)

// Validate ensure that the job pattern is valid, returning any error if invalid
func (job Job) Validate() error {
	if job.Pattern == "* * * * *" {
		return nil
	}
	components := strings.Split(job.Pattern, " ")
	if len(components) != 5 {
		return fmt.Errorf("Invalid number of date components")
	}

	dateUnits := []string{
		"minute",
		"hour",
		"day on month",
		"month",
		"day of week",
	}

	for i, component := range components {
		if component == "*" {
			continue
		}

		unit := dateUnits[i]

		if strings.ContainsRune(component, '/') {
			if err := validateExpression(component, unit, i); err != nil {
				return err
			}
		} else if strings.ContainsRune(component, '-') {
			if err := validateRange(component, unit, i); err != nil {
				return err
			}
		} else if strings.ContainsRune(component, ',') {
			if err := validateList(component, unit, i); err != nil {
				return err
			}
		} else {
			v, err := strconv.Atoi(component)
			if err != nil {
				return fmt.Errorf("Invalid %s value: %s", unit, err.Error())
			}
			if !validateDateComponent(v, i) {
				return fmt.Errorf("Invalid %s value", unit)
			}
		}
	}

	return nil
}

func validateExpression(component string, unit string, i int) error {
	parts := strings.Split(component, "/")
	if len(parts) > 2 {
		return fmt.Errorf("Invalid %s expression", unit)
	}
	value, err := strconv.Atoi(parts[1])
	if err != nil {
		return fmt.Errorf("Invalid %s expression: %s", unit, err.Error())
	}
	if !validateDateComponent(value, i) {
		return fmt.Errorf("Invalid %s expression", unit)
	}

	return nil
}

func validateRange(component string, unit string, i int) error {
	parts := strings.Split(component, "-")
	if len(parts) > 2 {
		return fmt.Errorf("Invalid %s range", unit)
	}
	left, err := strconv.Atoi(parts[0])
	if err != nil {
		return fmt.Errorf("Invalid %s range: %s", unit, err.Error())
	}
	right, err := strconv.Atoi(parts[1])
	if err != nil {
		return fmt.Errorf("Invalid %s range: %s", unit, err.Error())
	}
	if left > right || left == right {
		return fmt.Errorf("Invalid %s range", unit)
	}
	if !validateDateComponent(left, i) || !validateDateComponent(right, i) {
		return fmt.Errorf("Invalid %s expression", unit)
	}

	return nil
}

func validateList(component string, unit string, i int) error {
	for _, part := range strings.Split(component, ",") {
		value, err := strconv.Atoi(part)
		if err != nil {
			return fmt.Errorf("Invalid %s list: %s", unit, err.Error())
		}
		if !validateDateComponent(value, i) {
			return fmt.Errorf("Invalid %s list", unit)
		}
	}

	return nil
}

func validateDateComponent(v int, unit int) bool {
	switch unit {
	case 0:
		return validateMinute(v)
	case 1:
		return validateHour(v)
	case 2:
		return validateDayOfMonth(v)
	case 3:
		return validateMonth(v)
	case 4:
		return validateDayOfWeek(v)
	}

	return false
}

func validateMinute(v int) bool {
	return v >= 0 && v <= 60
}

func validateHour(v int) bool {
	return v >= 0 && v <= 24
}

func validateDayOfMonth(v int) bool {
	return v >= 1 && v <= 31
}

func validateMonth(v int) bool {
	return v >= 1 && v <= 12
}

func validateDayOfWeek(v int) bool {
	return v >= 0 && v <= 6
}
