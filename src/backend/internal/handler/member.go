package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"teamboard/internal/dto/request"
	"teamboard/internal/service"
)

type MemberHandler struct {
	memberManager service.MemberManager
}

func NewMemberHandler(memberManager service.MemberManager) *MemberHandler {
	return &MemberHandler{memberManager: memberManager}
}

func (h *MemberHandler) Me(c echo.Context) error {
	authCtx := getAuthContext(c)

	result, err := h.memberManager.GetCurrentMember(c.Request().Context(), authCtx.MemberID)
	if err != nil {
		return handleError(c, err)
	}
	return c.JSON(http.StatusOK, result)
}

func (h *MemberHandler) RegenerateMyKey(c echo.Context) error {
	authCtx := getAuthContext(c)

	result, err := h.memberManager.RegenerateAPIKey(c.Request().Context(), authCtx.TeamID, authCtx.MemberID)
	if err != nil {
		return handleError(c, err)
	}

	return c.JSON(http.StatusOK, result)
}

func (h *MemberHandler) UpdateMe(c echo.Context) error {
	authCtx := getAuthContext(c)

	var req request.UpdateMeRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, newErrorResponse("BAD_REQUEST", "Invalid request body"))
	}

	result, err := h.memberManager.UpdateMe(c.Request().Context(), authCtx.MemberID, &req)
	if err != nil {
		return handleError(c, err)
	}

	return c.JSON(http.StatusOK, result)
}

func (h *MemberHandler) List(c echo.Context) error {
	authCtx := getAuthContext(c)
	result, err := h.memberManager.ListByTeam(c.Request().Context(), authCtx.TeamID)
	if err != nil {
		return handleError(c, err)
	}
	return c.JSON(http.StatusOK, result)
}

func (h *MemberHandler) Create(c echo.Context) error {
	authCtx := getAuthContext(c)
	if authCtx.Role != "admin" {
		return c.JSON(http.StatusForbidden, newErrorResponse("FORBIDDEN", "Admin access required"))
	}

	var req request.CreateMemberRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, newErrorResponse("BAD_REQUEST", "Invalid request body"))
	}

	result, err := h.memberManager.Create(c.Request().Context(), authCtx.TeamID, &req)
	if err != nil {
		return handleError(c, err)
	}

	return c.JSON(http.StatusCreated, result)
}

func (h *MemberHandler) Update(c echo.Context) error {
	authCtx := getAuthContext(c)
	if authCtx.Role != "admin" {
		return c.JSON(http.StatusForbidden, newErrorResponse("FORBIDDEN", "Admin access required"))
	}

	memberID := c.Param("id")
	var req request.UpdateMemberRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, newErrorResponse("BAD_REQUEST", "Invalid request body"))
	}

	result, err := h.memberManager.Update(c.Request().Context(), authCtx.TeamID, memberID, &req)
	if err != nil {
		return handleError(c, err)
	}

	return c.JSON(http.StatusOK, result)
}

func (h *MemberHandler) Delete(c echo.Context) error {
	authCtx := getAuthContext(c)
	if authCtx.Role != "admin" {
		return c.JSON(http.StatusForbidden, newErrorResponse("FORBIDDEN", "Admin access required"))
	}

	memberID := c.Param("id")
	if err := h.memberManager.Delete(c.Request().Context(), authCtx.TeamID, memberID); err != nil {
		return handleError(c, err)
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *MemberHandler) RegenerateKey(c echo.Context) error {
	authCtx := getAuthContext(c)
	if authCtx.Role != "admin" {
		return c.JSON(http.StatusForbidden, newErrorResponse("FORBIDDEN", "Admin access required"))
	}

	memberID := c.Param("id")
	result, err := h.memberManager.RegenerateAPIKey(c.Request().Context(), authCtx.TeamID, memberID)
	if err != nil {
		return handleError(c, err)
	}

	return c.JSON(http.StatusOK, result)
}
