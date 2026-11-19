export const SUPPORTED_LOCALES = ['en', 'zh-CN'] as const

export type SupportedLocale = (typeof SUPPORTED_LOCALES)[number]

export const DEFAULT_LOCALE: SupportedLocale = 'en'
export const LOCALE_STORAGE_KEY = 'locale'
export const LOCALE_COOKIE_NAME = 'locale'

export function isSupportedLocale(value: unknown): value is SupportedLocale {
  return typeof value === 'string' && (SUPPORTED_LOCALES as readonly string[]).includes(value)
}

/**
 * Normalize an incoming BCP 47 tag to one of the supported locales.
 * Unknown or malformed values fall back to English.
 */
export function normalizeLocale(value?: string | null): SupportedLocale {
  if (!value) return DEFAULT_LOCALE

  const primary = value.trim().toLowerCase().replace(/_/g, '-').split(',')[0]?.split(';')[0]?.trim()
  if (!primary) return DEFAULT_LOCALE
  if (primary === 'zh' || primary.startsWith('zh-')) return 'zh-CN'
  if (primary === 'en' || primary.startsWith('en-')) return 'en'
  return DEFAULT_LOCALE
}

export function localeLabel(locale: SupportedLocale): string {
  return locale === 'zh-CN' ? '简体中文' : 'English'
}
