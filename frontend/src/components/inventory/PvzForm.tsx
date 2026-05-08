"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useRouter } from "next/navigation";
import { useForm } from "react-hook-form";
import { z } from "zod";
import { ApiError } from "@/lib/api";
import { createPvz, updatePvz } from "@/lib/inventory";
import type { Pvz } from "@/types/inventory";
import { Alert } from "@/components/ui/Alert";
import { Button } from "@/components/ui/Button";
import { Field } from "@/components/ui/Field";
import { Input } from "@/components/ui/Input";

const schema = z.object({
  city: z.string().trim().min(2, "Укажите город"),
  address: z.string().trim().min(5, "Укажите адрес"),
  company_name: z.string().trim().min(2, "Укажите название компании или ПВЗ"),
  contact_phone: z.string().trim().min(6, "Укажите контактный телефон"),
  contact_telegram: z.string().optional(),
  description: z.string().optional(),
  lat: z
    .string()
    .optional()
    .transform((value) => {
      const trimmed = value?.trim() ?? "";
      return trimmed ? Number(trimmed) : undefined;
    })
    .refine((value) => value === undefined || Number.isFinite(value), "Широта должна быть числом"),
  lng: z
    .string()
    .optional()
    .transform((value) => {
      const trimmed = value?.trim() ?? "";
      return trimmed ? Number(trimmed) : undefined;
    })
    .refine((value) => value === undefined || Number.isFinite(value), "Долгота должна быть числом"),
});

type FormInput = z.input<typeof schema>;
type FormValues = z.output<typeof schema>;

type Props = {
  mode: "create" | "edit";
  pvz?: Pvz;
};

export function PvzForm({ mode, pvz }: Props) {
  const router = useRouter();
  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
    setError,
  } = useForm<FormInput, unknown, FormValues>({
    resolver: zodResolver(schema),
    defaultValues: {
      city: pvz?.city ?? "",
      address: pvz?.address ?? "",
      company_name: pvz?.company_name ?? "",
      contact_phone: pvz?.contact_phone ?? "",
      contact_telegram: pvz?.contact_telegram ?? "",
      description: pvz?.description ?? "",
      lat: typeof pvz?.lat === "number" ? String(pvz.lat) : "",
      lng: typeof pvz?.lng === "number" ? String(pvz.lng) : "",
    },
  });

  async function onSubmit(values: FormValues) {
    try {
      if (mode === "edit" && pvz) {
        await updatePvz(pvz.id, values);
      } else {
        await createPvz(values);
      }
      router.push("/owner/pvz?submitted=1");
      router.refresh();
    } catch (error) {
      if (error instanceof ApiError) {
        setError("root", { message: inventoryErrorMessage(error) });
        return;
      }
      setError("root", { message: "Не удалось сохранить ПВЗ. Попробуйте еще раз." });
    }
  }

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-5">
      {errors.root?.message ? <Alert tone="error">{errors.root.message}</Alert> : null}
      <div className="grid gap-5 md:grid-cols-2">
        <Field label="Город" htmlFor="pvz-city" error={errors.city?.message}>
          <Input id="pvz-city" autoComplete="address-level2" {...register("city")} />
        </Field>
        <Field label="Название компании / ПВЗ" htmlFor="pvz-company" error={errors.company_name?.message}>
          <Input id="pvz-company" autoComplete="organization" {...register("company_name")} />
        </Field>
      </div>

      <Field label="Адрес" htmlFor="pvz-address" error={errors.address?.message}>
        <Input id="pvz-address" autoComplete="street-address" {...register("address")} />
      </Field>

      <div className="grid gap-5 md:grid-cols-2">
        <Field label="Контактный телефон" htmlFor="pvz-phone" error={errors.contact_phone?.message}>
          <Input id="pvz-phone" autoComplete="tel" {...register("contact_phone")} />
        </Field>
        <Field label="Telegram" htmlFor="pvz-telegram" hint="Необязательно" error={errors.contact_telegram?.message}>
          <Input id="pvz-telegram" placeholder="@username" {...register("contact_telegram")} />
        </Field>
      </div>

      <Field label="Описание" htmlFor="pvz-description" hint="Необязательно" error={errors.description?.message}>
        <textarea
          id="pvz-description"
          rows={4}
          className="w-full rounded-2xl border border-[var(--border)] bg-[var(--surface)] px-3.5 py-2.5 text-sm text-[var(--foreground)] outline-none transition placeholder:text-[var(--muted)] focus:border-[var(--accent)] focus:ring-4 focus:ring-[var(--accent-soft)]"
          {...register("description")}
        />
      </Field>

      <div className="grid gap-5 md:grid-cols-2">
        <Field label="Широта" htmlFor="pvz-lat" hint="Необязательно" error={errors.lat?.message}>
          <Input id="pvz-lat" inputMode="decimal" {...register("lat")} />
        </Field>
        <Field label="Долгота" htmlFor="pvz-lng" hint="Необязательно" error={errors.lng?.message}>
          <Input id="pvz-lng" inputMode="decimal" {...register("lng")} />
        </Field>
      </div>

      <div className="flex flex-col gap-3 sm:flex-row">
        <Button type="submit" disabled={isSubmitting}>
          {isSubmitting ? "Сохраняем..." : mode === "edit" ? "Сохранить и отправить на модерацию" : "Отправить на модерацию"}
        </Button>
        <Button type="button" variant="secondary" onClick={() => router.push("/owner/pvz")}>
          Отмена
        </Button>
      </div>
    </form>
  );
}

function inventoryErrorMessage(error: ApiError) {
  if (error.status === 503) {
    return "Проверка активных вакансий временно недоступна, попробуйте позже.";
  }
  if (error.status === 412) {
    return "У этого ПВЗ есть активная вакансия. Редактирование сейчас недоступно.";
  }
  return error.message;
}
