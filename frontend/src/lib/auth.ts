import type {
  ApiUserWire,
  AuthPayloadResponse,
  VerificationCodeResponse,
  VerifyEmailResponse,
} from "@/types/auth";
import { apiRequest } from "./api";

export type RegisterInput = {
  email: string;
  password: string;
  role: "worker" | "owner";
};

export type LoginInput = {
  email: string;
  password: string;
};

export type UpdateProfileInput = {
  full_name?: string;
  phone?: string;
  avatar_url?: string;
};

export function register(input: RegisterInput) {
  return apiRequest<AuthPayloadResponse>("/v1/auth/register", {
    method: "POST",
    body: input,
  });
}

export function login(input: LoginInput) {
  return apiRequest<AuthPayloadResponse>("/v1/auth/login", {
    method: "POST",
    body: input,
  });
}

export function refreshSession() {
  return apiRequest<AuthPayloadResponse>("/v1/auth/refresh", {
    method: "POST",
    body: {},
  });
}

export function logout() {
  return apiRequest<void>("/v1/auth/logout", {
    method: "POST",
    body: {},
  });
}

export function resendVerificationCode() {
  return apiRequest<VerificationCodeResponse>("/v1/auth/resend-code", {
    method: "POST",
    body: {},
  });
}

export function verifyEmail(code: string) {
  return apiRequest<VerifyEmailResponse>("/v1/auth/verify-email", {
    method: "POST",
    body: { code },
  });
}

export function getMe() {
  return apiRequest<ApiUserWire>("/v1/me");
}

export function updateMe(input: UpdateProfileInput) {
  return apiRequest<ApiUserWire>("/v1/me", {
    method: "PATCH",
    body: input,
  });
}
