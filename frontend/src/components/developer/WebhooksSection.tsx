import { useEffect, useState } from 'react'
import type { CSSProperties } from 'react';
import { useTranslation } from 'react-i18next'
import toast from 'react-hot-toast'
import {
  Alert,
  Button,
  Card,
  Checkbox,
  Collapse,
  Descriptions,
  Drawer,
  Empty,
  Form,
  Input,
  Modal,
  Segmented,
  Select,
  Space,
  Spin,
  Table,
  Tag,
  Typography,
} from 'antd'
import type { ColumnsType } from 'antd/es/table'
import { CopyOutlined, DeleteOutlined, SendOutlined } from '@ant-design/icons'
import { API_BASE_URL } from '../../config/api'
import { csrfHeader } from '../../lib/csrf'
import { C } from '../../theme'
import type { ApiKey, DeveloperWebhook, WebhookDeliveryLog } from './types'

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
import { formatDate, formatDateTime, formatTime } from '../../i18n/formatters'
import {
  highlightJson,
  copyToClipboard,
  WEBHOOK_EVENTS,
  WEBHOOK_EVENT_KEYS,
  WEBHOOK_EVENT_NAMES,
  groupDevDeliveries,
  getDevLifecycleStatus,
} from './types'

interface WebhooksSectionProps {
  token: string
  apiKeys: ApiKey[]
}

interface WebhookFormValues {
  url: string
  apiKeyId?: string | null
  events?: string[]
}

type DeliveryViewMode = 'chronological' | 'grouped'

/** Severity bucket for an event name, mirrored from the previous UI. */
function eventSeverity(eventName: string): { label: string; color: string } {
  if (eventName.includes('failed') || eventName.includes('declined')) return { label: 'fail', color: 'red' }
  if (eventName.includes('approved') || eventName.includes('completed') || eventName.includes('processed')) {
    return { label: 'pass', color: 'green' }
  }
  if (eventName.includes('manual') || eventName.includes('review')) return { label: 'review', color: 'gold' }
  return { label: 'info', color: 'default' }
}

function deliveryStatusColor(delivery: WebhookDeliveryLog): string {
  if (delivery.status === 'delivered') return 'success'
  if (delivery.status === 'failed') return 'error'
  return 'default'
}

function responseStatusColor(status: number): string {
  if (status < 300) return 'success'
  if (status < 500) return 'warning'
  return 'error'
}

const preStyle: CSSProperties = {
  margin: 0,
  padding: 12,
  background: C.codeBg,
  border: `1px solid ${T.border}`,
  borderRadius: 6,
  fontSize: 11,
  lineHeight: 1.5,
  fontFamily: C.mono,
  color: C.text,
  overflowX: 'auto',
  maxHeight: 260,
  whiteSpace: 'pre',
}

const sectionLabelStyle: CSSProperties = {
  fontFamily: T.mono,
  fontSize: 11,
  letterSpacing: '0.06em',
  textTransform: 'uppercase',
}

export function WebhooksSection({ token, apiKeys }: WebhooksSectionProps) {
  const { t } = useTranslation()
  const [form] = Form.useForm<WebhookFormValues>()
  const authHeaders = (token === 'session' ? {} : { Authorization: `Bearer ${token}` }) as Record<string, string>

  const [webhooks, setWebhooks] = useState<DeveloperWebhook[]>([])
  const [listLoading, setListLoading] = useState(false)
  const [listError, setListError] = useState<string | null>(null)
  const [addLoading, setAddLoading] = useState(false)
  const [addError, setAddError] = useState<string | null>(null)
  const [revealedSecrets, setRevealedSecrets] = useState<Record<string, string>>({})
  const [testingWebhookId, setTestingWebhookId] = useState<string | null>(null)
  const [confirmTarget, setConfirmTarget] = useState<DeveloperWebhook | null>(null)
  const [removingId, setRemovingId] = useState<string | null>(null)

  // Delivery log drawer state
  const [drawerHookId, setDrawerHookId] = useState<string | null>(null)
  const [webhookDeliveries, setWebhookDeliveries] = useState<Record<string, WebhookDeliveryLog[]>>({})
  const [deliveriesLoading, setDeliveriesLoading] = useState<string | null>(null)
  const [deliveryViewMode, setDeliveryViewMode] = useState<Record<string, DeliveryViewMode>>({})
  const [expandedSessionGroups, setExpandedSessionGroups] = useState<Record<string, Set<string>>>({})

  // Single delivery detail modal
  const [detail, setDetail] = useState<{ webhookId: string; delivery: WebhookDeliveryLog } | null>(null)
  const [resendingDeliveryId, setResendingDeliveryId] = useState<string | null>(null)

  // Fall back to the form's initial (all events) before the field registers.
  const selectedEvents = (Form.useWatch('events', form) as string[] | undefined) ?? [...WEBHOOK_EVENT_NAMES]
  const activeHook = webhooks.find(hook => hook.id === drawerHookId) ?? null
  const activeDeliveries = drawerHookId ? webhookDeliveries[drawerHookId] ?? [] : []
  const activeViewMode: DeliveryViewMode = drawerHookId ? deliveryViewMode[drawerHookId] ?? 'grouped' : 'grouped'

  const loadWebhooks = async () => {
    if (!token) return
    setListLoading(true)
    setListError(null)
    try {
      const res = await fetch(`${API_BASE_URL}/api/developer/webhooks`, {
        headers: authHeaders,
        credentials: 'include' as RequestCredentials,
      })
      if (!res.ok) throw new Error('Failed to load webhooks')
      const data = await res.json()
      setWebhooks(data.webhooks ?? [])
    } catch {
      setListError(t('developer.common.networkError'))
    } finally {
      setListLoading(false)
    }
  }

  // Fetch webhooks on mount / token change
  // eslint-disable-next-line react-hooks/exhaustive-deps
  useEffect(() => { void loadWebhooks() }, [token])

  const copyKey = async (key: string) => {
    const ok = await copyToClipboard(key)
    if (ok) toast.success(t('developer.common.copied'))
    else toast.error(t('developer.common.copyFailedManual'))
  }

  const addWebhook = async (values: WebhookFormValues) => {
    if (!token) return
    const url = (values.url ?? '').trim()
    const events = values.events ?? []
    if (!url) {
      toast.error(t('developer.webhooks.enterUrl'))
      return
    }
    if (events.length === 0) {
      toast.error(t('developer.webhooks.selectEvent'))
      return
    }

    setAddLoading(true)
    setAddError(null)
    try {
      const res = await fetch(`${API_BASE_URL}/api/developer/webhooks`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', ...authHeaders, ...csrfHeader() },
        credentials: 'include' as RequestCredentials,
        body: JSON.stringify({ url, events, api_key_id: values.apiKeyId || null }),
      })
      const data = await res.json()
      if (!res.ok) throw new Error(data.message || 'Failed to add webhook')
      setWebhooks(prev => [data.webhook, ...prev])
      // Store the full secret so the user can copy it right after creation
      if (data.webhook?.secret_key) {
        setRevealedSecrets(prev => ({ ...prev, [data.webhook.id]: data.webhook.secret_key }))
      }
      form.resetFields()
      toast.success(t('developer.webhooks.added'))
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : t('developer.webhooks.addFailed')
      setAddError(message)
      toast.error(message)
    } finally {
      setAddLoading(false)
    }
  }

  const removeWebhook = async (id: string) => {
    if (!token) return
    setRemovingId(id)
    try {
      const res = await fetch(`${API_BASE_URL}/api/developer/webhooks/${id}`, {
        method: 'DELETE',
        headers: { ...authHeaders, ...csrfHeader() },
        credentials: 'include' as RequestCredentials,
      })
      if (!res.ok) throw new Error('Failed to remove webhook')
      setWebhooks(prev => prev.filter(webhook => webhook.id !== id))
      setRevealedSecrets(prev => {
        const copy = { ...prev }
        delete copy[id]
        return copy
      })
      if (drawerHookId === id) setDrawerHookId(null)
      toast.success(t('developer.webhooks.removed'))
    } catch {
      toast.error(t('developer.webhooks.removeFailed'))
    } finally {
      setRemovingId(null)
    }
  }

  const testWebhook = async (id: string) => {
    if (!token) return
    setTestingWebhookId(id)
    const controller = new AbortController()
    const timer = setTimeout(() => controller.abort(), 15000)
    try {
      const res = await fetch(`${API_BASE_URL}/api/developer/webhooks/${id}/test`, {
        method: 'POST',
        headers: { ...authHeaders, ...csrfHeader() },
        credentials: 'include' as RequestCredentials,
        signal: controller.signal,
      })
      clearTimeout(timer)
      const data = await res.json()
      if (data.success) {
        toast.success(t('developer.webhooks.testDelivered', { status: data.status_code }))
      } else {
        toast.error(data.error || t('developer.webhooks.testFailed', { status: data.status_code }))
      }
    } catch (err: unknown) {
      clearTimeout(timer)
      const isAbort = err instanceof DOMException && err.name === 'AbortError'
      toast.error(isAbort ? t('developer.webhooks.testTimeout') : t('developer.webhooks.testFailedPlain'))
    } finally {
      setTestingWebhookId(null)
    }
  }

  const revealSecret = async (id: string) => {
    if (revealedSecrets[id]) {
      setRevealedSecrets(prev => {
        const copy = { ...prev }
        delete copy[id]
        return copy
      })
      return
    }
    if (!token) return
    try {
      const res = await fetch(`${API_BASE_URL}/api/developer/webhooks/${id}/secret`, {
        headers: authHeaders,
        credentials: 'include' as RequestCredentials,
      })
      const data = await res.json()
      if (res.ok && data.secret_key) {
        setRevealedSecrets(prev => ({ ...prev, [id]: data.secret_key }))
      } else {
        toast.error(t('developer.webhooks.revealFailed'))
      }
    } catch {
      toast.error(t('developer.webhooks.revealFailed'))
    }
  }

  const loadDeliveries = async (webhookId: string) => {
    if (!token) return
    setDeliveriesLoading(webhookId)
    try {
      const res = await fetch(`${API_BASE_URL}/api/developer/webhooks/${webhookId}/deliveries`, {
        headers: authHeaders,
        credentials: 'include' as RequestCredentials,
      })
      if (res.ok) {
        const data = await res.json()
        setWebhookDeliveries(prev => ({ ...prev, [webhookId]: data.deliveries ?? [] }))
      }
    } catch {
      /* network error */
    } finally {
      setDeliveriesLoading(null)
    }
  }

  const openDeliveries = async (webhookId: string) => {
    setDrawerHookId(webhookId)
    if (webhookDeliveries[webhookId]) return // already fetched
    await loadDeliveries(webhookId)
  }

  const resendDelivery = async (webhookId: string, deliveryId: string) => {
    if (!token) return
    setResendingDeliveryId(deliveryId)
    try {
      const res = await fetch(`${API_BASE_URL}/api/developer/webhooks/${webhookId}/deliveries/${deliveryId}/resend`, {
        method: 'POST',
        headers: { ...authHeaders, ...csrfHeader() },
        credentials: 'include' as RequestCredentials,
      })
      if (res.ok) {
        toast.success(t('developer.webhooks.resent'))
        // Refresh deliveries for this webhook
        const listRes = await fetch(`${API_BASE_URL}/api/developer/webhooks/${webhookId}/deliveries`, {
          headers: authHeaders,
          credentials: 'include' as RequestCredentials,
        })
        if (listRes.ok) {
          const data = await listRes.json()
          setWebhookDeliveries(prev => ({ ...prev, [webhookId]: data.deliveries ?? [] }))
        }
      } else {
        toast.error(t('developer.webhooks.resendFailed'))
      }
    } catch {
      toast.error(t('developer.webhooks.resendFailed'))
    } finally {
      setResendingDeliveryId(null)
    }
  }

  const deliveryColumns = (compact: boolean): ColumnsType<WebhookDeliveryLog> => [
    {
      title: t('developer.webhooks.events'),
      dataIndex: 'event',
      key: 'event',
      render: (event: string | null) => (
        <span style={{ fontFamily: T.mono, fontSize: 11 }}>{event || 'unknown'}</span>
      ),
    },
    {
      title: t('developer.webhooks.response'),
      key: 'status',
      width: 120,
      render: (_, delivery) => (
        <Tag color={deliveryStatusColor(delivery)}>
          {delivery.status === 'delivered' ? delivery.response_status ?? 'delivered' : delivery.status}
        </Tag>
      ),
    },
    {
      title: 'Time',
      dataIndex: 'created_at',
      key: 'created_at',
      width: compact ? 90 : 180,
      render: (value: string) => (
        <Typography.Text type="secondary" style={{ fontSize: 11, whiteSpace: 'nowrap' }}>
          {compact ? formatTime(value) : formatDateTime(value)}
        </Typography.Text>
      ),
    },
    {
      title: '',
      key: 'actions',
      width: 110,
      render: (_, delivery) => (
        (delivery.status === 'failed' || delivery.status === 'pending') && delivery.payload ? (
          <Button
            size="small"
            type="link"
            loading={resendingDeliveryId === delivery.id}
            onClick={event => {
              event.stopPropagation()
              void resendDelivery(drawerHookId ?? '', delivery.id)
            }}
          >
            {resendingDeliveryId === delivery.id ? 'Resending…' : 'Resend'}
          </Button>
        ) : null
      ),
    },
  ]

  const webhookColumns: ColumnsType<DeveloperWebhook> = [
    {
      title: t('developer.webhooks.endpointUrl'),
      dataIndex: 'url',
      key: 'url',
      width: 320,
      render: (url: string) => (
        <Typography.Text
          style={{ fontFamily: T.mono, fontSize: 12 }}
          ellipsis={{ tooltip: url }}
          copyable={{ text: url, tooltips: false }}
        >
          {url}
        </Typography.Text>
      ),
    },
    {
      title: t('developer.webhooks.scopeToKey'),
      key: 'api_key',
      width: 180,
      render: (_, hook) => (
        hook.api_key_id
          ? <Tag color="blue">{hook.api_key_preview || 'scoped'}</Tag>
          : <Tag>{t('developer.webhooks.allKeysPill')}</Tag>
      ),
    },
    {
      title: t('developer.webhooks.events'),
      key: 'events',
      width: 280,
      render: (_, hook) => {
        const events = hook.events ?? []
        return (
          <Space size={[0, 4]} wrap>
            {events.slice(0, 3).map(eventName => (
              <Tag key={`${hook.id}-${eventName}`} style={{ fontFamily: T.mono, fontSize: 10, marginInlineEnd: 0 }}>
                {eventName}
              </Tag>
            ))}
            {events.length > 3 && <Tag style={{ marginInlineEnd: 0 }}>+{events.length - 3}</Tag>}
          </Space>
        )
      },
    },
    {
      title: 'Status',
      key: 'status',
      width: 140,
      render: (_, hook) => (
        <Space size={4} wrap>
          <Tag color={hook.is_active ? 'success' : 'default'}>{hook.is_active ? 'Active' : 'Paused'}</Tag>
          {hook.is_sandbox && <Tag color="purple">sandbox</Tag>}
        </Space>
      ),
    },
    {
      title: 'Created',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 130,
      render: (value: string) => (
        <span style={{ fontFamily: T.mono, fontSize: 11, color: T.muted }}>{formatDate(value)}</span>
      ),
    },
    {
      title: '',
      key: 'actions',
      width: 280,
      render: (_, hook) => (
        <Space size={4} wrap>
          <Button
            size="small"
            icon={<SendOutlined />}
            loading={testingWebhookId === hook.id}
            onClick={() => void testWebhook(hook.id)}
          >
            Test
          </Button>
          <Button size="small" onClick={() => void openDeliveries(hook.id)}>Deliveries</Button>
          <Button
            size="small"
            danger
            icon={<DeleteOutlined />}
            onClick={() => setConfirmTarget(hook)}
          >
            Remove
          </Button>
        </Space>
      ),
    },
  ]

  const groups = groupDevDeliveries(activeDeliveries)
  const defaultExpandedGroups = new Set(groups.slice(0, 3).map(group => group.groupId))
  const activeGroupKeys = drawerHookId
    ? [...(expandedSessionGroups[drawerHookId] ?? defaultExpandedGroups)]
    : []

  return (
    <>
      {/* § 04 Webhooks — configure */}
      <div className="section-title">
        <h2><span className="num">§ 04</span>{t('developer.webhooks.sectionTitle')}</h2>
        <span className="meta">{webhooks.length} {t('developer.webhooks.endpoints')} {t('developer.webhooks.signedIdempotent')}</span>
      </div>

      <Card
        title={t('developer.webhooks.configure')}
        extra={
          <Typography.Text type="secondary" style={{ fontFamily: T.mono, fontSize: 11 }}>
            {t('developer.webhooks.subtitle')}
          </Typography.Text>
        }
        style={{ marginBottom: 24 }}
      >
        {addError && (
          <Alert
            type="error"
            showIcon
            title={addError}
            closable={{ onClose: () => setAddError(null) }}
            style={{ marginBottom: 16 }}
          />
        )}

        <Form
          form={form}
          layout="vertical"
          initialValues={{ url: '', apiKeyId: '', events: [...WEBHOOK_EVENT_NAMES] }}
          onFinish={addWebhook}
        >
          <Form.Item
            label={t('developer.webhooks.endpointUrl')}
            name="url"
            rules={[
              { required: true, whitespace: true, message: t('developer.webhooks.enterUrl') },
              { type: 'url', message: t('developer.webhooks.httpsTls') },
            ]}
          >
            <Input
              placeholder="https://yourapp.com/webhook"
              prefix="POST"
              style={{ fontFamily: T.mono }}
            />
          </Form.Item>

          <Form.Item
            label={t('developer.webhooks.scopeToKey')}
            name="apiKeyId"
            extra={t('developer.webhooks.limitKeys')}
          >
            <Select
              allowClear
              placeholder={t('developer.webhooks.allKeys')}
              options={[
                { value: '', label: t('developer.webhooks.allKeys') },
                ...apiKeys
                  .filter(key => key.status === 'active')
                  .map(key => ({ value: key.id, label: `${key.key_preview} — ${key.name}` })),
              ]}
            />
          </Form.Item>

          <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', gap: 8, marginBottom: 8 }}>
            <Typography.Text type="secondary" style={{ fontFamily: T.mono, fontSize: 11 }}>
              {t('developer.webhooks.eventsLine', { selected: selectedEvents.length, total: WEBHOOK_EVENT_NAMES.length })}
            </Typography.Text>
            <Space size={0}>
              <Button
                type="link"
                size="small"
                onClick={() => form.setFieldValue('events', [...WEBHOOK_EVENT_NAMES])}
              >
                {t('developer.webhooks.selectAll')}
              </Button>
              <Button type="link" size="small" onClick={() => form.setFieldValue('events', [])}>
                {t('developer.webhooks.clear')}
              </Button>
            </Space>
          </div>

          <Form.Item
            name="events"
            rules={[{ required: true, message: t('developer.webhooks.selectEvent') }]}
          >
            <Checkbox.Group style={{ width: '100%' }}>
              <Space orientation="vertical" size={4} style={{ width: '100%' }}>
                {WEBHOOK_EVENT_NAMES.map(eventName => {
                  const severity = eventSeverity(eventName)
                  const description = WEBHOOK_EVENT_KEYS[eventName]
                    ? t(`developer.webhooks.webhookEvents.${WEBHOOK_EVENT_KEYS[eventName]}`)
                    : WEBHOOK_EVENTS[eventName]
                  return (
                    <Checkbox key={eventName} value={eventName}>
                      <Space size={8} wrap>
                        <span style={{ fontFamily: T.mono, fontSize: 12 }}>{eventName}</span>
                        <Typography.Text type="secondary" style={{ fontSize: 11 }}>
                          {description}
                        </Typography.Text>
                        <Tag color={severity.color} style={{ marginInlineEnd: 0, fontSize: 10 }}>
                          {severity.label}
                        </Tag>
                      </Space>
                    </Checkbox>
                  )
                })}
              </Space>
            </Checkbox.Group>
          </Form.Item>

          <div style={{ display: 'flex', justifyContent: 'flex-end' }}>
            <Button type="primary" htmlType="submit" loading={addLoading}>
              {addLoading ? t('developer.webhooks.adding') : t('developer.webhooks.addWebhook')}
            </Button>
          </div>
        </Form>
      </Card>

      {/* § 05 Active endpoints */}
      <div className="section-title">
        <h2><span className="num">§ 05</span>{t('developer.webhooks.activeEndpoints')}</h2>
        <span className="meta">{t('developer.webhooks.inspectDeliveries')}</span>
      </div>

      {listError && (
        <Alert
          type="error"
          showIcon
          title={listError}
          action={<Button size="small" onClick={() => void loadWebhooks()}>{t('developer.webhooks.endpoints')}</Button>}
          style={{ marginBottom: 12 }}
        />
      )}

      <Card>
        <Table<DeveloperWebhook>
          rowKey="id"
          columns={webhookColumns}
          dataSource={webhooks}
          loading={listLoading}
          pagination={false}
          scroll={{ x: 1000 }}
          locale={{ emptyText: <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} /> }}
        />
      </Card>

      {/* Delivery log drawer */}
      <Drawer
        open={!!activeHook}
        onClose={() => setDrawerHookId(null)}
        size={760}
        title={
          <span style={{ fontFamily: T.mono, fontSize: 12 }}>{activeHook?.url}</span>
        }
        extra={activeHook && (
          <Space size={8}>
            <Tag color={activeHook.is_active ? 'success' : 'default'}>
              {activeHook.is_active ? 'Active' : 'Paused'}
            </Tag>
            <Button size="small" danger icon={<DeleteOutlined />} onClick={() => setConfirmTarget(activeHook)}>
              Remove
            </Button>
          </Space>
        )}
      >
        {activeHook && (
          <Descriptions
            column={1}
            size="small"
            bordered
            items={[
              {
                key: 'secret',
                label: t('developer.webhooks.signingSecret'),
                children: (
                  <Space.Compact style={{ width: '100%' }}>
                    <Input.Password
                      readOnly
                      value={revealedSecrets[activeHook.id] || activeHook.secret_key || '••••••••'}
                      visibilityToggle={{
                        visible: !!revealedSecrets[activeHook.id],
                        onVisibleChange: () => void revealSecret(activeHook.id),
                      }}
                      style={{ fontFamily: T.mono, fontSize: 12 }}
                    />
                    <Button
                      icon={<CopyOutlined />}
                      disabled={!revealedSecrets[activeHook.id]}
                      title={t('developer.webhooks.copySecret')}
                      onClick={() => void copyKey(revealedSecrets[activeHook.id])}
                    />
                  </Space.Compact>
                ),
              },
              {
                key: 'events',
                label: t('developer.webhooks.events'),
                children: (
                  <Space size={[0, 4]} wrap>
                    {(activeHook.events ?? []).length === 0
                      ? <Typography.Text type="secondary">—</Typography.Text>
                      : (activeHook.events ?? []).map(eventName => (
                        <Tag key={`${activeHook.id}-${eventName}`} style={{ marginInlineEnd: 0, fontFamily: T.mono, fontSize: 10 }}>
                          {eventName}
                        </Tag>
                      ))}
                  </Space>
                ),
              },
            ]}
          />
        )}

        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', margin: '20px 0 12px' }}>
          <span style={{ ...sectionLabelStyle, color: T.muted }}>Recent deliveries</span>
          {activeDeliveries.length > 0 && (
            <Segmented
              size="small"
              value={activeViewMode}
              onChange={value => {
                if (!drawerHookId) return
                setDeliveryViewMode(prev => ({ ...prev, [drawerHookId]: value as DeliveryViewMode }))
              }}
              options={[
                { label: 'Timeline', value: 'chronological' },
                { label: 'By session', value: 'grouped' },
              ]}
            />
          )}
        </div>

        {drawerHookId && deliveriesLoading === drawerHookId ? (
          <div style={{ padding: 32, textAlign: 'center' }}>
            <Spin />
          </div>
        ) : activeDeliveries.length === 0 ? (
          <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} />
        ) : activeViewMode === 'chronological' ? (
          <Table<WebhookDeliveryLog>
            rowKey="id"
            size="small"
            columns={deliveryColumns(false)}
            dataSource={activeDeliveries}
            pagination={false}
            onRow={delivery => ({
              onClick: () => drawerHookId && setDetail({ webhookId: drawerHookId, delivery }),
              style: { cursor: 'pointer' },
            })}
          />
        ) : (
          <Collapse
            size="small"
            activeKey={activeGroupKeys}
            onChange={keys => {
              if (!drawerHookId) return
              setExpandedSessionGroups(prev => ({
                ...prev,
                [drawerHookId]: new Set(keys as string[]),
              }))
            }}
            items={groups.map(group => {
              const lifecycle = getDevLifecycleStatus(group.deliveries)
              return {
                key: group.groupId,
                label: (
                  <Space size={8} wrap>
                    <span style={{ fontFamily: T.mono, fontSize: 12 }}>{group.label}</span>
                    <Tag style={{ marginInlineEnd: 0 }}>{group.deliveries.length}</Tag>
                    <Tag
                      style={{
                        marginInlineEnd: 0,
                        color: lifecycle.color,
                        background: lifecycle.bg,
                        borderColor: lifecycle.color,
                        fontSize: 10,
                      }}
                    >
                      {lifecycle.label}
                    </Tag>
                    <Typography.Text type="secondary" style={{ fontSize: 10 }}>{group.dateLabel}</Typography.Text>
                  </Space>
                ),
                children: (
                  <Table<WebhookDeliveryLog>
                    rowKey="id"
                    size="small"
                    columns={deliveryColumns(true)}
                    dataSource={group.deliveries}
                    pagination={false}
                    onRow={delivery => ({
                      onClick: () => drawerHookId && setDetail({ webhookId: drawerHookId, delivery }),
                      style: { cursor: 'pointer' },
                    })}
                  />
                ),
              }
            })}
          />
        )}
      </Drawer>

      {/* Single delivery detail */}
      <Modal
        open={detail !== null}
        onCancel={() => setDetail(null)}
        footer={null}
        width={720}
        title={
          <Space size={8}>
            <span style={{ fontFamily: T.mono, fontSize: 12 }}>{detail?.delivery.event || 'unknown'}</span>
            <Tag color={detail ? deliveryStatusColor(detail.delivery) : 'default'}>{detail?.delivery.status}</Tag>
          </Space>
        }
      >
        {detail && (
          <Space orientation="vertical" size={12} style={{ width: '100%' }}>
            <Typography.Text type="secondary" style={{ fontFamily: T.mono, fontSize: 11 }}>
              {formatDateTime(detail.delivery.created_at)}
            </Typography.Text>

            <div>
              <div style={{ ...sectionLabelStyle, color: T.muted, marginBottom: 6 }}>
                {t('developer.webhooks.requestPayload')}
              </div>
              {detail.delivery.payload ? (
                <pre style={preStyle}>{highlightJson(JSON.stringify(detail.delivery.payload, null, 2))}</pre>
              ) : (
                <Typography.Text type="secondary" style={{ fontSize: 11 }}>No payload captured</Typography.Text>
              )}
            </div>

            <div>
              <div style={{ ...sectionLabelStyle, color: T.muted, marginBottom: 6 }}>
                {t('developer.webhooks.response')}
                {detail.delivery.response_status && (
                  <Tag
                    color={responseStatusColor(detail.delivery.response_status)}
                    style={{ marginInlineStart: 8, marginInlineEnd: 0, fontSize: 10 }}
                  >
                    {detail.delivery.response_status}
                  </Tag>
                )}
              </div>
              <pre style={{ ...preStyle, maxHeight: 160, color: C.muted, whiteSpace: 'pre-wrap', wordBreak: 'break-all' }}>
                {detail.delivery.response_body
                  ? (() => {
                      try {
                        return highlightJson(JSON.stringify(JSON.parse(detail.delivery.response_body), null, 2))
                      } catch {
                        return detail.delivery.response_body.slice(0, 500)
                      }
                    })()
                  : 'No response captured'}
              </pre>
            </div>

            {(detail.delivery.status === 'failed' || detail.delivery.status === 'pending') && detail.delivery.payload && (
              <div style={{ display: 'flex', justifyContent: 'flex-end' }}>
                <Button
                  icon={<SendOutlined />}
                  loading={resendingDeliveryId === detail.delivery.id}
                  onClick={async () => {
                    await resendDelivery(detail.webhookId, detail.delivery.id)
                    setDetail(null)
                  }}
                >
                  {resendingDeliveryId === detail.delivery.id ? 'Resending…' : 'Resend'}
                </Button>
              </div>
            )}
          </Space>
        )}
      </Modal>

      {/* Remove confirmation */}
      <Modal
        open={confirmTarget !== null}
        title="Remove webhook"
        okButtonProps={{ danger: true }}
        confirmLoading={!!removingId}
        onCancel={() => setConfirmTarget(null)}
        onOk={async () => {
          if (!confirmTarget) return
          await removeWebhook(confirmTarget.id)
          setConfirmTarget(null)
        }}
      >
        <Typography.Text style={{ fontFamily: T.mono, fontSize: 12 }}>{confirmTarget?.url}</Typography.Text>
      </Modal>

      {/* Payload example reference */}
      <Collapse
        style={{ marginTop: 24, background: T.panel, borderColor: T.border }}
        items={[
          {
            key: 'payload',
            label: <span style={{ ...sectionLabelStyle, color: T.muted }}>{t('developer.webhooks.examplePayload')}</span>,
            children: (
              <pre className="code-block" style={{ margin: 0, border: 'none' }}>
{`{
  "user_id": "a1b2c3d4-...",
  "verification_id": "e5f6g7h8-...",
  "status": "verification.completed",
  "timestamp": "2026-03-19T12:00:00.000Z",
  "data": {
    "ocr_data": {
      "full_name": "Jane Doe",
      "date_of_birth": "1990-01-15",
      "id_number": "D1234567"
    },
    "face_match_score": 0.94
  }
}

// Verify signature (Node.js)
const crypto = require('crypto');
const expected = 'sha256=' + crypto
  .createHmac('sha256', YOUR_WEBHOOK_SECRET)
  .update(rawBody, 'utf8')
  .digest('hex');
if (expected !== req.headers['x-identity-verification-signature']) {
  throw new Error('Invalid signature');
}`}
              </pre>
            ),
          },
        ]}
      />
    </>
  )
}
