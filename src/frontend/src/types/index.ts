// ── Common ──

export interface ApiError {
  error: {
    code: string
    message: string
    details?: Array<{ field: string; message: string }>
  }
}

export interface PaginationMeta {
  total: number
  page: number
  per_page: number
  total_pages: number
}

export interface MemberSummary {
  id: string
  name: string
}

// ── Auth / Registration ──

export interface RegisterRequest {
  name: string
  admin_name: string
  admin_email: string
}

export interface RegisterResponse {
  team: {
    id: string
    name: string
    created_at: string
  }
  member: {
    id: string
    name: string
    email: string
    role: string
  }
  api_key: string
}

// ── Board ──

export interface Board {
  id: string
  name: string
  description?: string
  created_at: string
  updated_at: string
}

export interface Column {
  id: string
  name: string
  position: number
  status: ColumnStatus
  board_id?: string
  tasks?: Task[]
}

export type ColumnStatus = 'todo' | 'in_progress' | 'review' | 'done'

export interface BoardDetail {
  id: string
  name: string
  description?: string
  columns: Column[]
  created_at: string
  updated_at: string
}

export interface CreateBoardRequest {
  name: string
  description?: string
}

export interface UpdateBoardRequest {
  name?: string
  description?: string
}

// ── Column ──

export interface CreateColumnRequest {
  name: string
  position?: number
  status?: ColumnStatus
}

export interface RenameColumnRequest {
  name: string
}

export interface ReorderColumnRequest {
  position: number
}

// ── Task ──

export type TaskPriority = 'low' | 'medium' | 'high' | 'critical'

export interface Task {
  id: string
  title: string
  description?: string
  status: ColumnStatus
  priority: TaskPriority
  position: number
  column_id: string
  board_id: string
  assignee: MemberSummary | null
  due_date: string | null
  comment_count: number
  created_at: string
  updated_at: string
}

export interface TaskDetail extends Task {
  comments: Comment[]
}

export interface CreateTaskRequest {
  title: string
  description?: string
  assignee_id?: string | null
  due_date?: string | null
  priority?: TaskPriority
}

export interface UpdateTaskRequest {
  title?: string
  description?: string
  assignee_id?: string | null
  due_date?: string | null
  priority?: TaskPriority
}

export interface MoveTaskRequest {
  column_id: string
  position?: number
}

// ── Comment ──

export interface Comment {
  id: string
  body: string
  author: MemberSummary
  created_at: string
  updated_at: string
}

export interface CreateCommentRequest {
  body: string
}

export interface CommentsResponse {
  data: Comment[]
  meta: PaginationMeta
}

// ── Member ──

export type MemberRole = 'admin' | 'member'

export interface Member {
  id: string
  name: string
  email: string
  role: MemberRole
  created_at: string
}

export interface MemberWithKey extends Member {
  api_key: string
}

export interface CreateMemberRequest {
  name: string
  email: string
  role: MemberRole
}

export interface UpdateMemberRequest {
  name?: string
  role?: MemberRole
}

// ── Health ──

export interface HealthResponse {
  status: string
  postgres: string
  redis: string
  version: string
}
