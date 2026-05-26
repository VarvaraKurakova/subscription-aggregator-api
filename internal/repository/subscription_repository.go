package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/VarvaraKurakova/subscription-aggregator-api/internal/domain"
)

var ErrSubscriptionNotFound = errors.New("subscription not found")

type SubscriptionRepository struct {
	pool *pgxpool.Pool
}

func NewSubscriptionRepository(pool *pgxpool.Pool) *SubscriptionRepository {
	return &SubscriptionRepository{
		pool: pool,
	}
}

func (r *SubscriptionRepository) Create(ctx context.Context, sub domain.Subscription) (domain.Subscription, error) {
	query := `
		INSERT INTO subscriptions (
			service_name,
			price,
			user_id,
			start_date,
			end_date
		)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, service_name, price, user_id, start_date, end_date, created_at, updated_at
	`

	var created domain.Subscription

	err := r.pool.QueryRow(
		ctx,
		query,
		sub.ServiceName,
		sub.Price,
		sub.UserID,
		sub.StartDate,
		sub.EndDate,
	).Scan(
		&created.ID,
		&created.ServiceName,
		&created.Price,
		&created.UserID,
		&created.StartDate,
		&created.EndDate,
		&created.CreatedAt,
		&created.UpdatedAt,
	)
	if err != nil {
		return domain.Subscription{}, fmt.Errorf("create subscription: %w", err)
	}

	return created, nil
}

func (r *SubscriptionRepository) GetByID(ctx context.Context, id int64) (domain.Subscription, error) {
	query := `
		SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at
		FROM subscriptions
		WHERE id = $1
	`

	var sub domain.Subscription

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&sub.ID,
		&sub.ServiceName,
		&sub.Price,
		&sub.UserID,
		&sub.StartDate,
		&sub.EndDate,
		&sub.CreatedAt,
		&sub.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Subscription{}, ErrSubscriptionNotFound
		}

		return domain.Subscription{}, fmt.Errorf("get subscription by id: %w", err)
	}

	return sub, nil
}

func (r *SubscriptionRepository) List(ctx context.Context, filter domain.SubscriptionFilter) ([]domain.Subscription, error) {
	query := `
		SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at
		FROM subscriptions
		WHERE ($1::uuid IS NULL OR user_id = $1)
		  AND ($2::text IS NULL OR service_name = $2)
		ORDER BY id DESC
		LIMIT $3 OFFSET $4
	`

	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	}

	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	rows, err := r.pool.Query(
		ctx,
		query,
		filter.UserID,
		filter.ServiceName,
		limit,
		offset,
	)
	if err != nil {
		return nil, fmt.Errorf("list subscriptions: %w", err)
	}
	defer rows.Close()

	subscriptions := make([]domain.Subscription, 0)

	for rows.Next() {
		var sub domain.Subscription

		err := rows.Scan(
			&sub.ID,
			&sub.ServiceName,
			&sub.Price,
			&sub.UserID,
			&sub.StartDate,
			&sub.EndDate,
			&sub.CreatedAt,
			&sub.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan subscription: %w", err)
		}

		subscriptions = append(subscriptions, sub)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate subscriptions: %w", err)
	}

	return subscriptions, nil
}

func (r *SubscriptionRepository) Update(ctx context.Context, sub domain.Subscription) (domain.Subscription, error) {
	query := `
		UPDATE subscriptions
		SET
			service_name = $1,
			price = $2,
			user_id = $3,
			start_date = $4,
			end_date = $5,
			updated_at = NOW()
		WHERE id = $6
		RETURNING id, service_name, price, user_id, start_date, end_date, created_at, updated_at
	`

	var updated domain.Subscription

	err := r.pool.QueryRow(
		ctx,
		query,
		sub.ServiceName,
		sub.Price,
		sub.UserID,
		sub.StartDate,
		sub.EndDate,
		sub.ID,
	).Scan(
		&updated.ID,
		&updated.ServiceName,
		&updated.Price,
		&updated.UserID,
		&updated.StartDate,
		&updated.EndDate,
		&updated.CreatedAt,
		&updated.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Subscription{}, ErrSubscriptionNotFound
		}

		return domain.Subscription{}, fmt.Errorf("update subscription: %w", err)
	}

	return updated, nil
}

func (r *SubscriptionRepository) Delete(ctx context.Context, id int64) error {
	query := `
		DELETE FROM subscriptions
		WHERE id = $1
	`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete subscription: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrSubscriptionNotFound
	}

	return nil
}

func (r *SubscriptionRepository) ListForTotal(ctx context.Context, filter domain.TotalFilter) ([]domain.Subscription, error) {
	query := `
		SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at
		FROM subscriptions
		WHERE start_date <= $1
		  AND (end_date IS NULL OR end_date >= $2)
		  AND ($3::uuid IS NULL OR user_id = $3)
		  AND ($4::text IS NULL OR service_name = $4)
	`

	rows, err := r.pool.Query(
		ctx,
		query,
		filter.To,
		filter.From,
		filter.UserID,
		filter.ServiceName,
	)
	if err != nil {
		return nil, fmt.Errorf("list subscriptions for total: %w", err)
	}
	defer rows.Close()

	subscriptions := make([]domain.Subscription, 0)

	for rows.Next() {
		var sub domain.Subscription

		err := rows.Scan(
			&sub.ID,
			&sub.ServiceName,
			&sub.Price,
			&sub.UserID,
			&sub.StartDate,
			&sub.EndDate,
			&sub.CreatedAt,
			&sub.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan subscription for total: %w", err)
		}

		subscriptions = append(subscriptions, sub)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate subscriptions for total: %w", err)
	}

	return subscriptions, nil
}
