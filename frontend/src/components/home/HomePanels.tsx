"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { PvzCard } from "@/components/inventory/PvzCard";
import { VacancyRow } from "@/components/matching/VacancyRow";
import { Alert } from "@/components/ui/Alert";
import { Button } from "@/components/ui/Button";
import { listOwnerPvz } from "@/lib/inventory";
import { listPublicVacanciesClient, listVacanciesClient } from "@/lib/matching";
import { displayName, roleLabel } from "@/lib/profile";
import type { AuthStatus, User } from "@/types/auth";
import type { Pvz } from "@/types/inventory";
import type { Vacancy } from "@/types/matching";

export function GuestHomeContent({ status, user }: { status: AuthStatus; user: User | null }) {
  const [vacancies, setVacancies] = useState<Vacancy[]>([]);
  const [shiftFrom, setShiftFrom] = useState("");
  const [shiftTo, setShiftTo] = useState("");
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    async function load() {
      setLoading(true);
      setError(null);
      try {
        const result = await listPublicVacanciesClient({
          shift_start_from: toISOString(shiftFrom),
          shift_start_to: toISOString(shiftTo),
        }, 20);
        if (!cancelled) {
          setVacancies(result.items);
        }
      } catch {
        if (!cancelled) {
          setError("Не удалось загрузить смены.");
        }
      } finally {
        if (!cancelled) {
          setLoading(false);
        }
      }
    }
    void load();
    return () => {
      cancelled = true;
    };
  }, [shiftFrom, shiftTo]);

  return (
    <div className="space-y-6">
      <div className="grid gap-6 xl:grid-cols-[1.1fr_0.9fr]">
      <section className="rounded-[2rem] border border-[var(--border)] bg-[var(--panel)] p-6 shadow-[var(--shadow-soft)] sm:p-8">
        <p className="text-xs font-semibold uppercase tracking-[0.2em] text-[var(--accent)]">Беру Смену</p>
        <h1 className="mt-4 max-w-3xl text-3xl font-semibold tracking-tight text-[var(--foreground)] sm:text-4xl lg:text-5xl">
          Срочные смены для ПВЗ без долгих переписок.
        </h1>
        <p className="mt-4 max-w-2xl text-base leading-7 text-[var(--muted)] sm:text-lg">
          Владельцы публикуют смены, исполнители откликаются, назначение фиксируется в системе.
        </p>
        <div className="mt-8 flex flex-col gap-3 sm:flex-row">
          <Link href="/register/worker">
            <Button type="button" fullWidth className="sm:w-auto">Найти смену</Button>
          </Link>
          <Link href="/register/owner">
            <Button type="button" variant="secondary" fullWidth className="sm:w-auto">Найти исполнителя</Button>
          </Link>
        </div>
      </section>

      <aside className="rounded-[2rem] border border-[var(--border)] bg-[var(--card)] p-6 shadow-[var(--shadow-soft)] sm:p-8">
        {status !== "anonymous" && user ? (
          <Alert>
            <p className="font-medium">Вы вошли как {roleLabel(user.role).toLowerCase()}</p>
            <p className="mt-2 text-sm text-[var(--muted)]">
              {status === "authenticated_unverified"
                ? "Подтвердите почту, чтобы открыть рабочие разделы."
                : status === "authenticated_profile_incomplete"
                  ? "Завершите профиль, чтобы продолжить."
                  : "Можно переходить к сменам."}
            </p>
          </Alert>
        ) : (
          <div className="space-y-4">
            <h2 className="text-xl font-semibold">Как это работает</h2>
            <Step title="Владелец создает ПВЗ" text="После модерации можно публиковать смены." />
            <Step title="Исполнитель откликается" text="Отклик сразу попадает владельцу." />
            <Step title="Владелец назначает исполнителя" text="После смены стороны могут оставить отзыв." />
          </div>
        )}
      </aside>
      </div>

      <section className="grid gap-5 lg:grid-cols-[18rem_minmax(0,1fr)]">
        <aside className="rounded-[1.5rem] border border-[var(--border)] bg-[var(--panel)] p-4 shadow-[var(--shadow-soft)]">
          <h2 className="text-lg font-semibold">Смены</h2>
          <div className="mt-4 space-y-3">
            <label className="block text-sm font-medium">
              <span>Смена от</span>
              <input
                type="datetime-local"
                value={shiftFrom}
                onChange={(event) => setShiftFrom(event.target.value)}
                className="mt-1 min-h-11 w-full rounded-2xl border border-[var(--border)] bg-[var(--surface)] px-3 text-sm outline-none focus:border-[var(--accent)]"
              />
            </label>
            <label className="block text-sm font-medium">
              <span>Смена до</span>
              <input
                type="datetime-local"
                value={shiftTo}
                onChange={(event) => setShiftTo(event.target.value)}
                className="mt-1 min-h-11 w-full rounded-2xl border border-[var(--border)] bg-[var(--surface)] px-3 text-sm outline-none focus:border-[var(--accent)]"
              />
            </label>
            <Button type="button" variant="secondary" fullWidth onClick={() => { setShiftFrom(""); setShiftTo(""); }}>
              Сбросить
            </Button>
          </div>
        </aside>
        <div className="space-y-3">
          {loading ? (
            <p className="rounded-2xl border border-[var(--border)] bg-[var(--card)] p-5 text-sm text-[var(--muted)]">Загружаем смены...</p>
          ) : error ? (
            <p className="rounded-2xl border border-[var(--border)] bg-[var(--card)] p-5 text-sm text-[var(--danger)]">{error}</p>
          ) : vacancies.length === 0 ? (
            <p className="rounded-2xl border border-[var(--border)] bg-[var(--card)] p-6 text-sm text-[var(--muted)]">Подходящих смен пока нет.</p>
          ) : (
            vacancies.map((vacancy) => <VacancyRow key={vacancy.id} vacancy={vacancy} user={null} />)
          )}
        </div>
      </section>
    </div>
  );
}

export function OwnerHomeContent({ user }: { user: User }) {
  const [pvz, setPvz] = useState<Pvz[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    async function load() {
      try {
        const result = await listOwnerPvz();
        if (!cancelled) {
          setPvz(result.items);
        }
      } catch {
        if (!cancelled) {
          setError("Не удалось загрузить ПВЗ.");
        }
      } finally {
        if (!cancelled) {
          setLoading(false);
        }
      }
    }
    void load();
    return () => {
      cancelled = true;
    };
  }, []);

  return (
    <div className="space-y-6">
      <section className="rounded-[2rem] border border-[var(--border)] bg-[var(--panel)] p-6 shadow-[var(--shadow-soft)] sm:p-8">
        <p className="text-xs font-semibold uppercase tracking-[0.2em] text-[var(--accent)]">Кабинет владельца</p>
        <h1 className="mt-3 text-2xl font-semibold tracking-tight text-[var(--foreground)] sm:text-3xl">
          Здравствуйте{displayName(user, "") ? `, ${displayName(user, "")}` : ""}
        </h1>
        <p className="mt-2 max-w-2xl text-sm leading-6 text-[var(--muted)] sm:text-base">
          Управляйте ПВЗ, публикуйте смены и отслеживайте назначения.
        </p>
        <div className="mt-6 flex flex-wrap gap-3">
          <Link href="/owner/pvz"><Button type="button" className="rounded-xl">Мои ПВЗ</Button></Link>
          <Link href="/owner/vacancies"><Button type="button" variant="secondary" className="rounded-xl">Мои вакансии</Button></Link>
          <Link href="/assignments"><Button type="button" variant="secondary" className="rounded-xl">Мои назначения</Button></Link>
        </div>
      </section>
      <section>
        <div className="mb-4 flex flex-wrap items-center justify-between gap-3">
          <h2 className="text-xl font-semibold tracking-tight">Мои ПВЗ</h2>
          <Link href="/owner/pvz"><Button type="button" variant="secondary" className="rounded-xl">Все ПВЗ</Button></Link>
        </div>
        {loading ? (
          <p className="rounded-2xl border border-[var(--border)] bg-[var(--card)] p-5 text-sm text-[var(--muted)]">Загружаем ПВЗ...</p>
        ) : error ? (
          <p className="rounded-2xl border border-[var(--border)] bg-[var(--card)] p-5 text-sm text-[var(--danger)]">{error}</p>
        ) : pvz.length === 0 ? (
          <div className="rounded-2xl border border-[var(--border)] bg-[var(--card)] p-6">
            <p className="text-sm text-[var(--muted)]">У вас пока нет ПВЗ.</p>
            <Link href="/owner/pvz" className="mt-4 inline-flex"><Button type="button">Перейти к ПВЗ</Button></Link>
          </div>
        ) : (
          <div className="grid gap-4 lg:grid-cols-2">
            {pvz.slice(0, 4).map((item) => <PvzCard key={item.id} pvz={item} />)}
          </div>
        )}
      </section>
    </div>
  );
}

export function ModeratorHomeContent({ user }: { user: User }) {
  return (
    <section className="rounded-[2rem] border border-[var(--border)] bg-[var(--panel)] p-6 shadow-[var(--shadow-soft)] sm:p-8">
      <p className="text-xs font-semibold uppercase tracking-[0.2em] text-[var(--accent)]">Кабинет модератора</p>
      <h1 className="mt-3 text-2xl font-semibold tracking-tight text-[var(--foreground)] sm:text-3xl">
        Здравствуйте{displayName(user, "") ? `, ${displayName(user, "")}` : ""}
      </h1>
      <div className="mt-6 flex flex-wrap gap-3">
        <Link href="/moderation/pvz"><Button type="button" className="rounded-xl">Модерация ПВЗ</Button></Link>
        <Link href="/moderation/matching"><Button type="button" variant="secondary" className="rounded-xl">Модерация смен</Button></Link>
      </div>
    </section>
  );
}

export function WorkerHomeContent({ user }: { user: User }) {
  const [vacancies, setVacancies] = useState<Vacancy[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    async function load() {
      try {
        const result = await listVacanciesClient({}, 20);
        if (!cancelled) {
          setVacancies(result.items);
        }
      } catch {
        if (!cancelled) {
          setError("Не удалось загрузить ленту смен.");
        }
      } finally {
        if (!cancelled) {
          setLoading(false);
        }
      }
    }
    void load();
    return () => {
      cancelled = true;
    };
  }, []);

  return (
    <div className="space-y-6">
      <section className="rounded-[2rem] border border-[var(--border)] bg-[var(--panel)] p-6 shadow-[var(--shadow-soft)] sm:p-8">
        <p className="text-xs font-semibold uppercase tracking-[0.2em] text-[var(--accent)]">Кабинет исполнителя</p>
        <h1 className="mt-3 text-2xl font-semibold tracking-tight text-[var(--foreground)] sm:text-3xl">
          Смены рядом{displayName(user, "") ? `, ${displayName(user, "")}` : ""}
        </h1>
        <p className="mt-2 max-w-2xl text-sm leading-6 text-[var(--muted)] sm:text-base">
          Откликайтесь на подходящие смены и отслеживайте свои назначения.
        </p>
        <div className="mt-6 flex flex-wrap gap-3">
          <Link href="/vacancies"><Button type="button" className="rounded-xl">Лента смен</Button></Link>
          <Link href="/responses"><Button type="button" variant="secondary" className="rounded-xl">Мои отклики</Button></Link>
          <Link href="/assignments"><Button type="button" variant="secondary" className="rounded-xl">Мои назначения</Button></Link>
        </div>
      </section>
      <section>
        <div className="mb-4 flex flex-wrap items-center justify-between gap-3">
          <h2 className="text-xl font-semibold tracking-tight">Лента смен</h2>
          <Link href="/vacancies"><Button type="button" variant="secondary" className="rounded-xl">Все смены</Button></Link>
        </div>
        {loading ? (
          <p className="rounded-2xl border border-[var(--border)] bg-[var(--card)] p-5 text-sm text-[var(--muted)]">Загружаем смены...</p>
        ) : error ? (
          <p className="rounded-2xl border border-[var(--border)] bg-[var(--card)] p-5 text-sm text-[var(--danger)]">{error}</p>
        ) : vacancies.length === 0 ? (
          <p className="rounded-2xl border border-[var(--border)] bg-[var(--card)] p-6 text-sm text-[var(--muted)]">Подходящих смен пока нет.</p>
        ) : (
          <div className="space-y-3">
            {vacancies.map((vacancy) => <VacancyRow key={vacancy.id} vacancy={vacancy} user={user} />)}
          </div>
        )}
      </section>
    </div>
  );
}

function Step({ title, text }: { title: string; text: string }) {
  return (
    <div className="rounded-2xl bg-[var(--surface)] p-4">
      <p className="font-semibold">{title}</p>
      <p className="mt-1 text-sm leading-6 text-[var(--muted)]">{text}</p>
    </div>
  );
}

function toISOString(value: string) {
  if (!value) {
    return undefined;
  }
  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? undefined : date.toISOString();
}
