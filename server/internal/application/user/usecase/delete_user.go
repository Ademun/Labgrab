package usecase

import (
	"context"
	"fmt"
	"labgrab/internal/auth"
	"labgrab/internal/booking"
	"labgrab/internal/subscription"
	"labgrab/internal/user"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DeleteUserUsecase struct {
	AuthSvc         *auth.Service
	UserSvc         *user.Service
	SubscriptionSvc *subscription.Service
	BookingSvc      *booking.Service
	Pool            *pgxpool.Pool
}

func (uc *DeleteUserUsecase) Exec(ctx context.Context, session string) error {
	if err := uc.AuthSvc.ValidateSession(ctx, session); err != nil {
		return fmt.Errorf("user usecase: get user: validate session: %w", err)
	}

	userUUID, err := uc.AuthSvc.GetSessionData(ctx, session)
	if err != nil {
		return fmt.Errorf("user usecase: get user: get session data: %w", err)
	}

	tx, err := uc.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("user usecase: create transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := uc.AuthSvc.DeleteUserData(ctx, &auth.DeleteUserDataReq{
		UserUUID: userUUID,
		Tx:       tx,
	}); err != nil {
		return fmt.Errorf("user usecase: delete user data: %w", err)
	}

	if err := uc.SubscriptionSvc.DeleteSubscriptions(ctx, &subscription.DeleteSubscriptionsReq{
		UserUUID: userUUID,
		Tx:       tx,
	}); err != nil {
		return fmt.Errorf("user usecase: delete subscriptions: %w", err)
	}

	if err := uc.BookingSvc.DeleteBookings(ctx, &booking.DeleteBookingsReq{
		UserUUID: userUUID,
		Tx:       tx,
	}); err != nil {
		return fmt.Errorf("user usecase: delete bookings: %w", err)
	}

	if err := uc.UserSvc.DeleteUser(ctx, &user.DeleteUserReq{
		UserUUID: userUUID,
		Tx:       tx,
	}); err != nil {
		return fmt.Errorf("user usecase: delete user: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("user usecase: commit transaction: %w", err)
	}

	return nil
}
