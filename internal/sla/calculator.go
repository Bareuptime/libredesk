package sla

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/abhinavxd/libredesk/internal/business_hours/models"
)

var (
	ErrInvalidSLADuration = fmt.Errorf("invalid SLA duration")
	ErrMaxIterations      = fmt.Errorf("sla: exceeded maximum iterations - check configuration")
)

// CalculateDeadline computes the SLA deadline from a start time and SLA duration in minutes
// considering the provided holidays, working hours, and time zone.
func (m *Manager) CalculateDeadline(start time.Time, slaMinutes int, businessHours models.BusinessHours, timeZone string) (time.Time, error) {
	if slaMinutes <= 0 {
		return time.Time{}, ErrInvalidSLADuration
	}

	// If business is always open, return the deadline as the start time plus the SLA duration.
	if businessHours.IsAlwaysOpen {
		return start.Add(time.Duration(slaMinutes) * time.Minute), nil
	}

	// Load the specified time zone.
	loc, err := time.LoadLocation(timeZone)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid time zone %s: %v", timeZone, err)
	}

	// Convert start time to the specified time zone.
	currentTime := start.In(loc)
	remainingMinutes := slaMinutes
	maxIterations := ((slaMinutes+59)/60)*24 + 1

	// Unmarshal working hours.
	var workingHours map[string]models.WorkingHours
	if err := json.Unmarshal(businessHours.Hours, &workingHours); err != nil {
		return time.Time{}, fmt.Errorf("could not unmarshal working hours for SLA deadline calcuation: %v", err)
	}

	// Unmarshal holidays.
	var holidays = []models.Holiday{}
	if err := json.Unmarshal(businessHours.Holidays, &holidays); err != nil {
		return time.Time{}, fmt.Errorf("could not unmarshal holidays for SLA deadline calcuation: %v", err)
	}

	// Create a map of holidays.
	holidaysMap := make(map[string]struct{})
	for _, holiday := range holidays {
		holidaysMap[holiday.Date] = struct{}{}
	}

	iterations := 0
	for remainingMinutes > 0 {
		iterations++
		if iterations > maxIterations {
			return time.Time{}, ErrMaxIterations
		}

		// Skip holidays.
		dateStr := currentTime.Format(time.DateOnly)
		if _, isHoliday := holidaysMap[dateStr]; isHoliday {
			currentTime = nextDay(currentTime, loc)
			continue
		}

		// Get working hours for the current day.
		dayOfWeek := currentTime.Weekday().String()
		workHours, exists := workingHours[dayOfWeek]
		if !exists || workHours.ClosedAllDay {
			currentTime = nextDay(currentTime, loc)
			continue
		}

		// OpenAllDay scenario.
		var startOfWork, endOfWork time.Time
		if workHours.OpenAllDay {
			startOfWork = time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), 0, 0, 0, 0, loc)
			endOfWork = time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), 23, 59, 59, 0, loc)
		} else {
			var err error
			startOfWork, err = parseTime(currentTime, workHours.Open, loc)
			if err != nil {
				return time.Time{}, fmt.Errorf("invalid open time %s for %s: %v", workHours.Open, dayOfWeek, err)
			}
			endOfWork, err = parseTime(currentTime, workHours.Close, loc)
			if err != nil {
				return time.Time{}, fmt.Errorf("invalid close time %s for %s: %v", workHours.Close, dayOfWeek, err)
			}
		}

		// Adjust to start of work if current time is before it.
		if currentTime.Before(startOfWork) {
			currentTime = startOfWork
		}

		// Move to next day if current time is after end of work.
		if currentTime.After(endOfWork) {
			currentTime = nextDay(startOfWork, loc)
			continue
		}

		// Deduct minutes worked today from remaining SLA time.
		workMinutesLeft := int(endOfWork.Sub(currentTime).Minutes())
		if workMinutesLeft > remainingMinutes {
			return currentTime.Add(time.Duration(remainingMinutes) * time.Minute), nil
		}

		remainingMinutes -= workMinutesLeft
		currentTime = nextDay(startOfWork, loc)
	}

	return currentTime, nil
}

// nextDay advances the time to the start of the next day in the specified time zone.
func nextDay(t time.Time, loc *time.Location) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day()+1, 0, 0, 0, 0, loc)
}

// parseTime parses a time string in "HH:MM" format and returns a time.Time object for the given date and location.
func parseTime(date time.Time, timeStr string, loc *time.Location) (time.Time, error) {
	parsedTime, err := time.ParseInLocation("15:04", timeStr, loc)
	if err != nil {
		return time.Time{}, err
	}
	return time.Date(date.Year(), date.Month(), date.Day(), parsedTime.Hour(), parsedTime.Minute(), 0, 0, loc), nil
}
