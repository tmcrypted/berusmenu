import { apiRequest } from "@/lib/api";
import type {
  Assignment,
  AssignmentListResponse,
  AssignmentStatus,
  CreateVacancyInput,
  Response,
  ResponseListResponse,
  ResponseStatus,
  Review,
  ReviewListResponse,
  ReviewReply,
  ReviewReplyListResponse,
  ReviewTargetRole,
  UserRating,
  Vacancy,
  VacancyFilters,
  VacancyListResponse,
  VacancyStatus,
} from "@/types/matching";

type Wire = Record<string, unknown>;

const VACANCY_STATUS: Record<string, VacancyStatus> = {
  VACANCY_STATUS_PUBLISHED: "published",
  VACANCY_STATUS_ASSIGNED: "assigned",
  VACANCY_STATUS_COMPLETED: "completed",
  VACANCY_STATUS_CANCELLED: "cancelled",
  published: "published",
  assigned: "assigned",
  completed: "completed",
  cancelled: "cancelled",
};

const RESPONSE_STATUS: Record<string, ResponseStatus> = {
  RESPONSE_STATUS_PENDING: "pending",
  RESPONSE_STATUS_ACCEPTED: "accepted",
  RESPONSE_STATUS_REJECTED: "rejected",
  RESPONSE_STATUS_WITHDRAWN: "withdrawn",
  pending: "pending",
  accepted: "accepted",
  rejected: "rejected",
  withdrawn: "withdrawn",
};

const ASSIGNMENT_STATUS: Record<string, AssignmentStatus> = {
  ASSIGNMENT_STATUS_ACTIVE: "active",
  ASSIGNMENT_STATUS_COMPLETED: "completed",
  ASSIGNMENT_STATUS_CANCELLED: "cancelled",
  active: "active",
  completed: "completed",
  cancelled: "cancelled",
};

const REVIEW_TARGET_ROLE: Record<string, ReviewTargetRole> = {
  REVIEW_TARGET_ROLE_OWNER: "owner",
  REVIEW_TARGET_ROLE_WORKER: "worker",
  owner: "owner",
  worker: "worker",
};

function num(value: unknown) {
  const parsed = Number(value);
  return Number.isFinite(parsed) ? parsed : 0;
}

function str(value: unknown) {
  return typeof value === "string" ? value : "";
}

function optStr(value: unknown) {
  return value === null || value === undefined ? null : String(value);
}

function status<T extends string>(value: unknown, map: Record<string, T>, fallback: T) {
  return typeof value === "string" ? map[value] ?? fallback : fallback;
}

export function normalizeVacancy(raw: unknown): Vacancy {
  const r = (raw ?? {}) as Wire;
  return {
    id: num(r.id),
    owner_id: num(r.owner_id),
    pvz_id: num(r.pvz_id),
    status: status(r.status, VACANCY_STATUS, "unspecified"),
    title: optStr(r.title),
    description: optStr(r.description),
    shift_start_at: str(r.shift_start_at),
    shift_end_at: str(r.shift_end_at),
    pvz_name_snapshot: str(r.pvz_name_snapshot),
    pvz_address_snapshot: str(r.pvz_address_snapshot),
    pay_amount: num(r.pay_amount),
    slots_total: num(r.slots_total),
    created_at: str(r.created_at),
    updated_at: str(r.updated_at),
  };
}

export function normalizeResponse(raw: unknown): Response {
  const r = (raw ?? {}) as Wire;
  return {
    id: num(r.id),
    vacancy_id: num(r.vacancy_id),
    worker_id: num(r.worker_id),
    status: status(r.status, RESPONSE_STATUS, "unspecified"),
    message: optStr(r.message),
    created_at: str(r.created_at),
    updated_at: str(r.updated_at),
  };
}

export function normalizeAssignment(raw: unknown): Assignment {
  const r = (raw ?? {}) as Wire;
  return {
    id: num(r.id),
    vacancy_id: num(r.vacancy_id),
    response_id: num(r.response_id),
    owner_id: num(r.owner_id),
    worker_id: num(r.worker_id),
    status: status(r.status, ASSIGNMENT_STATUS, "unspecified"),
    accepted_at: str(r.accepted_at),
    pvz_name_snapshot: str(r.pvz_name_snapshot),
    pvz_address_snapshot: str(r.pvz_address_snapshot),
    shift_start_at_snapshot: str(r.shift_start_at_snapshot),
    shift_end_at_snapshot: str(r.shift_end_at_snapshot),
    pay_amount_snapshot: num(r.pay_amount_snapshot),
    owner_name_snapshot: str(r.owner_name_snapshot),
    worker_name_snapshot: str(r.worker_name_snapshot),
    completed_at: optStr(r.completed_at),
    cancelled_at: optStr(r.cancelled_at),
    cancelled_by_user_id: r.cancelled_by_user_id === undefined ? null : num(r.cancelled_by_user_id),
    cancel_reason: optStr(r.cancel_reason),
    created_at: str(r.created_at),
    updated_at: str(r.updated_at),
  };
}

export function normalizeReview(raw: unknown): Review {
  const r = (raw ?? {}) as Wire;
  return {
    id: num(r.id),
    assignment_id: num(r.assignment_id),
    author_user_id: num(r.author_user_id),
    target_user_id: num(r.target_user_id),
    target_role: status(r.target_role, REVIEW_TARGET_ROLE, "unspecified"),
    rating: num(r.rating),
    body: optStr(r.body),
    created_at: str(r.created_at),
    updated_at: str(r.updated_at),
  };
}

export function normalizeReviewReply(raw: unknown): ReviewReply {
  const r = (raw ?? {}) as Wire;
  return {
    id: num(r.id),
    review_id: num(r.review_id),
    author_user_id: num(r.author_user_id),
    target_user_id: num(r.target_user_id),
    body: str(r.body),
    created_at: str(r.created_at),
    updated_at: str(r.updated_at),
  };
}

export function normalizeUserRating(raw: unknown): UserRating {
  const r = (raw ?? {}) as Wire;
  return {
    user_id: num(r.user_id),
    avg_rating: num(r.avg_rating),
    reviews_count: num(r.reviews_count),
    updated_at: optStr(r.updated_at),
  };
}

export function normalizeVacancyList(raw: unknown): VacancyListResponse {
  const r = (raw ?? {}) as Wire;
  return {
    items: Array.isArray(r.items) ? r.items.map(normalizeVacancy) : [],
    next_page_token: typeof r.next_page_token === "string" ? r.next_page_token : undefined,
  };
}

export function normalizeResponseList(raw: unknown): ResponseListResponse {
  const r = (raw ?? {}) as Wire;
  return {
    items: Array.isArray(r.items) ? r.items.map(normalizeResponse) : [],
    next_page_token: typeof r.next_page_token === "string" ? r.next_page_token : undefined,
  };
}

export function normalizeAssignmentList(raw: unknown): AssignmentListResponse {
  const r = (raw ?? {}) as Wire;
  return {
    items: Array.isArray(r.items) ? r.items.map(normalizeAssignment) : [],
    next_page_token: typeof r.next_page_token === "string" ? r.next_page_token : undefined,
  };
}

export function normalizeReviewList(raw: unknown): ReviewListResponse {
  const r = (raw ?? {}) as Wire;
  return {
    items: Array.isArray(r.items) ? r.items.map(normalizeReview) : [],
    next_page_token: typeof r.next_page_token === "string" ? r.next_page_token : undefined,
  };
}

export function normalizeReviewReplyList(raw: unknown): ReviewReplyListResponse {
  const r = (raw ?? {}) as Wire;
  return {
    items: Array.isArray(r.items) ? r.items.map(normalizeReviewReply) : [],
    next_page_token: typeof r.next_page_token === "string" ? r.next_page_token : undefined,
  };
}

export function matchingQuery(filters: VacancyFilters = {}, pageSize = 20) {
  const params = new URLSearchParams({ page_size: String(pageSize) });
  for (const [key, value] of Object.entries(filters)) {
    if (value !== undefined && value !== "" && value !== "unspecified") {
      params.set(key, String(value));
    }
  }
  return params.toString();
}

export async function createVacancy(input: CreateVacancyInput) {
  return normalizeVacancy(await apiRequest<unknown>("/v1/vacancies", { method: "POST", body: input }));
}

export async function createResponse(vacancyId: number, message?: string) {
  return normalizeResponse(
    await apiRequest<unknown>(`/v1/vacancies/${vacancyId}/responses`, {
      method: "POST",
      body: { message: message?.trim() || undefined },
    }),
  );
}

export async function withdrawResponse(responseId: number) {
  return normalizeResponse(await apiRequest<unknown>(`/v1/responses/${responseId}/withdraw`, { method: "POST", body: {} }));
}

export async function acceptResponse(responseId: number) {
  return normalizeAssignment(await apiRequest<unknown>(`/v1/responses/${responseId}/accept`, { method: "POST", body: {} }));
}

export async function cancelAssignment(assignmentId: number, cancelReason?: string) {
  return normalizeAssignment(
    await apiRequest<unknown>(`/v1/assignments/${assignmentId}/cancel`, {
      method: "POST",
      body: { cancel_reason: cancelReason?.trim() || undefined },
    }),
  );
}

export async function completeAssignment(assignmentId: number) {
  return normalizeAssignment(await apiRequest<unknown>(`/v1/assignments/${assignmentId}/complete`, { method: "POST", body: {} }));
}

export async function createReview(assignmentId: number, rating: number, body?: string) {
  return normalizeReview(
    await apiRequest<unknown>(`/v1/assignments/${assignmentId}/reviews`, {
      method: "POST",
      body: { rating, body: body?.trim() || undefined },
    }),
  );
}

export async function updateReview(reviewId: number, rating: number, body?: string) {
  return normalizeReview(
    await apiRequest<unknown>(`/v1/reviews/${reviewId}`, {
      method: "PATCH",
      body: { rating, body: body?.trim() || undefined },
    }),
  );
}

export async function createReviewReply(reviewId: number, targetUserId: number, body: string) {
  return normalizeReviewReply(
    await apiRequest<unknown>(`/v1/reviews/${reviewId}/replies`, {
      method: "POST",
      body: { target_user_id: targetUserId, body },
    }),
  );
}

export async function moderationAcceptResponse(responseId: number, reason: string) {
  return normalizeAssignment(
    await apiRequest<unknown>(`/v1/moderation/matching/responses/${responseId}/accept`, {
      method: "POST",
      body: { reason },
    }),
  );
}

export async function moderationCancelAssignment(assignmentId: number, reason: string, cancelReason?: string) {
  return normalizeAssignment(
    await apiRequest<unknown>(`/v1/moderation/matching/assignments/${assignmentId}/cancel`, {
      method: "POST",
      body: { reason, cancel_reason: cancelReason?.trim() || undefined },
    }),
  );
}

export async function moderationCompleteAssignment(assignmentId: number, reason: string) {
  return normalizeAssignment(
    await apiRequest<unknown>(`/v1/moderation/matching/assignments/${assignmentId}/complete`, {
      method: "POST",
      body: { reason },
    }),
  );
}
