import { useTranslation } from 'react-i18next';
import { D } from './themeTokens';

interface ProgressIndicatorProps {
  currentStep: number;
  isMobile: boolean;
  stepLabels?: string[];
}

export const ProgressIndicator = ({ currentStep, isMobile, stepLabels }: ProgressIndicatorProps) => {
  const { t } = useTranslation();
  const steps = stepLabels || [t('demo.steps.start'), t('demo.steps.front'), t('demo.steps.back'), t('demo.steps.checking'), t('demo.steps.live'), t('demo.steps.results')];

  return (
    <div style={{
      display: 'grid',
      gridTemplateColumns: `repeat(${steps.length}, 1fr)`,
      gap: 8,
      fontFamily: D.mono,
      fontSize: 11,
      marginBottom: isMobile ? 24 : 32,
    }}>
      {steps.map((label, i) => {
        const stepNum = i + 1;
        const isDone = stepNum < currentStep;
        const isActive = stepNum === currentStep;
        return (
          <div key={i} style={{
            display: 'grid',
            gridTemplateColumns: 'auto 1fr',
            gap: 8,
            alignItems: 'center',
            padding: '6px 0',
            borderTop: `2px solid ${isDone ? D.accent : isActive ? D.text : D.border}`,
            color: isDone ? D.accentInk : isActive ? D.text : D.muted,
            transition: 'border-color 0.3s, color 0.3s',
          }}>
            <span style={{ fontVariantNumeric: 'tabular-nums' }}>{stepNum}</span>
            <span style={{
              fontSize: isMobile ? 8 : 10,
              textTransform: 'uppercase',
              letterSpacing: '0.06em',
              whiteSpace: 'nowrap',
              overflow: 'hidden',
              textOverflow: 'ellipsis',
            }}>{label}</span>
          </div>
        );
      })}
    </div>
  );
};
