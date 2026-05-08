"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import { z } from "zod";
import { logout, resendVerificationCode, verifyEmail } from "@/lib/auth";
import { ApiError } from "@/lib/api";
import { resolveApiAuthSnapshot } from "@/lib/auth-state";
import { profileCompletionPath } from "@/lib/profile";
import { useAuthStore } from "@/store/auth-store";
import { Alert } from "../ui/Alert";
import { Button } from "../ui/Button";
import { Field } from "../ui/Field";
import { Input } from "../ui/Input";
import { AuthGuard } from "./AuthGuard";
import { AuthShell } from "./AuthShell";

const schema = z.object({
  code: z
    .string()
    .trim()
    .min(4, "Введите код подтверждения")
    .max(12, "Слишком длинный код"),
});

type FormValues = z.infer<typeof schema>;

export function VerifyEmailForm() {
  const router = useRouter();
  const user = useAuthStore((state) => state.user);
  const hydrate = useAuthStore((state) => state.hydrate);
  const clearAuth = useAuthStore((state) => state.clearAuth);
  const [serverMessage, setServerMessage] = useState<string | null>(null);
  const [resendAvailableAt, setResendAvailableAt] = useState<string | null>(null);
  const [secondsLeft, setSecondsLeft] = useState(0);
  const [isResending, setIsResending] = useState(false);
  const [isLoggingOut, setIsLoggingOut] = useState(false);
  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
    setError,
  } = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: {
      code: "",
    },
  });

  useEffect(() => {
    if (!resendAvailableAt) {
      setSecondsLeft(0);
      return;
    }

    const tick = () => {
      const diff = Math.ceil((new Date(resendAvailableAt).getTime() - Date.now()) / 1000);
      setSecondsLeft(Math.max(diff, 0));
    };

    tick();
    const interval = window.setInterval(tick, 1000);

    return () => window.clearInterval(interval);
  }, [resendAvailableAt]);

  async function onSubmit(values: FormValues) {
    setServerMessage(null);

    try {
      const result = await verifyEmail(values.code);
      const snapshot = resolveApiAuthSnapshot(result.user);
      hydrate(snapshot);

      if (snapshot.status === "authenticated_profile_incomplete" && snapshot.user) {
        router.push(profileCompletionPath(snapshot.user.role));
        return;
      }

      router.push("/");
    } catch (error) {
      if (error instanceof ApiError) {
        setError("root", { message: error.message });
        return;
      }
      setError("root", {
        message: "Не удалось подтвердить почту. Попробуйте еще раз.",
      });
    }
  }

  async function handleResend() {
    setServerMessage(null);
    setIsResending(true);

    try {
      const data = await resendVerificationCode();
      setResendAvailableAt(data.resend_available_at);
      setServerMessage("Код отправлен повторно. Проверьте почту и локальные логи.");
    } catch (error) {
      if (error instanceof ApiError) {
        setServerMessage(error.message);
      } else {
        setServerMessage("Не удалось отправить код повторно.");
      }
    } finally {
      setIsResending(false);
    }
  }

  async function handleLogout() {
    setIsLoggingOut(true);

    try {
      await logout();
    } finally {
      clearAuth();
      router.push("/login");
      setIsLoggingOut(false);
    }
  }

  return (
    <AuthGuard mode="unverified">
      <AuthShell
        title="Подтвердите email"
        subtitle="Пока почта не подтверждена, backend разрешает только проверку кода, повторную отправку и выход из аккаунта."
      >
        <div className="space-y-5">
          <Alert>
            <p className="font-medium">Текущий email</p>
            <p className="mt-1 text-sm text-[var(--muted)]">{user?.email}</p>
          </Alert>
          {serverMessage ? <Alert tone="info">{serverMessage}</Alert> : null}
          {errors.root?.message ? <Alert tone="error">{errors.root.message}</Alert> : null}
          <form onSubmit={handleSubmit(onSubmit)} className="space-y-5">
            <Field
              label="Код подтверждения"
              htmlFor="verification-code"
              hint="В local/dev код печатается в логах auth-service или notification-service."
              error={errors.code?.message}
            >
              <Input
                id="verification-code"
                inputMode="numeric"
                autoComplete="one-time-code"
                placeholder="Например, 123456"
                {...register("code")}
              />
            </Field>
            <div className="grid gap-3 sm:grid-cols-2">
              <Button type="submit" fullWidth disabled={isSubmitting}>
                {isSubmitting ? "Проверяем..." : "Подтвердить email"}
              </Button>
              <Button
                type="button"
                variant="secondary"
                fullWidth
                disabled={isResending || secondsLeft > 0}
                onClick={() => void handleResend()}
              >
                {secondsLeft > 0
                  ? `Повторить через ${secondsLeft}с`
                  : isResending
                    ? "Отправляем..."
                    : "Отправить код снова"}
              </Button>
            </div>
          </form>
          <Button
            type="button"
            variant="ghost"
            fullWidth
            disabled={isLoggingOut}
            onClick={() => void handleLogout()}
          >
            {isLoggingOut ? "Выходим..." : "Выйти из аккаунта"}
          </Button>
        </div>
      </AuthShell>
    </AuthGuard>
  );
}
