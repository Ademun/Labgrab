package usecase

import (
	"context"
	"labgrab/internal/application/user/dto"
	"labgrab/internal/subscription"
	"labgrab/internal/user"
)

type NewUserUseCase struct {
	userSvc         *user.Service
	subscriptionSvc *subscription.Service
}

func NewNewUserUseCase(userSvc *user.Service) *NewUserUseCase {
	return &NewUserUseCase{userSvc: userSvc}
}

func (uc *NewUserUseCase) Exec(ctx context.Context, data *dto.NewUserReqDTO) (*dto.NewUserRespDTO, error) {
	userReq := &user.CreateUserReq{
		Name:        data.Name,
		Surname:     data.Surname,
		Patronymic:  data.Patronymic,
		GroupCode:   data.GroupCode,
		PhoneNumber: data.PhoneNumber,
		TelegramID:  data.TelegramID,
	}

	userResp, err := uc.userSvc.CreateUser(ctx, userReq)
	if err != nil {
		return nil, err
	}

	subReq := &subscription.CreateSubscriptionDataReq{
		UserUUID:            userResp.UUID,
		TimePreferences:     make(map[subscription.DayOfWeek][]int),
		BlacklistedTeachers: make([]string, 0),
	}

	if err := uc.subscriptionSvc.CreateSubscriptionData(ctx, userResp.Tx, subReq); err != nil {
		return nil, err
	}

	return &dto.NewUserRespDTO{
		UserUUID: userResp.UUID.String(),
	}, nil
}
