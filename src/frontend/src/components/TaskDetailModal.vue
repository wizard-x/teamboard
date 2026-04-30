<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { tasksApi, commentsApi } from '@/api'
import type { TaskDetail, Comment, Member, TaskPriority, UpdateTaskRequest, Column } from '@/types'

const props = defineProps<{
  taskId: string
  members: Member[]
  columns: Column[]
}>()

const emit = defineEmits<{
  close: []
  deleted: [taskId: string]
  updated: []
}>()

const task = ref<TaskDetail | null>(null)
const comments = ref<Comment[]>([])
const loading = ref(true)
const newComment = ref('')
const submittingComment = ref(false)
const editing = ref(false)
const editForm = ref<UpdateTaskRequest & { column_id?: string }>({})
const saving = ref(false)

const currentColumn = computed(() => {
  if (!task.value) return null
  return props.columns.find((c) => c.id === task.value!.column_id) ?? null
})

onMounted(async () => {
  try {
    task.value = await tasksApi.get(props.taskId)
    if (!task.value) return
    editForm.value = {
      title: task.value.title,
      description: task.value.description,
      priority: task.value.priority,
      assignee_id: task.value.assignee?.id ?? null,
      due_date: task.value.due_date,
      column_id: task.value.column_id,
    }
    const commentsRes = await commentsApi.list(props.taskId)
    comments.value = commentsRes.data
  } catch {
    // Error handled silently
  } finally {
    loading.value = false
  }
})

async function submitComment() {
  if (!newComment.value.trim() || !task.value) return
  submittingComment.value = true
  try {
    const comment = await commentsApi.create(task.value.id, { body: newComment.value.trim() })
    comments.value.push(comment)
    task.value.comment_count++
    newComment.value = ''
  } catch {
    // Silently fail
  } finally {
    submittingComment.value = false
  }
}

async function deleteComment(commentId: string) {
  try {
    await commentsApi.delete(commentId)
    comments.value = comments.value.filter((c) => c.id !== commentId)
    if (task.value) task.value.comment_count--
  } catch {
    // Silently fail
  }
}

async function saveTask() {
  if (!task.value) return
  saving.value = true
  try {
    const newColumnId = editForm.value.column_id
    const oldColumnId = task.value?.column_id

    // If column changed, move the task
    if (newColumnId && oldColumnId && newColumnId !== oldColumnId) {
      const targetCol = props.columns.find((c) => c.id === newColumnId)
      await tasksApi.move(task.value.id, {
        column_id: newColumnId,
      })
      task.value.column_id = newColumnId
      if (targetCol) {
        task.value.status = targetCol.status
      }
    }

    // Update other fields
    const { column_id: _cid, ...updateFields } = editForm.value
    const updated = await tasksApi.update(task.value.id, updateFields)
    task.value = { ...task.value, ...updated }
    editing.value = false
    emit('updated')
  } catch {
    // Silently fail
  } finally {
    saving.value = false
  }
}

async function deleteTask() {
  if (!task.value) return
  emit('deleted', task.value.id)
}

const priorityOptions: TaskPriority[] = ['low', 'medium', 'high', 'critical']

function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleString()
}

function columnName(columnId: string): string {
  return props.columns.find((c) => c.id === columnId)?.name ?? columnId
}

function formatDueDateForInput(dateStr: string | null | undefined): string {
  if (!dateStr) return ''
  const d = new Date(dateStr)
  return d.toISOString().slice(0, 16)
}
</script>

<template>
  <div class="modal-overlay" @click.self="emit('close')">
    <div class="modal modal-lg">
      <div class="modal-header">
        <h2>Task Details</h2>
        <button class="btn btn-sm btn-secondary" @click="emit('close')">✕</button>
      </div>

      <div v-if="loading" class="loading-container">
        <span class="spinner" />
      </div>

      <template v-else-if="task">
        <div class="modal-body">
          <!-- View Mode -->
          <template v-if="!editing">
            <div class="detail-row">
              <label>Title</label>
              <div class="detail-value">{{ task.title }}</div>
            </div>

            <div class="detail-row">
              <label>Description</label>
              <div class="detail-value">
                <span v-if="task.description" style="white-space: pre-wrap">{{ task.description }}</span>
                <span v-else class="empty">No description</span>
              </div>
            </div>

            <div class="detail-grid">
              <div class="detail-row">
                <label>Status</label>
                <div class="detail-value">
                  <span class="status-badge" :class="`status-${task.status}`">
                    {{ currentColumn?.name || task.status }}
                  </span>
                </div>
              </div>

              <div class="detail-row">
                <label>Priority</label>
                <div class="detail-value">
                  <span class="priority-badge" :class="`priority-${task.priority}`">{{ task.priority }}</span>
                </div>
              </div>

              <div class="detail-row">
                <label>Assignee</label>
                <div class="detail-value">
                  <span v-if="task.assignee">👤 {{ task.assignee.name }}</span>
                  <span v-else class="empty">Unassigned</span>
                </div>
              </div>

              <div class="detail-row">
                <label>Due Date</label>
                <div class="detail-value">
                  <span v-if="task.due_date">{{ formatDate(task.due_date) }}</span>
                  <span v-else class="empty">No due date</span>
                </div>
              </div>
            </div>

            <div class="detail-actions">
              <button class="btn btn-sm btn-secondary" @click="editing = true">✏️ Edit</button>
              <button class="btn btn-sm btn-danger" @click="deleteTask">🗑️ Delete</button>
            </div>
          </template>

          <!-- Edit Mode -->
          <template v-else>
            <div class="form-group">
              <label>Title</label>
              <input v-model="editForm.title" class="form-control" maxlength="200" />
            </div>
            <div class="form-group">
              <label>Description</label>
              <textarea v-model="editForm.description" class="form-control" maxlength="2000" />
            </div>
            <div class="detail-grid">
              <div class="form-group">
                <label>Status (Column)</label>
                <select v-model="editForm.column_id" class="form-control">
                  <option v-for="col in columns" :key="col.id" :value="col.id">
                    {{ col.name }}
                  </option>
                </select>
              </div>
              <div class="form-group">
                <label>Priority</label>
                <select v-model="editForm.priority" class="form-control">
                  <option v-for="p in priorityOptions" :key="p" :value="p">{{ p }}</option>
                </select>
              </div>
              <div class="form-group">
                <label>Assignee</label>
                <select v-model="editForm.assignee_id" class="form-control">
                  <option :value="null">Unassigned</option>
                  <option v-for="m in members" :key="m.id" :value="m.id">{{ m.name }}</option>
                </select>
              </div>
              <div class="form-group">
                <label>Due Date</label>
                <input v-model="editForm.due_date" type="datetime-local" class="form-control" />
              </div>
            </div>
            <div style="display: flex; gap: 8px; margin-top: 12px">
              <button class="btn btn-primary" :disabled="saving" @click="saveTask">
                {{ saving ? 'Saving...' : 'Save' }}
              </button>
              <button class="btn btn-secondary" @click="editing = false">Cancel</button>
            </div>
          </template>

          <!-- Comments Section -->
          <div class="comments-section">
            <h3>Comments ({{ comments.length }})</h3>

            <div class="comment-list">
              <div v-for="comment in comments" :key="comment.id" class="comment-item">
                <div class="comment-header">
                  <strong>{{ comment.author.name }}</strong>
                  <span class="comment-date">{{ formatDate(comment.created_at) }}</span>
                  <button class="btn btn-sm btn-secondary" style="margin-left: auto; padding: 2px 6px" @click="deleteComment(comment.id)">✕</button>
                </div>
                <div class="comment-body">{{ comment.body }}</div>
              </div>
              <div v-if="comments.length === 0" class="empty">No comments yet</div>
            </div>

            <form class="comment-form" @submit.prevent="submitComment">
              <textarea
                v-model="newComment"
                class="form-control"
                placeholder="Add a comment..."
                maxlength="2000"
                rows="2"
              />
              <button class="btn btn-sm btn-primary" :disabled="submittingComment || !newComment.trim()">
                {{ submittingComment ? 'Posting...' : 'Post' }}
              </button>
            </form>
          </div>
        </div>
      </template>
    </div>
  </div>
</template>

<style scoped>
.modal-lg {
  max-width: 720px;
}

.detail-row {
  margin-bottom: 12px;
}

.detail-row label {
  display: block;
  font-size: 12px;
  font-weight: 600;
  color: var(--color-text-muted);
  text-transform: uppercase;
  letter-spacing: 0.5px;
  margin-bottom: 4px;
}

.detail-value {
  font-size: 14px;
}

.detail-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
}

.empty {
  color: var(--color-text-muted);
  font-style: italic;
}

.priority-badge {
  display: inline-block;
  padding: 2px 8px;
  border-radius: 999px;
  font-size: 12px;
  font-weight: 600;
}

.priority-low { background: #d1e7dd; color: #0f5132; }
.priority-medium { background: #fff3cd; color: #856404; }
.priority-high { background: #ffe0cc; color: #9a3412; }
.priority-critical { background: #f8d7da; color: #842029; }

.status-badge {
  display: inline-block;
  padding: 2px 8px;
  border-radius: 999px;
  font-size: 12px;
  font-weight: 600;
}

.status-todo { background: #e2e8f0; color: #475569; }
.status-in_progress { background: #dbeafe; color: #1e40af; }
.status-review { background: #fef3c7; color: #92400e; }
.status-done { background: #d1fae5; color: #065f46; }

.detail-actions {
  display: flex;
  gap: 8px;
  margin-top: 16px;
  padding-top: 16px;
  border-top: 1px solid var(--color-border);
}

.comments-section {
  margin-top: 20px;
  padding-top: 16px;
  border-top: 1px solid var(--color-border);
}

.comments-section h3 {
  font-size: 14px;
  font-weight: 600;
  margin-bottom: 12px;
}

.comment-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-bottom: 12px;
}

.comment-item {
  background: var(--color-bg);
  border-radius: var(--radius);
  padding: 8px 12px;
}

.comment-header {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  margin-bottom: 4px;
}

.comment-date {
  color: var(--color-text-muted);
  font-size: 12px;
}

.comment-body {
  font-size: 14px;
  white-space: pre-wrap;
}

.comment-form {
  display: flex;
  gap: 8px;
  align-items: flex-end;
}

.comment-form textarea {
  flex: 1;
  min-height: 48px;
}
</style>
