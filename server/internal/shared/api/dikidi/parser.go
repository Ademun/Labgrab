package dikidi

import (
	"fmt"
	"labgrab/internal/shared/domain"
	"labgrab/internal/shared/errors"
	"labgrab/pkg/config"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Parser struct {
	numberRegexp     *regexp.Regexp
	auditoriumRegexp *regexp.Regexp
	spotRegexp       *regexp.Regexp
	topicRegexp      *regexp.Regexp

	namePrefix string

	topicMap    map[string]domain.LabTopic
	typeMap     map[string]domain.LabType
	defaultType domain.LabType
}

func NewParser(cfg *config.ParserConfig) (*Parser, error) {
	numberRegexp, err := regexp.Compile(cfg.NumberRegexpPattern)
	if err != nil {
		return nil, fmt.Errorf("invalid number_regexp pattern: %v", err)
	}

	auditoriumRegexp, err := regexp.Compile(cfg.AuditoriumRegexpPattern)
	if err != nil {
		return nil, fmt.Errorf("invalid auditorium_regexp pattern: %v", err)
	}

	spotRegexp, err := regexp.Compile(cfg.SpotRegexpPattern)
	if err != nil {
		return nil, fmt.Errorf("invalid spot_regexp pattern: %v", err)
	}

	topicRegexp, err := regexp.Compile(cfg.TopicRegexpPattern)
	if err != nil {
		return nil, fmt.Errorf("invalid topic_regexp pattern: %v", err)
	}

	topicMap := make(map[string]domain.LabTopic)
	for k, v := range cfg.TopicMap {
		topicMap[k] = domain.LabTopic(v)
	}

	typeMap := make(map[string]domain.LabType)
	for k, v := range cfg.TypeMap {
		typeMap[k] = domain.LabType(v)
	}

	return &Parser{
		numberRegexp:     numberRegexp,
		auditoriumRegexp: auditoriumRegexp,
		spotRegexp:       spotRegexp,
		topicRegexp:      topicRegexp,
		namePrefix:       cfg.NamePrefix,
		topicMap:         topicMap,
		typeMap:          typeMap,
		defaultType:      domain.LabType(cfg.DefaultType),
	}, nil
}

func (p *Parser) ParseServiceData(data *APIServiceData) ([]Event, error) {
	events := make([]Event, 0)
	pErrors := make([]error, 0)
	masters := data.Masters
	if len(masters) == 0 {
		return events, nil
	}

	for id, master := range masters {
		event, err := p.ParseEventInfo(master.Username, master.ServiceName)
		if err != nil {
			pErrors = append(pErrors, err)
			continue
		}

		schedule := make(domain.Schedule, 0)
		times := data.Times
		for _, timeStr := range times[id] {
			datetime, lesson, err := p.ParseTimeString(timeStr)
			if err != nil {
				pErrors = append(pErrors, err)
				continue
			}
			if _, ok := schedule[datetime]; !ok {
				schedule[datetime] = make(map[domain.Lesson][]string)
			}
			schedule[datetime][lesson] = make([]string, 0)
		}
		event.Schedule = schedule
		events = append(events, *event)
	}

	if len(pErrors) > 0 {
		return nil, &errors.ErrParsing{Errors: pErrors}
	}

	return events, nil
}

func (p *Parser) ParseRecords(data []APIRecord) ([]Booking, error) {
	bookings := make([]Booking, 0)
	pErrors := make([]error, 0)

	for _, record := range data {
		if len(record.Services) < 1 || len(record.Employees) < 1 {
			pErrors = append(pErrors, fmt.Errorf("invalid record, no services or employees"))
			continue
		}
		serviceName := record.Services[0].Name
		username := record.Employees[0].Username
		eventInfo, err := p.ParseEventInfo(username, serviceName)
		if err != nil {
			pErrors = append(pErrors, err)
			continue
		}

		bookingID, err := strconv.Atoi(record.ID)
		if err != nil {
			pErrors = append(pErrors, err)
			continue
		}

		startTime, err := time.Parse("2006-01-02 15:04:05", record.Time)
		if err != nil {
			pErrors = append(pErrors, err)
			continue
		}
		endTime, err := time.Parse("2006-01-02 15:04:05", record.TimeTo)
		if err != nil {
			pErrors = append(pErrors, err)
			continue
		}

		bookings = append(bookings, Booking{
			ID:         bookingID,
			Name:       eventInfo.Name,
			Type:       eventInfo.Type,
			Topic:      eventInfo.Topic,
			Number:     eventInfo.Number,
			Auditorium: eventInfo.Auditorium,
			Spot:       eventInfo.Spot,
			Start:      startTime,
			End:        endTime,
		})
	}

	if len(pErrors) > 0 {
		return nil, &errors.ErrParsing{Errors: pErrors}
	}

	return bookings, nil
}

func (p *Parser) ParseEventInfo(username, serviceName string) (*Event, error) {
	number, err := p.ParseEventNumber(username, serviceName)
	if err != nil {
		return nil, err
	}
	auditorium, err := p.ParseEventAuditorium(username, serviceName)
	if err != nil {
		return nil, err
	}
	spot, err := p.ParseEventSpot(username, serviceName)
	if err != nil {
		return nil, err
	}
	topic, err := p.ParseEventTopic(username, serviceName)
	if err != nil {
		return nil, err
	}
	labType := p.ParseEventType(username, serviceName)
	name := p.ParseEventName(username)

	return &Event{
		Name:       name,
		Number:     number,
		Auditorium: auditorium,
		Spot:       spot,
		Topic:      topic,
		Type:       labType,
	}, nil
}

func (p *Parser) ParseEventName(username string) string {
	name := p.numberRegexp.ReplaceAllString(username, "")
	name = p.auditoriumRegexp.ReplaceAllString(name, "")
	name = p.spotRegexp.ReplaceAllString(name, "")
	name = strings.TrimPrefix(name, p.namePrefix)
	name = strings.TrimSpace(name)
	name = strings.Join(strings.Fields(name), " ")
	return name
}

func (p *Parser) ParseEventNumber(username, serviceName string) (int, error) {
	if match := p.numberRegexp.FindStringSubmatch(username); match != nil {
		return strconv.Atoi(match[1])
	}
	if match := p.numberRegexp.FindStringSubmatch(serviceName); match != nil {
		return strconv.Atoi(match[1])
	}
	return 0, fmt.Errorf("lab number not found")
}

func (p *Parser) ParseEventAuditorium(username, serviceName string) (int, error) {
	if match := p.auditoriumRegexp.FindStringSubmatch(username); match != nil {
		return strconv.Atoi(match[1])
	}
	if match := p.auditoriumRegexp.FindStringSubmatch(serviceName); match != nil {
		return strconv.Atoi(match[1])
	}
	return 0, fmt.Errorf("lab auditorium not found")
}

func (p *Parser) ParseEventSpot(username, serviceName string) (*int, error) {
	if match := p.spotRegexp.FindStringSubmatch(username); match != nil {
		spot, err := strconv.Atoi(match[1])
		return &spot, err
	}
	if match := p.spotRegexp.FindStringSubmatch(serviceName); match != nil {
		spot, err := strconv.Atoi(match[1])
		return &spot, err
	}
	return nil, nil
}

func (p *Parser) ParseEventType(username, serviceName string) domain.LabType {
	for keyword := range p.typeMap {
		if strings.Contains(username, keyword) || strings.Contains(serviceName, keyword) {
			return p.typeMap[keyword]
		}
	}
	return p.defaultType
}

func (p *Parser) ParseEventTopic(username, serviceName string) (domain.LabTopic, error) {
	if match := p.topicRegexp.FindStringSubmatch(username); match != nil {
		if topic, ok := p.topicMap[match[1]]; ok {
			return topic, nil
		}
	}
	if match := p.topicRegexp.FindStringSubmatch(serviceName); match != nil {
		if topic, ok := p.topicMap[match[1]]; ok {
			return topic, nil
		}
	}
	return "", fmt.Errorf("topic not found")
}

func (p *Parser) ParseTimeString(timeString string) (time.Time, domain.Lesson, error) {
	datetime, err := time.Parse("2006-01-02 15:04:05", timeString)
	if err != nil {
		return time.Time{}, 0, err
	}
	lesson := domain.LocalTimeToLesson(datetime)

	return datetime, lesson, nil
}
