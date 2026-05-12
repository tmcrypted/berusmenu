import { Button } from "@/components/ui/Button";
import { Field } from "@/components/ui/Field";
import { Input } from "@/components/ui/Input";
import { Select } from "@/components/ui/Select";
import type { VacancyFilters } from "@/types/matching";

type Props = {
  filters: VacancyFilters;
  action?: string;
  showStatus?: boolean;
  showPvz?: boolean;
};

export function VacancyFilters({ filters, action, showStatus = true, showPvz = true }: Props) {
  return (
    <form action={action} className="rounded-[1.5rem] border border-[var(--border)] bg-[var(--panel)] p-4 shadow-[var(--shadow-soft)]">
      <div className="grid gap-4 md:grid-cols-5">
        {showStatus ? (
          <Field label="Статус" htmlFor="status">
            <Select id="status" name="status" defaultValue={filters.status ?? ""}>
              <option value="">Любой</option>
              <option value="published">Опубликована</option>
              <option value="assigned">Назначена</option>
              <option value="completed">Завершена</option>
              <option value="cancelled">Отменена</option>
            </Select>
          </Field>
        ) : null}
        {showPvz ? (
          <Field label="PVZ ID" htmlFor="pvz_id">
            <Input id="pvz_id" name="pvz_id" inputMode="numeric" defaultValue={filters.pvz_id ?? ""} />
          </Field>
        ) : null}
        <Field label="Смена от" htmlFor="shift_start_from">
          <Input id="shift_start_from" name="shift_start_from" type="datetime-local" defaultValue={toLocalInput(filters.shift_start_from)} />
        </Field>
        <Field label="Смена до" htmlFor="shift_start_to">
          <Input id="shift_start_to" name="shift_start_to" type="datetime-local" defaultValue={toLocalInput(filters.shift_start_to)} />
        </Field>
        <div className="flex items-end">
          <Button type="submit" fullWidth>
            Показать
          </Button>
        </div>
      </div>
    </form>
  );
}

function toLocalInput(value?: string) {
  if (!value) {
    return "";
  }
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return "";
  }
  return date.toISOString().slice(0, 16);
}
