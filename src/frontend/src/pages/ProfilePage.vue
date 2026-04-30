<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { membersApi } from '@/api'
import { useAuth } from '@/composables/useAuth'
import { useNotifications } from '@/composables/useNotifications'

const router = useRouter()
const { currentMember, setAuth, clearAuth, isAuthenticated } = useAuth()
const { notify } = useNotifications()

const editName = ref('')
const editing = ref(false)
const saving = ref(false)
const regenerating = ref(false)
const newApiKey = ref('')

onMounted(() => {
  if (!isAuthenticated.value) {
    router.push('/login')
    return
  }
  editName.value = currentMember.value?.name ?? ''
})

async function saveName() {
  if (!editName.value.trim()) {
    notify('error', 'Name cannot be empty')
    return
  }
  saving.value = true
  try {
    const updated = await membersApi.updateMe({ name: editName.value.trim() })
    if (currentMember.value) {
      setAuth(localStorage.getItem('api_key')!, { ...currentMember.value, name: updated.name })
    }
    editing.value = false
    notify('success', 'Name updated')
  } catch (e: any) {
    notify('error', e.response?.data?.error?.message || 'Failed to update name')
  } finally {
    saving.value = false
  }
}

async function regenerateKey() {
  if (!confirm('Are you sure? Your current API key will stop working immediately.')) return
  regenerating.value = true
  newApiKey.value = ''
  try {
    const result = await membersApi.regenerateMyKey()
    newApiKey.value = result.api_key
    const storedKey = localStorage.getItem('api_key')
    if (currentMember.value && storedKey) {
      setAuth(newApiKey.value, currentMember.value)
    }
    notify('success', 'API key regenerated')
  } catch (e: any) {
    notify('error', e.response?.data?.error?.message || 'Failed to regenerate key')
  } finally {
    regenerating.value = false
  }
}

function copyKey() {
  navigator.clipboard.writeText(newApiKey.value)
  notify('success', 'Copied to clipboard')
}

function logout() {
  clearAuth()
  router.push('/login')
}
</script>

<template>
  <div class="page-container">
    <div class="page-header">
      <h1>Profile</h1>
    </div>

    <div v-if="currentMember" class="profile-content">
      <!-- Profile Card -->
      <div class="card profile-card">
        <div class="profile-avatar">
          {{ currentMember.name.charAt(0).toUpperCase() }}
        </div>

        <div class="profile-info">
          <template v-if="editing">
            <div class="form-group">
              <label>Name</label>
              <input v-model="editName" class="form-control" maxlength="100" />
            </div>
            <div style="display: flex; gap: 8px">
              <button class="btn btn-sm btn-primary" :disabled="saving" @click="saveName">
                {{ saving ? 'Saving...' : 'Save' }}
              </button>
              <button class="btn btn-sm btn-secondary" @click="editing = false">Cancel</button>
            </div>
          </template>
          <template v-else>
            <h2>{{ currentMember.name }}</h2>
            <div class="profile-details">
              <div class="detail-item">
                <span class="detail-label">Email</span>
                <span>{{ currentMember.email }}</span>
              </div>
              <div class="detail-item">
                <span class="detail-label">Role</span>
                <span :class="['badge', currentMember.role === 'admin' ? 'badge-admin' : 'badge-member']">
                  {{ currentMember.role }}
                </span>
              </div>
              <div class="detail-item">
                <span class="detail-label">Joined</span>
                <span>{{ new Date(currentMember.created_at).toLocaleDateString() }}</span>
              </div>
            </div>
            <button class="btn btn-sm btn-secondary" style="margin-top: 12px" @click="editing = true">
              ✏️ Edit Name
            </button>
          </template>
        </div>
      </div>

      <!-- API Key Section -->
      <div class="card api-key-card">
        <h3>API Key</h3>
        <p class="hint">Your API key authenticates all API requests. Keep it secret.</p>

        <div v-if="newApiKey" class="new-key-display">
          <label>New API Key — save it now, you won't see it again:</label>
          <div class="key-row">
            <code>{{ newApiKey }}</code>
            <button class="btn btn-sm btn-secondary" @click="copyKey">Copy</button>
          </div>
        </div>

        <div class="key-actions">
          <button
            class="btn btn-sm btn-danger"
            :disabled="regenerating"
            @click="regenerateKey"
          >
            {{ regenerating ? 'Regenerating...' : '🔄 Regenerate API Key' }}
          </button>
        </div>
      </div>

      <!-- Danger Zone -->
      <div class="card danger-card">
        <h3>Session</h3>
        <p class="hint">Sign out of your current session.</p>
        <button class="btn btn-danger" @click="logout">Logout</button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.profile-content {
  display: flex;
  flex-direction: column;
  gap: 20px;
  max-width: 640px;
}

.profile-card {
  display: flex;
  gap: 20px;
  align-items: flex-start;
}

.profile-avatar {
  width: 64px;
  height: 64px;
  border-radius: 50%;
  background: var(--color-primary);
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 28px;
  font-weight: 700;
  flex-shrink: 0;
}

.profile-info {
  flex: 1;
  min-width: 0;
}

.profile-info h2 {
  font-size: 20px;
  font-weight: 600;
  margin-bottom: 12px;
}

.profile-details {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.detail-item {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 14px;
}

.detail-label {
  color: var(--color-text-muted);
  font-size: 12px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  width: 80px;
}

.api-key-card h3,
.danger-card h3 {
  font-size: 16px;
  font-weight: 600;
  margin-bottom: 4px;
}

.hint {
  font-size: 13px;
  color: var(--color-text-muted);
  margin-bottom: 12px;
}

.new-key-display {
  margin-bottom: 16px;
  padding: 12px;
  background: var(--color-bg);
  border-radius: var(--radius);
  border: 1px solid var(--color-border);
}

.new-key-display label {
  display: block;
  font-size: 12px;
  font-weight: 600;
  color: var(--color-danger);
  margin-bottom: 8px;
}

.key-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.key-row code {
  flex: 1;
  font-size: 13px;
  word-break: break-all;
  background: var(--color-surface);
  padding: 8px 12px;
  border-radius: var(--radius);
  border: 1px solid var(--color-border);
}

.key-actions {
  display: flex;
  gap: 8px;
}

.danger-card {
  border-color: var(--color-danger);
}
</style>
