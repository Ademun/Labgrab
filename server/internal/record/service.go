package record

import (
	"context"
	"labgrab/internal/shared/domain"
	"labgrab/internal/shared/errors"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
)

var tracer = otel.Tracer("record-service")

type Service struct {
	repo   *Repo
	logger *zap.SugaredLogger
}

func NewService(repo *Repo, logger *zap.SugaredLogger) *Service {
	return &Service{repo: repo, logger: logger}
}

func (s *Service) CreateRecord(ctx context.Context, req *CreateRecordReq) error {
	ctx, span := tracer.Start(ctx, "record_service.create_record")
	defer span.End()

	span.SetAttributes(
		attribute.String("user.uuid", req.UserUUID.String()),
		attribute.String("lab.type", string(req.LabType)),
		attribute.String("lab.topic", string(req.LabTopic)),
		attribute.Int("lab.auditorium", req.LabAuditorium),
		attribute.Int("lab.lesson", req.Lesson),
		attribute.Int("record.id", req.RecordID),
	)

	dbRec := &DBRecord{
		RecordID:      req.RecordID,
		LabType:       req.LabType,
		LabTopic:      req.LabTopic,
		LabAuditorium: req.LabAuditorium,
		Lesson:        req.Lesson,
		StartTime:     req.StartTime,
		EndTime:       req.EndTime,
		UserUUID:      req.UserUUID,
	}

	if err := s.repo.CreateRecord(ctx, dbRec); err != nil {
		err = &errors.ErrServiceProcedure{
			Procedure: "CreateRecord",
			Step:      "Repository call",
			Err:       err,
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to create record in repository")
		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

func (s *Service) GetRecords(ctx context.Context, userUUID uuid.UUID) ([]GetRecordsRes, error) {
	ctx, span := tracer.Start(ctx, "record_service.get_records")
	defer span.End()

	span.SetAttributes(attribute.String("user.uuid", userUUID.String()))

	recs, err := s.repo.GetRecords(ctx, userUUID)
	if err != nil {
		err = &errors.ErrServiceProcedure{
			Procedure: "GetRecords",
			Step:      "Repository call",
			Err:       err,
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to retrieve records from repository")
		return nil, err
	}

	result := make([]GetRecordsRes, len(recs))
	for i, rec := range recs {
		result[i] = GetRecordsRes{
			RecordID:      rec.RecordID,
			LabType:       rec.LabType,
			LabTopic:      rec.LabTopic,
			LabAuditorium: rec.LabAuditorium,
			Lesson:        rec.Lesson,
			StartTime:     rec.StartTime,
			EndTime:       rec.EndTime,
			Status:        rec.Status,
			UserUUID:      rec.UserUUID,
		}
	}

	span.SetAttributes(attribute.Int("records.count", len(result)))
	span.SetStatus(codes.Ok, "")
	return result, nil
}

func (s *Service) DeleteRecord(ctx context.Context, recordID int) error {
	ctx, span := tracer.Start(ctx, "record_service.delete_record")
	defer span.End()

	span.SetAttributes(attribute.Int("record.id", recordID))

	if err := s.repo.DeleteRecord(ctx, recordID); err != nil {
		err = &errors.ErrServiceProcedure{
			Procedure: "DeleteRecord",
			Step:      "Repository call",
			Err:       err,
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to delete record in repository")
		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

func (s *Service) CloseExpiredRecords(ctx context.Context) error {
	ctx, span := tracer.Start(ctx, "record_service.close_expired_records")
	defer span.End()

	if err := s.repo.CloseExpiredRecords(ctx); err != nil {
		err = &errors.ErrServiceProcedure{
			Procedure: "CloseExpiredRecords",
			Step:      "Repository call",
			Err:       err,
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to close expired records in repository")
		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

func (s *Service) FilterAvailableSlots(ctx context.Context, req *FilterSlotsReq) (domain.Schedule, error) {
	ctx, span := tracer.Start(ctx, "record_service.filter_available_slots")
	defer span.End()

	span.SetAttributes(
		attribute.String("user.uuid", req.UserUUID.String()),
		attribute.Int("slots.input_count", len(req.MatchingTimeslots)),
	)

	filter := &DBSlotFilter{
		UserUUID:          req.UserUUID,
		MatchingTimeslots: req.MatchingTimeslots,
	}

	schedule, err := s.repo.FilterAvailableSlots(ctx, filter)
	if err != nil {
		err = &errors.ErrServiceProcedure{
			Procedure: "FilterAvailableSlots",
			Step:      "Repository call",
			Err:       err,
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to filter available slots in repository")
		return nil, err
	}

	span.SetAttributes(attribute.Int("slots.output_count", len(schedule)))
	span.SetStatus(codes.Ok, "")
	return schedule, nil
}
