import { useTranslation } from 'react-i18next'
import { Select, Space, Typography } from 'antd'
import {
  SUPPORTED_LOCALES,
  getLocale,
  localeLabel,
  setLocale,
  type SupportedLocale,
} from '../../i18n'

export function LanguageSwitcher({ variant = 'default' }: { variant?: 'default' | 'toolbar' | 'console' }) {
  const { t } = useTranslation()
  const current = getLocale()
  const size = variant === 'default' ? 'small' : 'middle'
  const options = SUPPORTED_LOCALES.map((locale) => ({ value: locale, label: localeLabel(locale) }))
  return (
    <Space size={6}>
      {variant === 'default' && <Typography.Text type="secondary" style={{ fontSize: 11 }}>{t('common.language')}</Typography.Text>}
      <Select<SupportedLocale>
        aria-label={t('common.language')}
        size={size}
        value={current}
        options={options}
        style={{ minWidth: 112 }}
        onChange={(value) => void setLocale(value)}
      />
    </Space>
  )
}

export default LanguageSwitcher
