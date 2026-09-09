// Time is stored and transported as UTC (timestamptz). It is displayed in the
// timezone that owns the datum — the schedule's or the user's IANA zone — never
// the browser's (rancangan Bab 4.5). These helpers are the presentation edge.

/** Fallback zone when a record carries none. */
export const DEFAULT_TZ = 'Asia/Jakarta';

/** WIB/WITA/WIT label for the Indonesian zones; falls back to the zone name. */
export function zoneLabel(tz: string = DEFAULT_TZ): string {
  switch (tz) {
    case 'Asia/Jakarta':
      return 'WIB';
    case 'Asia/Makassar':
      return 'WITA';
    case 'Asia/Jayapura':
      return 'WIT';
    default:
      return tz;
  }
}

/** `07:00` in the given zone. */
export function formatTime(value: string | Date, tz: string = DEFAULT_TZ): string {
  return format(value, tz, { hour: '2-digit', minute: '2-digit', hour12: false });
}

/** `07:00 WIB` — the form most schedule/attendance screens want. */
export function formatTimeWithZone(value: string | Date, tz: string = DEFAULT_TZ): string {
  return `${formatTime(value, tz)} ${zoneLabel(tz)}`;
}

/** `09 Sep 2026` in the given zone. */
export function formatDate(value: string | Date, tz: string = DEFAULT_TZ, locale = 'id-ID'): string {
  return format(value, tz, { day: '2-digit', month: 'short', year: 'numeric' }, locale);
}

/** `09 Sep 2026, 07:00 WIB`. */
export function formatDateTime(value: string | Date, tz: string = DEFAULT_TZ, locale = 'id-ID'): string {
  return `${formatDate(value, tz, locale)}, ${formatTimeWithZone(value, tz)}`;
}

/**
 * Intl formatting pinned to an IANA zone, falling back to the default zone when
 * the name is unknown so a bad `timezone` column never blanks the screen.
 */
function format(
  value: string | Date,
  tz: string,
  options: Intl.DateTimeFormatOptions,
  locale = 'id-ID',
): string {
  const date = value instanceof Date ? value : new Date(value);
  if (Number.isNaN(date.getTime())) return '';
  try {
    return new Intl.DateTimeFormat(locale, { ...options, timeZone: tz }).format(date);
  } catch {
    return new Intl.DateTimeFormat(locale, { ...options, timeZone: DEFAULT_TZ }).format(date);
  }
}
