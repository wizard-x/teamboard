import axios, { type AxiosInstance, type AxiosError } from 'axios'
import type {
  RegisterRequest,
  RegisterResponse,
  Board,
  BoardDetail,
  CreateBoardRequest,
  UpdateBoardRequest,
  Column,
  CreateColumnRequest,
  RenameColumnRequest,
  ReorderColumnRequest,
  Task,
  TaskDetail,
  CreateTaskRequest,
  UpdateTaskRequest,
  MoveTaskRequest,
  Comment,
  CreateCommentRequest,
  CommentsResponse,
  Member,
  MemberWithKey,
  CreateMemberRequest,
  UpdateMemberRequest,
  HealthResponse,
  ApiError,
} from '@/types'

const API_BASE = '/api/v1'

function createClient(): AxiosInstance {
  const instance = axios.create({ baseURL: API_BASE })

  instance.interceptors.request.use((config) => {
    const apiKey = localStorage.getItem('api_key')
    if (apiKey) {
      config.headers['X-API-Key'] = apiKey
    }
    return config
  })

  instance.interceptors.response.use(
    (res) => res,
    (error: AxiosError<ApiError>) => {
      if (error.response?.status === 401) {
        localStorage.removeItem('api_key')
        localStorage.removeItem('member')
        window.location.href = '/login'
      }
      return Promise.reject(error)
    }
  )

  return instance
}

const client = createClient()

// ── Helper to extract data from responses ──
function unwrap<T>(res: { data: { data: T } }): T {
  return res.data.data
}

// ────────────────────────────────────────────
// Health
// ────────────────────────────────────────────

export const healthApi = {
  check(): Promise<HealthResponse> {
    return axios.get('/health').then((r) => r.data)
  },
}

// ────────────────────────────────────────────
// Auth / Registration
// ────────────────────────────────────────────

export const authApi = {
  register(req: RegisterRequest): Promise<RegisterResponse> {
    return client.post<{ data: RegisterResponse }>('/teams/register', req).then((r) => r.data.data)
  },
}

// ────────────────────────────────────────────
// Boards
// ────────────────────────────────────────────

export const boardsApi = {
  list(): Promise<Board[]> {
    return client.get<{ data: Board[] }>('/boards').then((r) => r.data.data)
  },

  get(id: string): Promise<BoardDetail> {
    return client.get<{ data: BoardDetail }>(`/boards/${id}`).then(unwrap)
  },

  create(req: CreateBoardRequest): Promise<BoardDetail> {
    return client.post<{ data: BoardDetail }>('/boards', req).then(unwrap)
  },

  update(id: string, req: UpdateBoardRequest): Promise<Board> {
    return client.put<{ data: Board }>(`/boards/${id}`, req).then(unwrap)
  },

  delete(id: string): Promise<void> {
    return client.delete(`/boards/${id}`)
  },
}

// ────────────────────────────────────────────
// Columns
// ────────────────────────────────────────────

export const columnsApi = {
  create(boardId: string, req: CreateColumnRequest): Promise<Column> {
    return client.post<{ data: Column }>(`/boards/${boardId}/columns`, req).then(unwrap)
  },

  rename(id: string, req: RenameColumnRequest): Promise<Column> {
    return client.put<{ data: Column }>(`/columns/${id}`, req).then(unwrap)
  },

  reorder(id: string, req: ReorderColumnRequest): Promise<Column[]> {
    return client.put<{ data: Column[] }>(`/columns/${id}/position`, req).then((r) => r.data.data)
  },

  delete(id: string): Promise<void> {
    return client.delete(`/columns/${id}`)
  },
}

// ────────────────────────────────────────────
// Tasks
// ────────────────────────────────────────────

export const tasksApi = {
  create(columnId: string, req: CreateTaskRequest): Promise<Task> {
    return client.post<{ data: Task }>(`/columns/${columnId}/tasks`, req).then(unwrap)
  },

  get(id: string): Promise<TaskDetail> {
    return client.get<{ data: TaskDetail }>(`/tasks/${id}`).then(unwrap)
  },

  update(id: string, req: UpdateTaskRequest): Promise<Task> {
    return client.put<{ data: Task }>(`/tasks/${id}`, req).then(unwrap)
  },

  move(id: string, req: MoveTaskRequest): Promise<Task> {
    return client.put<{ data: Task }>(`/tasks/${id}/move`, req).then(unwrap)
  },

  delete(id: string): Promise<void> {
    return client.delete(`/tasks/${id}`)
  },
}

// ────────────────────────────────────────────
// Comments
// ────────────────────────────────────────────

export const commentsApi = {
  list(taskId: string, page = 1, perPage = 20): Promise<CommentsResponse> {
    return client
      .get<CommentsResponse>(`/tasks/${taskId}/comments`, {
        params: { page, per_page: perPage },
      })
      .then((r) => r.data)
  },

  create(taskId: string, req: CreateCommentRequest): Promise<Comment> {
    return client.post<{ data: Comment }>(`/tasks/${taskId}/comments`, req).then(unwrap)
  },

  delete(id: string): Promise<void> {
    return client.delete(`/comments/${id}`)
  },
}

// ────────────────────────────────────────────
// Members
// ────────────────────────────────────────────

export const membersApi = {
  me(): Promise<Member> {
    return client.get<{ data: Member }>('/members/me').then(unwrap)
  },

  updateMe(req: { name: string }): Promise<Member> {
    return client.put<{ data: Member }>('/members/me', req).then(unwrap)
  },

  regenerateMyKey(): Promise<{ api_key: string }> {
    return client.post<{ data: { api_key: string } }>('/members/me/regenerate-key').then(unwrap)
  },

  list(): Promise<Member[]> {
    return client.get<{ data: Member[] }>('/members').then((r) => r.data.data)
  },

  create(req: CreateMemberRequest): Promise<MemberWithKey> {
    return client.post<{ data: MemberWithKey }>('/members', req).then(unwrap)
  },

  update(id: string, req: UpdateMemberRequest): Promise<Member> {
    return client.put<{ data: Member }>(`/members/${id}`, req).then(unwrap)
  },

  remove(id: string): Promise<void> {
    return client.delete(`/members/${id}`)
  },

  regenerateKey(id: string): Promise<{ api_key: string }> {
    return client.post<{ data: { api_key: string } }>(`/members/${id}/regenerate-key`).then(unwrap)
  },
}
