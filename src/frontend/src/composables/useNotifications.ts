import { ref } from 'vue'

interface Notification {
  id: number
  type: 'success' | 'error' | 'info'
  message: string
}

const notifications = ref<Notification[]>([])
let nextId = 0

export function useNotifications() {
  function notify(type: Notification['type'], message: string) {
    const id = nextId++
    notifications.value.push({ id, type, message })
    setTimeout(() => {
      notifications.value = notifications.value.filter((n) => n.id !== id)
    }, 4000)
  }

  function dismiss(id: number) {
    notifications.value = notifications.value.filter((n) => n.id !== id)
  }

  return { notifications, notify, dismiss }
}
