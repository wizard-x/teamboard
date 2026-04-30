package handler

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"teamboard/internal/dto/request"
	"teamboard/internal/service"
)

type TaskHandler struct {
	taskManager    service.TaskManager
	commentManager service.CommentManager
}

func NewTaskHandler(taskManager service.TaskManager, commentManager service.CommentManager) *TaskHandler {
	return &TaskHandler{taskManager: taskManager, commentManager: commentManager}
}

func (h *TaskHandler) Create(c echo.Context) error {
	authCtx := getAuthContext(c)
	columnID := c.Param("columnId")

	var req request.CreateTaskRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, newErrorResponse("BAD_REQUEST", "Invalid request body"))
	}
	req.ColumnID = columnID

	result, err := h.taskManager.Create(c.Request().Context(), authCtx.TeamID, authCtx.MemberID, &req)
	if err != nil {
		return handleError(c, err)
	}
	return c.JSON(http.StatusCreated, result)
}

func (h *TaskHandler) Get(c echo.Context) error {
	authCtx := getAuthContext(c)
	taskID := c.Param("id")

	result, err := h.taskManager.GetByID(c.Request().Context(), authCtx.TeamID, taskID)
	if err != nil {
		return handleError(c, err)
	}
	return c.JSON(http.StatusOK, result)
}

func (h *TaskHandler) Update(c echo.Context) error {
	authCtx := getAuthContext(c)
	taskID := c.Param("id")

	var req request.UpdateTaskRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, newErrorResponse("BAD_REQUEST", "Invalid request body"))
	}

	result, err := h.taskManager.Update(c.Request().Context(), authCtx.TeamID, taskID, &req)
	if err != nil {
		return handleError(c, err)
	}
	return c.JSON(http.StatusOK, result)
}

func (h *TaskHandler) Move(c echo.Context) error {
	authCtx := getAuthContext(c)
	taskID := c.Param("id")

	var req request.MoveTaskRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, newErrorResponse("BAD_REQUEST", "Invalid request body"))
	}

	result, err := h.taskManager.Move(c.Request().Context(), authCtx.TeamID, taskID, &req)
	if err != nil {
		return handleError(c, err)
	}
	return c.JSON(http.StatusOK, result)
}

func (h *TaskHandler) Delete(c echo.Context) error {
	authCtx := getAuthContext(c)
	taskID := c.Param("id")

	if err := h.taskManager.Delete(c.Request().Context(), authCtx.TeamID, taskID); err != nil {
		return handleError(c, err)
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *TaskHandler) CreateComment(c echo.Context) error {
	authCtx := getAuthContext(c)
	taskID := c.Param("taskId")

	var req request.CreateCommentRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, newErrorResponse("BAD_REQUEST", "Invalid request body"))
	}

	result, err := h.commentManager.Create(c.Request().Context(), authCtx.TeamID, taskID, authCtx.MemberID, &req)
	if err != nil {
		return handleError(c, err)
	}
	return c.JSON(http.StatusCreated, result)
}

func (h *TaskHandler) ListComments(c echo.Context) error {
	authCtx := getAuthContext(c)
	taskID := c.Param("taskId")

	page := getPage(c)
	perPage := getPerPage(c)

	result, err := h.commentManager.ListByTask(c.Request().Context(), authCtx.TeamID, taskID, page, perPage)
	if err != nil {
		return handleError(c, err)
	}
	return c.JSON(http.StatusOK, result)
}

func (h *TaskHandler) DeleteComment(c echo.Context) error {
	authCtx := getAuthContext(c)
	commentID := c.Param("id")

	isAdmin := authCtx.Role == "admin"
	if err := h.commentManager.Delete(c.Request().Context(), authCtx.TeamID, commentID, authCtx.MemberID, isAdmin); err != nil {
		return handleError(c, err)
	}
	return c.NoContent(http.StatusNoContent)
}

// Helpers

func getPage(c echo.Context) int {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}
	return page
}

func getPerPage(c echo.Context) int {
	perPage, _ := strconv.Atoi(c.QueryParam("per_page"))
	if perPage < 1 {
		perPage = 20
	}
	if perPage > 100 {
		perPage = 100
	}
	return perPage
}
