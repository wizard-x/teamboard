import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { createRouter, createMemoryHistory } from 'vue-router'
import TaskCard from '@/components/TaskCard.vue'
import type { Task } from '@/types'

function createMockTask(overrides: Partial<Task> = {}): Task {
  return {
    id: 'task-1',
    title: 'Test Task',
    description: 'A test task description',
    status: 'todo',
    priority: 'medium',
    position: 0,
    column_id: 'col-1',
    board_id: 'board-1',
    assignee: { id: 'mem-1', name: 'John Doe' },
    due_date: '2026-02-01T00:00:00Z',
    comment_count: 3,
    created_at: '2026-01-15T11:00:00Z',
    updated_at: '2026-01-15T14:00:00Z',
    ...overrides,
  }
}

describe('TaskCard', () => {
  it('renders task title', () => {
    const task = createMockTask({ title: 'My Task Title' })
    const wrapper = mount(TaskCard, { props: { task } })
    expect(wrapper.text()).toContain('My Task Title')
  })

  it('renders assignee name', () => {
    const task = createMockTask({ assignee: { id: 'mem-1', name: 'Jane Doe' } })
    const wrapper = mount(TaskCard, { props: { task } })
    expect(wrapper.text()).toContain('Jane Doe')
  })

  it('does not render assignee section when unassigned', () => {
    const task = createMockTask({ assignee: null })
    const wrapper = mount(TaskCard, { props: { task } })
    expect(wrapper.text()).not.toContain('👤')
  })

  it('renders comment count when > 0', () => {
    const task = createMockTask({ comment_count: 5 })
    const wrapper = mount(TaskCard, { props: { task } })
    expect(wrapper.text()).toContain('💬 5')
  })

  it('does not render comment count when 0', () => {
    const task = createMockTask({ comment_count: 0 })
    const wrapper = mount(TaskCard, { props: { task } })
    expect(wrapper.text()).not.toContain('💬')
  })

  it('renders due date', () => {
    const task = createMockTask({ due_date: '2026-03-15T00:00:00Z' })
    const wrapper = mount(TaskCard, { props: { task } })
    expect(wrapper.text()).toContain('📅')
  })

  it('does not render due date when null', () => {
    const task = createMockTask({ due_date: null })
    const wrapper = mount(TaskCard, { props: { task } })
    expect(wrapper.text()).not.toContain('📅')
  })

  it('renders description indicator when task has description', () => {
    const task = createMockTask({ description: 'Has desc' })
    const wrapper = mount(TaskCard, { props: { task } })
    expect(wrapper.text()).toContain('📝')
  })

  it('does not render description indicator when no description', () => {
    const task = createMockTask({ description: undefined })
    const wrapper = mount(TaskCard, { props: { task } })
    expect(wrapper.text()).not.toContain('📝')
  })

  it('emits click event when clicked', async () => {
    const task = createMockTask()
    const wrapper = mount(TaskCard, { props: { task } })
    await wrapper.trigger('click')
    expect(wrapper.emitted('click')).toHaveLength(1)
  })

  it('has draggable attribute', () => {
    const task = createMockTask()
    const wrapper = mount(TaskCard, { props: { task } })
    expect(wrapper.find('.task-card').attributes('draggable')).toBe('true')
  })

  it('emits dragstart when drag starts', async () => {
    const task = createMockTask()
    const wrapper = mount(TaskCard, { props: { task } })
    await wrapper.trigger('dragstart')
    expect(wrapper.emitted('dragstart')).toHaveLength(1)
  })

  it('applies overdue class for past due dates', () => {
    const pastDate = new Date()
    pastDate.setFullYear(pastDate.getFullYear() - 1)
    const task = createMockTask({ due_date: pastDate.toISOString() })
    const wrapper = mount(TaskCard, { props: { task } })
    const dueEl = wrapper.find('.due-date')
    expect(dueEl.exists()).toBe(true)
    expect(dueEl.classes()).toContain('overdue')
  })

  it('renders priority indicator with correct color for critical', () => {
    const task = createMockTask({ priority: 'critical' })
    const wrapper = mount(TaskCard, { props: { task } })
    const priorityBar = wrapper.find('.task-priority')
    expect(priorityBar.exists()).toBe(true)
    expect(priorityBar.attributes('style')).toContain('var(--priority-critical)')
  })

  it('renders priority indicator with correct color for high', () => {
    const task = createMockTask({ priority: 'high' })
    const wrapper = mount(TaskCard, { props: { task } })
    const priorityBar = wrapper.find('.task-priority')
    expect(priorityBar.attributes('style')).toContain('var(--priority-high)')
  })

  it('renders priority indicator with correct color for low', () => {
    const task = createMockTask({ priority: 'low' })
    const wrapper = mount(TaskCard, { props: { task } })
    const priorityBar = wrapper.find('.task-priority')
    expect(priorityBar.attributes('style')).toContain('var(--priority-low)')
  })
})
