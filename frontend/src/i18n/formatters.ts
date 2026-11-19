import i18n from './index'

export function currentLocale(): string {
  return i18n.resolvedLanguage || i18n.language || 'en'
}

export function formatDate(
  value: string | number | Date,
  options: Intl.DateTimeFormatOptions = { year: 'numeric', month: 'short', day: '2-digit' },
): string {
  const date = value instanceof Date ? value : new Date(value)
  if (Number.isNaN(date.getTime())) return ''
  return new Intl.DateTimeFormat(currentLocale(), options).format(date)
}

export function formatDateTime(
  value: string | number | Date,
  options: Intl.DateTimeFormatOptions = {
    year: 'numeric',
    month: 'short',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  },
): string {
  const date = value instanceof Date ? value : new Date(value)
  if (Number.isNaN(date.getTime())) return ''
  return new Intl.DateTimeFormat(currentLocale(), options).format(date)
}

export function formatNumber(value: number, options?: Intl.NumberFormatOptions): string {
  return new Intl.NumberFormat(currentLocale(), options).format(value)
}

export function formatTime(
  value: string | number | Date,
  options: Intl.DateTimeFormatOptions = { hour: '2-digit', minute: '2-digit' },
): string {
  const date = value instanceof Date ? value : new Date(value)
  if (Number.isNaN(date.getTime())) return ''
  return new Intl.DateTimeFormat(currentLocale(), options).format(date)
}

export function formatRelative(value: number, unit: Intl.RelativeTimeFormatUnit): string {
  return new Intl.RelativeTimeFormat(currentLocale(), { numeric: 'auto' }).format(value, unit)
}
