package models

import "time"

type User struct {
	UserID   string `json:"user_id" db:"user_id"`
	Username string `json:"username" db:"username"`
	TeamName string `json:"team_name" db:"team_name"`
	IsActive bool   `json:"is_active" db:"is_active"`
}

type TeamMember struct {
	UserID   string `json:"user_id" db:"user_id"`
	Username string `json:"username" db:"username"`
	IsActive bool   `json:"is_active" db:"is_active"`
}

type Team struct {
	TeamName string       `json:"team_name" db:"team_name"`
	Members  []TeamMember `json:"members"`
}

type PullRequest struct {
	PullRequestID     string     `json:"pull_request_id" db:"pull_request_id"`
	PullRequestName   string     `json:"pull_request_name" db:"pull_request_name"`
	AuthorID          string     `json:"author_id" db:"author_id"`
	Status            string     `json:"status" db:"status"`
	AssignedReviewers []string   `json:"assigned_reviewers"`
	CreatedAt         *time.Time `json:"createdAt,omitempty" db:"created_at"`
	MergedAt          *time.Time `json:"mergedAt,omitempty" db:"merged_at"`
}

type PullRequestShort struct {
	PullRequestID   string `json:"pull_request_id" db:"pull_request_id"`
	PullRequestName string `json:"pull_request_name" db:"pull_request_name"`
	AuthorID        string `json:"author_id" db:"author_id"`
	Status          string `json:"status" db:"status"`
}

type CreateTeamRequest struct {
	TeamName string       `json:"team_name" binding:"required"`
	Members  []TeamMember `json:"members" binding:"required"`
}

type SetIsActiveRequest struct {
	UserID   string `json:"user_id" binding:"required"`
	IsActive bool   `json:"is_active"`
}

type CreatePRRequest struct {
	PullRequestID   string `json:"pull_request_id" binding:"required"`
	PullRequestName string `json:"pull_request_name" binding:"required"`
	AuthorID        string `json:"author_id" binding:"required"`
}

type MergePRRequest struct {
	PullRequestID string `json:"pull_request_id" binding:"required"`
}

type ReassignReviewerRequest struct {
	PullRequestID string `json:"pull_request_id" binding:"required"`
	OldUserID     string `json:"old_user_id" binding:"required"`
}

type ErrorCode string

const (
	ErrTeamExists  ErrorCode = "TEAM_EXISTS"
	ErrPRExists    ErrorCode = "PR_EXISTS"
	ErrPRMerged    ErrorCode = "PR_MERGED"
	ErrNotAssigned ErrorCode = "NOT_ASSIGNED"
	ErrNoCandidate ErrorCode = "NO_CANDIDATE"
	ErrNotFound    ErrorCode = "NOT_FOUND"
)

type ErrorDetail struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
}

type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

type TeamResponse struct {
	Team Team `json:"team"`
}

type UserResponse struct {
	User User `json:"user"`
}

type PRResponse struct {
	PR PullRequest `json:"pr"`
}

type ReassignResponse struct {
	PR         PullRequest `json:"pr"`
	ReplacedBy string      `json:"replaced_by"`
}

type UserReviewsResponse struct {
	UserID       string             `json:"user_id"`
	PullRequests []PullRequestShort `json:"pull_requests"`
}
