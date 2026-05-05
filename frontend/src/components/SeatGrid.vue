<script setup lang="ts">
import { computed, type PropType } from 'vue';
import type { Seat } from '@/types/seats';

const props = defineProps({
  seats: {
    type: Array as PropType<Seat[]>,
    required: true,
  },
});

interface RowSeatData {
  row: string;
  seatMap: (Seat & { seatNum: number } | null)[];
}

const rowSeats = computed<RowSeatData[]>(() => {
  const rows: Record<string, { seatMap: (Seat & { seatNum: number } | null)[] }> = {};

  props.seats.forEach(seat => {
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
</script>

<template>
  <div class="overflow-x-auto pb-4">
    <!-- Main Grid: 1st col for labels, next 15 for seats -->
    <div class="min-w-[800px] grid grid-cols-[auto_repeat(15,minmax(0,1fr))] gap-2 text-center mx-auto">
      
      <!-- Header Row -->
      <div class="font-bold text-gray-500 py-2"></div> <!-- Spacer -->
      <div v-for="col in 15" :key="`col-${col}`" class="text-xs font-semibold text-gray-400 py-2">
        {{ col }}
      </div>

      <!-- Seat Rows -->
      <template v-for="rowData of rowSeats" :key="rowData.row">
        <!-- Row Label -->
        <div class="font-bold text-gray-700 py-2">{{ rowData.row }}</div>
        
        <!-- Seats for this Row -->
        <div
          v-for="(seat, index) of rowData.seatMap"
          :key="index"
          :class="{
            'bg-green-100 text-green-700 border-green-200': seat?.status === 'AVAILABLE',
            'bg-yellow-100 text-yellow-700 border-yellow-200': seat?.status === 'HOLD',
            'bg-red-100 text-red-700 border-red-200 opacity-60': seat?.status === 'SOLD',
            'bg-transparent border-transparent text-gray-200': !seat,
            'border rounded-md p-2 flex items-center justify-center transition-colors': true,
          }"
          :title="seat ? `Seat ${seat.number} - ${seat.status}` : ''"
        >
          <svg v-if="seat" xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
             <path d="M3 4a1 1 0 011-1h12a1 1 0 011 1v2a1 1 0 01-1 1H4a1 1 0 01-1-1V4zm0 4a1 1 0 011-1h12a1 1 0 011 1v6a1 1 0 01-1 1H4a1 1 0 01-1-1v-6zm2 0v6h10v-6H5z" />
          </svg>
          <span v-else class="text-lg">·</span>
        </div>
      </template>
    </div>

    <!-- Legend -->
    <div class="mt-8 flex gap-6 justify-center">
      <div class="flex items-center gap-2">
        <div class="w-4 h-4 rounded border border-green-200 bg-green-100"></div>
        <span class="text-sm text-gray-600 font-medium">Available</span>
      </div>
      <div class="flex items-center gap-2">
        <div class="w-4 h-4 rounded border border-yellow-200 bg-yellow-100"></div>
        <span class="text-sm text-gray-600 font-medium">Hold</span>
      </div>
      <div class="flex items-center gap-2">
        <div class="w-4 h-4 rounded border border-red-200 bg-red-100 opacity-60"></div>
        <span class="text-sm text-gray-600 font-medium">Sold</span>
      </div>
    </div>
  </div>
</template>
