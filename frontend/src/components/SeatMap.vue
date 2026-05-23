<script setup lang="ts">
import { ref, computed, type PropType } from 'vue';
import { useRouter } from 'vue-router';
import { useMutation } from '@tanstack/vue-query';
import { holdTickets } from '@/api/checkout';
import type { Seat } from '@/types/seats';
import type { HeldTicket, CheckoutSession } from '@/types/cart';

const props = defineProps({
  seats: {
    type: Array as PropType<Seat[]>,
    required: true,
  },
  showtimeId: {
    type: Number,
    required: true,
  },
});

const router = useRouter();

const MAX_SELECTION = 5;
const selectedSeats = ref<number[]>([]);
const selectionWarning = ref(false);
const holdError = ref<string | null>(null);

interface RowSeatData {
  row: string;
  seatMap: (Seat & { seatNum: number } | null)[];
}

const rowSeats = computed<RowSeatData[]>(() => {
  const rows: Record<string, { seatMap: (Seat & { seatNum: number } | null)[] }> = {};

  props.seats.forEach((seat) => {
    const match = seat.number.match(/^([A-Z])(\d+)$/);
    if (!match) return;

    const row = match[1];
    const seatNum = parseInt(match[2], 10);
    if (isNaN(seatNum) || seatNum < 1 || seatNum > 15) return;

    if (!rows[row]) {
      rows[row] = { seatMap: new Array(15).fill(null) };
    }

    rows[row].seatMap[seatNum - 1] = { ...seat, seatNum };
  });

  return Object.entries(rows)
    .sort((a, b) => a[0].localeCompare(b[0]))
    .map(([row, data]) => ({ row, seatMap: data.seatMap }));
});

const selectedSeatObjects = computed(() =>
  props.seats.filter((s) => selectedSeats.value.includes(s.id)),
);

function toggleSeat(seat: Seat) {
    holdError.value = null;
    selectionWarning.value = false;

    if (seat.status !== "AVAILABLE") return;

    const idx = selectedSeats.value.indexOf(seat.id);
    if (idx !== -1) {
        selectedSeats.value.splice(idx, 1);
        return;
    }

    if (selectedSeats.value.length >= MAX_SELECTION) {
        selectionWarning.value = true;
        setTimeout(() => {
            selectionWarning.value = false;
        }, 3000);
        return;
    }

    selectedSeats.value.push(seat.id);
}

function isSelected(seat: Seat | null): boolean {
    return seat !== null && selectedSeats.value.includes(seat.id);
}

const holdMutation = useMutation<HeldTicket[], Error, void>({
    mutationFn: () =>
        holdTickets({
            idSeats: [...selectedSeats.value],
            idShowtime: props.showtimeId,
            email: "",
            idUser: 0,
        }),
    onSuccess: (heldTickets) => {
        holdError.value = null;
        const raw = sessionStorage.getItem("checkout-session");
        if (raw) {
            const session: CheckoutSession = JSON.parse(raw);
            session.step = "PAYMENT";
            session.ticketIds = heldTickets.map((t) => t.id);
            sessionStorage.setItem("checkout-session", JSON.stringify(session));
        }
        router.push("/checkout/pay");
    },
    onError: (err: unknown) => {
        const axiosErr = err as {
            response?: { status: number; data?: { error?: string } };
        };
        const status = axiosErr.response?.status;
        const message = axiosErr.response?.data?.error;

        if (status === 409) {
            holdError.value =
                message ??
                "One or more selected seats were just taken. Please try again.";
        } else if (status === 400) {
            holdError.value =
                message ??
                "Invalid selection. Please check your seats and try again.";
        } else {
            holdError.value =
                message ?? "Failed to hold seats. Please try again.";
        }
    },
});

function handleConfirmSeats() {
    if (selectedSeats.value.length === 0) return;
    holdError.value = null;
    holdMutation.mutate();
}
</script>

<template>
    <div>
        <!-- Selection Warning -->
        <Transition name="slide-fade">
            <div
                v-if="selectionWarning"
                class="mb-4 bg-amber-50 border border-amber-200 text-amber-800 px-4 py-3 rounded-lg text-sm font-medium"
            >
                You can select a maximum of {{ MAX_SELECTION }} seats. Deselect
                a seat before choosing another.
            </div>
        </Transition>

        <!-- Hold Error -->
        <Transition name="slide-fade">
            <div
                v-if="holdError"
                class="mb-4 bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg text-sm font-medium"
            >
                {{ holdError }}
            </div>
        </Transition>

        <!-- Selected Seats Summary -->
        <div
            v-if="selectedSeats.length > 0"
            class="mb-6 bg-blue-50 border border-blue-200 rounded-lg px-4 py-3"
        >
            <p class="text-sm font-semibold text-blue-900 mb-1">
                Selected ({{ selectedSeats.length }}/{{ MAX_SELECTION }}):
            </p>
            <p class="text-sm text-blue-700">
                {{ selectedSeatObjects.map((s) => s.number).join(", ") }}
            </p>
        </div>

        <!-- Seat Grid -->
        <div class="overflow-x-auto pb-4">
            <div
                class="min-w-[800px] grid grid-cols-[auto_repeat(15,minmax(0,1fr))] gap-2 text-center mx-auto"
            >
                <!-- Header Row -->
                <div class="font-bold text-gray-500 py-2"></div>
                <div
                    v-for="col in 15"
                    :key="`col-${col}`"
                    class="text-xs font-semibold text-gray-400 py-2"
                >
                    {{ col }}
                </div>

                <!-- Seat Rows -->
                <template v-for="rowData of rowSeats" :key="rowData.row">
                    <div class="font-bold text-gray-700 py-2">
                        {{ rowData.row }}
                    </div>
                    <button
                        v-for="(seat, index) of rowData.seatMap"
                        :key="index"
                        :disabled="
                            seat?.status === 'SOLD' || seat?.status === 'HOLD'
                        "
                        :class="{
                            'bg-green-100 text-green-700 border-green-200 hover:bg-green-200':
                                seat?.status === 'AVAILABLE' &&
                                !isSelected(seat),
                            'bg-blue-600 text-white border-blue-700 shadow-md':
                                isSelected(seat),
                            'bg-yellow-100 text-yellow-700 border-yellow-200':
                                seat?.status === 'HOLD',
                            'bg-red-100 text-red-700 border-red-200 opacity-60 cursor-not-allowed':
                                seat?.status === 'SOLD',
                            'bg-transparent border-transparent text-gray-200 cursor-default':
                                !seat,
                            'border rounded-md p-2 flex items-center justify-center transition-all duration-150 cursor-pointer': true,
                        }"
                        :title="
                            seat
                                ? `Seat ${seat.number} - ${seat.status}${isSelected(seat) ? ' (selected)' : ''}`
                                : ''
                        "
                        @click="seat ? toggleSeat(seat) : null"
                    >
                        <svg
                            v-if="seat"
                            xmlns="http://www.w3.org/2000/svg"
                            class="h-5 w-5"
                            viewBox="0 0 20 20"
                            fill="currentColor"
                        >
                            <path
                                d="M3 4a1 1 0 011-1h12a1 1 0 011 1v2a1 1 0 01-1 1H4a1 1 0 01-1-1V4zm0 4a1 1 0 011-1h12a1 1 0 011 1v6a1 1 0 01-1 1H4a1 1 0 01-1-1v-6zm2 0v6h10v-6H5z"
                            />
                        </svg>
                        <span v-else class="text-lg">·</span>
                    </button>
                </template>
            </div>

            <!-- Legend -->
            <div class="mt-8 flex gap-6 justify-center flex-wrap">
                <div class="flex items-center gap-2">
                    <div
                        class="w-4 h-4 rounded border border-green-200 bg-green-100"
                    ></div>
                    <span class="text-sm text-gray-600 font-medium"
                        >Available</span
                    >
                </div>
                <div class="flex items-center gap-2">
                    <div
                        class="w-4 h-4 rounded border border-blue-700 bg-blue-600"
                    ></div>
                    <span class="text-sm text-gray-600 font-medium"
                        >Selected</span
                    >
                </div>
                <div class="flex items-center gap-2">
                    <div
                        class="w-4 h-4 rounded border border-yellow-200 bg-yellow-100"
                    ></div>
                    <span class="text-sm text-gray-600 font-medium">Hold</span>
                </div>
                <div class="flex items-center gap-2">
                    <div
                        class="w-4 h-4 rounded border border-red-200 bg-red-100 opacity-60"
                    ></div>
                    <span class="text-sm text-gray-600 font-medium">Sold</span>
                </div>
            </div>
        </div>

        <!-- Confirm Seats Button -->
        <div class="mt-8 flex flex-col items-center gap-3">
            <button
                :disabled="
                    selectedSeats.length === 0 || holdMutation.isPending.value
                "
                class="px-8 py-3 font-semibold rounded-lg shadow-md transition-all duration-200 disabled:opacity-50 disabled:cursor-not-allowed"
                :class="
                    selectedSeats.length === 0
                        ? 'bg-gray-300 text-gray-500'
                        : 'bg-blue-600 hover:bg-blue-700 text-white hover:shadow-lg'
                "
                @click="handleConfirmSeats"
            >
                <span v-if="holdMutation.isPending.value"
                    >Holding seats...</span
                >
                <span v-else>Confirm Seats ({{ selectedSeats.length }})</span>
            </button>
        </div>
    </div>
</template>

<style scoped>
.slide-fade-enter-active,
.slide-fade-leave-active {
    transition: all 0.3s ease;
}

.slide-fade-enter-from,
.slide-fade-leave-to {
    opacity: 0;
    transform: translateY(-8px);
}
</style>
