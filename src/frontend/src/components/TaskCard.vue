<script setup lang="ts">
import type { Task } from '@/types'

defineProps<{ task: Task }>()

defineEmits<{
  click: []
  dragstart: []
}>()

const priorityColors: Record<string, string> = {
  critical: 'var(--priority-critical)',
  high: 'var(--priority-high)',
  medium: 'var(--priority-medium)',
  low: 'var(--priority-low)',
}

function formatDate(dateStr: string | null): string {
  if (!dateStr) return ''
  return new Date(dateStr).toLocaleDateString(undefined, { month: 'short', day: 'numeric' })
}

function isOverdue(dateStr: string | null): boolean {
  if (!dateStr) return false
  return new Date(dateStr) < new Date()
}
</script>

<template>
  <div
    class="task-card"
    draggable="true"
    @dragstart="$emit('dragstart')"
    @click="$emit('click')"
  >
    <div class="task-priority" :style="{ background: priorityColors[task.priority] || priorityColors.medium }" />

    <div class="task-content">
      <div class="task-title">{{ task.title }}</div>

      <div class="task-meta">
        <span v-if="task.assignee" class="meta-item" title="Assignee">
          👤 {{ task.assignee.name }}
        </span>

        <span
          v-if="task.due_date"
          class="meta-item due-date"
          :class="{ overdue: isOverdue(task.due_date) }"
          title="Due date"
        >
          📅 {{ formatDate(task.due_date) }}
        </span>

        <span v-if="task.comment_count > 0" class="meta-item" title="Comments">
          💬 {{ task.comment_count }}
        </span>

        <span v-if="task.description" class="meta-item" title="Has description">
          📝
        </span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.task-card {
  background: var(--color-surface);
  border-radius: var(--radius);
  box-shadow: var(--shadow);
  display: flex;
  cursor: grab;
  transition: box-shadow 0.15s ease, transform 0.1s ease;
  overflow: hidden;
}

.task-card:hover {
  box-shadow: var(--shadow-lg);
}

.task-card:active {
  cursor: grabbing;
  transform: rotate(2deg);
}

.task-priority {
  width: 4px;
  flex-shrink: 0;
}

.task-content {
  padding: 10px 12px;
  flex: 1;
  min-width: 0;
}

.task-title {
  font-size: 14px;
  font-weight: 500;
  margin-bottom: 6px;
  word-wrap: break-word;
}

.task-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  font-size: 12px;
  color: var(--color-text-muted);
}

.meta-item {
  display: inline-flex;
  align-items: center;
  gap: 2px;
}

.due-date.overdue {
  color: var(--color-danger);
  font-weight: 600;
}
</style>
