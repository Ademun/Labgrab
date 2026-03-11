package dto

import (
	"labgrab/internal/shared/domain"
	"labgrab/internal/shared/types"
)

type GetTimeRestrictionsResDTO struct {
	Restrictions map[int]map[types.DayOfWeek][]domain.Lesson `json:"restrictions"`
}

type SetTimeRestrictionsReqDTO struct {
	Restrictions map[int]map[types.DayOfWeek][]domain.Lesson `json:"restrictions"`
}

type GetTeacherPreferencesResDTO struct {
	Preferences []string `json:"preferences"`
}

type SetTeacherPreferencesReqDTO struct {
	Preferences []string `json:"preferences"`
}
