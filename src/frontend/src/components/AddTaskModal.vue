<script setup lang="ts">
import { ref } from 'vue'
import { tasksApi } from '@/api'
import { useNotifications } from '@/composables/useNotifications'
import type { Task, Member, TaskPriority } from '@/types'

const props = defineProps<{
  columnId: string
  members: Member[]
}>()

const emit = defineEmits<{
  close: []
  created: [task: Task]
}>()

const { notify } = useNotifications()

const form = ref({
  title: '',
  description: '',
  assignee_id: '' as string,
  due_date: '',
  priority: 'medium' as TaskPriority,
})
const creating = ref(false)
const errors = ref<Array<{ field: string; message: string }>>([])

const priorityOptions: TaskPriority[] = ['low', 'medium', 'high', 'critical']

async function createTask() {
  if (!form.value.title.trim()) {
    errors.value = [{ field: 'title', message: 'Title is required' }]
    return
  }

  creating.value = true
  errors.value = []

  try {
    const task = await tasksApi.create(props.columnId, {
      title: form.value.title.trim(),
      description: form.value.description.trim() || undefined,
      assignee_id: form.value.assignee_id || undefined,
      due_date: form.value.due_date ? new Date(form.value.due_date).toISOString() : undefined,
      priority: form.value.priority,
    })
    emit('created', task)
    notify('success', 'Task created')
  } catch (e: any) {
    const err = e.response?.data?.error
    if (err?.details) {
      errors.value = err.details
    }
    notify('error', err?.message || 'Failed to create task')
  } finally {
    creating.value = false
  }
}

function getFieldError(field: string): string | undefined {
  return errors.value.find((e) => e.field === field)?.message
}
</script>

<template>
  <div class="modal-overlay" @click.self="emit('close')">
    <div class="modal">
      <div class="modal-header">
        <h2>Add Task</h2>
        <button class="btn btn-sm btn-secondary" @click="emit('close')">✕</button>
      </div>
      <form @submit.prevent="createTask">
        <div class="modal-body">
          <div class="form-group">
            <label for="taskTitle">Title</label>
            <input
              id="taskTitle"
              v-model="form.title"
              class="form-control"
              placeholder="What needs to be done?"
              maxlength="200"
            />
            <span v-if="getFieldError('title')" class="field-error">{{ getFieldError('title') }}</span>
          </div>

          <div class="form-group">
            <label for="taskDesc">Description</label>
            <textarea
              id="taskDesc"
              v-model="form.description"
              class="form-control"
              placeholder="Add more details..."
              maxlength="2000"
            />
          </div>

          <div class="form-row">
            <div class="form-group" style="flex: 1">
              <label for="taskPriority">Priority</label>
              <select id="taskPriority" v-model="form.priority" class="form-control">
                <option v-for="p in priorityOptions" :key="p" :value="p">{{ p }}</option>
              </select>
            </div>

            <div class="form-group" style="flex: 1">
              <label for="taskAssignee">Assignee</label>
              <select id="taskAssignee" v-model="form.assignee_id" class="form-control">
                <option value="">Unassigned</option>
                <option v-for="m in members" :key="m.id" :value="m.id">{{ m.name }}</option>
              </select>
            </div>
          </div>

          <div class="form-group">
            <label for="taskDue">Due Date</label>
            <input id="taskDue" v-model="form.due_date" type="datetime-local" class="form-control" />
          </div>
        </div>

        <div class="modal-footer">
          <button type="button" class="btn btn-secondary" @click="emit('close')">Cancel</button>
          <button type="submit" class="btn btn-primary" :disabled="creating">
            <span v-if="creating" class="spinner" style="width: 14px; height: 14px; border-width: 2px" />
            {{ creating ? 'Creating...' : 'Create Task' }}
          </button>
        </div>
      </form>
    </div>
  </div>
</template>

<style scoped>
.form-row {
  display: flex;
  gap: 12px;
}

.field-error {
  display: block;
  color: var(--color-danger);
  font-size: 12px;
  margin-top: 4px;
}
</style>
