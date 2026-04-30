<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { authApi } from '@/api'
import { useAuth } from '@/composables/useAuth'
import { useNotifications } from '@/composables/useNotifications'

const router = useRouter()
const { setAuth } = useAuth()
const { notify } = useNotifications()

const apiKey = ref('')
const loading = ref(false)
const error = ref('')

async function handleLogin() {
  if (!apiKey.value.trim()) {
    error.value = 'Please enter your API key'
    return
  }

  loading.value = true
  error.value = ''

  try {
    // Store key temporarily to authenticate
    localStorage.setItem('api_key', apiKey.value.trim())

    // Fetch current user info from /members/me
    const { membersApi } = await import('@/api')
    const member = await membersApi.me()

    setAuth(apiKey.value.trim(), member)
    notify('success', 'Logged in successfully')
    router.push('/boards')
  } catch (e: any) {
    localStorage.removeItem('api_key')
    if (e.response?.status === 401) {
      error.value = 'Invalid API key'
    } else {
      error.value = e.response?.data?.error?.message || 'Login failed'
    }
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="login-page">
    <div class="login-card card">
      <h1>Login</h1>
      <p class="subtitle">Enter your API key to access your team board</p>

      <div v-if="error" class="error-banner">{{ error }}</div>

      <form @submit.prevent="handleLogin">
        <div class="form-group">
          <label for="apiKey">API Key</label>
          <input
            id="apiKey"
            v-model="apiKey"
            type="password"
            class="form-control"
            placeholder="tb_k1_..."
            autocomplete="off"
          />
        </div>

        <button class="btn btn-primary" style="width: 100%" :disabled="loading">
          <span v-if="loading" class="spinner" style="width: 16px; height: 16px; border-width: 2px" />
          {{ loading ? 'Logging in...' : 'Login' }}
        </button>
      </form>

      <p class="register-link">
        Don't have an account?
        <router-link to="/register">Register your team</router-link>
      </p>
    </div>
  </div>
</template>

<style scoped>
.login-page {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: calc(100vh - 56px);
  padding: 24px;
}

.login-card {
  width: 100%;
  max-width: 400px;
}

.login-card h1 {
  font-size: 24px;
  margin-bottom: 4px;
}

.subtitle {
  color: var(--color-text-muted);
  font-size: 14px;
  margin-bottom: 24px;
}

.register-link {
  margin-top: 16px;
  font-size: 14px;
  color: var(--color-text-muted);
  text-align: center;
}
</style>
