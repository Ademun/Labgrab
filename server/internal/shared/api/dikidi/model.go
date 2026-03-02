package dikidi

import "labgrab/internal/shared/domain"

type Service struct {
	ID int
}

type Event struct {
	ID         int
	ServiceID  int
	Name       string
	Type       domain.LabType
	Topic      domain.LabTopic
	Number     int
	Auditorium int
	Spot       *int
	Schedule   domain.Schedule
	Link       string
}
