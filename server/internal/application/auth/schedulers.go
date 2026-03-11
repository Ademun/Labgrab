package auth

import (
	"context"
	"labgrab/internal/application/auth/usecase"
	"labgrab/internal/auth"
	"time"

	"github.com/go-co-op/gocron/v2"
	"go.uber.org/zap"
)

type Scheduler struct {
	scheduler      gocron.Scheduler
	authStaleUsers *usecase.AuthStaleUsersUsecase
	logger         *zap.SugaredLogger
}

func NewScheduler(
	authSvc *auth.Service,
	logger *zap.SugaredLogger,
) *Scheduler {
	return &Scheduler{
		authStaleUsers: &usecase.AuthStaleUsersUsecase{
			AuthSvc: authSvc,
		},
		logger: logger,
	}
}

func (s *Scheduler) Start(ctx context.Context) error {
	scheduler, err := gocron.NewScheduler()
	if err != nil {
		return err
	}

	_, err = scheduler.NewJob(
		gocron.DurationJob(time.Minute*10),
		gocron.NewTask(s.AuthStaleUsers, ctx),
		gocron.WithSingletonMode(gocron.LimitModeReschedule),
	)
	if err != nil {
		return err
	}

	scheduler.Start()
	s.scheduler = scheduler
	return nil
}

func (s *Scheduler) Stop() error {
	return s.scheduler.Shutdown()
}

func (s *Scheduler) AuthStaleUsers(ctx context.Context) {
	if err := s.authStaleUsers.Exec(ctx); err != nil {
		s.logger.Errorf("auth scheduler: auth stale users: %v", err)
		return
	}
}
