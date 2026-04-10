package postgres

import (
	"database/sql"
	"errors"
	"log"

	"github-release-notification-api/internal/model"
)

type RepositoryRepository struct {
	db *sql.DB
}

func NewRepositoryRepository(db *sql.DB) *RepositoryRepository {
	return &RepositoryRepository{db: db}
}

func (r *RepositoryRepository) GetByFullName(fullName string) (*model.Repository, error) {
	query := `
		SELECT id, full_name, owner, name, last_seen_tag, last_checked_at, created_at, updated_at
		FROM repositories
		WHERE full_name = $1
	`

	var repo model.Repository
	err := r.db.QueryRow(query, fullName).Scan(
		&repo.ID,
		&repo.FullName,
		&repo.Owner,
		&repo.Name,
		&repo.LastSeenTag,
		&repo.LastCheckedAt,
		&repo.CreatedAt,
		&repo.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &repo, nil
}

func (r *RepositoryRepository) GetByID(id int64) (*model.Repository, error) {
	query := `
		SELECT id, full_name, owner, name, last_seen_tag, last_checked_at, created_at, updated_at
		FROM repositories
		WHERE id = $1
	`

	var repo model.Repository
	err := r.db.QueryRow(query, id).Scan(
		&repo.ID,
		&repo.FullName,
		&repo.Owner,
		&repo.Name,
		&repo.LastSeenTag,
		&repo.LastCheckedAt,
		&repo.CreatedAt,
		&repo.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &repo, nil
}

func (r *RepositoryRepository) Create(repo *model.Repository) (*model.Repository, error) {
	query := `
		INSERT INTO repositories (full_name, owner, name, last_seen_tag)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRow(query, repo.FullName, repo.Owner, repo.Name, repo.LastSeenTag).Scan(
		&repo.ID,
		&repo.CreatedAt,
		&repo.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return repo, nil
}

func (r *RepositoryRepository) UpdateLastSeenTag(repositoryID int64, tag *string) error {
	query := `
		UPDATE repositories
		SET last_seen_tag = $1, updated_at = NOW()
		WHERE id = $2
	`

	_, err := r.db.Exec(query, tag, repositoryID)
	return err
}

func (r *RepositoryRepository) GetTrackedRepositories() ([]model.Repository, error) {
	query := `
		SELECT DISTINCT r.id, r.full_name, r.owner, r.name, r.last_seen_tag, r.last_checked_at, r.created_at, r.updated_at
		FROM repositories r
		JOIN subscriptions s ON s.repository_id = r.id
		WHERE s.active = TRUE AND s.confirmed = TRUE
		ORDER BY r.full_name
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("failed to close rows: %v", err)
		}
	}()

	var repos []model.Repository
	for rows.Next() {
		var repo model.Repository
		if err := rows.Scan(
			&repo.ID,
			&repo.FullName,
			&repo.Owner,
			&repo.Name,
			&repo.LastSeenTag,
			&repo.LastCheckedAt,
			&repo.CreatedAt,
			&repo.UpdatedAt,
		); err != nil {
			return nil, err
		}
		repos = append(repos, repo)
	}

	return repos, rows.Err()
}

func (r *RepositoryRepository) UpdateLastCheckedAt(repositoryID int64) error {
	query := `
		UPDATE repositories
		SET last_checked_at = NOW(), updated_at = NOW()
		WHERE id = $1
	`

	_, err := r.db.Exec(query, repositoryID)
	return err
}
