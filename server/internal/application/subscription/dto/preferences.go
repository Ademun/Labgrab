package dto

import (
	"labgrab/internal/shared/domain"
	"labgrab/internal/shared/types"
)

type GetTimePreferencesResDTO struct {
	Preferences map[int]map[types.DayOfWeek][]domain.Lesson `json:"preferences"`
}

type SetTimePreferencesReqDTO struct {
	Preferences map[int]map[types.DayOfWeek][]domain.Lesson `json:"preferences"`
}

type GetTeacherPreferencesResDTO struct {
	Preferences []string `json:"preferences"`
}

type SetTeacherPreferencesReqDTO struct {
	Preferences []string `json:"preferences"`
}
