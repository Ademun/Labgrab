package lab_polling

import (
	"labgrab/internal/shared/domain"
	"math"
	"time"
)

type LessonSchedule struct {
	Start, End time.Time
}

var LessonLookup = map[int]LessonSchedule{
	1: {parseLessonTime("08:50"), parseLessonTime("10:20")},
	2: {parseLessonTime("10:35"), parseLessonTime("12:05")},
	3: {parseLessonTime("12:35"), parseLessonTime("14:05")},
	4: {parseLessonTime("14:15"), parseLessonTime("15:45")},
	5: {parseLessonTime("15:55"), parseLessonTime("17:20")},
	6: {parseLessonTime("17:30"), parseLessonTime("19:00")},
	7: {parseLessonTime("19:10"), parseLessonTime("20:30")},
	8: {parseLessonTime("20:40"), parseLessonTime("22:00")},
}

func localTimeToLesson(lTime time.Time) domain.Lesson {
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
			return domain.Lesson(lessonNum)
		}
	}

	return 0
}

func parseLessonTime(s string) time.Time {
	t, _ := time.Parse("15:04", s)
	return t
}
