import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { Alert, Button, Card, Form, Input, Spin, Typography } from 'antd';
import { API_BASE_URL } from '../config/api';
import { injectFonts } from '../theme';
import '../styles/patterns.css';

export const AdminLogin = () => {
  const navigate = useNavigate();
  const { t } = useTranslation();
  const [email, setEmail] = useState('');
  const [otpCode, setOtpCode] = useState('');
  const [step, setStep] = useState<'email' | 'otp'>('email');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [selfHostedCode, setSelfHostedCode] = useState<string | null>(null);
  const [checkingAuth, setCheckingAuth] = useState(false);

  useEffect(() => { injectFonts(); }, []);

  // Check existing cookie auth — no escalation, OTP-only
  useEffect(() => {
    setCheckingAuth(true);
    fetch(`${API_BASE_URL}/api/admin/dashboard`, { credentials: 'include' })
      .then(res => {
        if (res.ok) navigate('/admin/verifications');
        else setCheckingAuth(false);
      })
      .catch(() => setCheckingAuth(false));
  }, [navigate]);

  // Auto-focus is handled by antd Form/Input when the OTP step mounts.

  const handleSendOtp = async () => {
    setLoading(true);
    setError('');
    setSelfHostedCode(null);

    try {
      const res = await fetch(`${API_BASE_URL}/api/auth/reviewer/otp/send`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify({ email }),
      });
      const data = await res.json();

      if (!res.ok) {
        setError(data.message || 'Failed to send verification code');
        return;
      }

      // Self-hosted: show code inline
      if (data.self_hosted && data.code) {
        setSelfHostedCode(data.code);
      }

      setStep('otp');
    } catch (caught: unknown) {
      const err = caught as Error;
      setError(err.message ?? t('admin.couldNotReachServer'));
    } finally {
      setLoading(false);
    }
  };

  const handleVerifyOtp = async () => {
    setLoading(true);
    setError('');

    try {
      const res = await fetch(`${API_BASE_URL}/api/auth/reviewer/otp/verify`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify({ email, code: otpCode }),
      });
      const data = await res.json();

      if (!res.ok) {
        setError(data.message || 'Invalid verification code');
        return;
      }

      if (!data.token) {
        setError('Login failed: invalid response from server');
        return;
      }

      // Token is now set as httpOnly cookie by the server
      navigate('/admin/verifications');
    } catch (caught: unknown) {
      const err = caught as Error;
      setError(err.message ?? t('admin.couldNotReachServer'));
    } finally {
      setLoading(false);
    }
  };

  // Spinner while checking existing auth
  if (checkingAuth) {
    return (
      <div style={{
        minHeight: '100vh',
        background: 'var(--paper)',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        fontFamily: 'var(--sans)',
      }}>
        <div style={{ textAlign: 'center' }}>
          <Spin size="large" />
          <Typography.Paragraph type="secondary" style={{ marginTop: 16 }}>
            {t('admin.authenticating')}
          </Typography.Paragraph>
        </div>
      </div>
    );
  }

  return (
    <div className="pattern-shield pattern-faint pattern-fade-edges" style={{
      minHeight: '100vh',
      background: 'var(--paper)',
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      fontFamily: 'var(--sans)',
      position: 'relative',
      overflow: 'hidden',
    }}>
      <div style={{ width: '100%', maxWidth: 420, padding: '0 24px', position: 'relative', zIndex: 1 }}>
        <div style={{ textAlign: 'center', marginBottom: 24 }}>
          <img
            src="/identity-logo.svg"
            alt="Identity Verification"
            style={{ height: 32, margin: '0 auto 24px' }}
            onError={(event) => { event.currentTarget.style.display = 'none' }}
          />
          <Typography.Text type="secondary" style={{ display: 'block', marginBottom: 8, textTransform: 'uppercase', letterSpacing: '0.12em', fontSize: 12 }}>
            {t('admin.eyebrow')}
          </Typography.Text>
          <Typography.Title level={2} style={{ margin: '0 0 8px' }}>{t('admin.title')}</Typography.Title>
          <Typography.Paragraph type="secondary" style={{ marginBottom: 0 }}>
            {step === 'email' ? t('admin.enterReviewerEmail') : t('admin.enterCodeSentTo', { email })}
          </Typography.Paragraph>
        </div>

        <Card styles={{ body: { padding: 32 } }}>
          {error && <Alert type="error" showIcon title={error} style={{ marginBottom: 20 }} />}

          {step === 'email' ? (
            <Form layout="vertical" requiredMark={false} onFinish={handleSendOtp}>
              <Form.Item
                label={t('admin.emailAddress')}
                name="email"
                rules={[{ required: true, type: 'email', message: t('admin.emailAddress') }]}
                style={{ marginBottom: 20 }}
              >
                <Input
                  size="large"
                  type="email"
                  value={email}
                  onChange={(event) => setEmail(event.target.value)}
                  placeholder={t('admin.emailPlaceholder')}
                  autoFocus
                />
              </Form.Item>
              <Button type="primary" htmlType="submit" block size="large" loading={loading}>
                {loading ? t('admin.sendingCode') : t('admin.sendCode')}
              </Button>
            </Form>
          ) : (
            <Form layout="vertical" requiredMark={false} onFinish={handleVerifyOtp}>
              {selfHostedCode && (
                <Alert
                  type="success"
                  showIcon
                  title={t('admin.selfHostedCode')}
                  description={<Typography.Text code copyable style={{ fontSize: 20, letterSpacing: '0.12em' }}>{selfHostedCode}</Typography.Text>}
                  style={{ marginBottom: 20 }}
                />
              )}
              <Form.Item label={t('admin.verificationCode')} style={{ marginBottom: 20 }}>
                <Input.OTP
                  length={6}
                  value={otpCode}
                  onChange={(value) => setOtpCode(value.replace(/\D/g, '').slice(0, 6))}
                  autoFocus
                />
              </Form.Item>
              <Button type="primary" htmlType="submit" block size="large" loading={loading} disabled={otpCode.length !== 6}>
                {loading ? t('admin.verifying') : t('admin.verify')}
              </Button>
              <Button
                type="default"
                block
                style={{ marginTop: 12 }}
                onClick={() => { setStep('email'); setOtpCode(''); setError(''); setSelfHostedCode(null) }}
              >
                {t('admin.useDifferentEmail')}
              </Button>
            </Form>
          )}
        </Card>

        <div style={{ textAlign: 'center', marginTop: 24 }}>
          <Typography.Link href="/">{t('admin.backToPortal')}</Typography.Link>
        </div>
      </div>
    </div>
  )
}
