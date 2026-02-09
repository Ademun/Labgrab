package domain

import "time"

type Lesson int

type LabType string

const (
	LabTypeDefence     LabType = "Defence"
	LabTypePerformance LabType = "Performance"
)

type LabTopic string

const (
	LabTopicVirtual     LabTopic = "Virtual"
	LabTopicElectricity LabTopic = "Electricity"
	LabTopicMechanics   LabTopic = "Mechanics"
	LabTopicOptics      LabTopic = "Optics"
	LabTopicsRigidBody  LabTopic = "Rigid Body"
)

type Schedule map[time.Time]map[Lesson][]string
