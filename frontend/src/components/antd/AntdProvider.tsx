import { useEffect, useState, type ReactNode } from 'react'
import { App as AntApp, ConfigProvider, theme } from 'antd'
import enUS from 'antd/locale/en_US'
import zhCN from 'antd/locale/zh_CN'
import i18n, { getLocale } from '../../i18n'
import { getTheme } from '../../theme'

function useDocumentTheme() {
  const [value, setValue] = useState<'light' | 'dark'>(() => getTheme())
  useEffect(() => {
    const observer = new MutationObserver(() => setValue(getTheme()))
    observer.observe(document.documentElement, { attributes: true, attributeFilter: ['data-theme'] })
    return () => observer.disconnect()
  }, [])
  return value
}

function useAntdLocale() {
  const [value, setValue] = useState(() => getLocale())
  useEffect(() => {
    const onChange = () => setValue(getLocale())
    i18n.on('languageChanged', onChange)
    return () => i18n.off('languageChanged', onChange)
  }, [])
  return value === 'zh-CN' ? zhCN : enUS
}

/** antd 全局主题、语言与反馈容器。 */
export function AntdProvider({ children }: { children: ReactNode }) {
  const documentTheme = useDocumentTheme()
  const locale = useAntdLocale()
  return (
    <ConfigProvider
      locale={locale}
      theme={{
        algorithm: documentTheme === 'dark' ? theme.darkAlgorithm : theme.defaultAlgorithm,
        token: {
          colorPrimary: '#00d4d4',
          colorInfo: '#00d4d4',
          borderRadius: 4,
          fontFamily: 'inherit',
          colorBgBase: documentTheme === 'dark' ? '#0b0b0d' : '#f7f6f3',
        },
        components: {
          Layout: {
            bodyBg: documentTheme === 'dark' ? '#0b0b0d' : '#f7f6f3',
            headerBg: documentTheme === 'dark' ? '#111114' : '#ffffff',
            siderBg: documentTheme === 'dark' ? '#0b0b0d' : '#ffffff',
          },
          Menu: {
            itemBg: 'transparent',
          },
        },
      }}
    >
      <AntApp>{children}</AntApp>
    </ConfigProvider>
  )
}
