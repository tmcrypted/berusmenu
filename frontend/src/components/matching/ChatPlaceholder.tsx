import type { Assignment, Response } from "@/types/matching";

type Props = {
  response?: Response;
  assignment?: Assignment;
};

export function ChatPlaceholder({ response, assignment }: Props) {
  return (
    <section className="rounded-[1.5rem] border border-dashed border-[var(--border-strong)] bg-[var(--surface)] p-5">
      <p className="text-xs font-semibold uppercase tracking-[0.16em] text-[var(--accent)]">
        Chat placeholder
      </p>
      <h2 className="mt-2 text-lg font-semibold text-[var(--foreground)]">Чат-заглушка</h2>
      <p className="mt-2 text-sm leading-6 text-[var(--muted)]">
        Communication-service еще не запущен, поэтому сообщения здесь read-only mock. Реальные chat endpoints не вызываются.
      </p>
      <div className="mt-4 space-y-3 text-sm">
        {response?.message ? (
          <div className="max-w-[85%] rounded-2xl bg-[var(--card)] p-3">
            <p className="font-medium">Сообщение отклика</p>
            <p className="mt-1 text-[var(--muted)]">{response.message}</p>
          </div>
        ) : null}
        <div className="ml-auto max-w-[85%] rounded-2xl bg-[var(--accent-soft)] p-3 text-[var(--foreground)]">
          {assignment ? "Смена назначена. Здесь появится рабочий чат." : "После отклика здесь появится диалог с владельцем."}
        </div>
      </div>
    </section>
  );
}
