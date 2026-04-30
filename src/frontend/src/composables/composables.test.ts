import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { createRouter, createMemoryHistory } from 'vue-router'
import NotificationList from '@/components/NotificationList.vue'
import { useNotifications } from '@/composables/useNotifications'
import { useAuth } from '@/composables/useAuth'

describe('NotificationList', () => {
  it('renders no notifications initially', () => {
    const wrapper = mount(NotificationList)
    expect(wrapper.findAll('.notification')).toHaveLength(0)
  })

  it('renders notifications when added', async () => {
    const { notify, notifications } = useNotifications()
    // Clear any stale notifications
    notifications.value.splice(0)
    notify('success', 'Test success')
    notify('error', 'Test error')

    const wrapper = mount(NotificationList)
    await wrapper.vm.$nextTick()

    const notifs = wrapper.findAll('.notification')
    expect(notifs.length).toBeGreaterThanOrEqual(2)
  })
})

describe('useNotifications composable', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    const { notifications } = useNotifications()
    notifications.value.splice(0)
  })

  it('adds and removes notifications', () => {
    const { notifications, notify, dismiss } = useNotifications()
    notify('info', 'Hello')
    expect(notifications.value).toHaveLength(1)
    expect(notifications.value[0].message).toBe('Hello')
    expect(notifications.value[0].type).toBe('info')

    dismiss(notifications.value[0].id)
    expect(notifications.value).toHaveLength(0)
  })

  it('auto-removes notifications after timeout', () => {
    const { notifications, notify } = useNotifications()
    notify('success', 'Will disappear')
    expect(notifications.value).toHaveLength(1)

    vi.advanceTimersByTime(4000)
    expect(notifications.value).toHaveLength(0)
  })
})

describe('useAuth composable', () => {
  beforeEach(() => {
    localStorage.clear()
    const { clearAuth } = useAuth()
    clearAuth()
  })

  it('starts unauthenticated when no key in storage', () => {
    const { isAuthenticated } = useAuth()
    expect(isAuthenticated.value).toBe(false)
  })

  it('becomes authenticated after setAuth', () => {
    const { isAuthenticated, setAuth } = useAuth()
    setAuth('tb_k1_test', {
      id: 'm1',
      name: 'Test',
      email: 'test@test.com',
      role: 'admin',
      created_at: '2026-01-01T00:00:00Z',
    })
    expect(isAuthenticated.value).toBe(true)
  })

  it('clears auth on clearAuth', () => {
    const { isAuthenticated, setAuth, clearAuth } = useAuth()
    setAuth('tb_k1_test', {
      id: 'm1',
      name: 'Test',
      email: 'test@test.com',
      role: 'admin',
      created_at: '2026-01-01T00:00:00Z',
    })
    expect(isAuthenticated.value).toBe(true)

    clearAuth()
    expect(isAuthenticated.value).toBe(false)
  })

  it('detects admin role', () => {
    const { isAdmin, setAuth } = useAuth()
    setAuth('key', {
      id: 'm1',
      name: 'Admin',
      email: 'a@b.com',
      role: 'admin',
      created_at: '2026-01-01T00:00:00Z',
    })
    expect(isAdmin.value).toBe(true)
  })

  it('detects non-admin role', () => {
    const { isAdmin, setAuth } = useAuth()
    setAuth('key', {
      id: 'm1',
      name: 'Member',
      email: 'm@b.com',
      role: 'member',
      created_at: '2026-01-01T00:00:00Z',
    })
    expect(isAdmin.value).toBe(false)
  })

  it('persists to localStorage', () => {
    const { setAuth } = useAuth()
    setAuth('tb_k1_test', {
      id: 'm1',
      name: 'Test',
      email: 'test@test.com',
      role: 'admin',
      created_at: '2026-01-01T00:00:00Z',
    })
    expect(localStorage.setItem).toHaveBeenCalledWith('api_key', 'tb_k1_test')
    expect(localStorage.setItem).toHaveBeenCalledWith('member', expect.any(String))
  })
})
