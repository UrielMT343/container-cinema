<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import { useMutation } from '@tanstack/vue-query';
import { confirmTickets } from '@/api/checkout';
import type { ConfirmedTicket } from '@/types/cart';

const router = useRouter();
const email = ref('');
const payError = ref<string | null>(null);
const confirmed = ref(false);
const ticketCount = ref(0);

onMounted(() => {
  const raw = sessionStorage.getItem('checkout-session');
  if (!raw) {
    router.push('/');
    return;
  }
  try {
    const session = JSON.parse(raw);
    ticketCount.value = session.ticketIds.length;
  } catch {
    router.push('/');
  }
});

const payMutation = useMutation<ConfirmedTicket[], Error, void>({
  mutationFn: () => {
    const raw = sessionStorage.getItem('checkout-session');
    if (!raw) throw new Error('No checkout session found');
    const session = JSON.parse(raw);
    return confirmTickets({
      email: email.value.trim(),
      ticketIds: session.ticketIds as string[],
    });
  },
  onSuccess: () => {
    sessionStorage.removeItem('checkout-session');
    confirmed.value = true;
    payError.value = null;
  },
  onError: (err: unknown) => {
    const axiosErr = err as { response?: { status: number; data?: { error?: string } } };
    const message = axiosErr.response?.data?.error;

    if (axiosErr.response?.status === 400) {
      payError.value = message ?? 'Invalid request. Please check your email and try again.';
    } else {
      payError.value = message ?? 'Payment failed. Please try again.';
    }
  },
});

function handlePay() {
  if (!email.value.trim()) return;
  payError.value = null;
  payMutation.mutate();
}

function goHome() {
  router.push('/');
}
</script>

<template>
  <div class="min-h-screen bg-gray-50 py-12 px-4">
    <div class="max-w-lg mx-auto">
      <!-- Success State -->
      <div v-if="confirmed" class="bg-white rounded-xl shadow-lg p-8 text-center">
        <div class="mx-auto w-16 h-16 bg-green-100 rounded-full flex items-center justify-center mb-6">
          <svg xmlns="http://www.w3.org/2000/svg" class="h-8 w-8 text-green-600" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" />
          </svg>
        </div>
        <h1 class="text-2xl font-bold text-gray-900 mb-2">Tickets Confirmed!</h1>
        <p class="text-gray-600 mb-8">
          Your tickets have been sent to your email. Enjoy the show!
        </p>
        <button
          class="px-6 py-3 bg-blue-600 hover:bg-blue-700 text-white font-semibold rounded-lg shadow-md hover:shadow-lg transition-all duration-200"
          @click="goHome"
        >
          Back to Home
        </button>
      </div>

      <!-- Checkout Form -->
      <div v-else class="bg-white rounded-xl shadow-lg p-8">
        <button
          class="mb-6 inline-flex items-center gap-2 text-blue-600 hover:text-blue-800 transition-colors"
          @click="$router.back()"
        >
          <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
            <path fill-rule="evenodd" d="M9.707 16.707a1 1 0 01-1.414 0l-6-6a1 1 0 010-1.414l6-6a1 1 0 011.414 1.414L5.414 9H16a1 1 0 110 2H5.414l4.293 4.293a1 1 0 010 1.414z" clip-rule="evenodd" />
          </svg>
          Back
        </button>

        <h1 class="text-2xl font-bold text-gray-900 mb-6">Checkout</h1>

        <!-- Order Summary -->
        <div class="bg-gray-50 border border-gray-200 rounded-lg p-4 mb-6">
          <h2 class="text-sm font-semibold text-gray-700 uppercase tracking-wide mb-2">
            Order Summary
          </h2>
          <div class="flex justify-between items-center">
            <span class="text-gray-600">Seats held</span>
            <span class="text-lg font-bold text-gray-900">{{ ticketCount }}</span>
          </div>
          <p class="text-xs text-gray-500 mt-2">
            Seats are reserved for 15 minutes.
          </p>
        </div>

        <!-- Email Form -->
        <form @submit.prevent="handlePay">
          <label for="email" class="block text-sm font-medium text-gray-700 mb-2">
            Email Address
          </label>
          <input
            id="email"
            v-model="email"
            type="email"
            required
            placeholder="you@example.com"
            class="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 outline-none transition-colors"
          />
          <p class="mt-2 text-xs text-gray-500">
            Your tickets will be sent to this email address.
          </p>

          <!-- Error -->
          <div v-if="payError" class="mt-4 bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg text-sm">
            {{ payError }}
          </div>

          <!-- Submit -->
          <button
            type="submit"
            :disabled="!email.trim() || payMutation.isPending.value"
            class="mt-6 w-full px-6 py-3 bg-blue-600 hover:bg-blue-700 text-white font-semibold rounded-lg shadow-md hover:shadow-lg transition-all duration-200 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            <span v-if="payMutation.isPending.value">Processing...</span>
            <span v-else>Complete Purchase</span>
          </button>
        </form>
      </div>
    </div>
  </div>
</template>
