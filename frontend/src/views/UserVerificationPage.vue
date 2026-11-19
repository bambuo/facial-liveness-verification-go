<template>
  <div class="uv-page">
    <!-- Session resolving -->
    <div v-if="!sessionReady" class="uv-center">
      <a-spin />
      <p class="uv-hint">Loading verification session…</p>
    </div>

    <!-- Mobile handoff in progress -->
    <div v-else-if="mobileRedirecting" class="uv-center">
      <a-spin />
      <p class="uv-hint">Preparing your mobile session…</p>
    </div>

    <!-- Session token could not be resolved -->
    <div v-else-if="sessionError" class="uv-center">
      <a-result status="error" title="This link is unavailable">
        <template #subtitle>
          {{ sessionError === 'expired'
            ? 'This verification link has expired. Ask the organisation to send a new one.'
            : 'This verification link is invalid or has already been used.' }}
        </template>
        <template #extra>
          <a-button @click="goBack">Go back</a-button>
        </template>
      </a-result>
    </div>

    <!-- Terminal completion screen -->
    <div v-else-if="phase === 'completed'" class="uv-center">
      <div class="uv-completion">
        <div class="uv-logo">
          <img v-if="branding?.logo_url" :src="branding.logo_url" :alt="branding.company_name || 'Logo'" />
          <span v-else>{{ branding?.company_name || 'Identity Verification' }}</span>
        </div>
        <a-result :status="completionStatus" :title="completionTitle">
          <template #subtitle>{{ completionBody }}</template>
        </a-result>
        <a-descriptions
          v-if="completionRows.length"
          :data="completionRows"
          :column="1"
          bordered
          size="small"
          class="uv-descriptions"
        />
        <p class="uv-hint">You can close this window.</p>
        <p v-if="hasCustomBranding" class="uv-hint uv-powered">Powered by Identity Verification</p>
      </div>
    </div>

    <!-- Mobile handoff: continue on phone -->
    <div v-else-if="phase === 'mobile-qr'" class="uv-center">
      <a-button v-if="showBack" class="uv-back" size="small" @click="phase = 'choice'">← Back to options</a-button>
      <a-card class="uv-card uv-handoff">
        <template v-if="handoffState === 'idle'">
          <div class="uv-handoff-head">
            <strong>Continue on your phone</strong>
            <a-tag color="arcoblue" size="small">RECOMMENDED</a-tag>
          </div>
          <p class="uv-hint">
            A guided camera experience on your phone. Link expires in 30 minutes.
          </p>
          <a-button type="primary" long :loading="handoffBusy" @click="createHandoff">
            {{ handoffBusy ? 'Generating link…' : 'Generate mobile link' }}
          </a-button>
          <a-alert v-if="handoffError" type="error" class="uv-gap">{{ handoffError }}</a-alert>
        </template>

        <template v-else-if="handoffState === 'waiting'">
          <p class="uv-handoff-title">Open this link on your phone</p>
          <a-input :model-value="handoffUrl" readonly class="uv-mono">
            <template #append>
              <a-button size="small" @click="copyHandoff">Copy</a-button>
            </template>
          </a-input>
          <a-space class="uv-gap" wrap>
            <a-button type="primary" @click="openHandoff">Open on this device</a-button>
            <a-button size="small" @click="cancelHandoff()">Cancel</a-button>
          </a-space>
          <div class="uv-handoff-status">
            <a-badge status="processing" />
            <span>Waiting for your phone…</span>
            <span class="uv-mono">{{ handoffCountdown }}</span>
          </div>
        </template>

        <template v-else>
          <div class="uv-handoff-done">
            <a-tag :color="handoffResult?.status === 'failed' ? 'red' : handoffResult?.status === 'manual_review' ? 'orange' : 'green'">
              {{ handoffResultLabel }}
            </a-tag>
            <p class="uv-hint">Completed on your phone.</p>
          </div>
        </template>
      </a-card>
    </div>

    <!-- Desktop: EndUserVerification flow -->
    <div v-else-if="phase === 'desktop'" class="uv-desktop">
      <a-button v-if="showBack" class="uv-back" size="small" @click="backToChoice">← Back to options</a-button>
      <a-card class="uv-card">
        <a-steps :current="activeStepIdx" size="small" class="uv-steps">
          <a-step v-for="item in steps" :key="item.step" :title="item.label" />
        </a-steps>

        <!-- 1: Initialize -->
        <div v-if="currentStep === 1" class="uv-step uv-step-center">
          <a-spin :size="36" />
          <h2>Initializing secure session</h2>
          <p class="uv-hint">Establishing an encrypted verification session.</p>
          <a-alert v-if="stepError" type="error" class="uv-gap">{{ stepError }}</a-alert>
          <a-button v-if="stepError" class="uv-gap" @click="backToChoice">Back to options</a-button>
        </div>

        <!-- 2: Front document -->
        <div v-else-if="currentStep === 2" class="uv-step">
          <div class="uv-step-head">
            <h2>{{ isAgeOnly ? 'Upload your ID' : 'Upload the front of your ID' }}</h2>
            <p class="uv-hint">
              {{ isAgeOnly
                ? `We only read your date of birth to confirm you are ${ageThreshold}+.`
                : 'Passport, driving licence or national ID. Keep all four corners visible and avoid glare.' }}
            </p>
          </div>

          <a-form layout="vertical" :model="{}">
            <a-form-item label="Document type">
              <a-select v-model="documentType" :style="{ width: '100%' }">
                <a-option value="national_id">National ID</a-option>
                <a-option value="passport">Passport</a-option>
                <a-option value="drivers_license">Driving licence</a-option>
              </a-select>
            </a-form-item>
          </a-form>

          <a-upload
            class="uv-drop"
            accept="image/*"
            :auto-upload="false"
            :show-file-list="false"
            :limit="1"
            @change="onFrontPicked"
          >
            <template #upload-button>
              <div class="uv-drop-inner">
                <img v-if="frontPreview" :src="frontPreview" alt="Front of ID preview" />
                <template v-else>
                  <div class="uv-drop-plus">+</div>
                  <p class="uv-drop-label">Upload front of ID</p>
                  <p class="uv-hint uv-drop-hint">JPG or PNG, up to 10 MB</p>
                </template>
              </div>
            </template>
          </a-upload>

          <a-button
            v-if="frontFile"
            type="primary"
            long
            class="uv-gap"
            :loading="isLoading"
            @click="uploadFront"
          >
            {{ isLoading ? 'Uploading…' : 'Continue' }}
          </a-button>
          <a-alert v-if="stepError" type="error" class="uv-gap">{{ stepError }}</a-alert>
        </div>

        <!-- 3: Reading front document -->
        <div v-else-if="currentStep === 3" class="uv-step uv-step-center">
          <a-spin :size="36" />
          <h2>{{ skipBack ? 'Verifying your document' : 'Reading your document' }}</h2>
          <p class="uv-hint">
            {{ skipBack
              ? 'Matching the details on your ID against the security features.'
              : 'Extracting the details from the front of your ID.' }}
          </p>
          <p class="uv-hint uv-mono">Please do not close this window.</p>
        </div>

        <!-- 4: Back document -->
        <div v-else-if="currentStep === 4" class="uv-step">
          <div class="uv-step-head">
            <h2>Upload the back of your ID</h2>
            <p class="uv-hint">The back carries the machine-readable zone we cross-check the front against.</p>
          </div>

          <a-descriptions
            v-if="ocrRows.length"
            :data="ocrRows"
            :column="1"
            bordered
            size="small"
            class="uv-gap"
            title="Details read from the front"
          />

          <a-upload
            class="uv-drop"
            accept="image/*"
            :auto-upload="false"
            :show-file-list="false"
            :limit="1"
            @change="onBackPicked"
          >
            <template #upload-button>
              <div class="uv-drop-inner">
                <img v-if="backPreview" :src="backPreview" alt="Back of ID preview" />
                <template v-else>
                  <div class="uv-drop-plus">+</div>
                  <p class="uv-drop-label">Upload back of ID</p>
                  <p class="uv-hint uv-drop-hint">JPG or PNG, up to 10 MB</p>
                </template>
              </div>
            </template>
          </a-upload>

          <a-button
            v-if="backFile"
            type="primary"
            long
            class="uv-gap"
            :loading="isLoading"
            @click="uploadBack"
          >
            {{ isLoading ? 'Uploading…' : 'Continue' }}
          </a-button>
          <a-alert v-if="stepError" type="error" class="uv-gap">{{ stepError }}</a-alert>
        </div>

        <!-- 5: Cross-validation -->
        <div v-else-if="currentStep === 5" class="uv-step uv-step-center">
          <a-spin :size="36" />
          <h2>Cross-checking your document</h2>
          <p class="uv-hint">Comparing the front and back of your ID.</p>
          <a-space direction="vertical" class="uv-checklist">
            <span>— Barcode and MRZ read</span>
            <span>— Personal details compared</span>
            <span>— Document authenticity signals</span>
          </a-space>
        </div>

        <!-- 6: Live capture -->
        <div v-else-if="currentStep === 6" class="uv-step">
          <div class="uv-step-head">
            <h2>Take a live selfie</h2>
            <p class="uv-hint">Follow the on-screen instructions. Your face must stay inside the oval.</p>
          </div>

          <div class="uv-camera">
            <video ref="videoEl" playsinline muted autoplay />
            <canvas ref="overlayEl" width="640" height="480" />
            <div v-if="cameraState === 'prompt'" class="uv-camera-overlay">
              <icon-camera class="uv-camera-icon" />
              <h3>Camera access required</h3>
              <p class="uv-hint">We use your camera only for this verification.</p>
              <a-button type="primary" :loading="cameraBusy" @click="startCamera">Enable camera</a-button>
            </div>
            <div v-else-if="cameraState === 'error'" class="uv-camera-overlay">
              <icon-exclamation-circle class="uv-camera-icon uv-camera-icon-error" />
              <h3>Camera access failed</h3>
              <p class="uv-hint">{{ cameraError }}</p>
              <a-button @click="retryCamera">Try again</a-button>
            </div>
          </div>

          <div v-if="cameraState === 'ready'" class="uv-camera-status">
            <a-tag :color="faceDetected ? 'green' : 'red'">
              {{ faceDetected ? 'Face detected — ready' : 'Position your face in frame' }}
            </a-tag>
            <span class="uv-mono">Liveness {{ Math.round(livenessScore * 100) }}%</span>
            <span class="uv-mono">Stability {{ Math.round(faceStability * 100) }}%</span>
          </div>

          <div v-if="countdown !== null" class="uv-countdown">
            <strong>{{ countdown }}</strong>
            <span class="uv-hint">Hold still…</span>
          </div>

          <a-button
            v-if="cameraState === 'ready' && countdown === null"
            type="primary"
            long
            class="uv-gap"
            :disabled="!faceDetected || livenessScore < 0.4 || faceStability < 0.5 || captureBusy"
            :loading="captureBusy"
            @click="startChallenge"
          >
            {{ captureBusy ? 'Processing…'
              : !faceDetected ? 'Position your face'
              : livenessScore < 0.4 ? 'Improve lighting'
              : faceStability < 0.5 ? 'Hold steady'
              : 'Start capture' }}
          </a-button>
          <a-button v-if="cameraState === 'ready'" long class="uv-gap-sm" @click="retryCamera">
            Restart camera
          </a-button>

          <a-alert v-if="stepError" type="error" class="uv-gap">{{ stepError }}</a-alert>

          <ul class="uv-tips">
            <li>Ensure good lighting on your face</li>
            <li>Keep your face centred in the oval</li>
            <li>Wait for the green indicator before capturing</li>
          </ul>
        </div>

        <!-- 65: Voice challenge -->
        <div v-else-if="currentStep === 65" class="uv-step uv-step-center">
          <h2>Voice verification</h2>
          <p class="uv-hint">Read the digits aloud to add a voice check to this verification.</p>

          <div class="uv-mic" :class="{ 'uv-mic-active': voiceIsRecording }">
            <icon-voice />
          </div>

          <a-card v-if="voiceDigits && (voiceExpiresIn === null || voiceExpiresIn > 0)" class="uv-voice-digits">
            <div class="uv-mono uv-hint">Speak these digits</div>
            <strong>{{ voiceDigits }}</strong>
            <div v-if="voiceExpiresIn !== null" class="uv-hint uv-mono">
              Expires in {{ voiceExpiresIn }}s
            </div>
          </a-card>

          <p v-if="voiceIsRecording" class="uv-hint uv-mono">Recording… {{ voiceSeconds }}s</p>
          <p v-else-if="voiceHasRecording" class="uv-hint uv-mono">Recording captured ({{ voiceSeconds }}s)</p>

          <a-space direction="vertical" class="uv-gap" fill>
            <a-button
              v-if="!voiceDigits"
              type="primary"
              long
              :loading="voiceBusy"
              @click="requestVoiceChallenge"
            >
              Get challenge
            </a-button>
            <a-button
              v-if="voiceDigits && !voiceIsRecording && !voiceHasRecording && (voiceExpiresIn === null || voiceExpiresIn > 0)"
              type="primary"
              long
              @click="startVoiceRecording"
            >
              Start recording
            </a-button>
            <a-button v-if="voiceIsRecording" type="primary" long @click="stopVoiceRecording">
              Stop recording
            </a-button>
            <a-button
              v-if="voiceHasRecording && !voiceIsRecording"
              type="primary"
              long
              :loading="voiceBusy"
              @click="submitVoice"
            >
              Submit voice sample
            </a-button>
            <a-button
              v-if="voiceDigits && voiceExpiresIn !== null && voiceExpiresIn <= 0 && !stepError"
              type="primary"
              long
              @click="restartVoiceChallenge"
            >
              Request new challenge
            </a-button>
          </a-space>

          <a-alert v-if="stepError" type="error" class="uv-gap">{{ stepError }}</a-alert>
          <a-button v-if="stepError" long @click="resetVoice">Try again</a-button>

          <a-upload
            class="uv-gap"
            accept="audio/wav"
            :auto-upload="false"
            :show-file-list="false"
            :limit="1"
            @change="onVoiceFilePicked"
          >
            <a-button long>Upload a WAV recording instead</a-button>
          </a-upload>
        </div>

        <!-- 7: Final result -->
        <div v-else-if="currentStep === 7" class="uv-step uv-step-center">
          <template v-if="!finalResult">
            <a-spin :size="36" />
            <h2>Processing your verification</h2>
            <p class="uv-hint">Matching your selfie to your document and running liveness checks.</p>
          </template>
          <template v-else>
            <a-tag :color="resultTagColor" size="large">{{ resultTagLabel }}</a-tag>
            <h2>{{ resultHeading }}</h2>
            <p class="uv-hint">{{ resultBody }}</p>

            <a-descriptions
              v-if="resultRows.length"
              :data="resultRows"
              :column="1"
              bordered
              size="small"
              class="uv-gap uv-result-grid"
            />

            <a-button
              v-if="finalResult.final_result === 'failed' && finalResult.retry_available === true"
              type="primary"
              long
              class="uv-gap"
              :loading="retryProcessing"
              @click="handleRetry"
            >
              {{ retryProcessing ? 'Restarting…' : 'Try again' }}
            </a-button>
            <p v-if="finalResult.final_result === 'failed' && finalResult.retry_available === false" class="uv-hint uv-mono">
              No retries remaining. Contact the organisation if you need help.
            </p>
          </template>
        </div>
      </a-card>
    </div>

    <!-- Choice (default) -->
    <div v-else class="uv-center">
      <a-button v-if="showBack" class="uv-back" size="small" @click="goBack">← Back</a-button>
      <div class="uv-choice">
        <div class="uv-choice-logo">
          <img v-if="branding?.logo_url" :src="branding.logo_url" :alt="branding.company_name || 'Logo'" />
          <span v-else>{{ branding?.company_name || 'Identity Verification' }}</span>
        </div>

        <a-alert v-if="viewOnly" type="warning" class="uv-gap">
          <template #title>Preview mode</template>
          Start a verification server-side (<code class="uv-mono">POST /api/v2/verify/initialize</code>) and reopen this page
          with <code class="uv-mono">?session=&lt;session_token&gt;</code> to run the live flow.
        </a-alert>

        <div class="uv-choice-head">
          <h1>{{ choiceTitle }}</h1>
          <p class="uv-hint">{{ choiceSubtitle }}</p>
        </div>

        <div class="uv-choice-grid">
          <a-card class="uv-option">
            <a-tag color="green" size="small">RECOMMENDED</a-tag>
            <div class="uv-mono uv-option-kicker">Mobile</div>
            <div>
              <h3>Continue on your phone</h3>
              <ul class="uv-tips">
                <li>Native camera experience</li>
                <li>Better lighting and framing guidance</li>
                <li>Automatic handoff back to this screen</li>
              </ul>
            </div>
            <a-button
              type="primary"
              long
              class="uv-option-action"
              :disabled="viewOnly"
              @click="phase = 'mobile-qr'"
            >
              Scan a QR code
            </a-button>
          </a-card>

          <a-card v-if="!isMobile" class="uv-option">
            <div class="uv-option-spacer" />
            <div class="uv-mono uv-option-kicker">Desktop</div>
            <div>
              <h3>Use this device</h3>
              <p class="uv-hint">Continue here with your webcam. You may be asked to hold your ID up to the camera.</p>
            </div>
            <a-button long class="uv-option-action" :disabled="viewOnly" @click="startDesktop">
              Continue on desktop
            </a-button>
          </a-card>
        </div>

        <p v-if="hasCustomBranding" class="uv-hint uv-powered">Powered by Identity Verification</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Message } from '@arco-design/web-vue'
import type { FileItem } from '@arco-design/web-vue'
import { api, apiMessage } from '../lib/api'
import { useAuthStore } from '../stores/auth'

type Phase = 'choice' | 'mobile-qr' | 'desktop' | 'completed'
type StepDef = { step: number; label: string }

const FULL_STEPS: StepDef[] = [
  { step: 1, label: 'Start' },
  { step: 2, label: 'Front ID' },
  { step: 3, label: 'Scanning' },
  { step: 4, label: 'Back ID' },
  { step: 5, label: 'Checking' },
  { step: 6, label: 'Selfie' },
  { step: 7, label: 'Done' },
]
const DOCUMENT_ONLY_STEPS: StepDef[] = [
  { step: 1, label: 'Start' },
  { step: 2, label: 'Front ID' },
  { step: 3, label: 'Scanning' },
  { step: 4, label: 'Back ID' },
  { step: 5, label: 'Checking' },
  { step: 7, label: 'Done' },
]
const IDENTITY_STEPS: StepDef[] = [
  { step: 1, label: 'Start' },
  { step: 2, label: 'Front ID' },
  { step: 3, label: 'Scanning' },
  { step: 6, label: 'Selfie' },
  { step: 7, label: 'Done' },
]
const AGE_ONLY_STEPS: StepDef[] = [
  { step: 1, label: 'Start' },
  { step: 2, label: 'Upload ID' },
  { step: 7, label: 'Done' },
]

const MAX_POLL_ATTEMPTS = 60

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

const query = (key: string) => {
  const value = route.query[key]
  return typeof value === 'string' ? value : ''
}

const sessionParam = query('session')
const urlApiKey = query('api_key')
const userId = query('user_id')
const redirectUrl = query('redirect_url')
const showBack = query('show_back') !== 'false'
const requestMode = query('verification_mode')
const ageThreshold = Number(query('age_threshold')) || 18
const isAgeOnly = requestMode === 'age_only'
const isDocumentOnly = requestMode === 'document_only'
const isIdentity = requestMode === 'identity'
const slug = typeof route.params.slug === 'string' ? route.params.slug : ''

const phase = ref<Phase>('choice')
const sessionToken = ref(sessionParam)
const sessionVerificationId = ref('')
const sessionUserId = ref('')
const sessionReady = ref(!sessionParam)
const sessionError = ref<'' | 'expired' | 'invalid'>('')
const mobileRedirecting = ref(false)
const isMobile = ref(typeof window !== 'undefined' ? window.innerWidth < 768 : false)
const viewOnly = computed(() => !sessionParam && (!urlApiKey || !userId))

const apiKey = ref(urlApiKey || auth.apiKey)
const branding = ref<{ logo_url?: string | null; accent_color?: string | null; company_name?: string | null } | null>(null)
const pageConfig = ref<Record<string, any> | null>(null)
const hasCustomBranding = computed(() => !!(branding.value?.logo_url || branding.value?.company_name || branding.value?.accent_color))

const currentStep = ref(1)
const verificationId = ref('')
const isLoading = ref(false)
const stepError = ref('')
const documentType = ref('national_id')
const frontFile = ref<File | null>(null)
const frontPreview = ref('')
const backFile = ref<File | null>(null)
const backPreview = ref('')
const ocrData = ref<Record<string, any> | null>(null)
const skipBack = ref(false)
const finalResult = ref<Record<string, any> | null>(null)
const verificationResult = ref<Record<string, any> | null>(null)
const retryProcessing = ref(false)
const liveMetrics = ref<Record<string, any> | null>(null)

const voiceEnabled = ref(false)
const voiceDigits = ref('')
const voiceExpiresIn = ref<number | null>(null)
const voiceSeconds = ref(0)
const voiceIsRecording = ref(false)
const voiceHasRecording = ref(false)
const voiceBusy = ref(false)
let voiceBlob: Blob | null = null
let voiceRecorder: MediaRecorder | null = null
let voiceChunks: Blob[] = []
let voiceTimer: number | null = null
let voiceDurationTimer: number | null = null
let voiceExpiryTimer: number | null = null

const steps = computed<StepDef[]>(() => {
  const base = isAgeOnly ? AGE_ONLY_STEPS
    : (isIdentity || skipBack.value)
      ? (isDocumentOnly
        ? [{ step: 1, label: 'Start' }, { step: 2, label: 'Front ID' }, { step: 3, label: 'Scanning' }, { step: 7, label: 'Done' }]
        : IDENTITY_STEPS)
      : isDocumentOnly ? DOCUMENT_ONLY_STEPS
        : FULL_STEPS
  return voiceEnabled.value
    ? [...base.filter((item) => item.step !== 7), { step: 65, label: 'Voice' }, { step: 7, label: 'Done' }]
    : base
})
const activeStepIdx = computed(() => {
  const index = steps.value.findIndex((item) => item.step === currentStep.value)
  return index >= 0 ? index : steps.value.length - 1
})

const choiceTitle = computed(() => (pageConfig.value?.headerTitle as string)
  || (branding.value?.company_name ? `Verify your identity with ${branding.value.company_name}` : 'Verify your identity'))
const choiceSubtitle = computed(() => (pageConfig.value?.headerSubtitle as string)
  || 'Choose how you would like to complete this verification. Both options take about two minutes.')

const ocrRows = computed(() => {
  const data = ocrData.value
  if (!data) return []
  const rows: { label: string; value: string }[] = []
  if (data.full_name) rows.push({ label: 'Full name', value: String(data.full_name) })
  if (data.id_number) rows.push({ label: 'Document number', value: String(data.id_number) })
  if (data.date_of_birth) rows.push({ label: 'Date of birth', value: String(data.date_of_birth) })
  if (data.expiry_date) rows.push({ label: 'Expiry', value: String(data.expiry_date) })
  if (data.issuing_country) rows.push({ label: 'Issuing country', value: String(data.issuing_country) })
  return rows
})

const resultStatus = computed(() => (finalResult.value?.final_result ?? finalResult.value?.status ?? '') as string)
const resultTagColor = computed(() => (resultStatus.value === 'verified' ? 'green' : resultStatus.value === 'failed' ? 'red' : 'orange'))
const resultTagLabel = computed(() => (resultStatus.value === 'verified' ? 'VERIFIED' : resultStatus.value === 'failed' ? 'FAILED' : 'UNDER REVIEW'))
const resultHeading = computed(() => {
  if (isAgeOnly) {
    return resultStatus.value === 'verified' ? 'Age verified' : 'Age could not be verified'
  }
  if (resultStatus.value === 'verified') return 'Identity verified'
  if (resultStatus.value === 'failed') return 'Verification failed'
  return 'Under review'
})
const resultBody = computed(() => {
  if (isAgeOnly) {
    return resultStatus.value === 'verified'
      ? `You meet the ${ageThreshold}+ age requirement.`
      : (finalResult.value?.user_message || 'We could not confirm your age from this document.')
  }
  if (resultStatus.value === 'verified') return 'Your identity has been successfully verified.'
  if (resultStatus.value === 'failed') return finalResult.value?.user_message || 'We could not verify your identity with these details.'
  return finalResult.value?.user_message || 'A reviewer will check this verification manually.'
})
const resultRows = computed(() => {
  const data = finalResult.value
  if (!data) return []
  const rows: { label: string; value: string }[] = []
  rows.push({ label: 'Status', value: resultStatus.value || 'processing' })
  const liveness = data.liveness_results?.liveness_score ?? data.liveness_score ?? liveMetrics.value?.liveness_score
  if (typeof liveness === 'number') rows.push({ label: 'Liveness', value: `${Math.round(liveness * 100)}%` })
  const passed = data.liveness_results?.liveness_passed ?? liveMetrics.value?.liveness_passed
  if (typeof passed === 'boolean') rows.push({ label: 'Liveness passed', value: passed ? 'Yes' : 'No' })
  const face = data.face_match_results?.similarity_score ?? data.face_match_results?.score ?? data.face_match_score ?? liveMetrics.value?.face_match_score
  if (typeof face === 'number') rows.push({ label: 'Face match', value: `${Math.round(face * 100)}%` })
  const cross = data.cross_validation_results?.score ?? data.cross_validation_results?.overall_score
  if (typeof cross === 'number') rows.push({ label: 'Document cross-check', value: `${Math.round(cross * 100)}%` })
  if (data.rejection_reason) rows.push({ label: 'Reason', value: String(data.rejection_reason) })
  return rows
})

const completionStatus = computed<'success' | 'error' | 'warning'>(() => {
  const status = (verificationResult.value?.status ?? '') as string
  if (status === 'verified' || status === 'completed') return 'success'
  if (status === 'failed') return 'error'
  return 'warning'
})
const completionTitle = computed(() => {
  const status = (verificationResult.value?.status ?? '') as string
  if (status === 'verified' || status === 'completed') return pageConfig.value?.completionTitle || 'Verification complete'
  if (status === 'failed') return 'Verification failed'
  return 'Verification under review'
})
const completionBody = computed(() => {
  const status = (verificationResult.value?.status ?? '') as string
  if (status === 'verified' || status === 'completed') {
    return pageConfig.value?.completionMessage || 'Your identity has been verified successfully.'
  }
  if (status === 'failed') return 'We could not verify your identity. Contact the organisation that requested this check.'
  return 'A reviewer will check this verification manually.'
})
const completionRows = computed(() => {
  const result = verificationResult.value
  if (!result) return []
  const rows: { label: string; value: string }[] = []
  if (typeof result.confidence_score === 'number') rows.push({ label: 'Confidence', value: `${Math.round(result.confidence_score * 100)}%` })
  if (typeof result.face_match_score === 'number') rows.push({ label: 'Face match', value: `${Math.round(result.face_match_score * 100)}%` })
  if (typeof result.liveness_score === 'number') rows.push({ label: 'Liveness', value: `${Math.round(result.liveness_score * 100)}%` })
  return rows
})

/* Mobile handoff */
const handoffState = ref<'idle' | 'waiting' | 'done'>('idle')
const handoffBusy = ref(false)
const handoffError = ref('')
const handoffUrl = ref('')
const handoffToken = ref('')
const handoffExpiresAt = ref(0)
const handoffCountdown = ref('')
const handoffResult = ref<Record<string, any> | null>(null)
let handoffPoll: number | null = null
let handoffTimer: number | null = null

const handoffResultLabel = computed(() => {
  const status = (handoffResult.value?.status ?? '') as string
  if (status === 'failed') return 'Failed'
  if (status === 'manual_review') return 'Under review'
  return 'Verified'
})

function authHeaders(): Record<string, string> {
  return sessionToken.value ? { 'X-Session-Token': sessionToken.value } : {}
}

async function resolveSession() {
  if (!sessionParam) return
  try {
    const { data } = await api.get('/v2/verify/session-info', { headers: { 'X-Session-Token': sessionParam } })
    sessionVerificationId.value = data.verification_id || ''
    sessionUserId.value = data.user_id || ''
    sessionToken.value = sessionParam
    auth.setSessionToken(sessionParam)
  } catch (error) {
    const status = (error as { response?: { status?: number } }).response?.status
    sessionError.value = status === 410 ? 'expired' : 'invalid'
  } finally {
    sessionReady.value = true
  }
}

async function loadBranding() {
  if (sessionParam) return
  try {
    const url = slug
      ? `/v2/verify/page-config/slug/${encodeURIComponent(slug)}`
      : urlApiKey
        ? `/v2/verify/page-config?api_key=${encodeURIComponent(urlApiKey)}`
        : ''
    if (!url) return
    const { data } = await api.get(url)
    if (data?.branding) branding.value = data.branding
    const config = data?.config ?? data?.page_config
    if (config) pageConfig.value = config
  } catch {
    /* branding is optional */
  }
}

async function autoRedirectMobile() {
  if (!isMobile.value || viewOnly.value || !sessionReady.value || sessionError.value || mobileRedirecting.value) return
  mobileRedirecting.value = true
  try {
    const { data } = await api.post('/verify/handoff/create', {
      ...(sessionToken.value ? {} : { api_key: apiKey.value }),
      user_id: sessionUserId.value || userId,
      ...(sessionVerificationId.value ? { verification_id: sessionVerificationId.value } : {}),
    }, { headers: { 'Content-Type': 'application/json', ...authHeaders() } })
    let url = `/verify/mobile?token=${data.token}`
    if (redirectUrl) url += `&redirect_url=${encodeURIComponent(redirectUrl)}`
    router.replace(url)
  } catch {
    mobileRedirecting.value = false
    phase.value = 'desktop'
    startDesktop()
  }
}

function goBack() {
  if (window.history.length > 1) router.back()
  else window.close()
}

function backToChoice() {
  phase.value = 'choice'
  stepError.value = ''
}

onMounted(async () => {
  await Promise.all([resolveSession(), loadBranding()])
  await autoRedirectMobile()
  window.addEventListener('resize', onResize)
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', onResize)
  stopCamera()
  cleanupVoice()
  cancelHandoff()
})

function onResize() {
  isMobile.value = window.innerWidth < 768
}

/* ── Desktop flow: initialize → front → back → selfie → finalize ─────────── */

function sleep(ms: number) {
  return new Promise((resolve) => window.setTimeout(resolve, ms))
}

async function fetchStatus(): Promise<Record<string, any> | null> {
  if (!verificationId.value) return null
  try {
    const { data } = await api.get(`/v2/verify/${verificationId.value}/status`, { headers: authHeaders() })
    return data
  } catch {
    return null
  }
}

async function startDesktop() {
  phase.value = 'desktop'
  currentStep.value = 1
  stepError.value = ''
  if (sessionVerificationId.value) {
    verificationId.value = sessionVerificationId.value
    const status = await fetchStatus()
    if (status?.final_result === 'failed') {
      showResult(status)
      return
    }
    currentStep.value = 2
    return
  }
  if (!apiKey.value || !userId) {
    stepError.value = 'Missing parameters. Open this page with ?session=<token>, or with ?api_key=…&user_id=…'
    return
  }
  isLoading.value = true
  try {
    auth.setAPIKey(apiKey.value)
    const { data } = await api.post('/v2/verify/initialize', { user_id: userId }, { headers: authHeaders() })
    verificationId.value = data.verification_id
    if (data.session_token) {
      sessionToken.value = data.session_token
      auth.setSessionToken(data.session_token)
    }
    currentStep.value = 2
  } catch (error) {
    stepError.value = apiMessage(error)
  } finally {
    isLoading.value = false
  }
}

function onFrontPicked(_fileList: FileItem[], item: FileItem) {
  if (!item?.file) return
  if (frontPreview.value) URL.revokeObjectURL(frontPreview.value)
  frontFile.value = item.file
  frontPreview.value = URL.createObjectURL(item.file)
  stepError.value = ''
}

function onBackPicked(_fileList: FileItem[], item: FileItem) {
  if (!item?.file) return
  if (backPreview.value) URL.revokeObjectURL(backPreview.value)
  backFile.value = item.file
  backPreview.value = URL.createObjectURL(item.file)
  stepError.value = ''
}

async function uploadFront() {
  if (!frontFile.value || !verificationId.value) return
  isLoading.value = true
  stepError.value = ''
  try {
    const form = new FormData()
    form.append('document_type', documentType.value)
    form.append('document', frontFile.value)
    const { data } = await api.post(`/v2/verify/${verificationId.value}/front-document`, form, { headers: authHeaders() })
    ocrData.value = data.ocr_data ?? null
    Message.success('Document uploaded')
    if (isAgeOnly && data.age_verification) {
      showResult(data)
      return
    }
    skipBack.value = isIdentity || data.requires_back === false || data.detected_document_type === 'passport'
    currentStep.value = 3
    await pollFrontOCR()
  } catch (error) {
    stepError.value = apiMessage(error)
  } finally {
    isLoading.value = false
  }
}

async function pollFrontOCR() {
  for (let attempt = 0; attempt < MAX_POLL_ATTEMPTS; attempt++) {
    await sleep(2000)
    const status = await fetchStatus()
    if (!status) continue
    if (status.final_result === 'failed' || status.final_result === 'manual_review') {
      showResult(status)
      return
    }
    if (status.status !== 'FRONT_PROCESSING') {
      currentStep.value = skipBack.value ? 6 : 4
      return
    }
  }
  stepError.value = 'Document processing timed out. Please try again.'
  currentStep.value = 2
}

async function uploadBack() {
  if (!backFile.value || !verificationId.value) return
  isLoading.value = true
  stepError.value = ''
  try {
    const form = new FormData()
    form.append('document', backFile.value)
    form.append('document_type', documentType.value)
    await api.post(`/v2/verify/${verificationId.value}/back-document`, form, { headers: authHeaders() })
    Message.success('Back of ID uploaded')
    currentStep.value = 5
    await api.post(`/v2/verify/${verificationId.value}/cross-validation`, {}, { headers: authHeaders() })
    await pollCrossValidation()
  } catch (error) {
    stepError.value = apiMessage(error)
  } finally {
    isLoading.value = false
  }
}

async function pollCrossValidation() {
  for (let attempt = 0; attempt < MAX_POLL_ATTEMPTS; attempt++) {
    await sleep(2000)
    const status = await fetchStatus()
    if (!status) continue
    if (status.final_result) {
      showResult(status)
      return
    }
    if (status.status === 'AWAITING_LIVE' || status.status === 'FACE_MATCHING') {
      currentStep.value = 6
      return
    }
  }
  stepError.value = 'Cross-check is taking too long. Please try again.'
  currentStep.value = 4
}

function notifyParent(result: Record<string, any>) {
  if (window.parent !== window) {
    window.parent.postMessage({ type: 'VERIFICATION_COMPLETE', result }, '*')
  }
}

function showResult(data: Record<string, any>) {
  const result = { ...data }
  finalResult.value = result
  currentStep.value = 7
  if (result.final_result === 'failed' && result.retry_available === true) return
  const status = result.final_result ?? result.status
  const summary = {
    verification_id: result.verification_id ?? verificationId.value,
    status,
    user_id: result.user_id ?? sessionUserId.value ?? userId,
    confidence_score: result.confidence_score,
    face_match_score: result.face_match_results?.similarity_score ?? result.face_match_score ?? liveMetrics.value?.face_match_score,
    liveness_score: result.liveness_results?.liveness_score ?? result.liveness_score ?? liveMetrics.value?.liveness_score,
  }
  notifyParent(summary)
  if (redirectUrl) {
    window.setTimeout(() => {
      const target = new URL(redirectUrl)
      target.searchParams.set('verification_id', String(summary.verification_id ?? ''))
      target.searchParams.set('status', String(status ?? ''))
      window.location.href = target.toString()
    }, 3000)
    return
  }
  verificationResult.value = summary
  phase.value = 'completed'
}

async function pollFinalResult() {
  for (let attempt = 0; attempt < MAX_POLL_ATTEMPTS; attempt++) {
    await sleep(3000)
    const status = await fetchStatus()
    if (!status) continue
    if (status.status === 'AWAITING_VOICE') {
      voiceEnabled.value = true
      currentStep.value = 65
      return
    }
    if (status.final_result) {
      showResult(status)
      return
    }
  }
  stepError.value = 'Verification is taking longer than expected. Refresh to check the status.'
}

async function finalize() {
  currentStep.value = 7
  finalResult.value = null
  try {
    const { data } = await api.post(`/v2/verify/${verificationId.value}/finalize`, {}, { headers: authHeaders() })
    if (data?.final_result) {
      showResult(data)
      return
    }
  } catch {
    /* finalize may lag behind the live-capture write — fall back to polling */
  }
  await pollFinalResult()
}

async function handleRetry() {
  if (!verificationId.value) return
  retryProcessing.value = true
  try {
    await api.post(`/v2/verify/${verificationId.value}/restart`, {}, { headers: authHeaders() })
    if (frontPreview.value) URL.revokeObjectURL(frontPreview.value)
    if (backPreview.value) URL.revokeObjectURL(backPreview.value)
    frontFile.value = null
    frontPreview.value = ''
    backFile.value = null
    backPreview.value = ''
    ocrData.value = null
    finalResult.value = null
    skipBack.value = false
    voiceEnabled.value = false
    resetVoice()
    currentStep.value = 2
  } catch (error) {
    Message.error(apiMessage(error))
  } finally {
    retryProcessing.value = false
  }
}

/* ── Live capture (camera + liveness overlay) ────────────────────────────── */

const videoEl = ref<HTMLVideoElement | null>(null)
const overlayEl = ref<HTMLCanvasElement | null>(null)
const cameraState = ref<'prompt' | 'initializing' | 'ready' | 'error'>('prompt')
const cameraBusy = ref(false)
const captureBusy = ref(false)
const cameraError = ref('')
const faceDetected = ref(false)
const livenessScore = ref(0)
const faceStability = ref(0)
const countdown = ref<number | null>(null)

let mediaStream: MediaStream | null = null
let rafId: number | null = null
let analysisCanvas: HTMLCanvasElement | null = null
const faceHistory: boolean[] = []
const stabHistory: number[] = []

async function startCamera() {
  cameraState.value = 'initializing'
  cameraError.value = ''
  cameraBusy.value = true
  try {
    if (!navigator.mediaDevices?.getUserMedia) throw new Error('Camera not supported in this browser')
    mediaStream = await navigator.mediaDevices.getUserMedia({
      video: { width: { ideal: 640 }, height: { ideal: 480 }, facingMode: 'user' },
      audio: false,
    })
    const video = videoEl.value
    if (video) {
      video.srcObject = mediaStream
      await video.play()
    }
    cameraState.value = 'ready'
    rafId = window.requestAnimationFrame(cameraLoop)
  } catch (error) {
    cameraState.value = 'error'
    const err = error as DOMException
    cameraError.value =
      err.name === 'NotAllowedError' ? 'Camera permission denied. Please enable camera access.'
        : err.name === 'NotFoundError' ? 'No camera found.'
          : err.name === 'NotReadableError' ? 'Camera is already in use by another app.'
            : `Camera error: ${err.message || 'Unknown error'}`
  } finally {
    cameraBusy.value = false
  }
}

function stopCamera() {
  if (rafId !== null) {
    window.cancelAnimationFrame(rafId)
    rafId = null
  }
  mediaStream?.getTracks().forEach((track) => track.stop())
  mediaStream = null
  const video = videoEl.value
  if (video) video.srcObject = null
  cameraState.value = 'prompt'
  countdown.value = null
}

function cameraLoop() {
  const video = videoEl.value
  const overlay = overlayEl.value
  if (cameraState.value !== 'ready') return
  if (!video || !overlay) {
    rafId = window.requestAnimationFrame(cameraLoop)
    return
  }
  if (video.readyState >= 2 && video.videoWidth > 0) {
    if (!analysisCanvas) analysisCanvas = document.createElement('canvas')
    analysisCanvas.width = 160
    analysisCanvas.height = 120
    const ctx = analysisCanvas.getContext('2d', { willReadFrequently: true })
    if (ctx) {
      ctx.drawImage(video, 0, 0, 160, 120)
      analyzeFrame(ctx.getImageData(0, 0, 160, 120))
      drawOverlay(overlay)
    }
  }
  rafId = window.requestAnimationFrame(cameraLoop)
}

function analyzeFrame(image: ImageData) {
  const { data, width, height } = image
  const cx = width / 2
  const cy = height / 2
  const size = Math.min(width, height) * 0.5
  let skin = 0
  let warm = 0
  let total = 0
  let brightness = 0
  let variance = 0
  for (let y = cy - size / 2; y < cy + size / 2; y++) {
    for (let x = cx - size / 2; x < cx + size / 2; x++) {
      if (x < 0 || x >= width || y < 0 || y >= height) continue
      const i = (Math.floor(y) * width + Math.floor(x)) * 4
      const r = data[i]
      const g = data[i + 1]
      const b = data[i + 2]
      const br = (r + g + b) / 3
      brightness += br
      variance += Math.abs(r - g) + Math.abs(g - b) + Math.abs(r - b)
      if ((r > 50 && g > 25 && b > 15 && r > b) || (br > 40 && br < 250 && r >= g) || (r > 30 && r + g > b * 1.5)) skin++
      if (r > g && r > b && br > 50) warm++
      total++
    }
  }
  const skinRatio = total ? skin / total : 0
  const warmRatio = total ? warm / total : 0
  const avgBrightness = total ? brightness / total : 0
  const avgVariance = total ? variance / total : 0
  livenessScore.value = Math.max(
    0.5,
    Math.min(1, skinRatio * 3) * 0.6
      + (avgBrightness > 30 && avgBrightness < 250 ? 1 : 0.7) * 0.2
      + Math.min(1, avgVariance / 20) * 0.2,
  )
  stabHistory.push(skinRatio > 0.02 || warmRatio > 0.05 ? 1 : 0)
  if (stabHistory.length > 5) stabHistory.shift()
  faceStability.value = Math.max(0.5, stabHistory.reduce((a, b) => a + b, 0) / Math.max(stabHistory.length, 1))

  const detected =
    (avgBrightness > 30 && avgBrightness < 250 && (skinRatio > 0.02 || warmRatio > 0.05))
    || (livenessScore.value > 0.3 && avgBrightness > 40 && avgVariance > 5)
  faceHistory.push(detected)
  if (faceHistory.length > 6) faceHistory.shift()
  const positive = faceHistory.filter(Boolean).length
  faceDetected.value = faceHistory.length >= 3 && positive >= Math.ceil(faceHistory.length * 0.5)
}

function drawOverlay(canvas: HTMLCanvasElement) {
  const ctx = canvas.getContext('2d')
  if (!ctx) return
  const { width, height } = canvas
  const cx = width / 2
  const cy = height / 2
  const radiusX = Math.min(width, height) * 0.22
  const radiusY = radiusX * 1.35
  ctx.clearRect(0, 0, width, height)
  ctx.strokeStyle = faceDetected.value ? '#10b981' : '#ef4444'
  ctx.lineWidth = 2
  ctx.setLineDash([8, 6])
  ctx.beginPath()
  ctx.ellipse(cx, cy, radiusX, radiusY, 0, 0, 2 * Math.PI)
  ctx.stroke()
  ctx.setLineDash([])
  ctx.fillStyle = ctx.strokeStyle
  ctx.font = '14px monospace'
  ctx.textAlign = 'center'
  ctx.fillText(faceDetected.value ? 'FACE DETECTED' : 'POSITION FACE IN FRAME', cx, cy + radiusY + 24)
  if (faceDetected.value) {
    ctx.font = '12px monospace'
    ctx.fillText(
      `LIVENESS ${Math.round(livenessScore.value * 100)}%  |  STABILITY ${Math.round(faceStability.value * 100)}%`,
      cx,
      cy + radiusY + 44,
    )
  }
}

function retryCamera() {
  stopCamera()
  stepError.value = ''
  faceHistory.length = 0
  stabHistory.length = 0
  void startCamera()
}

function startChallenge() {
  if (!faceDetected.value || livenessScore.value < 0.4 || faceStability.value < 0.5) {
    stepError.value = !faceDetected.value
      ? 'No face detected. Position your face in the frame.'
      : livenessScore.value < 0.4
        ? 'Ensure good lighting and your face is clearly visible.'
        : 'Hold your face steady in the frame.'
    return
  }
  stepError.value = ''
  countdown.value = 3
  const timer = window.setInterval(() => {
    if (countdown.value === null) {
      window.clearInterval(timer)
      return
    }
    if (!faceDetected.value || livenessScore.value < 0.4 || faceStability.value < 0.5) {
      window.clearInterval(timer)
      countdown.value = null
      stepError.value = 'Face lost during countdown. Please try again.'
      return
    }
    if (countdown.value <= 1) {
      window.clearInterval(timer)
      countdown.value = null
      void captureFrame()
    } else {
      countdown.value -= 1
    }
  }, 1000)
}

async function captureFrame() {
  const video = videoEl.value
  if (!video) return
  const canvas = document.createElement('canvas')
  canvas.width = video.videoWidth || 640
  canvas.height = video.videoHeight || 480
  canvas.getContext('2d')?.drawImage(video, 0, 0, canvas.width, canvas.height)
  const blob = await new Promise<Blob | null>((resolve) => canvas.toBlob(resolve, 'image/jpeg', 0.8))
  if (!blob) {
    stepError.value = 'Failed to capture image. Please try again.'
    return
  }
  const ok = await submitSelfie(blob)
  if (ok) stopCamera()
}

async function submitSelfie(blob: Blob): Promise<boolean> {
  if (!verificationId.value) return false
  captureBusy.value = true
  stepError.value = ''
  try {
    const form = new FormData()
    form.append('selfie', blob, 'selfie.jpg')
    const { data } = await api.post(`/v2/verify/${verificationId.value}/live-capture`, form, { headers: authHeaders() })
    liveMetrics.value = {
      liveness_score: data.liveness_results?.liveness_score,
      liveness_passed: data.liveness_results?.liveness_passed,
      face_match_score: data.face_match_results?.confidence,
    }
    const status = await fetchStatus()
    if (status?.status === 'AWAITING_VOICE') {
      voiceEnabled.value = true
      currentStep.value = 65
      return true
    }
    await finalize()
    return true
  } catch (error) {
    stepError.value = apiMessage(error)
    return false
  } finally {
    captureBusy.value = false
  }
}

/* ── Voice challenge ─────────────────────────────────────────────────────── */

function resetVoice() {
  if (voiceRecorder && voiceRecorder.state === 'recording') {
    voiceRecorder.stream.getTracks().forEach((track) => track.stop())
    voiceRecorder.stop()
  }
  voiceRecorder = null
  if (voiceTimer !== null) window.clearTimeout(voiceTimer)
  if (voiceDurationTimer !== null) window.clearInterval(voiceDurationTimer)
  if (voiceExpiryTimer !== null) window.clearInterval(voiceExpiryTimer)
  voiceTimer = null
  voiceDurationTimer = null
  voiceExpiryTimer = null
  voiceDigits.value = ''
  voiceExpiresIn.value = null
  voiceHasRecording.value = false
  voiceIsRecording.value = false
  voiceSeconds.value = 0
  voiceBlob = null
  voiceChunks = []
  stepError.value = ''
}

function cleanupVoice() {
  resetVoice()
  voiceEnabled.value = false
}

async function requestVoiceChallenge() {
  if (!verificationId.value) return
  voiceBusy.value = true
  stepError.value = ''
  try {
    const { data } = await api.post(`/v2/verify/${verificationId.value}/voice-challenge`, {}, { headers: authHeaders() })
    voiceDigits.value = data.challenge_digits
    voiceExpiresIn.value = data.expires_in_seconds
    if (voiceExpiryTimer !== null) window.clearInterval(voiceExpiryTimer)
    const start = Date.now()
    const total = Number(data.expires_in_seconds) || 0
    voiceExpiryTimer = window.setInterval(() => {
      const remaining = total - Math.floor((Date.now() - start) / 1000)
      voiceExpiresIn.value = Math.max(0, remaining)
      if (remaining <= 0 && voiceExpiryTimer !== null) {
        window.clearInterval(voiceExpiryTimer)
        voiceExpiryTimer = null
      }
    }, 1000)
  } catch (error) {
    stepError.value = apiMessage(error)
  } finally {
    voiceBusy.value = false
  }
}

async function startVoiceRecording() {
  stepError.value = ''
  voiceChunks = []
  voiceBlob = null
  voiceHasRecording.value = false
  try {
    const stream = await navigator.mediaDevices.getUserMedia({ audio: true })
    const recorder = new MediaRecorder(stream, {
      mimeType: MediaRecorder.isTypeSupported('audio/webm;codecs=opus') ? 'audio/webm;codecs=opus' : 'audio/webm',
    })
    recorder.ondataavailable = (event) => {
      if (event.data.size > 0) voiceChunks.push(event.data)
    }
    recorder.onstop = () => {
      const raw = new Blob(voiceChunks, { type: recorder.mimeType })
      void blobToWav(raw).then((wav) => {
        voiceBlob = wav
        voiceHasRecording.value = true
      }).catch(() => {
        voiceBlob = raw
        voiceHasRecording.value = true
      })
      voiceIsRecording.value = false
      if (voiceDurationTimer !== null) window.clearInterval(voiceDurationTimer)
      voiceDurationTimer = null
      stream.getTracks().forEach((track) => track.stop())
    }
    voiceRecorder = recorder
    recorder.start(100)
    voiceIsRecording.value = true
    voiceSeconds.value = 0
    const startedAt = Date.now()
    voiceDurationTimer = window.setInterval(() => {
      voiceSeconds.value = Math.floor((Date.now() - startedAt) / 1000)
    }, 200)
    voiceTimer = window.setTimeout(() => {
      if (voiceRecorder?.state === 'recording') voiceRecorder.stop()
    }, 10000)
  } catch (error) {
    stepError.value = (error as Error).message || 'Microphone access denied.'
  }
}

function stopVoiceRecording() {
  if (voiceTimer !== null) window.clearTimeout(voiceTimer)
  if (voiceRecorder?.state === 'recording') voiceRecorder.stop()
}

function onVoiceFilePicked(_fileList: FileItem[], item: FileItem) {
  if (!item?.file) return
  voiceBlob = item.file
  voiceHasRecording.value = true
  voiceIsRecording.value = false
  stepError.value = ''
}

async function submitVoice() {
  if (!verificationId.value || !voiceBlob) return
  voiceBusy.value = true
  stepError.value = ''
  try {
    const form = new FormData()
    form.append('file', voiceBlob, 'voice.wav')
    const { data } = await api.post(`/v2/verify/${verificationId.value}/voice-capture`, form, { headers: authHeaders() })
    if (voiceExpiryTimer !== null) window.clearInterval(voiceExpiryTimer)
    if (data?.final_result) showResult(data)
    else await finalize()
  } catch (error) {
    stepError.value = apiMessage(error)
  } finally {
    voiceBusy.value = false
  }
}

async function restartVoiceChallenge() {
  resetVoice()
  await requestVoiceChallenge()
}

async function blobToWav(blob: Blob): Promise<Blob> {
  const buffer = await blob.arrayBuffer()
  const AudioCtx: typeof AudioContext = window.AudioContext
    || (window as unknown as { webkitAudioContext: typeof AudioContext }).webkitAudioContext
  const context = new AudioCtx()
  const decoded = await context.decodeAudioData(buffer)
  await context.close()
  const channels = decoded.numberOfChannels
  const mono = new Float32Array(decoded.length)
  for (let channel = 0; channel < channels; channel++) {
    const samples = decoded.getChannelData(channel)
    for (let i = 0; i < samples.length; i++) mono[i] += samples[i] / channels
  }
  const out = new ArrayBuffer(44 + mono.length * 2)
  const view = new DataView(out)
  const writeString = (offset: number, value: string) => {
    for (let i = 0; i < value.length; i++) view.setUint8(offset + i, value.charCodeAt(i))
  }
  writeString(0, 'RIFF')
  view.setUint32(4, 36 + mono.length * 2, true)
  writeString(8, 'WAVE')
  writeString(12, 'fmt ')
  view.setUint32(16, 16, true)
  view.setUint16(20, 1, true)
  view.setUint16(22, 1, true)
  view.setUint32(24, decoded.sampleRate, true)
  view.setUint32(28, decoded.sampleRate * 2, true)
  view.setUint16(32, 2, true)
  view.setUint16(34, 16, true)
  writeString(36, 'data')
  view.setUint32(40, mono.length * 2, true)
  let offset = 44
  for (let i = 0; i < mono.length; i++) {
    const sample = Math.max(-1, Math.min(1, mono[i]))
    view.setInt16(offset, sample < 0 ? sample * 0x8000 : sample * 0x7fff, true)
    offset += 2
  }
  return new Blob([out], { type: 'audio/wav' })
}

/* ── Mobile handoff (continue on phone) ──────────────────────────────────── */

async function createHandoff() {
  handoffBusy.value = true
  handoffError.value = ''
  try {
    const { data } = await api.post('/verify/handoff/create', {
      ...(sessionToken.value ? {} : { api_key: apiKey.value }),
      user_id: sessionUserId.value || userId,
      ...(sessionVerificationId.value ? { verification_id: sessionVerificationId.value } : {}),
    }, { headers: { 'Content-Type': 'application/json', ...authHeaders() } })
    if (!data?.token || !data?.expires_at) throw new Error('Handoff response missing required fields')
    handoffToken.value = data.token
    handoffUrl.value = `${window.location.origin}/verify/mobile?token=${data.token}${redirectUrl ? `&redirect_url=${encodeURIComponent(redirectUrl)}` : ''}`
    handoffExpiresAt.value = new Date(data.expires_at).getTime()
    handoffState.value = 'waiting'
    const initialLeft = Math.max(0, Math.floor((handoffExpiresAt.value - Date.now()) / 1000))
    handoffCountdown.value = `${Math.floor(initialLeft / 60)}:${String(initialLeft % 60).padStart(2, '0')}`
    startHandoffTimers()
  } catch (error) {
    handoffError.value = apiMessage(error)
  } finally {
    handoffBusy.value = false
  }
}

function startHandoffTimers() {
  handoffTimer = window.setInterval(() => {
    const left = Math.max(0, Math.floor((handoffExpiresAt.value - Date.now()) / 1000))
    handoffCountdown.value = `${Math.floor(left / 60)}:${String(left % 60).padStart(2, '0')}`
    if (left <= 0) cancelHandoff()
  }, 1000)
  handoffPoll = window.setInterval(async () => {
    try {
      const { data } = await api.get(`/verify/handoff/${handoffToken.value}/status`)
      if (data.status === 'pending') return
      handoffResult.value = data.result ?? { status: data.status }
      handoffState.value = 'done'
      cancelHandoff(true)
      handleVerificationComplete(handoffResult.value as Record<string, any>)
    } catch {
      /* retry on the next tick */
    }
  }, 3000)
}

async function copyHandoff() {
  try {
    await navigator.clipboard.writeText(handoffUrl.value)
    Message.success('Link copied')
  } catch {
    Message.warning('Copy the link manually')
  }
}

function openHandoff() {
  window.location.href = handoffUrl.value
}

function cancelHandoff(keepState = false) {
  if (handoffTimer !== null) window.clearInterval(handoffTimer)
  if (handoffPoll !== null) window.clearInterval(handoffPoll)
  handoffTimer = null
  handoffPoll = null
  if (!keepState) {
    handoffState.value = 'idle'
    handoffUrl.value = ''
    handoffToken.value = ''
    handoffResult.value = null
    handoffCountdown.value = ''
  }
}

function handleVerificationComplete(result: Record<string, any>) {
  const summary = {
    verification_id: result.verification_id ?? 'mobile-handoff',
    status: result.status ?? 'manual_review',
    user_id: result.user_id ?? userId,
    confidence_score: result.confidence_score,
    face_match_score: result.face_match_score,
    liveness_score: result.liveness_score,
  }
  notifyParent(summary)
  if (redirectUrl) {
    window.setTimeout(() => {
      const target = new URL(redirectUrl)
      target.searchParams.set('verification_id', String(summary.verification_id))
      target.searchParams.set('status', String(summary.status))
      window.location.href = target.toString()
    }, 1500)
    return
  }
  verificationResult.value = summary
  phase.value = 'completed'
}
</script>

<style scoped>
.uv-page { min-height: 100vh; background: #f7f6f3; color: #0b0b0d; }
.uv-center { min-height: 100vh; display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 14px; padding: 32px 16px; position: relative; }
.uv-desktop { min-height: 100vh; display: flex; align-items: center; justify-content: center; padding: 48px 16px; position: relative; }
.uv-card { width: min(540px, 100%); border-radius: 4px; }
.uv-back { position: absolute; top: 20px; left: 20px; font-family: ui-monospace, Menlo, monospace; }
.uv-hint { color: #6b6b70; font-size: 13px; line-height: 1.55; margin: 0; }
.uv-mono { font-family: ui-monospace, SFMono-Regular, Menlo, monospace; }
.uv-gap { margin-top: 14px; }
.uv-gap-sm { margin-top: 8px; }
.uv-power, .uv-powered { font-size: 11px; letter-spacing: .04em; }
.uv-logo, .uv-choice-logo { font-family: ui-monospace, Menlo, monospace; font-size: 13px; letter-spacing: .12em; text-transform: uppercase; color: #6b6b70; text-align: center; margin-bottom: 18px; }
.uv-logo img, .uv-choice-logo img { height: 36px; max-width: 60%; object-fit: contain; display: inline-block; }

.uv-completion { width: min(440px, 100%); text-align: center; }
.uv-descriptions { text-align: left; margin-top: 4px; }

.uv-choice { width: min(680px, 100%); }
.uv-choice-head { text-align: center; margin: 12px 0 26px; }
.uv-choice-head h1 { font-size: 26px; font-weight: 600; margin: 0 0 8px; }
.uv-choice-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 16px; align-items: stretch; }
.uv-option { display: flex; flex-direction: column; gap: 14px; border-radius: 4px; }
.uv-option h3 { margin: 0 0 6px; font-size: 16px; }
.uv-option-kicker { font-size: 12px; letter-spacing: .08em; text-transform: uppercase; color: #6b6b70; }
.uv-option-spacer { height: 22px; }
.uv-option-action { margin-top: auto; }
@media (max-width: 767px) { .uv-choice-grid { grid-template-columns: 1fr; } }

.uv-steps { margin-bottom: 26px; }
.uv-steps :deep(.arco-steps-item) { flex: 1 1 0; min-width: 0; padding-right: 6px; }
.uv-steps :deep(.arco-steps-item:last-child) { padding-right: 0; }
.uv-steps :deep(.arco-steps-item-node) { display: none; }
.uv-steps :deep(.arco-steps-item-title) { font-size: 11px; line-height: 1.2; letter-spacing: 0; color: #6b6b70; }
.uv-steps :deep(.arco-steps-item-active .arco-steps-item-title) { color: #0b0b0d; font-weight: 600; }
.uv-steps :deep(.arco-steps-item-title) { font-size: 11px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.uv-step { display: flex; flex-direction: column; gap: 14px; }
.uv-step-center { align-items: center; text-align: center; padding: 18px 0; }
.uv-step h2 { font-size: 20px; font-weight: 600; margin: 0; }
.uv-step-head h2 { margin-bottom: 6px; }
.uv-checklist { display: flex; flex-direction: column; gap: 6px; font-family: ui-monospace, Menlo, monospace; font-size: 12px; color: #6b6b70; margin-top: 12px; }
.uv-tips { margin: 4px 0 0; padding-left: 18px; color: #6b6b70; font-size: 12px; line-height: 1.7; }

.uv-drop { display: block; width: 100%; padding: 22px 16px; border: 1px dashed #c9c6bd; border-radius: 4px; background: #fff; cursor: pointer; text-align: center; }
.uv-drop:hover, .uv-drop.arco-upload-draggable:hover { border-color: #0b0b0d; }
.uv-drop-inner { display: flex; flex-direction: column; align-items: center; gap: 6px; }
.uv-drop-inner img { max-height: 160px; border: 1px solid #e4e2dc; }
.uv-drop-plus { font-family: ui-monospace, Menlo, monospace; font-size: 24px; color: #6b6b70; }
.uv-drop-label { margin: 0; font-weight: 500; }
.uv-drop-hint { font-size: 11px; letter-spacing: .04em; }

.uv-handoff { display: flex; flex-direction: column; gap: 12px; }
.uv-handoff-head { display: flex; align-items: center; gap: 8px; }
.uv-handoff-title { font-weight: 600; margin: 0; }
.uv-handoff-status { display: flex; align-items: center; gap: 8px; font-family: ui-monospace, Menlo, monospace; font-size: 12px; color: #6b6b70; }
.uv-handoff-done { display: flex; flex-direction: column; align-items: center; gap: 8px; padding: 8px 0; }

.uv-camera { position: relative; border-radius: 4px; overflow: hidden; background: #000; aspect-ratio: 4 / 3; }
.uv-camera video { width: 100%; height: 100%; object-fit: cover; display: block; }
.uv-camera canvas { position: absolute; inset: 0; width: 100%; height: 100%; }
.uv-camera-overlay { position: absolute; inset: 0; display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 8px; padding: 24px; text-align: center; color: #e2e8f0; background: rgba(0, 0, 0, .7); }
.uv-camera-overlay h3 { margin: 0; }
.uv-camera-overlay .uv-hint { color: #cbd5e1; }
.uv-camera-icon { font-size: 32px; }
.uv-camera-icon-error { color: #ef4444; }
.uv-camera-status { display: flex; align-items: center; gap: 12px; flex-wrap: wrap; font-size: 12px; color: #6b6b70; }
.uv-countdown { display: flex; flex-direction: column; align-items: center; font-family: ui-monospace, Menlo, monospace; }
.uv-countdown strong { font-size: 40px; line-height: 1; }
.uv-voice-digits { text-align: center; background: #faf9f6; }
.uv-voice-digits strong { display: block; font-size: 28px; letter-spacing: .2em; margin: 8px 0; }
.uv-mic { width: 76px; height: 76px; border-radius: 50%; border: 2px solid #e4e2dc; display: flex; align-items: center; justify-content: center; font-size: 26px; color: #6b6b70; }
.uv-mic-active { border-color: #0b0b0d; color: #0b0b0d; }
.uv-result-grid { text-align: left; }
</style>
