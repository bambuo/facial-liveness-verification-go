import { useState, useEffect } from 'react'
import { createRef, useRef } from 'react';
import type { CSSProperties, ClipboardEvent, KeyboardEvent } from 'react';
import { useTranslation } from 'react-i18next'
import toast from 'react-hot-toast'
import { Button, Card, Divider, Form, Input, Space, Typography } from 'antd'
import type { InputRef } from 'antd'
import { GithubOutlined } from '@ant-design/icons'
import { API_BASE_URL } from '../../config/api'
import { csrfHeader } from '../../lib/csrf'
import type { AuthStep } from './types'

export function AuthGate({ onAuth }: { onAuth: (token: string, apiKey?: string) => void }) {
  const { t } = useTranslation()
  const [step, setStep] = useState<AuthStep>('enter_email')
  const [email, setEmail] = useState('')
  const [otpDigits, setOtpDigits] = useState<string[]>(['', '', '', '', '', ''])
  const [registrationToken, setRegistrationToken] = useState('')
  const [name, setName] = useState('')
  const [company, setCompany] = useState('')
  const [loading, setLoading] = useState(false)
  const [resendCooldown, setResendCooldown] = useState(0)
  const [githubConfigured, setGithubConfigured] = useState(false)
  const [emailError, setEmailError] = useState('')
  const [isReturning] = useState(() => localStorage.getItem('identity:has-session') === 'true')

  const markReturning = () => localStorage.setItem('identity:has-session', 'true')

  const otpRefs = useRef(
    Array.from({ length: 6 }, () => createRef<InputRef>())
  ).current

  // Check if GitHub OAuth is configured
  useEffect(() => {
    fetch(`${API_BASE_URL}/api/auth/developer/github/url?state=check`, { credentials: 'include' })
      .then(r => r.json())
      .then(d => setGithubConfigured(d.configured === true))
      .catch(() => {})
  }, [])

  // Handle GitHub callback from URL params
  useEffect(() => {
    const params = new URLSearchParams(window.location.search)
    const code = params.get('code')
    const returnedState = params.get('state')
    if (!code) return

    // Clear URL immediately to prevent re-firing on effect re-run
    window.history.replaceState({}, '', window.location.pathname)

    // Verify OAuth state to prevent CSRF
    const storedState = sessionStorage.getItem('github_oauth_state')
    sessionStorage.removeItem('github_oauth_state')
    if (!returnedState || returnedState !== storedState) {
      toast.error(t('auth.oauthStateMismatch'))
      return
    }

    setLoading(true)
    fetch(`${API_BASE_URL}/api/auth/developer/github/callback`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', ...csrfHeader() },
      credentials: 'include',
      body: JSON.stringify({ code, state: returnedState }),
    })
      .then(r => r.json().then(data => ({ ok: r.ok, data })))
      .then(({ ok, data }) => {
        if (!ok) throw new Error(data.message || 'GitHub login failed')
        markReturning()
        onAuth('session', data.api_key?.key)
      })
      .catch((err: unknown) => {
        toast.error(err instanceof Error ? err.message : t('auth.githubFailed'))
      })
      .finally(() => setLoading(false))
  }, [onAuth])

  // Resend cooldown timer
  useEffect(() => {
    if (resendCooldown <= 0) return
    const timer = setTimeout(() => setResendCooldown(c => c - 1), 1000)
    return () => clearTimeout(timer)
  }, [resendCooldown])

  const sendOtp = async () => {
    const trimmed = email.trim()
    if (!trimmed) {
      setEmailError('Please enter your email address')
      return
    }
    if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(trimmed)) {
      setEmailError('Please enter a valid email address')
      return
    }
    setEmailError('')
    setEmail(trimmed)
    setLoading(true)
    try {
      const res = await fetch(`${API_BASE_URL}/api/auth/developer/otp/send`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', ...csrfHeader() },
        credentials: 'include',
        body: JSON.stringify({ email: trimmed }),
      })
      const data = await res.json()
      if (!res.ok) throw new Error(data.message || 'Failed to send code')
      setStep('verify_otp')
      setResendCooldown(30)

      if (data.code && data.self_hosted) {
        // Self-hosted mode: auto-fill and auto-submit the OTP
        const digits = data.code.split('')
        setOtpDigits(digits)
        toast.success(t('auth.codeAutoFilled'))
        // Small delay to let state update, then auto-verify
        setTimeout(() => verifyOtp(data.code), 300)
      } else {
        setOtpDigits(['', '', '', '', '', ''])
        toast.success(t('auth.codeSent'))
      }
    } catch (err: unknown) {
      toast.error(err instanceof Error ? err.message : t('auth.sendCodeFailed'))
    } finally {
      setLoading(false)
    }
  }

  const handleOtpDigit = (index: number, value: string) => {
    if (!/^\d?$/.test(value)) return
    const next = [...otpDigits]
    next[index] = value
    setOtpDigits(next)

    // Auto-advance
    if (value && index < 5) {
      otpRefs[index + 1].current?.focus()
    }

    // Auto-submit when all 6 digits entered
    if (value && index === 5 && next.every(d => d.length === 1)) {
      verifyOtp(next.join(''))
    }
  }

  const handleOtpKeyDown = (index: number, e: KeyboardEvent) => {
    if (e.key === 'Backspace' && !otpDigits[index] && index > 0) {
      otpRefs[index - 1].current?.focus()
    }
  }

  const handleOtpPaste = (e: ClipboardEvent) => {
    e.preventDefault()
    const pasted = e.clipboardData.getData('text').replace(/\D/g, '').slice(0, 6)
    if (pasted.length === 0) return
    const next = [...otpDigits]
    for (let i = 0; i < pasted.length; i++) {
      next[i] = pasted[i]
    }
    setOtpDigits(next)
    if (pasted.length === 6) {
      verifyOtp(pasted)
    } else {
      otpRefs[Math.min(pasted.length, 5)].current?.focus()
    }
  }

  const verifyOtp = async (code: string) => {
    setLoading(true)
    try {
      const res = await fetch(`${API_BASE_URL}/api/auth/developer/otp/verify`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', ...csrfHeader() },
        credentials: 'include',
        body: JSON.stringify({ email, code }),
      })
      const data = await res.json()
      if (!res.ok) throw new Error(data.message || 'Invalid code')

      if (data.is_new) {
        // New user — go to registration step
        setRegistrationToken(data.registration_token)
        setStep('complete_registration')
      } else {
        // Existing user — login complete (cookie set by server)
        markReturning()
        onAuth(data.token)
      }
    } catch (err: unknown) {
      toast.error(err instanceof Error ? err.message : t('auth.verificationFailed'))
      setOtpDigits(['', '', '', '', '', ''])
      otpRefs[0].current?.focus()
    } finally {
      setLoading(false)
    }
  }

  const completeRegistration = async () => {
    setLoading(true)
    try {
      const res = await fetch(`${API_BASE_URL}/api/auth/developer/otp/complete-registration`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', ...csrfHeader() },
        credentials: 'include',
        body: JSON.stringify({ registration_token: registrationToken, name, company }),
      })
      const data = await res.json()
      if (!res.ok) throw new Error(data.message || 'Registration failed')
      // Cookie set by server
      markReturning()
      onAuth(data.token, data.api_key?.key)
    } catch (err: unknown) {
      toast.error(err instanceof Error ? err.message : t('auth.registrationFailed'))
    } finally {
      setLoading(false)
    }
  }

  const startGitHub = async () => {
    setLoading(true)
    try {
      const res = await fetch(`${API_BASE_URL}/api/auth/developer/github/url`, { credentials: 'include' })
      const data = await res.json()
      if (!data.url) throw new Error('GitHub OAuth not configured')
      sessionStorage.setItem('github_oauth_state', data.state)
      window.location.href = data.url
    } catch (err: unknown) {
      toast.error(err instanceof Error ? err.message : t('auth.githubFailed'))
      setLoading(false)
    }
  }

  return (
    <div className="pattern-guilloche-rainbow pattern-full" style={{ minHeight: 'calc(100vh - 120px)', display: 'flex', alignItems: 'center', justifyContent: 'center', padding: 24, '--pattern-opacity': '0.01' } as CSSProperties}>
      <Card style={{ width: '100%', maxWidth: 400 }} styles={{ body: { padding: 40 } }}>
        <Typography.Text type="secondary" style={{ fontFamily: 'ui-monospace, Menlo, monospace', fontSize: 11, letterSpacing: '0.08em', display: 'block', marginBottom: 24 }}>
          identity-verification / developer-portal
        </Typography.Text>

        {step === 'enter_email' && (
          <>
            <Typography.Title level={4} style={{ marginTop: 0, marginBottom: 8 }}>
              {isReturning ? t('auth.signIn') : t('auth.signUp')}
            </Typography.Title>
            <Typography.Paragraph type="secondary" style={{ marginBottom: 20 }}>
              {isReturning ? t('auth.enterEmailSignIn') : t('auth.enterEmailSignUp')}
            </Typography.Paragraph>
            <Form layout="vertical" requiredMark={false} onFinish={() => sendOtp()}>
              <Form.Item
                label={t('auth.email')}
                validateStatus={emailError ? 'error' : undefined}
                help={emailError || undefined}
                style={{ marginBottom: 12 }}
              >
                <Input
                  size="large"
                  type="email"
                  value={email}
                  onChange={e => { setEmail(e.target.value); setEmailError('') }}
                  placeholder={t('auth.emailPlaceholder')}
                  autoFocus
                />
              </Form.Item>
              <Button type="primary" htmlType="submit" block size="large" loading={loading}>
                {loading ? t('auth.sendingCode') : t('auth.continueEmail')}
              </Button>
            </Form>

            {githubConfigured && (
              <>
                <Divider plain>{t('auth.or')}</Divider>
                <Button block size="large" icon={<GithubOutlined />} onClick={startGitHub} disabled={loading}>
                  {t('auth.continueGitHub')}
                </Button>
              </>
            )}

            {!isReturning && (
              <Typography.Paragraph type="secondary" style={{ textAlign: 'center', marginTop: 16, marginBottom: 0 }}>
                {t('auth.alreadyHaveAccount')}
              </Typography.Paragraph>
            )}
          </>
        )}

        {step === 'verify_otp' && (
          <>
            <Typography.Title level={4} style={{ marginTop: 0, marginBottom: 8 }}>
              {t('auth.checkEmail')}
            </Typography.Title>
            <Typography.Paragraph type="secondary" style={{ marginBottom: 20 }}>
              {t('auth.sentCodeTo', { email })}
            </Typography.Paragraph>
            <Form layout="vertical" requiredMark={false} onFinish={() => verifyOtp(otpDigits.join(''))}>
              <Form.Item style={{ marginBottom: 16 }}>
                <div style={{ display: 'flex', gap: 8, justifyContent: 'center' }} onPaste={handleOtpPaste}>
                  {otpDigits.map((digit, i) => (
                    <Input
                      key={i}
                      ref={otpRefs[i]}
                      inputMode="numeric"
                      maxLength={1}
                      value={digit}
                      onChange={e => handleOtpDigit(i, e.target.value)}
                      onKeyDown={e => handleOtpKeyDown(i, e)}
                      autoFocus={i === 0}
                      disabled={loading}
                      style={{ width: 44, height: 52, textAlign: 'center', fontSize: 22, fontWeight: 600 }}
                    />
                  ))}
                </div>
              </Form.Item>
              <Button type="primary" htmlType="submit" block size="large" loading={loading} disabled={otpDigits.some(digit => !digit)}>
                {loading ? t('auth.verifying') : t('auth.verify')}
              </Button>
            </Form>
            <Space style={{ width: '100%', justifyContent: 'space-between', marginTop: 12 }}>
              <Button type="link" onClick={() => { setStep('enter_email'); setOtpDigits(['', '', '', '', '', '']) }}>
                {t('auth.useDifferentEmail')}
              </Button>
              <Button type="link" onClick={() => sendOtp()} disabled={resendCooldown > 0 || loading}>
                {resendCooldown > 0 ? t('auth.resendIn', { seconds: resendCooldown }) : t('auth.resendCode')}
              </Button>
            </Space>
          </>
        )}

        {step === 'complete_registration' && (
          <>
            <Typography.Title level={4} style={{ marginTop: 0, marginBottom: 8 }}>
              {t('auth.createYourAccount')}
            </Typography.Title>
            <Typography.Paragraph type="secondary" style={{ marginBottom: 20 }}>
              {t('auth.almostThere')}
            </Typography.Paragraph>
            <Form layout="vertical" requiredMark={false} onFinish={completeRegistration}>
              <Form.Item label={t('auth.name')} style={{ marginBottom: 12 }}>
                <Input size="large" value={name} onChange={e => setName(e.target.value)} placeholder={t('auth.namePlaceholder')} autoFocus />
              </Form.Item>
              <Form.Item label={t('auth.companyOptional')} style={{ marginBottom: 16 }}>
                <Input size="large" value={company} onChange={e => setCompany(e.target.value)} placeholder={t('auth.companyPlaceholder')} />
              </Form.Item>
              <Button type="primary" htmlType="submit" block size="large" loading={loading}>
                {loading ? t('auth.creatingAccount') : t('auth.createAccount')}
              </Button>
            </Form>
          </>
        )}
      </Card>
    </div>
  )
}
