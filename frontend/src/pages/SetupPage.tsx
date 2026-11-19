import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import toast from 'react-hot-toast'
import { Alert, Button, Card, Form, Input, Result, Space, Spin, Typography } from 'antd'
import { API_BASE_URL } from '../config/api'
import { injectFonts } from '../theme'
import { useTranslation } from 'react-i18next'
import { LanguageSwitcher } from '../components/common/LanguageSwitcher'
import { getLocale } from '../i18n'

type SetupState = 'loading' | 'form' | 'success' | 'error'

interface SetupResult {
  token: string
  developer: { id: string; email: string; name: string; company: string | null }
  api_key?: { key: string; id: string; name: string; is_sandbox: boolean; created_at: string }
}

export function SetupPage() {
  const navigate = useNavigate()
  const { t } = useTranslation()
  const [state, setState] = useState<SetupState>('loading')
  const [name, setName] = useState('')
  const [email, setEmail] = useState('')
  const [company, setCompany] = useState('')
  const [loading, setLoading] = useState(false)
  const [result, setResult] = useState<SetupResult | null>(null)

  useEffect(() => { injectFonts() }, [])

  // Check if setup is needed
  useEffect(() => {
    fetch(`${API_BASE_URL}/api/setup/status`)
      .then(r => {
        if (!r.ok) throw new Error(`HTTP ${r.status}`)
        return r.json()
      })
      .then(data => {
        if (data.needs_setup) {
          setState('form')
        } else {
          navigate('/', { replace: true })
        }
      })
      .catch(() => {
        setState('error')
      })
  }, [navigate])

  const handleSubmit = async () => {
    if (!name.trim() || !email.trim()) return

    setLoading(true)
    try {
      const res = await fetch(`${API_BASE_URL}/api/setup/initialize`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', 'Accept-Language': getLocale() },
        credentials: 'include',
        body: JSON.stringify({ name: name.trim(), email: email.trim(), company: company.trim() || undefined, locale: getLocale() }),
      })

      if (!res.ok) {
        const data = await res.json().catch(() => ({}))
        throw new Error(data.message || `Setup failed (${res.status})`)
      }

      const data = await res.json()

      // Auth cookie set by server (httpOnly)
      setResult(data)
      setState('success')
      toast.success(t('setup.accountCreated'))
    } catch (err: unknown) {
      toast.error(err instanceof Error ? err.message : t('setup.failed'))
    } finally {
      setLoading(false)
    }
  }

  if (state === 'loading') {
    return (
      <div style={{ background: 'var(--paper)', minHeight: '100vh', display: 'grid', placeItems: 'center' }}>
        <Space orientation="vertical" align="center">
          <Spin size="large" />
          <Typography.Text type="secondary">{t('setup.checkingStatus')}</Typography.Text>
        </Space>
      </div>
    )
  }

  if (state === 'error') {
    return (
      <div style={{ background: 'var(--paper)', minHeight: '100vh', display: 'grid', placeItems: 'center', padding: 24 }}>
        <Result
          status="error"
          title={t('setup.couldNotReachApi')}
          subTitle={t('setup.apiUnreachableHint')}
          extra={<Button type="primary" onClick={() => window.location.reload()}>{t('setup.retry')}</Button>}
        />
      </div>
    )
  }

  if (state === 'success' && result) {
    return (
      <div style={{ background: 'var(--paper)', minHeight: '100vh', color: 'var(--ink)' }}>
        <div style={{ maxWidth: 560, margin: '0 auto', padding: '56px 24px' }}>
          <Result
            status="success"
            title={t('setup.allSet')}
            subTitle={t('setup.welcome', { name: result.developer.name })}
          />
          {result.api_key && (
            <Alert
              type="success"
              showIcon
              title={t('setup.apiKeyNotice')}
              description={
                <Space orientation="vertical" size={8} style={{ width: '100%' }}>
                  <Typography.Text code copyable style={{ wordBreak: 'break-all' }}>{result.api_key.key}</Typography.Text>
                  <Typography.Text type="secondary" style={{ fontSize: 12 }}>
                    {t('setup.typeLabel', { type: result.api_key.is_sandbox ? t('setup.sandbox') : t('setup.production'), name: result.api_key.name })}
                  </Typography.Text>
                </Space>
              }
              style={{ marginBottom: 20 }}
            />
          )}
          <Card title={t('setup.nextSteps')} style={{ marginBottom: 20 }}>
            <Typography.Paragraph>1. {t('setup.nextStep1')}</Typography.Paragraph>
            <Typography.Paragraph>
              2. {t('setup.nextStep2Before')}<Typography.Link href="/docs">{t('setup.nextStep2Link')}</Typography.Link>{t('setup.nextStep2After')}
            </Typography.Paragraph>
            <Typography.Paragraph style={{ marginBottom: 0 }}>
              3. {t('setup.nextStep3Before')}<Typography.Link href="/demo">{t('setup.nextStep3Link')}</Typography.Link>{t('setup.nextStep3After')}
            </Typography.Paragraph>
          </Card>
          <Button type="primary" block size="large" onClick={() => navigate('/', { replace: true })}>
            {t('setup.goToDashboard')}
          </Button>
        </div>
      </div>
    )
  }

  return (
    <div style={{ background: 'var(--paper)', minHeight: '100vh', color: 'var(--ink)', display: 'grid', placeItems: 'center' }}>
      <div style={{ maxWidth: 440, width: '100%', padding: 24 }}>
        <div style={{ display: 'flex', justifyContent: 'flex-end', marginBottom: 16 }}>
          <LanguageSwitcher />
        </div>
        <div style={{ textAlign: 'center', marginBottom: 24 }}>
          <img src="/identity-logo.svg" alt="Identity Verification" style={{ height: 36, marginBottom: 20 }} />
          <Typography.Title level={3} style={{ marginBottom: 8 }}>{t('setup.welcomeTitle')}</Typography.Title>
          <Typography.Paragraph type="secondary" style={{ marginBottom: 0 }}>{t('setup.welcomeDesc')}</Typography.Paragraph>
        </div>
        <Form layout="vertical" requiredMark={false} onFinish={handleSubmit}>
          <Card styles={{ body: { padding: 24 } }}>
            <Form.Item label={t('setup.nameLabel')} name="name" rules={[{ required: true }]} style={{ marginBottom: 16 }}>
              <Input size="large" value={name} onChange={event => setName(event.target.value)} placeholder={t('setup.namePlaceholder')} autoFocus />
            </Form.Item>
            <Form.Item label={t('setup.emailLabel')} name="email" rules={[{ required: true, type: 'email' }]} style={{ marginBottom: 16 }}>
              <Input size="large" type="email" value={email} onChange={event => setEmail(event.target.value)} placeholder={t('setup.emailPlaceholder')} />
            </Form.Item>
            <Form.Item label={<>{t('setup.companyLabel')} <Typography.Text type="secondary">{t('setup.optional')}</Typography.Text></>} style={{ marginBottom: 20 }}>
              <Input size="large" value={company} onChange={event => setCompany(event.target.value)} placeholder={t('setup.companyPlaceholder')} />
            </Form.Item>
            <Button type="primary" htmlType="submit" block size="large" loading={loading} disabled={!name.trim() || !email.trim()}>
              {loading ? t('setup.creatingAccount') : t('setup.createAccount')}
            </Button>
          </Card>
        </Form>
        <Typography.Paragraph type="secondary" style={{ textAlign: 'center', marginTop: 16, fontSize: 12 }}>
          {t('setup.oneTimeSetup')}
        </Typography.Paragraph>
      </div>
    </div>
  )
}
