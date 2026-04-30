import { describe, it, expect, vi } from 'vitest'
import axios from 'axios'
import {
  healthApi,
  authApi,
  boardsApi,
  columnsApi,
  tasksApi,
  commentsApi,
  membersApi,
} from '@/api'

// Mock axios
vi.mock('axios', () => {
  const mockInstance = {
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
    delete: vi.fn(),
    interceptors: {
      request: { use: vi.fn() },
      response: { use: vi.fn() },
    },
  }
  const mockCreate = vi.fn(() => mockInstance)
  return {
    default: {
      create: mockCreate,
      get: vi.fn(),
    },
    __mockInstance: mockInstance,
  }
})

// Get the mocked instance
const mockAxios = axios as any

describe('API Client', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    localStorage.clear()
  })

  describe('healthApi', () => {
    it('calls GET /health', async () => {
      mockAxios.get.mockResolvedValue({
        data: { status: 'ok', postgres: 'ok', redis: 'ok', version: '1.0.0' },
      })
      const result = await healthApi.check()
      expect(mockAxios.get).toHaveBeenCalledWith('/health')
      expect(result.status).toBe('ok')
    })
  })

  describe('authApi', () => {
    it('calls POST /teams/register', async () => {
      const response = {
        data: {
          data: {
            team: { id: 't1', name: 'Team', created_at: '2026-01-01T00:00:00Z' },
            member: { id: 'm1', name: 'Admin', email: 'a@b.com', role: 'admin' },
            api_key: 'tb_k1_test',
          },
        },
      }
      const instance = mockAxios.create()
      instance.post.mockResolvedValue(response)

      const result = await authApi.register({
        name: 'Team',
        admin_name: 'Admin',
        admin_email: 'a@b.com',
      })
      expect(result.api_key).toBe('tb_k1_test')
    })
  })

  describe('boardsApi', () => {
    it('lists boards', async () => {
      const instance = mockAxios.create()
      instance.get.mockResolvedValue({
        data: {
          data: [
            { id: 'b1', name: 'Board 1', created_at: '2026-01-01T00:00:00Z', updated_at: '2026-01-01T00:00:00Z' },
          ],
        },
      })
      const result = await boardsApi.list()
      expect(result).toHaveLength(1)
      expect(result[0].name).toBe('Board 1')
    })

    it('gets board detail', async () => {
      const instance = mockAxios.create()
      instance.get.mockResolvedValue({
        data: {
          data: {
            id: 'b1',
            name: 'Board 1',
            columns: [],
            created_at: '2026-01-01T00:00:00Z',
            updated_at: '2026-01-01T00:00:00Z',
          },
        },
      })
      const result = await boardsApi.get('b1')
      expect(result.id).toBe('b1')
    })

    it('creates a board', async () => {
      const instance = mockAxios.create()
      instance.post.mockResolvedValue({
        data: {
          data: {
            id: 'b2',
            name: 'New Board',
            columns: [],
            created_at: '2026-01-01T00:00:00Z',
            updated_at: '2026-01-01T00:00:00Z',
          },
        },
      })
      const result = await boardsApi.create({ name: 'New Board' })
      expect(result.name).toBe('New Board')
    })

    it('updates a board', async () => {
      const instance = mockAxios.create()
      instance.put.mockResolvedValue({
        data: {
          data: { id: 'b1', name: 'Updated', created_at: '2026-01-01T00:00:00Z', updated_at: '2026-01-01T00:00:00Z' },
        },
      })
      const result = await boardsApi.update('b1', { name: 'Updated' })
      expect(result.name).toBe('Updated')
    })

    it('deletes a board', async () => {
      const instance = mockAxios.create()
      instance.delete.mockResolvedValue({ status: 204 })
      await boardsApi.delete('b1')
      expect(instance.delete).toHaveBeenCalledWith('/boards/b1')
    })
  })

  describe('tasksApi', () => {
    it('creates a task', async () => {
      const instance = mockAxios.create()
      instance.post.mockResolvedValue({
        data: {
          data: {
            id: 't1',
            title: 'New Task',
            status: 'todo',
            priority: 'high',
            position: 0,
            column_id: 'c1',
            board_id: 'b1',
            assignee: null,
            due_date: null,
            comment_count: 0,
            created_at: '2026-01-01T00:00:00Z',
            updated_at: '2026-01-01T00:00:00Z',
          },
        },
      })
      const result = await tasksApi.create('c1', { title: 'New Task', priority: 'high' })
      expect(result.title).toBe('New Task')
      expect(result.priority).toBe('high')
    })

    it('moves a task', async () => {
      const instance = mockAxios.create()
      instance.put.mockResolvedValue({
        data: {
          data: {
            id: 't1',
            title: 'Task',
            status: 'in_progress',
            column_id: 'c2',
            position: 1,
            priority: 'medium',
            board_id: 'b1',
            assignee: null,
            due_date: null,
            comment_count: 0,
            created_at: '2026-01-01T00:00:00Z',
            updated_at: '2026-01-01T00:00:00Z',
          },
        },
      })
      const result = await tasksApi.move('t1', { column_id: 'c2', position: 1 })
      expect(result.column_id).toBe('c2')
    })

    it('updates a task', async () => {
      const instance = mockAxios.create()
      instance.put.mockResolvedValue({
        data: {
          data: {
            id: 't1',
            title: 'Updated Task',
            status: 'todo',
            priority: 'critical',
            position: 0,
            column_id: 'c1',
            board_id: 'b1',
            assignee: null,
            due_date: null,
            comment_count: 0,
            created_at: '2026-01-01T00:00:00Z',
            updated_at: '2026-01-01T00:00:00Z',
          },
        },
      })
      const result = await tasksApi.update('t1', { priority: 'critical' })
      expect(result.priority).toBe('critical')
    })
  })

  describe('commentsApi', () => {
    it('lists comments with pagination', async () => {
      const instance = mockAxios.create()
      instance.get.mockResolvedValue({
        data: {
          data: [
            { id: 'cm1', body: 'Hello', author: { id: 'm1', name: 'User' }, created_at: '2026-01-01T00:00:00Z', updated_at: '2026-01-01T00:00:00Z' },
          ],
          meta: { total: 1, page: 1, per_page: 20, total_pages: 1 },
        },
      })
      const result = await commentsApi.list('t1')
      expect(result.data).toHaveLength(1)
      expect(result.meta.total).toBe(1)
    })

    it('creates a comment', async () => {
      const instance = mockAxios.create()
      instance.post.mockResolvedValue({
        data: {
          data: {
            id: 'cm2',
            body: 'Great work!',
            author: { id: 'm1', name: 'User' },
            created_at: '2026-01-01T00:00:00Z',
            updated_at: '2026-01-01T00:00:00Z',
          },
        },
      })
      const result = await commentsApi.create('t1', { body: 'Great work!' })
      expect(result.body).toBe('Great work!')
    })

    it('deletes a comment', async () => {
      const instance = mockAxios.create()
      instance.delete.mockResolvedValue({ status: 204 })
      await commentsApi.delete('cm1')
      expect(instance.delete).toHaveBeenCalledWith('/comments/cm1')
    })
  })

  describe('membersApi', () => {
    it('lists members', async () => {
      const instance = mockAxios.create()
      instance.get.mockResolvedValue({
        data: {
          data: [
            { id: 'm1', name: 'Admin', email: 'a@b.com', role: 'admin', created_at: '2026-01-01T00:00:00Z' },
          ],
        },
      })
      const result = await membersApi.list()
      expect(result).toHaveLength(1)
      expect(result[0].role).toBe('admin')
    })

    it('creates a member and returns api_key', async () => {
      const instance = mockAxios.create()
      instance.post.mockResolvedValue({
        data: {
          data: {
            id: 'm2',
            name: 'Member',
            email: 'm@b.com',
            role: 'member',
            created_at: '2026-01-01T00:00:00Z',
            api_key: 'tb_k1_newkey',
          },
        },
      })
      const result = await membersApi.create({ name: 'Member', email: 'm@b.com', role: 'member' })
      expect(result.api_key).toBe('tb_k1_newkey')
    })

    it('regenerates API key', async () => {
      const instance = mockAxios.create()
      instance.post.mockResolvedValue({
        data: { data: { api_key: 'tb_k3_newregen' } },
      })
      const result = await membersApi.regenerateKey('m1')
      expect(result.api_key).toBe('tb_k3_newregen')
    })
  })

  describe('columnsApi', () => {
    it('creates a column', async () => {
      const instance = mockAxios.create()
      instance.post.mockResolvedValue({
        data: {
          data: { id: 'c2', name: 'Blocked', position: 3, status: 'todo', board_id: 'b1' },
        },
      })
      const result = await columnsApi.create('b1', { name: 'Blocked', position: 3 })
      expect(result.name).toBe('Blocked')
    })

    it('renames a column', async () => {
      const instance = mockAxios.create()
      instance.put.mockResolvedValue({
        data: {
          data: { id: 'c1', name: 'Testing', position: 0, status: 'todo', board_id: 'b1' },
        },
      })
      const result = await columnsApi.rename('c1', { name: 'Testing' })
      expect(result.name).toBe('Testing')
    })

    it('reorders a column', async () => {
      const instance = mockAxios.create()
      instance.put.mockResolvedValue({
        data: {
          data: [
            { id: 'c2', name: 'B', position: 0, status: 'todo', board_id: 'b1' },
            { id: 'c1', name: 'A', position: 1, status: 'todo', board_id: 'b1' },
          ],
        },
      })
      const result = await columnsApi.reorder('c1', { position: 1 })
      expect(result).toHaveLength(2)
    })

    it('deletes a column', async () => {
      const instance = mockAxios.create()
      instance.delete.mockResolvedValue({ status: 204 })
      await columnsApi.delete('c1')
      expect(instance.delete).toHaveBeenCalledWith('/columns/c1')
    })
  })
})
