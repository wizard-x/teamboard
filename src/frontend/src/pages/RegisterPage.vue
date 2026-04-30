<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { authApi } from '@/api'
import { useAuth } from '@/composables/useAuth'
import { useNotifications } from '@/composables/useNotifications'

const router = useRouter()
const { setAuth } = useAuth()
const { notify } = useNotifications()

const form = ref({
  name: '',
  admin_name: '',
  admin_email: '',
})
const loading = ref(false)
const errors = ref<Array<{ field: string; message: string }>>([])
const globalError = ref('')
const generatedKey = ref('')

async function handleRegister() {
  loading.value = true
  errors.value = []
  globalError.value = ''

  try {
    const result = await authApi.register(form.value)
    generatedKey.value = result.api_key
    setAuth(result.api_key, {
      id: result.member.id,
      name: result.member.name,
      email: result.member.email,
      role: result.member.role as 'admin' | 'member',
      created_at: new Date().toISOString(),
    })
    notify('success', 'Team registered successfully!')
  } catch (e: any) {
    if (e.response?.data?.error) {
      const err = e.response.data.error
      if (err.details) {
        errors.value = err.details
      }
      globalError.value = err.message
    } else {
      globalError.value = 'Registration failed. Please try again.'
    }
  } finally {
    loading.value = false
  }
}

function copyKey() {
  navigator.clipboard.writeText(generatedKey.value)
  notify('success', 'API key copied to clipboard!')
}

function goToBoards() {
  router.push('/boards')
}

function getFieldError(field: string): string | undefined {
  return errors.value.find((e) => e.field === field)?.message
}
</script>

<template>
  <div class="register-page">
    <div class="register-card card">
      <template v-if="!generatedKey">
        <h1>Register Your Team</h1>
        <p class="subtitle">Create a team to start managing tasks</p>

        <div v-if="globalError" class="error-banner">{{ globalError }}</div>

        <form @submit.prevent="handleRegister">
          <div class="form-group">
            <label for="teamName">Team Name</label>
            <input
              id="teamName"
              v-model="form.name"
              class="form-control"
              placeholder="Engineering Team"
              maxlength="100"
            />
            <span v-if="getFieldError('name')" class="field-error">{{ getFieldError('name') }}</span>
          </div>

          <div class="form-group">
            <label for="adminName">Your Name</label>
            <input
              id="adminName"
              v-model="form.admin_name"
              class="form-control"
              placeholder="John Doe"
              maxlength="100"
            />
            <span v-if="getFieldError('admin_name')" class="field-error">{{ getFieldError('admin_name') }}</span>
          </div>

          <div class="form-group">
            <label for="adminEmail">Your Email</label>
            <input
              id="adminEmail"
              v-model="form.admin_email"
              type="email"
              class="form-control"
              placeholder="john@example.com"
              maxlength="255"
            />
            <span v-if="getFieldError('admin_email')" class="field-error">{{ getFieldError('admin_email') }}</span>
          </div>

          <button class="btn btn-primary" style="width: 100%" :disabled="loading">
            <span v-if="loading" class="spinner" style="width: 16px; height: 16px; border-width: 2px" />
            {{ loading ? 'Creating...' : 'Create Team' }}
          </button>
        </form>

        <p class="login-link">
          Already have an account?
          <router-link to="/login">Login</router-link>
        </p>
      </template>

      <template v-else>
        <h1>🎉 Team Created!</h1>
        <p class="subtitle">Save your API key now — you won't be able to see it again.</p>

        <div class="key-display">
          <code>{{ generatedKey }}</code>
          <button class="btn btn-sm btn-secondary" @click="copyKey">Copy</button>
        </div>

        <button class="btn btn-primary" style="width: 100%; margin-top: 20px" @click="goToBoards">
          Go to Boards →
        </button>
      </template>
    </div>
  </div>
</template>

<style scoped>
.register-page {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: calc(100vh - 56px);
  padding: 24px;
}

.register-card {
  width: 100%;
  max-width: 460px;
}

.register-card h1 {
  font-size: 24px;
  margin-bottom: 4px;
}

.subtitle {
  color: var(--color-text-muted);
  font-size: 14px;
  margin-bottom: 24px;
}

.field-error {
  display: block;
  color: var(--color-danger);
  font-size: 12px;
  margin-top: 4px;
}

.key-display {
  background: var(--color-bg);
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  padding: 12px 16px;
  display: flex;
  align-items: center;
  gap: 12px;
}

.key-display code {
  flex: 1;
  font-size: 14px;
  word-break: break-all;
}

.login-link {
  margin-top: 16px;
  font-size: 14px;
  color: var(--color-text-muted);
  text-align: center;
}
</style>
