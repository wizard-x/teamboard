package response

import "time"

// --- Team ---

type RegisterTeamResponse struct {
	Data RegisterTeamData `json:"data"`
}

type RegisterTeamData struct {
	Team   TeamBrief    `json:"team"`
	Member MemberDetail `json:"member"`
	APIKey string       `json:"api_key"`
}

type TeamBrief struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type TeamResponse struct {
	Data TeamBrief `json:"data"`
}

// --- Member ---

type MemberResponse struct {
	Data MemberDetail `json:"data"`
}

type MemberWithKeyResponse struct {
	Data MemberWithKey `json:"data"`
}

type MemberDetail struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

type MemberWithKey struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	APIKey    string    `json:"api_key"`
	CreatedAt time.Time `json:"created_at"`
}

type MemberListResponse struct {
	Data []MemberDetail `json:"data"`
}

type APIKeyResponse struct {
	Data APIKeyData `json:"data"`
}

type APIKeyData struct {
	APIKey string `json:"api_key"`
}

// --- Member summary (for embedding) ---

type MemberSummary struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
