import { describe, it, expect } from 'vitest'
import type {
  Task,
  Board,
  Column,
  TaskPriority,
  ColumnStatus,
  Member,
  Comment,
  ApiError,
} from '@/types'

describe('TypeScript types', () => {
  it('creates a valid Task object', () => {
    const task: Task = {
      id: '01HZ3TASK000000000001',
      title: 'Implement auth middleware',
      description: 'Add JWT validation',
      status: 'todo',
      priority: 'high',
      position: 0,
      column_id: '01HZ3COL00000000001',
      board_id: '01HZ3BOARD0000000000000001',
      assignee: { id: '01HZ3MEM0001', name: 'John Doe' },
      due_date: '2026-02-01T00:00:00Z',
      comment_count: 3,
      created_at: '2026-01-15T11:00:00Z',
      updated_at: '2026-01-15T14:00:00Z',
    }
    expect(task.id).toBeTruthy()
    expect(task.priority).toBe('high')
  })

  it('creates a Task with null optional fields', () => {
    const task: Task = {
      id: 't1',
      title: 'Simple task',
      status: 'todo',
      priority: 'medium',
      position: 0,
      column_id: 'c1',
      board_id: 'b1',
      assignee: null,
      due_date: null,
      comment_count: 0,
      created_at: '2026-01-01T00:00:00Z',
      updated_at: '2026-01-01T00:00:00Z',
    }
    expect(task.assignee).toBeNull()
    expect(task.due_date).toBeNull()
  })

  it('validates TaskPriority values', () => {
    const priorities: TaskPriority[] = ['low', 'medium', 'high', 'critical']
    expect(priorities).toHaveLength(4)
  })

  it('validates ColumnStatus values', () => {
    const statuses: ColumnStatus[] = ['todo', 'in_progress', 'review', 'done']
    expect(statuses).toHaveLength(4)
  })

  it('creates a valid Board object', () => {
    const board: Board = {
      id: 'b1',
      name: 'Sprint 42',
      description: 'Sprint board for Q1 2026',
      created_at: '2026-01-15T10:30:00Z',
      updated_at: '2026-01-15T10:30:00Z',
    }
    expect(board.name).toBe('Sprint 42')
  })

  it('creates a Column with tasks', () => {
    const column: Column = {
      id: 'c1',
      name: 'Todo',
      position: 0,
      status: 'todo',
      tasks: [],
    }
    expect(column.tasks).toEqual([])
  })

  it('creates a Comment object', () => {
    const comment: Comment = {
      id: 'cm1',
      body: 'This looks good',
      author: { id: 'm1', name: 'John' },
      created_at: '2026-01-15T12:00:00Z',
      updated_at: '2026-01-15T12:00:00Z',
    }
    expect(comment.body).toBe('This looks good')
  })

  it('creates a Member object', () => {
    const member: Member = {
      id: 'm1',
      name: 'John Doe',
      email: 'john@example.com',
      role: 'admin',
      created_at: '2026-01-15T10:00:00Z',
    }
    expect(member.role).toBe('admin')
  })

  it('creates an ApiError structure', () => {
    const error: ApiError = {
      error: {
        code: 'VALIDATION_ERROR',
        message: 'Invalid input',
        details: [
          { field: 'name', message: 'Name is required' },
        ],
      },
    }
    expect(error.error.code).toBe('VALIDATION_ERROR')
    expect(error.error.details).toHaveLength(1)
  })
})
