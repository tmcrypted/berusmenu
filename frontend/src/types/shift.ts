export type ShiftCard = {
  id: string;
  /** ISO 8601 для атрибута dateTime */
  dateIso: string;
  dateLabel: string;
  role: string;
  address: string;
  metro: string | null;
  timeWindow: string;
  payLabel: string;
  pvzName: string;
};
