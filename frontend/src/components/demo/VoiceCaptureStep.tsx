import { useTranslation } from 'react-i18next';
import { D } from './themeTokens';
import { TipBar, StepLabel, DemoPrimaryBtn, AmbientGlow } from './DemoShared';

interface VoiceCaptureStepProps {
  isProcessing: boolean;
  challengeDigits: string | null;
  expiresIn: number | null;
  isRecording: boolean;
  recordingDuration: number;
  onRequestChallenge: () => void;
  onStartRecording: () => void;
  onStopRecording: () => void;
  onSubmit: () => void;
  hasRecording: boolean;
  stepError?: string | null;
  onRetry?: () => void;
  step?: number;
  totalSteps?: number;
}

export const VoiceCaptureStep = ({
  isProcessing,
  challengeDigits,
  expiresIn,
  isRecording,
  recordingDuration,
  onRequestChallenge,
  onStartRecording,
  onStopRecording,
  onSubmit,
  hasRecording,
  stepError,
  onRetry,
  step,
  totalSteps,
}: VoiceCaptureStepProps) => {
  const { t } = useTranslation();
  const hasChallenge = !!challengeDigits;
  const expired = expiresIn !== null && expiresIn <= 0;

  return (
    <div className="demo-fade-up" style={{ padding: '8px 0', position: 'relative' }}>
      <AmbientGlow />

      <StepLabel step={step ?? 6} total={totalSteps ?? 7} label={t('demo.voice.stepLabel')} />

      <h2 style={{
        fontSize: 22, fontWeight: 700, color: D.text,
        marginBottom: 6, letterSpacing: '-0.02em',
      }}>
        {t('demo.voice.title')}
      </h2>

      <p style={{ color: D.muted, fontSize: 13, lineHeight: 1.55, marginBottom: 16 }}>
        {t('demo.voice.desc')}
      </p>

      <div style={{ maxWidth: 420, margin: '0 auto', display: 'flex', flexDirection: 'column', gap: 14 }}>
        {/* Microphone icon placeholder */}
        <div style={{
          width: 120, height: 120, borderRadius: '50%', margin: '0 auto',
          background: `linear-gradient(135deg, ${D.cyan}18, ${D.cyan}08)`,
          border: `2px solid ${isRecording ? D.cyan : D.border}`,
          display: 'flex', alignItems: 'center', justifyContent: 'center',
          transition: 'border-color 0.3s, box-shadow 0.3s',
          boxShadow: isRecording ? `0 0 24px ${D.cyan}40` : 'none',
        }}>
          <svg width="40" height="40" viewBox="0 0 24 24" fill="none" stroke={isRecording ? D.cyan : D.muted} strokeWidth="1.5">
            <path d="M12 1a3 3 0 0 0-3 3v8a3 3 0 0 0 6 0V4a3 3 0 0 0-3-3z" />
            <path d="M19 10v2a7 7 0 0 1-14 0v-2" />
            <line x1="12" y1="19" x2="12" y2="23" />
            <line x1="8" y1="23" x2="16" y2="23" />
          </svg>
        </div>

        {/* Challenge digits display */}
        {hasChallenge && !expired && (
          <div style={{
            background: `${D.cyan}0a`, border: `1px solid ${D.cyan}30`,
            borderRadius: 12, padding: '16px 20px', textAlign: 'center',
          }}>
            <div style={{
              fontFamily: D.mono, fontSize: 11, color: D.cyan,
              letterSpacing: '0.08em', marginBottom: 8, textTransform: 'uppercase',
            }}>
              {t('demo.voice.speakDigits')}
            </div>
            <div style={{
              fontFamily: D.mono, fontSize: 32, fontWeight: 700, color: D.text,
              letterSpacing: '0.2em',
            }}>
              {challengeDigits}
            </div>
            {expiresIn !== null && (
              <div style={{
                fontFamily: D.mono, fontSize: 11, color: expiresIn < 30 ? D.red : D.muted,
                marginTop: 8,
              }}>
                {t('demo.voice.expiresIn', { seconds: expiresIn })}
              </div>
            )}
          </div>
        )}

        {/* Recording duration */}
        {isRecording && (
          <div style={{
            fontFamily: D.mono, fontSize: 14, color: D.cyan,
            textAlign: 'center',
          }}>
            {t('demo.voice.recording', { seconds: recordingDuration })}
          </div>
        )}

        {/* Recording complete indicator */}
        {hasRecording && !isRecording && !isProcessing && (
          <div style={{
            fontFamily: D.mono, fontSize: 12, color: D.green || '#22c55e',
            textAlign: 'center',
          }}>
            {t('demo.voice.captured', { seconds: recordingDuration })}
          </div>
        )}

        <TipBar text={t('demo.voice.tip')} />

        {/* Action buttons */}
        {!hasChallenge && !isProcessing && (
          <DemoPrimaryBtn onClick={onRequestChallenge} disabled={isProcessing}>
            {t('demo.voice.getChallenge')}
          </DemoPrimaryBtn>
        )}

        {hasChallenge && !expired && !isRecording && !hasRecording && (
          <DemoPrimaryBtn onClick={onStartRecording} disabled={isProcessing}>
            {t('demo.voice.startRecording')}
          </DemoPrimaryBtn>
        )}

        {isRecording && (
          <DemoPrimaryBtn onClick={onStopRecording} disabled={false}>
            {t('demo.voice.stopRecording')}
          </DemoPrimaryBtn>
        )}

        {hasRecording && !isRecording && !isProcessing && (
          <DemoPrimaryBtn onClick={onSubmit} disabled={isProcessing}>
            {t('demo.voice.submit')}
          </DemoPrimaryBtn>
        )}

        {isProcessing && (
          <DemoPrimaryBtn onClick={() => {}} disabled>
            {t('demo.voice.verifying')}
          </DemoPrimaryBtn>
        )}

        {expired && !stepError && (
          <DemoPrimaryBtn onClick={() => { onRetry?.(); onRequestChallenge(); }} disabled={isProcessing}>
            {t('demo.voice.requestNew')}
          </DemoPrimaryBtn>
        )}
      </div>

      {stepError && (
        <div style={{ marginTop: 10, textAlign: 'center', display: 'flex', flexDirection: 'column', gap: 8, alignItems: 'center' }}>
          <p style={{
            fontSize: 12, color: D.red,
            fontFamily: D.mono,
          }}>
            {stepError}
          </p>
          {onRetry && (
            <DemoPrimaryBtn onClick={onRetry} disabled={isProcessing}>
              {t('demo.voice.tryAgain')}
            </DemoPrimaryBtn>
          )}
        </div>
      )}
    </div>
  );
};
