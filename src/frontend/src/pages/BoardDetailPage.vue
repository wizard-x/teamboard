<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { boardsApi, tasksApi, columnsApi } from '@/api'
import { useNotifications } from '@/composables/useNotifications'
import { useAuth } from '@/composables/useAuth'
import type { BoardDetail, Column, Task, Member } from '@/types'
import TaskCard from '@/components/TaskCard.vue'
import TaskDetailModal from '@/components/TaskDetailModal.vue'
import AddTaskModal from '@/components/AddTaskModal.vue'
import BoardSettings from '@/components/BoardSettings.vue'
import { membersApi } from '@/api'

const route = useRoute()
const router = useRouter()
const { notify } = useNotifications()
const { isAdmin } = useAuth()

const board = ref<BoardDetail | null>(null)
const members = ref<Member[]>([])
const loading = ref(true)
const selectedTaskId = ref<string | null>(null)
const addTaskColumnId = ref<string | null>(null)
const showSettings = ref(false)

// Drag state
const draggingTaskId = ref<string | null>(null)
const dragSourceColumnId = ref<string | null>(null)
const dragOverColumnId = ref<string | null>(null)

// Column drag state
const draggingColumnId = ref<string | null>(null)
const dragOverColumnTarget = ref<string | null>(null)

// Filter state
const filterText = ref('')
const filterPriority = ref<string>('')
const filterAssignee = ref<string>('')
const showFilters = ref(false)

const hasActiveFilters = computed(() =>
  filterText.value !== '' || filterPriority.value !== '' || filterAssignee.value !== ''
)

const filteredColumns = computed(() => {
  if (!board.value) return []
  const text = filterText.value.toLowerCase().trim()
  return board.value.columns.map((col) => ({
    ...col,
    tasks: (col.tasks ?? []).filter((task) => {
      if (text && !task.title.toLowerCase().includes(text) && !(task.description?.toLowerCase().includes(text))) return false
      if (filterPriority.value && task.priority !== filterPriority.value) return false
      if (filterAssignee.value) {
        if (filterAssignee.value === '__none__') {
          if (task.assignee) return false
        } else {
          if (task.assignee?.id !== filterAssignee.value) return false
        }
      }
      return true
    }),
  }))
})

function clearFilters() {
  filterText.value = ''
  filterPriority.value = ''
  filterAssignee.value = ''
}

const boardId = computed(() => route.params.id as string)

onMounted(async () => {
  await loadBoard()
  await loadMembers()
})

async function loadBoard() {
  loading.value = true
  try {
    board.value = await boardsApi.get(boardId.value)
  } catch (e: any) {
    notify('error', 'Failed to load board')
    router.push('/boards')
  } finally {
    loading.value = false
  }
}

async function loadMembers() {
  try {
    members.value = await membersApi.list()
  } catch {
    // Silently fail — members are optional for board view
  }
}

// ── Task Drag & Drop ──

function onTaskDragStart(taskId: string, columnId: string) {
  draggingTaskId.value = taskId
  dragSourceColumnId.value = columnId
  // Clear column drag
  draggingColumnId.value = null
}

function onTaskDragOver(e: DragEvent, columnId: string) {
  if (!draggingTaskId.value) return
  e.preventDefault()
  if (e.dataTransfer) {
    e.dataTransfer.dropEffect = 'move'
  }
  dragOverColumnId.value = columnId
}

function onTaskDragLeave(columnId: string) {
  if (dragOverColumnId.value === columnId) {
    dragOverColumnId.value = null
  }
}

async function onTaskDrop(targetColumnId: string, position?: number) {
  const taskId = draggingTaskId.value
  const sourceId = dragSourceColumnId.value
  if (!taskId) return

  draggingTaskId.value = null
  dragSourceColumnId.value = null
  dragOverColumnId.value = null

  const sourceColumn = board.value?.columns.find((c) => c.id === sourceId)
  const targetColumn = board.value?.columns.find((c) => c.id === targetColumnId)

  if (!sourceColumn || !targetColumn || !board.value) return

  const taskIndex = sourceColumn.tasks?.findIndex((t) => t.id === taskId) ?? -1
  if (taskIndex === -1) return

  const [task] = sourceColumn.tasks!.splice(taskIndex, 1)

  sourceColumn.tasks?.forEach((t, i) => {
    t.position = i
  })

  task.column_id = targetColumnId
  task.status = targetColumn.status

  const insertPos = position ?? targetColumn.tasks?.length ?? 0
  targetColumn.tasks!.splice(insertPos, 0, task)

  targetColumn.tasks?.forEach((t, i) => {
    t.position = i
  })

  try {
    await tasksApi.move(taskId, {
      column_id: targetColumnId,
      position: insertPos,
    })
  } catch {
    notify('error', 'Failed to move task. Reverting...')
    await loadBoard()
  }
}

function onTaskColumnDrop(e: DragEvent, columnId: string) {
  e.preventDefault()
  dragOverColumnId.value = null
  onTaskDrop(columnId)
}

// ── Column Drag & Drop ──

function onColumnDragStart(columnId: string, e: DragEvent) {
  draggingColumnId.value = columnId
  // Clear task drag
  draggingTaskId.value = null
  dragSourceColumnId.value = null
  if (e.dataTransfer) {
    e.dataTransfer.effectAllowed = 'move'
    e.dataTransfer.dropEffect = 'move'
  }
}

function onColumnDragOver(e: DragEvent, columnId: string) {
  if (!draggingColumnId.value) return
  e.preventDefault()
  e.stopPropagation()
  if (e.dataTransfer) {
    e.dataTransfer.dropEffect = 'move'
  }
  dragOverColumnTarget.value = columnId
}

function onColumnDragLeave(columnId: string) {
  if (dragOverColumnTarget.value === columnId) {
    dragOverColumnTarget.value = null
  }
}

async function onColumnDrop(targetColumnId: string) {
  const sourceId = draggingColumnId.value
  if (!sourceId || sourceId === targetColumnId || !board.value) {
    draggingColumnId.value = null
    dragOverColumnTarget.value = null
    return
  }

  const columns = board.value.columns
  const sourceIdx = columns.findIndex((c) => c.id === sourceId)
  const targetIdx = columns.findIndex((c) => c.id === targetColumnId)
  if (sourceIdx === -1 || targetIdx === -1) return

  // Move column in local state
  const [moved] = columns.splice(sourceIdx, 1)
  columns.splice(targetIdx, 0, moved)

  // Update positions
  columns.forEach((c, i) => (c.position = i))

  draggingColumnId.value = null
  dragOverColumnTarget.value = null

  try {
    await columnsApi.reorder(sourceId, { position: targetIdx })
  } catch {
    notify('error', 'Failed to reorder columns. Reverting...')
    await loadBoard()
  }
}

function onColumnDropOnColumn(e: DragEvent, columnId: string) {
  e.preventDefault()
  e.stopPropagation()
  dragOverColumnTarget.value = null
  onColumnDrop(columnId)
}

function onColumnDragEnd() {
  draggingColumnId.value = null
  dragOverColumnTarget.value = null
}

// ── Task Actions ──

function openTaskDetail(taskId: string) {
  selectedTaskId.value = taskId
}

function closeTaskDetail() {
  selectedTaskId.value = null
}

function openAddTask(columnId: string) {
  addTaskColumnId.value = columnId
}

function closeAddTask() {
  addTaskColumnId.value = null
}

async function onTaskCreated(task: Task) {
  if (!board.value) return
  const column = board.value.columns.find((c) => c.id === task.column_id)
  if (column) {
    if (!column.tasks) column.tasks = []
    column.tasks.push(task)
  }
  addTaskColumnId.value = null
}

async function deleteTask(taskId: string) {
  if (!board.value) return
  try {
    await tasksApi.delete(taskId)
    for (const column of board.value.columns) {
      if (column.tasks) {
        column.tasks = column.tasks.filter((t) => t.id !== taskId)
        column.tasks.forEach((t, i) => (t.position = i))
      }
    }
    selectedTaskId.value = null
    notify('success', 'Task deleted')
  } catch {
    notify('error', 'Failed to delete task')
  }
}

async function deleteColumn(columnId: string) {
  if (!board.value) return
  try {
    await columnsApi.delete(columnId)
    board.value.columns = board.value.columns.filter((c) => c.id !== columnId)
    board.value.columns.forEach((c, i) => (c.position = i))
    notify('success', 'Column deleted')
  } catch (e: any) {
    notify('error', e.response?.data?.error?.message || 'Failed to delete column')
  }
}

async function addColumn() {
  if (!board.value) return
  const name = prompt('Column name:')
  if (!name?.trim()) return
  try {
    const column = await columnsApi.create(board.value.id, { name: name.trim() })
    board.value.columns.push({ ...column, tasks: [] })
    notify('success', 'Column added')
  } catch (e: any) {
    notify('error', e.response?.data?.error?.message || 'Failed to add column')
  }
}

async function deleteBoard() {
  if (!confirm('Delete this board and all its tasks? This cannot be undone.')) return
  try {
    await boardsApi.delete(boardId.value)
    notify('success', 'Board deleted')
    router.push('/boards')
  } catch (e: any) {
    notify('error', e.response?.data?.error?.message || 'Failed to delete board')
  }
}

function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleDateString()
}
</script>

<template>
  <div class="board-page">
    <div v-if="loading" class="loading-container">
      <span class="spinner" />
      Loading board...
    </div>

    <template v-else-if="board">
      <div class="board-header">
        <div class="board-title">
          <button class="btn btn-sm btn-secondary" @click="router.push('/boards')">← Back</button>
          <h1>{{ board.name }}</h1>
          <span v-if="board.description" class="board-desc">{{ board.description }}</span>
        </div>
        <div class="board-actions">
          <button
            class="btn btn-sm btn-secondary"
            :class="{ 'btn-active-filter': showFilters || hasActiveFilters }"
            @click="showFilters = !showFilters"
          >🔍 Filter</button>
          <button class="btn btn-sm btn-secondary" @click="addColumn">+ Column</button>
          <button class="btn btn-sm btn-secondary" @click="showSettings = true">⚙ Settings</button>
          <button v-if="isAdmin" class="btn btn-sm btn-danger" @click="deleteBoard">Delete Board</button>
        </div>
      </div>

      <div v-if="showFilters" class="filter-bar">
        <div class="filter-row">
          <input
            v-model="filterText"
            class="form-control filter-input"
            type="search"
            placeholder="Search tasks..."
            clearable
          />
          <select v-model="filterPriority" class="form-control filter-select">
            <option value="">All priorities</option>
            <option value="low">🟢 Low</option>
            <option value="medium">🟡 Medium</option>
            <option value="high">🟠 High</option>
            <option value="critical">🔴 Critical</option>
          </select>
          <select v-model="filterAssignee" class="form-control filter-select">
            <option value="">All assignees</option>
            <option value="__none__">Unassigned</option>
            <option v-for="m in members" :key="m.id" :value="m.id">{{ m.name }}</option>
          </select>
          <button v-if="hasActiveFilters" class="btn btn-sm btn-secondary" @click="clearFilters">✕ Clear</button>
        </div>
      </div>

      <div class="board-columns">
        <div
          v-for="column in filteredColumns"
          :key="column.id"
          class="column"
          :class="{
            'column-drag-over': dragOverColumnId === column.id && draggingTaskId,
            'column-dragging': draggingColumnId === column.id,
            'column-drop-target': dragOverColumnTarget === column.id && draggingColumnId,
          }"
          @dragover="(e) => { onTaskDragOver(e, column.id); onColumnDragOver(e, column.id) }"
          @dragleave="() => { onTaskDragLeave(column.id); onColumnDragLeave(column.id) }"
          @drop="onColumnDropOnColumn($event, column.id); if (draggingTaskId) onTaskColumnDrop($event, column.id)"
        >
          <div
            class="column-header"
            draggable="true"
            @dragstart="onColumnDragStart(column.id, $event)"
            @dragend="onColumnDragEnd"
          >
            <span class="column-grip" title="Drag to reorder">⠿</span>
            <h3>{{ column.name }}</h3>
            <span class="task-count">{{ column.tasks?.length || 0 }}</span>
            <div class="column-actions">
              <button
                class="btn btn-sm btn-secondary"
                style="padding: 2px 6px; font-size: 18px"
                @click="openAddTask(column.id)"
                title="Add task"
              >+</button>
              <button
                class="btn btn-sm btn-secondary"
                style="padding: 2px 6px; font-size: 12px"
                @click="deleteColumn(column.id)"
                title="Delete column"
              >✕</button>
            </div>
          </div>

          <div class="column-tasks">
            <TaskCard
              v-for="task in column.tasks"
              :key="task.id"
              :task="task"
              :status-label="column.name"
              :draggable="true"
              @dragstart="onTaskDragStart(task.id, column.id)"
              @click="openTaskDetail(task.id)"
            />

            <div v-if="!column.tasks?.length" class="column-empty">
              No tasks
            </div>
          </div>
        </div>
      </div>

      <!-- Task Detail Modal -->
      <TaskDetailModal
        v-if="selectedTaskId"
        :task-id="selectedTaskId"
        :members="members"
        :columns="board.columns"
        @close="closeTaskDetail"
        @deleted="deleteTask"
        @updated="loadBoard"
      />

      <!-- Add Task Modal -->
      <AddTaskModal
        v-if="addTaskColumnId"
        :column-id="addTaskColumnId"
        :members="members"
        @close="closeAddTask"
        @created="onTaskCreated"
      />

      <!-- Board Settings Modal -->
      <BoardSettings
        v-if="showSettings"
        :board="board"
        @close="showSettings = false"
        @updated="loadBoard"
      />
    </template>
  </div>
</template>

<style scoped>
.board-page {
  height: calc(100vh - 56px);
  display: flex;
  flex-direction: column;
}

.board-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 24px;
  background: var(--color-surface);
  border-bottom: 1px solid var(--color-border);
  flex-shrink: 0;
}

.board-title {
  display: flex;
  align-items: center;
  gap: 12px;
}

.board-title h1 {
  font-size: 20px;
  font-weight: 600;
}

.board-desc {
  font-size: 13px;
  color: var(--color-text-muted);
}

.board-actions {
  display: flex;
  gap: 8px;
}

.btn-active-filter {
  background: var(--color-primary) !important;
  color: #fff !important;
  border-color: var(--color-primary) !important;
}

.filter-bar {
  padding: 12px 24px;
  background: var(--color-surface);
  border-bottom: 1px solid var(--color-border);
  flex-shrink: 0;
}

.filter-row {
  display: flex;
  gap: 8px;
  align-items: center;
}

.filter-input {
  max-width: 260px;
}

.filter-select {
  max-width: 180px;
}

.board-columns {
  display: flex;
  gap: 16px;
  padding: 16px 24px;
  overflow-x: auto;
  flex: 1;
  align-items: flex-start;
}

.column {
  min-width: 300px;
  max-width: 320px;
  background: var(--color-bg);
  border-radius: var(--radius);
  display: flex;
  flex-direction: column;
  max-height: 100%;
  flex-shrink: 0;
  border: 2px solid transparent;
  transition: border-color 0.15s ease, opacity 0.15s ease;
}

.column.column-drag-over {
  border-color: var(--color-primary);
  background: color-mix(in srgb, var(--color-primary) 5%, var(--color-bg));
}

.column.column-dragging {
  opacity: 0.5;
}

.column.column-drop-target {
  border-color: var(--color-success);
  border-style: dashed;
}

.column-header {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 12px 8px;
  font-weight: 600;
  font-size: 14px;
  cursor: grab;
  user-select: none;
}

.column-header:active {
  cursor: grabbing;
}

.column-grip {
  color: var(--color-text-muted);
  font-size: 16px;
  line-height: 1;
  opacity: 0.5;
  transition: opacity 0.15s;
}

.column-header:hover .column-grip {
  opacity: 1;
}

.column-header h3 {
  flex: 1;
  font-size: 14px;
}

.task-count {
  background: var(--color-border);
  color: var(--color-text-muted);
  border-radius: 999px;
  padding: 1px 8px;
  font-size: 12px;
  font-weight: 400;
}

.column-actions {
  display: flex;
  gap: 4px;
}

.column-tasks {
  padding: 4px 8px 12px;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.column-empty {
  text-align: center;
  padding: 16px;
  color: var(--color-text-muted);
  font-size: 13px;
}
</style>
