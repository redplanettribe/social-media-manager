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
	ErrIvalidTimeZone   = errors.New("invalid time zone")
)

type TimeSlot struct {
	DayOfWeek time.Weekday `json:"day_of_week"`
	Hour      int          `json:"hour"`
	Minute    int          `json:"minute"`
}

type WeeklyPostSchedule struct {
	TimeZone   string        `json:"time_zone"`
	Slots      []TimeSlot    `json:"slots"`
	TimeMargin time.Duration `json:"time_margin"`
}

// NewWeeklyPostSchedule creates a new WeeklyPostSchedule.
func NewWeeklyPostSchedule(timeZone string, slots []TimeSlot) *WeeklyPostSchedule {
	return &WeeklyPostSchedule{
		TimeZone:   timeZone,
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

// Decode populates a WeeklyPostSchedule from a JSON string.
func Decode(encoded string) (*WeeklyPostSchedule, error) {
	var schedule WeeklyPostSchedule
	err := json.Unmarshal([]byte(encoded), &schedule)
	if err != nil {
		return nil, err
	}
	return &schedule, nil
}

// IsTime checks if the given time matches any scheduled slot.
func (w *WeeklyPostSchedule) IsTime(t time.Time) bool {
	loc, err := time.LoadLocation(w.TimeZone)
	if err != nil {
		return false
	}
	localT := t.In(loc)
	for _, slot := range w.Slots {
		slotTime := time.Date(localT.Year(), localT.Month(), localT.Day(), slot.Hour, slot.Minute, 0, 0, loc)
		if localT.Weekday() == slot.DayOfWeek &&
			localT.After(slotTime.Add(-w.TimeMargin)) &&
			localT.Before(slotTime.Add(w.TimeMargin)) {
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

	// validate hour and minute
	if hour < 0 || hour > 23 {
		return ErrInvalidHour
	}
	if minute < 0 || minute > 59 {
		return ErrInvalidMinute
	}

	// make sure the slot doesn't already exist
	for _, slot := range w.Slots {
		if slot.DayOfWeek == dayOfWeek && slot.Hour == hour && slot.Minute == minute {
			return nil
		}
	}

	w.Slots = append(w.Slots, TimeSlot{DayOfWeek: dayOfWeek, Hour: hour, Minute: minute})
	return nil
}

// SetTimeZone sets the timezone of the schedule
func (w *WeeklyPostSchedule) SetTimeZone(timeZone string) error {
	_, err := time.LoadLocation(timeZone)
	if err != nil {
		return errors.Join(ErrIvalidTimeZone, err)
	}

	w.TimeZone = timeZone
	return nil
}
