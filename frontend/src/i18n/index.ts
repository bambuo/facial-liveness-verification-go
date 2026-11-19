import i18n from 'i18next'
import { initReactI18next } from 'react-i18next'
import {
  DEFAULT_LOCALE,
  LOCALE_COOKIE_NAME,
  LOCALE_STORAGE_KEY,
  SUPPORTED_LOCALES,
  isSupportedLocale,
  normalizeLocale,
  type SupportedLocale,
} from './locales'
import en from './locales/en.json'

type Resource = Record<string, unknown>

const loaders: Partial<Record<SupportedLocale, () => Promise<{ default: Resource }>>> = {
  'zh-CN': () => import('./locales/zh-CN.json'),
}

let initialized = false

function readCookie(name: string): string | null {
  const escaped = name.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
  const match = document.cookie.match(new RegExp(`(?:^|; )${escaped}=([^;]*)`))
  return match ? decodeURIComponent(match[1]) : null
}

function writeCookie(name: string, value: string): void {
  const secure = window.location.protocol === 'https:' ? '; Secure' : ''
  document.cookie = `${name}=${encodeURIComponent(value)}; Path=/; Max-Age=31536000; SameSite=Lax${secure}`
}

function matchBrowserLocale(): SupportedLocale {
  const languages = navigator.languages?.length ? navigator.languages : [navigator.language]
  for (const tag of languages) {
    if (typeof tag !== 'string') continue
    const primary = tag.toLowerCase()
    if (primary === 'zh' || primary.startsWith('zh-')) return 'zh-CN'
    if (primary === 'en' || primary.startsWith('en-')) return 'en'
  }
  return DEFAULT_LOCALE
}

function resolveInitialLocale(): SupportedLocale {
  const params = new URLSearchParams(window.location.search)
  const query = params.get('lang')
  if (query) return normalizeLocale(query)

  const stored = localStorage.getItem(LOCALE_STORAGE_KEY)
  if (stored && isSupportedLocale(stored)) return stored

  const cookie = readCookie(LOCALE_COOKIE_NAME)
  if (cookie && isSupportedLocale(cookie)) return cookie

  return matchBrowserLocale()
}

function persistLocale(locale: SupportedLocale): void {
  localStorage.setItem(LOCALE_STORAGE_KEY, locale)
  writeCookie(LOCALE_COOKIE_NAME, locale)
}

function applyDocumentLocale(locale: SupportedLocale): void {
  document.documentElement.lang = locale
}

function removeLangQuery(): void {
  const url = new URL(window.location.href)
  if (!url.searchParams.has('lang')) return
  url.searchParams.delete('lang')
  const next = `${url.pathname}${url.search}${url.hash}`
  window.history.replaceState({}, '', next)
}

async function ensureResource(locale: SupportedLocale): Promise<void> {
  if (locale === 'en' || i18n.hasResourceBundle(locale, 'translation')) return
  const loader = loaders[locale]
  if (!loader) return
  const mod = await loader()
  i18n.addResourceBundle(locale, 'translation', mod.default, true, true)
}

export async function initI18n(): Promise<void> {
  if (initialized) return

  const initial = resolveInitialLocale()
  const resources: Record<string, { translation: Resource }> = {
    en: { translation: en as Resource },
  }

  if (initial === 'zh-CN') {
    const mod = await loaders['zh-CN']!()
    resources['zh-CN'] = { translation: mod.default }
  }

  await i18n.use(initReactI18next).init({
    lng: initial,
    fallbackLng: DEFAULT_LOCALE,
    supportedLngs: [...SUPPORTED_LOCALES],
    load: 'currentOnly',
    resources,
    defaultNS: 'translation',
    interpolation: { escapeValue: false },
    returnEmptyString: false,
    react: { useSuspense: false },
  })

  persistLocale(initial)
  applyDocumentLocale(initial)
  removeLangQuery()
  initialized = true
}

export async function setLocale(locale: string): Promise<void> {
  const next = normalizeLocale(locale)
  await ensureResource(next)
  await i18n.changeLanguage(next)
  persistLocale(next)
  applyDocumentLocale(next)
}

export function getLocale(): SupportedLocale {
  return normalizeLocale(i18n.resolvedLanguage || i18n.language)
}

export { useTranslation } from 'react-i18next'
export {
  DEFAULT_LOCALE,
  LOCALE_COOKIE_NAME,
  LOCALE_STORAGE_KEY,
  SUPPORTED_LOCALES,
  isSupportedLocale,
  localeLabel,
  normalizeLocale,
  type SupportedLocale,
} from './locales'
export default i18n
