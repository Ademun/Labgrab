package usecase

import (
	"context"
	"fmt"
	"labgrab/internal/application/user/dto"
	"labgrab/internal/auth"
	"labgrab/internal/user"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuthUserUseCase struct {
	authSvc *auth.Service
	userSvc *user.Service
	pool    *pgxpool.Pool
}

func NewAuthUserUseCase(authSvc *auth.Service, userSvc *user.Service, pool *pgxpool.Pool) *AuthUserUseCase {
	return &AuthUserUseCase{
		authSvc: authSvc,
		userSvc: userSvc,
		pool:    pool,
	}
}

func (uc *AuthUserUseCase) Exec(ctx context.Context, data *dto.AuthUserReqDTO) (*dto.AuthUserRespDTO, error) {
	telegramAuthData := &auth.TelegramAuthData{
		Id:        data.Id,
		FirstName: data.FirstName,
		LastName:  data.LastName,
		Username:  data.Username,
		PhotoURL:  data.PhotoURL,
		AuthDate:  data.AuthDate,
		Hash:      data.Hash,
	}

	if err := uc.authSvc.ValidateTelegramAuthData(ctx, telegramAuthData); err != nil {
		return nil, err
	}

	exists, err := uc.userSvc.ExistsByTelegramID(ctx, data.Id)
	if err != nil {
		return nil, err
	}

	var userUUID uuid.UUID
	isNew := true
	if exists {
		isNew = false
		existingUUID, err := uc.userSvc.GetUserUUIDByTelegramID(ctx, data.Id)
		if err != nil {
			return nil, err
		}
		userUUID = existingUUID
	} else {
		tx, err := uc.pool.Begin(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to begin transaction: %v", err)
		}

		req := &user.CreateUserReq{
			TelegramID: data.Id,
			Username:   data.Username,
			Tx:         tx,
		}

		if data.FirstName != "" {
			req.Name = &data.FirstName
		}
		if data.LastName != "" {
			req.Surname = &data.LastName
		}
		if data.PhotoURL != "" {
			req.PhotoUrl = &data.PhotoURL
		}

		createdUUID, err := uc.userSvc.CreateUser(ctx, req)
		if err != nil {
			if err := tx.Rollback(ctx); err != nil {
				return nil, fmt.Errorf("failed to rollback transaction: %v", err)
			}
			return nil, err
		}
		if err := tx.Commit(ctx); err != nil {
			if err := tx.Rollback(ctx); err != nil {
				return nil, fmt.Errorf("failed to rollback transaction: %v", err)
			}
			return nil, fmt.Errorf("failed to commit transaction: %v", err)
		}

		userUUID = createdUUID
	}

	session, err := uc.authSvc.CreateSession(ctx, userUUID)

	return &dto.AuthUserRespDTO{
		Session: session,
		IsNew:   isNew,
	}, nil
}
