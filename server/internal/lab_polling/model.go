package lab_polling

import (
	"labgrab/internal/shared/domain"
)

type Event struct {
	Name       string
	Type       domain.LabType
	Topic      domain.LabTopic
	Number     int
	Auditorium int
	Spot       *int
	Schedule   domain.Schedule
}

type EventResult struct {
	Data *Event
	Err  error
}
