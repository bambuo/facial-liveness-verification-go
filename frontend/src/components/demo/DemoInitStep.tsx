import { useTranslation } from 'react-i18next';
import { D } from './themeTokens';
import { ContinueOnPhone } from '../ContinueOnPhone';
import { Button, Input, Select, Space, Typography } from 'antd'

type VerificationMode = 'full' | 'document_only' | 'identity' | 'age_only';

interface DemoInitStepProps {
  apiKey: string;
  userId: string;
  isLoading: boolean;
  isMobile: boolean;
  mobileHandoffDone: boolean;
  mobileResult: any;
  verificationMode: VerificationMode;
  ageThreshold: number;
  onApiKeyChange: (key: string) => void;
  onFillDeveloperKey: () => void;
  fillingKey: boolean;
  onUserIdChange: (id: string) => void;
  onVerificationModeChange: (mode: VerificationMode) => void;
  onAgeThresholdChange: (threshold: number) => void;
  onStart: () => void;
  onMobileHandoffDone: (done: boolean) => void;
  onMobileResult: (result: any) => void;
  onMobileVerificationComplete: (verificationId: string) => void;
}

export const DemoInitStep = ({
  apiKey,
  userId,
  isLoading,
  isMobile,
  mobileHandoffDone,
  mobileResult,
  verificationMode,
  ageThreshold,
  onApiKeyChange,
  onFillDeveloperKey,
  fillingKey,
  onUserIdChange,
  onVerificationModeChange,
  onAgeThresholdChange,
  onStart,
  onMobileHandoffDone,
  onMobileResult,
  onMobileVerificationComplete,
}: DemoInitStepProps) => {
  const { t } = useTranslation();
  return (
    <div style={{ padding: '8px 0' }}>
      {mobileHandoffDone ? (
        <div style={{ maxWidth: 400, margin: '0 auto', textAlign: 'center', padding: '32px 0' }}>
          <div style={{
            width: 56, height: 56, margin: '0 auto 16px',
            display: 'flex', alignItems: 'center', justifyContent: 'center',
            background: mobileResult?.status === 'verified' || mobileResult?.status === 'completed' ? D.greenDim : D.redDim,
            border: `1px solid ${mobileResult?.status === 'verified' || mobileResult?.status === 'completed' ? D.green : D.red}`,
            fontSize: 24,
            color: mobileResult?.status === 'verified' || mobileResult?.status === 'completed' ? D.green : D.red,
          }}>
            {mobileResult?.status === 'verified' || mobileResult?.status === 'completed' ? '\u2713' : '\u2717'}
          </div>
          <h2 style={{ fontSize: 20, fontWeight: 600, color: D.text, marginBottom: 6 }}>
            {mobileResult?.status === 'verified' || mobileResult?.status === 'completed' ? 'Verification Complete' : mobileResult?.status === 'failed' ? 'Verification Failed' : 'Under Review'}
          </h2>
          <p style={{ color: D.muted, fontSize: 13 }}>{t('demo.init.completedMobile')}</p>
          {mobileResult?.confidence_score != null && (
            <p style={{ color: D.muted, fontSize: 13, marginTop: 8 }}>
              Confidence: {Math.round(mobileResult.confidence_score * 100)}%
            </p>
          )}
          {(mobileResult?.status === 'failed' || mobileResult?.status === 'manual_review') && (
            <Button
              onClick={() => { onMobileHandoffDone(false); onMobileResult(null); }}
              style={{ marginTop: 16, background: 'none', border: 'none', color: D.accentInk, cursor: 'pointer', fontSize: 13 }}
            >
              {t('demo.init.tryAgain')}
            </Button>
          )}
        </div>
      ) : (
        <>
          <h2 style={{ fontSize: 18, fontWeight: 600, color: D.text, textAlign: 'center', marginBottom: 6 }}>
            {t('demo.init.title')}
          </h2>
          <p style={{ color: D.muted, fontSize: 13, textAlign: 'center', marginBottom: 28 }}>
            {t('demo.init.subtitle')}
          </p>

          <div style={{ width: '100%', display: 'flex', flexDirection: 'column', gap: 16, marginBottom: 80 }}>
            <div style={{ width: '100%', maxWidth: 420, margin: '0 auto' }}>
              <label style={{ display: 'block', fontFamily: D.mono, fontSize: 11, letterSpacing: '0.05em', textTransform: 'uppercase' as const, color: D.muted, marginBottom: 8 }}>
                {t('demo.init.apiKey')}
              </label>
              <Space.Compact block>
                <Input
                  type="text"
                  value={apiKey}
                  onChange={(e) => onApiKeyChange(e.target.value)}
                  placeholder={t('demo.init.apiKeyPlaceholder')}
                  style={{ background: D.panel, color: D.text, width: 'calc(100% - 84px)', fontSize: 14, fontFamily: D.sans }}
                />
                <Button
                  onClick={onFillDeveloperKey}
                  disabled={fillingKey}
                  loading={fillingKey}
                  style={{ width: 84, paddingInline: 6, color: D.accentInk, fontFamily: D.mono, fontSize: 12, whiteSpace: 'nowrap' }}
                >
                  {fillingKey ? t('demo.init.fillingKey') : t('demo.init.useDeveloperKey')}
                </Button>
              </Space.Compact>
              <p style={{ marginTop: 4, fontSize: 11, color: D.dim }}>
                {t('demo.init.getKeyPrefix')}<a href="/developer" style={{ color: D.accentInk, textDecoration: 'none' }}>{t('demo.init.developerPage')}</a>
              </p>
            </div>
            <div style={{ width: '100%', maxWidth: 420, margin: '0 auto' }}>
              <label style={{ display: 'block', fontFamily: D.mono, fontSize: 11, letterSpacing: '0.05em', textTransform: 'uppercase' as const, color: D.muted, marginBottom: 8 }}>
                {t('demo.init.userId')}
              </label>
              <Input
                type="text"
                value={userId}
                onChange={(e) => onUserIdChange(e.target.value)}
                placeholder={t('demo.init.userIdPlaceholder')}
                style={{ background: D.panel, border: `1px solid ${D.borderStrong}`, color: D.text, padding: '10px 14px', width: '100%', fontSize: 14, fontFamily: D.sans, outline: 'none', boxSizing: 'border-box' }}
              />
            </div>
            <div>
              <label style={{ display: 'block', fontFamily: D.mono, fontSize: 11, letterSpacing: '0.05em', textTransform: 'uppercase' as const, color: D.muted, marginBottom: 8 }}>
                {t('demo.init.mode')}
              </label>
              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 16 }}>
                {([
                  { mode: 'full' as VerificationMode, label: t('demo.init.full'), desc: t('demo.init.fullDesc') },
                  { mode: 'document_only' as VerificationMode, label: t('demo.init.documentOnly'), desc: t('demo.init.documentOnlyDesc') },
                  { mode: 'identity' as VerificationMode, label: t('demo.init.identity'), desc: t('demo.init.identityDesc') },
                  { mode: 'age_only' as VerificationMode, label: t('demo.init.ageOnly'), desc: t('demo.init.ageOnlyDesc') },
                ]).map(({ mode, label, desc }) => (
                  <Button
                    key={mode}
                    block
                    aria-pressed={verificationMode === mode}
                    onClick={() => onVerificationModeChange(mode)}
                    style={{
                      minHeight: 64,
                      height: 'auto',
                      width: '100%',
                      display: 'flex',
                      flexDirection: 'column',
                      alignItems: 'flex-start',
                      justifyContent: 'center',
                      gap: 2,
                      padding: '10px 22px',
                      borderRadius: 0,
                      whiteSpace: 'normal',
                      textAlign: 'left',
                      background: verificationMode === mode ? D.accentSoft : D.panel,
                      border: `1px solid ${verificationMode === mode ? D.accent : D.border}`,
                      color: verificationMode === mode ? D.accentInk : D.text,
                      boxShadow: 'none',
                    }}
                  >
                    <Typography.Text strong style={{ color: verificationMode === mode ? D.accentInk : D.text, fontSize: 14, lineHeight: 1.35 }}>
                      {label}
                    </Typography.Text>
                    <Typography.Text
                      type="secondary"
                      ellipsis
                      style={{ display: 'block', width: '100%', color: D.muted, fontSize: 12, lineHeight: 1.35, textAlign: 'left' }}
                    >
                      {desc}
                    </Typography.Text>
                  </Button>
                ))}
              </div>
            </div>
            {verificationMode === 'age_only' && (
              <div>
                <label style={{ display: 'block', fontFamily: D.mono, fontSize: 11, letterSpacing: '0.05em', textTransform: 'uppercase' as const, color: D.muted, marginBottom: 8 }}>
                  {t('demo.init.minAge')}
                </label>
                <Select
                  value={ageThreshold}
                  onChange={(e) => onAgeThresholdChange(Number(e))}
                  style={{ background: D.panel, border: `1px solid ${D.borderStrong}`, color: D.text, padding: '10px 14px', width: '100%', fontSize: 14, outline: 'none', boxSizing: 'border-box' }}
                 options={[{ value: 18, label: "18+" }, { value: 21, label: "21+" }, { value: 25, label: "25+" }]} />
              </div>
            )}
          </div>

          <div style={{ display: 'grid', gridTemplateColumns: isMobile ? '1fr' : '1fr 1fr', gap: isMobile ? 16 : 32, width: '100%' }}>
            <ContinueOnPhone
              apiKey={apiKey}
              userId={userId}
              source="demo"
              verificationMode={verificationMode}
              ageThreshold={verificationMode === 'age_only' ? ageThreshold : undefined}
              onComplete={(result) => {
                if (result.verification_id) {
                  onMobileVerificationComplete(result.verification_id);
                } else {
                  onMobileResult(result);
                  onMobileHandoffDone(true);
                }
              }}
            />
            <div style={{ background: D.panel, border: `1px solid ${D.border}`, padding: 20, display: 'flex', flexDirection: 'column', alignItems: 'center', textAlign: 'center', gap: 10 }}>
              <div style={{ fontSize: 13, color: D.muted, fontWeight: 500 }}>{t('demo.init.orDevice')}</div>
              <p style={{ fontSize: 12, color: D.dim, lineHeight: 1.5 }}>{t('demo.init.orDeviceDesc')}</p>
              <Button
                onClick={onStart}
                disabled={isLoading || !apiKey.trim() || !userId.trim()}
                style={{ background: 'transparent', color: D.muted, border: `1px solid ${D.border}`, padding: '9px 0', width: '100%', fontFamily: D.mono, fontWeight: 500, fontSize: 13, cursor: isLoading || !apiKey.trim() || !userId.trim() ? 'not-allowed' : 'pointer', opacity: isLoading || !apiKey.trim() || !userId.trim() ? 0.5 : 1 }}
              >
                {isLoading ? t('demo.init.starting') : t('demo.init.start')}
              </Button>
            </div>
          </div>
        </>
      )}
    </div>
  );
};
