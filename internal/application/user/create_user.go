package user

import (
	"labgrab/internal/subscription"
	"labgrab/internal/user"

	"go.uber.org/zap"
)

type CreateUserUseCase struct {
	userSvc         *user.Service
	subscriptionSvc *subscription.Service
	logger          *zap.SugaredLogger
}

func NewCreateUserUseCase(userSvc *user.Service, subscriptionSvc *subscription.Service, logger *zap.SugaredLogger) *CreateUserUseCase {
	return &CreateUserUseCase{
		userSvc:         userSvc,
		subscriptionSvc: subscriptionSvc,
		logger:          logger,
	}
}
