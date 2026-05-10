"use client";

import { useEffect, useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { Button } from "@/components/ui/Button";
import { Field } from "@/components/ui/Field";
import { Input } from "@/components/ui/Input";
import { Select } from "@/components/ui/Select";
import { ApiError } from "@/lib/api";
import { listOwnerPvz } from "@/lib/inventory";
import { createVacancy } from "@/lib/matching";
import type { Pvz } from "@/types/inventory";

export function VacancyForm() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const [pvz, setPvz] = useState<Pvz[]>([]);
  const [pvzId, setPvzId] = useState(searchParams.get("pvz_id") ?? "");
  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [start, setStart] = useState("");
  const [end, setEnd] = useState("");
  const [pay, setPay] = useState("");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    async function load() {
      try {
        const response = await listOwnerPvz();
        if (!cancelled) {
          setPvz(response.items);
        }
      } catch {
        if (!cancelled) {
          setError("Не удалось загрузить PVZ.");
        }
      }
    }
    void load();
    return () => {
      cancelled = true;
    };
  }, []);

  async function submit() {
    setBusy(true);
    setError(null);
    try {
      const created = await createVacancy({
        pvz_id: Number(pvzId),
        title: title.trim() || undefined,
        description: description.trim() || undefined,
        shift_start_at: toRFC3339(start),
        shift_end_at: toRFC3339(end),
        pay_amount: Number(pay),
        slots_total: 1,
      });
      router.push(`/owner/vacancies/${created.id}`);
      router.refresh();
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Не удалось создать вакансию.");
    } finally {
      setBusy(false);
    }
  }

  const approved = pvz.filter((item) => item.status === "approved");

  return (
    <div className="space-y-5">
      {error ? <p className="rounded-2xl border border-[var(--danger)]/30 bg-[var(--surface)] p-3 text-sm text-[var(--danger)]">{error}</p> : null}
      <Field label="PVZ" htmlFor="vacancy-pvz">
        <Select id="vacancy-pvz" value={pvzId} onChange={(event) => setPvzId(event.target.value)}>
          <option value="">Выберите одобренный PVZ</option>
          {approved.map((item) => (
            <option key={item.id} value={item.id}>
              {item.company_name} · {item.address}
            </option>
          ))}
        </Select>
      </Field>
      <Field label="Название" htmlFor="vacancy-title" hint="Необязательно">
        <Input id="vacancy-title" value={title} onChange={(event) => setTitle(event.target.value)} />
      </Field>
      <Field label="Описание" htmlFor="vacancy-description" hint="Необязательно">
        <textarea id="vacancy-description" rows={4} value={description} onChange={(event) => setDescription(event.target.value)} className="w-full rounded-2xl border border-[var(--border)] bg-[var(--surface)] px-3.5 py-2.5 text-sm outline-none" />
      </Field>
      <div className="grid gap-5 md:grid-cols-3">
        <Field label="Начало" htmlFor="vacancy-start">
          <Input id="vacancy-start" type="datetime-local" value={start} onChange={(event) => setStart(event.target.value)} />
        </Field>
        <Field label="Конец" htmlFor="vacancy-end">
          <Input id="vacancy-end" type="datetime-local" value={end} onChange={(event) => setEnd(event.target.value)} />
        </Field>
        <Field label="Оплата, руб." htmlFor="vacancy-pay">
          <Input id="vacancy-pay" inputMode="numeric" value={pay} onChange={(event) => setPay(event.target.value)} />
        </Field>
      </div>
      <div className="flex flex-wrap gap-3">
        <Button type="button" disabled={busy || !pvzId || !start || !end || !pay} onClick={() => void submit()}>
          {busy ? "Создаем..." : "Создать вакансию"}
        </Button>
        <Button type="button" variant="secondary" onClick={() => router.push("/owner/vacancies")}>
          Отмена
        </Button>
      </div>
    </div>
  );
}

function toRFC3339(value: string) {
  return value ? `${value}:00+03:00` : "";
}
