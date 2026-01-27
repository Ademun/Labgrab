package lab_polling

import "labgrab/internal/shared/types"

type Type string

const (
	TypeDefence     Type = "Defence"
	TypePerformance Type = "Performance"
)

type Topic string

const (
	TopicVirtual     Topic = "Virtual"
	TopicElectricity Topic = "Electricity"
	TopicMechanics   Topic = "Mechanics"
	TopicOptics      Topic = "Optics"
	TopicRigidBody   Topic = "Rigid Body"
)

type Event struct {
	Name       string
	Type       Type
	Topic      Topic
	Number     int
	Auditorium int
	Spot       *int
	Schedule   map[types.DayOfWeek]map[int][]string
}
