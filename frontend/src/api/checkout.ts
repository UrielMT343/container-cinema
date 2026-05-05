import apiClient from "@/lib/axios";
import type {
  HoldRequestPayload,
  HeldTicket,
  ConfirmRequestPayload,
  ConfirmedTicket,
} from "@/types/cart";
import type { ErrorResponse } from "@/types/showtimes";

export const beginCheckout = async (): Promise<void> => {
    await apiClient.post("/public/checkout/begin");
};

export const holdTickets = async (
    payload: HoldRequestPayload,
): Promise<HeldTicket[]> => {
    const response = await apiClient.post<HeldTicket[] | ErrorResponse>(
        "/user/ticket/hold",
        payload,
    );
    if (response.data && "error" in response.data) {
        throw new Error(response.data.error);
    }
    return response.data as HeldTicket[];
};

export const confirmTickets = async (
    payload: ConfirmRequestPayload,
): Promise<ConfirmedTicket[]> => {
    const response = await apiClient.patch<ConfirmedTicket[] | ErrorResponse>(
        "/user/ticket/pay",
        payload,
    );
    if (response.data && "error" in response.data) {
        throw new Error(response.data.error);
    }
    return response.data as ConfirmedTicket[];
};
