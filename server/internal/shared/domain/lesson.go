package domain

import (
	"math"
	"time"
)

type LessonSchedule struct {
	Start, End time.Time
}

var LessonLookup = map[int]LessonSchedule{
	1: {ParseLessonTime("08:50"), ParseLessonTime("10:20")},
	2: {ParseLessonTime("10:35"), ParseLessonTime("12:05")},
	3: {ParseLessonTime("12:35"), ParseLessonTime("14:05")},
	4: {ParseLessonTime("14:15"), ParseLessonTime("15:45")},
	5: {ParseLessonTime("15:55"), ParseLessonTime("17:20")},
	6: {ParseLessonTime("17:30"), ParseLessonTime("19:00")},
	7: {ParseLessonTime("19:10"), ParseLessonTime("20:30")},
	8: {ParseLessonTime("20:40"), ParseLessonTime("22:00")},
}

func LocalTimeToLesson(lTime time.Time) Lesson {
	minute := float64(lTime.Minute())
	roundedMinute := int(math.Round(minute/10.0) * 10)

	hour := lTime.Hour()
	if roundedMinute == 60 {
		hour++
		roundedMinute = 0
	}

	totalMinutes := hour*60 + roundedMinute

	for lessonNum, schedule := range LessonLookup {
		start := schedule.Start.Hour()*60 + schedule.Start.Minute()
		end := schedule.End.Hour()*60 + schedule.End.Minute()

		if totalMinutes >= start && totalMinutes <= end {
			return Lesson(lessonNum)
		}
	}

	return 0
}

func ParseLessonTime(s string) time.Time {
	t, _ := time.Parse("15:04", s)
	return t
}
