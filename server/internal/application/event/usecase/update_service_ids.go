package usecase

import (
	"context"
	"fmt"
	"labgrab/internal/event"
)

type UpdateServiceIDsUsecase struct {
	EventSvc *event.Service
}

func (uc *UpdateServiceIDsUsecase) Exec(ctx context.Context) error {
	var clientCookies string
	if err := uc.EventSvc.UpdateServiceIDs(ctx, &clientCookies); err != nil {
		return fmt.Errorf("event: usecase: update service ids: %w", err)
	}

	return nil
}
