import { ShiftList } from "@/components/ShiftList";

export default function Home() {
  return (
    <div className="min-h-[calc(100vh-3.5rem)] bg-[var(--background)]">
      <section className="mx-auto max-w-6xl px-4 py-8 sm:px-6 sm:py-12">
        <div className="mb-8 max-w-2xl">
          <h1 className="text-2xl font-bold tracking-tight text-[var(--foreground)] sm:text-3xl">
            Смены на сегодня и завтра
          </h1>
          <p className="mt-2 text-base leading-relaxed text-[var(--muted)] sm:text-lg">
            Заглушка: здесь будут реальные объявления от владельцев ПВЗ. Пока —
            демо-карточки.
          </p>
        </div>
        <ShiftList />
      </section>

      <footer className="border-t border-[var(--border)] py-6 text-center text-sm text-[var(--muted)]">
        © {new Date().getFullYear()} Беру Смену · berusmenu.ru
      </footer>
    </div>
  );
}
