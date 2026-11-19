import { ReactNode } from 'react'
import { Link, useLocation } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { LanguageSwitcher } from '../common/LanguageSwitcher'

interface LayoutProps {
  children: ReactNode
}

const getGitHubUrl = () => 'https://github.com/example/identity-verification'

export function Layout({ children }: LayoutProps) {
  const location = useLocation()
  const { t } = useTranslation()

  const isAdminRoute = location.pathname.startsWith('/admin')
  const isStandaloneRoute =
    location.pathname.startsWith('/user-verification') ||
    location.pathname.startsWith('/verify/mobile') ||
    location.pathname.startsWith('/v/') ||
    location.pathname === '/developer/page-builder'

  // Admin consoles and Page Builder render the switcher inside their own toolbars.
  const hasInlineLanguage =
    location.pathname === '/developer/page-builder' ||
    location.pathname.startsWith('/admin/verifications') ||
    location.pathname.startsWith('/admin/developers')

  const floatingLanguage = hasInlineLanguage ? null : (
    <div style={{ position: 'fixed', top: 12, right: 12, zIndex: 1000 }}>
      <LanguageSwitcher />
    </div>
  )

  if (isAdminRoute) {
    return <>{floatingLanguage}{children}</>
  }

  if (isStandaloneRoute) {
    return <>{floatingLanguage}{children}</>
  }

  // ─────────────────────────────────────────
  // Community edition: no navbar, minimal footer (v2 styling)
  // ─────────────────────────────────────────
  const pathname = location.pathname
  const showBackNav = pathname !== '/' && pathname !== '/developer' && pathname !== '/setup'

  return (
    <div className="min-h-screen flex flex-col" style={{ background: 'var(--paper)', color: 'var(--ink)' }}>
      <main className="flex-1">
        {showBackNav && (
          <div style={{ borderBottom: '1px solid var(--rule)', padding: '10px 24px' }}>
            <Link to="/" className="mono" style={{ color: 'var(--mid)', fontSize: 12.5, textDecoration: 'none', display: 'inline-flex', alignItems: 'center', gap: 6 }}>
              &#8592; {t('layout.backToDevPortal')}
            </Link>
          </div>
        )}
        {children}
      </main>

      <footer style={{ background: 'var(--panel)', borderTop: '1px solid var(--rule)' }}>
        <div className="max-w-7xl mx-auto py-8 px-4 sm:px-6 lg:px-8">
          <div className="flex flex-col sm:flex-row items-center justify-between gap-4">
            <a href="https://example.com" target="_blank" rel="noopener noreferrer"
              className="flex items-center gap-3 hover:opacity-80 transition-opacity">
              <img src="/identity-logo.svg" alt="Identity Verification" className="h-6 w-auto" />
              <span className="mono" style={{ color: 'var(--soft)', fontSize: 12 }}>{t('layout.poweredBy')}</span>
            </a>
            <div className="flex items-center gap-6">
              {[
                { label: t('layout.devPortal'), href: '/' },
                { label: t('layout.docs'), href: '/docs' },
                { label: t('layout.demo'), href: '/demo' },
                { label: t('layout.verifyCredential'), href: '/verify-credential' },
              ].map(({ label, href }) => (
                <Link key={href} to={href}
                  className="mono hover:opacity-100 transition-opacity"
                  style={{ color: 'var(--mid)', fontSize: 12 }}>{label}</Link>
              ))}
              <LanguageSwitcher />
              <a href={getGitHubUrl()} target="_blank" rel="noopener noreferrer"
                className="mono hover:opacity-100 transition-opacity"
                style={{ color: 'var(--mid)', fontSize: 12 }}>{t('layout.github')}</a>
            </div>
          </div>
        </div>
      </footer>
    </div>
  )
}
