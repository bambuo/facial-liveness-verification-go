<template>
  <div class="demo-shell">
    <div class="demo-inner">
      <div class="demo-eyebrow">Live demo · v2 API</div>
      <h1 class="demo-heading">Identity verification demo</h1>
      <p class="demo-lede">
        Initialize a session, upload a document, capture a selfie and read the deterministic result.
        Every step below calls this deployment's live API with the API key you provide.
      </p>

      <div class="demo-privacy">
        <icon-safe class="demo-privacy-icon" />
        <p>
          <strong>Privacy.</strong>
          Documents are processed for this verification only. Face embeddings are held in memory for
          matching and are never stored.
        </p>
      </div>

      <a-card class="demo-card" :bordered="true">
        <a-steps v-if="visibleSteps.length" :current="currentIndex" size="small" class="demo-steps">
          <a-step
            v-for="step in visibleSteps"
            :key="step.key"
            :title="''"
            :description="step.description"
          />
        </a-steps>
        <!-- ── 1. Init ─────────────────────────────────────────── -->
        <section v-if="currentStep === 1" class="step">
          <h2 class="step-title">Start a verification</h2>
          <p class="step-lede">
            Use a developer API key. A sandbox key can be created for you when a developer session is
            active in this browser.
          </p>

          <a-form :model="initModel" layout="vertical" class="narrow">
            <a-form-item label="API key" extra="Keys are created in the developer portal.">
              <div class="key-row">
                <a-input v-model="apiKey" placeholder="ik_..." allow-clear class="grow" />
                <a-button :loading="fillingKey" @click="fillDeveloperKey">
                  {{ fillingKey ? 'Creating…' : 'Use developer key' }}
                </a-button>
              </div>
            </a-form-item>

            <a-form-item label="User ID" extra="A UUID that identifies your end user.">
              <a-input v-model="userId" placeholder="xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx" allow-clear />
            </a-form-item>

            <a-form-item label="Verification mode">
              <div class="mode-grid">
                <button
                  v-for="mode in MODES"
                  :key="mode.value"
                  type="button"
                  class="mode-card"
                  :class="{ active: verificationMode === mode.value }"
                  @click="verificationMode = mode.value"
                >
                  <strong>{{ mode.label }}</strong>
                  <span>{{ mode.description }}</span>
                </button>
              </div>
            </a-form-item>

            <a-form-item v-if="verificationMode === 'age_only'" label="Minimum age">
              <a-select v-model="ageThreshold">
                <a-option :value="18">18+</a-option>
                <a-option :value="21">21+</a-option>
                <a-option :value="25">25+</a-option>
              </a-select>
            </a-form-item>
          </a-form>

          <a-alert v-if="verificationMode !== 'full'" type="warning" class="inline-alert">
            This deployment initializes every verification in full mode. The selected mode is sent to
            the API but is not applied yet, so document, selfie and finalize steps still run.
          </a-alert>
          <a-alert v-if="initError" type="error" class="inline-alert">{{ initError }}</a-alert>

          <a-space direction="vertical" size="medium" class="narrow start-actions">
            <a-button
              type="primary"
              size="large"
              long
              :loading="isLoading"
              :disabled="!apiKey.trim() || !userId.trim()"
              @click="startVerification"
            >
              {{ isLoading ? 'Starting…' : 'Start on this device' }}
            </a-button>
            <router-link class="device-link" :to="{ path: '/verify/mobile', query: apiKey ? { api_key: apiKey } : {} }">
              Continue on a phone instead →
            </router-link>
          </a-space>
        </section>
        <!-- ── 2. Front document ───────────────────────────────── -->
        <section v-else-if="currentStep === 2" class="step">
          <div class="step-label">{{ stepCounter }}</div>
          <h2 class="step-title">Front of your ID</h2>
          <p class="step-lede">
            Passport, driving licence or national ID. Keep all four corners inside the frame.
          </p>

          <div class="narrow stack">
            <div class="field">
              <label class="field-label">Document type</label>
              <a-select v-model="documentType" placeholder="Select a document type" class="full">
                <a-option value="national_id">National ID</a-option>
                <a-option value="drivers_license">Driving licence</a-option>
                <a-option value="passport">Passport</a-option>
                <a-option value="other">Other</a-option>
              </a-select>
            </div>

            <div class="tip-bar"><i />A flat, well-lit photo gives the best OCR result.</div>

            <div class="id-viewfinder">
              <img v-if="previewUrl" :src="previewUrl" alt="Document preview" class="id-preview" />
              <div v-else class="id-mock">
                <div class="id-mock-top"><span class="id-chip" /><span class="id-photo" /></div>
                <div><span class="id-line w62" /><span class="id-line w42" /></div>
                <div><span class="id-line thin w90" /><span class="id-line thin w85" /></div>
              </div>
              <span class="corner tl" /><span class="corner tr" /><span class="corner bl" /><span class="corner br" />
              <div v-if="isLoading" class="viewfinder-overlay"><a-spin /><small>PROCESSING</small></div>
            </div>

            <a-upload
              accept="image/jpeg,image/png,application/pdf"
              :auto-upload="false"
              :show-file-list="false"
              draggable
              class="full"
              @change="onFrontChange"
            >
              <template #upload-button>
                <div class="dropzone">
                  <icon-upload class="dropzone-icon" />
                  <p>{{ selectedFile ? 'Choose a different file' : 'Click or drop the front of your document' }}</p>
                  <small>JPEG, PNG or PDF · up to 10 MB</small>
                </div>
              </template>
            </a-upload>

            <div v-if="selectedFile" class="file-row">
              <div>
                <p class="file-name">{{ selectedFile.name }}</p>
                <small class="muted">{{ fileSize(selectedFile.size) }}</small>
              </div>
              <a-tag color="gray">{{ documentType || 'no type' }}</a-tag>
            </div>

            <a-alert v-if="frontError" type="error">{{ frontError }}</a-alert>

            <a-button
              type="primary"
              size="large"
              long
              :loading="isLoading"
              :disabled="!selectedFile || !documentType || isLoading"
              @click="uploadFront"
            >
              {{ !documentType ? 'Select a document type first' : isLoading ? 'Scanning…' : 'Scan front of ID' }}
            </a-button>
          </div>
        </section>

        <!-- ── 3. Processing ───────────────────────────────────── -->
        <section v-else-if="currentStep === 3" class="step centered">
          <a-spin :size="36" />
          <p class="processing-message">{{ processingMessages[processingIndex] }}</p>
          <small class="muted mono">OCR AND CLASSIFICATION RUNNING</small>
          <a-space wrap class="processing-tags">
            <a-tag v-for="tag in processingTags" :key="tag" color="arcoblue">{{ tag }}</a-tag>
          </a-space>
        </section>
        <!-- ── 4. Back document ────────────────────────────────── -->
        <section v-else-if="currentStep === 4" class="step">
          <div class="step-label">{{ stepCounter }}</div>
          <h2 class="step-title">Back of your ID</h2>
          <p class="step-lede">
            The reverse side carries the MRZ. Keep glare and shadows away from the machine-readable zone.
          </p>

          <div class="narrow stack">
            <a-descriptions v-if="hasOcr" :column="1" size="small" bordered title="Extracted from the front">
              <a-descriptions-item v-if="ocr.full_name" label="Name">{{ ocr.full_name }}</a-descriptions-item>
              <a-descriptions-item v-if="ocr.date_of_birth" label="Date of birth">{{ ocr.date_of_birth }}</a-descriptions-item>
              <a-descriptions-item v-if="ocr.id_number || ocr.document_number" label="Document number">
                {{ ocr.id_number || ocr.document_number }}
              </a-descriptions-item>
              <a-descriptions-item v-if="ocr.expiry_date" label="Expires">{{ ocr.expiry_date }}</a-descriptions-item>
            </a-descriptions>

            <div class="tip-bar"><i>MRZ lines are checksum-validated when present.</i></div>

            <div class="id-viewfinder">
              <img v-if="backPreviewUrl" :src="backPreviewUrl" alt="Back document preview" class="id-preview" />
              <div v-else class="id-mock">
                <div class="mrz-lines">
                  <span v-for="line in 16" :key="line" :style="{ width: `${48 + Math.abs(Math.sin(line * 1.3)) * 38}%` }" />
                </div>
                <div><span class="id-line thin w90" /><span class="id-line thin w85" /><span class="id-line thin w80" /></div>
              </div>
              <span class="corner tl" /><span class="corner tr" /><span class="corner bl" /><span class="corner br" />
              <div v-if="isLoading" class="viewfinder-overlay"><a-spin /><small>PROCESSING</small></div>
            </div>

            <a-upload
              accept="image/jpeg,image/png,image/webp"
              :auto-upload="false"
              :show-file-list="false"
              draggable
              class="full"
              @change="onBackChange"
            >
              <template #upload-button>
                <div class="dropzone">
                  <icon-upload class="dropzone-icon" />
                  <p>{{ backFile ? 'Choose a different file' : 'Click or drop the back of your document' }}</p>
                  <small>JPEG, PNG or WebP · up to 10 MB</small>
                </div>
              </template>
            </a-upload>

            <div v-if="backFile" class="file-row">
              <div>
                <p class="file-name">{{ backFile.name }}</p>
                <small class="muted">{{ fileSize(backFile.size) }}</small>
              </div>
              <a-tag color="gray">{{ documentType || 'no type' }}</a-tag>
            </div>

            <a-alert v-if="backError" type="error">{{ backError }}</a-alert>

            <a-button
              type="primary"
              size="large"
              long
              :loading="isLoading"
              :disabled="!backFile || isLoading"
              @click="uploadBack"
            >
              {{ isLoading ? 'Uploading…' : 'Scan back of ID' }}
            </a-button>
          </div>
        </section>

        <!-- ── 5. Cross-checking ───────────────────────────────── -->
        <section v-else-if="currentStep === 5" class="step centered">
          <div class="step-label">{{ stepCounter }}</div>
          <a-spin :size="36" />
          <p class="processing-message">{{ checkingMessages[checkingIndex] }}</p>
          <small class="muted mono">CROSS-VALIDATING DOCUMENT FIELDS</small>
          <a-space wrap class="processing-tags">
            <a-tag v-for="tag in checkingTags" :key="tag" color="arcoblue">{{ tag }}</a-tag>
          </a-space>
          <a-alert v-if="checkingError" type="error" class="inline-alert">
            {{ checkingError }}
            <template #action>
              <a-button size="mini" @click="runCrossValidation">Retry</a-button>
            </template>
          </a-alert>
        </section>
        <!-- ── 6. Live capture ─────────────────────────────────── -->
        <section v-else-if="currentStep === 6" class="step">
          <div class="step-label">{{ stepCounter }}</div>
          <h2 class="step-title">Live selfie</h2>
          <p class="step-lede">
            Look straight at the camera in even light. The selfie is compared with the portrait on your
            document and scored for liveness.
          </p>

          <div class="narrow stack">
            <div class="face-viewfinder" :class="{ live: cameraState === 'ready' }">
              <video v-if="cameraState === 'ready'" ref="videoEl" autoplay playsinline muted class="face-video" />
              <div v-else class="face-placeholder">
                <icon-user class="face-icon" />
                <span class="face-dots" />
              </div>
              <span class="face-ring" />
              <div v-if="cameraState === 'starting'" class="viewfinder-overlay face-overlay">
                <a-spin /><small>STARTING CAMERA</small>
              </div>
              <div v-else-if="cameraState === 'ready' && isLoading" class="viewfinder-overlay face-overlay">
                <a-spin /><small>CHECKING</small>
              </div>
            </div>

            <div class="cue-row">
              <div
                v-for="(cue, index) in LIVENESS_CUES"
                :key="cue.label"
                class="cue"
                :class="{ active: cueIndex === index }"
              >
                <span class="cue-icon">{{ cue.icon }}</span>
                <small>{{ cue.label }}</small>
              </div>
            </div>

            <div class="tip-bar">
              <i>Hold still for a second; the selfie must be a live capture, not a photo of a photo.</i>
            </div>

            <a-alert v-if="cameraError" type="error">{{ cameraError }}</a-alert>
            <a-alert v-if="liveError" type="error">{{ liveError }}</a-alert>

            <template v-if="cameraState !== 'ready'">
              <a-button type="primary" size="large" long :loading="cameraState === 'starting'" @click="startCamera">
                Start camera
              </a-button>

              <a-upload
                accept="image/jpeg,image/png"
                :auto-upload="false"
                :show-file-list="false"
                :disabled="isLoading"
                @change="onSelfieChange"
              >
                <template #upload-button>
                  <a-button long :loading="isLoading">Upload a selfie instead</a-button>
                </template>
              </a-upload>

              <a-checkbox v-model="voiceEnabled">
                Add a voice challenge step after the selfie
              </a-checkbox>

              <a-button long :disabled="isLoading" @click="skipLiveCapture">Skip live capture</a-button>
            </template>

            <template v-else>
              <a-space class="live-actions">
                <a-button type="primary" size="large" :loading="isLoading" @click="captureSelfie">Capture selfie</a-button>
                <a-button size="large" :disabled="isLoading" @click="stopCamera">Cancel</a-button>
              </a-space>
            </template>
          </div>
        </section>
        <!-- ── 7. Voice challenge ──────────────────────────────── -->
        <section v-else-if="currentStep === 7" class="step">
          <div class="step-label">{{ stepCounter }}</div>
          <h2 class="step-title">Voice challenge</h2>
          <p class="step-lede">
            Read the digits aloud in one continuous take. The recording is transcribed and matched to the
            challenge before the session is finalized.
          </p>

          <div class="narrow stack">
            <div class="voice-orb" :class="{ ready: !!voiceChallenge }">
              <icon-voice class="voice-icon" />
            </div>

            <div v-if="voiceChallenge && !voiceExpired" class="challenge-box">
              <small class="mono">SAY THESE DIGITS</small>
              <div class="challenge-digits">{{ voiceChallenge }}</div>
              <small v-if="voiceExpiresIn !== null" :class="voiceExpiresIn < 30 ? 'danger' : 'muted'">
                Expires in {{ voiceExpiresIn }}s
              </small>
            </div>

            <div v-if="voiceChallenge && voiceExpired" class="tip-bar danger">
              <i>The challenge expired. Request a new one before uploading.</i>
            </div>

            <div class="tip-bar"><i>Upload a mono 16 kHz WAV file; browsers record WebM, which this API rejects.</i></div>

            <a-upload
              accept="audio/wav,audio/x-wav,.wav"
              :auto-upload="false"
              :show-file-list="false"
              :disabled="isLoading || !voiceChallenge"
              @change="onVoiceChange"
            >
              <template #upload-button>
                <a-button long :disabled="isLoading || !voiceChallenge">
                  {{ voiceFile ? `Selected: ${voiceFile.name}` : 'Choose a WAV recording' }}
                </a-button>
              </template>
            </a-upload>

            <a-alert v-if="voiceError" type="error">{{ voiceError }}</a-alert>

            <a-button v-if="!voiceChallenge" type="primary" size="large" long :loading="isLoading" @click="requestVoiceChallenge">
              Get challenge digits
            </a-button>
            <template v-else-if="voiceExpired">
              <a-button type="primary" size="large" long :loading="isLoading" @click="restartVoiceChallenge">
                Request a new challenge
              </a-button>
            </template>
            <a-button
              v-else
              type="primary"
              size="large"
              long
              :loading="isLoading"
              :disabled="!voiceFile || isLoading"
              @click="submitVoice"
            >
              {{ isLoading ? 'Verifying…' : 'Submit voice sample' }}
            </a-button>

            <a-button long :disabled="isLoading" @click="skipVoice">Skip voice and finish</a-button>
          </div>
        </section>
        <!-- ── 8. Results ──────────────────────────────────────── -->
        <section v-else-if="currentStep === 8" class="step">
          <a-result :status="resultStatus" :title="resultTitle" :subtitle="resultSubtitle">
            <template #extra>
              <a-space wrap>
                <a-button
                  v-if="finalResult === 'verified'"
                  type="primary"
                  @click="goToCredential"
                >
                  Get verifiable credential
                </a-button>
                <a-button @click="currentStep = 9">Try address verification</a-button>
                <a-button @click="resetDemo">Start new demo</a-button>
              </a-space>
            </template>
          </a-result>

          <a-descriptions :column="1" size="small" bordered title="Overview" class="result-block">
            <a-descriptions-item label="Verification ID">
              <span class="mono">{{ verification?.verification_id || verificationId || '—' }}</span>
            </a-descriptions-item>
            <a-descriptions-item label="Final result">
              <a-tag :color="finalResult === 'verified' ? 'green' : finalResult === 'failed' ? 'red' : 'orange'">
                {{ finalResult || 'pending' }}
              </a-tag>
            </a-descriptions-item>
            <a-descriptions-item label="Session status">
              <span class="mono">{{ verification?.status || '—' }}</span>
            </a-descriptions-item>
            <a-descriptions-item label="Pipeline step">
              <span class="mono">{{ verification?.current_step ?? '—' }} / {{ verification?.total_steps ?? '—' }}</span>
            </a-descriptions-item>
            <a-descriptions-item label="Mode">
              {{ verification?.verification_mode || verificationMode }}
              <a-tag v-if="verification?.is_sandbox" color="gray" class="ml-6">sandbox</a-tag>
            </a-descriptions-item>
          </a-descriptions>

          <div class="result-grid">
            <a-card class="result-card" :bordered="true">
              <template #title>Cross-validation</template>
              <template v-if="stepDetails.crossValidation">
                <div class="score">
                  <div class="score-head">
                    <span>Field score</span>
                    <strong>{{ percent(stepDetails.crossValidation.score) }}</strong>
                  </div>
                  <a-progress
                    :percent="clamp01(stepDetails.crossValidation.score)"
                    :show-text="false"
                    size="small"
                    :color="toneColor(stepDetails.crossValidation.score)"
                  />
                </div>
                <div class="kv">
                  <span>Documents match</span>
                  <strong>{{ stepDetails.documentsMatch ? 'Yes' : 'No' }}</strong>
                </div>
              </template>
              <span v-else class="muted">Not completed</span>
            </a-card>

            <a-card class="result-card" :bordered="true">
              <template #title>Liveness</template>
              <template v-if="stepDetails.liveness">
                <div class="score">
                  <div class="score-head">
                    <span>Liveness score</span>
                    <strong>{{ percent(stepDetails.liveness.liveness_score) }}</strong>
                  </div>
                  <a-progress
                    :percent="clamp01(stepDetails.liveness.liveness_score)"
                    :show-text="false"
                    size="small"
                    :color="stepDetails.liveness.liveness_passed ? 'rgb(var(--green-6))' : 'rgb(var(--red-6))'"
                  />
                </div>
                <div class="kv">
                  <span>Result</span>
                  <a-tag :color="stepDetails.liveness.liveness_passed ? 'green' : 'red'">
                    {{ stepDetails.liveness.liveness_passed ? 'passed' : 'failed' }}
                  </a-tag>
                </div>
              </template>
              <span v-else class="muted">Not completed</span>
            </a-card>

            <a-card class="result-card" :bordered="true">
              <template #title>Face detection</template>
              <template v-if="stepDetails.faceDetection !== null">
                <div class="score">
                  <div class="score-head">
                    <span>Detection confidence</span>
                    <strong>{{ percent(stepDetails.faceDetection) }}</strong>
                  </div>
                  <a-progress
                    :percent="clamp01(stepDetails.faceDetection)"
                    :show-text="false"
                    size="small"
                    color="rgb(var(--arcoblue-6))"
                  />
                </div>
                <small class="muted">Similarity vs. document portrait is not exposed by the status endpoint.</small>
              </template>
              <span v-else class="muted">Not completed</span>
            </a-card>

            <a-card v-if="stepDetails.voice" class="result-card" :bordered="true">
              <template #title>Voice</template>
              <div class="score">
                <div class="score-head">
                  <span>Match score</span>
                  <strong>{{ percent(stepDetails.voice.score) }}</strong>
                </div>
                <a-progress
                  :percent="clamp01(stepDetails.voice.score)"
                  :show-text="false"
                  size="small"
                  :color="toneColor(stepDetails.voice.score)"
                />
              </div>
              <div class="kv" v-if="stepDetails.voice.transcription">
                <span>Transcript</span>
                <span class="mono">{{ stepDetails.voice.transcription }}</span>
              </div>
            </a-card>
          </div>

          <a-alert v-if="finalResult === 'failed'" type="error" class="result-block">
            {{ verification?.message || 'The verification was rejected by one of the gates.' }}
            <template #action>
              <a-button v-if="retryCount < 3" size="mini" :loading="retrying" @click="retryVerification">Try again</a-button>
              <span v-else class="muted">Retry limit reached</span>
            </template>
          </a-alert>

          <a-collapse :bordered="false" class="result-block">
            <a-collapse-item header="Raw API response" key="raw">
              <pre class="demo-code">{{ rawPayload }}</pre>
            </a-collapse-item>
          </a-collapse>
        </section>
        <!-- ── 9. Address (optional) ───────────────────────────── -->
        <section v-else-if="currentStep === 9" class="step">
          <h2 class="step-title">
            Proof of address
            <small class="optional">optional</small>
          </h2>
          <p class="step-lede">
            Upload a recent utility bill, bank statement or tax document. The name is matched against the
            identity document on file.
          </p>

          <div v-if="!addressResult" class="narrow stack">
            <div class="field">
              <label class="field-label">Document type</label>
              <a-select v-model="addressDocType" class="full">
                <a-option value="utility_bill">Utility bill</a-option>
                <a-option value="bank_statement">Bank statement</a-option>
                <a-option value="tax_document">Tax document</a-option>
              </a-select>
            </div>

            <a-upload
              accept="image/jpeg,image/png,application/pdf"
              :auto-upload="false"
              :show-file-list="false"
              draggable
              class="full"
              @change="onAddressChange"
            >
              <template #upload-button>
                <div class="dropzone">
                  <icon-upload class="dropzone-icon" />
                  <p>{{ addressFile ? 'Choose a different file' : 'Click or drop your address document' }}</p>
                  <small>JPEG, PNG or PDF · up to 10 MB</small>
                </div>
              </template>
            </a-upload>

            <div v-if="addressFile" class="file-row">
              <div>
                <p class="file-name">{{ addressFile.name }}</p>
                <small class="muted">{{ fileSize(addressFile.size) }}</small>
              </div>
              <img v-if="addressPreview" :src="addressPreview" alt="Address document preview" class="thumb" />
            </div>

            <a-alert v-if="addressError" type="error">{{ addressError }}</a-alert>

            <a-button
              v-if="addressFile"
              type="primary"
              size="large"
              long
              :loading="addressUploading"
              @click="uploadAddress"
            >
              {{ addressUploading ? 'Processing…' : 'Verify address document' }}
            </a-button>

            <a-button long @click="resetDemo">Skip and start a new demo</a-button>
          </div>

          <div v-else class="narrow stack">
            <a-card class="result-card" :bordered="true">
              <a-result
                :status="addressStatusTone"
                :title="addressStatusLabel"
                class="address-result"
              />
              <a-descriptions :column="1" size="small" bordered>
                <a-descriptions-item label="Overall score">{{ percent(addressResult.score) }}</a-descriptions-item>
                <a-descriptions-item label="Name match">{{ percent(addressResult.name_match_score) }}</a-descriptions-item>
                <a-descriptions-item label="Document fresh">{{ addressResult.document_fresh ? 'Yes' : 'Unknown' }}</a-descriptions-item>
                <a-descriptions-item v-if="addressResult.address" label="Extracted address">
                  {{ addressResult.address }}
                </a-descriptions-item>
              </a-descriptions>
              <a-alert v-if="addressResult.reasons?.length" type="warning" class="inline-alert">
                <template #title>Validation notes</template>
                <ul class="note-list">
                  <li v-for="(reason, index) in addressResult.reasons" :key="index">{{ reason }}</li>
                </ul>
              </a-alert>
            </a-card>
            <a-button type="primary" long @click="resetDemo">Start a new demo</a-button>
          </div>
        </section>

        <!-- ── 10. Credential ──────────────────────────────────── -->
        <section v-else-if="currentStep === 10" class="step">
          <h2 class="step-title">Verifiable credential</h2>

          <template v-if="!credential">
            <p class="step-lede">
              A W3C JWT-VC is signed by this deployment for verified identities. Credentials are stored per
              verification and can be checked or revoked at any time.
            </p>

            <a-card class="result-card narrow" :bordered="true">
              <a-space direction="vertical" size="small" class="full">
                <div class="credential-step"><span class="step-chip">1</span><div><strong>Issue</strong><p>The API signs a credential for the verified subject.</p></div></div>
                <div class="credential-step"><span class="step-chip">2</span><div><strong>Store</strong><p>Your customer keeps the JWT; only its JTI is tracked here.</p></div></div>
                <div class="credential-step"><span class="step-chip">3</span><div><strong>Verify</strong><p>Anyone can check the JTI status to confirm it is still active.</p></div></div>
              </a-space>
            </a-card>

            <a-alert v-if="credentialError" type="error" class="inline-alert">{{ credentialError }}</a-alert>

            <a-space direction="vertical" size="medium" class="narrow">
              <a-button type="primary" size="large" long :loading="credentialLoading" @click="issueCredential">
                {{ credentialLoading ? 'Issuing…' : 'Issue credential' }}
              </a-button>
              <a-button long @click="currentStep = 8">Back to results</a-button>
              <a-button long @click="resetDemo">Start new demo</a-button>
            </a-space>
          </template>

          <template v-else>
            <p class="step-lede">
              {{ credentialAlreadyIssued ? 'This verification already holds an active credential.' : 'Credential issued for this verification.' }}
            </p>

            <a-descriptions :column="1" size="small" bordered>
              <a-descriptions-item label="JTI"><span class="mono break">{{ credential.jti }}</span></a-descriptions-item>
              <a-descriptions-item v-if="credentialParts" label="Algorithm">
                <span class="mono">{{ credentialParts.header?.alg || '—' }}</span>
              </a-descriptions-item>
              <a-descriptions-item v-if="credentialParts" label="Issuer">
                <span class="mono break">{{ credentialParts.payload?.iss || '—' }}</span>
              </a-descriptions-item>
              <a-descriptions-item v-if="credentialParts" label="Type">
                <span class="mono">{{ (credentialParts.payload?.vc?.type || []).join(', ') || '—' }}</span>
              </a-descriptions-item>
              <a-descriptions-item label="Expires">{{ formatDate(credential.expires_at) }}</a-descriptions-item>
            </a-descriptions>

            <a-descriptions
              v-if="credentialClaims"
              :column="1"
              size="small"
              bordered
              title="Credential subject"
              class="result-block"
            >
              <a-descriptions-item v-for="(value, key) in credentialClaims" :key="key" :label="String(key)">
                <span class="mono">{{ value === null ? 'null' : String(value) }}</span>
              </a-descriptions-item>
            </a-descriptions>

            <a-card v-if="credential.jwt" class="result-card result-block" :bordered="true">
              <template #title>
                <div class="card-head">
                  <span>Raw JWT</span>
                  <a-button size="mini" @click="copyJwt">{{ copied ? 'Copied' : 'Copy' }}</a-button>
                </div>
              </template>
              <pre class="demo-code">{{ credential.jwt }}</pre>
            </a-card>

            <a-alert v-if="credentialStatus" :type="credentialStatus.active ? 'success' : 'error'" class="result-block">
              {{ credentialStatus.active ? 'Credential is active.' : `Credential is inactive: ${credentialStatus.reason || 'unknown'}` }}
            </a-alert>
            <a-alert v-if="credentialError" type="error" class="result-block">{{ credentialError }}</a-alert>

            <a-space wrap class="result-block">
              <a-button :loading="checkingCredential" @click="checkCredentialStatus">Check status</a-button>
              <a-button status="danger" :loading="revoking" @click="revokeCredential">Revoke</a-button>
              <a-button @click="emailOpen = !emailOpen">Send by email</a-button>
            </a-space>

            <a-card v-if="emailOpen" class="result-card result-block" :bordered="true">
              <template #title>Send credential</template>
              <div class="send-row">
                <a-input v-model="sendEmail" placeholder="holder@example.com" class="grow" @press-enter="sendCredential" />
                <a-button type="primary" :loading="sending" :disabled="!sendEmail.trim()" @click="sendCredential">
                  Send
                </a-button>
              </div>
              <a-alert v-if="sendResult" type="success" class="inline-alert">{{ sendResult }}</a-alert>
            </a-card>

            <a-space wrap class="result-block">
              <a-button @click="currentStep = 8">Back to results</a-button>
              <a-button @click="resetDemo">Start new demo</a-button>
            </a-space>
          </template>
        </section>
      </a-card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue'
import { Message } from '@arco-design/web-vue'
import { useRoute, useRouter } from 'vue-router'
import { api, apiMessage } from '../lib/api'
import { useAuthStore } from '../stores/auth'

type VerificationMode = 'full' | 'document_only' | 'identity' | 'age_only'

const MODES: Array<{ value: VerificationMode; label: string; description: string }> = [
  { value: 'full', label: 'Full', description: 'Document, selfie and face match' },
  { value: 'document_only', label: 'Document only', description: 'OCR and cross-validation' },
  { value: 'identity', label: 'Identity', description: 'Document and selfie only' },
  { value: 'age_only', label: 'Age only', description: 'Estimate age from the ID' },
]

const LIVENESS_CUES = [
  { icon: '😐', label: 'Look ahead' },
  { icon: '🙂', label: 'Smile' },
  { icon: '↔', label: 'Turn slightly' },
]

const PROCESSING_MESSAGES = [
  'Running OCR on the document',
  'Classifying the document type',
  'Extracting the machine-readable zone',
]

const PROCESSING_TAGS = ['OCR', 'CLASSIFICATION', 'FIELD EXTRACTION']

const CHECKING_MESSAGES = [
  'Comparing document fields',
  'Checking the expiry date',
  'Preparing the selfie gate',
]

const CHECKING_TAGS = ['CROSS-VALIDATION', 'MRZ', 'EXPIRY']

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

// 1 init · 2 front · 3 processing · 4 back · 5 checking · 6 live · 7 voice · 8 results · 9 address · 10 credential
const currentStep = ref(1)
const isLoading = ref(false)
const initError = ref('')

const apiKey = ref(typeof route.query.api_key === 'string' ? route.query.api_key : auth.apiKey)
const fillingKey = ref(false)
const userId = ref(crypto.randomUUID())
const verificationMode = ref<VerificationMode>('full')
const ageThreshold = ref(18)
const initModel = reactive({})

const verificationId = ref('')
const verification = ref<any>(null)
const stepDetails = reactive<{
  ocr: any
  back: any
  crossValidation: any
  documentsMatch: boolean | null
  liveness: any
  faceDetection: number | null
  voice: any
}>({ ocr: null, back: null, crossValidation: null, documentsMatch: null, liveness: null, faceDetection: null, voice: null })

const selectedFile = ref<File | null>(null)
const previewUrl = ref<string | null>(null)
const documentType = ref('')
const frontError = ref('')
const skipBack = ref(false)

const backFile = ref<File | null>(null)
const backPreviewUrl = ref<string | null>(null)
const backError = ref('')

const checkingError = ref('')

const cameraState = ref<'idle' | 'starting' | 'ready' | 'error'>('idle')
const cameraError = ref('')
const liveError = ref('')
const videoEl = ref<HTMLVideoElement | null>(null)
let mediaStream: MediaStream | null = null

const voiceEnabled = ref(false)
const voiceAuthEnabled = ref(false)
const voiceChallenge = ref<string | null>(null)
const voiceExpiresIn = ref<number | null>(null)
const voiceFile = ref<File | null>(null)
const voiceError = ref('')

const addressFile = ref<File | null>(null)
const addressPreview = ref<string | null>(null)
const addressDocType = ref('utility_bill')
const addressResult = ref<any>(null)
const addressUploading = ref(false)
const addressError = ref('')

const credential = ref<{ jwt: string; jti: string; expires_at: string } | null>(null)
const credentialAlreadyIssued = ref(false)
const credentialError = ref('')
const credentialLoading = ref(false)
const credentialStatus = ref<{ active: boolean; reason?: string } | null>(null)
const checkingCredential = ref(false)
const revoking = ref(false)
const copied = ref(false)
const emailOpen = ref(false)
const sendEmail = ref('')
const sending = ref(false)
const sendResult = ref('')

const retryCount = ref(0)
const retrying = ref(false)

const processingIndex = ref(0)
const checkingIndex = ref(0)
const cueIndex = ref(0)
let processingTimer: ReturnType<typeof setTimeout> | null = null
let voiceCountdown: ReturnType<typeof setInterval> | null = null
const rotationTimers: Array<ReturnType<typeof setInterval>> = []

// ── Init ────────────────────────────────────────────────────

async function fillDeveloperKey() {
  const cached = sessionStorage.getItem('iv_demo_sandbox_key')
  if (cached) {
    apiKey.value = cached
    if (!userId.value) userId.value = crypto.randomUUID()
    Message.success('Sandbox key loaded')
    return
  }
  fillingKey.value = true
  try {
    const { data } = await api.post('/developer/api-key', {
      name: `Demo sandbox ${new Date().toISOString().slice(0, 16).replace('T', ' ')}`,
      is_sandbox: true,
    })
    const created = typeof data.api_key === 'string' ? data.api_key : data.api_key?.key
    if (!created) throw new Error('Key creation failed')
    sessionStorage.setItem('iv_demo_sandbox_key', created)
    apiKey.value = created
    if (!userId.value) userId.value = crypto.randomUUID()
    Message.success('Sandbox key created')
  } catch (error) {
    Message.error(apiMessage(error))
  } finally {
    fillingKey.value = false
  }
}

async function startVerification() {
  const key = apiKey.value.trim()
  if (!key) { Message.warning('Enter an API key'); return }
  if (!userId.value.trim()) { Message.warning('Enter a user ID'); return }
  isLoading.value = true
  initError.value = ''
  try {
    auth.setAPIKey(key)
    const body: Record<string, unknown> = { user_id: userId.value.trim(), source: 'demo' }
    if (verificationMode.value !== 'full') {
      body.verification_mode = verificationMode.value
      if (verificationMode.value === 'age_only') body.age_threshold = ageThreshold.value
    }
    const { data } = await api.post('/v2/verify/initialize', body)
    verificationId.value = data.verification_id
    verification.value = { ...data }
    if (data.session_token) auth.setSessionToken(data.session_token)
    resetDocumentState()
    currentStep.value = 2
    Message.success('Verification session started')
  } catch (error) {
    initError.value = apiMessage(error)
  } finally {
    isLoading.value = false
  }
}

// ── Front document ──────────────────────────────────────────

function fileSize(bytes: number) {
  return `${(bytes / 1024 / 1024).toFixed(2)} MB`
}

function onFrontChange(_fileList: unknown, item: { file?: File }) {
  const file = item?.file
  if (!file) return
  if (!['image/jpeg', 'image/png', 'application/pdf'].includes(file.type)) {
    Message.error('Use a JPEG, PNG or PDF file')
    return
  }
  if (file.size > 10 * 1024 * 1024) {
    Message.error('File must be under 10 MB')
    return
  }
  selectedFile.value = file
  frontError.value = ''
  if (previewUrl.value) URL.revokeObjectURL(previewUrl.value)
  previewUrl.value = file.type.startsWith('image/') ? URL.createObjectURL(file) : null
}

async function uploadFront() {
  if (!selectedFile.value || !verificationId.value) { Message.warning('Choose a document first'); return }
  isLoading.value = true
  frontError.value = ''
  try {
    const form = new FormData()
    form.append('document', selectedFile.value)
    form.append('document_type', documentType.value || 'national_id')
    const { data } = await api.post(`/v2/verify/${verificationId.value}/front-document`, form)
    stepDetails.ocr = data.ocr_data || null
    skipBack.value = data.requires_back === false
    currentStep.value = 3
    Message.success('Front document accepted')
    if (processingTimer) clearTimeout(processingTimer)
    processingTimer = setTimeout(() => {
      if (currentStep.value === 3) currentStep.value = skipBack.value ? 5 : 4
    }, 1400)
  } catch (error) {
    frontError.value = apiMessage(error)
  } finally {
    isLoading.value = false
  }
}

// ── Back document ───────────────────────────────────────────

function onBackChange(_fileList: unknown, item: { file?: File }) {
  const file = item?.file
  if (!file) return
  if (!['image/jpeg', 'image/png', 'image/webp'].includes(file.type)) {
    Message.error('Use a JPEG, PNG or WebP file')
    return
  }
  if (file.size > 10 * 1024 * 1024) {
    Message.error('File must be under 10 MB')
    return
  }
  backFile.value = file
  backError.value = ''
  if (backPreviewUrl.value) URL.revokeObjectURL(backPreviewUrl.value)
  backPreviewUrl.value = URL.createObjectURL(file)
}

async function uploadBack() {
  if (!backFile.value || !verificationId.value) { Message.warning('Choose the back of your document'); return }
  isLoading.value = true
  backError.value = ''
  try {
    const form = new FormData()
    form.append('document', backFile.value)
    form.append('document_type', documentType.value || 'national_id')
    const { data } = await api.post(`/v2/verify/${verificationId.value}/back-document`, form)
    stepDetails.back = data
    stepDetails.documentsMatch = typeof data.documents_match === 'boolean' ? data.documents_match : null
    currentStep.value = 5
    await runCrossValidation()
  } catch (error) {
    backError.value = apiMessage(error)
  } finally {
    isLoading.value = false
  }
}

// ── Cross-validation ────────────────────────────────────────

async function runCrossValidation() {
  if (!verificationId.value) return
  checkingError.value = ''
  try {
    const { data } = await api.post(`/v2/verify/${verificationId.value}/cross-validation`)
    stepDetails.crossValidation = data.cross_validation_results || { score: data.documents_match ? 1 : 0 }
    if (typeof data.documents_match === 'boolean') stepDetails.documentsMatch = data.documents_match
    currentStep.value = 6
  } catch (error) {
    checkingError.value = apiMessage(error)
  }
}

// ── Live capture ────────────────────────────────────────────

async function startCamera() {
  cameraError.value = ''
  liveError.value = ''
  if (!navigator.mediaDevices?.getUserMedia) {
    cameraError.value = 'This browser does not expose a camera API.'
    return
  }
  cameraState.value = 'starting'
  try {
    mediaStream = await navigator.mediaDevices.getUserMedia({
      video: { width: { ideal: 640 }, height: { ideal: 480 }, facingMode: 'user' },
      audio: false,
    })
    cameraState.value = 'ready'
    await nextTick()
    if (videoEl.value) {
      videoEl.value.srcObject = mediaStream
      await videoEl.value.play().catch(() => undefined)
    }
  } catch (error) {
    cameraState.value = 'error'
    const name = (error as { name?: string }).name
    cameraError.value =
      name === 'NotAllowedError' ? 'Camera permission was denied.'
        : name === 'NotFoundError' ? 'No camera was found on this device.'
          : name === 'NotReadableError' ? 'The camera is already in use by another app.'
            : 'The camera could not be started.'
  }
}

function stopCamera() {
  mediaStream?.getTracks().forEach((track) => track.stop())
  mediaStream = null
  if (videoEl.value) videoEl.value.srcObject = null
  cameraState.value = 'idle'
}

async function captureSelfie() {
  const video = videoEl.value
  if (!video || !video.videoWidth) {
    cameraError.value = 'The camera is not ready yet.'
    return
  }
  const canvas = document.createElement('canvas')
  canvas.width = video.videoWidth
  canvas.height = video.videoHeight
  const context = canvas.getContext('2d')
  if (!context) {
    cameraError.value = 'Capturing a frame failed.'
    return
  }
  context.drawImage(video, 0, 0, canvas.width, canvas.height)
  const blob = await new Promise<Blob | null>((resolve) => canvas.toBlob(resolve, 'image/jpeg', 0.85))
  if (!blob) {
    cameraError.value = 'Capturing a frame failed.'
    return
  }
  await submitSelfie(blob, 'selfie.jpg')
}

async function onSelfieChange(_fileList: unknown, item: { file?: File }) {
  const file = item?.file
  if (!file) return
  if (!['image/jpeg', 'image/png'].includes(file.type)) {
    Message.error('Use a JPEG or PNG selfie')
    return
  }
  await submitSelfie(file, file.name)
}

async function submitSelfie(blob: Blob, filename: string) {
  if (!verificationId.value) { Message.warning('Start a verification first'); return }
  isLoading.value = true
  liveError.value = ''
  try {
    const form = new FormData()
    form.append('selfie', blob, filename)
    const { data } = await api.post(`/v2/verify/${verificationId.value}/live-capture`, form)
    stepDetails.liveness = data.liveness_results || null
    stepDetails.faceDetection = data.face_match_results?.confidence ?? null
    stopCamera()
    Message.success('Selfie captured')
    if (voiceEnabled.value) {
      voiceAuthEnabled.value = true
      currentStep.value = 7
    } else {
      await finalizeVerification()
    }
  } catch (error) {
    liveError.value = apiMessage(error)
  } finally {
    isLoading.value = false
  }
}

async function skipLiveCapture() {
  stopCamera()
  await refreshStatus()
  currentStep.value = 8
  Message.info('Continuing without a live capture')
}

async function finalizeVerification() {
  try {
    const { data } = await api.post(`/v2/verify/${verificationId.value}/finalize`)
    verification.value = { ...(verification.value || {}), ...data }
  } catch (error) {
    liveError.value = apiMessage(error)
  }
  await refreshStatus()
  currentStep.value = 8
}

async function refreshStatus() {
  if (!verificationId.value) return
  try {
    const { data } = await api.get(`/v2/verify/${verificationId.value}/status`)
    verification.value = { ...(verification.value || {}), ...data }
  } catch {
    // Keep the last known snapshot when the status call fails.
  }
}

// ── Voice challenge ─────────────────────────────────────────

async function requestVoiceChallenge() {
  if (!verificationId.value) return
  isLoading.value = true
  voiceError.value = ''
  try {
    const { data } = await api.post(`/v2/verify/${verificationId.value}/voice-challenge`)
    voiceChallenge.value = data.challenge_digits
    startVoiceCountdown(Number(data.expires_in_seconds) || 0)
  } catch (error) {
    voiceError.value = apiMessage(error)
  } finally {
    isLoading.value = false
  }
}

function startVoiceCountdown(seconds: number) {
  if (voiceCountdown) clearInterval(voiceCountdown)
  voiceExpiresIn.value = seconds
  const started = Date.now()
  voiceCountdown = setInterval(() => {
    const remaining = seconds - Math.floor((Date.now() - started) / 1000)
    voiceExpiresIn.value = remaining > 0 ? remaining : 0
    if (remaining <= 0 && voiceCountdown) {
      clearInterval(voiceCountdown)
      voiceCountdown = null
    }
  }, 1000)
}

async function restartVoiceChallenge() {
  voiceChallenge.value = null
  voiceFile.value = null
  voiceError.value = ''
  await requestVoiceChallenge()
}

function onVoiceChange(_fileList: unknown, item: { file?: File }) {
  const file = item?.file
  if (!file) return
  if (!/\.wav$/i.test(file.name) && !file.type.includes('wav')) {
    Message.error('The voice endpoint accepts WAV only')
    return
  }
  voiceFile.value = file
  voiceError.value = ''
}

async function submitVoice() {
  if (!verificationId.value || !voiceFile.value) { Message.warning('Choose a WAV recording'); return }
  isLoading.value = true
  voiceError.value = ''
  try {
    const form = new FormData()
    form.append('file', voiceFile.value, voiceFile.value.name)
    const { data } = await api.post(`/v2/verify/${verificationId.value}/voice-capture`, form)
    stepDetails.voice = data.voice_match_results || null
    if (voiceCountdown) { clearInterval(voiceCountdown); voiceCountdown = null }
    Message.success('Voice sample processed')
    await finalizeVerification()
  } catch (error) {
    voiceError.value = apiMessage(error)
  } finally {
    isLoading.value = false
  }
}

async function skipVoice() {
  if (voiceCountdown) { clearInterval(voiceCountdown); voiceCountdown = null }
  await finalizeVerification()
}

// ── Address document ────────────────────────────────────────

function onAddressChange(_fileList: unknown, item: { file?: File }) {
  const file = item?.file
  if (!file) return
  if (!['image/jpeg', 'image/png', 'application/pdf'].includes(file.type)) {
    Message.error('Use a JPEG, PNG or PDF file')
    return
  }
  if (file.size > 10 * 1024 * 1024) {
    Message.error('File must be under 10 MB')
    return
  }
  addressFile.value = file
  addressError.value = ''
  if (addressPreview.value) URL.revokeObjectURL(addressPreview.value)
  addressPreview.value = file.type.startsWith('image/') ? URL.createObjectURL(file) : null
}

async function uploadAddress() {
  if (!addressFile.value || !verificationId.value) { Message.warning('Choose an address document'); return }
  addressUploading.value = true
  addressError.value = ''
  try {
    const form = new FormData()
    form.append('document', addressFile.value)
    form.append('document_type', addressDocType.value)
    const { data } = await api.post(`/v2/verify/${verificationId.value}/address-document`, form)
    addressResult.value = data.address_verification || null
    Message.success('Address document processed')
  } catch (error) {
    addressError.value = apiMessage(error)
  } finally {
    addressUploading.value = false
  }
}

// ── Credential ──────────────────────────────────────────────

function decodeJwtPart(part: string) {
  try {
    const normalized = part.replace(/-/g, '+').replace(/_/g, '/')
    const padded = normalized.padEnd(Math.ceil(normalized.length / 4) * 4, '=')
    const binary = atob(padded)
    const bytes = Uint8Array.from(binary, (char) => char.charCodeAt(0))
    return JSON.parse(new TextDecoder().decode(bytes))
  } catch {
    return null
  }
}

async function issueCredential() {
  if (!verificationId.value) {
    credentialError.value = 'Start and complete a verification before issuing a credential.'
    return
  }
  credentialLoading.value = true
  credentialError.value = ''
  credentialStatus.value = null
  try {
    const { data } = await api.get(`/v2/verify/${verificationId.value}/credential`)
    credentialAlreadyIssued.value = !!data.already_issued
    credential.value = { jwt: data.credential || '', jti: data.jti || '', expires_at: data.expires_at || '' }
    if (!data.credential && data.message) Message.info(data.message)
    else Message.success('Credential issued')
  } catch (error) {
    credentialError.value = apiMessage(error)
  } finally {
    credentialLoading.value = false
  }
}

async function checkCredentialStatus() {
  if (!credential.value) return
  checkingCredential.value = true
  credentialError.value = ''
  try {
    const { data } = await api.get(`/v2/credentials/${encodeURIComponent(credential.value.jti)}/status`)
    credentialStatus.value = { active: !!data.active, reason: data.reason }
  } catch (error) {
    credentialError.value = apiMessage(error)
  } finally {
    checkingCredential.value = false
  }
}

async function revokeCredential() {
  if (!credential.value) return
  revoking.value = true
  credentialError.value = ''
  try {
    await api.post(`/v2/credentials/${encodeURIComponent(credential.value.jti)}/revoke`, { reason: 'revoked_from_demo' })
    credentialStatus.value = { active: false, reason: 'revoked' }
    Message.success('Credential revoked')
  } catch (error) {
    credentialError.value = apiMessage(error)
  } finally {
    revoking.value = false
  }
}

async function sendCredential() {
  if (!sendEmail.value.trim()) { Message.warning('Enter an email address'); return }
  sending.value = true
  sendResult.value = ''
  credentialError.value = ''
  try {
    const { data } = await api.post(`/v2/verify/${verificationId.value}/credential/send`, { email: sendEmail.value.trim() })
    sendResult.value = data.email_sent
      ? `Credential emailed to ${sendEmail.value.trim()}`
      : `Queued for ${sendEmail.value.trim()} (SMTP is disabled on this deployment)`
  } catch (error) {
    credentialError.value = apiMessage(error)
  } finally {
    sending.value = false
  }
}

async function copyJwt() {
  if (!credential.value?.jwt) return
  try {
    await navigator.clipboard.writeText(credential.value.jwt)
    copied.value = true
    setTimeout(() => { copied.value = false }, 2000)
  } catch {
    Message.warning('Clipboard access was blocked')
  }
}

// ── Navigation ──────────────────────────────────────────────

function goToCredential() {
  credential.value = null
  credentialError.value = ''
  credentialStatus.value = null
  credentialAlreadyIssued.value = false
  currentStep.value = 10
}

async function retryVerification() {
  if (!verificationId.value) return
  retrying.value = true
  try {
    await api.post(`/v2/verify/${verificationId.value}/restart`)
    retryCount.value += 1
    resetDocumentState()
    verification.value = null
    currentStep.value = 2
    Message.success('Verification restarted')
  } catch (error) {
    Message.error(apiMessage(error))
  } finally {
    retrying.value = false
  }
}

function resetDocumentState() {
  selectedFile.value = null
  documentType.value = ''
  frontError.value = ''
  backFile.value = null
  backError.value = ''
  skipBack.value = false
  checkingError.value = ''
  liveError.value = ''
  cameraError.value = ''
  voiceEnabled.value = false
  voiceAuthEnabled.value = false
  voiceChallenge.value = null
  voiceFile.value = null
  voiceError.value = ''
  addressFile.value = null
  addressResult.value = null
  addressError.value = ''
  stepDetails.ocr = null
  stepDetails.back = null
  stepDetails.crossValidation = null
  stepDetails.documentsMatch = null
  stepDetails.liveness = null
  stepDetails.faceDetection = null
  stepDetails.voice = null
  if (previewUrl.value) URL.revokeObjectURL(previewUrl.value)
  previewUrl.value = null
  if (backPreviewUrl.value) URL.revokeObjectURL(backPreviewUrl.value)
  backPreviewUrl.value = null
  if (addressPreview.value) URL.revokeObjectURL(addressPreview.value)
  addressPreview.value = null
}

function resetDemo() {
  stopCamera()
  resetDocumentState()
  if (voiceCountdown) { clearInterval(voiceCountdown); voiceCountdown = null }
  if (processingTimer) { clearTimeout(processingTimer); processingTimer = null }
  verification.value = null
  verificationId.value = ''
  credential.value = null
  credentialStatus.value = null
  credentialError.value = ''
  retryCount.value = 0
  initError.value = ''
  userId.value = crypto.randomUUID()
  currentStep.value = 1
  router.replace({ path: '/demo' })
}

onMounted(async () => {
  rotationTimers.push(setInterval(() => { processingIndex.value = (processingIndex.value + 1) % PROCESSING_MESSAGES.length }, 1800))
  rotationTimers.push(setInterval(() => { checkingIndex.value = (checkingIndex.value + 1) % CHECKING_MESSAGES.length }, 1800))
  rotationTimers.push(setInterval(() => { cueIndex.value = (cueIndex.value + 1) % LIVENESS_CUES.length }, 1600))

  const queryVerification = typeof route.query.verification_id === 'string' ? route.query.verification_id : ''
  const queryStep = Number(route.query.step)
  if (queryVerification) {
    verificationId.value = queryVerification
    await refreshStatus()
    currentStep.value = 8
  } else if (Number.isFinite(queryStep) && queryStep >= 1 && queryStep <= 10 && queryStep !== 1) {
    currentStep.value = Math.min(Math.max(queryStep, 1), 10)
  }
})

onBeforeUnmount(() => {
  stopCamera()
  rotationTimers.forEach((timer) => clearInterval(timer))
  rotationTimers.length = 0
  if (voiceCountdown) clearInterval(voiceCountdown)
  if (processingTimer) clearTimeout(processingTimer)
})

watch(currentStep, (value) => {
  if (value === 7) voiceAuthEnabled.value = true
})

const processingMessages = PROCESSING_MESSAGES
const processingTags = PROCESSING_TAGS
const checkingMessages = CHECKING_MESSAGES
const checkingTags = CHECKING_TAGS

const ocr = computed<Record<string, any>>(() => stepDetails.ocr || {})
const hasOcr = computed(() => Object.keys(ocr.value).length > 0)

const finalResult = computed(() => {
  const value = verification.value?.final_result
  if (value) return String(value)
  const status = String(verification.value?.status || '')
  if (status === 'HARD_REJECTED') return 'failed'
  if (status === 'COMPLETE') return 'manual_review'
  return ''
})

const resultStatus = computed(() => (finalResult.value === 'verified' ? 'success' : finalResult.value === 'failed' ? 'error' : 'warning'))

const resultTitle = computed(() => {
  if (finalResult.value === 'verified') return 'Identity verified'
  if (finalResult.value === 'failed') return 'Verification failed'
  if (finalResult.value === 'manual_review') return 'Sent for manual review'
  return 'Verification finished'
})

const resultSubtitle = computed(() => {
  if (finalResult.value === 'verified') return 'Every configured gate passed for this session.'
  if (finalResult.value === 'failed') return verification.value?.message || 'A gate rejected this verification.'
  if (finalResult.value === 'manual_review') return 'The result needs a reviewer before it can be accepted.'
  return 'No final result was returned for this session.'
})

const voiceExpired = computed(() => voiceExpiresIn.value !== null && voiceExpiresIn.value <= 0)

const credentialParts = computed(() => {
  const jwt = credential.value?.jwt
  if (!jwt) return null
  const segments = jwt.split('.')
  return {
    header: segments[0] ? decodeJwtPart(segments[0]) : null,
    payload: segments[1] ? decodeJwtPart(segments[1]) : null,
  }
})

const credentialClaims = computed<Record<string, unknown> | null>(() => {
  const claims = credentialParts.value?.payload?.vc?.credentialSubject
  return claims && typeof claims === 'object' ? claims as Record<string, unknown> : null
})

const rawPayload = computed(() => JSON.stringify({
  ...(verification.value || {}),
  step_responses: {
    ocr: stepDetails.ocr,
    back_document: stepDetails.back,
    cross_validation: stepDetails.crossValidation,
    liveness: stepDetails.liveness,
    face_detection_confidence: stepDetails.faceDetection,
    voice: stepDetails.voice,
  },
}, null, 2))

const addressStatusTone = computed(() => {
  const status = String(addressResult.value?.status || '').toUpperCase()
  return status === 'PASS' ? 'success' : status === 'FAIL' ? 'error' : 'warning'
})

const addressStatusLabel = computed(() => {
  const status = String(addressResult.value?.status || '').toUpperCase()
  return status === 'PASS' ? 'Address verified' : status === 'FAIL' ? 'Address rejected' : 'Address needs review'
})

const visibleSteps = computed(() => {
  const keys: Array<{ key: string; description: string }> = [
    { key: 'start', description: 'Start' },
    { key: 'front', description: 'Document' },
  ]
  if (!skipBack.value) keys.push({ key: 'back', description: 'Back' })
  keys.push({ key: 'checking', description: 'Checking' }, { key: 'live', description: 'Selfie' })
  if (voiceAuthEnabled.value) keys.push({ key: 'voice', description: 'Voice' })
  keys.push({ key: 'results', description: 'Result' })
  return keys.map((step, index) => ({ ...step, number: String(index + 1) }))
})

const currentKey = computed(() => {
  switch (currentStep.value) {
    case 1: return 'start'
    case 2: return 'front'
    case 3: return 'front'
    case 4: return skipBack.value ? 'checking' : 'back'
    case 5: return 'checking'
    case 6: return 'live'
    case 7: return 'voice'
    default: return 'results'
  }
})

const currentIndex = computed(() => {
  const index = visibleSteps.value.findIndex((step) => step.key === currentKey.value)
  return index < 0 ? 0 : index
})

const stepCounter = computed(() => {
  const total = visibleSteps.value.length
  const label = visibleSteps.value[currentIndex.value]?.description || ''
  return `Step ${currentIndex.value + 1} of ${total} — ${label}`
})

function clamp01(value: unknown) {
  const number = Number(value)
  if (!Number.isFinite(number)) return 0
  return Math.min(Math.max(number, 0), 1)
}

function percent(value: unknown) {
  const number = Number(value)
  if (!Number.isFinite(number)) return '—'
  return `${Math.round(clamp01(number) * 100)}%`
}

function toneColor(value: unknown) {
  const number = Number(value)
  if (!Number.isFinite(number)) return 'rgb(var(--gray-5))'
  return number >= 0.7 ? 'rgb(var(--green-6))' : number >= 0.4 ? 'rgb(var(--orange-6))' : 'rgb(var(--red-6))'
}

function formatDate(value: unknown) {
  if (!value) return '—'
  const date = new Date(String(value))
  return Number.isNaN(date.getTime()) ? String(value) : date.toLocaleString()
}
</script>

<style scoped>
.demo-shell { background: #f7f6f3; min-height: 100vh; }
.demo-inner { max-width: 760px; margin: 0 auto; padding: 44px 20px 72px; }
.demo-eyebrow { font-family: ui-monospace, Menlo, monospace; font-size: 11px; letter-spacing: .1em; text-transform: uppercase; color: #6b6b70; }
.demo-heading { font-family: ui-monospace, Menlo, monospace; font-size: 26px; font-weight: 600; margin: 10px 0 8px; color: #0b0b0d; }
.demo-lede { color: #6b6b70; font-size: 14px; line-height: 1.65; margin: 0 0 18px; max-width: 640px; }
.demo-privacy { display: flex; gap: 10px; align-items: flex-start; background: #eef6ff; border: 1px solid #e4e2dc; padding: 12px 14px; margin-bottom: 28px; }
.demo-privacy p { margin: 0; font-size: 12px; line-height: 1.6; color: #6b6b70; }
.demo-privacy strong { color: #0b0b0d; }
.demo-privacy-icon { color: #2563eb; font-size: 16px; margin-top: 1px; flex-shrink: 0; }
.demo-card { border-radius: 4px; }
.demo-card :deep(.arco-card-body) { padding: 26px 24px 30px; }
.demo-steps { margin-bottom: 26px; }
.demo-steps :deep(.arco-steps-item-title) { text-transform: uppercase; font-size: 11px; letter-spacing: .05em; }
.step { padding: 4px 0; }
.step.centered { display: flex; flex-direction: column; align-items: center; gap: 10px; padding: 36px 0; text-align: center; }
.step-title { font-size: 20px; font-weight: 600; margin: 0 0 6px; color: #0b0b0d; letter-spacing: -.01em; }
.step.centered .step-title { margin: 0; }
.step-lede { color: #6b6b70; font-size: 13px; line-height: 1.6; margin: 0 0 18px; }
.step-label { font-family: ui-monospace, Menlo, monospace; font-size: 10px; letter-spacing: .12em; text-transform: uppercase; color: #2563eb; margin-bottom: 8px; }
.narrow { max-width: 440px; margin: 0 auto; }
.field { display: flex; flex-direction: column; gap: 6px; }
.field-label { font-size: 12px; color: #4e5969; }
.stack { display: flex; flex-direction: column; gap: 14px; }
.full { width: 100%; }
.grow { flex: 1; min-width: 0; }
.key-row { width: 100%; display: flex; gap: 8px; align-items: center; }
.key-row > :first-child { flex: 1; min-width: 0; }
.send-row { display: flex; gap: 8px; align-items: center; }
.send-row > :first-child { flex: 1; min-width: 0; }
.mode-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 8px; width: 100%; }
.mode-card { text-align: left; padding: 10px 12px; background: transparent; border: 1px solid #e4e2dc; cursor: pointer; transition: border-color .15s, background .15s; display: flex; flex-direction: column; gap: 2px; }
.mode-card strong { font-size: 13px; color: #0b0b0d; font-weight: 600; }
.mode-card span { font-size: 10px; color: #6b6b70; }
.mode-card.active { background: #eef6ff; border-color: #2563eb; }
.mode-card.active strong { color: #1d4ed8; }
.inline-alert { margin-top: 14px; }
.start-actions { width: 100%; margin-top: 20px; }
.device-link { display: block; text-align: center; font-family: ui-monospace, Menlo, monospace; font-size: 12px; color: #6b6b70; text-decoration: none; }
.device-link:hover { color: #0b0b0d; }
.tip-bar { display: flex; align-items: center; gap: 8px; background: #eef6ff; border: 1px solid #e4e2dc; padding: 9px 13px; font-size: 12px; color: #6b6b70; }
.tip-bar i { width: 5px; height: 5px; background: #2563eb; flex-shrink: 0; }
.tip-bar.danger { background: #fff1f0; border-color: #ffccc7; color: #cb2634; }
.tip-bar.danger i { background: #cb2634; }
.id-viewfinder { position: relative; width: 100%; max-width: 420px; margin: 0 auto; aspect-ratio: 1.586; background: #f2f3f5; border: 1px solid #e4e2dc; overflow: hidden; }
.id-preview { position: absolute; inset: 16px; width: calc(100% - 32px); height: calc(100% - 32px); object-fit: cover; }
.id-mock { position: absolute; inset: 16px; background: #fff; padding: 13px 15px; display: flex; flex-direction: column; justify-content: space-between; }
.id-mock-top { display: flex; align-items: center; justify-content: space-between; }
.id-chip { width: 30px; height: 22px; background: linear-gradient(135deg, #c8a84b, #e8c96a); opacity: .75; }
.id-photo { width: 38px; height: 38px; background: #e4e2dc; }
.id-line { display: block; height: 5px; background: #e4e2dc; margin-bottom: 8px; }
.id-line.thin { height: 4px; background: #eef6ff; margin-bottom: 4px; }
.w42 { width: 42%; } .w62 { width: 62%; } .w80 { width: 80%; } .w85 { width: 85%; } .w90 { width: 90%; }
.corner { position: absolute; width: 22px; height: 22px; border: solid #2563eb; }
.corner.tl { top: 10px; left: 10px; border-width: 2px 0 0 2px; }
.corner.tr { top: 10px; right: 10px; border-width: 2px 2px 0 0; }
.corner.bl { bottom: 10px; left: 10px; border-width: 0 0 2px 2px; }
.corner.br { bottom: 10px; right: 10px; border-width: 0 2px 2px 0; }
.viewfinder-overlay { position: absolute; inset: 0; background: rgba(11, 11, 13, .88); display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 12px; color: #fff; }
.viewfinder-overlay small { font-family: ui-monospace, Menlo, monospace; font-size: 10px; letter-spacing: .18em; }
.mrz-lines { display: flex; flex-direction: column; gap: 3px; justify-content: center; flex: 1; }
.mrz-lines span { display: block; height: 4px; background: #eef6ff; }
.dropzone { border: 1px dashed #c9c9cd; background: #fff; padding: 24px 16px; text-align: center; cursor: pointer; width: 100%; }
.dropzone p { margin: 8px 0 2px; font-size: 13px; color: #4e5969; }
.dropzone small { font-size: 11px; color: #86909c; }
.dropzone-icon { font-size: 22px; color: #86909c; }
.file-row { display: flex; align-items: center; justify-content: space-between; gap: 12px; background: #fff; border: 1px solid #e4e2dc; padding: 10px 14px; }
.file-name { margin: 0; font-size: 13px; font-weight: 500; color: #0b0b0d; word-break: break-all; }
.thumb { width: 56px; height: 40px; object-fit: cover; border: 1px solid #e4e2dc; }
.processing-message { font-size: 16px; font-weight: 700; color: #0b0b0d; margin: 8px 0 0; }
.processing-tags { justify-content: center; margin-top: 6px; }
.mono { font-family: ui-monospace, Menlo, monospace; }
.face-viewfinder { position: relative; width: 200px; height: 240px; margin: 0 auto; border-radius: 114px 114px 94px 94px; background: #f2f3f5; overflow: hidden; }
.face-video { width: 100%; height: 100%; object-fit: cover; transform: scaleX(-1); }
.face-placeholder { position: absolute; inset: 0; display: flex; align-items: center; justify-content: center; }
.face-icon { font-size: 46px; color: #c9cdd4; }
.face-dots { position: absolute; inset: 0; background-image: radial-gradient(circle, #dbe6f3 1px, transparent 1px); background-size: 18px 18px; opacity: .7; }
.face-ring { position: absolute; inset: 0; border: 2px solid #2563eb; border-radius: 114px 114px 94px 94px; pointer-events: none; }
.face-viewfinder.live .face-ring { border-color: #00b42a; }
.face-overlay { border-radius: 114px 114px 94px 94px; }
.cue-row { display: flex; justify-content: center; gap: 22px; margin: 4px 0; }
.cue { display: flex; flex-direction: column; align-items: center; gap: 6px; opacity: .55; transition: opacity .3s; }
.cue.active { opacity: 1; }
.cue-icon { width: 40px; height: 40px; display: flex; align-items: center; justify-content: center; font-size: 17px; border: 1px solid #e4e2dc; background: transparent; }
.cue.active .cue-icon { background: #eef6ff; border-color: #2563eb; }
.cue small { font-family: ui-monospace, Menlo, monospace; font-size: 9px; letter-spacing: .06em; color: #86909c; }
.cue.active small { color: #1d4ed8; }
.live-actions { width: 100%; justify-content: center; }
.voice-orb { width: 110px; height: 110px; margin: 0 auto; border-radius: 50%; display: flex; align-items: center; justify-content: center; border: 2px solid #e4e2dc; background: #fff; }
.voice-orb.ready { border-color: #722ed1; box-shadow: 0 0 22px rgba(114, 46, 209, .18); }
.voice-icon { font-size: 38px; color: #6b6b70; }
.voice-orb.ready .voice-icon { color: #722ed1; }
.challenge-box { text-align: center; border: 1px solid #e4e2dc; background: #faf5ff; padding: 14px 18px; display: flex; flex-direction: column; gap: 6px; }
.challenge-box small { font-size: 10px; letter-spacing: .08em; color: #722ed1; }
.challenge-digits { font-family: ui-monospace, Menlo, monospace; font-size: 30px; font-weight: 700; letter-spacing: .2em; color: #0b0b0d; }
.danger { color: #cb2634; }
.result-block { margin-top: 18px; }
.result-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 12px; margin-top: 18px; }
.result-card :deep(.arco-card-body) { padding: 14px 16px; }
.result-card { height: 100%; }
.result-card :deep(.arco-card-header) { padding: 0 16px; height: 40px; }
.result-card :deep(.arco-card-header-title) { font-family: ui-monospace, Menlo, monospace; font-size: 11px; letter-spacing: .04em; text-transform: uppercase; color: #6b6b70; }
.score { margin-bottom: 8px; }
.score-head { display: flex; justify-content: space-between; font-size: 12px; color: #6b6b70; margin-bottom: 4px; }
.score-head strong { font-family: ui-monospace, Menlo, monospace; color: #0b0b0d; }
.kv { display: flex; justify-content: space-between; align-items: center; font-size: 12px; color: #6b6b70; gap: 10px; }
.kv strong { color: #0b0b0d; }
.card-head { display: flex; align-items: center; justify-content: space-between; gap: 12px; }
.demo-code { background: #0f172a; color: #e2e8f0; border: 1px solid #e4e2dc; padding: 12px; font-family: ui-monospace, Menlo, monospace; font-size: 10px; line-height: 1.6; max-height: 320px; overflow: auto; margin: 0; white-space: pre-wrap; word-break: break-all; }
.optional { font-size: 11px; font-weight: 400; color: #86909c; margin-left: 8px; }
.note-list { margin: 0; padding-left: 16px; font-size: 12px; }
.credential-step { display: flex; gap: 12px; align-items: flex-start; }
.credential-step strong { font-size: 13px; color: #0b0b0d; }
.credential-step p { margin: 2px 0 0; font-size: 12px; color: #6b6b70; line-height: 1.5; }
.step-chip { width: 24px; height: 24px; flex-shrink: 0; display: flex; align-items: center; justify-content: center; background: #eef6ff; border: 1px solid #e4e2dc; font-family: ui-monospace, Menlo, monospace; font-size: 10px; color: #2563eb; }
.break { word-break: break-all; }
.ml-6 { margin-left: 6px; }
.address-result :deep(.arco-result-title) { margin-top: 0; }
.address-result :deep(.arco-result) { padding: 12px 0 18px; }
@media (max-width: 720px) {
  .result-grid { grid-template-columns: 1fr; }
  .mode-grid { grid-template-columns: 1fr; }
  .demo-inner { padding: 28px 14px 56px; }
  .demo-card :deep(.arco-card-body) { padding: 18px 14px 22px; }
}
</style>
