package cron

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var monthMap = map[string]string{
	"JAN": "1",
	"FEB": "2",
	"MAR": "3",
	"APR": "4",
	"MAY": "5",
	"JUN": "6",
	"JUL": "7",
	"AUG": "8",
	"SEP": "9",
	"OCT": "10",
	"NOV": "11",
	"DEC": "12",
}

var weekdayMap = map[string]string{
	"SUN": "0",
	"MON": "1",
	"TUE": "2",
	"WED": "3",
	"THU": "4",
	"FRI": "5",
	"SAT": "6",
}

var alphabeticalPattern = regexp.MustCompile("[A-Z]{3}")

// Validate will ensure that the job pattern is valid and return an error with any validation error
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
		} else if alphabeticalPattern.MatchString(component) {
			if err := validateName(component, unit, i); err != nil {
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

func validateName(component string, unit string, i int) error {
	var m map[string]string
	if i == 3 {
		m = monthMap
	} else if i == 4 {
		m = weekdayMap
	} else {
		return fmt.Errorf("Invalid %s value", unit)
	}

	if _, ok := m[component]; !ok {
		return fmt.Errorf("Invalid %s value", unit)
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

// getRealPattern will return each of the 5 components from the given pattern converting any named values to their
// numerical equals. This assumes the pattern has already been validated and will panic on invalid patterns.
func getRealPattern(pattern string) []string {
	if pattern == "* * * * *" {
		return []string{"*", "*", "*", "*", "*"}
	}

	components := strings.Split(strings.ToUpper(pattern), " ")
	minute := components[0]
	hour := components[1]
	dayOfMonth := components[2]
	month := components[3]
	dayOfWeek := components[4]

	// Replace any named values (I.E. JAN or WED) with their numerical values
	if alphabeticalPattern.MatchString(month) {
		month = monthMap[month]
	}
	if alphabeticalPattern.MatchString(dayOfWeek) {
		dayOfWeek = weekdayMap[dayOfWeek]
	}

	return []string{minute, hour, dayOfMonth, month, dayOfWeek}
}
