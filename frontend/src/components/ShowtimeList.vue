<script setup lang="ts">
import { useQuery } from '@tanstack/vue-query';
import { getShowtimes } from '@/api/showtimes';
import type { Showtime } from '@/types/showtimes';

const { data: showtimes, isLoading, error } = useQuery<Showtime[], Error>({
  queryKey: ['public-showtimes'],
  queryFn: getShowtimes,
});
</script>

<template>
  <div class="container mx-auto px-4 py-8">
    <h1 class="text-3xl font-bold mb-8 text-gray-900">Now Showing</h1>
    <div v-if="isLoading" class="text-center py-12">
      <p class="text-gray-500 text-lg">Loading showtimes...</p>
    </div>
    <div v-else-if="error" class="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg">
      Error loading showtimes: {{ error.message }}
    </div>
    <div v-else-if="showtimes?.length" class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
      <div v-for="showtime in showtimes" :key="showtime.id">
        <router-link
          :to="`/showtimes/${showtime.id}/seats`"
          class="block h-full group"
        >
          <div class="h-full border border-gray-100 rounded-xl p-6 shadow-sm bg-white cursor-pointer group-hover:shadow-xl group-hover:-translate-y-1 transition-all duration-300 flex flex-col justify-between">
            <div>
              <h2 class="text-2xl font-bold mb-4 text-gray-900 group-hover:text-blue-600 transition-colors">
                {{ showtime.movie.name }}
              </h2>
              <div class="space-y-3 mb-4">
                <p class="text-gray-600 flex items-start gap-2">
                  <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-gray-400 mt-0.5 shrink-0" viewBox="0 0 20 20" fill="currentColor">
                    <path fill-rule="evenodd" d="M4 4a2 2 0 012-2h8a2 2 0 012 2v12a2 2 0 01-2 2H6a2 2 0 01-2-2V4zm2 0h8v12H6V4z" clip-rule="evenodd" />
                  </svg>
                  <span>Auditorium: <span class="font-medium text-gray-800">{{ showtime.auditorium.name }}</span></span>
                </p>
                <p class="text-gray-600 flex items-start gap-2">
                  <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-gray-400 mt-0.5 shrink-0" viewBox="0 0 20 20" fill="currentColor">
                    <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm1-12a1 1 0 10-2 0v4a1 1 0 00.293.707l2.828 2.829a1 1 0 101.415-1.415L11 9.586V6z" clip-rule="evenodd" />
                  </svg>
                  <span class="font-medium text-gray-800">{{ new Date(showtime.startTime).toLocaleString() }}</span>
                </p>
              </div>
            </div>
            <div class="flex justify-between items-center pt-4 border-t border-gray-50 mt-auto">
              <p class="text-gray-400 text-xs">ID: {{ showtime.id }}</p>
              <span class="text-blue-600 group-hover:text-blue-700 text-sm font-semibold transition-colors">
                Select Seats &rarr;
              </span>
            </div>
          </div>
        </router-link>
      </div>
    </div>
    <div v-else class="text-center py-12 text-gray-500 text-lg">
      No showtimes available.
    </div>
  </div>
</template>
