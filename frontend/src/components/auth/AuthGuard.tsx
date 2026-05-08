"use client";

import type { ReactNode } from "react";
import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { profileCompletionPath } from "@/lib/profile";
import type { UserRole } from "@/types/auth";
import { Spinner } from "../ui/Spinner";
import { useResolvedAuthSnapshot } from "./AuthStoreHydrator";

type Props = {
  children: ReactNode;
  mode: "anonymous" | "unverified" | "profile_completion";
  role?: UserRole;
};

export function AuthGuard({ children, mode, role }: Props) {
  const router = useRouter();
  const { initialized, status, user } = useResolvedAuthSnapshot();

  useEffect(() => {
    if (!initialized) {
      return;
    }

    if (mode === "anonymous") {
      if (status === "anonymous") {
        return;
      }
      if (status === "authenticated_unverified") {
        router.replace("/verify-email");
        return;
      }
      if (status === "authenticated_profile_incomplete" && user) {
        router.replace(profileCompletionPath(user.role));
        return;
      }
      router.replace("/");
      return;
    }

    if (mode === "unverified") {
      if (status === "anonymous") {
        router.replace("/login");
        return;
      }
      if (status === "authenticated_unverified") {
        return;
      }
      if (status === "authenticated_profile_incomplete" && user) {
        router.replace(profileCompletionPath(user.role));
        return;
      }
      router.replace("/");
      return;
    }

    if (status === "anonymous") {
      router.replace("/login");
      return;
    }
    if (status === "authenticated_unverified") {
      router.replace("/verify-email");
      return;
    }
    if (status === "authenticated_profile_incomplete" && user) {
      if (!role || role === user.role || user.role === "unspecified") {
        return;
      }
      router.replace(profileCompletionPath(user.role));
      return;
    }
    if (status === "authenticated_ready") {
      router.replace("/");
    }
  }, [initialized, mode, role, router, status, user]);

  if (!initialized || status === "loading") {
    return (
      <div className="flex min-h-[calc(100vh_-_var(--header-height))] items-center justify-center">
        <Spinner />
      </div>
    );
  }

  if (mode === "anonymous" && status === "anonymous") {
    return <>{children}</>;
  }

  if (mode === "unverified" && status === "authenticated_unverified") {
    return <>{children}</>;
  }

  if (
    mode === "profile_completion" &&
    status === "authenticated_profile_incomplete" &&
    user &&
    (!role || role === user.role || user.role === "unspecified")
  ) {
    return <>{children}</>;
  }

  return (
      <div className="flex min-h-[calc(100vh_-_var(--header-height))] items-center justify-center">
      <Spinner />
    </div>
  );
}
