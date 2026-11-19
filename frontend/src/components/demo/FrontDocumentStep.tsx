import { useTranslation } from 'react-i18next';
import { D } from './themeTokens';
import { IDViewfinder, TipBar, StepLabel, DemoPrimaryBtn, AmbientGlow } from './DemoShared';
import { Button, Select, Upload } from 'antd'
import type { ChangeEvent } from 'react';

interface FrontDocumentStepProps {
  selectedFile: File | null;
  previewUrl: string | null;
  documentType: string;
  isLoading: boolean;
  isAgeOnly?: boolean;
  ageThreshold?: number;
  totalSteps?: number;
  onFileSelect: (e: ChangeEvent<HTMLInputElement>) => void;
  onDocumentTypeChange: (type: string) => void;
  onUpload: () => void;
}

export const FrontDocumentStep = ({
  selectedFile,
  previewUrl,
  documentType,
  isLoading,
  isAgeOnly,
  ageThreshold,
  totalSteps,
  onFileSelect,
  onDocumentTypeChange,
  onUpload,
}: FrontDocumentStepProps) => {
  const { t } = useTranslation();
  const handleBeforeUpload = (file: File) => {
    onFileSelect({ target: { files: [file] } } as unknown as ChangeEvent<HTMLInputElement>);
    return false;
  };
  return (
    <div className="demo-fade-up" style={{ padding: '8px 0', position: 'relative' }}>
      <AmbientGlow />

      <StepLabel
        step={2}
        total={totalSteps ?? 6}
        label={isAgeOnly ? t('demo.front.uploadId') : t('demo.front.stepLabel')}
      />

      <h2 style={{
        fontSize: 22, fontWeight: 700, color: D.text,
        marginBottom: 6, letterSpacing: '-0.02em',
      }}>
        {isAgeOnly ? t('demo.front.titleAge') : t('demo.front.title')}
      </h2>

      <p style={{ color: D.muted, fontSize: 13, lineHeight: 1.55, marginBottom: 16 }}>
        {isAgeOnly
          ? t('demo.front.ageDesc', { age: ageThreshold ?? 18 })
          : t('demo.front.desc')}
      </p>

      <div style={{ maxWidth: 420, margin: '0 auto', display: 'flex', flexDirection: 'column', gap: 14 }}>
        {/* Document type selector */}
        <Select
          value={documentType}
          onChange={(e) => onDocumentTypeChange(e)}
          style={{
            width: '100%', padding: '10px 14px',
            border: `1px solid ${D.borderStrong}`, background: D.panel,
            color: D.text, fontFamily: D.mono, fontSize: 12, outline: 'none',
          }}
         options={[{ value: "", label: t('demo.front.selectType') }, { value: "national_id", label: t('demo.front.nationalId') }, { value: "drivers_license", label: t('demo.front.driversLicense') }, { value: "passport", label: t('demo.front.passport') }, { value: "other", label: t('demo.front.other') }]} />

        <TipBar text={t('demo.front.tip')} />

        {/* ID Viewfinder */}
        <IDViewfinder
          variant="front"
          processing={isLoading}
          processingLabel={t('demo.front.processingLabel')}
          previewUrl={previewUrl}
        />

        {!selectedFile ? (
          <Upload
            accept="image/jpeg,image/png,application/pdf"
            showUploadList={false}
            beforeUpload={handleBeforeUpload}
            disabled={isLoading || !documentType}
          >
            <DemoPrimaryBtn onClick={() => undefined} disabled={isLoading || !documentType}>
              {!documentType ? t('demo.front.selectFirst') : t('demo.front.chooseFront')}
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
                <p style={{ fontWeight: 500, color: D.text, fontSize: 13, margin: 0 }}>{selectedFile.name}</p>
                <p style={{ color: D.dim, fontSize: 11, margin: '2px 0 0' }}>{(selectedFile.size / 1024 / 1024).toFixed(2)} MB</p>
              </div>
              <Upload accept="image/jpeg,image/png,application/pdf" showUploadList={false} beforeUpload={handleBeforeUpload}>
                <Button type="link" style={{ color: D.accentInk, fontSize: 12, fontFamily: D.mono }}>
                  {t('demo.front.change')}
                </Button>
              </Upload>
            </div>
            <DemoPrimaryBtn onClick={onUpload} disabled={isLoading || !documentType}>
              {isLoading ? t('demo.front.processing') : t('demo.front.scanFront')}
            </DemoPrimaryBtn>
          </>
        )}
      </div>
    </div>
  );
};
