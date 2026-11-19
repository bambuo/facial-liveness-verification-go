import { useState, useEffect, useRef } from 'react';
import type { ReactNode } from 'react';
/**
 * Shared layout for all /docs/* pages.
 * Provides: header bar, sidebar with page-level + section-level nav, main content area.
 */
import { Link, useLocation, useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { MagnifyingGlassIcon } from '@heroicons/react/24/outline';
import { C, injectFonts } from '../../theme';
import { Pill } from './shared';
import '../../styles/patterns.css';
import { Button, Input } from 'antd'
import type { InputRef } from 'antd'

// Theme-aware token alias for the docs shell.
const T = {
  ...C,
  bg: 'var(--paper)',
  panel: 'var(--panel)',
  sidebar: 'var(--panel)',
  surface: 'var(--panel-2)',
  surfaceHover: 'var(--panel-3)',
  border: 'var(--rule)',
  borderStrong: 'var(--rule-strong)',
  text: 'var(--ink)',
  muted: 'var(--mid)',
  dim: 'var(--soft)',
  cyan: 'var(--accent-ink)',
  cyanDim: 'var(--accent-soft)',
  cyanBorder: 'color-mix(in oklab, var(--accent) 30%, transparent)',
  accent: 'var(--accent)',
  accentInk: 'var(--accent-ink)',
  accentSoft: 'var(--accent-soft)',
  code: 'var(--accent-ink)',
  codeBg: 'var(--paper-2)',
}

// ─── Page-level navigation (shown at top of sidebar on every docs page) ──────

export interface DocPage {
  path: string;
  label: string;
  labelKey: string;
}

export const DOC_PAGES: DocPage[] = [
  { path: '/docs',           label: 'Getting Started', labelKey: 'docs.nav.gettingStarted' },
  { path: '/docs/guides',    label: 'Guides', labelKey: 'docs.nav.guides' },
  { path: '/docs/sdk',       label: 'SDK & Embed', labelKey: 'docs.nav.sdk' },
  { path: '/docs/features',  label: 'Features', labelKey: 'docs.nav.features' },
  { path: '/docs/reference', label: 'Reference', labelKey: 'docs.nav.reference' },
  { path: '/docs/review',    label: 'Review Dashboard', labelKey: 'docs.nav.review' },
];

// ─── Search index (all sections across all docs pages) ───────────────────────

interface SearchItem {
  page: string;
  path: string;
  section: string;
  id: string;
}

const SEARCH_INDEX: SearchItem[] = [
  { page: 'Getting Started', path: '/docs', section: 'Quick Start', id: 'quick-start' },
  { page: 'Getting Started', path: '/docs', section: 'Authentication', id: 'auth' },
  { page: 'Getting Started', path: '/docs', section: 'Verification Flow', id: 'flow' },
  { page: 'Getting Started', path: '/docs', section: '1 · Start Session', id: 'step-1' },
  { page: 'Getting Started', path: '/docs', section: '2 · Upload Front', id: 'step-2' },
  { page: 'Getting Started', path: '/docs', section: '3 · Upload Back', id: 'step-3' },
  { page: 'Getting Started', path: '/docs', section: '4 · Live Capture', id: 'step-4' },
  { page: 'Getting Started', path: '/docs', section: '5 · Get Results', id: 'step-5' },
  { page: 'Getting Started', path: '/docs', section: 'Cross-Validation', id: 'selfie' },
  { page: 'Guides', path: '/docs/guides', section: 'Integration Options', id: 'integration' },
  { page: 'Guides', path: '/docs/guides', section: 'Guides & Tutorials', id: 'guides' },
  { page: 'Guides', path: '/docs/guides', section: 'End-to-End Tutorial', id: 'guide-e2e' },
  { page: 'Guides', path: '/docs/guides', section: 'Mobile Handoff', id: 'guide-mobile' },
  { page: 'Guides', path: '/docs/guides', section: 'Building Custom UI', id: 'guide-custom-ui' },
  { page: 'Guides', path: '/docs/guides', section: 'Self-Hosting', id: 'self-hosting' },
  { page: 'Guides', path: '/docs/guides', section: 'Prerequisites', id: 'sh-prerequisites' },
  { page: 'Guides', path: '/docs/guides', section: 'Install', id: 'sh-install' },
  { page: 'Guides', path: '/docs/guides', section: 'External Database', id: 'sh-external-db' },
  { page: 'Guides', path: '/docs/guides', section: 'Useful Commands', id: 'sh-commands' },
  { page: 'SDK & Embed', path: '/docs/sdk', section: 'JavaScript SDK', id: 'sdk' },
  { page: 'SDK & Embed', path: '/docs/sdk', section: 'Embed Component', id: 'embed' },
  { page: 'Features', path: '/docs/features', section: 'Analysis Engine', id: 'analysis' },
  { page: 'Features', path: '/docs/features', section: 'Statuses', id: 'statuses' },
  { page: 'Features', path: '/docs/features', section: 'Batch API', id: 'batch' },
  { page: 'Features', path: '/docs/features', section: 'Address Verification', id: 'address' },
  { page: 'Features', path: '/docs/features', section: 'AML / Sanctions', id: 'aml' },
  { page: 'Features', path: '/docs/features', section: 'Age Estimation', id: 'age-estimation' },
  { page: 'Features', path: '/docs/features', section: 'Velocity Checks', id: 'velocity' },
  { page: 'Features', path: '/docs/features', section: 'IP Geolocation', id: 'ip-geolocation' },
  { page: 'Features', path: '/docs/features', section: 'Voice Auth', id: 'voice-auth' },
  { page: 'Features', path: '/docs/features', section: 'Compliance Rules', id: 'compliance' },
  { page: 'Features', path: '/docs/features', section: 'Monitoring', id: 'monitoring' },
  { page: 'Reference', path: '/docs/reference', section: 'Rate Limits', id: 'rate-limits' },
  { page: 'Reference', path: '/docs/reference', section: 'Changelog', id: 'changelog' },
  { page: 'Reference', path: '/docs/reference', section: 'Support', id: 'support' },
  { page: 'Review Dashboard', path: '/docs/review', section: 'Overview', id: 'overview' },
  { page: 'Review Dashboard', path: '/docs/review', section: 'Getting Access', id: 'access' },
  { page: 'Review Dashboard', path: '/docs/review', section: 'Roles', id: 'roles' },
  { page: 'Review Dashboard', path: '/docs/review', section: 'Invite Team Members', id: 'reviewers' },
  { page: 'Review Dashboard', path: '/docs/review', section: 'OTP Login', id: 'reviewer-login' },
  { page: 'Review Dashboard', path: '/docs/review', section: 'The Dashboard', id: 'dashboard' },
  { page: 'Review Dashboard', path: '/docs/review', section: 'Stats Bar', id: 'stats' },
  { page: 'Review Dashboard', path: '/docs/review', section: 'Verification Table', id: 'table' },
  { page: 'Review Dashboard', path: '/docs/review', section: 'Filters & Search', id: 'filters' },
  { page: 'Review Dashboard', path: '/docs/review', section: 'Detail Panel', id: 'detail' },
  { page: 'Review Dashboard', path: '/docs/review', section: 'Review Actions', id: 'actions' },
  { page: 'Review Dashboard', path: '/docs/review', section: 'Approve', id: 'approve' },
  { page: 'Review Dashboard', path: '/docs/review', section: 'Reject', id: 'reject' },
  { page: 'Review Dashboard', path: '/docs/review', section: 'Override', id: 'override' },
  { page: 'Review Dashboard', path: '/docs/review', section: 'Status Reference', id: 'statuses' },
  { page: 'Review Dashboard', path: '/docs/review', section: 'Webhooks', id: 'webhooks' },
];

// ─── Section nav item (within a page) ────────────────────────────────────────

export interface NavItem {
  id: string;
  label: string;
  depth: number; // 0 = section, 1 = subsection
}

// ─── Layout component ────────────────────────────────────────────────────────

interface DocLayoutProps {
  /** Slug shown in header: "identity-verification / {slug}" */
  slug: string;
  /** Section nav items for the current page */
  nav: NavItem[];
  children: ReactNode;
}

export const DocLayout = ({ slug, nav, children }: DocLayoutProps) => {
  const { t } = useTranslation();
  const location = useLocation();
  const navigate = useNavigate();
  const [active, setActive] = useState(nav[0]?.id ?? '');
  const [searchQuery, setSearchQuery] = useState('');
  const [searchOpen, setSearchOpen] = useState(false);
  const searchRef = useRef<HTMLDivElement>(null);
  const searchInputRef = useRef<InputRef>(null);

  useEffect(() => { injectFonts(); }, []);

  // Scroll spy
  useEffect(() => {
    const ids = nav.map(n => n.id);
    const obs = new IntersectionObserver(
      entries => entries.forEach(e => { if (e.isIntersecting) setActive(e.target.id); }),
      { rootMargin: '-80px 0px -55% 0px' }
    );
    ids.forEach(id => { const el = document.getElementById(id); if (el) obs.observe(el); });
    return () => obs.disconnect();
  }, [nav]);

  // Hash-based scrolling for cross-page search navigation
  useEffect(() => {
    if (location.hash) {
      const id = location.hash.slice(1);
      setTimeout(() => {
        document.getElementById(id)?.scrollIntoView({ behavior: 'smooth', block: 'start' });
      }, 150);
    }
  }, [location.hash, location.pathname]);

  // Cmd+K / Ctrl+K to focus search, Escape to close
  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
        e.preventDefault();
        searchInputRef.current?.focus();
        setSearchOpen(true);
      }
      if (e.key === 'Escape') {
        setSearchOpen(false);
        searchInputRef.current?.blur();
      }
    };
    document.addEventListener('keydown', handler);
    return () => document.removeEventListener('keydown', handler);
  }, []);

  // Click outside to close search dropdown
  useEffect(() => {
    if (!searchOpen) return;
    const handler = (e: MouseEvent) => {
      if (searchRef.current && !searchRef.current.contains(e.target as Node)) {
        setSearchOpen(false);
      }
    };
    document.addEventListener('mousedown', handler);
    return () => document.removeEventListener('mousedown', handler);
  }, [searchOpen]);

  // Search filtering and grouping
  const searchResults = searchQuery.trim().length > 0
    ? SEARCH_INDEX.filter(item =>
        item.section.toLowerCase().includes(searchQuery.toLowerCase()) ||
        item.page.toLowerCase().includes(searchQuery.toLowerCase())
      ).slice(0, 12)
    : [];

  const groupedResults = searchResults.reduce<Record<string, { page: string; items: SearchItem[] }>>((acc, item) => {
    if (!acc[item.path]) acc[item.path] = { page: item.page, items: [] };
    acc[item.path].items.push(item);
    return acc;
  }, {});

  const handleSearchSelect = (item: SearchItem) => {
    setSearchQuery('');
    setSearchOpen(false);
    if (location.pathname === item.path) {
      document.getElementById(item.id)?.scrollIntoView({ behavior: 'smooth', block: 'start' });
    } else {
      navigate(`${item.path}#${item.id}`);
    }
  };

  const scrollTo = (id: string) => document.getElementById(id)?.scrollIntoView({ behavior: 'smooth', block: 'start' });

  const isCurrentPage = (path: string) => location.pathname === path;

  return (
    <div className="pattern-wave pattern-faint pattern-fade-edges pattern-full" style={{ fontFamily: T.sans, background: T.bg, color: T.text, margin: '0 -24px 0', minHeight: '100vh', overflowX: 'hidden' }}>

      {/* ── Header bar ── */}
      <div style={{ borderBottom: `1px solid ${T.border}`, padding: '18px 32px', display: 'flex', alignItems: 'center', gap: 16, position: 'sticky', top: 0, background: T.bg, zIndex: 10 }}>
        <div>
          <span style={{ fontFamily: T.mono, fontSize: '1.05rem', fontWeight: 600, color: T.text }}>
            <span style={{ color: T.cyan }}>identity-verification</span>
            <span style={{ color: T.dim }}> / </span>
            <span>{slug}</span>
          </span>
        </div>
        {/* Search bar */}
        <div ref={searchRef} style={{ position: 'relative', flex: '0 1 280px' }}>
          <div style={{ position: 'relative' }}>
            <MagnifyingGlassIcon style={{ position: 'absolute', left: 10, top: '50%', transform: 'translateY(-50%)', width: 14, height: 14, color: T.dim, pointerEvents: 'none' }} />
            <Input
              ref={searchInputRef}
              type="text"
              value={searchQuery}
              onChange={e => { setSearchQuery(e.target.value); setSearchOpen(true); }}
              onFocus={() => { if (searchQuery.trim()) setSearchOpen(true); }}
              placeholder="Search docs..."
              style={{
                width: '100%', padding: '6px 48px 6px 30px',
                background: T.surface, border: `1px solid ${T.border}`, borderRadius: 0,
                fontFamily: T.mono, fontSize: '0.75rem', color: T.text,
                outline: 'none',
              }}
            />
            <span style={{ position: 'absolute', right: 10, top: '50%', transform: 'translateY(-50%)', fontFamily: T.mono, fontSize: '0.6rem', color: T.dim, pointerEvents: 'none' }}>
              {typeof navigator !== 'undefined' && /Mac|iPhone|iPad/.test(navigator.userAgent) ? '⌘K' : 'Ctrl+K'}
            </span>
          </div>

          {/* Search results dropdown */}
          {searchOpen && searchResults.length > 0 && (
            <div style={{
              position: 'absolute', top: '100%', left: 0, right: 0, marginTop: 4,
              background: T.panel, border: `1px solid ${T.border}`,
              borderRadius: 0, maxHeight: 320, overflowY: 'auto', zIndex: 50,
              boxShadow: '0 8px 32px rgba(0,0,0,0.4)',
            }}>
              {Object.entries(groupedResults).map(([path, group]) => (
                <div key={path}>
                  <div style={{ fontFamily: T.mono, fontSize: '0.6rem', color: T.dim, letterSpacing: '0.1em', textTransform: 'uppercase', padding: '8px 12px 4px' }}>
                    {group.page}
                  </div>
                  {group.items.map(item => (
                    <Button
                      key={`${item.path}#${item.id}`}
                      onClick={() => handleSearchSelect(item)}
                      style={{
                        display: 'block', width: '100%', textAlign: 'left',
                        padding: '6px 12px 6px 20px',
                        background: 'transparent', border: 'none', cursor: 'pointer',
                        fontFamily: T.sans, fontSize: '0.8rem', color: T.muted,
                        transition: 'background 0.1s, color 0.1s',
                      }}
                      onMouseEnter={e => { e.currentTarget.style.background = T.surfaceHover; e.currentTarget.style.color = T.cyan; }}
                      onMouseLeave={e => { e.currentTarget.style.background = 'transparent'; e.currentTarget.style.color = T.muted; }}
                    >
                      {item.section}
                    </Button>
                  ))}
                </div>
              ))}
            </div>
          )}
        </div>

        <div style={{ marginLeft: 'auto', display: 'flex', gap: 8, alignItems: 'center' }}>
          <Link
            to="/docs/markdown"
            style={{
              fontFamily: T.mono, fontSize: '0.72rem', fontWeight: 500,
              color: T.muted, textDecoration: 'none',
              padding: '4px 10px', borderRadius: 0,
              border: `1px solid ${T.border}`,
              transition: 'color 0.15s, border-color 0.15s',
            }}
            onMouseEnter={e => { e.currentTarget.style.color = T.cyan; e.currentTarget.style.borderColor = T.cyan; }}
            onMouseLeave={e => { e.currentTarget.style.color = T.muted; e.currentTarget.style.borderColor = T.border; }}
          >.md</Link>
          <Pill color={T.green} bg={T.greenDim}>v1.7.0</Pill>
          <Pill color={T.muted} bg="rgba(74,85,104,0.13)">March 2026</Pill>
        </div>
      </div>

      {/* ── Two-column layout ── */}
      <div style={{ display: 'flex', maxWidth: 1440, margin: '0 auto' }}>

        {/* Sidebar */}
        <aside className="hidden lg:block" style={{ width: 230, flexShrink: 0, position: 'sticky', top: 57, height: 'calc(100vh - 57px)', overflowY: 'auto', borderRight: `1px solid ${T.border}`, padding: '20px 0', background: T.sidebar }}>

          {/* Page-level nav */}
          <div style={{ fontFamily: T.mono, fontSize: '0.62rem', color: T.dim, letterSpacing: '0.12em', textTransform: 'uppercase', padding: '0 20px 8px' }}>{t('docs.pages')}</div>
          {DOC_PAGES.map(page => (
            <Link
              key={page.path}
              to={page.path}
              style={{
                display: 'block', width: '100%', textAlign: 'left',
                padding: '5px 20px',
                fontFamily: T.sans, fontSize: '0.8rem', fontWeight: 600,
                color: isCurrentPage(page.path) ? T.cyan : T.muted,
                background: isCurrentPage(page.path) ? T.cyanDim : 'transparent',
                borderLeft: isCurrentPage(page.path) ? `2px solid ${T.cyan}` : '2px solid transparent',
                textDecoration: 'none',
                transition: 'all 0.15s',
              }}
            >
              {t(page.labelKey)}
            </Link>
          ))}

          {/* Section-level nav (within current page) */}
          {nav.length > 0 && (
            <>
              <div style={{ height: 1, background: T.border, margin: '12px 20px' }} />
              <div style={{ fontFamily: T.mono, fontSize: '0.62rem', color: T.dim, letterSpacing: '0.12em', textTransform: 'uppercase', padding: '0 20px 8px' }}>{t('docs.onThisPage')}</div>
              {nav.map(item => (
                <Button
                  key={item.id}
                  onClick={() => scrollTo(item.id)}
                  style={{
                    display: 'block', width: '100%', textAlign: 'left',
                    padding: item.depth === 0 ? '5px 20px' : '4px 20px 4px 34px',
                    fontFamily: T.sans,
                    fontSize: item.depth === 0 ? '0.8rem' : '0.75rem',
                    fontWeight: item.depth === 0 ? 600 : 400,
                    color: active === item.id ? T.cyan : item.depth === 0 ? T.text : T.muted,
                    background: active === item.id ? T.cyanDim : 'transparent',
                    borderLeft: active === item.id ? `2px solid ${T.cyan}` : '2px solid transparent',
                    border: 'none', cursor: 'pointer',
                    transition: 'all 0.15s',
                  }}
                >
                  {t(item.label)}
                </Button>
              ))}
            </>
          )}
        </aside>

        {/* Main content */}
        <main style={{ flex: 1, padding: '72px 52px', maxWidth: 980, minWidth: 0 }}>
          {children}

          {/* Footer */}
          <div style={{ marginTop: 64, paddingTop: 24, borderTop: `1px solid ${T.border}`, fontFamily: T.mono, fontSize: '0.7rem', color: T.dim }}>
            &copy; 2026 Identity Verification &mdash; Open source under MIT License
          </div>
        </main>
      </div>
    </div>
  );
};
