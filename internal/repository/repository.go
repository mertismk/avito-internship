package repository

import (
	"database/sql"
	"fmt"

	"github.com/mertismk/avito-internship/internal/models"
)

type Repository struct {
	db *sql.DB
}

func New(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateTeam(teamName string) error {
	_, err := r.db.Exec(`INSERT INTO teams (team_name) VALUES ($1)`, teamName)
	return err
}

func (r *Repository) TeamExists(teamName string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(
		`SELECT EXISTS(SELECT 1 FROM teams WHERE team_name = $1)`,
		teamName,
	).Scan(&exists)
	return exists, err
}

func (r *Repository) GetTeamMembers(teamName string) ([]models.TeamMember, error) {
	rows, err := r.db.Query(`
		SELECT user_id, username, is_active 
		FROM users 
		WHERE team_name = $1
		ORDER BY user_id
	`, teamName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []models.TeamMember
	for rows.Next() {
		var m models.TeamMember
		if err := rows.Scan(&m.UserID, &m.Username, &m.IsActive); err != nil {
			return nil, err
		}
		members = append(members, m)
	}
	return members, nil
}

func (r *Repository) UpsertUser(user models.TeamMember, teamName string) error {
	_, err := r.db.Exec(`
		INSERT INTO users (user_id, username, team_name, is_active)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id) 
		DO UPDATE SET 
			username = EXCLUDED.username,
			team_name = EXCLUDED.team_name,
			is_active = EXCLUDED.is_active
	`, user.UserID, user.Username, teamName, user.IsActive)
	return err
}

func (r *Repository) GetUser(userID string) (*models.User, error) {
	var user models.User
	err := r.db.QueryRow(`
		SELECT user_id, username, team_name, is_active
		FROM users WHERE user_id = $1
	`, userID).Scan(&user.UserID, &user.Username, &user.TeamName, &user.IsActive)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *Repository) SetUserIsActive(userID string, isActive bool) error {
	result, err := r.db.Exec(
		`UPDATE users SET is_active = $1 WHERE user_id = $2`,
		isActive, userID,
	)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *Repository) GetActiveTeamMembers(teamName, excludeUserID string) ([]models.User, error) {
	rows, err := r.db.Query(`
		SELECT user_id, username, team_name, is_active
		FROM users
		WHERE team_name = $1 AND is_active = true AND user_id != $2
		ORDER BY user_id
	`, teamName, excludeUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.UserID, &u.Username, &u.TeamName, &u.IsActive); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

func (r *Repository) CreatePR(pr models.CreatePRRequest) error {
	_, err := r.db.Exec(`
		INSERT INTO pull_requests (pull_request_id, pull_request_name, author_id, status)
		VALUES ($1, $2, $3, 'OPEN')
	`, pr.PullRequestID, pr.PullRequestName, pr.AuthorID)
	return err
}

func (r *Repository) PRExists(prID string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(
		`SELECT EXISTS(SELECT 1 FROM pull_requests WHERE pull_request_id = $1)`,
		prID,
	).Scan(&exists)
	return exists, err
}

func (r *Repository) GetPR(prID string) (*models.PullRequest, error) {
	var pr models.PullRequest
	err := r.db.QueryRow(`
		SELECT pull_request_id, pull_request_name, author_id, status, created_at, merged_at
		FROM pull_requests WHERE pull_request_id = $1
	`, prID).Scan(
		&pr.PullRequestID,
		&pr.PullRequestName,
		&pr.AuthorID,
		&pr.Status,
		&pr.CreatedAt,
		&pr.MergedAt,
	)
	if err != nil {
		return nil, err
	}

	reviewers, err := r.GetPRReviewers(prID)
	if err != nil {
		return nil, err
	}
	pr.AssignedReviewers = reviewers

	return &pr, nil
}

func (r *Repository) MergePR(prID string) error {
	_, err := r.db.Exec(`
		UPDATE pull_requests 
		SET status = 'MERGED', merged_at = NOW()
		WHERE pull_request_id = $1 AND status = 'OPEN'
	`, prID)
	return err
}

func (r *Repository) AddReviewer(prID, userID string) error {
	_, err := r.db.Exec(
		`INSERT INTO pr_reviewers (pull_request_id, user_id) VALUES ($1, $2)`,
		prID, userID,
	)
	return err
}

func (r *Repository) RemoveReviewer(prID, userID string) error {
	result, err := r.db.Exec(
		`DELETE FROM pr_reviewers WHERE pull_request_id = $1 AND user_id = $2`,
		prID, userID,
	)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("reviewer not found")
	}
	return nil
}

func (r *Repository) GetPRReviewers(prID string) ([]string, error) {
	rows, err := r.db.Query(
		`SELECT user_id FROM pr_reviewers WHERE pull_request_id = $1 ORDER BY assigned_at`,
		prID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reviewers []string
	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			return nil, err
		}
		reviewers = append(reviewers, userID)
	}
	return reviewers, nil
}

func (r *Repository) IsUserReviewer(prID, userID string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(
		`SELECT EXISTS(SELECT 1 FROM pr_reviewers WHERE pull_request_id = $1 AND user_id = $2)`,
		prID, userID,
	).Scan(&exists)
	return exists, err
}

func (r *Repository) GetUserPRs(userID string) ([]models.PullRequestShort, error) {
	rows, err := r.db.Query(`
		SELECT pr.pull_request_id, pr.pull_request_name, pr.author_id, pr.status
		FROM pull_requests pr
		INNER JOIN pr_reviewers rev ON pr.pull_request_id = rev.pull_request_id
		WHERE rev.user_id = $1
		ORDER BY pr.created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prs []models.PullRequestShort
	for rows.Next() {
		var pr models.PullRequestShort
		if err := rows.Scan(&pr.PullRequestID, &pr.PullRequestName, &pr.AuthorID, &pr.Status); err != nil {
			return nil, err
		}
		prs = append(prs, pr)
	}
	return prs, nil
}
