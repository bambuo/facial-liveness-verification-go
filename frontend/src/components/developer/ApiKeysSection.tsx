import { useEffect, useRef, useState } from 'react'
import type { CSSProperties, Dispatch, ReactNode, SetStateAction } from 'react';
import { useTranslation } from 'react-i18next'
import toast from 'react-hot-toast'
import { Link } from 'react-router-dom'
import {
  Alert,
  Button,
  Card,
  Descriptions,
  Empty,
  Form,
  Input,
  Modal,
  Progress,
  Space,
  Spin,
  Switch,
  Table,
  Tabs,
  Tag,
  Typography,
} from 'antd'
import type { ColumnsType } from 'antd/es/table'
import { API_BASE_URL, getDocumentationApiUrl } from '../../config/api'
import { isCommunity } from '../../config/edition'
import { csrfHeader } from '../../lib/csrf'
import { C } from '../../theme'
import {
  TrashIcon,
  PlusIcon,
  ClipboardDocumentIcon,
  ChevronRightIcon,
  ChevronDownIcon,
  CodeBracketIcon,
} from '@heroicons/react/24/outline'

// Theme-aware token alias: structural colors follow the active theme.
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
}
import type { ApiKey, ApiActivity, VerificationDetail, DeveloperStats } from './types'
import { formatDate, formatDateTime, formatNumber } from '../../i18n/formatters'
import { copyToClipboard } from './types'

// ─── JSON syntax highlighting ───────────────────────────────────
const jsonTokenColors = {
  key: T.cyan,
  string: T.green,
  number: T.amber,
  boolean: T.purple,
  null: T.red,
  brace: T.dim,
  comma: 'rgba(255,255,255,0.25)',
} as const;

function highlightJson(obj: unknown): ReactNode[] {
  const raw = JSON.stringify(obj, null, 2);
  if (!raw) return [];
  const nodes: ReactNode[] = [];
  const tokenRe = /("(?:[^"\\]|\\.)*")\s*:|("(?:[^"\\]|\\.)*")|(-?\d+(?:\.\d+)?(?:[eE][+-]?\d+)?)\b|(true|false)|(null)|([{}\[\]])|([,:])|\n( *)/g;
  let lastIndex = 0;
  let match: RegExpExecArray | null;
  let i = 0;
  while ((match = tokenRe.exec(raw)) !== null) {
    if (match.index > lastIndex) nodes.push(raw.slice(lastIndex, match.index));
    lastIndex = match.index + match[0].length;
    if (match[1]) {
      nodes.push(<span key={i++} style={{ color: jsonTokenColors.key }}>{match[1]}</span>);
      nodes.push(<span key={i++} style={{ color: jsonTokenColors.comma }}>: </span>);
    } else if (match[2]) {
      nodes.push(<span key={i++} style={{ color: jsonTokenColors.string }}>{match[2]}</span>);
    } else if (match[3]) {
      nodes.push(<span key={i++} style={{ color: jsonTokenColors.number }}>{match[3]}</span>);
    } else if (match[4]) {
      nodes.push(<span key={i++} style={{ color: jsonTokenColors.boolean }}>{match[4]}</span>);
    } else if (match[5]) {
      nodes.push(<span key={i++} style={{ color: jsonTokenColors.null }}>{match[5]}</span>);
    } else if (match[6]) {
      nodes.push(<span key={i++} style={{ color: jsonTokenColors.brace }}>{match[6]}</span>);
    } else if (match[7]) {
      nodes.push(<span key={i++} style={{ color: jsonTokenColors.comma }}>{match[7]}</span>);
    } else if (match[8] !== undefined) {
      nodes.push('\n' + match[8]);
    }
  }
  if (lastIndex < raw.length) nodes.push(raw.slice(lastIndex));
  return nodes;
}

// Render an OCR field value. Most fields are primitives, but some
// (e.g. confidence_scores: Record<string, number>) are nested objects
// that need to render as a readable key/value list instead of being
// flattened to "[object Object]" via String() coercion.
const renderOcrFieldValue = (value: unknown): ReactNode => {
  if (value == null) return '-'
  if (typeof value === 'object') {
    const entries = Object.entries(value as Record<string, unknown>)
    if (entries.length === 0) return '-'
    return (
      <div style={{ display: 'flex', flexDirection: 'column', gap: 2 }}>
        {entries.map(([k, v]) => (
          <div key={k} style={{ fontSize: 11, lineHeight: 1.4 }}>
            <span style={{ color: T.muted }}>{k.replace(/_/g, ' ')}:</span>{' '}
            <span style={{ color: T.text }}>{typeof v === 'number' ? v.toFixed(2) : String(v)}</span>
          </div>
        ))}
      </div>
    )
  }
  return String(value)
}

const monoText: CSSProperties = { fontFamily: T.mono, fontSize: 12 }
const mutedMonoText: CSSProperties = { fontFamily: T.mono, fontSize: 12, color: T.muted }

// Columns for the per-session request log table.
const apiLogColumns: ColumnsType<ApiActivity> = [
  {
    title: 'Time',
    dataIndex: 'timestamp',
    width: 170,
    render: value => <span style={mutedMonoText}>{formatDateTime(value)}</span>,
  },
  {
    title: 'Method',
    dataIndex: 'method',
    width: 90,
    render: value => <span style={monoText}>{value}</span>,
  },
  {
    title: 'Endpoint',
    dataIndex: 'endpoint',
    render: value => (
      <Typography.Text code style={{ fontSize: 12, wordBreak: 'break-all' }}>{value}</Typography.Text>
    ),
  },
  {
    title: 'Status',
    dataIndex: 'status_code',
    width: 90,
    render: value => <Tag color={value >= 400 ? 'error' : 'success'}>{value}</Tag>,
  },
  {
    title: 'Latency',
    dataIndex: 'response_time_ms',
    width: 90,
    render: value => <span style={mutedMonoText}>{value}ms</span>,
  },
]

// --- CreateKeyModal ---

function CreateKeyModal({ open, onClose, onCreated, token }: {
  open: boolean
  onClose: () => void
  onCreated: (key: ApiKey, fullKey: string) => void
  token: string
}) {
  const { t } = useTranslation()
  const [form] = Form.useForm<{ name: string; is_sandbox: boolean }>()
  const [loading, setLoading] = useState(false)

  const submit = async (values: { name: string; is_sandbox: boolean }) => {
    const { name, is_sandbox: isSandbox } = values
    setLoading(true)
    try {
      const res = await fetch(`${API_BASE_URL}/api/developer/api-key`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', ...(token !== 'session' ? { Authorization: `Bearer ${token}` } : {}), ...csrfHeader() }, credentials: 'include' as RequestCredentials,
        body: JSON.stringify({ name, is_sandbox: isSandbox }),
      })
      const data = await res.json()
      if (!res.ok) throw new Error(data.message || 'Failed to create key')
      const newKey: ApiKey = {
        id: data.key_id,
        name: data.name,
        key_preview: `${data.key_prefix}...`,
        is_sandbox: data.is_sandbox,
        is_active: true,
        last_used_at: null,
        created_at: data.created_at,
        expires_at: data.expires_at,
        status: 'active',
      }
      form.resetFields()
      onCreated(newKey, data.api_key)
    } catch (err: unknown) {
      toast.error(err instanceof Error ? err.message : t('developer.apiKeys.createFailed'))
    } finally {
      setLoading(false)
    }
  }

  return (
    <Modal
      open={open}
      title={t('developer.apiKeys.createTitle')}
      onCancel={onClose}
      footer={null}
      destroyOnHidden
      width={420}
    >
      <Form
        form={form}
        layout="vertical"
        initialValues={{ name: '', is_sandbox: false }}
        onFinish={submit}
      >
        <Form.Item
          name="name"
          label={t('developer.apiKeys.keyName')}
          rules={[{ required: true, whitespace: true }]}
        >
          <Input placeholder={t('developer.apiKeys.keyNamePlaceholder')} autoComplete="off" />
        </Form.Item>
        <Form.Item
          name="is_sandbox"
          label={t('developer.apiKeys.sandboxMode')}
          valuePropName="checked"
          style={{ marginBottom: 16 }}
        >
          <Switch size="small" />
        </Form.Item>
        <Space style={{ width: '100%', justifyContent: 'flex-end' }}>
          <Button onClick={onClose}>{t('developer.apiKeys.cancel')}</Button>
          <Button type="primary" htmlType="submit" loading={loading}>
            {loading ? t('developer.apiKeys.creating') : t('developer.apiKeys.createKey')}
          </Button>
        </Space>
      </Form>
    </Modal>
  )
}

// --- ApiKeysSection ---

interface ApiKeysSectionProps {
  token: string
  apiKeys: ApiKey[]
  setApiKeys: Dispatch<SetStateAction<ApiKey[]>>
  stats: DeveloperStats | null
  newFullKey: string | null
  setNewFullKey: Dispatch<SetStateAction<string | null>>
  onUnauthorized: () => void
  /** Rendered between the usage stats strip and the API keys table */
  renderAfterStats?: ReactNode
  /** When true, hides Create Key and delete actions (operator sees keys read-only) */
  isOperator?: boolean
}

export function ApiKeysSection({ token, apiKeys, setApiKeys, stats, newFullKey, setNewFullKey, onUnauthorized, renderAfterStats, isOperator = false }: ApiKeysSectionProps) {
  const { t } = useTranslation()
  const authHeaders = (token === 'session' ? {} : { Authorization: `Bearer ${token}` }) as Record<string, string>
  const [showCreate, setShowCreate] = useState(false)
  const [deleteId, setDeleteId] = useState<string | null>(null)
  const [deleteLoading, setDeleteLoading] = useState(false)
  const [expandedKeyId, setExpandedKeyId] = useState<string | null>(null)
  const [keyLogs, setKeyLogs] = useState<Record<string, ApiActivity[]>>({})
  const [keySessionOutcomes, setKeySessionOutcomes] = useState<Record<string, Record<string, string>>>({})
  const [logsLoadingForKey, setLogsLoadingForKey] = useState<string | null>(null)
  const [selectedLog, setSelectedLog] = useState<ApiActivity | null>(null)
  const [logSearchByKey, setLogSearchByKey] = useState<Record<string, string>>({})
  const [expandedSessionByKey, setExpandedSessionByKey] = useState<Record<string, string | null>>({})
  const [verificationDetail, setVerificationDetail] = useState<VerificationDetail | null>(null)
  const [verificationDetailLoading, setVerificationDetailLoading] = useState<string | null>(null)
  const [verificationDetailError, setVerificationDetailError] = useState<string | null>(null)
  const [detailTab, setDetailTab] = useState<'scores' | 'json'>('scores')
  const newKeyBannerRef = useRef<HTMLDivElement | null>(null)

  // When a new key is issued, scroll the banner into view so the dev
  // sees the one-time full key immediately instead of having to hunt
  // for it at a different scroll position.
  useEffect(() => {
    if (newFullKey && newKeyBannerRef.current) {
      newKeyBannerRef.current.scrollIntoView({ behavior: 'smooth', block: 'center' })
    }
  }, [newFullKey])

  const copyKey = async (key: string) => {
    const ok = await copyToClipboard(key)
    if (ok) toast.success(t('developer.common.copied'))
    else toast.error(t('developer.common.copyFailedManual'))
  }

  const handleCreated = (key: ApiKey, fullKey: string) => {
    setApiKeys(prev => [...prev, key])
    setNewFullKey(fullKey)
    setShowCreate(false)
  }

  const handleDelete = async (id: string) => {
    if (!token) return
    setDeleteLoading(true)
    try {
      const res = await fetch(`${API_BASE_URL}/api/developer/api-key/${id}`, {
        method: 'DELETE',
        headers: { ...authHeaders, ...csrfHeader() },
        credentials: 'include' as RequestCredentials,
      })
      if (res.status === 401) { onUnauthorized(); return }
      if (!res.ok) throw new Error('Delete failed')
      setApiKeys(prev => prev.filter(k => k.id !== id))
      toast.success(t('developer.apiKeys.keyDeleted'))
    } catch {
      toast.error(t('developer.apiKeys.deleteFailed'))
    } finally {
      setDeleteLoading(false)
      setDeleteId(null)
    }
  }

  const fetchKeyLogs = async (keyId: string) => {
    if (!token) return
    setLogsLoadingForKey(keyId)
    try {
      const res = await fetch(`${API_BASE_URL}/api/developer/activity?api_key_id=${encodeURIComponent(keyId)}`, {
        headers: authHeaders,
        credentials: 'include' as RequestCredentials,
      })
      const data = await res.json()
      if (!res.ok) throw new Error(data.message || 'Failed to load logs')
      setKeyLogs(prev => ({ ...prev, [keyId]: data.recent_activities ?? [] }))
      setKeySessionOutcomes(prev => ({ ...prev, [keyId]: data.session_outcomes ?? {} }))
    } catch {
      toast.error(t('developer.apiKeys.logsLoadFailed'))
      setKeyLogs(prev => ({ ...prev, [keyId]: [] }))
    } finally {
      setLogsLoadingForKey(null)
    }
  }

  const toggleKeyLogs = async (keyId: string) => {
    if (expandedKeyId === keyId) {
      setExpandedKeyId(null)
      return
    }
    setExpandedKeyId(keyId)
    await fetchKeyLogs(keyId)
  }

  const inferResourceLabel = (endpoint: string) => {
    const clean = endpoint.split('?')[0]
    const parts = clean.split('/').filter(Boolean)
    if (parts.length === 0) return 'unknown'
    const apiIndex = parts.indexOf('api')
    if (apiIndex >= 0 && parts[apiIndex + 1]) return parts[apiIndex + 1]
    return parts[0]
  }

  const extractSessionId = (endpoint: string) => {
    const uuidMatch = endpoint.match(/[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}/i)
    if (uuidMatch) return uuidMatch[0]
    const queryMatch = endpoint.match(/(?:session_id|verification_id)=([^&]+)/i)
    if (queryMatch && queryMatch[1]) return decodeURIComponent(queryMatch[1])
    return 'no-session'
  }

  const groupLogsBySession = (logs: ApiActivity[]) => {
    const sorted = [...logs].sort((a, b) => new Date(a.timestamp).getTime() - new Date(b.timestamp).getTime())
    const explicit = sorted.map(log => {
      const sid = extractSessionId(log.endpoint)
      return sid === 'no-session' ? null : sid
    })

    const prevSession: Array<string | null> = []
    const nextSession: Array<string | null> = new Array(sorted.length).fill(null)

    let currentPrev: string | null = null
    for (let i = 0; i < sorted.length; i++) {
      if (explicit[i]) currentPrev = explicit[i]
      prevSession[i] = currentPrev
    }

    let currentNext: string | null = null
    for (let i = sorted.length - 1; i >= 0; i--) {
      if (explicit[i]) currentNext = explicit[i]
      nextSession[i] = currentNext
    }

    return sorted.reduce<Record<string, ApiActivity[]>>((acc, log, idx) => {
      const rawSession = extractSessionId(log.endpoint)
      const assignedSession =
        rawSession !== 'no-session'
          ? rawSession
          : (prevSession[idx] || nextSession[idx] || 'no-session')
      if (!acc[assignedSession]) acc[assignedSession] = []
      acc[assignedSession].push(log)
      return acc
    }, {})
  }

  const toggleSessionGroup = (keyId: string, sessionId: string) => {
    setExpandedSessionByKey(prev => ({
      ...prev,
      [keyId]: prev[keyId] === sessionId ? null : sessionId,
    }))
  }

  const inferActionLabel = (method: string, endpoint: string) => {
    const lowerMethod = method.toUpperCase()
    const lowerEndpoint = endpoint.toLowerCase()
    if (lowerMethod === 'POST' && lowerEndpoint.includes('/start')) return 'Start verification flow'
    if (lowerMethod === 'POST' && lowerEndpoint.includes('/upload')) return 'Upload verification artifact'
    if (lowerMethod === 'GET' && lowerEndpoint.includes('/status')) return 'Read verification status'
    if (lowerMethod === 'POST') return 'Create/submit resource'
    if (lowerMethod === 'GET') return 'Read resource'
    if (lowerMethod === 'PUT' || lowerMethod === 'PATCH') return 'Update resource'
    if (lowerMethod === 'DELETE') return 'Delete resource'
    return 'Process API call'
  }

  const fetchVerificationDetail = async (verificationId: string) => {
    if (!token) return
    setVerificationDetailLoading(verificationId)
    setVerificationDetailError(null)
    try {
      const res = await fetch(`${API_BASE_URL}/api/developer/verifications/${verificationId}`, {
        headers: authHeaders,
        credentials: 'include' as RequestCredentials,
      })
      const data = await res.json()
      if (!res.ok) throw new Error(data.message || 'Failed to load verification details')
      setVerificationDetail(data)
      setDetailTab('scores')
    } catch (err: unknown) {
      setVerificationDetailError(err instanceof Error ? err.message : 'Failed to load verification details')
      setVerificationDetail(null)
    } finally {
      setVerificationDetailLoading(null)
    }
  }

  const closeVerificationDetail = () => {
    setVerificationDetail(null)
    setVerificationDetailError(null)
  }

  const apiUrl = getDocumentationApiUrl();
  const curlSnippet = `curl -X POST ${apiUrl}/api/verification/sessions \\
  -H "X-API-Key: YOUR_API_KEY" \\
  -H "Content-Type: application/json" \\
  -d '{"mode":"sandbox"}'`

  const deletingKey = deleteId ? apiKeys.find(k => k.id === deleteId) ?? null : null

  // ─── API key table columns ────────────────────────────────────
  const keyColumns: ColumnsType<ApiKey> = [
    {
      title: t('developer.apiKeys.name'),
      dataIndex: 'name',
      width: 220,
      ellipsis: { showTitle: true },
      render: value => (
        <Typography.Text strong ellipsis={{ tooltip: value }} style={{ maxWidth: '100%' }}>
          {value}
        </Typography.Text>
      ),
    },
    {
      title: t('developer.apiKeys.key'),
      dataIndex: 'key_preview',
      width: 220,
      render: (value: string) => (
        <Typography.Text code style={{ wordBreak: 'break-all' }}>{value}</Typography.Text>
      ),
    },
    {
      title: t('developer.apiKeys.type'),
      dataIndex: 'is_sandbox',
      width: 100,
      render: (value: boolean) => (
        <Tag color={value ? 'warning' : 'success'}>
          {value ? t('developer.apiKeys.sandbox') : t('developer.apiKeys.live')}
        </Tag>
      ),
    },
    {
      title: t('developer.apiKeys.status'),
      dataIndex: 'status',
      width: 110,
      render: (value: ApiKey['status'], key) => (
        <Tag color={key.is_active && value === 'active' ? 'success' : 'default'}>{value}</Tag>
      ),
    },
    {
      title: t('developer.apiKeys.created'),
      dataIndex: 'created_at',
      width: 140,
      render: value => <span style={mutedMonoText}>{formatDate(value)}</span>,
    },
    {
      title: t('developer.apiKeys.lastUsed'),
      dataIndex: 'last_used_at',
      width: 140,
      render: value => <span style={mutedMonoText}>{value ? formatDate(value) : '—'}</span>,
    },
    {
      title: '',
      key: 'actions',
      width: 170,
      render: (_: unknown, key) => (
        <Space size={4}>
          <Button
            size="small"
            onClick={() => toggleKeyLogs(key.id)}
            icon={
              expandedKeyId === key.id
                ? <ChevronDownIcon style={{ width: 11, height: 11 }} />
                : <ChevronRightIcon style={{ width: 11, height: 11 }} />
            }
          >
            {t('developer.apiKeys.logs')}
          </Button>
          {!isOperator && (
            <Button
              size="small"
              danger
              aria-label={t('developer.apiKeys.key')}
              icon={<TrashIcon style={{ width: 11, height: 11 }} />}
              onClick={() => setDeleteId(key.id)}
            />
          )}
        </Space>
      ),
    },
  ]

  // ─── Expanded row: per-key request log grouped by session ─────
  const renderKeyLogs = (key: ApiKey) => {
    if (logsLoadingForKey === key.id) {
      return (
        <Space size={8} style={{ padding: '8px 0' }}>
          <Spin size="small" />
          <Typography.Text type="secondary">{t('developer.apiKeys.loadingLogs')}</Typography.Text>
        </Space>
      )
    }

    const logs = keyLogs[key.id] ?? []
    if (logs.length === 0) {
      return <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description={t('developer.apiKeys.noCalls')} />
    }

    const searchTerm = (logSearchByKey[key.id] ?? '').trim().toLowerCase()
    const filteredLogs = logs.filter(log => {
      const sessionId = extractSessionId(log.endpoint).toLowerCase()
      return !searchTerm || sessionId.includes(searchTerm) || log.endpoint.toLowerCase().includes(searchTerm)
    })

    const grouped = groupLogsBySession(filteredLogs)

    const sessionIds = Object.keys(grouped).sort((a, b) => {
      const aLatest = grouped[a].reduce((max, log) => Math.max(max, new Date(log.timestamp).getTime()), 0)
      const bLatest = grouped[b].reduce((max, log) => Math.max(max, new Date(log.timestamp).getTime()), 0)
      return bLatest - aLatest
    })

    return (
      <Space orientation="vertical" size={8} style={{ width: '100%' }}>
        <Input
          allowClear
          size="small"
          value={logSearchByKey[key.id] ?? ''}
          onChange={e => setLogSearchByKey(prev => ({ ...prev, [key.id]: e.target.value }))}
          placeholder={t('developer.apiKeys.searchPlaceholder')}
        />
        {sessionIds.length === 0 ? (
          <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description={t('developer.apiKeys.noSessions')} />
        ) : sessionIds.map(sessionId => {
          const sessionLogs = grouped[sessionId]
          const latestTs = sessionLogs.reduce((max, log) => Math.max(max, new Date(log.timestamp).getTime()), 0)
          const isOpen = expandedSessionByKey[key.id] === sessionId
          const verificationOutcome = (keySessionOutcomes[key.id] ?? {})[sessionId]
          const normalizedOutcome = verificationOutcome ? verificationOutcome.toLowerCase() : ''
          const isFailedOutcome = normalizedOutcome === 'failed'
          const isSuccessOutcome = normalizedOutcome === 'verified'
          const outcomeLabel = isFailedOutcome
            ? 'failed'
            : isSuccessOutcome
              ? 'succeeded'
              : verificationOutcome || 'in_progress'
          const outcomeTagStyle: CSSProperties = {
            marginInlineEnd: 0,
            fontSize: 10,
            fontFamily: T.mono,
            textTransform: 'uppercase',
            letterSpacing: '0.04em',
            cursor: sessionId === 'no-session' ? 'default' : 'pointer',
          }

          return (
            <div key={sessionId} style={{ border: `1px solid ${T.border}`, borderRadius: 8, overflow: 'hidden' }}>
              <div
                style={{
                  display: 'flex',
                  justifyContent: 'space-between',
                  alignItems: 'flex-start',
                  gap: 8,
                  padding: '8px 10px',
                  background: T.surface,
                  flexWrap: 'wrap',
                }}
              >
                <div>
                  <Space size={4} wrap>
                    <Button
                      type="text"
                      size="small"
                      style={{ paddingInline: 4 }}
                      onClick={() => toggleSessionGroup(key.id, sessionId)}
                      icon={
                        isOpen
                          ? <ChevronDownIcon style={{ width: 11, height: 11 }} />
                          : <ChevronRightIcon style={{ width: 11, height: 11 }} />
                      }
                    >
                      <span style={{ fontFamily: T.mono, fontSize: 12 }}>
                        Session: {sessionId === 'no-session' ? 'unscoped request' : sessionId}
                      </span>
                    </Button>
                    {sessionId !== 'no-session' && (
                      <Button
                        type="text"
                        size="small"
                        aria-label={t('developer.apiKeys.copyKey')}
                        icon={<ClipboardDocumentIcon style={{ width: 12, height: 12, color: T.muted }} />}
                        onClick={() => copyKey(sessionId)}
                      />
                    )}
                  </Space>
                  <Space size={8} wrap style={{ marginTop: 2 }}>
                    <Typography.Text type="secondary" style={{ fontSize: 11 }}>
                      {t('developer.apiKeys.logsSummary', { count: sessionLogs.length, time: formatDateTime(latestTs) })}
                    </Typography.Text>
                    {sessionId !== 'no-session' ? (
                      <Tag
                        icon={<CodeBracketIcon style={{ width: 10, height: 10 }} />}
                        color={isFailedOutcome ? 'error' : isSuccessOutcome ? 'success' : 'default'}
                        style={outcomeTagStyle}
                        onClick={() => fetchVerificationDetail(sessionId)}
                      >
                        {verificationDetailLoading === sessionId ? 'loading...' : outcomeLabel}
                      </Tag>
                    ) : (
                      <Tag style={outcomeTagStyle}>{outcomeLabel}</Tag>
                    )}
                  </Space>
                </div>
                <Button type="link" size="small" onClick={() => toggleSessionGroup(key.id, sessionId)}>
                  {isOpen ? 'hide' : 'open'}
                </Button>
              </div>

              {isOpen && (
                <Table<ApiActivity>
                  rowKey={(record, index) => `${record.timestamp}-${index ?? 0}`}
                  columns={apiLogColumns}
                  dataSource={sessionLogs}
                  size="small"
                  pagination={false}
                  scroll={{ x: 'max-content' }}
                  onRow={record => ({ onClick: () => setSelectedLog(record), style: { cursor: 'pointer' } })}
                />
              )}
            </div>
          )
        })}
      </Space>
    )
  }

  return (
    <>
      {/* Usage strip */}
      {stats && (() => {
        const cards = [
          { label: t('developer.apiKeys.requestsThisMonth'), value: formatNumber(stats.monthly_usage) },
          { label: t('developer.apiKeys.verified'),            value: formatNumber(stats.successful_requests) },
          // Community edition has no quota — hide the limit card.
          // Service keys / operators have no quota (monthly_limit null) — show "Unlimited".
          ...(isCommunity ? [] : [
            stats.monthly_limit === null
              ? { label: t('developer.apiKeys.quota'), value: t('developer.apiKeys.unlimited') }
              : {
                  label: t('developer.apiKeys.limitRemaining'),
                  value: formatNumber(Math.max(0, stats.monthly_limit - stats.monthly_usage)),
                },
          ]),
        ]
        const colsClass = cards.length === 2 ? ' cols-2' : ''
        return (
          <div className={`stats${colsClass}`}>
            {cards.map(({ label, value }) => (
              <div key={label} className="stat">
                <div className="l">{label}</div>
                <div className="v">{value}</div>
              </div>
            ))}
          </div>
        )
      })()}

      {renderAfterStats}

      {/* New key banner — placed just above the API Keys card so it
          lands in the dev's viewport immediately after creating a key */}
      {newFullKey && (
        <div ref={newKeyBannerRef} style={{ marginBottom: 12 }}>
          <Alert
            type="success"
            showIcon
            title="Key created — copy now, it won't be shown again"
            description={
              <Typography.Text code style={{ fontSize: 13, wordBreak: 'break-all' }}>
                {newFullKey}
              </Typography.Text>
            }
            action={
              <Space orientation="vertical" size={4}>
                <Button type="primary" size="small" onClick={() => copyKey(newFullKey)}>
                  Copy
                </Button>
                <Button size="small" onClick={() => setNewFullKey(null)}>
                  Dismiss
                </Button>
              </Space>
            }
          />
        </div>
      )}

      {/* API Keys section */}
      <div className="section-title">
        <h2><span className="num">§ 02</span>{t('developer.apiKeys.sectionTitle')}</h2>
        <span className="meta">{apiKeys.length} {apiKeys.length === 1 ? t('developer.apiKeys.keyCountOne') : t('developer.apiKeys.keyCountMany')}</span>
      </div>
      <section className="card" style={{ marginBottom: 32 }}>
        <div className="card-head">
          <h3>{t('developer.apiKeys.yourKeys')}</h3>
          <span className="sub">{t('developer.apiKeys.rotateHint')}</span>
          {!isOperator && (
            <div className="right">
              <Button
                type="primary"
                size="small"
                icon={<PlusIcon style={{ width: 12, height: 12 }} />}
                onClick={() => setShowCreate(true)}
              >
                {t('developer.apiKeys.createKey')}
              </Button>
            </div>
          )}
        </div>

        <Table<ApiKey>
          rowKey="id"
          columns={keyColumns}
          dataSource={apiKeys}
          size="small"
          scroll={{ x: 1000 }}
          pagination={{ pageSize: 10, hideOnSinglePage: true, showSizeChanger: false, size: 'small' }}
          locale={{
            emptyText: <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description={t('developer.apiKeys.noKeys')} />,
          }}
          expandable={{
            showExpandColumn: false,
            expandedRowKeys: expandedKeyId ? [expandedKeyId] : [],
            expandedRowRender: key => renderKeyLogs(key),
          }}
        />
      </section>

      {/* Quick start */}
      <div className="section-title">
        <h2><span className="num">§ 03</span>{t('developer.apiKeys.quickStart')}</h2>
        <span className="meta">{t('developer.apiKeys.quickStartMeta')}</span>
      </div>
      <pre className="code-block">
        <span className="c">{t('developer.apiKeys.quickStartComment')}</span>
        {'\n'}
        {curlSnippet}
      </pre>
      <div style={{ marginTop: 12, marginBottom: 8 }}>
        <Link to="/docs" style={{ color: T.accent, fontSize: 13, fontFamily: T.mono }}>
          {t('developer.apiKeys.fullDocumentation')}
        </Link>
      </div>

      {/* Create key modal — unreachable for operators (button hidden), but guarded here too */}
      {!isOperator && (
        <CreateKeyModal
          open={showCreate && Boolean(token)}
          onClose={() => setShowCreate(false)}
          onCreated={handleCreated}
          token={token}
        />
      )}

      {/* Delete confirmation */}
      <Modal
        open={Boolean(deleteId)}
        title={t('developer.apiKeys.confirm')}
        okText={t('developer.apiKeys.confirm')}
        cancelText={t('developer.apiKeys.cancel')}
        okButtonProps={{ danger: true, loading: deleteLoading }}
        cancelButtonProps={{ disabled: deleteLoading }}
        onCancel={() => { if (!deleteLoading) setDeleteId(null) }}
        onOk={() => { if (deleteId) void handleDelete(deleteId) }}
        destroyOnHidden
      >
        {deletingKey && (
          <Space orientation="vertical" size={4}>
            <Typography.Text strong>{deletingKey.name}</Typography.Text>
            <Typography.Text code>{deletingKey.key_preview}</Typography.Text>
          </Space>
        )}
      </Modal>

      {/* API call debug modal */}
      {selectedLog && (
        <Modal
          open
          onCancel={() => setSelectedLog(null)}
          width={760}
          title={
            <Space orientation="vertical" size={2}>
              <Typography.Text type="secondary" style={{ fontSize: 11, letterSpacing: '0.08em', textTransform: 'uppercase' }}>
                {t('developer.apiKeys.debugDetails')}
              </Typography.Text>
              <span style={{ fontSize: 15, fontWeight: 600 }}>
                {selectedLog.method} {selectedLog.endpoint}
              </span>
            </Space>
          }
          footer={<Button onClick={() => setSelectedLog(null)}>{t('developer.apiKeys.close')}</Button>}
        >
          <Descriptions
            size="small"
            bordered
            column={2}
            items={[
              {
                key: 'session',
                label: t('developer.apiKeys.logSessionId'),
                children: (
                  <Typography.Text code>
                    {extractSessionId(selectedLog.endpoint) === 'no-session'
                      ? 'not detected in endpoint'
                      : extractSessionId(selectedLog.endpoint)}
                  </Typography.Text>
                ),
              },
              {
                key: 'resource',
                label: t('developer.apiKeys.logResource'),
                children: <Typography.Text code>{inferResourceLabel(selectedLog.endpoint)}</Typography.Text>,
              },
              {
                key: 'action',
                label: t('developer.apiKeys.logAction'),
                children: inferActionLabel(selectedLog.method, selectedLog.endpoint),
              },
              {
                key: 'status',
                label: t('developer.apiKeys.logStatusLatency'),
                children: (
                  <Tag color={selectedLog.status_code >= 400 ? 'error' : 'success'}>
                    {selectedLog.status_code} in {selectedLog.response_time_ms}ms
                  </Tag>
                ),
              },
              {
                key: 'timestamp',
                label: t('developer.apiKeys.logTimestamp'),
                children: <Typography.Text code>{formatDateTime(selectedLog.timestamp)}</Typography.Text>,
              },
              {
                key: 'ip',
                label: 'IP',
                children: <Typography.Text code>{selectedLog.ip_address || 'not captured'}</Typography.Text>,
              },
              {
                key: 'agent',
                label: 'User-Agent',
                span: 2,
                children: (
                  <Typography.Text code style={{ wordBreak: 'break-all' }}>
                    {selectedLog.user_agent || 'not captured'}
                  </Typography.Text>
                ),
              },
            ]}
          />
          <Alert
            style={{ marginTop: 12 }}
            type={selectedLog.status_code >= 400 ? 'error' : 'success'}
            showIcon
            title={
              selectedLog.error_message
              || (selectedLog.status_code >= 400
                ? t('developer.apiKeys.requestFailed')
                : t('developer.apiKeys.requestSuccess'))
            }
            description={t('developer.apiKeys.payloadNotCaptured')}
          />
        </Modal>
      )}

      {/* Verification detail modal */}
      <Modal
        open={Boolean(verificationDetail || verificationDetailError)}
        onCancel={closeVerificationDetail}
        footer={<Button onClick={closeVerificationDetail}>{t('developer.apiKeys.close')}</Button>}
        width={800}
        styles={{ body: { maxHeight: '70vh', overflowY: 'auto' } }}
        title={
          verificationDetailError ? (
            <Typography.Text type="danger">{t('developer.apiKeys.failedDetails')}</Typography.Text>
          ) : (
            <Space size={8} wrap>
              <Typography.Text type="secondary" style={{ fontSize: 11, letterSpacing: '0.08em', textTransform: 'uppercase' }}>
                {t('developer.apiKeys.verificationDetail')}
              </Typography.Text>
              {verificationDetail && (
                <Typography.Text code>{verificationDetail.verification_id}</Typography.Text>
              )}
              {verificationDetail && (
                <Button
                  type="text"
                  size="small"
                  aria-label={t('developer.apiKeys.copyKey')}
                  icon={<ClipboardDocumentIcon style={{ width: 14, height: 14 }} />}
                  onClick={() => copyKey(verificationDetail.verification_id)}
                />
              )}
              {verificationDetail?.is_sandbox && <Tag color="warning">{t('developer.apiKeys.sandbox')}</Tag>}
            </Space>
          )
        }
      >
        {verificationDetailError ? (
          <Alert
            type="error"
            showIcon
            title={t('developer.apiKeys.failedDetails')}
            description={verificationDetailError}
          />
        ) : verificationDetail ? (
          <Tabs
            activeKey={detailTab}
            onChange={key => setDetailTab(key as 'scores' | 'json')}
            items={[
              {
                key: 'scores',
                label: 'Scores',
                children: (
                  <div style={{ display: 'grid', gridTemplateColumns: 'repeat(2, minmax(0, 1fr))', gap: 12 }}>
                    {/* Status + step */}
                    <Card size="small" title={t('developer.apiKeys.status')}>
                      <Typography.Text code style={{ fontSize: 13 }}>{verificationDetail.status}</Typography.Text>
                      <div style={{ color: T.muted, fontSize: 12, marginTop: 4 }}>
                        Step {verificationDetail.current_step} of {verificationDetail.total_steps}
                      </div>
                    </Card>

                    {/* Final result */}
                    <Card size="small" title={t('developer.apiKeys.finalResult')}>
                      <Tag
                        color={
                          verificationDetail.final_result === 'verified'
                            ? 'success'
                            : verificationDetail.final_result === 'failed'
                              ? 'error'
                              : verificationDetail.final_result === 'manual_review'
                                ? 'warning'
                                : 'default'
                        }
                      >
                        {verificationDetail.final_result || 'In progress'}
                      </Tag>
                    </Card>

                    {/* Cross-validation */}
                    {verificationDetail.cross_validation_results && (
                      <Card size="small" title={t('developer.apiKeys.crossValidation')}>
                        <Space size={8} style={{ width: '100%' }} align="center">
                          <Progress
                            style={{ flex: 1, margin: 0 }}
                            size="small"
                            showInfo={false}
                            percent={(verificationDetail.cross_validation_results.overall_score ?? 0) * 100}
                            strokeColor={(verificationDetail.cross_validation_results.overall_score ?? 0) >= 0.8 ? T.green : T.amber}
                          />
                          <span style={monoText}>
                            {((verificationDetail.cross_validation_results.overall_score ?? 0) * 100).toFixed(0)}%
                          </span>
                        </Space>
                        <Tag
                          style={{ marginTop: 6 }}
                          color={
                            verificationDetail.cross_validation_results.verdict === 'PASS'
                              ? 'success'
                              : verificationDetail.cross_validation_results.verdict === 'FAIL'
                                ? 'error'
                                : 'warning'
                          }
                        >
                          {verificationDetail.cross_validation_results.verdict}
                        </Tag>
                      </Card>
                    )}

                    {/* Face match */}
                    {verificationDetail.face_match_results && (
                      <Card size="small" title={t('developer.apiKeys.faceMatch')}>
                        <Space size={8} style={{ width: '100%' }} align="center">
                          <Progress
                            style={{ flex: 1, margin: 0 }}
                            size="small"
                            showInfo={false}
                            percent={(verificationDetail.face_match_results.similarity_score ?? 0) * 100}
                            strokeColor={verificationDetail.face_match_results.passed ? T.green : T.red}
                          />
                          <span style={monoText}>
                            {((verificationDetail.face_match_results.similarity_score ?? 0) * 100).toFixed(1)}%
                          </span>
                        </Space>
                        <Tag style={{ marginTop: 6 }} color={verificationDetail.face_match_results.passed ? 'success' : 'error'}>
                          {verificationDetail.face_match_results.passed ? 'passed' : 'failed'}
                        </Tag>
                      </Card>
                    )}

                    {/* Liveness */}
                    {verificationDetail.liveness_results && (
                      <Card size="small" title={t('developer.apiKeys.liveness')}>
                        <Space size={8} style={{ width: '100%' }} align="center">
                          <Progress
                            style={{ flex: 1, margin: 0 }}
                            size="small"
                            showInfo={false}
                            percent={(verificationDetail.liveness_results.score ?? 0) * 100}
                            strokeColor={verificationDetail.liveness_results.passed ? T.green : T.red}
                          />
                          <span style={monoText}>
                            {((verificationDetail.liveness_results.score ?? 0) * 100).toFixed(1)}%
                          </span>
                        </Space>
                        <Tag style={{ marginTop: 6 }} color={verificationDetail.liveness_results.passed ? 'success' : 'error'}>
                          {verificationDetail.liveness_results.passed ? 'passed' : 'failed'}
                        </Tag>
                      </Card>
                    )}

                    {/* Risk score */}
                    {verificationDetail.risk_score && (
                      <Card size="small" title={t('developer.apiKeys.riskScore')}>
                        <div style={{ fontFamily: T.mono, fontSize: 20, fontWeight: 600, color: T.text }}>
                          {verificationDetail.risk_score.overall_score}
                        </div>
                        <Tag
                          color={
                            verificationDetail.risk_score.risk_level === 'low'
                              ? 'success'
                              : verificationDetail.risk_score.risk_level === 'high'
                                ? 'error'
                                : 'warning'
                          }
                        >
                          {verificationDetail.risk_score.risk_level}
                        </Tag>
                      </Card>
                    )}

                    {/* OCR data */}
                    {verificationDetail.ocr_data && (
                      <Card
                        size="small"
                        title={t('developer.apiKeys.ocrData')}
                        style={{ gridColumn: '1 / -1' }}
                      >
                        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(180px, 1fr))', gap: 8 }}>
                          {Object.entries(verificationDetail.ocr_data as Record<string, unknown>).map(([field, value]) => (
                            <div key={field}>
                              <div style={{ color: T.muted, fontSize: 10, textTransform: 'uppercase', letterSpacing: '0.06em' }}>
                                {field.replace(/_/g, ' ')}
                              </div>
                              <div style={{ fontFamily: T.mono, fontSize: 12, color: T.text, marginTop: 2 }}>
                                {renderOcrFieldValue(value)}
                              </div>
                            </div>
                          ))}
                        </div>
                      </Card>
                    )}

                    {/* Rejection info */}
                    {verificationDetail.rejection_reason && (
                      <div style={{ gridColumn: '1 / -1' }}>
                        <Alert
                          type="error"
                          showIcon
                          title={t('developer.apiKeys.rejection')}
                          description={
                            <Space orientation="vertical" size={2}>
                              <Typography.Text code style={{ color: 'inherit' }}>
                                {verificationDetail.rejection_reason}
                              </Typography.Text>
                              {verificationDetail.rejection_detail && (
                                <span style={{ fontSize: 12 }}>{verificationDetail.rejection_detail}</span>
                              )}
                            </Space>
                          }
                        />
                      </div>
                    )}

                    {/* Manual review reason */}
                    {verificationDetail.manual_review_reason && (
                      <div style={{ gridColumn: '1 / -1' }}>
                        <Alert
                          type="warning"
                          showIcon
                          title={t('developer.apiKeys.manualReview')}
                          description={
                            <Typography.Text code style={{ color: 'inherit' }}>
                              {verificationDetail.manual_review_reason}
                            </Typography.Text>
                          }
                        />
                      </div>
                    )}

                    {/* Message */}
                    {verificationDetail.message && (
                      <div style={{ gridColumn: '1 / -1' }}>
                        <Alert type="info" showIcon title={verificationDetail.message} />
                      </div>
                    )}
                  </div>
                ),
              },
              {
                key: 'json',
                label: 'Raw JSON',
                children: (
                  <div>
                    <Space style={{ width: '100%', justifyContent: 'space-between', marginBottom: 12 }} wrap>
                      <Typography.Text type="secondary" style={{ fontSize: 12 }}>
                        {t('developer.apiKeys.responseFrom')}{' '}
                        <Typography.Text code>GET /api/developer/verifications/:id</Typography.Text>
                      </Typography.Text>
                      <Button
                        size="small"
                        icon={<ClipboardDocumentIcon style={{ width: 12, height: 12 }} />}
                        onClick={() => copyKey(JSON.stringify(verificationDetail, null, 2))}
                      >
                        Copy
                      </Button>
                    </Space>
                    <pre
                      style={{
                        background: T.codeBg,
                        borderRadius: 8,
                        padding: '16px 18px',
                        margin: 0,
                        fontFamily: T.mono,
                        fontSize: 12,
                        color: T.code,
                        lineHeight: 1.6,
                        overflowX: 'auto',
                        maxHeight: 500,
                        overflowY: 'auto',
                      }}
                    >
                      {highlightJson(verificationDetail)}
                    </pre>
                  </div>
                ),
              },
            ]}
          />
        ) : null}
      </Modal>
    </>
  )
}

