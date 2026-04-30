<script setup lang="ts">
import { useNotifications } from '@/composables/useNotifications'

const { notifications, dismiss } = useNotifications()
</script>

<template>
  <div class="notifications">
    <div
      v-for="n in notifications"
      :key="n.id"
      :class="['notification', `notification-${n.type}`]"
    >
      <span>{{ n.message }}</span>
      <button class="dismiss" @click="dismiss(n.id)">✕</button>
    </div>
  </div>
</template>
