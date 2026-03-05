package telegram

import (
	"labgrab/internal/shared/domain"
	"time"
)

type NotifyEventReq struct {
	UserID        int
	LabName       string
	LabType       domain.LabType
	LabTopic      domain.LabTopic
	LabNumber     int
	LabAuditorium int
	Spot          *int
	Schedule      domain.Schedule
	PageURL       string
}

type NotifyEnrollmentReq struct {
	UserID        int
	LabName       string
	LabType       domain.LabType
	LabTopic      domain.LabTopic
	LabNumber     int
	LabAuditorium int
	Spot          *int
	Date          time.Time
	Lesson        domain.Lesson
}

var ruMonths = map[time.Month]string{
	time.January:   "января",
	time.February:  "февраля",
	time.March:     "марта",
	time.April:     "апреля",
	time.May:       "мая",
	time.June:      "июня",
	time.July:      "июля",
	time.August:    "августа",
	time.September: "сентября",
	time.October:   "октября",
	time.November:  "ноября",
	time.December:  "декабря",
}

var ruWeekdays = map[time.Weekday]string{
	time.Monday:    "понедельник",
	time.Tuesday:   "вторник",
	time.Wednesday: "среда",
	time.Thursday:  "четверг",
	time.Friday:    "пятница",
	time.Saturday:  "суббота",
	time.Sunday:    "воскресенье",
}

var lessonIcons = map[domain.Lesson]string{
	1: "1️⃣",
	2: "2️⃣",
	3: "3️⃣",
	4: "4️⃣",
	5: "5️⃣",
	6: "6️⃣",
	7: "7️⃣",
	8: "8️⃣",
}
