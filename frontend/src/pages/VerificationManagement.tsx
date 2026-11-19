import { useState, useEffect, useCallback } from 'react'
import { createElement } from 'react';
import type { ElementType } from 'react';
import { useNavigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { SearchOutlined } from '@ant-design/icons'
import { API_BASE_URL } from '../config/api'
import { fetchCsrfToken, getCsrfToken, csrfHeader, clearCsrfToken } from '../lib/csrf'
import { C, injectFonts } from '../theme'
import { LanguageSwitcher } from '../components/common/LanguageSwitcher'
import { formatDate as formatLocaleDate, formatDateTime as formatLocaleDateTime, formatNumber, formatTime as formatLocaleTime } from '../i18n/formatters'
import '../styles/patterns.css'
import {
  ShieldCheckIcon,
  XCircleIcon,
  ClockIcon,
  ExclamationTriangleIcon,
  ArrowPathIcon,
  DocumentTextIcon,
  ArrowLeftIcon,
  CheckCircleIcon,
  ArrowsUpDownIcon,
  PhotoIcon,
  ComputerDesktopIcon,
  UsersIcon,
} from '@heroicons/react/24/outline'
import {
  Alert,
  Button,
  Card,
  Col,
  Descriptions,
  Empty,
  Input,
  Modal,
  Row,
  Segmented,
  Select,
  Spin,
  Statistic,
  Table,
  Tag,
} from 'antd'
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

interface Verification {
  id: string
  user_id: string | null
  status: string
  document_type?: string
  created_at: string
  is_sandbox?: boolean
  source?: string
  failure_reason?: string
  developer_id?: string
  document_thumbnail?: string | null
  selfie_thumbnail?: string | null
}

interface DocumentInfo {
  id: string
  file_name: string
  document_type: string
  url: string | null
}

interface GateDebug {
  gates: {
    ocr: { extracted: boolean | null; quality_score: number | null; quality_analysis: any; fields: Record<string, string> | null; back_fields: Record<string, string> | null; barcode_data: any }
    cross_validation: { score: number | null; verdict: string | null; mismatches: Array<{ field: string; front: string; back: string }>; results: any }
    liveness: { score: number | null; passed: boolean | null }
    deepfake: { is_real: boolean | null; real_probability: number | null; fake_probability: number | null }
    face_match: { score: number | null; passed: boolean | null }
    photo_consistency: { score: number | null }
    address: { status: string | null; score: number | null }
  }
  risk: { overall_score: number; risk_level: string; factors: Array<{ factor?: string; name?: string; weight?: number; score?: number }>; computed_at: string } | null
  aml: { risk_level: string; match_found: boolean; match_count: number; matches: Array<{ listed_name: string; list_source: string; score: number; match_type: string }>; lists_checked: string[]; screened_at: string } | null
  timing: { session_started_at: string | null; processing_completed_at: string | null }
  decision: { failure_reason: string | null; manual_review_reason: string | null; reviewed_by: string | null; reviewed_at: string | null }
  duplicates: { flags: DuplicateFlag[] | null; fingerprints: Array<{ fingerprint_type: string; hash_value: string; created_at: string }> } | null
}

interface DuplicateFlag {
  type: string
  matched_verification_id: string
  hamming_distance: number
}

interface VerificationDetail {
  id: string
  user_id: string
  status: string
  document_type?: string
  created_at: string
  is_sandbox?: boolean
  source?: string
  failure_reason?: string
  document_url: string | null
  selfie_url: string | null
  documents: DocumentInfo[]
  user?: { id: string; full_name?: string; email?: string } | null
  developer?: { id: string; company_name?: string; email?: string } | null
  debug?: GateDebug
  duplicate_flags?: DuplicateFlag[] | null
}

interface Stats {
  total: number
  pending: number
  verified: number
  failed: number
  manual_review: number
}

interface ConfirmAction {
  verificationId: string
  decision: 'approve' | 'reject' | 'override'
  newStatus?: string
}

type StatusFilter = 'all' | 'manual_review' | 'verified' | 'failed' | 'pending'

// ─── Helpers ────────────────────────────────────────────────

const STATUS_CONFIG: Record<string, { tagColor: string; color: string; icon: ElementType; label: string }> = {
  verified:      { tagColor: 'success',    color: T.green, icon: ShieldCheckIcon,        label: 'verifyMgmt.statusVerified' },
  failed:        { tagColor: 'error',      color: T.red,   icon: XCircleIcon,            label: 'verifyMgmt.statusFailed' },
  manual_review: { tagColor: 'warning',    color: T.amber, icon: ExclamationTriangleIcon, label: 'verifyMgmt.statusManualReview' },
  pending:       { tagColor: 'processing', color: T.cyan,  icon: ClockIcon,              label: 'verifyMgmt.statusPending' },
  processing:    { tagColor: 'processing', color: T.blue,  icon: ArrowPathIcon,          label: 'verifyMgmt.processing' },
}

const getStatusConfig = (status: string) => STATUS_CONFIG[status] || STATUS_CONFIG.pending

export const truncateId = (id: string | null | undefined) => {
  if (!id) return '—'
  return id.length > 12 ? `${id.slice(0, 6)}...${id.slice(-4)}` : id
}

const formatDate = (iso: string) => {
  return formatLocaleDate(iso, { month: 'short', day: 'numeric', year: 'numeric' })
}

const formatTime = (iso: string) => {
  return formatLocaleTime(iso)
}

const authFetchOpts = (): RequestInit => ({ credentials: 'include' })
const authFetchOptsJson = (): RequestInit => ({ credentials: 'include', headers: { 'Content-Type': 'application/json', ...csrfHeader() } })

/** Returns theme color for a 0-1 score: green ≥0.7, amber ≥0.4, red <0.4, dim if null */
const scoreColor = (score: number | null): string => {
  if (score === null || score === undefined) return T.dim
  if (score >= 0.7) return T.green
  if (score >= 0.4) return T.amber
  return T.red
}

/** Format a 0-1 score as percentage, or '—' if null */
const fmtScore = (score: number | null): string => {
  if (score === null || score === undefined) return '—'
  return `${(score * 100).toFixed(0)}%`
}

/** Inline score bar component */
function ScoreBar({ label, score, passed }: { label: string; score: number | null; passed?: boolean | null }) {
  const { t } = useTranslation()
  const pct = score !== null && score !== undefined ? Math.round(score * 100) : 0
  const color = scoreColor(score)
  return (
    <div style={{ marginBottom: 8 }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 3 }}>
        <span style={{ color: T.muted, fontSize: 11, fontWeight: 500 }}>{label}</span>
        <span style={{ color, fontSize: 12, fontFamily: T.mono, fontWeight: 600 }}>
          {fmtScore(score)}
          {passed !== null && passed !== undefined && (
            <span style={{ marginLeft: 6, color: passed ? T.green : T.red, fontSize: 10 }}>
              {passed ? t('verifyMgmt.pass') : t('verifyMgmt.fail')}
            </span>
          )}
        </span>
      </div>
      <div style={{ height: 3, background: 'var(--rule)', overflow: 'hidden' }}>
        <div style={{ width: `${pct}%`, height: '100%', background: color, transition: 'width 0.3s ease' }} />
      </div>
    </div>
  )
}

// ─── Component ──────────────────────────────────────────────

export function VerificationManagement() {
  const { t } = useTranslation()
  const navigate = useNavigate()

  useEffect(() => { injectFonts() }, [])

  // ── State ──
  const [verifications, setVerifications] = useState<Verification[]>([])
  const [stats, setStats] = useState<Stats>({ total: 0, pending: 0, verified: 0, failed: 0, manual_review: 0 })
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [statusFilter, setStatusFilter] = useState<StatusFilter>('all')
  const [searchQuery, setSearchQuery] = useState('')
  const [page, setPage] = useState(1)
  const [totalPages, setTotalPages] = useState(1)
  const [expandedId, setExpandedId] = useState<string | null>(null)
  const [detail, setDetail] = useState<VerificationDetail | null>(null)
  const [detailLoading, setDetailLoading] = useState(false)
  const [confirmAction, setConfirmAction] = useState<ConfirmAction | null>(null)
  const [actionLoading, setActionLoading] = useState(false)
  const [actionReason, setActionReason] = useState('')
  const [overrideStatus, setOverrideStatus] = useState('verified')
  const [toast, setToast] = useState<{ message: string; type: 'success' | 'error' } | null>(null)
  const [authReady, setAuthReady] = useState(() => !!getCsrfToken())
  const [userRole, setUserRole] = useState<'admin' | 'reviewer' | 'platform'>('reviewer')
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
    fetch(`${API_BASE_URL}/api/admin/dashboard`, { credentials: 'include' })
      .then(res => {
        if (res.ok) { setAuthReady(true); fetchCsrfToken(); return }
        clearCsrfToken(); navigate('/admin/login')
      })
      .catch(() => { clearCsrfToken(); navigate('/admin/login') })
  }, [navigate])

  // ── Detect user role (org admin vs reviewer vs platform admin) ──
  useEffect(() => {
    if (!authReady) return
    // Probe analytics — org admins and platform admins get 200, regular reviewers get 403
    fetch(`${API_BASE_URL}/api/admin/analytics?period=7d`, { credentials: 'include' })
      .then(res => {
        if (res.ok) {
          // Could be org admin or platform admin — try developers list to distinguish
          fetch(`${API_BASE_URL}/api/admin/developers?limit=1`, { credentials: 'include' })
            .then(r => setUserRole(r.ok ? 'platform' : 'admin'))
            .catch(() => setUserRole('admin'))
        } else {
          setUserRole('reviewer')
        }
      })
      .catch(() => setUserRole('reviewer'))
  }, [authReady])

  // ── Toast auto-dismiss ──
  useEffect(() => {
    if (!toast) return
    const timer = setTimeout(() => setToast(null), 4000)
    return () => clearTimeout(timer)
  }, [toast])

  // ── Fetch verifications ──
  const fetchVerifications = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      const params = new URLSearchParams({ page: String(page), limit: String(LIMIT) })
      if (statusFilter !== 'all') params.set('status', statusFilter)

      const res = await fetch(`${API_BASE_URL}/api/admin/verifications?${params}`, authFetchOpts())
      if (res.status === 401) {
        navigate('/admin/login')
        return
      }
      if (!res.ok) throw new Error(`HTTP ${res.status}`)

      const data = await res.json()
      setVerifications(data.verifications || [])
      setTotalPages(data.pagination?.pages || 1)
    } catch (caught: unknown) {
      const err = caught as Error;
      setError(err.message || t('verifyMgmt.failedLoadList'))
    } finally {
      setLoading(false)
    }
  }, [page, statusFilter, navigate])

  // ── Fetch stats ──
  const fetchStats = useCallback(async () => {
    try {
      const res = await fetch(`${API_BASE_URL}/api/admin/dashboard`, authFetchOpts())
      if (!res.ok) return
      const data = await res.json()
      if (data.stats) setStats(data.stats)
    } catch { /* non-critical */ }
  }, [])

  useEffect(() => { if (authReady) fetchVerifications() }, [authReady, fetchVerifications])
  useEffect(() => { if (authReady) fetchStats() }, [authReady, fetchStats])

  // ── Fetch detail when expanding a row ──
  const toggleExpand = async (id: string) => {
    if (expandedId === id) { setExpandedId(null); setDetail(null); return }
    setExpandedId(id)
    setDetail(null)
    setDetailLoading(true)
    try {
      const res = await fetch(`${API_BASE_URL}/api/admin/verification/${id}`, authFetchOpts())
      if (!res.ok) throw new Error(`HTTP ${res.status}`)
      const data = await res.json()
      setDetail(data.verification)
    } catch { setDetail(null) }
    finally { setDetailLoading(false) }
  }

  // ── Execute review action ──
  const executeAction = async () => {
    if (!confirmAction) return
    setActionLoading(true)
    try {
      const body: Record<string, string> = { decision: confirmAction.decision }
      if (confirmAction.decision === 'override') body.new_status = confirmAction.newStatus || overrideStatus
      if (actionReason.trim()) body.reason = actionReason.trim()

      const res = await fetch(`${API_BASE_URL}/api/admin/verification/${confirmAction.verificationId}/review`, {
        method: 'PUT', ...authFetchOptsJson(), body: JSON.stringify(body),
      })
      if (!res.ok) { const d = await res.json().catch(() => ({})); throw new Error(d.message || `HTTP ${res.status}`) }

      const resolvedStatus = confirmAction.decision === 'approve' ? 'verified'
        : confirmAction.decision === 'reject' ? 'failed'
        : confirmAction.newStatus || overrideStatus

      setVerifications(prev => prev.map(v => v.id === confirmAction.verificationId ? { ...v, status: resolvedStatus } : v))
      if (detail?.id === confirmAction.verificationId) setDetail(d => d ? { ...d, status: resolvedStatus } : d)
      const successKey = confirmAction.decision === 'approve'
        ? 'approveSuccess'
        : confirmAction.decision === 'reject'
          ? 'rejectSuccess'
          : 'overrideSuccess'
      setToast({ message: t(`verifyMgmt.${successKey}`, { status: resolvedStatus }), type: 'success' })
      setConfirmAction(null)
      setActionReason('')
      fetchStats()
    } catch (caught: unknown) {
      const err = caught as Error;
      setToast({ message: err.message || t('verifyMgmt.actionFailed'), type: 'error' })
    } finally {
      setActionLoading(false)
    }
  }

  // ── Search filter (client-side on loaded page) ──
  const filtered = searchQuery
    ? verifications.filter(v =>
        v.id.toLowerCase().includes(searchQuery.toLowerCase()) ||
        (v.user_id?.toLowerCase().includes(searchQuery.toLowerCase()) ?? false)
      )
    : verifications

  const closeConfirm = () => {
    if (actionLoading) return
    setConfirmAction(null)
    setActionReason('')
  }

  // ── Status Tag renderer ──
  const renderStatusTag = (status: string) => {
    const cfg = getStatusConfig(status)
    return (
      <Tag color={cfg.tagColor} style={{ marginInlineEnd: 0 }}>
        <span style={{ display: 'inline-flex', alignItems: 'center', gap: 4 }}>
          {createElement(cfg.icon, { style: { width: 12, height: 12 } })}
          {t(cfg.label)}
        </span>
      </Tag>
    )
  }

  const filterOptions: Array<{ label: string; value: StatusFilter }> = [
    { label: t('verifyMgmt.all'), value: 'all' },
    { label: t('verifyMgmt.review'), value: 'manual_review' },
    { label: t('verifyMgmt.statusPending'), value: 'pending' },
    { label: t('verifyMgmt.statusVerified'), value: 'verified' },
    { label: t('verifyMgmt.statusFailed'), value: 'failed' },
  ]

  // ── Table columns ──
  const columns: ColumnsType<Verification> = [
    {
      title: t('verifyMgmt.colPreview'),
      key: 'preview',
      width: 90,
      render: (_, v) => (
        <div style={{ display: 'flex', alignItems: 'center', gap: 4 }} onClick={e => e.stopPropagation()}>
          {v.document_thumbnail ? (
            <img
              src={v.document_thumbnail}
              alt="Doc"
              style={{ width: 32, height: 32, objectFit: 'cover', border: `1px solid ${T.border}` }}
              onError={e => { (e.target as HTMLImageElement).style.display = 'none' }}
            />
          ) : (
            <div style={{ width: 32, height: 32, background: T.codeBg, border: `1px solid ${T.border}`, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
              <PhotoIcon style={{ width: 14, height: 14, color: T.dim }} />
            </div>
          )}
          {v.selfie_thumbnail ? (
            <img
              src={v.selfie_thumbnail}
              alt={t('verifyMgmt.live')}
              style={{ width: 32, height: 32, objectFit: 'cover', border: `1px solid ${T.border}` }}
              onError={e => { (e.target as HTMLImageElement).style.display = 'none' }}
            />
          ) : null}
        </div>
      ),
    },
    {
      title: t('verifyMgmt.colVerificationId'),
      key: 'verification_id',
      render: (_, v) => (
        <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
          <span style={{ color: T.text, fontSize: 13, fontFamily: T.mono }} title={v.id}>
            {truncateId(v.id)}
          </span>
          {v.is_sandbox && (
            <Tag color="purple" style={{ marginInlineEnd: 0, fontSize: 9, fontWeight: 600, textTransform: 'uppercase', letterSpacing: '0.05em' }}>
              {t('verifyMgmt.sandbox')}
            </Tag>
          )}
        </div>
      ),
    },
    {
      title: t('verifyMgmt.colUserId'),
      key: 'user_id',
      render: (_, v) => (
        <span style={{ color: T.muted, fontSize: 13, fontFamily: T.mono }} title={v.user_id ?? undefined}>
          {truncateId(v.user_id)}
        </span>
      ),
    },
    {
      title: t('verifyMgmt.colStatus'),
      key: 'status',
      width: 150,
      render: (_, v) => renderStatusTag(v.status),
    },
    {
      title: t('verifyMgmt.colType'),
      key: 'document_type',
      width: 130,
      render: (_, v) => (
        <Tag style={{ marginInlineEnd: 0 }}>{(v.document_type || 'unknown').replace(/_/g, ' ')}</Tag>
      ),
    },
    {
      title: t('verifyMgmt.colCreated'),
      key: 'created_at',
      width: 150,
      render: (_, v) => (
        <div style={{ display: 'flex', flexDirection: 'column', justifyContent: 'center' }}>
          <span style={{ color: T.muted, fontSize: 12 }}>{formatDate(v.created_at)}</span>
          <span style={{ color: T.dim, fontSize: 10, fontFamily: T.mono }}>{formatTime(v.created_at)}</span>
        </div>
      ),
    },
    {
      title: t('verifyMgmt.colActions'),
      key: 'actions',
      width: 200,
      render: (_, v) => (
        <div style={{ display: 'flex', alignItems: 'center', gap: 6 }} onClick={e => e.stopPropagation()}>
          {(v.status === 'verified' || v.status === 'failed') ? (
            <span style={{
              display: 'inline-flex', alignItems: 'center', gap: 4,
              color: v.status === 'verified' ? T.green : T.red,
              fontSize: 11, fontWeight: 600, fontFamily: T.mono,
            }}>
              {v.status === 'verified' ? (
                <><CheckCircleIcon style={{ width: 14, height: 14 }} /> {t('verifyMgmt.complete')}</>
              ) : (
                <><XCircleIcon style={{ width: 14, height: 14 }} /> {t('verifyMgmt.complete')}</>
              )}
            </span>
          ) : null}
          {(v.status === 'manual_review' || v.status === 'pending') && (
            <>
              <Button
                size="small"
                onClick={() => setConfirmAction({ verificationId: v.id, decision: 'approve' })}
                title={t('verifyMgmt.approve')}
                style={{
                  background: T.greenDim, border: `1px solid rgba(52,211,153,0.3)`,
                  color: T.green, fontSize: 11, fontWeight: 500, fontFamily: T.mono,
                }}
              >
                {t('verifyMgmt.approve')}
              </Button>
              <Button
                size="small"
                onClick={() => setConfirmAction({ verificationId: v.id, decision: 'reject' })}
                title={t('verifyMgmt.reject')}
                style={{
                  background: T.redDim, border: `1px solid rgba(248,113,113,0.3)`,
                  color: T.red, fontSize: 11, fontWeight: 500, fontFamily: T.mono,
                }}
              >
                {t('verifyMgmt.reject')}
              </Button>
            </>
          )}
          {userRole !== 'reviewer' && (
            <Button
              size="small"
              onClick={() => setConfirmAction({ verificationId: v.id, decision: 'override' })}
              title={t('verifyMgmt.override')}
              icon={<ArrowsUpDownIcon style={{ width: 11, height: 11 }} />}
              style={{
                background: T.amberDim, border: `1px solid rgba(251,191,36,0.2)`,
                color: T.amber, fontSize: 11, fontWeight: 500, fontFamily: T.mono,
              }}
            >
              {t('verifyMgmt.override')}
            </Button>
          )}
        </div>
      ),
    },
  ]

  // ── Detail sections ──
  const renderDocumentsSection = () => {
    const d = detail!
    return (
      <Card size="small" title={t('verifyMgmt.uploadedDocuments')}>
        {d.documents && d.documents.length > 0 ? (
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(160px, 1fr))', gap: 12 }}>
            {d.documents.map(doc => (
              <div key={doc.id} style={{ background: T.codeBg, border: `1px solid ${T.border}`, overflow: 'hidden' }}>
                {doc.url ? (
                  <img
                    src={doc.url}
                    alt={doc.file_name}
                    style={{ width: '100%', height: 160, objectFit: 'cover', display: 'block' }}
                    onError={e => { (e.target as HTMLImageElement).style.display = 'none' }}
                  />
                ) : (
                  <div style={{
                    width: '100%', height: 160, display: 'flex', alignItems: 'center', justifyContent: 'center',
                    background: T.codeBg,
                  }}>
                    <PhotoIcon style={{ width: 32, height: 32, color: T.dim }} />
                  </div>
                )}
                <div style={{ padding: '8px 10px' }}>
                  <div style={{ color: T.muted, fontSize: 11, fontFamily: T.mono }}>{doc.file_name}</div>
                  <div style={{ color: T.dim, fontSize: 10, marginTop: 2 }}>{doc.document_type}</div>
                </div>
              </div>
            ))}

            {/* Selfie */}
            {d.selfie_url && (
              <div style={{ background: T.codeBg, border: `1px solid ${T.border}`, overflow: 'hidden' }}>
                <img
                  src={d.selfie_url}
                  alt={t('verifyMgmt.live')}
                  style={{ width: '100%', height: 160, objectFit: 'cover', display: 'block' }}
                  onError={e => { (e.target as HTMLImageElement).style.display = 'none' }}
                />
                <div style={{ padding: '8px 10px' }}>
                  <div style={{ color: T.muted, fontSize: 11, fontFamily: T.mono }}>{t('verifyMgmt.live')}</div>
                  <div style={{ color: T.dim, fontSize: 10, marginTop: 2 }}>{t('verifyMgmt.selfie')}</div>
                </div>
              </div>
            )}
          </div>
        ) : (
          <div style={{
            background: T.codeBg, border: `1px solid ${T.border}`, padding: 24, textAlign: 'center',
          }}>
            <DocumentTextIcon style={{ width: 24, height: 24, color: T.dim, margin: '0 auto 8px' }} />
            <div style={{ color: T.dim, fontSize: 12 }}>{t('verifyMgmt.noDocuments')}</div>
          </div>
        )}
      </Card>
    )
  }

  const renderMetadataSection = (v: Verification) => {
    const d = detail!
    const metaRows: Array<[string, string]> = [
      [t('verifyMgmt.fieldVerificationId'), d.id],
      [t('verifyMgmt.fieldUserId'), d.user_id],
      [t('verifyMgmt.fieldStatus'), d.status],
      [t('verifyMgmt.fieldDocumentType'), d.document_type || '-'],
      [t('verifyMgmt.fieldSource'), d.source || 'api'],
      [t('verifyMgmt.sandbox'), d.is_sandbox ? t('verifyMgmt.yes') : t('verifyMgmt.no')],
      [t('verifyMgmt.fieldCreated'), `${formatDate(d.created_at)} ${formatTime(d.created_at)}`],
      ...(d.failure_reason ? [[t('verifyMgmt.fieldFailureReason'), d.failure_reason]] as Array<[string, string]> : []),
      ...(d.developer?.company_name ? [[t('verifyMgmt.fieldDeveloper'), d.developer.company_name]] as Array<[string, string]> : []),
      ...(d.developer?.email ? [[t('verifyMgmt.fieldDeveloperEmail'), d.developer.email]] as Array<[string, string]> : []),
    ]

    return (
      <Card size="small" title={t('verifyMgmt.verificationData')}>
        <Descriptions
          size="small"
          column={1}
          items={metaRows.map(([label, value]) => ({
            key: label,
            label,
            children: label === t('verifyMgmt.fieldStatus') ? renderStatusTag(value) : value,
          }))}
        />

        {/* Action buttons in detail view */}
        <div style={{ display: 'flex', gap: 8, marginTop: 16 }}>
          <Button
            onClick={() => setConfirmAction({ verificationId: v.id, decision: 'approve' })}
            icon={<CheckCircleIcon style={{ width: 16, height: 16 }} />}
            style={{
              flex: 1, background: T.greenDim, border: `1px solid rgba(52,211,153,0.3)`,
              color: T.green, fontSize: 13, fontWeight: 600, fontFamily: T.mono,
            }}
          >
            {t('verifyMgmt.approve')}
          </Button>
          <Button
            onClick={() => setConfirmAction({ verificationId: v.id, decision: 'reject' })}
            icon={<XCircleIcon style={{ width: 16, height: 16 }} />}
            style={{
              flex: 1, background: T.redDim, border: `1px solid rgba(248,113,113,0.3)`,
              color: T.red, fontSize: 13, fontWeight: 600, fontFamily: T.mono,
            }}
          >
            {t('verifyMgmt.reject')}
          </Button>
          {userRole !== 'reviewer' && (
            <Button
              onClick={() => setConfirmAction({ verificationId: v.id, decision: 'override' })}
              icon={<ArrowsUpDownIcon style={{ width: 16, height: 16 }} />}
              style={{
                flex: 1, background: T.amberDim, border: `1px solid rgba(251,191,36,0.2)`,
                color: T.amber, fontSize: 13, fontWeight: 600, fontFamily: T.mono,
              }}
            >
              {t('verifyMgmt.override')}
            </Button>
          )}
        </div>
      </Card>
    )
  }

  const renderDuplicateSection = () => {
    const d = detail!
    const flags = d.duplicate_flags
    if (flags && flags.length > 0) {
      return (
        <Alert
          type="warning"
          showIcon
          style={{ marginTop: 16 }}
          title={`Duplicate Detected (${flags.length} match${flags.length !== 1 ? 'es' : ''})`}
          description={
            <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
              {flags.map((flag, i) => (
                <div key={i} style={{ display: 'flex', alignItems: 'center', gap: 8, fontSize: 12 }}>
                  <Tag
                    color={flag.type === 'document_phash' ? 'cyan' : 'purple'}
                    style={{ marginInlineEnd: 0, fontSize: 10, fontWeight: 600 }}
                  >
                    {flag.type === 'document_phash' ? t('verifyMgmt.docTypeDocument') : t('verifyMgmt.docTypeFace')}
                  </Tag>
                  <span style={{ color: T.muted }}>{t('verifyMgmt.matchesLabel')}</span>
                  <Button
                    type="link"
                    size="small"
                    style={{ padding: 0, fontSize: 11, fontFamily: T.mono }}
                    onClick={() => toggleExpand(flag.matched_verification_id)}
                  >
                    {truncateId(flag.matched_verification_id)}
                  </Button>
                  <span style={{ color: T.dim, fontSize: 10 }}>
                    {t('verifyMgmt.distance', { value: flag.hamming_distance })}
                  </span>
                </div>
              ))}
            </div>
          }
        />
      )
    }

    // Also show duplicate flags from debug.duplicates if available
    const debugFlags = d.debug?.duplicates?.flags
    if (debugFlags && debugFlags.length > 0 && !d.duplicate_flags?.length) {
      return (
        <Alert
          type="warning"
          showIcon
          style={{ marginTop: 16 }}
          title={`Duplicate Detected (${debugFlags.length} match${debugFlags.length !== 1 ? 'es' : ''})`}
        />
      )
    }

    return null
  }

  const renderGateAnalysis = () => {
    const debug = detail!.debug
    if (!debug) return null
    const risk = debug.risk
    const aml = debug.aml
    const crossValidation = debug.gates.cross_validation

    return (
      <div style={{ marginTop: 16 }}>
        <div style={{ color: T.dim, fontSize: 10, fontWeight: 600, textTransform: 'uppercase', letterSpacing: '0.1em', marginBottom: 12 }}>
          {t('verifyMgmt.gateAnalysis')}
        </div>

        <Row gutter={[12, 12]}>
          {/* Gate Scores */}
          <Col xs={24} lg={8}>
            <Card size="small" title={t('verifyMgmt.gateScores')}>
              <ScoreBar label={t('verifyMgmt.ocrQuality')} score={debug.gates.ocr.quality_score} />
              <ScoreBar label={t('verifyMgmt.crossValidation')} score={crossValidation.score} passed={crossValidation.verdict === 'PASS' ? true : crossValidation.verdict === 'REJECT' ? false : null} />
              <ScoreBar label={t('verifyMgmt.liveness')} score={debug.gates.liveness.score} passed={debug.gates.liveness.passed} />
              <ScoreBar label={t('verifyMgmt.deepfake')} score={debug.gates.deepfake.real_probability} passed={debug.gates.deepfake.is_real} />
              <ScoreBar label={t('verifyMgmt.faceMatch')} score={debug.gates.face_match.score} passed={debug.gates.face_match.passed} />
              {debug.gates.photo_consistency.score !== null && (
                <ScoreBar label={t('verifyMgmt.photoConsistency')} score={debug.gates.photo_consistency.score} />
              )}
              {debug.gates.address.score !== null && (
                <ScoreBar label={t('verifyMgmt.addressMatch')} score={debug.gates.address.score} />
              )}
            </Card>
          </Col>

          {/* OCR Extracted Fields */}
          <Col xs={24} lg={8}>
            <Card size="small" title={t('verifyMgmt.ocrFields')}>
              {debug.gates.ocr.fields ? (
                <Descriptions
                  size="small"
                  column={1}
                  items={Object.entries(debug.gates.ocr.fields).map(([key, val]) => ({
                    key,
                    label: key.replace(/_/g, ' '),
                    children: val || '—',
                  }))}
                />
              ) : (
                <div style={{ color: T.dim, fontSize: 11 }}>{t('verifyMgmt.noOcr')}</div>
              )}
              {crossValidation.mismatches && crossValidation.mismatches.length > 0 && (
                <div style={{ marginTop: 12 }}>
                  <div style={{ color: T.amber, fontSize: 11, fontWeight: 600, marginBottom: 4 }}>{t('verifyMgmt.crossMismatches')}</div>
                  {crossValidation.mismatches.map((m, i) => (
                    <div key={i} style={{ color: T.muted, fontSize: 10, fontFamily: T.mono, padding: '2px 0' }}>
                      {m.field}: front="{m.front}" vs back="{m.back}"
                    </div>
                  ))}
                </div>
              )}
            </Card>
          </Col>

          {/* Risk & AML */}
          <Col xs={24} lg={8}>
            <Card size="small">
              {risk ? (
                <>
                  <div style={{ color: T.text, fontSize: 12, fontWeight: 600, marginBottom: 12 }}>{t('verifyMgmt.riskAssessment')}</div>
                  <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 8 }}>
                    <span style={{
                      fontSize: 20, fontWeight: 700, fontFamily: T.mono,
                      color: risk.overall_score <= 30 ? T.green : risk.overall_score <= 60 ? T.amber : T.red,
                    }}>
                      {risk.overall_score}
                    </span>
                    <Tag
                      color={risk.risk_level === 'low' ? 'success' : risk.risk_level === 'medium' ? 'warning' : 'error'}
                      style={{ marginInlineEnd: 0, fontSize: 10, fontWeight: 600, textTransform: 'uppercase' }}
                    >
                      {risk.risk_level}
                    </Tag>
                  </div>
                  {risk.factors && (
                    <div>
                      {(Array.isArray(risk.factors) ? risk.factors : []).map((f, i) => (
                        <div key={i} style={{ display: 'flex', justifyContent: 'space-between', padding: '2px 0' }}>
                          <span style={{ color: T.dim, fontSize: 10 }}>{f.factor || f.name}</span>
                          <span style={{ color: T.muted, fontSize: 10, fontFamily: T.mono }}>{f.weight ?? f.score}</span>
                        </div>
                      ))}
                    </div>
                  )}
                </>
              ) : (
                <div style={{ color: T.dim, fontSize: 11 }}>{t('verifyMgmt.noRisk')}</div>
              )}

              {aml && (
                <div style={{ marginTop: 16, paddingTop: 12, borderTop: `1px solid ${T.border}` }}>
                  <div style={{ color: T.text, fontSize: 12, fontWeight: 600, marginBottom: 8 }}>{t('verifyMgmt.amlScreening')}</div>
                  <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 4 }}>
                    <Tag
                      color={aml.match_found ? 'error' : 'success'}
                      style={{ marginInlineEnd: 0, fontSize: 10, fontWeight: 600, textTransform: 'uppercase' }}
                    >
                      {aml.match_found ? t('verifyMgmt.matchFound') : t('verifyMgmt.clear')}
                    </Tag>
                    <Tag
                      color={aml.risk_level === 'confirmed_match' ? 'error' : aml.risk_level === 'potential_match' ? 'warning' : 'success'}
                      style={{ marginInlineEnd: 0, fontSize: 10, fontWeight: 600, textTransform: 'uppercase' }}
                    >
                      {aml.risk_level?.replace(/_/g, ' ') || 'clear'}
                    </Tag>
                  </div>
                  <div style={{ color: T.dim, fontSize: 10, marginTop: 4 }}>
                    Lists: {(aml.lists_checked || []).join(', ') || '—'}
                    {aml.screened_at && (
                      <span style={{ marginLeft: 8 }}>
                        Screened: {formatLocaleDateTime(aml.screened_at)}
                      </span>
                    )}
                  </div>
                  {aml.match_found && aml.matches?.length > 0 && (
                    <div style={{ marginTop: 8 }}>
                      <div style={{ color: T.muted, fontSize: 10, fontWeight: 600, marginBottom: 4 }}>
                        {aml.match_count ?? aml.matches.length} match{(aml.match_count ?? aml.matches.length) !== 1 ? 'es' : ''}
                      </div>
                      {aml.matches.map((m, i) => (
                        <div key={i} style={{
                          display: 'flex', justifyContent: 'space-between', alignItems: 'center',
                          padding: '3px 0', borderBottom: i < aml.matches.length - 1 ? `1px solid ${T.border}` : undefined,
                        }}>
                          <div style={{ flex: 1, minWidth: 0 }}>
                            <span style={{ color: T.text, fontSize: 11 }}>{m.listed_name}</span>
                            <span style={{ color: T.dim, fontSize: 10, marginLeft: 6 }}>{m.list_source}</span>
                          </div>
                          <div style={{ display: 'flex', alignItems: 'center', gap: 6, flexShrink: 0 }}>
                            <span style={{ color: T.dim, fontSize: 9, textTransform: 'uppercase' }}>{m.match_type?.replace(/_/g, '+')}</span>
                            <span style={{
                              fontFamily: T.mono, fontSize: 11, fontWeight: 600,
                              color: m.score >= 0.95 ? T.red : m.score >= 0.85 ? T.amber : T.muted,
                            }}>
                              {(m.score * 100).toFixed(0)}%
                            </span>
                          </div>
                        </div>
                      ))}
                    </div>
                  )}
                </div>
              )}

              {debug.timing.processing_completed_at && debug.timing.session_started_at && (
                <div style={{ marginTop: 16, paddingTop: 12, borderTop: `1px solid ${T.border}` }}>
                  <div style={{ color: T.text, fontSize: 12, fontWeight: 600, marginBottom: 4 }}>{t('verifyMgmt.processingTime')}</div>
                  <div style={{ color: T.cyan, fontSize: 14, fontFamily: T.mono, fontWeight: 600 }}>
                    {((new Date(debug.timing.processing_completed_at).getTime() - new Date(debug.timing.session_started_at).getTime()) / 1000).toFixed(1)}s
                  </div>
                </div>
              )}
            </Card>
          </Col>
        </Row>
      </div>
    )
  }

  const renderDetail = (v: Verification) => (
    <div style={{ padding: '8px 4px' }}>
      {detailLoading && (
        <div style={{ padding: 32, textAlign: 'center' }}>
          <Spin />
        </div>
      )}

      {!detailLoading && detail && (
        <div>
          <Row gutter={[20, 20]}>
            <Col xs={24} lg={10}>{renderDocumentsSection()}</Col>
            <Col xs={24} lg={14}>{renderMetadataSection(v)}</Col>
          </Row>
          {renderDuplicateSection()}
          {renderGateAnalysis()}
        </div>
      )}

      {!detailLoading && !detail && (
        <Alert type="error" showIcon title={t('verifyMgmt.failedLoad')} />
      )}
    </div>
  )

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
          border: `1px solid ${T.border}`, maxWidth: 400,
        }}>
          <ComputerDesktopIcon style={{ width: 48, height: 48, color: T.cyan, margin: '0 auto 20px' }} />
          <h2 style={{ color: T.text, fontSize: 20, fontWeight: 600, margin: '0 0 12px', fontFamily: T.sans }}>
            {t('verifyMgmt.desktopRequired')}
          </h2>
          <p style={{ color: T.muted, fontSize: 14, lineHeight: 1.6, margin: '0 0 24px' }}>
            {t('verifyMgmt.desktopRequiredDesc')}
          </p>
          <Button onClick={() => navigate(-1)}>{t('verifyMgmt.goBack')}</Button>
        </div>
      </div>
    )
  }

  // ── Render ──
  return (
    <div className="pattern-crosshatch pattern-faint pattern-fade-edges pattern-full" style={{ background: T.bg, minHeight: '100vh', fontFamily: T.sans }}>
      <div style={{ maxWidth: 1400, margin: '0 auto', padding: '32px 24px' }}>

        {/* ── Header ── */}
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', flexWrap: 'wrap', gap: 16, marginBottom: 32 }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 16 }}>
            <Button
              type="text"
              onClick={() => navigate('/admin/login')}
              style={{ color: T.muted, display: 'flex' }}
              title={t('verifyMgmt.back')}
            >
              <ArrowLeftIcon style={{ width: 20, height: 20 }} />
            </Button>
            <div>
              <div style={{ fontFamily: T.mono, fontSize: 11, color: T.muted, letterSpacing: '0.08em', marginBottom: 8 }}>
                {t('verifyMgmt.eyebrow')}
              </div>
              <h1 style={{ fontFamily: T.mono, color: T.text, fontSize: 24, fontWeight: 600, margin: 0 }}>
                {t('verifyMgmt.title')}
              </h1>
              <p style={{ color: T.dim, fontSize: 13, margin: '4px 0 0', fontFamily: T.mono }}>
                {t('verifyMgmt.subtitle')}
              </p>
            </div>
          </div>
          <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
            <LanguageSwitcher variant="console" />
            {userRole === 'platform' && (
              <Button
                onClick={() => navigate('/admin/developers')}
                icon={<UsersIcon style={{ width: 14, height: 14 }} />}
                style={{ fontSize: 13 }}
              >
                {t('verifyMgmt.developers')}
              </Button>
            )}
            {/* Role badge */}
            <Tag color={userRole === 'platform' ? 'purple' : userRole === 'admin' ? 'cyan' : undefined} style={{ marginInlineEnd: 0, fontSize: 11 }}>
              {userRole === 'platform' ? t('verifyMgmt.rolePlatform') : userRole === 'admin' ? t('verifyMgmt.roleAdmin') : t('verifyMgmt.roleReviewer')}
            </Tag>
            <Button
              onClick={() => { fetchVerifications(); fetchStats() }}
              icon={<ArrowPathIcon style={{ width: 14, height: 14 }} />}
              style={{ fontSize: 13 }}
            >
              {t('verifyMgmt.refresh')}
            </Button>
            <Button
              danger
              onClick={() => {
                fetch(`${API_BASE_URL}/api/auth/logout`, { method: 'POST', credentials: 'include', headers: csrfHeader() }).catch(() => {})
                clearCsrfToken()
                navigate('/admin/login')
              }}
              style={{ fontSize: 13 }}
            >
              {t('verifyMgmt.signOut')}
            </Button>
          </div>
        </div>

        {/* ── Stats Cards ── */}
        <Row gutter={[12, 12]} style={{ marginBottom: 28 }}>
          {[
            { label: t('verifyMgmt.total'), value: stats.total, color: T.text },
            { label: t('verifyMgmt.review'), value: stats.manual_review, color: T.amber },
            { label: t('verifyMgmt.statusPending'), value: stats.pending, color: T.cyan },
            { label: t('verifyMgmt.statusVerified'), value: stats.verified, color: T.green },
            { label: t('verifyMgmt.statusFailed'), value: stats.failed, color: T.red },
          ].map(s => (
            <Col key={s.label} flex="1 1 160px">
              <Card size="small">
                <Statistic
                  title={s.label}
                  value={formatNumber(s.value)}
                  styles={{ content: { color: s.color, fontFamily: T.mono, fontSize: 26, fontWeight: 600, lineHeight: 1.2 } }}
                />
              </Card>
            </Col>
          ))}
        </Row>

        {/* ── Filter Bar ── */}
        <div style={{
          display: 'flex', alignItems: 'center', gap: 12, marginBottom: 16,
          background: T.panel, border: `1px solid ${T.border}`, padding: '12px 16px',
        }}>
          <Segmented
            value={statusFilter}
            options={filterOptions}
            onChange={value => { setStatusFilter(value as StatusFilter); setPage(1) }}
          />
          <div style={{ flex: 1 }} />
          <Input
            allowClear
            prefix={<SearchOutlined style={{ color: T.dim }} />}
            placeholder={t('verifyMgmt.searchPlaceholder')}
            value={searchQuery}
            onChange={e => setSearchQuery(e.target.value)}
            style={{ width: 220, fontSize: 12, fontFamily: T.mono }}
          />
        </div>

        {error && !loading && (
          <Alert
            type="error"
            showIcon
            title={error}
            action={<Button size="small" onClick={fetchVerifications}>{t('verifyMgmt.retry')}</Button>}
            style={{ marginBottom: 16 }}
          />
        )}

        {/* ── Verifications Table ── */}
        <Table<Verification>
          rowKey="id"
          columns={columns}
          dataSource={filtered}
          loading={loading}
          scroll={{ x: 1020 }}
          locale={{
            emptyText: (
              <Empty
                image={Empty.PRESENTED_IMAGE_SIMPLE}
                description={searchQuery ? t('verifyMgmt.noMatch') : t('verifyMgmt.none')}
              />
            ),
          }}
          pagination={{
            current: page,
            pageSize: LIMIT,
            total: totalPages * LIMIT,
            showSizeChanger: false,
            hideOnSinglePage: true,
            onChange: p => setPage(p),
          }}
          expandable={{
            expandedRowKeys: expandedId ? [expandedId] : [],
            expandRowByClick: true,
            onExpand: (expanded, record) => {
              if (expanded) {
                if (expandedId !== record.id) toggleExpand(record.id)
              } else if (expandedId === record.id) {
                toggleExpand(record.id)
              }
            },
            expandedRowRender: record => renderDetail(record),
          }}
          onRow={() => ({ style: { cursor: 'pointer' } })}
        />
      </div>

      {/* ── Confirmation Modal ── */}
      <Modal
        open={!!confirmAction}
        title={confirmAction
          ? confirmAction.decision === 'approve' ? t('verifyMgmt.approveTitle')
            : confirmAction.decision === 'reject' ? t('verifyMgmt.rejectTitle')
            : t('verifyMgmt.overrideTitle')
          : undefined}
        onCancel={closeConfirm}
        mask={{ closable: !actionLoading }}
        keyboard={!actionLoading}
        footer={confirmAction ? [
          <Button key="cancel" onClick={closeConfirm} disabled={actionLoading}>
            {t('verifyMgmt.cancel')}
          </Button>,
          <Button
            key="confirm"
            onClick={executeAction}
            loading={actionLoading}
            style={{
              background: confirmAction.decision === 'approve' ? T.greenDim
                : confirmAction.decision === 'reject' ? T.redDim
                : T.amberDim,
              borderColor: confirmAction.decision === 'approve' ? 'rgba(52,211,153,0.4)'
                : confirmAction.decision === 'reject' ? 'rgba(248,113,113,0.4)'
                : 'rgba(251,191,36,0.4)',
              color: confirmAction.decision === 'approve' ? T.green
                : confirmAction.decision === 'reject' ? T.red
                : T.amber,
              fontWeight: 600,
              fontFamily: T.mono,
            }}
          >
            {confirmAction.decision === 'approve' ? t('verifyMgmt.confirmApprove')
              : confirmAction.decision === 'reject' ? t('verifyMgmt.confirmReject')
              : t('verifyMgmt.confirmOverride')}
          </Button>,
        ] : undefined}
      >
        {confirmAction && (
          <div>
            <div style={{ color: T.muted, fontSize: 13, marginBottom: 16 }}>
              {confirmAction.decision === 'approve'
                ? t('verifyMgmt.approveDesc')
                : confirmAction.decision === 'reject'
                ? t('verifyMgmt.rejectDesc')
                : t('verifyMgmt.overrideDesc')}
            </div>

            <div style={{ marginBottom: 12 }}>
              <div style={{ color: T.dim, fontSize: 11, fontWeight: 600, marginBottom: 4, fontFamily: T.mono }}>
                {t('verifyMgmt.verificationIdLabel')}
              </div>
              <div style={{
                background: T.codeBg, padding: '8px 12px',
                color: T.text, fontSize: 12, fontFamily: T.mono, border: `1px solid ${T.border}`,
              }}>
                {confirmAction.verificationId}
              </div>
            </div>

            {/* Override status selector */}
            {confirmAction.decision === 'override' && (
              <div style={{ marginBottom: 12 }}>
                <div style={{ color: T.dim, fontSize: 11, fontWeight: 600, marginBottom: 4, fontFamily: T.mono }}>
                  {t('verifyMgmt.newStatus')}
                </div>
                <Select
                  value={overrideStatus}
                  onChange={value => setOverrideStatus(value)}
                  style={{ width: '100%' }}
                  options={[
                    { value: 'verified', label: t('verifyMgmt.statusVerified') },
                    { value: 'failed', label: t('verifyMgmt.statusFailed') },
                    { value: 'manual_review', label: t('verifyMgmt.statusManualReview') },
                    { value: 'pending', label: t('verifyMgmt.statusPending') },
                  ]}
                />
              </div>
            )}

            {/* Reason (optional for approve, recommended for reject/override) */}
            <div>
              <div style={{ color: T.dim, fontSize: 11, fontWeight: 600, marginBottom: 4, fontFamily: T.mono }}>
                {t('verifyMgmt.reason')} {confirmAction.decision === 'approve' ? t('verifyMgmt.optional') : t('verifyMgmt.recommended')}
              </div>
              <Input.TextArea
                value={actionReason}
                onChange={e => setActionReason(e.target.value)}
                placeholder={
                  confirmAction.decision === 'approve' ? t('verifyMgmt.notesPlaceholder')
                  : confirmAction.decision === 'reject' ? t('verifyMgmt.rejectPlaceholder')
                  : t('verifyMgmt.overridePlaceholder')
                }
                rows={3}
                style={{ fontSize: 13, resize: 'vertical' }}
              />
            </div>
          </div>
        )}
      </Modal>

      {/* ── Toast Notification ── */}
      {toast && (
        <div style={{ position: 'fixed', bottom: 24, right: 24, zIndex: 200, width: 380, maxWidth: '90vw' }}>
          <Alert
            type={toast.type === 'success' ? 'success' : 'error'}
            showIcon
            title={toast.message}
            closable={{ onClose: () => setToast(null) }}
          />
        </div>
      )}
    </div>
  )
}
