package schedule

import (
	"fmt"
	"strings"
	"time"
)

func dayStringToInt(day string) (int, error) {
	day = strings.ToLower(day)
	switch day {
	case "понедельник":
		return 1, nil
	case "вторник":
		return 2, nil
	case "среда":
		return 3, nil
	case "четверг":
		return 4, nil
	case "пятница":
		return 5, nil
	case "суббота":
		return 6, nil
	case "воскресенье":
		return 7, nil
	case "сегодня":
		weekday := int(time.Now().Weekday())
		if weekday == 0 {
			weekday = 7
		}
		return weekday, nil
	default:
		return 0, fmt.Errorf("неизвестный день: %s", day)
	}
}

func computeActualDateForWeekday(weekday int) (time.Time, error) {
	now := time.Now()
	current := int(now.Weekday())
	if current == 0 {
		current = 7
	}
	diff := weekday - current
	if diff < 0 {
		diff += 7
	}
	// Если запрошенный день сегодня, то diff будет 0.
	return now.AddDate(0, 0, diff), nil
}
