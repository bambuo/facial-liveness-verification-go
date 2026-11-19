import { useState } from 'react';
import { Fragment, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import { DocLayout, NavItem } from '../components/docs/DocLayout';
import {
  H2, Lead, Divider, SectionAnchor, Callout,
  FieldRow, Pre, CodeTabs, CodeTabType, Pipeline,
  EndpointCard, StatusPill,
} from '../components/docs/shared';
import { getDocumentationApiUrl } from '../config/api';
import { isCloud } from '../config/edition';
import { C } from '../theme';
import '../styles/patterns.css';

// Theme-aware token alias for docs content.
const T = {
  ...C,
  bg: 'var(--paper)',
  panel: 'var(--panel)',
  sidebar: 'var(--panel)',
  surface: 'var(--panel-2)',
  surfaceHover: 'var(--panel-3)',
  border: 'var(--rule)',
  borderStrong: 'var(--rule-strong)',
  text: 'var(--ink)',
  muted: 'var(--mid)',
  dim: 'var(--soft)',
  cyan: 'var(--accent-ink)',
  cyanDim: 'var(--accent-soft)',
  cyanBorder: 'color-mix(in oklab, var(--accent) 30%, transparent)',
  accent: 'var(--accent)',
  accentInk: 'var(--accent-ink)',
  accentSoft: 'var(--accent-soft)',
  code: 'var(--accent-ink)',
  codeBg: 'var(--paper-2)',
}

const NAV: NavItem[] = [
  { id: 'quick-start', label: 'docs.api.quickStartTitle', depth: 0 },
  { id: 'auth', label: 'docs.api.authTitle', depth: 0 },
  { id: 'flow', label: 'Verification Flow', depth: 0 },
  { id: 'step-1', label: '1 · Start Session', depth: 1 },
  { id: 'step-2', label: '2 · Upload Front', depth: 1 },
  { id: 'step-3', label: '3 · Upload Back', depth: 1 },
  { id: 'step-4', label: '4 · Live Capture', depth: 1 },
  { id: 'step-5', label: '5 · Get Results', depth: 1 },
  { id: 'selfie', label: 'Cross-Validation', depth: 1 },
  { id: 'selfie-url', label: 'Get Selfie URL', depth: 1 },
];

export const DocsPage = () => {
  const { t } = useTranslation();
  const apiUrl = getDocumentationApiUrl();
  const [tab, setTab] = useState<CodeTabType>('curl');
  const [isMobile, setIsMobile] = useState(() =>
    typeof window !== 'undefined' && window.innerWidth < 768
  );

  useEffect(() => {
    const mq = window.matchMedia('(max-width: 767px)');
    const handler = (e: MediaQueryListEvent) => setIsMobile(e.matches);
    mq.addEventListener('change', handler);
    setIsMobile(mq.matches);
    return () => mq.removeEventListener('change', handler);
  }, []);

  return (
    <DocLayout slug="api-docs" nav={NAV}>

      {/* ══ QUICK START ══════════════════════════════════════════════════ */}
      <SectionAnchor id="quick-start" />
      <H2 index="01">{t('docs.api.quickStartTitle')}</H2>
      <Lead>
        {t('docs.api.quickStartLead')}
      </Lead>

      {isCloud && (
        <div style={{ margin: '24px 0 32px', borderRadius: 0, overflow: 'hidden', border: `1px solid ${T.border}`, background: T.codeBg }}>
          <video
            src="https://qffbflsboayyqvnqwqil.supabase.co/storage/v1/object/public/platform-assets/developer-integration-demo.mp4"
            style={{ width: '100%', aspectRatio: '16/9', display: 'block' }}
            controls
            playsInline
            preload="metadata"
            poster="/identity-logo.svg"
          />
        </div>
      )}

      <Pipeline />

      <CodeTabs tab={tab} onChange={setTab}
        js={`const BASE = '${apiUrl}';
const KEY  = 'your-api-key';
const headers = { 'X-API-Key': KEY };

// ─── 1. Start session ───────────────────────────────────────────
const { verification_id } = await fetch(\`\${BASE}/api/v2/verify/initialize\`, {
  method: 'POST',
  headers: { ...headers, 'Content-Type': 'application/json' },
  body: JSON.stringify({ user_id: 'user-uuid' }),
}).then(r => r.json());

// ─── 2. Upload front document ───────────────────────────────────
const fd1 = new FormData();
fd1.append('document_type', 'drivers_license');
fd1.append('document', frontFile);       // File from <input type="file">
await fetch(\`\${BASE}/api/v2/verify/\${verification_id}/front-document\`, { method: 'POST', headers, body: fd1 });

// ─── 3. Poll until OCR finishes ─────────────────────────────────
let r;
do {
  await new Promise(ok => setTimeout(ok, 2000));
  r = await fetch(\`\${BASE}/api/v2/verify/\${verification_id}/status\`, { headers }).then(r => r.json());
} while (!r.ocr_data);

// ─── 4. Upload back-of-ID (cross-validation auto-triggers) ──────
const fd2 = new FormData();
fd2.append('document_type', 'drivers_license');
fd2.append('document', backFile);        // field name is 'document' (not 'back_of_id')
await fetch(\`\${BASE}/api/v2/verify/\${verification_id}/back-document\`, { method: 'POST', headers, body: fd2 });

// ─── 5. Poll until cross-validation finishes ────────────────────
do {
  await new Promise(ok => setTimeout(ok, 2000));
  r = await fetch(\`\${BASE}/api/v2/verify/\${verification_id}/status\`, { headers }).then(r => r.json());
} while (!r.cross_validation_results);

if (r.final_result === 'failed') throw new Error('Cross-validation failed: ' + r.failure_reason);

// ─── 6. Submit live capture (file upload) ────────────────────────
const fd3 = new FormData();
fd3.append('selfie', captureBlob, 'capture.jpg');  // Blob from canvas.toBlob()
await fetch(\`\${BASE}/api/v2/verify/\${verification_id}/live-capture\`, {
  method: 'POST',
  headers,
  body: fd3,
});

// ─── 7. Poll for final result ────────────────────────────────────
do {
  await new Promise(ok => setTimeout(ok, 2000));
  r = await fetch(\`\${BASE}/api/v2/verify/\${verification_id}/status\`, { headers }).then(r => r.json());
} while (!r.final_result);

console.log(r.final_result);               // 'verified' | 'failed' | 'manual_review'
console.log(r.face_match_results?.similarity_score);   // 0.0 – 1.0
console.log(r.ocr_data.name);              // "Jane Smith"`}
        python={`import requests, time

BASE = '${apiUrl}'
HEADERS = {'X-API-Key': 'your-api-key'}

def poll(vid, until_key=None, until_final=False):
    """Poll /status until condition is met."""
    while True:
        time.sleep(2)
        r = requests.get(f'{BASE}/api/v2/verify/{vid}/status', headers=HEADERS).json()
        if until_key and r.get(until_key): return r
        if until_final and r.get('final_result') is not None: return r

# ─── 1. Start session ───────────────────────────────────────────
session = requests.post(f'{BASE}/api/v2/verify/initialize',
    headers={**HEADERS, 'Content-Type': 'application/json'},
    json={'user_id': 'user-uuid'}
).json()
vid = session['verification_id']

# ─── 2. Upload front document ───────────────────────────────────
with open('front.jpg', 'rb') as f:
    requests.post(f'{BASE}/api/v2/verify/{vid}/front-document', headers=HEADERS,
        data={'document_type': 'drivers_license'},
        files={'document': f})

# ─── 3. Poll until OCR is ready ─────────────────────────────────
r = poll(vid, until_key='ocr_data')
print('Name:', r['ocr_data']['name'])

# ─── 4. Upload back-of-ID (cross-validation auto-triggers) ──────
with open('back.jpg', 'rb') as f:
    requests.post(f'{BASE}/api/v2/verify/{vid}/back-document', headers=HEADERS,
        data={'document_type': 'drivers_license'},
        files={'document': f})

# ─── 5. Poll until cross-validation finishes ────────────────────
r = poll(vid, until_key='cross_validation_results')
if r.get('final_result') == 'failed':
    raise Exception('Cross-validation failed: ' + r.get('failure_reason', ''))

# ─── 6. Submit live capture (file upload) ────────────────────────
with open('capture.jpg', 'rb') as f:
    requests.post(f'{BASE}/api/v2/verify/{vid}/live-capture', headers=HEADERS,
        files={'selfie': f})  # field name is 'selfie' in the API

# ─── 7. Poll for final result ────────────────────────────────────
r = poll(vid, until_final=True)
print(r['final_result'])                  # verified / failed / manual_review
print(r.get('face_match_results', {}))    # match scores
print(r['ocr_data']['name'])              # "Jane Smith"`}
      />

      <Divider />

      {/* ══ AUTHENTICATION ═══════════════════════════════════════════════ */}
      <SectionAnchor id="auth" />
      <H2 index="02">{t('docs.api.authTitle')}</H2>
      <Lead>{t('docs.api.authLead')}</Lead>

      <div style={{ background: T.surface, border: `1px solid ${T.border}`, borderRadius: 0, padding: '20px 24px', marginBottom: 16 }}>
        <div style={{ display: 'flex', gap: 24, flexWrap: 'wrap' }}>
          <div>
            <div style={{ fontFamily: T.mono, fontSize: '0.68rem', color: T.muted, letterSpacing: '0.08em', textTransform: 'uppercase', marginBottom: 6 }}>{t('docs.api.baseUrl')}</div>
            <code style={{ fontFamily: T.mono, fontSize: '0.85rem', color: T.cyan }}>{apiUrl}</code>
          </div>
          <div>
            <div style={{ fontFamily: T.mono, fontSize: '0.68rem', color: T.muted, letterSpacing: '0.08em', textTransform: 'uppercase', marginBottom: 6 }}>{t('docs.api.authHeader')}</div>
            <code style={{ fontFamily: T.mono, fontSize: '0.85rem', color: T.cyan }}>X-API-Key: ik_your_api_key</code>
          </div>
        </div>
      </div>

      <Pre label={t('docs.api.curlExample')} code={`curl -X POST ${apiUrl}/api/v2/verify/initialize \\
  -H "X-API-Key: ik_your_api_key" \\
  -H "Content-Type: application/json" \\
  -d '{"user_id": "550e8400-e29b-41d4-a716-446655440000"}'`} />

      <Callout type="tip">
        {t('docs.api.sandboxTip')}
      </Callout>

      <Divider />

      {/* ══ VERIFICATION FLOW ════════════════════════════════════════════ */}
      <SectionAnchor id="flow" />
      <H2 index="03">{t('docs.api.flowTitle')}</H2>
      <Lead>
        {t('docs.api.flowLead')}
      </Lead>

      {/* Step 1 */}
      <SectionAnchor id="step-1" />
      <EndpointCard step={1} method="POST" path="/api/v2/verify/initialize" title={t('docs.api.startSessionTitle')}>
        <p style={{ fontFamily: T.sans, fontSize: '0.88rem', color: T.muted, lineHeight: 1.7, marginBottom: 16 }}>
          {t('docs.api.step1Desc')}
        </p>
        <div style={{ marginBottom: 12 }}>
          <FieldRow name="user_id" type="UUID string" req={true} desc={t('docs.api.fieldUserIdDesc')} />
          <FieldRow name="sandbox" type="boolean" req={false} desc={t('docs.api.fieldSandboxDesc')} />
          <FieldRow name="addons" type="object" req={false} desc={t('docs.api.fieldAddonsDesc')} />
          <FieldRow name="addons.aml_screening" type="boolean" req={false} desc={t('docs.api.fieldAmlDesc')} />
        </div>
        <CodeTabs tab={tab} onChange={setTab}
          curl={`curl -X POST ${apiUrl}/api/v2/verify/initialize \\
  -H "X-API-Key: your-key" \\
  -H "Content-Type: application/json" \\
  -d '{
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "addons": { "aml_screening": true }
  }'`}
          js={`const res = await fetch(\`${apiUrl}/api/v2/verify/initialize\`, {
  method: 'POST',
  headers: {
    'X-API-Key': 'your-key',
    'Content-Type': 'application/json',
  },
  body: JSON.stringify({
    user_id: '550e8400-e29b-41d4-a716-446655440000',
    addons: { aml_screening: true },
  }),
});
const data = await res.json();`}
          python={`import requests

res = requests.post(
    '${apiUrl}/api/v2/verify/initialize',
    headers={ 'X-API-Key': 'your-key' },
    json={
        'user_id': '550e8400-e29b-41d4-a716-446655440000',
        'addons': { 'aml_screening': True },
    },
)
data = res.json()`} />
        <Pre label={t('docs.api.response201')} code={`{
  "success": true,
  "verification_id": "550e8400-e29b-41d4-a716-446655440001",
  "status": "AWAITING_FRONT",
  "current_step": 1,
  "total_steps": 5,
  "message": "Verification session initialized"
}`} />
      </EndpointCard>

      {/* Step 2 */}
      <SectionAnchor id="step-2" />
      <EndpointCard step={2} method="POST" path="/api/v2/verify/:id/front-document" title={t('docs.api.uploadFrontTitle')}>
        <p style={{ fontFamily: T.sans, fontSize: '0.88rem', color: T.muted, lineHeight: 1.7, marginBottom: 16 }}>
          {t('docs.api.step2DescBefore')}<strong style={{ color: T.text }}>{t('docs.api.step2DescStrong')}</strong>{t('docs.api.step2DescAfter')}
          <code style={{ fontFamily: T.mono, color: T.cyan }}>ocr_data</code>
          {t('docs.api.step2DescEnd')}
        </p>
        <div style={{ marginBottom: 12 }}>
          <FieldRow name="document_type" type="string" req={true} desc="'passport' | 'drivers_license' | 'national_id' | 'other'" />
          <FieldRow name="document" type="File" req={true} desc={t('docs.api.fieldDocumentDesc')} />
        </div>
        <CodeTabs tab={tab} onChange={setTab}
          curl={`curl -X POST ${apiUrl}/api/v2/verify/550e8400-e29b-41d4-a716-446655440001/front-document \\
  -H "X-API-Key: your-key" \\
  -F "document_type=drivers_license" \\
  -F "document=@front.jpg"`}
          js={`const fd = new FormData();
fd.append('document_type', 'drivers_license');
fd.append('document', frontFile); // File from <input type="file">

const res = await fetch(
  \`${apiUrl}/api/v2/verify/\${verification_id}/front-document\`,
  { method: 'POST', headers: { 'X-API-Key': 'your-key' }, body: fd },
);
const data = await res.json();`}
          python={`import requests

res = requests.post(
    f'${apiUrl}/api/v2/verify/{verification_id}/front-document',
    headers={ 'X-API-Key': 'your-key' },
    data={ 'document_type': 'drivers_license' },
    files={ 'document': open('front.jpg', 'rb') },
)
data = res.json()`} />
        <Pre label={t('docs.api.response201')} code={`{
  "success": true,
  "verification_id": "550e8400-e29b-41d4-a716-446655440001",
  "status": "AWAITING_BACK",
  "current_step": 2,
  "document_id": "doc_abc123",
  "message": "Front document processed successfully",

  // OCR extraction results (included in response):
  "ocr_data": {
    "full_name": "JANE SMITH",
    "date_of_birth": "1990-06-15",
    "id_number": "DL123456789",
    "expiration_date": "2030-06-15",
    "address": "123 Main St, Anytown, US",
    "confidence_scores": { "full_name": 0.97, "date_of_birth": 0.96 }
  }

  // If Gate 1 (quality) fails → status becomes "HARD_REJECTED":
  // "rejection_reason": "POOR_QUALITY",
  // "rejection_detail": "Image too blurry for OCR extraction"
}`} />
        <Callout type="note">
          {t('docs.api.step2CalloutBefore')}
          <code style={{ fontFamily: T.mono }}>HARD_REJECTED</code>
          {t('docs.api.step2CalloutMiddle')}
          <code style={{ fontFamily: T.mono }}>rejection_reason</code>
          {t('docs.api.step2CalloutAnd')}
          <code style={{ fontFamily: T.mono }}>rejection_detail</code>
          {t('docs.api.step2CalloutEnd')}
        </Callout>
      </EndpointCard>

      {/* Step 3 */}
      <SectionAnchor id="step-3" />
      <EndpointCard step={3} method="POST" path="/api/v2/verify/:id/back-document" title={t('docs.api.uploadBackTitle')}
        badge={{ label: t('docs.api.badgeRequired'), color: T.red, bg: T.redDim }}>
        <p style={{ fontFamily: T.sans, fontSize: '0.88rem', color: T.muted, lineHeight: 1.7, marginBottom: 16 }}>
          {t('docs.api.step3DescBefore')}<strong style={{ color: T.text }}>{t('docs.api.step3DescStrong')}</strong>{t('docs.api.step3DescAfter')}
          <strong style={{ color: T.red }}>{t('docs.api.step3DescRequired')}</strong>{t('docs.api.step3DescMiddle')}
          <StatusPill status="failed" />{t('docs.api.step3DescEnd')}
        </p>

        <div style={{ background: T.surface, border: `1px solid ${T.border}`, borderRadius: 0, padding: '14px 18px', marginBottom: 16 }}>
          <div style={{ fontFamily: T.mono, fontSize: '0.68rem', color: T.muted, letterSpacing: '0.08em', textTransform: 'uppercase', marginBottom: 10 }}>{t('docs.api.step3ChecksTitle')}</div>
          <div style={{ display: 'grid', gridTemplateColumns: isMobile ? '1fr' : '1fr 1fr', gap: '6px 24px' }}>
            {(t('docs.api.step3Checks', { returnObjects: true }) as string[]).map(s => (
              <div key={s} style={{ fontFamily: T.sans, fontSize: '0.8rem', color: T.muted, display: 'flex', gap: 8, alignItems: 'center' }}>
                <span style={{ color: T.green, fontSize: '0.7rem' }}>✓</span> {s}
              </div>
            ))}
          </div>
        </div>

        <div style={{ marginBottom: 12 }}>
          <FieldRow name="document_type" type="string" req={true} desc={t('docs.api.fieldDocTypeStep3Desc')} />
          <FieldRow name="document" type="File" req={true} desc={t('docs.api.fieldDocumentBackDesc')} />
        </div>
        <CodeTabs tab={tab} onChange={setTab}
          curl={`curl -X POST ${apiUrl}/api/v2/verify/550e8400-e29b-41d4-a716-446655440001/back-document \\
  -H "X-API-Key: your-key" \\
  -F "document_type=drivers_license" \\
  -F "document=@back.jpg"`}
          js={`const fd = new FormData();
fd.append('document_type', 'drivers_license');
fd.append('document', backFile); // File from <input type="file">

const res = await fetch(
  \`${apiUrl}/api/v2/verify/\${verification_id}/back-document\`,
  { method: 'POST', headers: { 'X-API-Key': 'your-key' }, body: fd },
);
const data = await res.json();`}
          python={`import requests

res = requests.post(
    f'${apiUrl}/api/v2/verify/{verification_id}/back-document',
    headers={ 'X-API-Key': 'your-key' },
    data={ 'document_type': 'drivers_license' },
    files={ 'document': open('back.jpg', 'rb') },
)
data = res.json()`} />
        <Pre label={t('docs.api.response201')} code={`{
  "success": true,
  "verification_id": "550e8400-e29b-41d4-a716-446655440001",
  "status": "AWAITING_LIVE",
  "current_step": 3,
  "message": "Back document processed successfully",

  // Barcode / PDF417 extraction:
  "barcode_data": {
    "first_name": "JANE",
    "last_name": "SMITH",
    "date_of_birth": "19900615",
    "document_number": "DL123456789"
  },

  // Cross-validation (front OCR vs back barcode):
  "documents_match": true,
  "cross_validation_results": {
    "verdict": "PASS",            // "PASS" or "REVIEW"
    "has_critical_failure": false,
    "score": 0.95,                // 0.0 – 1.0
    "failures": []                // e.g. ["DOB_MISMATCH", "NAME_MISMATCH"]
  }
}`} />
        <Callout type="warning">
          {t('docs.api.step3CalloutBefore')}
          <code style={{ fontFamily: T.mono }}>GET /api/v2/verify/:id/status</code>
          {t('docs.api.step3CalloutMiddle')}
          <code style={{ fontFamily: T.mono }}>cross_validation_results</code>
          {t('docs.api.step3CalloutMiddle2')}
          <code style={{ fontFamily: T.mono }}>final_result</code>
          {t('docs.api.step3CalloutMiddle3')}
          <StatusPill status="failed" />
          {t('docs.api.step3CalloutEnd')}
          <code style={{ fontFamily: T.mono }}>failure_reason</code>
          {t('docs.api.step3CalloutTail')}
        </Callout>
      </EndpointCard>

      {/* Step 4 */}
      <SectionAnchor id="step-4" />
      <EndpointCard step={4} method="POST" path="/api/v2/verify/:id/live-capture" title={t('docs.api.liveCaptureTitle')}
        badge={{ label: t('docs.api.badgeFinalGate'), color: T.green, bg: T.greenDim }}>
        <p style={{ fontFamily: T.sans, fontSize: '0.88rem', color: T.muted, lineHeight: 1.7, marginBottom: 16 }}>
          {t('docs.api.step4DescBefore')}
          <strong style={{ color: T.text }}>{t('docs.api.step4DescStrong1')}</strong>{t('docs.api.step4DescMid')}
          <code style={{ fontFamily: T.mono, color: T.cyan }}>liveness_metadata</code>
          {t('docs.api.step4DescMid2')}
          <strong style={{ color: T.text }}>{t('docs.api.step4DescStrong2')}</strong>{t('docs.api.step4DescMid3')}
          <strong style={{ color: T.text }}>{t('docs.api.step4DescStrong3')}</strong>{t('docs.api.step4DescEnd')}
        </p>

        {/* Two-column mode comparison */}
        <div style={{ display: 'grid', gridTemplateColumns: isMobile ? '1fr' : '1fr 1fr', gap: 12, marginBottom: 24 }}>
          {[
            {
              title: t('docs.api.passiveTitle'), color: T.amber, label: t('docs.api.passiveLabel'),
              desc: t('docs.api.passiveDesc'),
              signals: t('docs.api.passiveSignals', { returnObjects: true }) as string[],
              security: t('docs.api.securityBasic'),
              when: t('docs.api.passiveBestFor'),
            },
            {
              title: t('docs.api.headTurnTitle'), color: T.green, label: t('docs.api.headTurnLabel'),
              desc: t('docs.api.headTurnDesc'),
              signals: t('docs.api.headTurnSignals', { returnObjects: true }) as string[],
              security: t('docs.api.securityStrong'),
              when: t('docs.api.headTurnBestFor'),
            },
          ].map(mode => (
            <div key={mode.title} style={{ background: T.surface, border: `1px solid ${T.border}`, borderRadius: 0, padding: '18px 20px', display: 'flex', flexDirection: 'column', gap: 10, borderTop: `3px solid ${mode.color}` }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                <span style={{ fontFamily: T.mono, fontSize: '0.82rem', fontWeight: 700, color: mode.color }}>{mode.title}</span>
                <span style={{ fontFamily: T.mono, fontSize: '0.65rem', fontWeight: 600, padding: '2px 8px', borderRadius: 0, background: `${mode.color}18`, color: mode.color, border: `1px solid ${mode.color}30`, letterSpacing: '0.06em' }}>{mode.label}</span>
              </div>
              <p style={{ fontFamily: T.sans, fontSize: '0.78rem', color: T.muted, lineHeight: 1.6, margin: 0 }}>{mode.desc}</p>
              <div>
                {mode.signals.map(s => (
                  <div key={s} style={{ fontFamily: T.sans, fontSize: '0.72rem', color: T.dim, padding: '2px 0', display: 'flex', gap: 6 }}>
                    <span style={{ color: mode.color, fontSize: '0.6rem', marginTop: 3 }}>▸</span> {s}
                  </div>
                ))}
              </div>
              <div style={{ marginTop: 'auto', paddingTop: 8, borderTop: `1px solid ${T.border}` }}>
                <div style={{ fontFamily: T.mono, fontSize: '0.65rem', color: T.dim, letterSpacing: '0.06em', textTransform: 'uppercase', marginBottom: 2 }}>{t('docs.api.security')}</div>
                <span style={{ fontFamily: T.sans, fontSize: '0.78rem', color: mode.color, fontWeight: 600 }}>{mode.security}</span>
              </div>
              <div>
                <div style={{ fontFamily: T.mono, fontSize: '0.65rem', color: T.dim, letterSpacing: '0.06em', textTransform: 'uppercase', marginBottom: 2 }}>{t('docs.api.bestFor')}</div>
                <span style={{ fontFamily: T.sans, fontSize: '0.75rem', color: T.muted }}>{mode.when}</span>
              </div>
            </div>
          ))}
        </div>

        {/* Client-side requirements */}
        <div style={{ background: T.surface, border: `1px solid ${T.border}`, borderRadius: 0, padding: '14px 18px', marginBottom: 16 }}>
          <div style={{ fontFamily: T.mono, fontSize: '0.68rem', color: T.muted, letterSpacing: '0.08em', textTransform: 'uppercase', marginBottom: 10 }}>{t('docs.api.clientTitle')}</div>
          <div style={{ display: 'grid', gridTemplateColumns: isMobile ? '1fr' : '100px 1fr 1fr', gap: '8px 16px' }}>
            {[
              [t('docs.api.clientPassive'), t('docs.api.clientPassiveBrowser'), t('docs.api.clientPassiveDeps')],
              [t('docs.api.clientHeadTurn'), t('docs.api.clientHeadTurnBrowser'), t('docs.api.clientHeadTurnDeps')],
            ].map(([tier, browser, deps]) => (
              <Fragment key={tier}>
                <span style={{ fontFamily: T.mono, fontSize: '0.75rem', color: T.cyan, fontWeight: 600 }}>{tier}</span>
                <span style={{ fontFamily: T.sans, fontSize: '0.78rem', color: T.muted, lineHeight: 1.5 }}>{browser}</span>
                <span style={{ fontFamily: T.sans, fontSize: '0.78rem', color: T.muted, lineHeight: 1.5 }}>{deps}</span>
              </Fragment>
            ))}
          </div>
        </div>

        <Callout type="tip">
          <strong>{t('docs.api.step4TipStrong')}</strong>
          {t('docs.api.step4TipBodyBefore')}<code style={{ fontFamily: T.mono }}>getUserMedia</code>{t('docs.api.step4TipBodyJoin')}<code style={{ fontFamily: T.mono }}>canvas.toDataURL</code>{t('docs.api.step4TipBodyAfter')}
        </Callout>

        {/* Fields */}
        <div style={{ marginBottom: 12 }}>
          <FieldRow name="selfie" type="File" req={true} desc={t('docs.api.fieldSelfieDesc')} />
          <FieldRow name="liveness_metadata" type="JSON string" req={false} desc={t('docs.api.fieldLivenessMetadataDesc')} />
        </div>

        <Callout type="warning">
          <strong>{t('docs.api.step4WarningStrong')}</strong>
          {t('docs.api.step4WarningMid')}
          <code style={{ fontFamily: T.mono }}>LIVENESS_FAILED</code>
          {t('docs.api.step4WarningMid2')}
          <code style={{ fontFamily: T.mono }}>getUserMedia()</code>
          {t('docs.api.step4WarningEnd')}
        </Callout>

        {/* liveness_metadata JSON shape */}
        <Pre label="liveness_metadata — head_turn" code={`{
  "challenge_type": "head_turn",
  "challenge_direction": "left",
  "frames": [
    { "frame_base64": "/9j/4AAQ...", "timestamp": 1000,  "phase": "turn1_start" },
    { "frame_base64": "/9j/4BBR...", "timestamp": 4000,  "phase": "turn1_peak" },
    { "frame_base64": "/9j/4CCG...", "timestamp": 7000,  "phase": "turn1_return" },
    { "frame_base64": "/9j/4DDH...", "timestamp": 8200,  "phase": "turn_start" },
    { "frame_base64": "/9j/4EEI...", "timestamp": 11200, "phase": "turn_peak" },
    { "frame_base64": "/9j/4FFJ...", "timestamp": 14200, "phase": "turn_return" }
  ],
  "start_timestamp": 0,
  "end_timestamp": 15000
}`} />

        {/* Passive upload example */}
        <CodeTabs tab={tab} onChange={setTab}
          curl={`# Passive liveness (just the file, no metadata)
curl -X POST ${apiUrl}/api/v2/verify/VERIFICATION_ID/live-capture \\
  -H "X-API-Key: your-key" \\
  -F "selfie=@capture.jpg"`}
          js={`// Passive liveness — just the file
const fd = new FormData();
fd.append('selfie', captureFile);

const res = await fetch(
  \`${apiUrl}/api/v2/verify/\${verification_id}/live-capture\`,
  { method: 'POST', headers: { 'X-API-Key': 'your-key' }, body: fd },
);
const data = await res.json();`}
          python={`# Passive liveness — just the file
import requests

res = requests.post(
    f'${apiUrl}/api/v2/verify/{verification_id}/live-capture',
    headers={'X-API-Key': 'your-key'},
    files={'selfie': open('capture.jpg', 'rb')},
)
data = res.json()`} />

        <CodeTabs tab={tab} onChange={setTab}
          curl={`# Head-turn liveness — capture frames from camera, server analyzes face
curl -X POST ${apiUrl}/api/v2/verify/VERIFICATION_ID/live-capture \\
  -H "X-API-Key: your-key" \\
  -F "selfie=@capture.jpg" \\
  -F 'liveness_metadata={
    "challenge_type": "head_turn",
    "challenge_direction": "left",
    "frames": [
      {"frame_base64":"<base64>","timestamp":1000,"phase":"turn1_start"},
      {"frame_base64":"<base64>","timestamp":4000,"phase":"turn1_peak"},
      {"frame_base64":"<base64>","timestamp":7000,"phase":"turn1_return"},
      {"frame_base64":"<base64>","timestamp":8200,"phase":"turn_start"},
      {"frame_base64":"<base64>","timestamp":11200,"phase":"turn_peak"},
      {"frame_base64":"<base64>","timestamp":14200,"phase":"turn_return"}
    ],
    "start_timestamp": 0,
    "end_timestamp": 15000
  }'`}
          js={`// Head-turn liveness — capture frames from camera, server analyzes face
const fd = new FormData();
fd.append('selfie', captureFile);
fd.append('liveness_metadata', JSON.stringify({
  challenge_type: 'head_turn',
  challenge_direction: 'left',
  frames: capturedFrames, // from canvas.toDataURL('image/jpeg', 0.7)
  // Each frame: { frame_base64, timestamp, phase }
  // Phases: turn1_start → turn1_peak → turn1_return → turn_start → turn_peak → turn_return
  start_timestamp: startTime,
  end_timestamp: Date.now(),
}));

const res = await fetch(
  \`${apiUrl}/api/v2/verify/\${verification_id}/live-capture\`,
  { method: 'POST', headers: { 'X-API-Key': 'your-key' }, body: fd },
);
const data = await res.json();`}
          python={`# Head-turn liveness — capture frames from camera, server analyzes face
import requests, json

metadata = {
    'challenge_type': 'head_turn',
    'challenge_direction': 'left',
    'frames': captured_frames,  # list of {frame_base64, timestamp, phase}
    # Phases: turn1_start → turn1_peak → turn1_return → turn_start → turn_peak → turn_return
    'start_timestamp': start_time,
    'end_timestamp': end_time,
}

res = requests.post(
    f'${apiUrl}/api/v2/verify/{verification_id}/live-capture',
    headers={'X-API-Key': 'your-key'},
    files={'selfie': open('capture.jpg', 'rb')},
    data={'liveness_metadata': json.dumps(metadata)},
)
data = res.json()`} />

        <Pre label={t('docs.api.response201')} code={`{
  "success": true,
  "verification_id": "550e8400-e29b-41d4-a716-446655440001",
  "status": "COMPLETE",
  "current_step": 5,
  "selfie_id": "selfie_abc789",
  "message": "Verification complete",

  // Face match (live capture vs document photo):
  "face_match_results": {
    "passed": true,
    "similarity_score": 0.94,     // cosine similarity 0.0 – 1.0
    "threshold_used": 0.6         // configurable confidence threshold
  },

  // Liveness detection:
  "liveness_results": {
    "liveness_passed": true,
    "liveness_score": 0.96,
    "liveness_mode": "head_turn"  // "passive" | "head_turn"
  },

  // Age estimation (risk signal, never pass/fail):
  "age_estimation": {
    "document_face_age": 33,      // estimated from document photo
    "live_face_age": 35,          // estimated from live capture
    "declared_age": 35,           // calculated from DOB
    "age_discrepancy": 0          // |live_face_age - declared_age|
  },

  // Velocity analysis (fraud detection):
  "velocity_analysis": {
    "ip_verifications_1h": 1,     // same IP in last hour
    "ip_verifications_24h": 2,    // same IP in last 24 hours
    "user_verifications_24h": 1,  // same user in last 24 hours
    "avg_step_duration_ms": 8500, // avg time between steps
    "fastest_step_ms": 5200,      // fastest single step
    "flags": [],                  // e.g. ["rapid_ip_reuse", "bot_like_timing"]
    "score": 0                    // 0-100, highest flag score
  },

  // IP geolocation risk:
  "geo_analysis": {
    "ip_country": "US",           // ISO alpha-2 from IP lookup
    "ip_region": "CA",
    "ip_city": "San Francisco",
    "document_country": "US",     // from front document OCR
    "is_tor": false,
    "is_datacenter": false,
    "flags": [],                  // e.g. ["country_mismatch", "tor_exit_node"]
    "score": 0                    // 0-100, highest flag score
  },

  // Final auto-decision:
  "final_result": "verified"      // "verified" | "manual_review" | "failed"
}`} />

        <Callout type="tip">
          <strong>{t('docs.api.step4ChoiceStrong')}</strong>
          {t('docs.api.step4ChoiceBody')}<strong>{t('docs.api.step4ChoiceStrong2')}</strong>{t('docs.api.step4ChoiceEnd')}
        </Callout>
        <Callout type="warning">
          {t('docs.api.step4MetadataWarnBefore')}<code style={{ fontFamily: T.mono }}>liveness_metadata</code>
          {t('docs.api.step4MetadataWarnMid')}<strong>{t('docs.api.step4MetadataWarnStrong')}</strong>
          {t('docs.api.step4MetadataWarnMid2')}
          <code style={{ fontFamily: T.mono }}>VALIDATION_ERROR</code>
          {t('docs.api.step4MetadataWarnEnd')}
        </Callout>
      </EndpointCard>

      {/* Step 5 */}
      <SectionAnchor id="step-5" />
      <EndpointCard step={5} method="GET" path="/api/v2/verify/:id/status" title={t('docs.api.resultsTitle')}>
        <p style={{ fontFamily: T.sans, fontSize: '0.88rem', color: T.muted, lineHeight: 1.7, marginBottom: 16 }}>
          {t('docs.api.step5Desc')}
        </p>

        <div style={{ background: T.surface, border: `1px solid ${T.border}`, borderRadius: 0, padding: '14px 18px', marginBottom: 16 }}>
          <div style={{ fontFamily: T.mono, fontSize: '0.68rem', color: T.muted, letterSpacing: '0.08em', textTransform: 'uppercase', marginBottom: 12 }}>{t('docs.api.pollingConditions')}</div>
          {[
            { after: t('docs.api.pollingAfter2'), condition: t('docs.api.pollingCondition2'), next: t('docs.api.pollingNext2') },
            { after: t('docs.api.pollingAfter3'), condition: t('docs.api.pollingCondition3'), next: t('docs.api.pollingNext3') },
            { after: t('docs.api.pollingAfter4'), condition: t('docs.api.pollingCondition4'), next: t('docs.api.pollingNext4') },
          ].map(r => (
            <div key={r.after} style={{ display: 'flex', gap: 12, padding: '8px 0', borderTop: `1px solid ${T.border}`, flexWrap: 'wrap' }}>
              <span style={{ fontFamily: T.sans, fontSize: '0.78rem', color: T.muted, width: 160, flexShrink: 0 }}>{r.after}</span>
              <code style={{ fontFamily: T.mono, fontSize: '0.75rem', color: T.cyan, flex: 1 }}>{r.condition}</code>
              <span style={{ fontFamily: T.sans, fontSize: '0.78rem', color: T.green, flexShrink: 0 }}>→ {r.next}</span>
            </div>
          ))}
        </div>

        <CodeTabs tab={tab} onChange={setTab}
          curl={`curl -X GET ${apiUrl}/api/v2/verify/550e8400-e29b-41d4-a716-446655440001/status \\
  -H "X-API-Key: your-key"`}
          js={`const res = await fetch(
  \`${apiUrl}/api/v2/verify/\${verification_id}/status\`,
  { headers: { 'X-API-Key': 'your-key' } },
);
const data = await res.json();`}
          python={`import requests

res = requests.get(
    f'${apiUrl}/api/v2/verify/{verification_id}/status',
    headers={ 'X-API-Key': 'your-key' },
)
data = res.json()`} />
        <Pre label={t('docs.api.responseCompleted')} code={`{
  "success": true,
  "verification_id": "550e8400-e29b-41d4-a716-446655440001",
  "status": "COMPLETE",
  "current_step": 5,
  "total_steps": 5,
  "created_at": "2026-03-06T12:00:00Z",
  "updated_at": "2026-03-06T12:05:30Z",

  // ── Upload progress ──────────────────────────────────────────
  "front_document_uploaded": true,
  "back_document_uploaded": true,
  "live_capture_uploaded": true,

  // ── Front document (OCR) ─────────────────────────────────────
  "ocr_data": {
    "full_name":         "JANE SMITH",
    "date_of_birth":     "1990-06-15",
    "id_number":         "DL123456789",
    "expiration_date":   "2030-06-15",
    "address":           "123 Main St, Anytown, US",
    "confidence_scores": { "full_name": 0.97, "date_of_birth": 0.96 }
  },

  // ── Back-of-ID (barcode + cross-validation) ──────────────────
  "barcode_data": {
    "first_name": "JANE", "last_name": "SMITH",
    "document_number": "DL123456789"
  },
  "documents_match": true,
  "cross_validation_results": {
    "verdict": "PASS",
    "has_critical_failure": false,
    "overall_score": 0.95,
    "failures": []
  },

  // ── Live capture (liveness + face match) ─────────────────────
  "face_match_results": {
    "passed": true,
    "similarity_score": 0.94,
    "threshold_used": 0.6
  },
  "liveness_results": {
    "passed": true,
    "score": 0.96
  },

  // ── AML / Sanctions / PEP (when AML_PROVIDER is configured) ──
  "aml_screening": {
    "risk_level": "clear",           // "clear" | "potential_match" | "confirmed_match"
    "match_found": false,
    "match_count": 0,
    "lists_checked": ["us_ofac_sdn", "eu_sanctions", "un_sanctions", "opensanctions-pep"],
    "screened_at": "2026-03-06T12:05:28Z"
    // PEP matches include match_type: "pep", list_source: "opensanctions-pep"
  },

  // ── Age estimation (risk signal, never pass/fail) ─────────
  "age_estimation": {
    "document_face_age": 33,       // estimated from document photo
    "live_face_age": 35,           // estimated from live capture
    "declared_age": 35,            // calculated from DOB
    "age_discrepancy": 0           // |live_face_age - declared_age|
  },

  // ── Velocity analysis (fraud detection) ─────────────────────
  "velocity_analysis": {
    "ip_verifications_1h": 1,
    "ip_verifications_24h": 2,
    "user_verifications_24h": 1,
    "avg_step_duration_ms": 8500,
    "fastest_step_ms": 5200,
    "flags": [],
    "score": 0
  },

  // ── IP geolocation risk ────────────────────────────────────────
  "geo_analysis": {
    "ip_country": "US",
    "ip_region": "CA",
    "ip_city": "San Francisco",
    "document_country": "US",
    "is_tor": false,
    "is_datacenter": false,
    "flags": [],
    "score": 0
  },

  // ── Final decision ───────────────────────────────────────────
  "final_result": "verified",       // "verified" | "manual_review" | "failed"
  "failure_reason": null,           // set if final_result is "failed"
  "manual_review_reason": null      // set if final_result is "manual_review"
}`} />
      </EndpointCard>

      {/* Cross-validation (optional) */}
      <SectionAnchor id="selfie" />
      <EndpointCard method="POST" path="/api/v2/verify/:id/cross-validation" title={t('docs.api.crossValidationTitle')}
        badge={{ label: t('docs.api.badgeOptional'), color: T.dim, bg: 'rgba(74,85,104,0.13)' }}>
        <p style={{ fontFamily: T.sans, fontSize: '0.88rem', color: T.muted, lineHeight: 1.7, marginBottom: 16 }}>
          {t('docs.api.crossValidationDesc')}
        </p>
        <Pre label={t('docs.api.response200')} code={`{
  "success": true,
  "verification_id": "550e8400-e29b-41d4-a716-446655440001",
  "status": "AWAITING_LIVE",
  "documents_match": true,
  "cross_validation_results": {
    "verdict": "PASS",
    "has_critical_failure": false,
    "score": 0.95,
    "failures": []
  }
}`} />
      </EndpointCard>

      {/* Live-capture selfie URL (optional) */}
      <SectionAnchor id="selfie-url" />
      <EndpointCard method="GET" path="/api/v2/verify/:id/selfie" title={t('docs.api.selfieTitle')}
        badge={{ label: t('docs.api.badgeOptional'), color: T.dim, bg: 'rgba(74,85,104,0.13)' }}>
        <p style={{ fontFamily: T.sans, fontSize: '0.88rem', color: T.muted, lineHeight: 1.7, marginBottom: 16 }}>
          {t('docs.api.selfieDescBefore')}<strong style={{ color: T.text }}>{t('docs.api.selfieDescStrong')}</strong>{t('docs.api.selfieDescAfter')}
          <code style={{ fontFamily: T.mono, color: T.cyan }}>404</code>
          {t('docs.api.selfieDescEnd')}
        </p>
        <CodeTabs tab={tab} onChange={setTab}
          curl={`curl -X GET ${apiUrl}/api/v2/verify/550e8400-e29b-41d4-a716-446655440001/selfie \\
  -H "X-API-Key: your-key"`}
          js={`const res = await fetch(
  \`${apiUrl}/api/v2/verify/\${verification_id}/selfie\`,
  { headers: { 'X-API-Key': 'your-key' } },
);
const data = await res.json();`}
          python={`import requests

res = requests.get(
    f'${apiUrl}/api/v2/verify/{verification_id}/selfie',
    headers={ 'X-API-Key': 'your-key' },
)
data = res.json()`} />
        <Pre label={t('docs.api.response200')} code={`{
  "verification_id": "550e8400-e29b-41d4-a716-446655440001",
  "selfie_url": "https://storage.example.com/selfies/…?token=…",
  "expires_in": 300
}`} />
        <Callout type="note">
          {t('docs.api.selfieNoteBefore')}<strong>{t('docs.api.selfieNoteStrong')}</strong>{t('docs.api.selfieNoteMid')}<strong>{t('docs.api.selfieNoteStrong2')}</strong>{t('docs.api.selfieNoteMid2')}
          <code style={{ fontFamily: T.mono }}>{`{ "selfie_url": null }`}</code>{t('docs.api.selfieNoteEnd')}
        </Callout>
      </EndpointCard>

    </DocLayout>
  );
};
