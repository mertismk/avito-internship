package service

import (
	"database/sql"
	"math/rand"
	"time"

	"github.com/mertismk/avito-internship/internal/models"
	"github.com/mertismk/avito-internship/internal/repository"
)

type Service struct {
	repo *repository.Repository
	rand *rand.Rand
}

func New(repo *repository.Repository) *Service {
	return &Service{
		repo: repo,
		rand: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (s *Service) CreateTeam(req models.CreateTeamRequest) (*models.Team, error) {
	exists, err := s.repo.TeamExists(req.TeamName)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, &AppError{
			Code:    models.ErrTeamExists,
			Message: "team_name already exists",
		}
	}

	if err := s.repo.CreateTeam(req.TeamName); err != nil {
		return nil, err
	}

	for _, member := range req.Members {
		if err := s.repo.UpsertUser(member, req.TeamName); err != nil {
			return nil, err
		}
	}

	return &models.Team{
		TeamName: req.TeamName,
		Members:  req.Members,
	}, nil
}

func (s *Service) GetTeam(teamName string) (*models.Team, error) {
	exists, err := s.repo.TeamExists(teamName)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, &AppError{
			Code:    models.ErrNotFound,
			Message: "team not found",
		}
	}

	members, err := s.repo.GetTeamMembers(teamName)
	if err != nil {
		return nil, err
	}

	return &models.Team{
		TeamName: teamName,
		Members:  members,
	}, nil
}

func (s *Service) SetUserIsActive(req models.SetIsActiveRequest) (*models.User, error) {
	if err := s.repo.SetUserIsActive(req.UserID, req.IsActive); err != nil {
		if err == sql.ErrNoRows {
			return nil, &AppError{
				Code:    models.ErrNotFound,
				Message: "user not found",
			}
		}
		return nil, err
	}

	return s.repo.GetUser(req.UserID)
}

func (s *Service) GetUserReviews(userID string) (*models.UserReviewsResponse, error) {
	_, err := s.repo.GetUser(userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &AppError{
				Code:    models.ErrNotFound,
				Message: "user not found",
			}
		}
		return nil, err
	}

	prs, err := s.repo.GetUserPRs(userID)
	if err != nil {
		return nil, err
	}

	return &models.UserReviewsResponse{
		UserID:       userID,
		PullRequests: prs,
	}, nil
}

func (s *Service) CreatePR(req models.CreatePRRequest) (*models.PullRequest, error) {
	exists, err := s.repo.PRExists(req.PullRequestID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, &AppError{
			Code:    models.ErrPRExists,
			Message: "PR id already exists",
		}
	}

	author, err := s.repo.GetUser(req.AuthorID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &AppError{
				Code:    models.ErrNotFound,
				Message: "author not found",
			}
		}
		return nil, err
	}

	if err := s.repo.CreatePR(req); err != nil {
		return nil, err
	}

	candidates, err := s.repo.GetActiveTeamMembers(author.TeamName, req.AuthorID)
	if err != nil {
		return nil, err
	}

	reviewers := s.selectReviewers(candidates, 2)
	for _, reviewerID := range reviewers {
		if err := s.repo.AddReviewer(req.PullRequestID, reviewerID); err != nil {
			return nil, err
		}
	}

	return s.repo.GetPR(req.PullRequestID)
}

func (s *Service) MergePR(req models.MergePRRequest) (*models.PullRequest, error) {
	pr, err := s.repo.GetPR(req.PullRequestID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &AppError{
				Code:    models.ErrNotFound,
				Message: "PR not found",
			}
		}
		return nil, err
	}

	if pr.Status == "MERGED" {
		return pr, nil
	}

	if err := s.repo.MergePR(req.PullRequestID); err != nil {
		return nil, err
	}

	return s.repo.GetPR(req.PullRequestID)
}

func (s *Service) ReassignReviewer(req models.ReassignReviewerRequest) (*models.PullRequest, string, error) {
	pr, err := s.repo.GetPR(req.PullRequestID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, "", &AppError{
				Code:    models.ErrNotFound,
				Message: "PR not found",
			}
		}
		return nil, "", err
	}

	if pr.Status == "MERGED" {
		return nil, "", &AppError{
			Code:    models.ErrPRMerged,
			Message: "cannot reassign on merged PR",
		}
	}

	isReviewer, err := s.repo.IsUserReviewer(req.PullRequestID, req.OldUserID)
	if err != nil {
		return nil, "", err
	}
	if !isReviewer {
		return nil, "", &AppError{
			Code:    models.ErrNotAssigned,
			Message: "reviewer is not assigned to this PR",
		}
	}

	oldReviewer, err := s.repo.GetUser(req.OldUserID)
	if err != nil {
		return nil, "", err
	}

	candidates, err := s.repo.GetActiveTeamMembers(oldReviewer.TeamName, pr.AuthorID)
	if err != nil {
		return nil, "", err
	}

	var available []models.User
	for _, candidate := range candidates {
		isAssigned := false
		for _, reviewerID := range pr.AssignedReviewers {
			if candidate.UserID == reviewerID {
				isAssigned = true
				break
			}
		}
		if !isAssigned {
			available = append(available, candidate)
		}
	}

	if len(available) == 0 {
		return nil, "", &AppError{
			Code:    models.ErrNoCandidate,
			Message: "no active replacement candidate in team",
		}
	}

	newReviewerID := s.selectReviewers(available, 1)[0]

	if err := s.repo.RemoveReviewer(req.PullRequestID, req.OldUserID); err != nil {
		return nil, "", err
	}

	if err := s.repo.AddReviewer(req.PullRequestID, newReviewerID); err != nil {
		return nil, "", err
	}

	pr, err = s.repo.GetPR(req.PullRequestID)
	if err != nil {
		return nil, "", err
	}

	return pr, newReviewerID, nil
}

func (s *Service) selectReviewers(candidates []models.User, count int) []string {
	if len(candidates) <= count {
		result := make([]string, len(candidates))
		for i, c := range candidates {
			result[i] = c.UserID
		}
		return result
	}

	shuffled := make([]models.User, len(candidates))
	copy(shuffled, candidates)

	for i := len(shuffled) - 1; i > 0; i-- {
		j := s.rand.Intn(i + 1)
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	}

	result := make([]string, count)
	for i := 0; i < count; i++ {
		result[i] = shuffled[i].UserID
	}
	return result
}

type AppError struct {
	Code    models.ErrorCode
	Message string
}

func (e *AppError) Error() string {
	return e.Message
}
