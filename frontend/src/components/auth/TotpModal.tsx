import { useState, useEffect } from 'react';
import { adminApi } from '../../lib/adminApiInstance';
import type { ApiError } from '../../lib/apiClient';
import { RetryAfterError } from '../../lib/apiClient';
import { Alert, Button, Form, Input, Modal, Space, Typography } from 'antd'

interface Props {
  tempToken: string;
  onSuccess: (jwtToken: string) => void;
  onCancel: () => void;
}

export function TotpModal({ tempToken, onSuccess, onCancel }: Props) {
  const [code, setCode] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const [retryUntil, setRetryUntil] = useState<number | null>(null);
  const [countdown, setCountdown] = useState(0);

  useEffect(() => {
    if (!retryUntil) return;
    const interval = setInterval(() => {
      const remaining = Math.ceil((retryUntil - Date.now()) / 1000);
      if (remaining <= 0) {
        setRetryUntil(null);
        setCountdown(0);
        clearInterval(interval);
      } else {
        setCountdown(remaining);
      }
    }, 500);
    return () => clearInterval(interval);
  }, [retryUntil]);

  const handleSubmit = async () => {
    setLoading(true);
    setError('');
    try {
      const { data } = await adminApi.post('/auth/totp/verify', {
        temp_token: tempToken,
        totp_code: code,
      });
      if (!data.token || typeof data.token !== 'string') {
        throw { message: 'Invalid response from server' } as ApiError;
      }
      onSuccess(data.token);
    } catch (err) {
      if (err instanceof RetryAfterError) {
        setRetryUntil(Date.now() + err.retryAfter * 1000);
        setCountdown(err.retryAfter);
        setError(`Too many attempts. Wait ${err.retryAfter}s.`);
        return;
      }
      const apiError = err as ApiError;
      setError(apiError.message ?? 'Invalid code. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  const disabled = code.length !== 6 || loading || Date.now() < (retryUntil ?? 0);

  return (
    <Modal
      open
      title="Two-Factor Authentication"
      onCancel={onCancel}
      footer={null}
      width={380}
      destroyOnHidden
    >
      <Typography.Paragraph type="secondary">
        Enter the 6-digit code from your authenticator app.
      </Typography.Paragraph>
      <Form layout="vertical" requiredMark={false} onFinish={handleSubmit}>
        <Form.Item>
          <Input.OTP
            length={6}
            value={code}
            onChange={(value) => setCode(value.replace(/\D/g, ''))}
            autoFocus
            style={{ width: '100%' }}
          />
        </Form.Item>
        {error && <Alert type="error" showIcon title={error} style={{ marginBottom: 16 }} />}
        <Space style={{ width: '100%' }}>
          <Button block onClick={onCancel}>Cancel</Button>
          <Button type="primary" htmlType="submit" block disabled={disabled} loading={loading}>
            {countdown > 0 ? `Wait ${countdown}s` : 'Verify'}
          </Button>
        </Space>
      </Form>
    </Modal>
  );
}
