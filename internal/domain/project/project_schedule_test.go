package project

import (
	"testing"
	"time"
)

func TestNewWeeklyPostSchedule(t *testing.T) {
	slots := []TimeSlot{
		{DayOfWeek: time.Monday, Hour: 10, Minute: 30},
	}
	schedule := NewWeeklyPostSchedule(slots)

	if len(schedule.Slots) != 1 {
		t.Errorf("expected 1 slot, got %d", len(schedule.Slots))
	}
	if schedule.TimeMargin != 5*time.Minute {
		t.Errorf("expected time margin of 5 minutes, got %v", schedule.TimeMargin)
	}
}

func TestEncodeDecode(t *testing.T) {
	slots := []TimeSlot{
		{DayOfWeek: time.Monday, Hour: 10, Minute: 30},
	}
	schedule := NewWeeklyPostSchedule(slots)

	encoded, err := schedule.Encode()
	if err != nil {
		t.Fatalf("failed to encode schedule: %v", err)
	}

	decoded, err := DecodeSchedule(encoded)
	if err != nil {
		t.Fatalf("failed to decode schedule: %v", err)
	}

	if len(decoded.Slots) != 1 {
		t.Errorf("expected 1 slot, got %d", len(decoded.Slots))
	}
	if decoded.Slots[0] != slots[0] {
		t.Errorf("expected slot %v, got %v", slots[0], decoded.Slots[0])
	}
}

func TestIsTime(t *testing.T) {
	slots := []TimeSlot{
		{DayOfWeek: time.Monday, Hour: 10, Minute: 30},
	}
	schedule := NewWeeklyPostSchedule(slots)

	testTime := time.Date(2023, time.October, 2, 10, 30, 0, 0, time.UTC) // Monday
	if !schedule.IsTime(testTime) {
		t.Errorf("expected IsTime to return true for %v", testTime)
	}

	testTime = testTime.Add(6 * time.Minute)
	if schedule.IsTime(testTime) {
		t.Errorf("expected IsTime to return false for %v", testTime)
	}
}

func TestAddSlot(t *testing.T) {
	schedule := NewWeeklyPostSchedule([]TimeSlot{})

	err := schedule.AddSlot(time.Monday, 10, 30)
	if err != nil {
		t.Fatalf("failed to add slot: %v", err)
	}

	if len(schedule.Slots) != 1 {
		t.Errorf("expected 1 slot, got %d", len(schedule.Slots))
	}

	err = schedule.AddSlot(time.Sunday, 25, 0)
	if err != ErrInvalidHour {
		t.Errorf("expected ErrInvalidHour, got %v", err)
	}

	err = schedule.AddSlot(time.Sunday, 10, 61)
	if err != ErrInvalidMinute {
		t.Errorf("expected ErrInvalidMinute, got %v", err)
	}

	err = schedule.AddSlot(time.Weekday(7), 10, 30)
	if err != ErrInvalidDayOfWeek {
		t.Errorf("expected ErrInvalidDayOfWeek, got %v", err)
	}
}

func TestRemoveSlot(t *testing.T) {
	slots := []TimeSlot{
		{DayOfWeek: time.Monday, Hour: 10, Minute: 30},
		{DayOfWeek: time.Tuesday, Hour: 14, Minute: 45},
	}
	schedule := NewWeeklyPostSchedule(slots)

	err := schedule.RemoveSlot(time.Monday, 10, 30)
	if err != nil {
		t.Fatalf("failed to remove slot: %v", err)
	}

	if len(schedule.Slots) != 1 {
		t.Errorf("expected 1 slot, got %d", len(schedule.Slots))
	}

	if schedule.Slots[0].DayOfWeek != time.Tuesday || schedule.Slots[0].Hour != 14 || schedule.Slots[0].Minute != 45 {
		t.Errorf("unexpected remaining slot: %v", schedule.Slots[0])
	}

	err = schedule.RemoveSlot(time.Sunday, 25, 0)
	if err != ErrInvalidHour {
		t.Errorf("expected ErrInvalidHour, got %v", err)
	}

	err = schedule.RemoveSlot(time.Sunday, 10, 61)
	if err != ErrInvalidMinute {
		t.Errorf("expected ErrInvalidMinute, got %v", err)
	}

	err = schedule.RemoveSlot(time.Weekday(7), 10, 30)
	if err != ErrInvalidDayOfWeek {
		t.Errorf("expected ErrInvalidDayOfWeek, got %v", err)
	}

	err = schedule.RemoveSlot(time.Monday, 10, 30)
	if err != nil {
		t.Errorf("expected no error for removing non-existent slot, got %v", err)
	}
}
