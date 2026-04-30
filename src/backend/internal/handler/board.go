package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"teamboard/internal/dto/request"
	"teamboard/internal/service"
)

type BoardHandler struct {
	boardManager service.BoardManager
}

func NewBoardHandler(boardManager service.BoardManager) *BoardHandler {
	return &BoardHandler{boardManager: boardManager}
}

func (h *BoardHandler) Create(c echo.Context) error {
	authCtx := getAuthContext(c)
	var req request.CreateBoardRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, newErrorResponse("BAD_REQUEST", "Invalid request body"))
	}

	result, err := h.boardManager.Create(c.Request().Context(), authCtx.TeamID, &req)
	if err != nil {
		return handleError(c, err)
	}

	return c.JSON(http.StatusCreated, result)
}

func (h *BoardHandler) List(c echo.Context) error {
	authCtx := getAuthContext(c)
	result, err := h.boardManager.ListByTeam(c.Request().Context(), authCtx.TeamID)
	if err != nil {
		return handleError(c, err)
	}
	return c.JSON(http.StatusOK, result)
}

func (h *BoardHandler) Get(c echo.Context) error {
	authCtx := getAuthContext(c)
	boardID := c.Param("id")

	result, err := h.boardManager.GetByID(c.Request().Context(), authCtx.TeamID, boardID)
	if err != nil {
		return handleError(c, err)
	}
	return c.JSON(http.StatusOK, result)
}

func (h *BoardHandler) Update(c echo.Context) error {
	authCtx := getAuthContext(c)
	boardID := c.Param("id")

	var req request.UpdateBoardRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, newErrorResponse("BAD_REQUEST", "Invalid request body"))
	}

	result, err := h.boardManager.Update(c.Request().Context(), authCtx.TeamID, boardID, &req)
	if err != nil {
		return handleError(c, err)
	}
	return c.JSON(http.StatusOK, result)
}

func (h *BoardHandler) Delete(c echo.Context) error {
	authCtx := getAuthContext(c)
	if authCtx.Role != "admin" {
		return c.JSON(http.StatusForbidden, newErrorResponse("FORBIDDEN", "Admin access required"))
	}

	boardID := c.Param("id")
	if err := h.boardManager.Delete(c.Request().Context(), authCtx.TeamID, boardID); err != nil {
		return handleError(c, err)
	}
	return c.NoContent(http.StatusNoContent)
}

// Column operations

func (h *BoardHandler) AddColumn(c echo.Context) error {
	authCtx := getAuthContext(c)
	boardID := c.Param("boardId")

	var req request.CreateColumnRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, newErrorResponse("BAD_REQUEST", "Invalid request body"))
	}

	result, err := h.boardManager.AddColumn(c.Request().Context(), authCtx.TeamID, boardID, &req)
	if err != nil {
		return handleError(c, err)
	}
	return c.JSON(http.StatusCreated, result)
}

func (h *BoardHandler) RenameColumn(c echo.Context) error {
	authCtx := getAuthContext(c)
	columnID := c.Param("id")

	var req request.UpdateColumnRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, newErrorResponse("BAD_REQUEST", "Invalid request body"))
	}

	result, err := h.boardManager.RenameColumn(c.Request().Context(), authCtx.TeamID, columnID, &req)
	if err != nil {
		return handleError(c, err)
	}
	return c.JSON(http.StatusOK, result)
}

func (h *BoardHandler) ReorderColumn(c echo.Context) error {
	authCtx := getAuthContext(c)
	columnID := c.Param("id")

	var req request.ReorderColumnRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, newErrorResponse("BAD_REQUEST", "Invalid request body"))
	}

	result, err := h.boardManager.ReorderColumn(c.Request().Context(), authCtx.TeamID, columnID, &req)
	if err != nil {
		return handleError(c, err)
	}
	return c.JSON(http.StatusOK, result)
}

func (h *BoardHandler) DeleteColumn(c echo.Context) error {
	authCtx := getAuthContext(c)
	columnID := c.Param("id")

	if err := h.boardManager.DeleteColumn(c.Request().Context(), authCtx.TeamID, columnID); err != nil {
		return handleError(c, err)
	}
	return c.NoContent(http.StatusNoContent)
}
