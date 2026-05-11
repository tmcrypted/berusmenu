import { cn } from "@/lib/cn";

type Props = {
  name?: string | null;
  src?: string | null;
  size?: "sm" | "md" | "lg";
};

const sizeClasses = {
  sm: "h-10 w-10 text-sm",
  md: "h-14 w-14 text-lg",
  lg: "h-24 w-24 text-3xl",
};

export function UserAvatar({ name, src, size = "md" }: Props) {
  const initial = name?.trim()?.[0]?.toUpperCase() || "?";

  if (src?.trim()) {
    return (
      <img
        src={src}
        alt={name?.trim() ? `Аватар ${name}` : "Аватар пользователя"}
        className={cn("shrink-0 rounded-full border border-[var(--border)] object-cover", sizeClasses[size])}
      />
    );
  }

  return (
    <div className={cn("flex shrink-0 items-center justify-center rounded-full border border-[var(--border)] bg-[var(--accent-soft)] font-semibold text-[var(--accent)]", sizeClasses[size])}>
      {initial}
    </div>
  );
}
