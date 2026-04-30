import { ref, computed } from 'vue'
import type { Member, MemberRole } from '@/types'

const apiKey = ref<string | null>(localStorage.getItem('api_key'))
const currentMember = ref<Member | null>(null)

// Restore member from localStorage
try {
  const stored = localStorage.getItem('member')
  if (stored) currentMember.value = JSON.parse(stored)
} catch {}

export function useAuth() {
  const isAuthenticated = computed(() => !!apiKey.value)

  const isAdmin = computed(() => currentMember.value?.role === 'admin')

  function setAuth(key: string, member: Member) {
    apiKey.value = key
    currentMember.value = member
    localStorage.setItem('api_key', key)
    localStorage.setItem('member', JSON.stringify(member))
  }

  function clearAuth() {
    apiKey.value = null
    currentMember.value = null
    localStorage.removeItem('api_key')
    localStorage.removeItem('member')
  }

  function getApiKey(): string | null {
    return apiKey.value
  }

  return {
    apiKey,
    currentMember,
    isAuthenticated,
    isAdmin,
    setAuth,
    clearAuth,
    getApiKey,
  }
}
