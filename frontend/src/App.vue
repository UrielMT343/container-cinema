<script setup lang="ts">
import SessionExpiredModal from '@/components/SessionExpiredModal.vue';
</script>

<template>
  <RouterView />
  <SessionExpiredModal />
</template>
