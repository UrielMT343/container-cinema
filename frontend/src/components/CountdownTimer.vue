<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue';
import { useRouter } from 'vue-router';
import type { CheckoutSession } from '@/types/cart';
import { isSessionExpired } from '@/composables/useSessionExpired';

const router = useRouter();
const display = ref('');
const expired = ref(false);
let intervalId: ReturnType<typeof setInterval> | null = null;

function pad(n: number): string {
  return String(n).padStart(2, '0');
}

function tick(): void {
  const raw = sessionStorage.getItem('checkout-session');
  if (!raw) {
    display.value = '';
    return;
  }

  let session: CheckoutSession;
  try {
    session = JSON.parse(raw);
  } catch {
    display.value = '';
    return;
  }

  const expiresAt = new Date(session.expiresAt).getTime();
  const now = Date.now();
  const diff = Math.max(0, expiresAt - now);

  if (diff <= 0) {
    display.value = '00:00';
    expired.value = true;
    if (intervalId !== null) {
      clearInterval(intervalId);
      intervalId = null;
    }
    isSessionExpired.value = true;
    return;
  }

  const minutes = Math.floor(diff / 60000);
  const seconds = Math.floor((diff % 60000) / 1000);
  display.value = `${pad(minutes)}:${pad(seconds)}`;
}

onMounted(() => {
  const raw = sessionStorage.getItem('checkout-session');
  if (raw) {
    let session: CheckoutSession;
    try {
      session = JSON.parse(raw);
    } catch {
      return;
    }
    if (session.expiresAt) {
      tick();
      intervalId = setInterval(tick, 1000);
    }
  }
});

onUnmounted(() => {
  if (intervalId !== null) {
    clearInterval(intervalId);
    intervalId = null;
  }
});
</script>

<template>
  <div
    v-if="display"
    class="flex items-center justify-center gap-2 mb-6"
    :class="{
      'text-amber-700 font-semibold': true,
    }"
  >
    <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
      <path stroke-linecap="round" stroke-linejoin="round" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
    </svg>
    <span
      class="text-sm font-mono font-bold px-3 py-1.5 rounded-lg border"
      :class="{
        'bg-red-50 text-red-700 border-red-200': display <= '01:00',
        'bg-amber-50 text-amber-700 border-amber-200': display > '01:00' && display <= '05:00',
        'bg-gray-50 text-gray-700 border-gray-200': display > '05:00',
      }"
    >
      {{ display }}
    </span>
  </div>
</template>
