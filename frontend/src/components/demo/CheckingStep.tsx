import { useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import { D } from './themeTokens';
import { StepLabel, AmbientGlow, DemoPrimaryBtn } from './DemoShared';

const MESSAGE_KEYS = ['demo.checking.msg1', 'demo.checking.msg2', 'demo.checking.msg3'];

const TAG_KEYS = ['demo.checking.tag1', 'demo.checking.tag2', 'demo.checking.tag3'];

interface CheckingStepProps {
  stepError?: string | null;
  onRetry?: () => void;
  step?: number;
  totalSteps?: number;
}

export const CheckingStep = ({ stepError, onRetry, step, totalSteps }: CheckingStepProps) => {
  const { t } = useTranslation();
  const messages = MESSAGE_KEYS.map(key => t(key));
  const tags = TAG_KEYS.map(key => t(key));
  const [msgIdx, setMsgIdx] = useState(0);

  useEffect(() => {
    const iv = setInterval(() => setMsgIdx(i => (i + 1) % messages.length), 1800);
    return () => clearInterval(iv);
  }, [messages.length]);

  return (
    <div
      className="demo-fade-up"
      style={{
        textAlign: 'center', padding: '48px 0', position: 'relative',
        display: 'flex', flexDirection: 'column', alignItems: 'center',
      }}
    >
      <AmbientGlow />

      <StepLabel step={step ?? 4} total={totalSteps ?? 6} label={t('demo.checking.stepLabel')} />

      {/* Spinner */}
      <div style={{
        width: 80, height: 80,
        border: `2.5px solid ${D.accentSoft}`,
        borderTopColor: D.accent, borderRadius: '50%',
        animation: 'dSpin 1s linear infinite',
        marginBottom: 18,
      }} />

      {/* Cycling message */}
      <p style={{
        fontFamily: D.sans, fontSize: 16, fontWeight: 700,
        color: D.text, marginBottom: 6,
      }}>
        {messages[msgIdx]}
      </p>

      <p style={{
        fontFamily: D.mono, fontSize: 11, color: D.muted,
        letterSpacing: '0.08em',
      }}>
        {t('demo.checking.moment')}
      </p>

      {/* Status tags */}
      <div style={{ display: 'flex', gap: 8, marginTop: 24, flexWrap: 'wrap', justifyContent: 'center' }}>
        {tags.map(tag => (
          <span key={tag} style={{
            padding: '5px 10px',
            background: D.accentSoft, border: `1px solid ${D.border}`,
            fontFamily: D.mono, fontSize: 10,
            color: D.accentInk, opacity: 0.7,
          }}>{tag}</span>
        ))}
      </div>

      {stepError && (
        <div style={{ marginTop: 16, textAlign: 'center' }}>
          <p style={{
            fontSize: 12, color: D.red,
            fontFamily: D.mono, marginBottom: onRetry ? 12 : 0,
          }}>
            {stepError}
          </p>
          {onRetry && (
            <DemoPrimaryBtn onClick={onRetry} style={{ maxWidth: 200, margin: '0 auto' }}>
              {t('demo.checking.retry')}
            </DemoPrimaryBtn>
          )}
        </div>
      )}
    </div>
  );
};
