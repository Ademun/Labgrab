package user

import (
	"context"
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

func (uc *CreateUserUseCase) Exec(ctx context.Context, req *CreateUserReqDTO) error {
	userReq := &user.CreateUserReq{
		Name:        req.Name,
		Surname:     req.Surname,
		Patronymic:  req.Patronymic,
		GroupCode:   req.GroupCode,
		PhoneNumber: req.PhoneNumber,
		Email:       req.Email,
		TelegramID:  req.TelegramID,
	}

	res, err := uc.userSvc.CreateUser(ctx, userReq)
	if err != nil {
		return err
	}

	subReq := &subscription.CreateSubscriptionDataReq{
		UserUUID: res.UUID,
	}

	err = uc.subscriptionSvc.CreateSubscriptionData(ctx, res.Tx, subReq)
	if err != nil {
		return err
	}

	return nil
}
