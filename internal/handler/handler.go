package handler

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mertismk/avito-internship/internal/models"
	"github.com/mertismk/avito-internship/internal/service"
)

type Handler struct {
	service *service.Service
}

func New(service *service.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) SetupRoutes(r *gin.Engine) {
	r.POST("/team/add", h.CreateTeam)
	r.GET("/team/get", h.GetTeam)

	r.POST("/users/setIsActive", h.SetUserIsActive)
	r.GET("/users/getReview", h.GetUserReviews)

	r.POST("/pullRequest/create", h.CreatePR)
	r.POST("/pullRequest/merge", h.MergePR)
	r.POST("/pullRequest/reassign", h.ReassignReviewer)

	r.GET("/health", h.HealthCheck)
}

func (h *Handler) CreateTeam(c *gin.Context) {
	var req models.CreateTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: models.ErrorDetail{
				Code:    models.ErrNotFound,
				Message: "invalid request body",
			},
		})
		return
	}

	team, err := h.service.CreateTeam(req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, models.TeamResponse{Team: *team})
}

func (h *Handler) GetTeam(c *gin.Context) {
	teamName := c.Query("team_name")
	if teamName == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: models.ErrorDetail{
				Code:    models.ErrNotFound,
				Message: "team_name query parameter is required",
			},
		})
		return
	}

	team, err := h.service.GetTeam(teamName)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, *team)
}

func (h *Handler) SetUserIsActive(c *gin.Context) {
	var req models.SetIsActiveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: models.ErrorDetail{
				Code:    models.ErrNotFound,
				Message: "invalid request body",
			},
		})
		return
	}

	user, err := h.service.SetUserIsActive(req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, models.UserResponse{User: *user})
}

func (h *Handler) GetUserReviews(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: models.ErrorDetail{
				Code:    models.ErrNotFound,
				Message: "user_id query parameter is required",
			},
		})
		return
	}

	resp, err := h.service.GetUserReviews(userID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *Handler) CreatePR(c *gin.Context) {
	var req models.CreatePRRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: models.ErrorDetail{
				Code:    models.ErrNotFound,
				Message: "invalid request body",
			},
		})
		return
	}

	pr, err := h.service.CreatePR(req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, models.PRResponse{PR: *pr})
}

func (h *Handler) MergePR(c *gin.Context) {
	var req models.MergePRRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: models.ErrorDetail{
				Code:    models.ErrNotFound,
				Message: "invalid request body",
			},
		})
		return
	}

	pr, err := h.service.MergePR(req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, models.PRResponse{PR: *pr})
}

func (h *Handler) ReassignReviewer(c *gin.Context) {
	var req models.ReassignReviewerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: models.ErrorDetail{
				Code:    models.ErrNotFound,
				Message: "invalid request body",
			},
		})
		return
	}

	pr, replacedBy, err := h.service.ReassignReviewer(req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, models.ReassignResponse{
		PR:         *pr,
		ReplacedBy: replacedBy,
	})
}

func (h *Handler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"service": "reviewer-service",
	})
}

func handleError(c *gin.Context, err error) {
	var appErr *service.AppError
	if errors.As(err, &appErr) {
		c.JSON(statusCode(appErr.Code), models.ErrorResponse{
			Error: models.ErrorDetail{
				Code:    appErr.Code,
				Message: appErr.Message,
			},
		})
		return
	}

	if errors.Is(err, sql.ErrNoRows) {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error: models.ErrorDetail{
				Code:    models.ErrNotFound,
				Message: "resource not found",
			},
		})
		return
	}

	c.JSON(http.StatusInternalServerError, models.ErrorResponse{
		Error: models.ErrorDetail{
			Code:    models.ErrNotFound,
			Message: "internal server error",
		},
	})
}

func statusCode(code models.ErrorCode) int {
	switch code {
	case models.ErrTeamExists, models.ErrPRExists:
		return http.StatusBadRequest
	case models.ErrPRMerged, models.ErrNotAssigned, models.ErrNoCandidate:
		return http.StatusConflict
	case models.ErrNotFound:
		return http.StatusNotFound
	default:
		return http.StatusInternalServerError
	}
}
