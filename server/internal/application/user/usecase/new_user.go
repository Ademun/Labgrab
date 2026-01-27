package usecase

import (
	"context"
	"fmt"
	"labgrab/internal/application/user/dto"
	"labgrab/internal/shared/types"
	"labgrab/internal/subscription"
	"labgrab/internal/user"

	"github.com/jackc/pgx/v5/pgxpool"
)

type NewUserUseCase struct {
	userSvc         *user.Service
	subscriptionSvc *subscription.Service
	pool            *pgxpool.Pool
}

func NewNewUserUseCase(userSvc *user.Service, subscriptionSvc *subscription.Service, pool *pgxpool.Pool) *NewUserUseCase {
	return &NewUserUseCase{
		userSvc:         userSvc,
		subscriptionSvc: subscriptionSvc,
		pool:            pool,
	}
}

func (uc *NewUserUseCase) Exec(ctx context.Context, data *dto.NewUserReqDTO) (*dto.NewUserRespDTO, error) {
	tx, err := uc.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("error starting transaction: %v", err)
	}

	userReq := &user.CreateUserReq{
		Name:        data.Name,
		Surname:     data.Surname,
		Patronymic:  data.Patronymic,
		GroupCode:   data.GroupCode,
		PhoneNumber: data.PhoneNumber,
		TelegramID:  data.TelegramID,
	}

	userUUID, err := uc.userSvc.CreateUser(ctx, userReq)
	if err != nil {
		if err := tx.Rollback(ctx); err != nil {
			return nil, fmt.Errorf("error rolling back transaction: %v", err)
		}
		return nil, err
	}

	subReq := &subscription.CreateSubscriptionDataReq{
		UserUUID:            userUUID,
		TimePreferences:     make(map[types.DayOfWeek][]int),
		BlacklistedTeachers: make([]string, 0),
	}

	if err := uc.subscriptionSvc.CreateSubscriptionData(ctx, subReq); err != nil {
		if err := tx.Rollback(ctx); err != nil {
			return nil, fmt.Errorf("error rolling back transaction: %v", err)
		}
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("error committing transaction: %v", err)
	}

	return &dto.NewUserRespDTO{
		UserUUID: userUUID.String(),
	}, nil
}
