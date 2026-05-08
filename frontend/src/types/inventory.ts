export type PvzStatus = "pending" | "approved" | "rejected" | "blocked" | "unspecified";

export type Pvz = {
  id: number;
  owner_id: number;
  status: PvzStatus;
  city: string;
  address: string;
  company_name: string;
  contact_phone: string;
  contact_telegram?: string | null;
  description?: string | null;
  lat?: number | null;
  lng?: number | null;
  rejection_reason?: string | null;
  moderated_by?: number | null;
  moderated_at?: string | null;
  created_at: string;
  updated_at: string;
};

export type PvzFormInput = {
  city: string;
  address: string;
  company_name: string;
  contact_phone: string;
  contact_telegram?: string;
  description?: string;
  lat?: number;
  lng?: number;
};

export type PvzListResponse = {
  items: Pvz[];
  next_page_token?: string;
};

export type ModerationDecisionInput = {
  reason?: string;
  comment?: string;
};

export type BlockUserAndDeletePvzInput = {
  reason: string;
  block_duration_minutes: number;
  comment?: string;
};

export type BlockUserAndDeletePvzResponse = {
  blocked_user_id: string;
  blocked_until: string;
  deleted_count: string;
  blocked_count: string;
};
