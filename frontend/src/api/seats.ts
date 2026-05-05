import apiClient from '@/lib/axios';
import type { ErrorResponse } from '@/types/showtimes';
import type { Seat } from '@/types/seats';

export const getSeatsByShowtime = async (showtimeId: number): Promise<Seat[]> => {
  const response = await apiClient.get<Seat[] | ErrorResponse>(`/public/seats/showtime/${showtimeId}`);
  if (response.data && 'error' in response.data) {
    throw new Error(response.data.error);
  }
  return response.data as Seat[];
};
