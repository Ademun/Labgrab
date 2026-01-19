package lab_polling

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
	TopicSolidBody   Topic = "SolidBody"
)

type DayOfWeek string

const (
	DayMon DayOfWeek = "MON"
	DayTue DayOfWeek = "TUE"
	DayWed DayOfWeek = "WED"
	DayThu DayOfWeek = "THU"
	DayFri DayOfWeek = "FRI"
	DaySat DayOfWeek = "SAT"
	DaySun DayOfWeek = "SUN"
)

type Lesson int

type Teacher string

type Schedule map[DayOfWeek]map[Lesson][]Teacher

type Event struct {
	Name       string
	Type       Type
	Topic      Topic
	Number     int
	Auditorium int
	Spot       *int
	Schedule   Schedule
}
