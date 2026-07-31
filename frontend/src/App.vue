<script setup lang="ts">
import { onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'
import AppToast from '@/components/common/AppToast.vue'

const auth = useAuthStore()

onMounted(() => {
  auth.checkAuth()
})
</script>

<template>
  <router-view />
  <AppToast />
</template>
