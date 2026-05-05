import apiClient from "@/lib/axios";
import type { Showtime, ErrorResponse } from "@/types/showtimes";

export const getShowtimes = async (): Promise<Showtime[]> => {
    const response = await apiClient.get<Showtime[] | ErrorResponse>(
        "/public/showtimes",
    );
    if (response.data && "error" in response.data) {
        throw new Error(response.data.error);
    }
    return response.data as Showtime[];
};
