package domain

import (
	"time"

	"github.com/google/uuid"
)

type Subscription struct {
	ID          int64
	ServiceName string
	Price       int
	UserID      uuid.UUID
	StartDate   time.Time
	EndDate     *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type SubscriptionFilter struct {
	UserID      *uuid.UUID
	ServiceName *string
	Limit       int
	Offset      int
}

type TotalFilter struct {
	From        time.Time
	To          time.Time
	UserID      *uuid.UUID
	ServiceName *string
}
