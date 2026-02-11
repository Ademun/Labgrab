package lab_enrollment

import (
	"time"

	"github.com/google/uuid"
)

type JobStatus string

const (
	JobStatusQueued     JobStatus = "Queued"
	JobStatusProcessing JobStatus = "Processing"
	JobStatusCompleted  JobStatus = "Completed"
)

type DBEnrollmentJob struct {
	UUID             uuid.UUID              `db:"job_uuid"`
	UserUUID         uuid.UUID              `db:"user_uuid"`
	SubscriptionUUID uuid.UUID              `db:"subscription_uuid"`
	Status           JobStatus              `db:"job_status"`
	AvailableDates   map[time.Time][]string `db:"available_dates"`
	CreatedAt        time.Time              `db:"created_at"`
	StartedAt        time.Time              `db:"started_at"`
	CompletedAt      time.Time              `db:"completed_at"`
}

type JobResult string

const (
	JobResultSuccess JobResult = "Success"
	JobResultFailed  JobResult = "Failed"
)

type DBJobResult struct {
	JobUUID        uuid.UUID `db:"job_uuid"`
	Result         JobResult `db:"result"`
	ErrorMessage   string    `db:"error_message"`
	EnrollmentUUID uuid.UUID `db:"enrollment_uuid"`
}

type DBEnrollment struct {
	UUID               uuid.UUID `db:"enrollment_uuid"`
	UserUUID           uuid.UUID `db:"user_uuid"`
	DikidiEnrollmentID int       `db:"dikidi_enrollment_id"`
	VisitTime          time.Time `db:"visit_time"`
	EnrolledAt         time.Time `db:"enrolled_at"`
}

type DBUserData struct {
	UserUUID        uuid.UUID `db:"user_uuid"`
	Session         string    `db:"session"`
	TransportCookie string    `db:"transport_cookie"`
	Cookies         []string  `db:"cookies"`
}
