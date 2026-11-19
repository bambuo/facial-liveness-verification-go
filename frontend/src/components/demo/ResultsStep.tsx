import { useTranslation } from 'react-i18next';
import { D } from './themeTokens';
import { formatDateTime } from '../../i18n/formatters';
import type { VerificationRequest } from './types';
import { Button } from 'antd'
import type { CSSProperties, ReactNode } from 'react';

// ─── JSON syntax highlighting ───────────────────────────────────
const jsonTokenColors = {
  key: D.accent,
  string: D.green,
  number: D.amber,
  boolean: D.purple,
  null: D.red,
  brace: D.dim,
  comma: 'rgba(255,255,255,0.25)',
} as const;

function highlightJson(obj: unknown): ReactNode[] {
  const raw = JSON.stringify(obj, null, 2);
  if (!raw) return [];
  // Tokenize JSON string into colored spans
  const nodes: ReactNode[] = [];
  // Regex matches: strings (including keys), numbers, booleans, null, braces/brackets, commas/colons
  const tokenRe = /("(?:[^"\\]|\\.)*")\s*:|("(?:[^"\\]|\\.)*")|(-?\d+(?:\.\d+)?(?:[eE][+-]?\d+)?)\b|(true|false)|(null)|([{}\[\]])|([,:])|\n( *)/g;
  let lastIndex = 0;
  let match: RegExpExecArray | null;
  let i = 0;
  while ((match = tokenRe.exec(raw)) !== null) {
    // Any unmatched text between tokens
    if (match.index > lastIndex) {
      nodes.push(raw.slice(lastIndex, match.index));
    }
    lastIndex = match.index + match[0].length;
    if (match[1]) {
      // Key (quoted string followed by colon)
      nodes.push(<span key={i++} style={{ color: jsonTokenColors.key }}>{match[1]}</span>);
      nodes.push(<span key={i++} style={{ color: jsonTokenColors.comma }}>: </span>);
    } else if (match[2]) {
      // String value
      nodes.push(<span key={i++} style={{ color: jsonTokenColors.string }}>{match[2]}</span>);
    } else if (match[3]) {
      // Number
      nodes.push(<span key={i++} style={{ color: jsonTokenColors.number }}>{match[3]}</span>);
    } else if (match[4]) {
      // Boolean
      nodes.push(<span key={i++} style={{ color: jsonTokenColors.boolean }}>{match[4]}</span>);
    } else if (match[5]) {
      // null
      nodes.push(<span key={i++} style={{ color: jsonTokenColors.null }}>{match[5]}</span>);
    } else if (match[6]) {
      // Braces / brackets
      nodes.push(<span key={i++} style={{ color: jsonTokenColors.brace }}>{match[6]}</span>);
    } else if (match[7]) {
      // Comma / colon
      nodes.push(<span key={i++} style={{ color: jsonTokenColors.comma }}>{match[7]}</span>);
    } else if (match[8] !== undefined) {
      // Newline + indentation
      nodes.push('\n' + match[8]);
    }
  }
  if (lastIndex < raw.length) nodes.push(raw.slice(lastIndex));
  return nodes;
}

interface ResultsStepProps {
  verificationRequest: VerificationRequest;
  isMobile: boolean;
  retryProcessing: boolean;
  onRetry: () => void;
  onStartNew: () => void;
  onGoToAddress: () => void;
  onGoToCredential?: () => void;
}

// Helper: score bar for 0-1 or 0-100 values
const ScoreBar = ({ value, max = 1, color, label, detail }: { value: number | null | undefined; max?: number; color: string; label: string; detail?: string }) => {
  if (value == null) return null;
  const pct = max === 1 ? Math.round(value * 100) : Math.round(value);
  return (
    <div style={{ marginBottom: 10 }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: 12, marginBottom: 4 }}>
        <span style={{ color: D.muted }}>{label}</span>
        <span style={{ color, fontWeight: 600, fontFamily: D.mono }}>{pct}%</span>
      </div>
      <div style={{ height: 3, background: D.border, overflow: 'hidden' }}>
        <div style={{ height: '100%', width: `${pct}%`, background: color, transition: 'width 0.6s ease' }} />
      </div>
      {detail && <div style={{ fontSize: 10, color: D.dim, marginTop: 2 }}>{detail}</div>}
    </div>
  );
};

// Helper: pass/fail badge
const Badge = ({ passed, label }: { passed: boolean | null | undefined; label?: string }) => {
  const { t } = useTranslation();
  if (passed == null) return <span style={{ color: D.dim, fontSize: 11 }}>N/A</span>;
  return (
    <span style={{
      display: 'inline-flex', alignItems: 'center', gap: 4,
      fontSize: 11, fontWeight: 600, padding: '2px 8px',
      fontFamily: D.mono, letterSpacing: '0.03em',
      background: passed ? D.greenDim : D.redDim,
      color: passed ? D.green : D.red,
      border: `1px solid ${passed ? 'rgba(52,211,153,0.25)' : 'rgba(248,113,113,0.25)'}`,
    }}>
      {passed ? '\u2713' : '\u2717'} {label ?? (passed ? t('demo.results.passed') : t('demo.results.failedShort'))}
    </span>
  );
};

export const ResultsStep = ({
  verificationRequest,
  isMobile,
  retryProcessing,
  onRetry,
  onStartNew,
  onGoToAddress,
  onGoToCredential,
}: ResultsStepProps) => {
  const { t } = useTranslation();
  // v2: final_result has user-facing status, status has internal machine state
  const status = verificationRequest?.final_result ?? verificationRequest?.status;
  const isVerified = status === 'verified';
  const isFailed = status === 'failed';
  const ageVerif = (verificationRequest as any)?.age_verification;
  const isAgeOnly = !!(ageVerif && !verificationRequest?.cross_validation_results);
  const statusTone = isVerified ? D.green : isFailed ? D.red : D.amber;
  const statusBg = isVerified ? D.greenDim : isFailed ? D.redDim : D.amberDim;
  const statusIcon = isVerified ? '\u2713' : isFailed ? '\u2717' : '\u26A0';
  const statusLabel = isAgeOnly
    ? (isVerified ? t('demo.results.ageVerified') : t('demo.results.ageVerificationFailed'))
    : (isVerified ? t('demo.results.complete') : isFailed ? t('demo.results.failed') : t('demo.results.underReview'));
  const statusText = isVerified
    ? t('demo.results.verifiedStatus')
    : isFailed
      ? t('demo.results.failedStatus')
      : t('demo.results.reviewStatus');

  const cv = verificationRequest?.cross_validation_results;
  const fm = verificationRequest?.face_match_results;
  const lv = verificationRequest?.liveness_results;
  const aml = verificationRequest?.aml_screening;
  const risk = verificationRequest?.risk_score;
  const ageEst = verificationRequest?.age_estimation;

  // Card style
  const cardStyle: CSSProperties = {
    background: D.panel, border: `1px solid ${D.border}`, padding: 16, marginBottom: 12, textAlign: 'left',
  };
  const cardTitle: CSSProperties = {
    fontFamily: D.mono, fontSize: 11, fontWeight: 600, color: D.muted, marginBottom: 10, letterSpacing: '0.04em', textTransform: 'uppercase' as const,
  };

  return (
    <div style={{ padding: '8px 0' }}>
      {/* Status Header */}
      <div style={{ textAlign: 'center', marginBottom: 24 }}>
        <div style={{ width: 56, height: 56, background: statusBg, border: `1px solid ${statusTone}`, display: 'flex', alignItems: 'center', justifyContent: 'center', margin: '0 auto 12px', fontSize: 22, color: statusTone }}>
          {statusIcon}
        </div>
        <h2 style={{ fontSize: 20, fontWeight: 600, color: D.text, marginBottom: 4 }}>{statusLabel}</h2>
        <p style={{ color: D.muted, fontSize: 13, margin: 0 }}>
          {isAgeOnly
            ? isVerified
              ? t('verification.ageMet', { threshold: ageVerif?.age_threshold ?? 18 })
              : (verificationRequest?.user_message || (verificationRequest as any)?.message || t('verification.ageFailed'))
            : <>
                {isVerified && t('verification.allGatesPassed')}
                {isFailed && (verificationRequest?.user_message || t('verification.failedDefault'))}
                {!isVerified && !isFailed && (verificationRequest?.user_message || t('verification.manualReviewDefault'))}
              </>}
        </p>
        {verificationRequest?.rejection_reason && (
          <div style={{ display: 'inline-flex', alignItems: 'center', gap: 6, marginTop: 8, padding: '4px 12px', background: D.redDim, border: `1px solid rgba(248,113,113,0.2)` }}>
            <span title={verificationRequest.rejection_reason} style={{ color: D.red, fontSize: 11, fontFamily: D.mono, fontWeight: 600 }}>
              {t(`demo.results.rejectionShort.${verificationRequest.rejection_reason}`, {
                defaultValue: verificationRequest.rejection_reason.replace(/_/g, ' ').toLowerCase(),
              })}
            </span>
          </div>
        )}
      </div>

      {/* Verification Overview Card */}
      <div style={cardStyle}>
        <div style={cardTitle}>{t('demo.results.overview')}</div>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 6, fontSize: 13 }}>
          <div style={{ display: 'flex', justifyContent: 'space-between' }}>
            <span style={{ color: D.muted }}>{t('demo.results.verificationId')}</span>
            <span style={{ color: D.text, fontFamily: D.mono, fontSize: 11 }}>{verificationRequest?.verification_id?.slice(0, 12)}{'\u2026'}</span>
          </div>
          <div style={{ display: 'flex', justifyContent: 'space-between' }}>
            <span style={{ color: D.muted }}>{t('demo.results.status')}</span>
            <span style={{ color: statusTone, fontWeight: 600 }}>{statusText}</span>
          </div>
          <div style={{ display: 'flex', justifyContent: 'space-between' }}>
            <span style={{ color: D.muted }}>{t('demo.results.pipelineStep')}</span>
            <span style={{ color: D.text, fontFamily: D.mono, fontSize: 11 }}>{verificationRequest?.current_step ?? 'N/A'}</span>
          </div>
          <div style={{ display: 'flex', justifyContent: 'space-between' }}>
            <span style={{ color: D.muted }}>{t('demo.results.created')}</span>
            <span style={{ color: D.text }}>{verificationRequest?.created_at ? formatDateTime(verificationRequest.created_at) : 'N/A'}</span>
          </div>
        </div>
      </div>

      {/* Age Verification Card */}
      {ageVerif && (
        <div style={cardStyle}>
          <div style={cardTitle}>{t('demo.results.ageVerification')}</div>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 6, fontSize: 13 }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
              <span style={{ color: D.muted }}>{t('demo.results.ageCheck')}</span>
              <Badge passed={ageVerif.is_of_age} label={ageVerif.is_of_age ? t('demo.results.passed') : t('demo.results.failedShort')} />
            </div>
            <div style={{ display: 'flex', justifyContent: 'space-between' }}>
              <span style={{ color: D.muted }}>{t('demo.results.minimumAge')}</span>
              <span style={{ color: D.text, fontWeight: 600 }}>{ageVerif.age_threshold}+</span>
            </div>
          </div>
        </div>
      )}

      {/* Result Cards Grid */}
      {!isAgeOnly && (
      <div style={{ display: 'grid', gridTemplateColumns: isMobile ? '1fr' : '1fr 1fr', gap: 12, marginBottom: 12 }}>
        {/* Cross-Validation */}
        <div style={cardStyle}>
          <div style={cardTitle}>{t('demo.results.crossValidation')}</div>
          {cv ? (
            <>
              <ScoreBar value={cv.overall_score} color={cv.verdict === 'PASS' ? D.green : cv.verdict === 'REJECT' ? D.red : D.amber} label={t('demo.results.overallScore')} />
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', fontSize: 12, marginBottom: 6 }}>
                <span style={{ color: D.muted }}>{t('demo.results.verdict')}</span>
                <Badge passed={cv.verdict === 'PASS'} label={cv.verdict} />
              </div>
              <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: 12 }}>
                <span style={{ color: D.muted }}>{t('demo.results.docExpired')}</span>
                <span style={{ color: cv.document_expired ? D.red : D.green, fontSize: 11, fontWeight: 600 }}>{cv.document_expired ? t('demo.results.yes') : t('demo.results.no')}</span>
              </div>
              {cv.field_scores && Object.keys(cv.field_scores).length > 0 && (
                <div style={{ marginTop: 8, borderTop: `1px solid ${D.border}`, paddingTop: 8 }}>
                  <div style={{ fontSize: 10, color: D.dim, marginBottom: 4, fontFamily: D.mono }}>{t('demo.results.fieldScores')}</div>
                  {Object.entries(cv.field_scores).map(([field, data]) => (
                    <div key={field} style={{ display: 'flex', justifyContent: 'space-between', fontSize: 11, marginBottom: 2 }}>
                      <span style={{ color: D.dim }}>{field}</span>
                      <span style={{ color: data.passed ? D.green : D.red, fontFamily: D.mono }}>{Math.round(data.score * 100)}%</span>
                    </div>
                  ))}
                </div>
              )}
            </>
          ) : (
            <span style={{ color: D.dim, fontSize: 12 }}>{t('demo.results.notCompleted')}</span>
          )}
        </div>

        {/* Face Match */}
        <div style={cardStyle}>
          <div style={cardTitle}>{t('demo.results.faceMatch')}</div>
          {fm ? (
            <>
              <ScoreBar value={fm.similarity_score} color={fm.passed ? D.green : D.red} label={t('demo.results.similarity')} />
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', fontSize: 12, marginBottom: 6 }}>
                <span style={{ color: D.muted }}>{t('demo.results.result')}</span>
                <Badge passed={fm.passed} />
              </div>
              <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: 12 }}>
                <span style={{ color: D.muted }}>{t('demo.results.threshold')}</span>
                <span style={{ color: D.text, fontFamily: D.mono, fontSize: 11 }}>{fm.threshold_used != null ? `${Math.round(fm.threshold_used * 100)}%` : 'N/A'}</span>
              </div>
            </>
          ) : (
            <span style={{ color: D.dim, fontSize: 12 }}>{t('demo.results.notCompleted')}</span>
          )}
        </div>

        {/* Age Estimation */}
        {ageEst && (
        <div style={cardStyle}>
          <div style={cardTitle}>{t('demo.results.ageEstimation')}</div>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 6, fontSize: 13 }}>
            {ageEst.live_face_age != null && (
              <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                <span style={{ color: D.muted }}>{t('demo.results.estimatedAge')}</span>
                <span style={{ color: D.text, fontWeight: 600, fontFamily: D.mono }}>{ageEst.live_face_age}</span>
              </div>
            )}
            {ageEst.declared_age != null && (
              <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                <span style={{ color: D.muted }}>{t('demo.results.declaredAge')}</span>
                <span style={{ color: D.text, fontWeight: 600, fontFamily: D.mono }}>{ageEst.declared_age}</span>
              </div>
            )}
            {ageEst.document_face_age != null && (
              <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                <span style={{ color: D.muted }}>{t('demo.results.documentPhotoAge')}</span>
                <span style={{ color: D.dim, fontFamily: D.mono, fontSize: 11 }}>{ageEst.document_face_age}</span>
              </div>
            )}
            {ageEst.age_discrepancy != null && (
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginTop: 4, paddingTop: 6, borderTop: `1px solid ${D.border}` }}>
                <span style={{ color: D.muted }}>{t('demo.results.discrepancy')}</span>
                <Badge
                  passed={ageEst.age_discrepancy < 5}
                  label={t(ageEst.age_discrepancy === 1 ? 'demo.results.year' : 'demo.results.years', { count: ageEst.age_discrepancy })}
                />
              </div>
            )}
          </div>
        </div>
        )}

        {/* Liveness */}
        <div style={cardStyle}>
          <div style={cardTitle}>{t('demo.results.livenessDetection')}</div>
          {lv ? (
            <>
              <ScoreBar value={lv.score} color={lv.passed ? D.green : D.red} label={t('demo.results.livenessScore')} />
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', fontSize: 12 }}>
                <span style={{ color: D.muted }}>{t('demo.results.result')}</span>
                <Badge passed={lv.passed} />
              </div>
            </>
          ) : (
            <span style={{ color: D.dim, fontSize: 12 }}>{t('demo.results.notCompleted')}</span>
          )}
        </div>

        {/* AML Screening */}
        <div style={cardStyle}>
          <div style={cardTitle}>{t('demo.results.amlSanctions')}</div>
          {aml ? (
            <>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', fontSize: 12, marginBottom: 6 }}>
                <span style={{ color: D.muted }}>{t('demo.results.riskLevel')}</span>
                <Badge passed={aml.risk_level === 'clear'} label={aml.risk_level?.replace('_', ' ')} />
              </div>
              <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: 12, marginBottom: 4 }}>
                <span style={{ color: D.muted }}>{t('demo.results.matches')}</span>
                <span style={{ color: aml.match_found ? D.red : D.green, fontWeight: 600, fontSize: 11 }}>{t('demo.results.matchesFound', { count: aml.match_count ?? 0 })}</span>
              </div>
              {aml.lists_checked && aml.lists_checked.length > 0 && (
                <div style={{ fontSize: 10, color: D.dim, marginTop: 4 }}>
                  {t('demo.results.lists')}: {aml.lists_checked.join(', ')}
                </div>
              )}
            </>
          ) : (
            <span style={{ color: D.dim, fontSize: 12 }}>{t('demo.results.notAvailable')}</span>
          )}
        </div>
      </div>
      )}

      {/* Risk Score — full width */}
      {risk && (
        <div style={cardStyle}>
          <div style={cardTitle}>{t('demo.results.riskScore')}</div>
          <div style={{ display: 'flex', alignItems: 'center', gap: 16 }}>
            <div style={{
              width: 52, height: 52, display: 'flex', alignItems: 'center', justifyContent: 'center', flexShrink: 0,
              fontSize: 18, fontWeight: 700, fontFamily: D.mono,
              background: risk.risk_level === 'low' ? D.greenDim : risk.risk_level === 'medium' ? D.amberDim : D.redDim,
              border: `2px solid ${risk.risk_level === 'low' ? D.green : risk.risk_level === 'medium' ? D.amber : D.red}`,
              color: risk.risk_level === 'low' ? D.green : risk.risk_level === 'medium' ? D.amber : D.red,
            }}>
              {risk.overall_score}
            </div>
            <div style={{ flex: 1 }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 4 }}>
                <span style={{ color: D.text, fontSize: 13, fontWeight: 600, textTransform: 'capitalize' }}>{t('demo.results.riskLevelValue', { level: risk.risk_level })}</span>
                <span style={{ color: D.dim, fontSize: 11, fontFamily: D.mono }}>{t('demo.results.scale')}</span>
              </div>
              <div style={{ height: 3, background: D.border, overflow: 'hidden' }}>
                <div style={{ height: '100%', width: `${risk.overall_score}%`, background: risk.risk_level === 'low' ? D.green : risk.risk_level === 'medium' ? D.amber : D.red, transition: 'width 0.6s ease' }} />
              </div>
              {risk.risk_factors && risk.risk_factors.length > 0 && (
                <div style={{ marginTop: 6, display: 'flex', gap: 4, flexWrap: 'wrap' }}>
                  {risk.risk_factors.map((f, i) => (
                    <span key={i} style={{ fontSize: 10, padding: '1px 6px', background: D.surface, border: `1px solid ${D.border}`, color: D.dim, fontFamily: D.mono }}>
                      {f.factor}
                    </span>
                  ))}
                </div>
              )}
            </div>
          </div>
        </div>
      )}

      {/* Raw API Response — always shown */}
      <div style={{ ...cardStyle, marginBottom: 20 }}>
        <div style={cardTitle}>{t('demo.results.rawResponse')}</div>
        <p style={{ color: D.dim, fontSize: 11, margin: '0 0 8px' }}>
          {t('demo.results.rawDesc')} <code style={{ color: D.accent, background: D.codeBg, padding: '1px 4px', fontSize: 10 }}>GET /api/v2/verify/:id/status</code>
        </p>
        <pre style={{ background: D.codeBg, color: D.code, padding: 12, border: `1px solid ${D.borderStrong}`, fontSize: 10, fontFamily: D.mono, overflowX: 'auto', maxHeight: 300, overflowY: 'auto', lineHeight: 1.5, margin: 0 }}>
          {highlightJson(verificationRequest)}
        </pre>
      </div>

      {/* Retry Button (failed only) */}
      {isFailed && verificationRequest?.retry_available === true && (
        <div style={{ textAlign: 'center', marginBottom: 16 }}>
          <Button
            onClick={onRetry}
            disabled={retryProcessing}
            style={{
              background: retryProcessing ? 'transparent' : D.accent,
              color: retryProcessing ? D.muted : D.bg, border: `1px solid ${D.accent}`,
              padding: '10px 32px', fontFamily: D.mono, fontWeight: 500, fontSize: 13,
              cursor: retryProcessing ? 'not-allowed' : 'pointer',
              opacity: retryProcessing ? 0.6 : 1,
            }}
          >
            {retryProcessing ? t('demo.results.restarting') : t('demo.results.tryAgain')}
          </Button>
        </div>
      )}
      {isFailed && verificationRequest?.retry_available === false && (
        <p style={{ color: D.red, fontSize: 11, textAlign: 'center', marginBottom: 16 }}>
          {t('demo.results.maxRetry')}
        </p>
      )}

      {/* Action Buttons */}
      <div style={{ display: 'flex', gap: 12, justifyContent: 'center', flexWrap: 'wrap' }}>
        {isVerified && onGoToCredential && (
          <Button
            onClick={onGoToCredential}
            style={{
              background: D.accent, color: 'var(--accent-fg)', border: `1px solid ${D.accent}`,
              padding: '10px 24px', fontFamily: D.mono, fontWeight: 500, fontSize: 13, cursor: 'pointer',
              display: 'flex', alignItems: 'center', gap: 8,
            }}
          >
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" />
              <path d="M9 12l2 2 4-4" />
            </svg>
            {t('demo.results.getCredential')}
          </Button>
        )}
        <Button
          onClick={onGoToAddress}
          style={{ background: D.purple, color: '#fff', border: `1px solid ${D.purple}`, padding: '10px 24px', fontFamily: D.mono, fontWeight: 500, fontSize: 13, cursor: 'pointer', display: 'flex', alignItems: 'center', gap: 8 }}
        >
          {t('demo.results.tryAddress')}
        </Button>
        <Button
          onClick={onStartNew}
          style={{ background: 'transparent', color: D.muted, border: `1px solid ${D.border}`, padding: '10px 24px', fontFamily: D.mono, fontWeight: 500, fontSize: 13, cursor: 'pointer' }}
        >
          {t('demo.results.startNew')}
        </Button>
      </div>
      <p style={{ color: D.dim, fontSize: 11, marginTop: 12, textAlign: 'center' }}>
        {t('demo.results.postFeatures')}
      </p>
    </div>
  );
};
