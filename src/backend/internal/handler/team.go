package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"teamboard/internal/dto/request"
	"teamboard/internal/service"
)

type TeamHandler struct {
	teamManager service.TeamManager
}

func NewTeamHandler(teamManager service.TeamManager) *TeamHandler {
	return &TeamHandler{teamManager: teamManager}
}

func (h *TeamHandler) Register(c echo.Context) error {
	var req request.RegisterTeamRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, newErrorResponse("BAD_REQUEST", "Invalid request body"))
	}

	result, err := h.teamManager.Register(c.Request().Context(), &req)
	if err != nil {
		return handleError(c, err)
	}

	return c.JSON(http.StatusCreated, result)
}

// AuthContext retrieves auth context from echo context.
type AuthContext struct {
	MemberID string
	TeamID   string
	Role     string
}

func getAuthContext(c echo.Context) *AuthContext {
	ac := c.Get("auth_context")
	if ac == nil {
		return nil
	}
	if authCtx, ok := ac.(*service.AuthContext); ok {
		return &AuthContext{
			MemberID: authCtx.MemberID,
			TeamID:   authCtx.TeamID,
			Role:     authCtx.Role,
		}
	}
	return nil
}
