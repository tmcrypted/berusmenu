export type UserRole = "worker" | "owner" | "moderator" | "unspecified";

export type ApiUserWire = Record<string, unknown>;

export type User = {
  id: number;
  email: string;
  email_verified_at?: string | null;
  role: UserRole;
  full_name: string;
  phone?: string | null;
  avatar_url?: string | null;
  created_at: string;
  updated_at: string;
};

export type AuthPayloadResponse = {
  user: ApiUserWire;
  access_expires_at: string;
};

export type VerificationCodeResponse = {
  code_expires_at: string;
  resend_available_at: string;
};

export type VerifyEmailResponse = {
  user: ApiUserWire;
};

export type AuthStatus =
  | "loading"
  | "anonymous"
  | "authenticated_unverified"
  | "authenticated_profile_incomplete"
  | "authenticated_ready";
