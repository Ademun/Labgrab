package record

import (
	"labgrab/internal/shared/domain"
	"time"

	"github.com/google/uuid"
)

type Status string

const (
	StatusOpen   Status = "Open"
	StatusClosed Status = "Closed"
)

// DBRecord record_service.records
type DBRecord struct {
	RecordID      int             `db:"record_id"`
	LabType       domain.LabType  `db:"lab_type"`
	LabTopic      domain.LabTopic `db:"lab_topic"`
	LabAuditorium int             `db:"lab_auditorium"`
	Lesson        int             `db:"lesson"`
	StartTime     time.Time       `db:"start_time"`
	EndTime       time.Time       `db:"end_time"`
	Status        Status          `db:"status"`
	UserUUID      uuid.UUID       `db:"user_uuid"`
}

type DBSlotFilter struct {
	UserUUID          uuid.UUID
	MatchingTimeslots domain.Schedule
}
