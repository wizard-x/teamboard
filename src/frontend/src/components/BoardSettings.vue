<script setup lang="ts">
import { ref } from 'vue'
import { boardsApi, columnsApi } from '@/api'
import { useNotifications } from '@/composables/useNotifications'
import type { BoardDetail } from '@/types'

const props = defineProps<{
  board: BoardDetail
}>()

const emit = defineEmits<{
  close: []
  updated: []
}>()

const { notify } = useNotifications()

const editBoardForm = ref({
  name: props.board.name,
  description: props.board.description || '',
})
const savingBoard = ref(false)

const editingColumnId = ref<string | null>(null)
const editColumnName = ref('')
const savingColumn = ref(false)

async function saveBoard() {
  savingBoard.value = true
  try {
    await boardsApi.update(props.board.id, editBoardForm.value)
    notify('success', 'Board updated')
    emit('updated')
    emit('close')
  } catch (e: any) {
    notify('error', e.response?.data?.error?.message || 'Failed to update board')
  } finally {
    savingBoard.value = false
  }
}

function startEditColumn(id: string, currentName: string) {
  editingColumnId.value = id
  editColumnName.value = currentName
}

async function saveColumnName() {
  if (!editingColumnId.value || !editColumnName.value.trim()) return
  savingColumn.value = true
  try {
    await columnsApi.rename(editingColumnId.value, { name: editColumnName.value.trim() })
    editingColumnId.value = null
    notify('success', 'Column renamed')
    emit('updated')
  } catch (e: any) {
    notify('error', e.response?.data?.error?.message || 'Failed to rename column')
  } finally {
    savingColumn.value = false
  }
}
</script>

<template>
  <div class="modal-overlay" @click.self="emit('close')">
    <div class="modal">
      <div class="modal-header">
        <h2>Board Settings</h2>
        <button class="btn btn-sm btn-secondary" @click="emit('close')">✕</button>
      </div>

      <div class="modal-body">
        <!-- Board Info -->
        <h3 style="margin-bottom: 12px; font-size: 14px">Board Info</h3>
        <div class="form-group">
          <label>Name</label>
          <input v-model="editBoardForm.name" class="form-control" maxlength="100" />
        </div>
        <div class="form-group">
          <label>Description</label>
          <textarea v-model="editBoardForm.description" class="form-control" maxlength="500" />
        </div>
        <button class="btn btn-sm btn-primary" :disabled="savingBoard" @click="saveBoard">
          {{ savingBoard ? 'Saving...' : 'Save Board' }}
        </button>

        <!-- Columns -->
        <h3 style="margin: 20px 0 12px; font-size: 14px">Columns</h3>
        <div class="columns-list">
          <div v-for="column in board.columns" :key="column.id" class="column-item">
            <template v-if="editingColumnId === column.id">
              <input v-model="editColumnName" class="form-control" style="flex: 1" />
              <button class="btn btn-sm btn-primary" :disabled="savingColumn" @click="saveColumnName">Save</button>
              <button class="btn btn-sm btn-secondary" @click="editingColumnId = null">Cancel</button>
            </template>
            <template v-else>
              <span>{{ column.name }}</span>
              <span class="column-position">Position: {{ column.position }}</span>
              <button class="btn btn-sm btn-secondary" @click="startEditColumn(column.id, column.name)">Rename</button>
            </template>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.columns-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.column-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  background: var(--color-bg);
  border-radius: var(--radius);
  font-size: 14px;
}

.column-position {
  color: var(--color-text-muted);
  font-size: 12px;
  margin-left: auto;
}
</style>
