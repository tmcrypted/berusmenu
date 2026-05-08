"use client";

import { ShiftList } from "@/components/ShiftList";
import { Alert } from "@/components/ui/Alert";
import { Button } from "@/components/ui/Button";
import { mockOwnerPvz, mockOwnerVacancies } from "@/data/mock-owner";
import { roleLabel } from "@/lib/profile";
import type { AuthStatus, User } from "@/types/auth";

export function GuestHomeContent({
  status,
  user,
}: {
  status: AuthStatus;
  user: User | null;
}) {
  return (
    <>
      <div className="grid gap-6 xl:grid-cols-[1.15fr_0.85fr]">
        <div className="rounded-[2rem] border border-[var(--border)] bg-[var(--panel)] p-6 shadow-[var(--shadow-soft)] sm:p-8">
          <div className="flex flex-wrap items-center gap-3">
            <span className="inline-flex rounded-full bg-[var(--accent-soft)] px-3 py-1 text-xs font-semibold uppercase tracking-[0.2em] text-[var(--accent)]">
              Срочные смены
            </span>
            <span className="inline-flex rounded-full border border-[var(--border)] px-3 py-1 text-xs font-medium text-[var(--muted)]">
              Беру Смену . ру
            </span>
          </div>
          <h1 className="mt-5 max-w-3xl text-3xl font-semibold tracking-tight text-[var(--foreground)] sm:text-4xl lg:text-5xl">
            Быстрый поиск подмены для ПВЗ и понятный вход в рабочие смены.
          </h1>
          <p className="mt-4 max-w-2xl text-base leading-7 text-[var(--muted)] sm:text-lg">
            Маркетплейс срочного подбора на смену в ПВЗ: владельцы публикуют вакансии, исполнители
            откликаются. Здесь — публичная витрина до полного API.
          </p>
          <div className="mt-8 grid gap-3 sm:grid-cols-[1.4fr_1fr]">
            <label className="rounded-[1.5rem] border border-[var(--border)] bg-[var(--card)] p-3 shadow-[var(--shadow-soft)]">
              <span className="mb-2 block text-xs font-semibold uppercase tracking-[0.16em] text-[var(--muted)]">
                Поиск
              </span>
              <input
                readOnly
                value="Сотрудник ПВЗ, Москва"
                className="w-full bg-transparent text-sm text-[var(--foreground)] outline-none"
              />
            </label>
            <label className="rounded-[1.5rem] border border-[var(--border)] bg-[var(--card)] p-3 shadow-[var(--shadow-soft)]">
              <span className="mb-2 block text-xs font-semibold uppercase tracking-[0.16em] text-[var(--muted)]">
                Дата
              </span>
              <input
                readOnly
                value="Сегодня и завтра"
                className="w-full bg-transparent text-sm text-[var(--foreground)] outline-none"
              />
            </label>
          </div>
          <div className="mt-6 flex flex-col gap-3 sm:flex-row">
            <Button type="button" fullWidth className="sm:w-auto">
              Ищу смену
            </Button>
            <Button type="button" variant="secondary" fullWidth className="sm:w-auto">
              Ищу сотрудника
            </Button>
          </div>
        </div>

        <aside className="space-y-4 rounded-[2rem] border border-[var(--border)] bg-[var(--card)] p-6 shadow-[var(--shadow-soft)] sm:p-8">
          <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-1">
            <div className="rounded-[1.5rem] bg-[var(--surface)] p-4">
              <p className="text-sm text-[var(--muted)]">Демо-карточек смен</p>
              <p className="mt-2 text-3xl font-semibold">4</p>
            </div>
            <div className="rounded-[1.5rem] bg-[var(--surface)] p-4">
              <p className="text-sm text-[var(--muted)]">Этап</p>
              <p className="mt-2 text-base font-semibold">Auth + профиль</p>
            </div>
          </div>
          {status !== "anonymous" && user ? (
            <Alert>
              <p className="font-medium">Вы вошли как {roleLabel(user.role).toLowerCase()}</p>
              <p className="mt-2 text-sm text-[var(--muted)]">
                {status === "authenticated_unverified"
                  ? "Доступ ограничен до подтверждения почты."
                  : status === "authenticated_profile_incomplete"
                    ? "Нужно завершить профиль перед полноценной работой."
                    : user.role === "owner"
                      ? "Следующий шаг: подать заявку на одобрение ПВЗ."
                      : "Профиль завершен, можно переходить к откликам."}
              </p>
            </Alert>
          ) : (
            <Alert>
              Войдите или зарегистрируйтесь, чтобы перейти от витрины к реальному flow.
            </Alert>
          )}
        </aside>
      </div>

      <section className="mt-8 grid gap-4 md:grid-cols-3">
        <div className="rounded-[1.75rem] border border-[var(--border)] bg-[var(--card)] p-5">
          <p className="text-sm font-semibold text-[var(--foreground)]">Для исполнителя</p>
          <p className="mt-2 text-sm leading-6 text-[var(--muted)]">
            Зарегистрироваться, подтвердить email, заполнить профиль и перейти к откликам.
          </p>
        </div>
        <div className="rounded-[1.75rem] border border-[var(--border)] bg-[var(--card)] p-5">
          <p className="text-sm font-semibold text-[var(--foreground)]">Для владельца ПВЗ</p>
          <p className="mt-2 text-sm leading-6 text-[var(--muted)]">
            ПВЗ проходят модерацию в inventory-service, вакансии — в matching-service.
          </p>
        </div>
        <div className="rounded-[1.75rem] border border-[var(--border)] bg-[var(--card)] p-5">
          <p className="text-sm font-semibold text-[var(--foreground)]">Для продукта</p>
          <p className="mt-2 text-sm leading-6 text-[var(--muted)]">
            Короткий путь от vacancy к assignment, без лишней инфраструктуры на MVP.
          </p>
        </div>
      </section>

      <section className="mt-8">
        <div className="mb-5 flex flex-col gap-2 sm:flex-row sm:items-end sm:justify-between">
          <div>
            <h2 className="text-2xl font-semibold tracking-tight text-[var(--foreground)]">
              Демонстрационная лента смен
            </h2>
            <p className="mt-2 text-sm text-[var(--muted)]">
              Публичная витрина для гостей и пользователей до стадии «мои отклики».
            </p>
          </div>
        </div>
        <ShiftList />
      </section>
    </>
  );
}

export function OwnerHomeContent({ user }: { user: User }) {
  return (
    <>
      <div className="rounded-[2rem] border border-[var(--border)] bg-[var(--panel)] p-6 shadow-[var(--shadow-soft)] sm:p-8">
        <p className="text-xs font-semibold uppercase tracking-[0.2em] text-[var(--accent)]">
          Кабинет владельца ПВЗ
        </p>
        <h1 className="mt-3 text-2xl font-semibold tracking-tight text-[var(--foreground)] sm:text-3xl">
          Здравствуйте{user.full_name?.trim() ? `, ${user.full_name.trim()}` : ""}
        </h1>
        <p className="mt-2 max-w-2xl text-sm leading-6 text-[var(--muted)] sm:text-base">
          Ниже — моки точек и вакансий. Когда подключим inventory-service и matching-service, эти
          блоки станут живыми.
        </p>
        <div className="mt-6 flex flex-wrap gap-3">
          <Button type="button" className="rounded-xl">
            Добавить ПВЗ (мок)
          </Button>
          <Button type="button" variant="secondary" className="rounded-xl">
            Создать вакансию (мок)
          </Button>
        </div>
      </div>

      <div className="mt-8 grid gap-6 lg:grid-cols-2">
        <section className="rounded-[1.75rem] border border-[var(--border)] bg-[var(--card)] p-6 shadow-[var(--shadow-soft)]">
          <h2 className="text-lg font-semibold text-[var(--foreground)]">Мои ПВЗ</h2>
          <ul className="mt-4 space-y-3">
            {mockOwnerPvz.map((pvz) => (
              <li
                key={pvz.id}
                className="rounded-xl border border-[var(--border)] bg-[var(--surface)] px-4 py-3"
              >
                <p className="font-medium text-[var(--foreground)]">{pvz.title}</p>
                <p className="text-sm text-[var(--muted)]">{pvz.address}</p>
                <p className="mt-1 text-xs font-medium text-[var(--accent)]">{pvz.statusLabel}</p>
              </li>
            ))}
          </ul>
        </section>
        <section className="rounded-[1.75rem] border border-[var(--border)] bg-[var(--card)] p-6 shadow-[var(--shadow-soft)]">
          <h2 className="text-lg font-semibold text-[var(--foreground)]">Мои вакансии</h2>
          <ul className="mt-4 space-y-3">
            {mockOwnerVacancies.map((v) => (
              <li
                key={v.id}
                className="rounded-xl border border-[var(--border)] bg-[var(--surface)] px-4 py-3"
              >
                <p className="font-medium text-[var(--foreground)]">{v.title}</p>
                <p className="text-sm text-[var(--muted)]">{v.pvzLabel}</p>
                <p className="mt-1 text-xs text-[var(--muted)]">
                  {v.windowLabel} · {v.payLabel}
                </p>
              </li>
            ))}
          </ul>
        </section>
      </div>
    </>
  );
}

export function WorkerHomeContent({ user }: { user: User }) {
  return (
    <>
      <div className="rounded-[2rem] border border-[var(--border)] bg-[var(--panel)] p-6 shadow-[var(--shadow-soft)] sm:p-8">
        <p className="text-xs font-semibold uppercase tracking-[0.2em] text-[var(--accent)]">
          Кабинет исполнителя
        </p>
        <h1 className="mt-3 text-2xl font-semibold tracking-tight text-[var(--foreground)] sm:text-3xl">
          Смены рядом{user.full_name?.trim() ? `, ${user.full_name.trim()}` : ""}
        </h1>
        <p className="mt-2 max-w-2xl text-sm leading-6 text-[var(--muted)] sm:text-base">
          Ниже публичная лента и моки. Отклики и чаты появятся после matching-service и
          communication-service.
        </p>
        <div className="mt-6 flex flex-wrap gap-3">
          <Button type="button" className="rounded-xl">
            Мои отклики (мок)
          </Button>
          <Button type="button" variant="secondary" className="rounded-xl">
            Сохранённые смены (мок)
          </Button>
        </div>
      </div>

      <section className="mt-8">
        <h2 className="text-xl font-semibold tracking-tight text-[var(--foreground)]">
          Лента смен для подмены
        </h2>
        <p className="mt-2 text-sm text-[var(--muted)]">
          То же демо, что и на публичной витрине — разделим персональные рекомендации, когда будет
          API.
        </p>
        <div className="mt-5">
          <ShiftList />
        </div>
      </section>
    </>
  );
}
