package domain

import "time"

type Lesson int

type LabType string

const (
	LabTypeDefence     LabType = "Defence"
	LabTypePerformance LabType = "Performance"
)

func (lt LabType) RU() string {
	switch lt {
	case LabTypeDefence:
		return "Защита"
	case LabTypePerformance:
		return "Выполнение"
	}
	return string(lt)
}

type LabTopic string

const (
	LabTopicVirtual     LabTopic = "Virtual"
	LabTopicElectricity LabTopic = "Electricity"
	LabTopicMechanics   LabTopic = "Mechanics"
	LabTopicOptics      LabTopic = "Optics"
	LabTopicsRigidBody  LabTopic = "Rigid Body"
)

func (lt LabTopic) RU() string {
	switch lt {
	case LabTopicVirtual:
		return "Виртуальная"
	case LabTopicElectricity:
		return "Электричество"
	case LabTopicMechanics:
		return "Механика"
	case LabTopicOptics:
		return "Оптика"
	case LabTopicsRigidBody:
		return "Твёрдое тело"
	}
	return string(lt)
}

func (lt LabTopic) Icon() string {
	switch lt {
	case LabTopicVirtual:
		return "💻"
	case LabTopicElectricity:
		return "⚡️"
	case LabTopicMechanics:
		return "⚙️"
	case LabTopicOptics:
		return "👁️"
	case LabTopicsRigidBody:
		return "💎"
	}
	return ""
}

type Schedule map[time.Time]map[Lesson][]string

type Service struct {
	ID int
}

type Event struct {
	ID         int
	ServiceID  int
	Name       string
	Type       LabType
	Topic      LabTopic
	Number     int
	Auditorium int
	Spot       *int
	Schedule   Schedule
	Link       string
}

type Booking struct {
	ID         int
	Name       string
	Type       LabType
	Topic      LabTopic
	Number     int
	Auditorium int
	Spot       *int
	Lesson     Lesson
	Start      time.Time
	End        time.Time
}
