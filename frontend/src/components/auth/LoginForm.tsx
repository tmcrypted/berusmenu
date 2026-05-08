"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import { z } from "zod";
import { login } from "@/lib/auth";
import { ApiError } from "@/lib/api";
import { resolveApiAuthSnapshot } from "@/lib/auth-state";
import { profileCompletionPath } from "@/lib/profile";
import { useAuthStore } from "@/store/auth-store";
import { Alert } from "../ui/Alert";
import { Button } from "../ui/Button";
import { Field } from "../ui/Field";
import { Input } from "../ui/Input";
import { PasswordInput } from "../ui/PasswordInput";
import { AuthGuard } from "./AuthGuard";
import { AuthShell } from "./AuthShell";

const schema = z.object({
  email: z.email("Введите корректный email"),
  password: z.string().min(8, "Пароль должен быть не короче 8 символов"),
});

type FormValues = z.infer<typeof schema>;

export function LoginForm() {
  const router = useRouter();
  const hydrate = useAuthStore((state) => state.hydrate);
  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
    setError,
  } = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: {
      email: "",
      password: "",
    },
  });

  async function onSubmit(values: FormValues) {
    try {
      const auth = await login(values);
      const snapshot = resolveApiAuthSnapshot(auth.user, auth.access_expires_at);
      hydrate(snapshot);

      if (snapshot.status === "authenticated_unverified") {
        router.push("/verify-email");
        return;
      }

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
        message: "Не удалось выполнить вход. Попробуйте еще раз.",
      });
    }
  }

  return (
    <AuthGuard mode="anonymous">
      <AuthShell
        title="Вход"
        subtitle="Войдите, чтобы откликаться на смены, подтверждать почту и продолжать заполнение профиля."
        footer={
          <>
            Нет аккаунта?{" "}
            <Link
              href="/register"
              className="font-semibold text-[var(--accent)] hover:text-[var(--accent-hover)]"
            >
              Создать аккаунт
            </Link>
          </>
        }
      >
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-5">
          {errors.root?.message ? <Alert tone="error">{errors.root.message}</Alert> : null}
          <Field label="Email" htmlFor="login-email" error={errors.email?.message}>
            <Input
              id="login-email"
              type="email"
              autoComplete="email"
              placeholder="you@example.com"
              {...register("email")}
            />
          </Field>
          <Field label="Пароль" htmlFor="login-password" error={errors.password?.message}>
            <PasswordInput
              id="login-password"
              autoComplete="current-password"
              placeholder="Минимум 8 символов"
              {...register("password")}
            />
          </Field>
          <Button type="submit" fullWidth disabled={isSubmitting}>
            {isSubmitting ? "Входим..." : "Войти"}
          </Button>
        </form>
      </AuthShell>
    </AuthGuard>
  );
}
