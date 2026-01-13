package lab

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

type Event struct {
	Type       Type
	Topic      Topic
	Number     int
	Auditorium int
	Schedule   map[DayOfWeek]map[int][]string
}
