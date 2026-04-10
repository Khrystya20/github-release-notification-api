package model

import "time"

type Repository struct {
	ID            int64
	FullName      string
	Owner         string
	Name          string
	LastSeenTag   *string
	LastCheckedAt *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
