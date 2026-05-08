/** Мобильный РФ в E.164: +7 и 10 цифр, первая из них 9. */
export const RU_MOBILE_E164 = /^\+7[9]\d{9}$/;

const DIGITS_ONLY = /\D/g;

/**
 * Приводит ввод к виду +7XXXXXXXXXX (ровно 12 символов при успехе).
 * Принимает +7…, 8…, 9… (10 цифр) и строки только с цифрами.
 */
export function normalizeRussianMobile(input: string): string {
  let digits = input.replace(DIGITS_ONLY, "");
  if (digits.length === 11 && digits.startsWith("8")) {
    digits = "7" + digits.slice(1);
  }
  if (digits.length === 11 && digits.startsWith("7")) {
    return `+${digits}`;
  }
  if (digits.length === 10 && digits.startsWith("9")) {
    return `+7${digits}`;
  }
  const trimmed = input.trim();
  if (trimmed.startsWith("+7") && /^\+7\d+$/.test(trimmed)) {
    return `+7${trimmed.slice(2).replace(DIGITS_ONLY, "")}`;
  }
  return trimmed;
}

export function formatRuMobileDisplay(e164: string): string {
  const normalized = normalizeRussianMobile(e164);
  const m = normalized.match(/^\+7(\d{3})(\d{3})(\d{2})(\d{2})$/);
  if (!m) {
    return e164;
  }
  return `+7 (${m[1]}) ${m[2]}-${m[3]}-${m[4]}`;
}
