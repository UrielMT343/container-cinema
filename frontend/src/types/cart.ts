export interface CheckoutBeginResponse {
  cartId: string;
  expiresAt: string;
}

export interface ConfirmRequestPayload {
  email: string;
  ticketIds: string[];
}

export interface ConfirmedTicket {
  id: string;
  idSeat: number;
  idShowtime: number;
  expiresAt: string;
  status: string;
}

export interface HoldRequestPayload {
  idSeats: number[];
  idShowtime: number;
  email: string;
  idUser: number;
}

export interface HeldTicket {
  id: string;
  idSeat: number;
  idShowtime: number;
  idUser: number;
  email: string;
  status: string;
}

export type CheckoutStep = 'SEATS' | 'PAYMENT';

export interface CheckoutSession {
  expiresAt: string;
  step: CheckoutStep;
  showtimeId: number;
  ticketIds?: string[];
}
