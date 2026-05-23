<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch } from 'vue';
import { useRoute } from 'vue-router';
import { useQuery, useMutation } from '@tanstack/vue-query';
import { getSeatsByShowtime } from '@/api/seats';
import { beginCheckout } from '@/api/checkout';
import type { Seat } from '@/types/seats';
import type { CheckoutBeginResponse, CheckoutSession } from '@/types/cart';
import SeatGrid from '@/components/SeatGrid.vue';
import SeatMap from '@/components/SeatMap.vue';
import CountdownTimer from '@/components/CountdownTimer.vue';

const route = useRoute();
const showtimeId = Number(route.params.id);
const showModal = ref(false);
const cartInitialized = ref(false);
const checkoutError = ref<string | null>(null);

onMounted(() => {
  const raw = sessionStorage.getItem('checkout-session');
  if (raw) {
    try {
      const session: CheckoutSession = JSON.parse(raw);
      if (session.step === 'SEATS' && session.showtimeId === showtimeId) {
        cartInitialized.value = true;
      }
    } catch {
      // ignore
    }
  }
});

const { data: seats, isLoading, error } = useQuery<Seat[], Error>({
  queryKey: ['showtime-seats', showtimeId],
  queryFn: () => getSeatsByShowtime(showtimeId),
  enabled: !isNaN(showtimeId),
});

const localSeats = ref<Seat[]>([]);
let eventSource: EventSource | null = null;

watch(() => seats.value, (newSeats) => {
  if (newSeats) localSeats.value = newSeats.map((s) => ({ ...s }));
}, { immediate: true });

watch(() => showtimeId, (id) => {
  if (!id || isNaN(id)) return;

  eventSource?.close();
  eventSource = new EventSource(`/api/v1/public/sse/showtimes/${id}`, {
    withCredentials: true,
  });

  eventSource.addEventListener('seat_update', (event) => {
    try {
      const parsed = JSON.parse(event.data);
      const payloads = Array.isArray(parsed) ? parsed : [parsed];
      payloads.forEach((p: { seatId: number; status: Seat['status'] }) => {
        const seat = localSeats.value.find((s) => s.id === p.seatId);
        if (seat) seat.status = p.status;
      });
    } catch {
      // ignore parse errors
    }
  });
}, { immediate: true });

onUnmounted(() => {
  eventSource?.close();
});

const beginMutation = useMutation<CheckoutBeginResponse, Error, void>({
  mutationFn: beginCheckout,
  onSuccess: (data) => {
    const session: CheckoutSession = {
      expiresAt: data.expiresAt,
      step: 'SEATS',
      showtimeId,
    };
    sessionStorage.setItem('checkout-session', JSON.stringify(session));
    cartInitialized.value = true;
    showModal.value = false;
    checkoutError.value = null;
  },
  onError: (err: unknown) => {
    const axiosErr = err as { response?: { status: number; data?: { error?: string } } };
    checkoutError.value =
      axiosErr.response?.data?.error ?? 'Failed to initialize checkout. Please try again.';
  },
});

const handleConfirm = () => {
  checkoutError.value = null;
  beginMutation.mutate();
};

const handleCancel = () => {
  showModal.value = false;
  checkoutError.value = null;
};
</script>

<template>
  <div class="container mx-auto px-4 py-8">
    <button
      @click="$router.back()"
      class="mb-6 inline-flex items-center gap-2 text-blue-600 hover:text-blue-800 transition-colors"
    >
      <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
        <path fill-rule="evenodd" d="M9.707 16.707a1 1 0 01-1.414 0l-6-6a1 1 0 010-1.414l6-6a1 1 0 011.414 1.414L5.414 9H16a1 1 0 110 2H5.414l4.293 4.293a1 1 0 010 1.414z" clip-rule="evenodd" />
      </svg>
      Back to Showtimes
    </button>

    <h1 class="text-3xl font-bold mb-6 text-gray-900">Seats for Showtime #{{ showtimeId }}</h1>

    <CountdownTimer />

    <div v-if="isLoading" class="text-center py-12">
      <p class="text-gray-500 text-lg">Loading seats...</p>
    </div>

    <div v-else-if="error" class="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg">
      Error loading seats: {{ error.message }}
    </div>

    <div v-else-if="seats?.length">
      <!-- Display-only grid before cart init -->
      <SeatGrid v-if="!cartInitialized" :seats="localSeats" />

      <!-- Interactive seat map after cart init -->
      <SeatMap v-else :seats="localSeats" :showtimeId="showtimeId" />

      <!-- Select Seats Button (only before cart init) -->
      <div v-if="!cartInitialized" class="mt-8 flex justify-center">
        <button
          @click="showModal = true"
          class="px-8 py-3 bg-blue-600 hover:bg-blue-700 text-white font-semibold rounded-lg shadow-md hover:shadow-lg transition-all duration-200"
        >
          Select Seats
        </button>
      </div>
    </div>

    <div v-else class="text-center py-12 text-gray-500 text-lg">
      No seats found for this showtime.
    </div>

    <!-- Confirmation Modal -->
    <Transition name="fade">
      <div
        v-if="showModal"
        class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 px-4"
        @click.self="handleCancel"
      >
        <div class="w-full max-w-md bg-white rounded-xl shadow-2xl p-6">
          <h2 class="text-xl font-bold text-gray-900 mb-2">Confirm Selection</h2>
          <p class="text-gray-600 mb-4">
            Are you sure you want to proceed with seat selection for this showtime?
          </p>

          <div v-if="checkoutError" class="mb-4 bg-red-50 border border-red-200 text-red-700 px-3 py-2 rounded-lg text-sm">
            {{ checkoutError }}
          </div>

          <div class="flex gap-3 justify-end">
            <button
              @click="handleCancel"
              :disabled="beginMutation.isPending.value"
              class="px-4 py-2 text-gray-700 bg-gray-100 hover:bg-gray-200 rounded-lg transition-colors disabled:opacity-50"
            >
              Cancel
            </button>
            <button
              @click="handleConfirm"
              :disabled="beginMutation.isPending.value"
              class="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white font-medium rounded-lg transition-colors disabled:opacity-50"
            >
              <span v-if="beginMutation.isPending.value">Initializing...</span>
              <span v-else>Confirm</span>
            </button>
          </div>
        </div>
      </div>
    </Transition>
  </div>
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
