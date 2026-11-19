import { useTranslation } from 'react-i18next';
import { D } from './themeTokens';
import { Button, Select, Upload } from 'antd'
import type { ChangeEvent } from 'react';

interface AddressStepProps {
  addressFile: File | null;
  addressPreview: string | null;
  addressDocType: string;
  addressResult: any;
  addressUploading: boolean;
  onAddressFileSelect: (e: ChangeEvent<HTMLInputElement>) => void;
  onAddressDocTypeChange: (type: string) => void;
  onUploadAddress: () => void;
  onStartNew: () => void;
}

export const AddressStep = ({
  addressFile,
  addressPreview,
  addressDocType,
  addressResult,
  addressUploading,
  onAddressFileSelect,
  onAddressDocTypeChange,
  onUploadAddress,
  onStartNew,
}: AddressStepProps) => {
  const { t } = useTranslation();
  const handleBeforeUpload = (file: File) => {
    onAddressFileSelect({ target: { files: [file] } } as unknown as ChangeEvent<HTMLInputElement>);
    return false;
  };
  return (
    <div style={{ padding: '8px 0' }}>
      <h2 style={{ fontSize: 18, fontWeight: 600, color: D.text, textAlign: 'center', marginBottom: 6 }}>
        {t('demo.address.title')}
        <span style={{ fontSize: 11, fontWeight: 400, color: D.dim, marginLeft: 8, verticalAlign: 'middle' }}>{t('demo.address.optional')}</span>
      </h2>
      <p style={{ color: D.muted, fontSize: 13, textAlign: 'center', marginBottom: 24 }}>
        {t('demo.address.desc')}
      </p>

      {!addressResult ? (
        <div style={{ maxWidth: 420, margin: '0 auto', display: 'flex', flexDirection: 'column', gap: 16 }}>
          <div>
            <label style={{ display: 'block', fontFamily: D.mono, fontSize: 11, letterSpacing: '0.05em', textTransform: 'uppercase' as const, color: D.muted, marginBottom: 8 }}>{t('demo.address.documentType')}</label>
            <Select
              value={addressDocType}
              onChange={(e) => onAddressDocTypeChange(e)}
              style={{ background: D.panel, border: `1px solid ${D.borderStrong}`, color: D.text, padding: '10px 14px', width: '100%', fontSize: 14, outline: 'none' }}
             options={[{ value: "utility_bill", label: t('demo.address.utilityBill') }, { value: "bank_statement", label: t('demo.address.bankStatement') }, { value: "tax_document", label: t('demo.address.taxDocument') }]} />
          </div>

          <Upload accept="image/jpeg,image/png,application/pdf" showUploadList={false} beforeUpload={handleBeforeUpload}>
            <div style={{ border: `1px solid ${D.borderStrong}`, background: D.panel, padding: '32px 16px', textAlign: 'center', cursor: 'pointer', width: '100%' }}>
              <svg style={{ width: 40, height: 40, margin: '0 auto 12px', color: D.muted }} fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
              </svg>
              <p style={{ color: D.muted, fontSize: 13 }}>{t('demo.address.uploadPrompt')}</p>
              <p style={{ color: D.dim, fontSize: 11, marginTop: 4 }}>{t('demo.address.fileHint')}</p>
            </div>
          </Upload>

          {addressFile && (
            <div style={{ background: D.surface, border: `1px solid ${D.border}`, padding: 16 }}>
              <p style={{ fontWeight: 500, color: D.text, fontSize: 13, margin: 0 }}>{addressFile.name}</p>
              <p style={{ color: D.muted, fontSize: 11, marginTop: 4 }}>{(addressFile.size / 1024 / 1024).toFixed(2)} MB</p>
              {addressPreview && (
                <img src={addressPreview} alt={t('demo.address.previewAlt')} style={{ width: '100%', height: 160, objectFit: 'contain', background: D.codeBg, marginTop: 12 }} />
              )}
            </div>
          )}

          {addressFile && (
            <Button
              onClick={onUploadAddress}
              disabled={addressUploading}
              style={{ background: D.purple, color: '#fff', border: `1px solid ${D.purple}`, padding: '11px 0', width: '100%', fontFamily: D.mono, fontWeight: 500, fontSize: 13, cursor: addressUploading ? 'not-allowed' : 'pointer', opacity: addressUploading ? 0.5 : 1 }}
            >
              {addressUploading ? t('demo.address.processing') : t('demo.address.verify')}
            </Button>
          )}

          <Button
            onClick={onStartNew}
            style={{ background: 'transparent', color: D.muted, border: `1px solid ${D.border}`, padding: '10px 0', fontFamily: D.mono, fontWeight: 500, fontSize: 13, cursor: 'pointer' }}
          >
            {t('demo.address.skipNew')}
          </Button>
        </div>
      ) : (
        <div style={{ maxWidth: 420, margin: '0 auto' }}>
          {/* Address Result Card */}
          <div style={{
            background: D.panel, border: `1px solid ${D.border}`, padding: 20, marginBottom: 20,
          }}>
            <div style={{ textAlign: 'center', marginBottom: 16 }}>
              <div style={{
                width: 48, height: 48, margin: '0 auto 12px',
                display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: 20,
                background: addressResult.status === 'verified' ? D.greenDim : addressResult.status === 'failed' ? D.redDim : D.amberDim,
                border: `1px solid ${addressResult.status === 'verified' ? D.green : addressResult.status === 'failed' ? D.red : D.amber}`,
                color: addressResult.status === 'verified' ? D.green : addressResult.status === 'failed' ? D.red : D.amber,
              }}>
                {addressResult.status === 'verified' ? '\u2713' : addressResult.status === 'failed' ? '\u2717' : '\u26A0'}
              </div>
              <h3 style={{ fontSize: 16, fontWeight: 600, color: D.text, marginBottom: 4 }}>
                {addressResult.status === 'verified' ? t('demo.address.verified') : addressResult.status === 'failed' ? t('demo.address.failed') : t('demo.address.review')}
              </h3>
            </div>

            <div style={{ display: 'flex', flexDirection: 'column', gap: 8, fontSize: 13 }}>
              <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                <span style={{ color: D.muted }}>{t('demo.address.overallScore')}</span>
                <span style={{ color: D.text, fontWeight: 600 }}>{Math.round((addressResult.score || 0) * 100)}%</span>
              </div>
              <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                <span style={{ color: D.muted }}>{t('demo.address.nameMatch')}</span>
                <span style={{ color: D.text, fontWeight: 600 }}>{Math.round((addressResult.name_match_score || 0) * 100)}%</span>
              </div>
              <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                <span style={{ color: D.muted }}>{t('demo.address.documentFresh')}</span>
                <span style={{ color: addressResult.document_fresh ? D.green : D.amber }}>{addressResult.document_fresh ? t('demo.address.yes') : t('demo.address.noUnknown')}</span>
              </div>
              {addressResult.address && (
                <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                  <span style={{ color: D.muted }}>{t('demo.address.address')}</span>
                  <span style={{ color: D.text, textAlign: 'right', maxWidth: 220 }}>{addressResult.address}</span>
                </div>
              )}
            </div>

            {addressResult.reasons && addressResult.reasons.length > 0 && (
              <div style={{ marginTop: 12, background: D.codeBg, border: `1px solid ${D.borderStrong}`, padding: 10 }}>
                <div style={{ fontSize: 11, color: D.muted, marginBottom: 6, fontWeight: 600, fontFamily: D.mono }}>{t('demo.address.validationNotes')}</div>
                {addressResult.reasons.map((r: string, i: number) => (
                  <p key={i} style={{ color: D.dim, fontSize: 11, margin: '2px 0', fontFamily: D.mono }}>{'\u2022'} {r}</p>
                ))}
              </div>
            )}
          </div>

          <Button
            onClick={onStartNew}
            style={{ background: D.accent, color: 'var(--accent-fg)', border: `1px solid ${D.accent}`, padding: '10px 24px', fontFamily: D.mono, fontWeight: 500, fontSize: 13, cursor: 'pointer', width: '100%' }}
          >
            {t('demo.address.startNew')}
          </Button>
        </div>
      )}
    </div>
  );
};
