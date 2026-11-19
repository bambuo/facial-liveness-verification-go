import { useEffect } from 'react';
import { Link } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { C, injectFonts } from '../theme';

export const NotFoundPage = () => {
  const { t } = useTranslation();
  useEffect(() => { injectFonts(); }, []);

  return (
    <div style={{ background: 'var(--paper)', fontFamily: C.sans, color: 'var(--ink)', minHeight: '100vh', display: 'flex', alignItems: 'center', justifyContent: 'center', padding: '40px 24px' }}>
      <div style={{ textAlign: 'center', maxWidth: 480 }}>
        <div style={{ fontFamily: C.mono, fontSize: 'clamp(80px, 16vw, 128px)', fontWeight: 500, color: 'var(--rule)', lineHeight: 1, marginBottom: 8, userSelect: 'none', letterSpacing: '-0.04em' }}>
          404
        </div>
        <div className="eyebrow" style={{ marginBottom: 20 }}>
          {t('notFound.eyebrow')}
        </div>
        <p style={{ color: 'var(--mid)', fontSize: 15, lineHeight: 1.6, marginBottom: 36 }}>
          {t('notFound.description')}
        </p>
        <div style={{ display: 'flex', gap: 12, justifyContent: 'center', flexWrap: 'wrap', marginBottom: 32 }}>
          <Link
            to="/"
            className="btn"
            style={{ textDecoration: 'none' }}
          >
            {t('notFound.home')}
          </Link>
          <Link
            to="/docs"
            className="btn ghost"
            style={{ border: '1px solid var(--rule-strong)', textDecoration: 'none' }}
          >
            {t('notFound.docs')}
          </Link>
        </div>
        <div style={{ display: 'flex', gap: 24, justifyContent: 'center', flexWrap: 'wrap' }}>
          {[
            { label: t('notFound.developerPortal'), to: '/developer' },
            { label: t('notFound.demo'), to: '/demo' },
          ].map(({ label, to }) => (
            <Link key={to} to={to} style={{ color: 'var(--mid)', fontSize: 13, textDecoration: 'none', fontFamily: C.mono }}>
              {label}
            </Link>
          ))}
        </div>
      </div>
    </div>
  );
};
