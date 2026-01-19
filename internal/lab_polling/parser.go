package lab_polling

import (
	"fmt"
	"labgrab/internal/shared/api/dikidi"
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

	timezone *time.Location

	topicMap    map[string]Topic
	typeMap     map[string]Type
	defaultType Type
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

	timezone, err := time.LoadLocation(cfg.Timezone)
	if err != nil {
		return nil, fmt.Errorf("invalid timezone: %v", err)
	}

	topicMap := make(map[string]Topic)
	for k, v := range cfg.TopicMap {
		topicMap[k] = Topic(v)
	}

	typeMap := make(map[string]Type)
	for k, v := range cfg.TypeMap {
		typeMap[k] = Type(v)
	}

	return &Parser{
		numberRegexp:     numberRegexp,
		auditoriumRegexp: auditoriumRegexp,
		spotRegexp:       spotRegexp,
		topicRegexp:      topicRegexp,
		namePrefix:       cfg.NamePrefix,
		timezone:         timezone,
		topicMap:         topicMap,
		typeMap:          typeMap,
		defaultType:      Type(cfg.DefaultType),
	}, nil
}

func (p *Parser) ParseSlot(slot *dikidi.APISlotData) ([]Event, error) {
	events := make([]Event, 0)
	errors := make([]error, 0)
	masters := slot.Data.Masters
	if len(masters) == 0 {
		return events, nil
	}

	for id, master := range masters {
		event, err := p.parseSlotInfo(master.Username, master.ServiceName)
		if err != nil {
			errors = append(errors, err)
			continue
		}

		schedule := make(Schedule)
		times := slot.Data.Times
		for _, timeStr := range times[id] {
			dayOfWeek, lesson, err := p.parseTimeString(timeStr)
			if err != nil {
				errors = append(errors, err)
				continue
			}
			schedule[dayOfWeek][lesson] = make([]Teacher, 0)
		}
		event.Schedule = schedule
		events = append(events, *event)
	}

	if len(errors) > 0 {
		return nil, &ErrSlotParsing{errors: errors}
	}

	return events, nil
}

func (p *Parser) parseSlotInfo(username, serviceName string) (*Event, error) {
	number, err := p.parseNumber(username, serviceName)
	if err != nil {
		return nil, err
	}
	auditorium, err := p.parseAuditorium(username, serviceName)
	if err != nil {
		return nil, err
	}
	spot, err := p.parseSpot(username, serviceName)
	if err != nil {
		return nil, err
	}
	topic, err := p.parseTopic(username, serviceName)
	if err != nil {
		return nil, err
	}
	labType := p.parseType(username, serviceName)
	name := p.parseName(username)

	return &Event{
		Name:       name,
		Number:     number,
		Auditorium: auditorium,
		Spot:       spot,
		Topic:      topic,
		Type:       labType,
	}, nil
}

func (p *Parser) parseName(username string) string {
	name := p.numberRegexp.ReplaceAllString(username, "")
	name = p.auditoriumRegexp.ReplaceAllString(name, "")
	name = p.spotRegexp.ReplaceAllString(name, "")
	name = strings.TrimPrefix(name, p.namePrefix)
	name = strings.TrimSpace(name)
	name = strings.Join(strings.Fields(name), " ")
	return name
}

func (p *Parser) parseNumber(username, serviceName string) (int, error) {
	if match := p.numberRegexp.FindStringSubmatch(username); match != nil {
		return strconv.Atoi(match[1])
	}
	if match := p.numberRegexp.FindStringSubmatch(serviceName); match != nil {
		return strconv.Atoi(match[1])
	}
	return 0, fmt.Errorf("lab number not found")
}

func (p *Parser) parseAuditorium(username, serviceName string) (int, error) {
	if match := p.auditoriumRegexp.FindStringSubmatch(username); match != nil {
		return strconv.Atoi(match[1])
	}
	if match := p.auditoriumRegexp.FindStringSubmatch(serviceName); match != nil {
		return strconv.Atoi(match[1])
	}
	return 0, fmt.Errorf("lab auditorium not found")
}

func (p *Parser) parseSpot(username, serviceName string) (*int, error) {
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

func (p *Parser) parseTopic(username, serviceName string) (Topic, error) {
	if match := p.topicRegexp.FindStringSubmatch(username); match != nil {
		if topic, ok := p.topicMap[strings.ToLower(match[1])]; ok {
			return topic, nil
		}
	}
	if match := p.topicRegexp.FindStringSubmatch(serviceName); match != nil {
		if topic, ok := p.topicMap[strings.ToLower(match[1])]; ok {
			return topic, nil
		}
	}
	return "", fmt.Errorf("topic not found")
}

func (p *Parser) parseType(username, serviceName string) Type {
	for keyword := range p.typeMap {
		if strings.Contains(username, keyword) || strings.Contains(serviceName, keyword) {
			return p.typeMap[keyword]
		}
	}
	return p.defaultType
}

func (p *Parser) parseTimeString(timeString string) (DayOfWeek, Lesson, error) {
	datetime, err := time.ParseInLocation("2006-01-02 15:04:05", timeString, p.timezone)
	if err != nil {
		return DayMon, 0, err
	}
	dayOfWeek := nativeWeekdayToDayOfWeek(datetime.Weekday())
	lesson := localTimeToLesson(datetime)

	return dayOfWeek, lesson, nil
}
