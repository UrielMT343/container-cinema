export interface Seat {
  id: number;
  idAuditorium: number;
  number: string;
  status: 'AVAILABLE' | 'HOLD' | 'SOLD';
}
