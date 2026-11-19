import { useTranslation } from 'react-i18next';
import { D } from './themeTokens';
import { OvalFaceViewfinder, TipBar, StepLabel, DemoPrimaryBtn, LivenessCues, AmbientGlow } from './DemoShared';
import { Button } from 'antd'
import type { ReactNode } from 'react';

interface LiveCaptureStepProps {
  isProcessing: boolean;
  showActiveLiveness: boolean;
  onStartLiveness: () => void;
  onSkipLiveCapture: () => void;
  renderActiveLiveness: () => ReactNode;
  stepError?: string | null;
  step?: number;
  totalSteps?: number;
}

export const LiveCaptureStep = ({
  isProcessing,
  showActiveLiveness,
  onStartLiveness,
  onSkipLiveCapture,
  renderActiveLiveness,
  stepError,
  step,
  totalSteps,
}: LiveCaptureStepProps) => {
  const { t } = useTranslation();
  return (
    <div className="demo-fade-up" style={{ padding: '8px 0', position: 'relative' }}>
      <AmbientGlow />

      <StepLabel step={step ?? 5} total={totalSteps ?? 6} label={t('demo.live.stepLabel')} />

      {!showActiveLiveness ? (
        <>
          <h2 style={{
            fontSize: 22, fontWeight: 700, color: D.text,
            marginBottom: 6, letterSpacing: '-0.02em',
          }}>
            {t('demo.live.title')}
          </h2>

          <p style={{ color: D.muted, fontSize: 13, lineHeight: 1.55, marginBottom: 12 }}>
            {t('demo.live.desc')}
          </p>

          <div style={{ maxWidth: 420, margin: '0 auto', display: 'flex', flexDirection: 'column', gap: 14 }}>
            {/* Oval face viewfinder (preview) */}
            <OvalFaceViewfinder processing={false} />

            {/* Liveness cues */}
            <LivenessCues />

            <TipBar text={t('demo.live.tip')} />

            <DemoPrimaryBtn onClick={onStartLiveness} disabled={isProcessing}>
              {t('demo.live.start')}
            </DemoPrimaryBtn>

            <Button
              onClick={onSkipLiveCapture}
              style={{
                background: 'transparent', color: D.muted,
                border: `1px solid ${D.border}`,
                padding: '11px 0', fontFamily: D.mono, fontWeight: 500, fontSize: 13,
                cursor: 'pointer', width: '100%',
              }}
            >
              {t('demo.live.skip')}
            </Button>
          </div>
        </>
      ) : (
        <>
          <h2 style={{
            fontSize: 22, fontWeight: 700, color: D.text,
            marginBottom: 6, letterSpacing: '-0.02em',
          }}>
            {t('demo.live.title')}
          </h2>

          <p style={{ color: D.muted, fontSize: 13, lineHeight: 1.55, marginBottom: 16 }}>
            {t('demo.live.descActive')}
          </p>

          <div style={{ margin: '0 auto' }}>
            {renderActiveLiveness()}
          </div>
        </>
      )}

      {stepError && (
        <p style={{
          marginTop: 10, fontSize: 12, color: D.red,
          fontFamily: D.mono, textAlign: 'center',
        }}>
          {stepError}
        </p>
      )}
    </div>
  );
};
