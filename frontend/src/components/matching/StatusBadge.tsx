import { assignmentStatusLabel, responseStatusLabel, vacancyStatusLabel } from "@/lib/matching-format";
import type { AssignmentStatus, ResponseStatus, VacancyStatus } from "@/types/matching";

type Props =
  | { kind: "vacancy"; status: VacancyStatus }
  | { kind: "response"; status: ResponseStatus }
  | { kind: "assignment"; status: AssignmentStatus };

export function StatusBadge(props: Props) {
  const label =
    props.kind === "vacancy"
      ? vacancyStatusLabel(props.status)
      : props.kind === "response"
        ? responseStatusLabel(props.status)
        : assignmentStatusLabel(props.status);

  return (
    <span className="inline-flex rounded-full border border-[var(--border)] bg-[var(--surface)] px-3 py-1 text-xs font-semibold text-[var(--muted)]">
      {label}
    </span>
  );
}
