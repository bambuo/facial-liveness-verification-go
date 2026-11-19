import { useTranslation } from 'react-i18next';
import { D } from './themeTokens';
import { IDViewfinder, TipBar, StepLabel, DemoPrimaryBtn, AmbientGlow } from './DemoShared';
import type { VerificationRequest } from './types';
import { Button, Upload } from 'antd'
import type { ChangeEvent } from 'react';

interface BackUploadStepProps {
  verificationRequest: VerificationRequest | null;
  backFile: File | null;
  backPreviewUrl: string | null;
  isLoading: boolean;
  totalSteps?: number;
  onBackFileSelect: (e: ChangeEvent<HTMLInputElement>) => void;
  onUpload: () => void;
}

export const BackUploadStep = ({
  verificationRequest,
  backFile,
  backPreviewUrl,
  isLoading,
  totalSteps,
  onBackFileSelect,
  onUpload,
}: BackUploadStepProps) => {
  const { t } = useTranslation();
  const handleBeforeUpload = (file: File) => {
    onBackFileSelect({ target: { files: [file] } } as unknown as ChangeEvent<HTMLInputElement>);
    return false;
  };
  const ocrData = verificationRequest?.ocr_data;

  return (
    <div className="demo-fade-up" style={{ padding: '8px 0', position: 'relative' }}>
      <AmbientGlow />

      <StepLabel step={3} total={totalSteps ?? 6} label={t('demo.back.stepLabel')} />

      <h2 style={{
        fontSize: 22, fontWeight: 700, color: D.text,
        marginBottom: 6, letterSpacing: '-0.02em',
      }}>
        {t('demo.back.title')}
      </h2>

      <p style={{ color: D.muted, fontSize: 13, lineHeight: 1.55, marginBottom: 16 }}>
        {t('demo.back.desc')}
      </p>

      <div style={{ maxWidth: 420, margin: '0 auto', display: 'flex', flexDirection: 'column', gap: 14 }}>
        {/* OCR results from front (collapsed summary) */}
        {ocrData && Object.keys(ocrData).length > 0 && (
          <div style={{
            background: D.surface, border: `1px solid ${D.border}`,
            padding: '12px 14px',
          }}>
            <div style={{
              fontFamily: D.mono, fontSize: 10, fontWeight: 600,
              color: D.muted, marginBottom: 8, letterSpacing: '0.06em',
              textTransform: 'uppercase',
            }}>
              {t('demo.back.frontResults')}
            </div>
            <div style={{ display: 'flex', flexWrap: 'wrap', gap: '6px 16px', fontSize: 12 }}>
              {ocrData.full_name && (
                <div>
                  <span style={{ color: D.dim }}>{t('demo.back.name')}: </span>
                  <span style={{ color: D.text, fontWeight: 500 }}>{ocrData.full_name}</span>
                </div>
              )}
              {ocrData.date_of_birth && (
                <div>
                  <span style={{ color: D.dim }}>{t('demo.back.dob')}: </span>
                  <span style={{ color: D.text, fontWeight: 500 }}>{ocrData.date_of_birth}</span>
                </div>
              )}
              {ocrData.document_number && (
                <div>
                  <span style={{ color: D.dim }}>{t('demo.back.docNumber')}: </span>
                  <span style={{ color: D.text, fontWeight: 500, fontFamily: D.mono }}>{ocrData.document_number}</span>
                </div>
              )}
            </div>
          </div>
        )}

        <TipBar text={t('demo.back.tip')} />

        {/* ID Viewfinder (back variant) */}
        <IDViewfinder
          variant="back"
          processing={isLoading}
          processingLabel={t('demo.back.processingLabel')}
          previewUrl={backPreviewUrl}
        />

        {!backFile ? (
          <Upload
            accept="image/jpeg,image/png,image/webp"
            showUploadList={false}
            beforeUpload={handleBeforeUpload}
            disabled={isLoading}
          >
            <DemoPrimaryBtn onClick={() => undefined} disabled={isLoading}>
              {t('demo.back.chooseBack')}
            </DemoPrimaryBtn>
          </Upload>
        ) : (
          <>
            <div style={{
              display: 'flex', alignItems: 'center', justifyContent: 'space-between',
              background: D.surface, border: `1px solid ${D.border}`,
              padding: '10px 14px',
            }}>
              <div>
                <p style={{ fontWeight: 500, color: D.text, fontSize: 13, margin: 0 }}>{backFile.name}</p>
                <p style={{ color: D.dim, fontSize: 11, margin: '2px 0 0' }}>{(backFile.size / 1024 / 1024).toFixed(2)} MB</p>
              </div>
              <Upload accept="image/jpeg,image/png,image/webp" showUploadList={false} beforeUpload={handleBeforeUpload}>
                <Button type="link" style={{ color: D.accentInk, fontSize: 12, fontFamily: D.mono }}>
                  {t('demo.back.change')}
                </Button>
              </Upload>
            </div>
            <DemoPrimaryBtn onClick={onUpload} disabled={isLoading}>
              {isLoading ? t('demo.back.processing') : t('demo.back.scanBack')}
            </DemoPrimaryBtn>
          </>
        )}
      </div>
    </div>
  );
};
