package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/VarvaraKurakova/subscription-aggregator-api/internal/domain"
)

var (
	ErrInvalidInput = errors.New("invalid input")
)

type SubscriptionRepository interface {
	Create(ctx context.Context, sub domain.Subscription) (domain.Subscription, error)
	GetByID(ctx context.Context, id int64) (domain.Subscription, error)
	List(ctx context.Context, filter domain.SubscriptionFilter) ([]domain.Subscription, error)
	Update(ctx context.Context, sub domain.Subscription) (domain.Subscription, error)
	Delete(ctx context.Context, id int64) error
	ListForTotal(ctx context.Context, filter domain.TotalFilter) ([]domain.Subscription, error)
}

type SubscriptionService struct {
	repo SubscriptionRepository
}

func NewSubscriptionService(repo SubscriptionRepository) *SubscriptionService {
	return &SubscriptionService{
		repo: repo,
	}
}

type CreateSubscriptionInput struct {
	ServiceName string
	Price       int
	UserID      string
	StartDate   string
	EndDate     *string
}

type UpdateSubscriptionInput struct {
	ID          int64
	ServiceName string
	Price       int
	UserID      string
	StartDate   string
	EndDate     *string
}

type ListSubscriptionsInput struct {
	UserID      *string
	ServiceName *string
	Limit       int
	Offset      int
}

type TotalSubscriptionsInput struct {
	From        string
	To          string
	UserID      *string
	ServiceName *string
}

type TotalSubscriptionsResult struct {
	Total       int
	PeriodFrom string
	PeriodTo   string
	UserID      *string
	ServiceName *string
}

func (s *SubscriptionService) Create(ctx context.Context, input CreateSubscriptionInput) (domain.Subscription, error) {
	sub, err := buildSubscriptionFromInput(
		input.ServiceName,
		input.Price,
		input.UserID,
		input.StartDate,
		input.EndDate,
	)
	if err != nil {
		return domain.Subscription{}, err
	}

	return s.repo.Create(ctx, sub)
}

func (s *SubscriptionService) GetByID(ctx context.Context, id int64) (domain.Subscription, error) {
	if id <= 0 {
		return domain.Subscription{}, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	return s.repo.GetByID(ctx, id)
}

func (s *SubscriptionService) List(ctx context.Context, input ListSubscriptionsInput) ([]domain.Subscription, error) {
	filter := domain.SubscriptionFilter{
		Limit:  input.Limit,
		Offset: input.Offset,
	}

	if input.UserID != nil && strings.TrimSpace(*input.UserID) != "" {
		parsedUserID, err := uuid.Parse(strings.TrimSpace(*input.UserID))
		if err != nil {
			return nil, fmt.Errorf("%w: invalid user_id", ErrInvalidInput)
		}

		filter.UserID = &parsedUserID
	}

	if input.ServiceName != nil && strings.TrimSpace(*input.ServiceName) != "" {
		serviceName := strings.TrimSpace(*input.ServiceName)
		filter.ServiceName = &serviceName
	}

	return s.repo.List(ctx, filter)
}

func (s *SubscriptionService) Update(ctx context.Context, input UpdateSubscriptionInput) (domain.Subscription, error) {
	if input.ID <= 0 {
		return domain.Subscription{}, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	sub, err := buildSubscriptionFromInput(
		input.ServiceName,
		input.Price,
		input.UserID,
		input.StartDate,
		input.EndDate,
	)
	if err != nil {
		return domain.Subscription{}, err
	}

	sub.ID = input.ID

	return s.repo.Update(ctx, sub)
}

func (s *SubscriptionService) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	return s.repo.Delete(ctx, id)
}

func (s *SubscriptionService) GetTotal(ctx context.Context, input TotalSubscriptionsInput) (TotalSubscriptionsResult, error) {
	from, err := domain.ParseMonthYear(strings.TrimSpace(input.From))
	if err != nil {
		return TotalSubscriptionsResult{}, fmt.Errorf("%w: invalid from date", ErrInvalidInput)
	}

	to, err := domain.ParseMonthYear(strings.TrimSpace(input.To))
	if err != nil {
		return TotalSubscriptionsResult{}, fmt.Errorf("%w: invalid to date", ErrInvalidInput)
	}

	if to.Before(from) {
		return TotalSubscriptionsResult{}, fmt.Errorf("%w: to date must be greater than or equal to from date", ErrInvalidInput)
	}

	filter := domain.TotalFilter{
		From: from,
		To:   to,
	}

	if input.UserID != nil && strings.TrimSpace(*input.UserID) != "" {
		parsedUserID, err := uuid.Parse(strings.TrimSpace(*input.UserID))
		if err != nil {
			return TotalSubscriptionsResult{}, fmt.Errorf("%w: invalid user_id", ErrInvalidInput)
		}

		filter.UserID = &parsedUserID
	}

	if input.ServiceName != nil && strings.TrimSpace(*input.ServiceName) != "" {
		serviceName := strings.TrimSpace(*input.ServiceName)
		filter.ServiceName = &serviceName
	}

	subscriptions, err := s.repo.ListForTotal(ctx, filter)
	if err != nil {
		return TotalSubscriptionsResult{}, err
	}

	total := calculateTotal(subscriptions, from, to)

	return TotalSubscriptionsResult{
		Total:       total,
		PeriodFrom: domain.FormatMonthYear(from),
		PeriodTo:   domain.FormatMonthYear(to),
		UserID:      input.UserID,
		ServiceName: input.ServiceName,
	}, nil
}

func buildSubscriptionFromInput(
	serviceName string,
	price int,
	userID string,
	startDate string,
	endDate *string,
) (domain.Subscription, error) {
	serviceName = strings.TrimSpace(serviceName)
	if serviceName == "" {
		return domain.Subscription{}, fmt.Errorf("%w: service_name is required", ErrInvalidInput)
	}

	if price <= 0 {
		return domain.Subscription{}, fmt.Errorf("%w: price must be positive", ErrInvalidInput)
	}

	parsedUserID, err := uuid.Parse(strings.TrimSpace(userID))
	if err != nil {
		return domain.Subscription{}, fmt.Errorf("%w: invalid user_id", ErrInvalidInput)
	}

	parsedStartDate, err := domain.ParseMonthYear(strings.TrimSpace(startDate))
	if err != nil {
		return domain.Subscription{}, fmt.Errorf("%w: invalid start_date", ErrInvalidInput)
	}

	var parsedEndDate *time.Time
	if endDate != nil && strings.TrimSpace(*endDate) != "" {
		value, err := domain.ParseMonthYear(strings.TrimSpace(*endDate))
		if err != nil {
			return domain.Subscription{}, fmt.Errorf("%w: invalid end_date", ErrInvalidInput)
		}

		if value.Before(parsedStartDate) {
			return domain.Subscription{}, fmt.Errorf("%w: end_date must be greater than or equal to start_date", ErrInvalidInput)
		}

		parsedEndDate = &value
	}

	return domain.Subscription{
		ServiceName: serviceName,
		Price:       price,
		UserID:      parsedUserID,
		StartDate:   parsedStartDate,
		EndDate:     parsedEndDate,
	}, nil
}

func calculateTotal(subscriptions []domain.Subscription, from, to time.Time) int {
	total := 0

	for _, sub := range subscriptions {
		effectiveStart := domain.MaxMonth(sub.StartDate, from)

		effectiveEnd := to
		if sub.EndDate != nil {
			effectiveEnd = domain.MinMonth(*sub.EndDate, to)
		}

		if effectiveEnd.Before(effectiveStart) {
			continue
		}

		months := domain.MonthsBetweenInclusive(effectiveStart, effectiveEnd)
		total += sub.Price * months
	}

	return total
}