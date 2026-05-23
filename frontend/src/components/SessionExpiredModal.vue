<script setup lang="ts">
import { useRouter } from 'vue-router';
import { isSessionExpired } from '@/composables/useSessionExpired';

const router = useRouter();

function handleReturnHome() {
  sessionStorage.clear();
  isSessionExpired.value = false;
  router.push('/');
}
</script>

<template>
  <Transition name="fade">
    <div
      v-if="isSessionExpired"
      class="fixed inset-0 z-[100] flex items-center justify-center bg-black/80"
    >
      <div class="w-full max-w-sm bg-white rounded-xl shadow-2xl p-8 text-center">
        <div class="mx-auto w-16 h-16 bg-red-100 rounded-full flex items-center justify-center mb-6">
          <svg xmlns="http://www.w3.org/2000/svg" class="h-8 w-8 text-red-600" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
        </div>
        <h2 class="text-2xl font-bold text-gray-900 mb-2">Session Expired</h2>
        <p class="text-gray-600 mb-8">
          Your reserved seats have been released. Please start over to select new seats.
        </p>
        <button
          @click="handleReturnHome"
          class="w-full px-6 py-3 bg-blue-600 hover:bg-blue-700 text-white font-semibold rounded-lg shadow-md hover:shadow-lg transition-all duration-200"
        >
          Return to Home
        </button>
      </div>
    </div>
  </Transition>
</template>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
