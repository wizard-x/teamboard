import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { createRouter, createMemoryHistory, type Router } from 'vue-router'
import AppHeader from '@/components/AppHeader.vue'
import { useAuth } from '@/composables/useAuth'

function createTestRouter(): Router {
  return createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/', component: { template: '<div/>' } },
      { path: '/login', component: { template: '<div/>' } },
      { path: '/register', component: { template: '<div/>' } },
      { path: '/boards', component: { template: '<div/>' } },
      { path: '/members', component: { template: '<div/>' } },
    ],
  })
}

describe('AppHeader', () => {
  let router: Router

  beforeEach(() => {
    localStorage.clear()
    const { clearAuth } = useAuth()
    clearAuth()
    router = createTestRouter()
    router.push('/')
  })

  it('shows Login and Register buttons when not authenticated', () => {
    const wrapper = mount(AppHeader, {
      global: { plugins: [router] },
    })
    expect(wrapper.text()).toContain('Login')
    expect(wrapper.text()).toContain('Register')
  })

  it('renders the logo', () => {
    const wrapper = mount(AppHeader, {
      global: { plugins: [router] },
    })
    expect(wrapper.text()).toContain('Team Board')
  })

  it('shows user info and nav when authenticated', async () => {
    const { setAuth } = useAuth()
    setAuth('test-key', {
      id: 'mem-1',
      name: 'Test User',
      email: 'test@example.com',
      role: 'admin',
      created_at: '2026-01-01T00:00:00Z',
    })

    const wrapper = mount(AppHeader, {
      global: { plugins: [router] },
    })

    expect(wrapper.text()).toContain('Test User')
    expect(wrapper.text()).toContain('Logout')
    expect(wrapper.text()).toContain('Boards')
    expect(wrapper.text()).toContain('Members')
  })
})
