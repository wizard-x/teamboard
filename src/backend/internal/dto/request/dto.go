package request

// --- Team Registration ---

type RegisterTeamRequest struct {
	Name       string `json:"name"`
	AdminName  string `json:"admin_name"`
	AdminEmail string `json:"admin_email"`
}

// --- Board ---

type CreateBoardRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

type UpdateBoardRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

// --- Column ---

type CreateColumnRequest struct {
	Name     string `json:"name"`
	Position *int   `json:"position"`
	Status   string `json:"status"`
}

type UpdateColumnRequest struct {
	Name string `json:"name"`
}

type ReorderColumnRequest struct {
	Position int `json:"position"`
}

// --- Task ---

type CreateTaskRequest struct {
	ColumnID    string     `json:"column_id"`
	Title       string     `json:"title"`
	Description *string    `json:"description"`
	AssigneeID  *string    `json:"assignee_id"`
	DueDate     *string    `json:"due_date"`
	Priority    string     `json:"priority"`
}

type UpdateTaskRequest struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	AssigneeID  *string `json:"assignee_id"`
	DueDate     *string `json:"due_date"`
	Priority    *string `json:"priority"`
}

type MoveTaskRequest struct {
	ColumnID string `json:"column_id"`
	Position *int   `json:"position"`
}

// --- Comment ---

type CreateCommentRequest struct {
	Body string `json:"body"`
}

// --- Member ---

type CreateMemberRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

type UpdateMemberRequest struct {
	Name string `json:"name"`
	Role string `json:"role"`
}

type UpdateMeRequest struct {
	Name string `json:"name"`
}
