import { useState, useEffect, useCallback } from 'react'
import { createElement } from 'react';
import { formatDate as formatLocaleDate, formatNumber } from '../i18n/formatters'
import { useNavigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { API_BASE_URL } from '../config/api'
import { fetchCsrfToken, getCsrfToken, csrfHeader, clearCsrfToken } from '../lib/csrf'
import { C, injectFonts } from '../theme'
import '../styles/patterns.css'
import {
  ArrowPathIcon,
  ArrowLeftIcon,
  MagnifyingGlassIcon,
  ShieldCheckIcon,
  UsersIcon,
  KeyIcon,
  CheckBadgeIcon,
  ExclamationTriangleIcon,
  ComputerDesktopIcon,
} from '@heroicons/react/24/outline'
import { LanguageSwitcher } from '../components/common/LanguageSwitcher'
import { Alert, Button, Card, Col, Descriptions, Input, Row, Statistic, Table, Tag } from 'antd'
import type { ColumnsType } from 'antd/es/table'

// Theme-aware token alias: structural colors follow the active light/dark theme.
const T = {
  ...C,
  bg: 'var(--paper)',
  panel: 'var(--panel)',
  surface: 'var(--panel-2)',
  surfaceHover: 'var(--panel-3)',
  border: 'var(--rule)',
  borderStrong: 'var(--rule-strong)',
  text: 'var(--ink)',
  muted: 'var(--mid)',
  dim: 'var(--soft)',
  cyan: 'var(--accent)',
  cyanDim: 'var(--accent-soft)',
  cyanBorder: 'color-mix(in oklab, var(--accent) 30%, transparent)',
  accent: 'var(--accent)',
  accentInk: 'var(--accent-ink)',
  accentSoft: 'var(--accent-soft)',
  code: 'var(--accent-ink)',
  codeBg: 'var(--paper-2)',
}

// ─── Types ──────────────────────────────────────────────────

interface VerificationStats {
  total: number
  verified: number
  failed: number
  pending: number
  manual_review: number
}

interface WebhookStats {
  total_webhooks: number
  active_webhooks: number
  total_deliveries: number
  successful_deliveries: number
  failed_deliveries: number
  pending_deliveries: number
}

interface Developer {
  id: string
  email: string
  full_name: string | null
  company_name: string | null
  is_verified: boolean
  created_at: string
  webhook_url: string | null
  verification_stats: VerificationStats
  webhook_stats: WebhookStats
}

// ─── Helpers ────────────────────────────────────────────────

const formatDate = (iso: string) => {
  const d = new Date(iso)
  return formatLocaleDate(d, { month: 'short', day: 'numeric', year: 'numeric' })
}

const authFetchOpts = (): RequestInit => ({ credentials: 'include' })

// ─── Component ──────────────────────────────────────────────

export function DevelopersList() {
  const { t } = useTranslation()
  const navigate = useNavigate()

  useEffect(() => { injectFonts() }, [])

  // ── State ──
  const [developers, setDevelopers] = useState<Developer[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [searchQuery, setSearchQuery] = useState('')
  const [page, setPage] = useState(1)
  const [totalCount, setTotalCount] = useState(0)
  const [expandedId, setExpandedId] = useState<string | null>(null)
  const [authReady, setAuthReady] = useState(() => !!getCsrfToken())
  const [isMobile, setIsMobile] = useState(false)

  const LIMIT = 25

  // ── Mobile viewport detection ──
  useEffect(() => {
    const mq = window.matchMedia('(max-width: 768px)')
    setIsMobile(mq.matches)
    const handler = (e: MediaQueryListEvent) => setIsMobile(e.matches)
    mq.addEventListener('change', handler)
    return () => mq.removeEventListener('change', handler)
  }, [])

  // ── Auth guard (cookie-only, no escalation) ──
  useEffect(() => {
    fetch(`${API_BASE_URL}/api/admin/developers?limit=1`, { credentials: 'include' })
      .then(res => {
        if (res.ok) { setAuthReady(true); fetchCsrfToken(); return }
        clearCsrfToken()
        // Not a platform admin — redirect to verifications or login
        if (res.status === 401) navigate('/admin/login')
        else navigate('/admin/verifications')
      })
      .catch(() => { clearCsrfToken(); navigate('/admin/login') })
  }, [navigate])

  // ── Fetch developers ──
  const fetchDevelopers = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      const params = new URLSearchParams({ page: String(page), limit: String(LIMIT) })
      const res = await fetch(`${API_BASE_URL}/api/admin/developers?${params}`, authFetchOpts())
      if (res.status === 401) { navigate('/admin/login'); return }
      if (!res.ok) throw new Error(`HTTP ${res.status}`)
      const data = await res.json()
      setDevelopers(data.developers || [])
      setTotalCount(data.pagination?.total || 0)
    } catch (caught: unknown) {
      const err = caught as Error;
      setError(err.message || t('developersList.loadFailed'))
    } finally {
      setLoading(false)
    }
  }, [page, navigate])

  useEffect(() => { if (authReady) fetchDevelopers() }, [authReady, fetchDevelopers])

  // ── Derived stats ──
  const verifiedCount = developers.filter(d => d.is_verified).length
  const unverifiedCount = developers.filter(d => !d.is_verified).length
  const withKeysCount = developers.filter(d => (d.webhook_stats?.total_webhooks ?? 0) > 0 || d.verification_stats?.total > 0).length

  // ── Search filter (client-side on loaded page) ──
  const filtered = searchQuery
    ? developers.filter(d =>
        (d.full_name || '').toLowerCase().includes(searchQuery.toLowerCase()) ||
        d.email.toLowerCase().includes(searchQuery.toLowerCase()) ||
        (d.company_name || '').toLowerCase().includes(searchQuery.toLowerCase())
      )
    : developers

  const columns: ColumnsType<Developer> = [
    { title: 'Name', dataIndex: 'full_name', render: value => value || '—' },
    { title: 'Email', dataIndex: 'email' },
    { title: 'Company', dataIndex: 'company_name', render: value => value || '—' },
    {
      title: 'Verified', dataIndex: 'is_verified', width: 110,
      render: value => <Tag color={value ? 'success' : 'warning'}>{value ? t('developersList.yes') : 'No'}</Tag>,
    },
    {
      title: 'Verifications', width: 120,
      render: (_, developer) => <span style={{ fontFamily: T.mono }}>{developer.verification_stats?.total ?? 0}</span>,
    },
    { title: 'Signed Up', dataIndex: 'created_at', width: 140, render: value => formatDate(value) },
  ]

  const renderDeveloperDetail = (developer: Developer) => {
    const stats = developer.verification_stats
    return (
      <Row gutter={[16, 16]}>
        <Col xs={24} lg={12}>
          <Card size="small" title={t('developersList.verificationBreakdown')}>
            <Row gutter={[8, 8]}>
              <Col span={8}><Statistic title="Total" value={stats?.total ?? 0} /></Col>
              <Col span={8}><Statistic title="Verified" value={stats?.verified ?? 0} styles={{ content: { color: T.green } }} /></Col>
              <Col span={8}><Statistic title="Failed" value={stats?.failed ?? 0} styles={{ content: { color: T.red } }} /></Col>
              <Col span={8}><Statistic title="Review" value={stats?.manual_review ?? 0} styles={{ content: { color: T.amber } }} /></Col>
              <Col span={8}><Statistic title="Pending" value={stats?.pending ?? 0} styles={{ content: { color: T.cyan } }} /></Col>
            </Row>
          </Card>
        </Col>
        <Col xs={24} lg={12}>
          <Card size="small" title={t('developersList.developerDetails')}>
            <Descriptions
              size="small"
              column={1}
              items={[
                ['Developer ID', developer.id],
                ['Email', developer.email],
                ['Name', developer.full_name || '—'],
                ['Company', developer.company_name || '—'],
                ['Email Verified', developer.is_verified ? 'Yes' : 'No'],
                ['Webhook URL', developer.webhook_url || 'None'],
                ['Webhooks', `${developer.webhook_stats?.active_webhooks ?? 0} active / ${developer.webhook_stats?.total_webhooks ?? 0} total`],
                ['Signed Up', formatDate(developer.created_at)],
              ].map(([label, children]) => ({ key: String(label), label, children }))}
            />
          </Card>
        </Col>
      </Row>
    )
  }

  // ── Mobile guard ──
  if (isMobile) {
    return (
      <div style={{
        background: T.bg, minHeight: '100vh', fontFamily: T.sans,
        display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center',
        padding: '32px 24px', textAlign: 'center',
      }}>
        <div style={{
          background: T.panel, padding: '48px 32px',
          border: `1px solid ${T.border}`, maxWidth: 400,         }}>
          <ComputerDesktopIcon style={{ width: 48, height: 48, color: T.cyan, margin: '0 auto 20px' }} />
          <h2 style={{ color: T.text, fontSize: 20, fontWeight: 600, margin: '0 0 12px', fontFamily: T.sans }}>
            {t('developersList.desktopRequired')}
          </h2>
          <p style={{ color: T.muted, fontSize: 14, lineHeight: 1.6, margin: '0 0 24px' }}>
            {t('developersList.desktopRequiredDesc')}
          </p>
          <Button
            onClick={() => navigate(-1)}
            style={{
              background: 'none', border: `1px solid ${T.border}`,               color: T.text, padding: '10px 24px', cursor: 'pointer', fontFamily: T.sans,
              fontSize: 14, fontWeight: 500, transition: 'border-color 0.2s',
            }}
            onMouseEnter={e => (e.currentTarget.style.borderColor = T.cyan)}
            onMouseLeave={e => (e.currentTarget.style.borderColor = T.border)}
          >
            {t('developersList.goBack')}
          </Button>
        </div>
      </div>
    )
  }

  // ── Render ──
  return (
    <div className="pattern-crosshatch pattern-faint pattern-fade-edges pattern-full" style={{ background: T.bg, minHeight: '100vh', fontFamily: T.sans }}>
      <div style={{ maxWidth: 1400, margin: '0 auto', padding: '32px 24px' }}>

        {/* ── Header ── */}
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 32 }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 16 }}>
            <Button
              onClick={() => navigate('/admin/login')}
              style={{ background: 'none', border: 'none', cursor: 'pointer', color: T.muted, display: 'flex', padding: 4 }}
              title={t('developersList.back')}
            >
              <ArrowLeftIcon style={{ width: 20, height: 20 }} />
            </Button>
            <div>
              <div style={{ fontFamily: T.mono, fontSize: 11, color: T.muted, letterSpacing: '0.08em', marginBottom: 8 }}>
                {t('developersList.eyebrow')}
              </div>
              <h1 style={{ fontFamily: T.mono, color: T.text, fontSize: 24, fontWeight: 600, margin: 0 }}>
                {t('developersList.title')}
              </h1>
              <p style={{ color: T.dim, fontSize: 13, margin: '4px 0 0', fontFamily: T.mono }}>
                {t('developersList.subtitle')}
              </p>
            </div>
          </div>
          <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
            <LanguageSwitcher variant="console" />
            <Button
              onClick={() => navigate('/admin/verifications')}
              className="btn-secondary"
              style={{ display: 'flex', alignItems: 'center', gap: 6, padding: '8px 16px', fontSize: 13 }}
            >
              <ShieldCheckIcon style={{ width: 14, height: 14 }} />
              {t('developersList.verifications')}
            </Button>
            <Button
              onClick={() => fetchDevelopers()}
              className="btn-secondary"
              style={{ display: 'flex', alignItems: 'center', gap: 6, padding: '8px 16px', fontSize: 13 }}
            >
              <ArrowPathIcon style={{ width: 14, height: 14 }} />
              {t('developersList.refresh')}
            </Button>
            <Button
              onClick={() => {
                fetch(`${API_BASE_URL}/api/auth/logout`, { method: 'POST', credentials: 'include', headers: csrfHeader() }).catch(() => {})
                clearCsrfToken()
                navigate('/admin/login')
              }}
              className="btn-outline"
              style={{ padding: '8px 16px', fontSize: 13 }}
              onMouseEnter={e => { e.currentTarget.style.borderColor = T.red; e.currentTarget.style.color = T.red }}
              onMouseLeave={e => { e.currentTarget.style.borderColor = 'var(--rule)'; e.currentTarget.style.color = 'var(--ink)' }}
            >
              {t('developersList.signOut')}
            </Button>
          </div>
        </div>

        {/* ── Stats Cards ── */}
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(4, 1fr)', gap: 12, marginBottom: 28 }}>
          {[
            { label: 'Total Developers', value: totalCount, color: T.text, icon: UsersIcon },
            { label: 'Verified', value: verifiedCount, color: T.green, icon: CheckBadgeIcon },
            { label: 'Unverified', value: unverifiedCount, color: T.amber, icon: ExclamationTriangleIcon },
            { label: 'With Activity', value: withKeysCount, color: T.cyan, icon: KeyIcon },
          ].map(s => (
            <div key={s.label} style={{
              background: T.panel, border: `1px solid ${T.border}`, padding: '16px 20px',
                          }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 6 }}>
                {createElement(s.icon, { style: { width: 14, height: 14, color: s.color } })}
                <span style={{ color: T.dim, fontSize: 11, fontWeight: 600, textTransform: 'uppercase', letterSpacing: '0.08em' }}>
                  {s.label}
                </span>
              </div>
              <div style={{ color: s.color, fontSize: 28, fontWeight: 600, fontFamily: T.mono, lineHeight: 1 }}>
                {formatNumber(s.value)}
              </div>
            </div>
          ))}
        </div>

        {/* ── Search Bar ── */}
        <div style={{
          display: 'flex', alignItems: 'center', gap: 12, marginBottom: 16,
          background: T.panel, border: `1px solid ${T.border}`, padding: '12px 16px',         }}>
          <div style={{ position: 'relative', flex: 1, maxWidth: 400 }}>
            <MagnifyingGlassIcon style={{ width: 14, height: 14, color: T.dim, position: 'absolute', left: 10, top: '50%', transform: 'translateY(-50%)' }} />
            <Input
              type="text"
              placeholder={t('developersList.searchPlaceholder')}
              value={searchQuery}
              onChange={e => setSearchQuery(e.target.value)}
              style={{
                background: T.surface, border: `1px solid ${T.border}`,                 padding: '6px 10px 6px 30px', color: T.text, fontSize: 12, fontFamily: T.mono,
                width: '100%', outline: 'none',
              }}
              onFocus={e => (e.currentTarget.style.borderColor = T.cyanBorder)}
              onBlur={e => (e.currentTarget.style.borderColor = T.border)}
            />
          </div>
          <span style={{ color: T.dim, fontSize: 12, fontFamily: T.mono }}>
            {filtered.length} developer{filtered.length !== 1 ? 's' : ''}
          </span>
        </div>

        {error && (
          <Alert
            type="error"
            showIcon
            title={error}
            action={<Button size="small" onClick={fetchDevelopers}>{t('developersList.retry')}</Button>}
            style={{ marginBottom: 16 }}
          />
        )}

        <Table<Developer>
          rowKey="id"
          columns={columns}
          dataSource={filtered}
          loading={loading}
          scroll={{ x: 900 }}
          locale={{ emptyText: searchQuery ? t('developersList.noMatch') : t('developersList.none') }}
          pagination={{
            current: page,
            pageSize: LIMIT,
            total: totalCount,
            showSizeChanger: false,
            onChange: setPage,
          }}
          expandable={{
            expandedRowKeys: expandedId ? [expandedId] : [],
            onExpand: (expanded, developer) => setExpandedId(expanded ? developer.id : null),
            expandedRowRender: renderDeveloperDetail,
          }}
        />
      </div>
    </div>
  )
}
