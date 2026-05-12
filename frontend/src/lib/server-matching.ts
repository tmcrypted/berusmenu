import "server-only";

import { serverApiRequest, type ServerApiResult } from "@/lib/server-api";
import {
  matchingQuery,
  normalizeAssignment,
  normalizeAssignmentList,
  normalizeResponseList,
  normalizeReviewList,
  normalizeReviewReplyList,
  normalizeUserRating,
  normalizeVacancy,
  normalizeVacancyList,
} from "@/lib/matching";
import type {
  Assignment,
  AssignmentListResponse,
  ResponseListResponse,
  ReviewListResponse,
  ReviewReplyListResponse,
  UserRating,
  Vacancy,
  VacancyFilters,
  VacancyListResponse,
} from "@/types/matching";

function mapResult<T, U>(result: ServerApiResult<T>, mapper: (value: T) => U): ServerApiResult<U> {
  return result.ok ? { ok: true, data: mapper(result.data) } : result;
}

export async function listVacancies(filters: VacancyFilters = {}, pageSize = 20) {
  return mapResult(await serverApiRequest<unknown>(`/v1/vacancies?${matchingQuery(filters, pageSize)}`), normalizeVacancyList);
}

export async function listModerationVacancies(filters: VacancyFilters = {}, pageSize = 20) {
  return mapResult(
    await serverApiRequest<unknown>(`/v1/moderation/matching/vacancies?${matchingQuery(filters, pageSize)}`),
    normalizeVacancyList,
  );
}

export async function getVacancy(id: number): Promise<ServerApiResult<Vacancy>> {
  return mapResult(await serverApiRequest<unknown>(`/v1/vacancies/${id}`), normalizeVacancy);
}

export async function listResponsesForVacancy(vacancyId: number): Promise<ServerApiResult<ResponseListResponse>> {
  return mapResult(await serverApiRequest<unknown>(`/v1/vacancies/${vacancyId}/responses?page_size=50`), normalizeResponseList);
}

export async function listMyResponses(pageToken?: string, pageSize = 20): Promise<ServerApiResult<ResponseListResponse>> {
  const query = pageToken ? `?page_size=${pageSize}&page_token=${encodeURIComponent(pageToken)}` : `?page_size=${pageSize}`;
  return mapResult(await serverApiRequest<unknown>(`/v1/responses/me${query}`), normalizeResponseList);
}

export async function getAssignment(id: number): Promise<ServerApiResult<Assignment>> {
  return mapResult(await serverApiRequest<unknown>(`/v1/assignments/${id}`), normalizeAssignment);
}

export async function listMyAssignments(pageToken?: string): Promise<ServerApiResult<AssignmentListResponse>> {
  const query = pageToken ? `?page_size=20&page_token=${encodeURIComponent(pageToken)}` : "?page_size=20";
  return mapResult(await serverApiRequest<unknown>(`/v1/assignments/me${query}`), normalizeAssignmentList);
}

export async function getAssignmentForVacancy(vacancyId: number): Promise<ServerApiResult<Assignment | null>> {
  const result = await serverApiRequest<unknown>(`/v1/vacancies/${vacancyId}/assignment`);
  if (!result.ok && result.status === 404) {
    return { ok: true, data: null };
  }
  return mapResult(result, normalizeAssignment);
}

export async function listReviews(targetUserId?: number): Promise<ServerApiResult<ReviewListResponse>> {
  const query = targetUserId ? `?page_size=20&target_user_id=${targetUserId}` : "?page_size=20";
  return mapResult(await serverApiRequest<unknown>(`/v1/reviews${query}`), normalizeReviewList);
}

export async function listReviewReplies(reviewId: number): Promise<ServerApiResult<ReviewReplyListResponse>> {
  return mapResult(await serverApiRequest<unknown>(`/v1/reviews/${reviewId}/replies?page_size=50`), normalizeReviewReplyList);
}

export async function getUserRating(userId: number): Promise<ServerApiResult<UserRating>> {
  return mapResult(await serverApiRequest<unknown>(`/v1/users/${userId}/rating`), normalizeUserRating);
}

export function emptyVacancyList(): VacancyListResponse {
  return { items: [] };
}
