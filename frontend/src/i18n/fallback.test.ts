import { describe, expect, it } from 'vitest'
import i18next from 'i18next'
import en from './locales/en.json'
import zhCN from './locales/zh-CN.json'

describe('frontend i18n fallback', () => {
  it('returns the English value when a Chinese key is missing', async () => {
    const zhMissing = JSON.parse(JSON.stringify(zhCN)) as typeof zhCN
    delete (zhMissing.setup as Record<string, unknown>).checkingStatus

    const instance = i18next.createInstance()
    await instance.init({
      lng: 'zh-CN',
      fallbackLng: 'en',
      resources: {
        en: { translation: en },
        'zh-CN': { translation: zhMissing },
      },
      interpolation: { escapeValue: false },
    })

    expect(instance.t('setup.checkingStatus')).toBe(en.setup.checkingStatus)
  })
})
