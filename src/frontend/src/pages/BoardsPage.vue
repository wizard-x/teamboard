<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { boardsApi } from '@/api'
import { useNotifications } from '@/composables/useNotifications'
import type { Board } from '@/types'

const router = useRouter()
const { notify } = useNotifications()

const boards = ref<Board[]>([])
const loading = ref(true)
const showCreate = ref(false)
const createForm = ref({ name: '', description: '' })
const creating = ref(false)

onMounted(async () => {
  try {
    boards.value = await boardsApi.list()
  } catch (e: any) {
    notify('error', 'Failed to load boards')
  } finally {
    loading.value = false
  }
})

async function createBoard() {
  if (!createForm.value.name.trim()) return
  creating.value = true
  try {
    const board = await boardsApi.create(createForm.value)
    boards.value.unshift({
      id: board.id,
      name: board.name,
      description: board.description,
      created_at: board.created_at,
      updated_at: board.updated_at,
    })
    showCreate.value = false
    createForm.value = { name: '', description: '' }
    notify('success', 'Board created!')
  } catch (e: any) {
    notify('error', e.response?.data?.error?.message || 'Failed to create board')
  } finally {
    creating.value = false
  }
}

function openBoard(id: string) {
  router.push(`/boards/${id}`)
}

function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleDateString()
}
</script>

<template>
  <div class="page-container">
    <div class="page-header">
      <h1>Boards</h1>
      <button class="btn btn-primary" @click="showCreate = true">+ New Board</button>
    </div>

    <div v-if="loading" class="loading-container">
      <span class="spinner" />
      Loading boards...
    </div>

    <div v-else-if="boards.length === 0" class="empty-state">
      <p>No boards yet. Create your first board to get started!</p>
      <button class="btn btn-primary" @click="showCreate = true">+ Create Board</button>
    </div>

    <div v-else class="boards-grid">
      <div
        v-for="board in boards"
        :key="board.id"
        class="board-card card"
        @click="openBoard(board.id)"
      >
        <h3>{{ board.name }}</h3>
        <p v-if="board.description" class="board-desc">{{ board.description }}</p>
        <span class="board-date">Created {{ formatDate(board.created_at) }}</span>
      </div>
    </div>

    <!-- Create Board Modal -->
    <div v-if="showCreate" class="modal-overlay" @click.self="showCreate = false">
      <div class="modal">
        <div class="modal-header">
          <h2>Create Board</h2>
          <button class="btn btn-sm btn-secondary" @click="showCreate = false">✕</button>
        </div>
        <form @submit.prevent="createBoard">
          <div class="modal-body">
            <div class="form-group">
              <label for="boardName">Board Name</label>
              <input
                id="boardName"
                v-model="createForm.name"
                class="form-control"
                placeholder="Sprint 42"
                maxlength="100"
              />
            </div>
            <div class="form-group">
              <label for="boardDesc">Description (optional)</label>
              <textarea
                id="boardDesc"
                v-model="createForm.description"
                class="form-control"
                placeholder="What's this board for?"
                maxlength="500"
              />
            </div>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" @click="showCreate = false">Cancel</button>
            <button type="submit" class="btn btn-primary" :disabled="creating">
              <span v-if="creating" class="spinner" style="width: 14px; height: 14px; border-width: 2px" />
              {{ creating ? 'Creating...' : 'Create' }}
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<style scoped>
.boards-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 16px;
}

.board-card {
  cursor: pointer;
  transition: box-shadow 0.15s ease, transform 0.15s ease;
}
.board-card:hover {
  box-shadow: var(--shadow-lg);
  transform: translateY(-2px);
}

.board-card h3 {
  font-size: 16px;
  margin-bottom: 4px;
}

.board-desc {
  font-size: 13px;
  color: var(--color-text-muted);
  margin-bottom: 8px;
}

.board-date {
  font-size: 12px;
  color: var(--color-text-muted);
}
</style>
