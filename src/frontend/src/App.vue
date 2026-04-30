<script setup lang="ts">
import { RouterView } from 'vue-router'
import AppHeader from '@/components/AppHeader.vue'
import NotificationList from '@/components/NotificationList.vue'
</script>

<template>
  <div class="app">
    <AppHeader />
    <main>
      <RouterView />
    </main>
    <NotificationList />
  </div>
</template>

<style scoped>
.app {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}
main {
  flex: 1;
}
</style>
