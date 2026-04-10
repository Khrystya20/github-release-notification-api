package postgres

import (
	"database/sql"
	"errors"
	"log"

	"github-release-notification-api/internal/model"
)

type SubscriptionRepository struct {
	db *sql.DB
}

func NewSubscriptionRepository(db *sql.DB) *SubscriptionRepository {
	return &SubscriptionRepository{db: db}
}

func (r *SubscriptionRepository) GetByEmailAndRepositoryID(email string, repositoryID int64) (*model.Subscription, error) {
	query := `
		SELECT id, email, repository_id, confirmed, active, confirm_token, unsubscribe_token,
		       confirmed_at, created_at, updated_at
		FROM subscriptions
		WHERE email = $1 AND repository_id = $2
	`

	var s model.Subscription
	err := r.db.QueryRow(query, email, repositoryID).Scan(
		&s.ID,
		&s.Email,
		&s.RepositoryID,
		&s.Confirmed,
		&s.Active,
		&s.ConfirmToken,
		&s.UnsubscribeToken,
		&s.ConfirmedAt,
		&s.CreatedAt,
		&s.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &s, nil
}

func (r *SubscriptionRepository) GetByConfirmToken(token string) (*model.Subscription, error) {
	query := `
		SELECT id, email, repository_id, confirmed, active, confirm_token, unsubscribe_token,
		       confirmed_at, created_at, updated_at
		FROM subscriptions
		WHERE confirm_token = $1
	`

	var s model.Subscription
	err := r.db.QueryRow(query, token).Scan(
		&s.ID,
		&s.Email,
		&s.RepositoryID,
		&s.Confirmed,
		&s.Active,
		&s.ConfirmToken,
		&s.UnsubscribeToken,
		&s.ConfirmedAt,
		&s.CreatedAt,
		&s.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &s, nil
}

func (r *SubscriptionRepository) GetByUnsubscribeToken(token string) (*model.Subscription, error) {
	query := `
		SELECT id, email, repository_id, confirmed, active, confirm_token, unsubscribe_token,
		       confirmed_at, created_at, updated_at
		FROM subscriptions
		WHERE unsubscribe_token = $1
	`

	var s model.Subscription
	err := r.db.QueryRow(query, token).Scan(
		&s.ID,
		&s.Email,
		&s.RepositoryID,
		&s.Confirmed,
		&s.Active,
		&s.ConfirmToken,
		&s.UnsubscribeToken,
		&s.ConfirmedAt,
		&s.CreatedAt,
		&s.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &s, nil
}

func (r *SubscriptionRepository) ConfirmByID(id int64) error {
	query := `
		UPDATE subscriptions
		SET confirmed = TRUE,
		    confirmed_at = NOW(),
		    updated_at = NOW()
		WHERE id = $1
	`

	_, err := r.db.Exec(query, id)
	return err
}

func (r *SubscriptionRepository) DeactivateByID(id int64) error {
	query := `
		UPDATE subscriptions
		SET active = FALSE,
		    updated_at = NOW()
		WHERE id = $1
	`

	_, err := r.db.Exec(query, id)
	return err
}

func (r *SubscriptionRepository) ReactivateByID(id int64, confirmToken, unsubscribeToken string) error {
	query := `
		UPDATE subscriptions
		SET active = TRUE,
		    confirmed = FALSE,
		    confirmed_at = NULL,
		    confirm_token = $1,
		    unsubscribe_token = $2,
		    updated_at = NOW()
		WHERE id = $3
	`

	_, err := r.db.Exec(query, confirmToken, unsubscribeToken, id)
	return err
}

func (r *SubscriptionRepository) Create(subscription *model.Subscription) (*model.Subscription, error) {
	query := `
		INSERT INTO subscriptions (
			email, repository_id, confirmed, active, confirm_token, unsubscribe_token
		)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRow(
		query,
		subscription.Email,
		subscription.RepositoryID,
		subscription.Confirmed,
		subscription.Active,
		subscription.ConfirmToken,
		subscription.UnsubscribeToken,
	).Scan(
		&subscription.ID,
		&subscription.CreatedAt,
		&subscription.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return subscription, nil
}

func (r *SubscriptionRepository) GetActiveByEmail(email string) ([]model.SubscriptionResponse, error) {
	query := `
		SELECT s.email, r.full_name, s.confirmed, r.last_seen_tag
		FROM subscriptions s
		JOIN repositories r ON r.id = s.repository_id
		WHERE s.email = $1 AND s.active = TRUE
		ORDER BY r.full_name
	`

	rows, err := r.db.Query(query, email)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("failed to close rows: %v", err)
		}
	}()

	var result []model.SubscriptionResponse
	for rows.Next() {
		var item model.SubscriptionResponse
		if err := rows.Scan(&item.Email, &item.Repo, &item.Confirmed, &item.LastSeenTag); err != nil {
			return nil, err
		}
		result = append(result, item)
	}

	return result, rows.Err()
}

func (r *SubscriptionRepository) GetActiveConfirmedByRepositoryID(repositoryID int64) ([]model.Subscription, error) {
	query := `
		SELECT id, email, repository_id, confirmed, active, confirm_token, unsubscribe_token,
		       confirmed_at, created_at, updated_at
		FROM subscriptions
		WHERE repository_id = $1 AND active = TRUE AND confirmed = TRUE
		ORDER BY email
	`

	rows, err := r.db.Query(query, repositoryID)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("failed to close rows: %v", err)
		}
	}()

	var result []model.Subscription
	for rows.Next() {
		var s model.Subscription
		if err := rows.Scan(
			&s.ID,
			&s.Email,
			&s.RepositoryID,
			&s.Confirmed,
			&s.Active,
			&s.ConfirmToken,
			&s.UnsubscribeToken,
			&s.ConfirmedAt,
			&s.CreatedAt,
			&s.UpdatedAt,
		); err != nil {
			return nil, err
		}
		result = append(result, s)
	}

	return result, rows.Err()
}
