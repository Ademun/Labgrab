package dto

import (
	"labgrab/internal/shared/domain"
	"labgrab/internal/shared/types"
)

// GetTimePreferncesResDTO represents response for getting time preferences
type GetTimePreferncesResDTO struct {
	Preferences map[int]map[types.DayOfWeek][]domain.Lesson `json:"preferences"`
}

// GetTeacherPreferencesResDTO represents response for getting teacher preferences
type GetTeacherPreferencesResDTO struct {
	Preferences []string `json:"preferences"`
}

// SetTimePreferncesReqDTO represents request for setting time preferences
type SetTimePreferncesReqDTO struct {
	Preferences map[int]map[types.DayOfWeek][]domain.Lesson `json:"preferences"`
}

// SetTeacherPreferencesReqDTO represents request for setting teacher preferences
type SetTeacherPreferencesReqDTO struct {
	Preferences []string `json:"preferences"`
}
