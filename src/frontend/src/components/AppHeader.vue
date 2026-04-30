<script setup lang="ts">
import { useAuth } from '@/composables/useAuth'
import { useRouter } from 'vue-router'

const { isAuthenticated, currentMember, clearAuth } = useAuth()
const router = useRouter()

function logout() {
  clearAuth()
  router.push('/login')
}
</script>

<template>
  <header class="app-header">
    <div class="header-inner">
      <router-link to="/" class="logo">📋 Team Board</router-link>
      <nav v-if="isAuthenticated" class="nav-links">
        <router-link to="/boards">Boards</router-link>
        <router-link to="/members">Members</router-link>
      </nav>
      <div class="header-right">
        <template v-if="isAuthenticated && currentMember">
          <router-link to="/profile" class="user-link">
            <span class="user-avatar-sm">{{ currentMember.name.charAt(0).toUpperCase() }}</span>
            <span class="user-info">{{ currentMember.name }}</span>
          </router-link>
          <button class="btn btn-sm btn-secondary" @click="logout">Logout</button>
        </template>
        <template v-else>
          <router-link to="/login" class="btn btn-sm btn-secondary">Login</router-link>
          <router-link to="/register" class="btn btn-sm btn-primary">Register</router-link>
        </template>
      </div>
    </div>
  </header>
</template>

<style scoped>
.app-header {
  background: var(--color-surface);
  border-bottom: 1px solid var(--color-border);
  position: sticky;
  top: 0;
  z-index: 100;
}

.header-inner {
  max-width: 1400px;
  margin: 0 auto;
  padding: 0 24px;
  height: 56px;
  display: flex;
  align-items: center;
  gap: 24px;
}

.logo {
  font-weight: 700;
  font-size: 18px;
  color: var(--color-text);
}
.logo:hover { text-decoration: none; }

.nav-links {
  display: flex;
  gap: 4px;
}

.nav-links a {
  padding: 6px 12px;
  border-radius: var(--radius);
  font-size: 14px;
  color: var(--color-text-muted);
  transition: all 0.15s ease;
}
.nav-links a:hover,
.nav-links a.router-link-active {
  background: var(--color-bg);
  color: var(--color-text);
  text-decoration: none;
}

.header-right {
  margin-left: auto;
  display: flex;
  align-items: center;
  gap: 12px;
}

.user-info {
  font-size: 14px;
  color: var(--color-text-muted);
}

.user-link {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 4px 8px;
  border-radius: var(--radius);
  transition: background 0.15s;
}
.user-link:hover {
  background: var(--color-bg);
  text-decoration: none;
}

.user-avatar-sm {
  width: 28px;
  height: 28px;
  border-radius: 50%;
  background: var(--color-primary);
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 13px;
  font-weight: 700;
}
</style>
