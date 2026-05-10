export type VacancyStatus = "published" | "assigned" | "completed" | "cancelled" | "unspecified";
export type ResponseStatus = "pending" | "accepted" | "rejected" | "withdrawn" | "unspecified";
export type AssignmentStatus = "active" | "completed" | "cancelled" | "unspecified";
export type ReviewTargetRole = "owner" | "worker" | "unspecified";

export type Vacancy = {
  id: number;
  owner_id: number;
  pvz_id: number;
  status: VacancyStatus;
  title?: string | null;
  description?: string | null;
  shift_start_at: string;
  shift_end_at: string;
  pvz_name_snapshot: string;
  pvz_address_snapshot: string;
  pay_amount: number;
  slots_total: number;
  created_at: string;
  updated_at: string;
};

export type VacancyListResponse = {
  items: Vacancy[];
  next_page_token?: string;
};

export type Response = {
  id: number;
  vacancy_id: number;
  worker_id: number;
  status: ResponseStatus;
  message?: string | null;
  created_at: string;
  updated_at: string;
};

export type ResponseListResponse = {
  items: Response[];
  next_page_token?: string;
};

export type Assignment = {
  id: number;
  vacancy_id: number;
  response_id: number;
  owner_id: number;
  worker_id: number;
  status: AssignmentStatus;
  accepted_at: string;
  pvz_name_snapshot: string;
  pvz_address_snapshot: string;
  shift_start_at_snapshot: string;
  shift_end_at_snapshot: string;
  pay_amount_snapshot: number;
  owner_name_snapshot: string;
  worker_name_snapshot: string;
  completed_at?: string | null;
  cancelled_at?: string | null;
  cancelled_by_user_id?: number | null;
  cancel_reason?: string | null;
  created_at: string;
  updated_at: string;
};

export type AssignmentListResponse = {
  items: Assignment[];
  next_page_token?: string;
};

export type Review = {
  id: number;
  assignment_id: number;
  author_user_id: number;
  target_user_id: number;
  target_role: ReviewTargetRole;
  rating: number;
  body?: string | null;
  created_at: string;
  updated_at: string;
};

export type ReviewListResponse = {
  items: Review[];
  next_page_token?: string;
};

export type ReviewReply = {
  id: number;
  review_id: number;
  author_user_id: number;
  target_user_id: number;
  body: string;
  created_at: string;
  updated_at: string;
};

export type ReviewReplyListResponse = {
  items: ReviewReply[];
  next_page_token?: string;
};

export type UserRating = {
  user_id: number;
  avg_rating: number;
  reviews_count: number;
  updated_at?: string | null;
};

export type VacancyFilters = {
  status?: VacancyStatus;
  pvz_id?: number;
  shift_start_from?: string;
  shift_start_to?: string;
  page_token?: string;
};

export type CreateVacancyInput = {
  pvz_id: number;
  title?: string;
  description?: string;
  shift_start_at: string;
  shift_end_at: string;
  pay_amount: number;
  slots_total: number;
};
