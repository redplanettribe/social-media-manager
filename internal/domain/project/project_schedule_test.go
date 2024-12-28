package project

import (
	"testing"
	"time"
)

func TestEncodeDecode(t *testing.T) {
	schedule := &WeeklyPostSchedule{
		TimeZone: "America/New_York",
		Slots: []TimeSlot{
			{DayOfWeek: time.Monday, Hour: 9, Minute: 30},
			{DayOfWeek: time.Wednesday, Hour: 14, Minute: 0},
		},
	}

	encoded, err := schedule.Encode()
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}

	decoded, err := Decode(encoded)
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}

	if decoded.TimeZone != schedule.TimeZone {
		t.Errorf("Decode() TimeZone = %v, want %v", decoded.TimeZone, schedule.TimeZone)
	}

	if len(decoded.Slots) != len(schedule.Slots) {
		t.Errorf("Decode() Slots length = %v, want %v", len(decoded.Slots), len(schedule.Slots))
	}

	for i, slot := range decoded.Slots {
		if slot != schedule.Slots[i] {
			t.Errorf("Decode() Slot = %v, want %v", slot, schedule.Slots[i])
		}
	}
}

func TestWeeklyPostSchedule_IsTime(t *testing.T) {
    schedule := &WeeklyPostSchedule{
        TimeZone:   "America/New_York",
        TimeMargin: 5 * time.Minute,
        Slots: []TimeSlot{
            {DayOfWeek: time.Monday, Hour: 9, Minute: 30},
        },
    }
    loc, _ := time.LoadLocation("America/New_York")

    tests := []struct {
        name    string
        testTime time.Time
        want    bool
    }{
        {
            name:    "exact time match",
            testTime: time.Date(2023, time.October, 2, 9, 30, 0, 0, loc),
            want:    true,
        },
        {
            name:    "within margin (2 min early)",
            testTime: time.Date(2023, time.October, 2, 9, 28, 0, 0, loc),
            want:    true,
        },
        {
            name:    "within margin (2 min late)",
            testTime: time.Date(2023, time.October, 2, 9, 32, 0, 0, loc),
            want:    true,
        },
        {
            name:    "outside margin (6 min early)",
            testTime: time.Date(2023, time.October, 2, 9, 24, 0, 0, loc),
            want:    false,
        },
        {
            name:    "outside margin (6 min late)",
            testTime: time.Date(2023, time.October, 2, 9, 36, 0, 0, loc),
            want:    false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := schedule.IsTime(tt.testTime)
            if got != tt.want {
                t.Errorf("IsTime(%v) = %v, want %v", tt.testTime, got, tt.want)
            }
        })
    }
}

func TestWeeklyPostSchedule_IsTime_DifferentTimeZone(t *testing.T) {
	schedule := &WeeklyPostSchedule{
		TimeZone:   "America/New_York",
		TimeMargin: 5 * time.Minute,
		Slots: []TimeSlot{
			{DayOfWeek: time.Monday, Hour: 9, Minute: 30},
		},
	}
	loc, _ := time.LoadLocation("America/Los_Angeles")

	tests := []struct {
		name    string
		testTime time.Time
		want    bool
	}{
		{
			name:    "exact time match",
			testTime: time.Date(2023, time.October, 2, 6, 30, 0, 0, loc),
			want:    true,
		},
		{
			name:    "within margin (2 min early)",
			testTime: time.Date(2023, time.October, 2, 6, 28, 0, 0, loc),
			want:    true,
		},
		{
			name:    "within margin (2 min late)",
			testTime: time.Date(2023, time.October, 2, 6, 32, 0, 0, loc),
			want:    true,
		},
		{
			name:    "outside margin (6 min early)",
			testTime: time.Date(2023, time.October, 2, 6, 24, 0, 0, loc),
			want:    false,
		},
		{
			name:    "outside margin (6 min late)",
			testTime: time.Date(2023, time.October, 2, 6, 36, 0, 0, loc),
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := schedule.IsTime(tt.testTime)
			if got != tt.want {
				t.Errorf("IsTime(%v) = %v, want %v", tt.testTime, got, tt.want)
			}
		})
	}
}

func TestWeeklyPostSchedule_IsTime_MultipleSlots(t *testing.T) {
	schedule := &WeeklyPostSchedule{
		TimeZone:   "America/New_York",
		TimeMargin: 5 * time.Minute,
		Slots: []TimeSlot{
			{DayOfWeek: time.Monday, Hour: 9, Minute: 30},
			{DayOfWeek: time.Wednesday, Hour: 14, Minute: 0},
		},
	}
	loc, _ := time.LoadLocation("America/New_York")

	tests := []struct {
		name    string
		testTime time.Time
		want    bool
	}{
		{
			name:    "exact time match slot 1",
			testTime: time.Date(2023, time.October, 2, 9, 30, 0, 0, loc),
			want:    true,
		},
		{
			name:    "exact time match slot 2",
			testTime: time.Date(2023, time.October, 4, 14, 0, 0, 0, loc),
			want:    true,
		},
		{
			name:    "within margin (2 min early) slot 1",
			testTime: time.Date(2023, time.October, 2, 9, 28, 0, 0, loc),
			want:    true,
		},
		{
			name:    "within margin (2 min late) slot 1",
			testTime: time.Date(2023, time.October, 2, 9, 32, 0, 0, loc),
			want:    true,
		},
		{
			name:    "within margin (2 min early) slot 2",
			testTime: time.Date(2023, time.October, 4, 13, 58, 0, 0, loc),
			want:    true,
		},
		{
			name:    "within margin (2 min late) slot 2",
			testTime: time.Date(2023, time.October, 4, 14, 2, 0, 0, loc),
			want:    true,
		},
		{
			name:    "outside margin (6 min early) slot 1",
			testTime: time.Date(2023, time.October, 2, 9, 24, 0, 0, loc),
			want:    false,
		},
		{
			name:    "outside margin (6 min late) slot 1",
			testTime: time.Date(2023, time.October, 2, 9, 36, 0, 0, loc),
			want:    false,
		},
		{
			name:    "outside margin (6 min early) slot 2",
			testTime: time.Date(2023, time.October, 4, 13, 54, 0, 0, loc),
			want:    false,
		},
		{
			name:    "outside margin (6 min late) slot 2",
			testTime: time.Date(2023, time.October, 4, 14, 6, 0, 0, loc),
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := schedule.IsTime(tt.testTime)
			if got != tt.want {
				t.Errorf("IsTime(%v) = %v, want %v", tt.testTime, got, tt.want)
			}
		})
	}
}

// Add slot tests
func TestWeeklyPostSchedule_AddSlot(t *testing.T) {


	tests := []struct {
		name     string
		slot     TimeSlot
		expected []TimeSlot
	}{
		{
			name: "add slot",
			slot: TimeSlot{DayOfWeek: time.Wednesday, Hour: 14, Minute: 0},
			expected: []TimeSlot{
				{DayOfWeek: time.Monday, Hour: 9, Minute: 30},
				{DayOfWeek: time.Wednesday, Hour: 14, Minute: 0},
			},
		},
		{
			name: "add slot duplicate",
			slot: TimeSlot{DayOfWeek: time.Monday, Hour: 9, Minute: 30},
			expected: []TimeSlot{
				{DayOfWeek: time.Monday, Hour: 9, Minute: 30},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schedule := &WeeklyPostSchedule{
				TimeZone:   "America/New_York",
				TimeMargin: 5 * time.Minute,
				Slots: []TimeSlot{
					{DayOfWeek: time.Monday, Hour: 9, Minute: 30},
				},
			}
			schedule.AddSlot(tt.slot.DayOfWeek, tt.slot.Hour, tt.slot.Minute)
			if len(schedule.Slots) != len(tt.expected) {
				t.Fatalf("AddSlot() Slots length = %v, want %v", len(schedule.Slots), len(tt.expected))
			}
		})
	}
}