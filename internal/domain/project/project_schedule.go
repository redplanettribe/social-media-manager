package project

import (
	"encoding/json"
	"errors"
	"time"
)

var (
	ErrInvalidDayOfWeek = errors.New("invalid day of week")
	ErrInvalidHour      = errors.New("invalid hour")
	ErrInvalidMinute    = errors.New("invalid minute")
)

type TimeSlot struct {
	DayOfWeek time.Weekday `json:"day_of_week" swaggertype:"integer" example:"1"` // 0 = Sunday, 1 = Monday, etc.
	Hour      int          `json:"hour"`
	Minute    int          `json:"minute"`
}

type WeeklyPostSchedule struct {
	Slots      []TimeSlot    `json:"slots"`
	TimeMargin time.Duration `json:"time_margin" swaggertype:"integer" example:"300000000"` // 5 minutes in nanoseconds
}

// NewWeeklyPostSchedule creates a new WeeklyPostSchedule.
func NewWeeklyPostSchedule(slots []TimeSlot) *WeeklyPostSchedule {
	return &WeeklyPostSchedule{
		Slots:      slots,
		TimeMargin: 5 * time.Minute,
	}
}

// Encode returns a string you can store in a TEXT column.
func (w *WeeklyPostSchedule) Encode() (string, error) {
	data, err := json.Marshal(w)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// DecodeSchedule populates a WeeklyPostSchedule from a JSON string.
func DecodeSchedule(encoded string) (*WeeklyPostSchedule, error) {
	var schedule WeeklyPostSchedule
	err := json.Unmarshal([]byte(encoded), &schedule)
	if err != nil {
		return nil, err
	}
	return &schedule, nil
}

// IsTime checks if the given time matches any scheduled slot within the time margin.
func (w *WeeklyPostSchedule) IsTime(t time.Time) bool {
	utcTime := t.UTC()
	for _, slot := range w.Slots {
		slotTime := time.Date(utcTime.Year(), utcTime.Month(), utcTime.Day(),
			slot.Hour, slot.Minute, 0, 0, time.UTC)
		if utcTime.Weekday() == slot.DayOfWeek &&
			utcTime.After(slotTime.Add(-w.TimeMargin)) &&
			utcTime.Before(slotTime.Add(w.TimeMargin)) {
			return true
		}
	}
	return false
}

// AddSlot adds a new slot to the schedule.
func (w *WeeklyPostSchedule) AddSlot(dayOfWeek time.Weekday, hour, minute int) error {
	if dayOfWeek < time.Sunday || dayOfWeek > time.Saturday {
		return ErrInvalidDayOfWeek
	}

	if hour < 0 || hour > 23 {
		return ErrInvalidHour
	}
	if minute < 0 || minute > 59 {
		return ErrInvalidMinute
	}

	for _, slot := range w.Slots {
		if slot.DayOfWeek == dayOfWeek && slot.Hour == hour && slot.Minute == minute {
			return nil
		}
	}

	w.Slots = append(w.Slots, TimeSlot{DayOfWeek: dayOfWeek, Hour: hour, Minute: minute})
	return nil
}

func (w *WeeklyPostSchedule) RemoveSlot(dayOfWeek time.Weekday, hour, minute int) error {
	if dayOfWeek < time.Sunday || dayOfWeek > time.Saturday {
		return ErrInvalidDayOfWeek
	}

	if hour < 0 || hour > 23 {
		return ErrInvalidHour
	}
	if minute < 0 || minute > 59 {
		return ErrInvalidMinute
	}

	for i, slot := range w.Slots {
		if slot.DayOfWeek == dayOfWeek && slot.Hour == hour && slot.Minute == minute {
			w.Slots = append(w.Slots[:i], w.Slots[i+1:]...)
			return nil
		}
	}

	return nil
}
