<template>
  <div class="mv-shell">
    <!-- Loading -->
    <div v-if="loading" class="mv-state">
      <a-spin :size="40" />
      <span class="mv-mono">PREPARING SECURE SESSION…</span>
    </div>

    <!-- Error -->
    <div v-else-if="error" class="mv-state mv-state-error">
      <div class="mv-error-mark">!</div>
      <h2>We could not open this verification</h2>
      <p class="mv-mono">{{ error }}</p>
    </div>

    <!-- Flow -->
    <template v-else>
      <div class="mv-statusbar">
        <span>{{ clock }}</span>
        <span class="mv-accent mv-statusbar-tag">SECURE SESSION</span>
        <span class="mv-signal"><i>●</i><i>●</i><i>●</i></span>
      </div>

      <div v-if="brandingCompany || brandingLogo" class="mv-brand">
        <img v-if="brandingLogo" :src="brandingLogo" :alt="brandingCompany || 'Logo'" />
        <span v-else class="mv-mono">{{ brandingCompany }}</span>
      </div>

      <a-steps :current="activeStepIdx" size="small" class="mv-steps">
        <a-step v-for="label in stepLabels" :key="label" :title="label" />
      </a-steps>

      <div class="mv-screen">
        <!-- Front ID -->
        <template v-if="screen === 'front'">
          <span class="mv-kicker">{{ stepCounter(1) }} · {{ isAgeOnly ? 'UPLOAD ID' : 'FRONT ID' }}</span>
          <h1>{{ isAgeOnly ? 'Upload your ID' : 'Take a photo of your ID' }}</h1>
          <p class="mv-hint">
            {{ isAgeOnly
              ? `We only read your date of birth to confirm you are ${ageThreshold}+.`
              : 'Place your ID on a flat, well-lit surface. Keep all four corners inside the frame.' }}
          </p>

          <a-select v-model="documentType" class="mv-field">
            <a-option value="national_id">National ID</a-option>
            <a-option value="passport">Passport</a-option>
            <a-option value="drivers_license">Driving licence</a-option>
          </a-select>

          <a-alert type="info" class="mv-tip">Avoid glare and shadows across the document.</a-alert>

          <div class="mv-viewfinder" :class="{ 'mv-viewfinder-filled': !!frontPreview }">
            <img v-if="frontPreview" :src="frontPreview" alt="Front of ID preview" />
            <div v-else class="mv-viewfinder-empty">
              <a-spin v-if="isProcessing" :size="28" />
              <template v-else>
                <icon-idcard class="mv-viewfinder-icon" />
                <span class="mv-mono">ALIGN YOUR ID HERE</span>
              </template>
            </div>
          </div>

          <a-button v-if="!frontFile" type="primary" long size="large" @click="openDocumentCamera('front')">
            Take front photo
          </a-button>
          <a-button v-else type="primary" long size="large" :loading="isProcessing" @click="uploadFront">
            {{ isProcessing ? 'Processing…' : 'Scan front of ID' }}
          </a-button>
          <p v-if="stepError" class="mv-step-error">{{ stepError }}</p>
        </template>

        <!-- Back ID -->
        <template v-else-if="screen === 'back'">
          <span class="mv-kicker">{{ stepCounter(2) }} · BACK ID</span>
          <h1>Now the back of your ID</h1>
          <p class="mv-hint">The back carries the machine-readable zone we compare with the front.</p>

          <a-alert type="info" class="mv-tip">Hold the phone steady and fill the frame.</a-alert>

          <div class="mv-viewfinder" :class="{ 'mv-viewfinder-filled': !!backPreview }">
            <img v-if="backPreview" :src="backPreview" alt="Back of ID preview" />
            <div v-else class="mv-viewfinder-empty">
              <a-spin v-if="isProcessing" :size="28" />
              <template v-else>
                <icon-idcard class="mv-viewfinder-icon" />
                <span class="mv-mono">ALIGN THE BACK HERE</span>
              </template>
            </div>
          </div>

          <a-button v-if="!backFile" type="primary" long size="large" @click="openDocumentCamera('back')">
            Take back photo
          </a-button>
          <a-button v-else type="primary" long size="large" :loading="isProcessing" @click="uploadBack">
            {{ isProcessing ? 'Processing…' : 'Scan back of ID' }}
          </a-button>
          <p v-if="stepError" class="mv-step-error">{{ stepError }}</p>
        </template>

        <!-- Checking -->
        <template v-else-if="screen === 'checking'">
          <div class="mv-centered">
            <span class="mv-kicker">{{ stepCounter(skipBack ? 2 : 3) }} · CHECKING</span>
            <a-spin :size="40" />
            <p class="mv-checking-msg">{{ checkingMessage }}</p>
            <p class="mv-mono mv-hint">THIS TAKES A MOMENT</p>
            <a-space wrap class="mv-gap">
              <a-tag>Document read</a-tag>
              <a-tag>Details matched</a-tag>
              <a-tag>Security checks</a-tag>
            </a-space>
            <p v-if="stepError" class="mv-step-error">{{ stepError }}</p>
          </div>
        </template>

        <!-- Live photo -->
        <template v-else-if="screen === 'live'">
          <span class="mv-kicker">{{ stepCounter(4) }} · LIVE PHOTO</span>

          <template v-if="!showLiveness && !selfieFile && !fallbackSelfie">
            <h1>Liveness check</h1>
            <p class="mv-hint">We record a short burst of frames to confirm you are physically present.</p>
            <a-alert type="info" class="mv-tip">Face the camera in even lighting. Remove sunglasses and hats.</a-alert>
            <a-button type="primary" long size="large" :disabled="isProcessing" @click="showLiveness = true">
              Start liveness check
            </a-button>
          </template>

          <template v-else-if="showLiveness">
            <h1>Liveness check</h1>
            <p class="mv-hint">Keep your face inside the oval until the capture finishes.</p>

            <div class="mv-camera">
              <video ref="videoEl" playsinline muted autoplay />
              <canvas ref="overlayEl" width="640" height="480" />
              <div v-if="cameraState === 'prompt'" class="mv-camera-state">
                <icon-camera class="mv-camera-icon" />
                <a-button type="primary" :loading="cameraBusy" @click="startCamera">Enable camera</a-button>
              </div>
              <div v-else-if="cameraState === 'error'" class="mv-camera-state">
                <icon-exclamation-circle class="mv-camera-icon mv-camera-icon-error" />
                <p class="mv-mono">{{ cameraError }}</p>
                <a-space>
                  <a-button size="small" @click="retryCamera">Retry</a-button>
                  <a-button size="small" @click="useFallbackSelfie">Upload a selfie instead</a-button>
                </a-space>
              </div>
            </div>

            <div class="mv-face" :class="faceDetected ? 'mv-face-ok' : 'mv-face-bad'">
              <a-badge :status="faceDetected ? 'success' : 'danger'" />
              <span class="mv-mono">{{ faceDetected ? 'FACE DETECTED — READY' : 'POSITION YOUR FACE' }}</span>
            </div>

            <div v-if="countdown !== null" class="mv-countdown">
              <strong>{{ countdown }}</strong>
              <span class="mv-mono">HOLD STILL</span>
            </div>

            <a-button
              v-if="cameraState === 'ready' && countdown === null"
              type="primary"
              long
              size="large"
              :disabled="!faceDetected || captureBusy"
              :loading="captureBusy"
              @click="startChallenge"
            >
              {{ captureBusy ? 'Processing…' : faceDetected ? 'Capture selfie' : 'Position your face' }}
            </a-button>
            <p v-if="stepError" class="mv-step-error">{{ stepError }}</p>
          </template>

          <template v-else>
            <h1>Upload a selfie</h1>
            <p class="mv-hint">Your camera is not available — choose a recent photo instead.</p>
            <div class="mv-viewfinder mv-viewfinder-selfie">
              <img v-if="selfiePreview" :src="selfiePreview" alt="Selfie preview" />
              <div v-else class="mv-viewfinder-empty">
                <icon-user class="mv-viewfinder-icon" />
                <span class="mv-mono">FACE THE CAMERA</span>
              </div>
            </div>
            <input ref="selfieInput" type="file" accept="image/*" capture="user" class="mv-hidden" @change="onSelfiePicked" />
            <a-button v-if="!selfieFile" type="primary" long size="large" @click="pickSelfie">Choose selfie</a-button>
            <a-button v-else type="primary" long size="large" :loading="isProcessing" @click="uploadSelfie">
              {{ isProcessing ? 'Processing…' : 'Submit selfie' }}
            </a-button>
            <p v-if="stepError" class="mv-step-error">{{ stepError }}</p>
          </template>
        </template>

        <!-- Voice -->
        <template v-else-if="screen === 'voice'">
          <div class="mv-centered">
            <span class="mv-kicker">{{ stepCounter(5) }} · VOICE</span>
            <h1 class="mv-h1-sm">Voice verification</h1>
            <p class="mv-hint">Read the digits aloud so we can bind your voice to this session.</p>

            <div class="mv-mic" :class="{ 'mv-mic-active': voiceIsRecording }"><icon-voice /></div>

            <a-card v-if="voiceDigits && (voiceExpiresIn === null || voiceExpiresIn > 0)" class="mv-voice-digits">
              <div class="mv-mono mv-hint">SPEAK THESE DIGITS</div>
              <strong>{{ voiceDigits }}</strong>
              <div v-if="voiceExpiresIn !== null" class="mv-mono mv-hint">Expires in {{ voiceExpiresIn }}s</div>
            </a-card>

            <p v-if="voiceIsRecording" class="mv-mono mv-recording">RECORDING… {{ voiceSeconds }}s</p>
            <p v-else-if="voiceHasRecording" class="mv-mono mv-recorded">RECORDING CAPTURED ({{ voiceSeconds }}s)</p>

            <a-button v-if="!voiceDigits" type="primary" long size="large" :loading="voiceBusy" @click="requestVoiceChallenge">
              Get challenge
            </a-button>
            <a-button
              v-if="voiceDigits && !voiceIsRecording && !voiceHasRecording && (voiceExpiresIn === null || voiceExpiresIn > 0)"
              type="primary"
              long
              size="large"
              @click="startVoiceRecording"
            >
              Start recording
            </a-button>
            <a-button v-if="voiceIsRecording" type="primary" long size="large" @click="stopVoiceRecording">
              Stop recording
            </a-button>
            <a-button v-if="voiceHasRecording && !voiceIsRecording" type="primary" long size="large" :loading="voiceBusy" @click="submitVoice">
              Submit voice sample
            </a-button>
            <a-button
              v-if="voiceDigits && voiceExpiresIn !== null && voiceExpiresIn <= 0"
              type="primary"
              long
              size="large"
              @click="restartVoiceChallenge"
            >
              Request new challenge
            </a-button>

            <p v-if="stepError" class="mv-step-error">{{ stepError }}</p>
            <a-button v-if="stepError" long @click="resetVoice">Try again</a-button>
          </div>
        </template>

        <!-- Complete -->
        <template v-else-if="screen === 'done'">
          <div class="mv-centered">
            <template v-if="!finalResult">
              <a-spin :size="40" />
              <p class="mv-checking-msg">Processing your verification</p>
              <p class="mv-mono mv-hint">{{ isDocumentOnly ? 'FINALIZING RESULT' : 'ANALYZING RESULTS' }}</p>
            </template>

            <template v-else-if="resultStatus === 'verified'">
              <div class="mv-success-mark">✓</div>
              <span class="mv-kicker">{{ isAgeOnly ? 'AGE VERIFIED' : isDocumentOnly ? 'DOCUMENT VERIFIED' : 'COMPLETE' }}</span>
              <h1>You are all set</h1>
              <p class="mv-hint">
                {{ redirectUrl ? 'Redirecting you back…'
                  : isAgeOnly ? 'Your age has been confirmed.'
                    : isDocumentOnly ? 'Your document has been verified.'
                      : 'Your identity has been verified.' }}
              </p>

              <div class="mv-checklist">
                <div v-for="item in successChecklist" :key="item" class="mv-checklist-row">
                  <span class="mv-accent">✓</span>
                  <span>{{ item }}</span>
                </div>
              </div>

              <a-alert v-if="patchFailed" type="warning" class="mv-gap">
                We could not notify the desktop session. You can close this window.
              </a-alert>
              <p v-if="redirectUrl" class="mv-mono mv-hint mv-gap">REDIRECTING…</p>
            </template>

            <template v-else>
              <div class="mv-fail-mark">{{ resultStatus === 'failed' ? '✕' : '?' }}</div>
              <h1>
                {{ resultStatus === 'failed'
                  ? isAgeOnly ? 'Age not verified' : isDocumentOnly ? 'Document not verified' : 'Verification failed'
                  : 'Under review' }}
              </h1>
              <p class="mv-hint">
                {{ resultStatus === 'failed'
                  ? (finalResult.message || 'We could not verify these details. You can try again.')
                  : 'A reviewer will check this verification manually.' }}
              </p>

              <template v-if="resultStatus === 'failed' && finalResult.retry_available === true">
                <a-button type="primary" long size="large" :loading="retryProcessing" @click="handleRetry">
                  {{ retryProcessing ? 'Restarting…' : 'Try again' }}
                </a-button>
                <a-button long :disabled="retryProcessing" @click="returnWithoutRetry">
                  Return without retrying
                </a-button>
              </template>
              <p v-if="resultStatus === 'failed' && finalResult.retry_available === false" class="mv-mono mv-step-error">
                No retries remaining.
              </p>

              <a-alert v-if="patchFailed" type="warning" class="mv-gap">
                We could not notify the desktop session. You can close this window.
              </a-alert>
              <p v-if="redirectUrl" class="mv-mono mv-hint mv-gap">REDIRECTING…</p>
            </template>
          </div>
        </template>
      </div>
    </template>

    <!-- Guided document camera overlay -->
    <a-modal
      v-model:visible="showCamera"
      fullscreen
      hide-title
      :footer="false"
      :mask-closable="false"
      class="mv-modal"
    >
      <div class="mv-camera mv-camera-full">
        <video ref="docVideoEl" playsinline muted autoplay />
        <canvas ref="docOverlayEl" width="640" height="480" />
        <div v-if="docCameraState === 'prompt'" class="mv-camera-state">
          <icon-camera class="mv-camera-icon" />
          <p class="mv-mono">CAMERA ACCESS REQUIRED</p>
          <a-button type="primary" :loading="docCameraBusy" @click="startDocumentCamera">Enable camera</a-button>
        </div>
        <div v-else-if="docCameraState === 'error'" class="mv-camera-state">
          <icon-exclamation-circle class="mv-camera-icon mv-camera-icon-error" />
          <p class="mv-mono">{{ docCameraError }}</p>
          <a-space>
            <a-button size="small" @click="fallbackDocumentUpload">Upload a photo instead</a-button>
            <a-button size="small" @click="closeDocumentCamera">Close</a-button>
          </a-space>
        </div>
        <div class="mv-camera-bar">
          <a-button size="small" @click="closeDocumentCamera">Cancel</a-button>
          <a-button
            type="primary"
            :disabled="docCameraState !== 'ready'"
            :loading="docCaptureBusy"
            @click="captureDocument"
          >
            Capture {{ cameraVariant }}
          </a-button>
        </div>
      </div>
    </a-modal>

    <input ref="frontInput" type="file" accept="image/*" capture="environment" class="mv-hidden" @change="onFrontPicked" />
    <input ref="backInput" type="file" accept="image/*" capture="environment" class="mv-hidden" @change="onBackPicked" />
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { api, apiMessage } from '../lib/api'

type Screen = 'front' | 'back' | 'checking' | 'live' | 'voice' | 'done'

const FULL_STEPS = ['Front ID', 'Back ID', 'Checking', 'Live Photo', 'Complete']
const FULL_VOICE_STEPS = ['Front ID', 'Back ID', 'Checking', 'Live Photo', 'Voice', 'Complete']
const DOCUMENT_ONLY_STEPS = ['Front ID', 'Back ID', 'Checking', 'Complete']
const IDENTITY_STEPS = ['Front ID', 'Checking', 'Live Photo', 'Complete']
const AGE_ONLY_STEPS = ['Upload ID', 'Complete']
const PASSPORT_DOC_ONLY_STEPS = ['Front ID', 'Checking', 'Complete']
const MAX_POLL_ATTEMPTS = 60

const route = useRoute()
const query = (key: string) => {
  const value = route.query[key]
  return typeof value === 'string' ? value : ''
}

const token = query('token')
const redirectUrl = query('redirect_url')
const requestMode = query('verification_mode')
const ageThreshold = Number(query('age_threshold')) || 18
const isAgeOnly = requestMode === 'age_only'
const isDocumentOnly = requestMode === 'document_only'
const isIdentity = requestMode === 'identity'

const loading = ref(true)
const error = ref('')
const screen = ref<Screen>('front')
const verificationId = ref('')
const processing = ref(false)
const stepError = ref('')
const patchFailed = ref(false)
const retryProcessing = ref(false)
const finalResult = ref<Record<string, any> | null>(null)
const skipBack = ref(false)
const documentType = ref('national_id')
const frontFile = ref<File | null>(null)
const frontPreview = ref('')
const backFile = ref<File | null>(null)
const backPreview = ref('')
const selfieFile = ref<File | null>(null)
const selfiePreview = ref('')
const checkingMessage = ref('Verifying your document…')
const brandingLogo = ref('')
const brandingCompany = ref('')
const clock = ref(formatClock(new Date()))

const isProcessing = computed(() => processing.value)

/* Camera state */
const showLiveness = ref(false)
const fallbackSelfie = ref(false)
const cameraState = ref<'prompt' | 'initializing' | 'ready' | 'error'>('prompt')
const cameraBusy = ref(false)
const captureBusy = ref(false)
const cameraError = ref('')
const faceDetected = ref(false)
const livenessScore = ref(0)
const faceStability = ref(0)
const countdown = ref<number | null>(null)
const videoEl = ref<HTMLVideoElement | null>(null)
const overlayEl = ref<HTMLCanvasElement | null>(null)
const selfieInput = ref<HTMLInputElement | null>(null)
const frontInput = ref<HTMLInputElement | null>(null)
const backInput = ref<HTMLInputElement | null>(null)

/* Guided document camera */
const showCamera = ref(false)
const cameraVariant = ref<'front' | 'back'>('front')
const docVideoEl = ref<HTMLVideoElement | null>(null)
const docOverlayEl = ref<HTMLCanvasElement | null>(null)
const docCameraState = ref<'prompt' | 'initializing' | 'ready' | 'error'>('prompt')
const docCameraBusy = ref(false)
const docCaptureBusy = ref(false)
const docCameraError = ref('')

/* Voice */
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

/* Timers / streams */
let mediaStream: MediaStream | null = null
let docStream: MediaStream | null = null
let rafId: number | null = null
let docRafId: number | null = null
let analysisCanvas: HTMLCanvasElement | null = null
let checkingTimer: number | null = null
let clockTimer: number | null = null
let redirectTimer: number | null = null
const faceHistory: boolean[] = []
const stabHistory: number[] = []

function formatClock(date: Date) {
  return `${String(date.getHours()).padStart(2, '0')}:${String(date.getMinutes()).padStart(2, '0')}`
}

function headers(): Record<string, string> {
  return { 'X-Handoff-Token': token }
}

const stepLabels = computed(() => {
  if (isAgeOnly) return AGE_ONLY_STEPS
  if (isIdentity || skipBack.value) {
    return isDocumentOnly ? PASSPORT_DOC_ONLY_STEPS : IDENTITY_STEPS
  }
  if (isDocumentOnly) return DOCUMENT_ONLY_STEPS
  return screen.value === 'voice' ? FULL_VOICE_STEPS : FULL_STEPS
})

const activeStepIdx = computed(() => {
  const order: Screen[] = ['front', 'back', 'checking', 'live', 'voice', 'done']
  const index = order.indexOf(screen.value)
  if (isAgeOnly) return index >= 5 ? 1 : 0
  if (isIdentity || skipBack.value) {
    if (isDocumentOnly) return index === 0 ? 0 : index === 2 ? 1 : 2
    return index === 0 ? 0 : index === 2 ? 1 : index === 3 ? 2 : 3
  }
  if (isDocumentOnly) return index >= 4 ? 3 : index
  return index
})

const resultStatus = computed(() => {
  const value = finalResult.value?.final_result ?? finalResult.value?.status ?? ''
  return typeof value === 'string' ? value : ''
})

const successChecklist = computed(() => {
  if (isAgeOnly) {
    return [
      'Document scanned',
      `Age requirement met (${finalResult.value?.age_verification?.age_threshold ?? ageThreshold}+)`,
    ]
  }
  if (isDocumentOnly) return ['Document scanned', 'Details confirmed']
  if (isIdentity) return ['Document verified', 'Liveness passed', 'Face matched']
  return ['Document verified', 'Details confirmed', 'Liveness passed', 'Face matched']
})

function stepCounter(current: number) {
  const total = isAgeOnly ? 1 : isIdentity || skipBack.value ? 4 : isDocumentOnly ? 4 : 5
  return `STEP ${current} OF ${total}`
}

function sleep(ms: number) {
  return new Promise((resolve) => window.setTimeout(resolve, ms))
}

async function apiGet(path: string) {
  const { data } = await api.get(path, { headers: headers() })
  return data
}

async function loadSession() {
  if (!token) {
    error.value = 'Missing handoff token. Scan the QR code on the desktop again.'
    loading.value = false
    return
  }
  try {
    const { data } = await api.get(`/verify/handoff/${token}/session`)
    if (!data?.user_id) throw new Error('This handoff session is incomplete. Scan the QR code again.')
    if (data.branding) {
      brandingLogo.value = data.branding.logo_url || ''
      brandingCompany.value = data.branding.company_name || ''
    }
    if (data.verification_id) {
      verificationId.value = data.verification_id
      loading.value = false
      await showResultIfTerminal(data.verification_id, true)
      return
    }
    await initializeVerification(data.user_id, data.source)
    loading.value = false
  } catch (err) {
    const status = (err as { response?: { status?: number } }).response?.status
    error.value = status === 410
      ? 'This handoff link has expired. Scan the QR code again.'
      : status === 409
        ? 'This link has already been used.'
        : apiMessage(err)
    loading.value = false
  }
}

async function initializeVerification(userId: string, source?: string) {
  try {
    const { data } = await api.post('/v2/verify/initialize', {
      user_id: userId,
      ...(source ? { source } : {}),
    }, { headers: { 'Content-Type': 'application/json', ...headers() } })
    verificationId.value = data.verification_id
    api.patch(`/verify/handoff/${token}/link`, { verification_id: data.verification_id }).catch(() => {})
  } catch (err) {
    error.value = apiMessage(err)
  }
}

async function showResultIfTerminal(id: string, failedOnly = false) {
  try {
    const status = await apiGet(`/v2/verify/${id}/status`)
    const terminal = failedOnly ? ['failed'] : ['failed', 'manual_review', 'verified']
    if (terminal.includes(status?.final_result)) {
      showFinalResult(status)
      return true
    }
  } catch {
    /* fall through to the normal path */
  }
  return false
}

/* ── Front / back document ─────────────────────────────────────────────── */

function onFrontPicked(event: Event) {
  const file = (event.target as HTMLInputElement).files?.[0]
  if (!file) return
  if (frontPreview.value) URL.revokeObjectURL(frontPreview.value)
  frontFile.value = file
  frontPreview.value = URL.createObjectURL(file)
  stepError.value = ''
}

function onBackPicked(event: Event) {
  const file = (event.target as HTMLInputElement).files?.[0]
  if (!file) return
  if (backPreview.value) URL.revokeObjectURL(backPreview.value)
  backFile.value = file
  backPreview.value = URL.createObjectURL(file)
  stepError.value = ''
}

async function uploadFront() {
  if (!frontFile.value || !verificationId.value) return
  processing.value = true
  stepError.value = ''
  try {
    const form = new FormData()
    form.append('document_type', documentType.value)
    form.append('document', frontFile.value)
    const { data } = await api.post(`/v2/verify/${verificationId.value}/front-document`, form, { headers: headers() })
    if (isAgeOnly && data.age_verification) {
      showFinalResult(data)
      return
    }
    if (data?.rejection_reason) {
      stepError.value = data.message || 'The photo of your ID is not clear enough. Please retake it in good lighting.'
      frontFile.value = null
      if (frontPreview.value) URL.revokeObjectURL(frontPreview.value)
      frontPreview.value = ''
      return
    }
    skipBack.value = isIdentity || data?.detected_document_type === 'passport' || data?.requires_back === false
    if (skipBack.value) {
      screen.value = 'checking'
      await pollFrontOCRForIdentity()
      return
    }
    screen.value = 'back'
  } catch (err) {
    stepError.value = apiMessage(err)
  } finally {
    processing.value = false
  }
}

async function pollFrontOCRForIdentity() {
  for (let attempt = 0; attempt < MAX_POLL_ATTEMPTS; attempt++) {
    await sleep(2000)
    try {
      const data = await apiGet(`/v2/verify/${verificationId.value}/status`)
      if (data?.final_result) {
        showFinalResult(data)
        return
      }
      if (data?.status !== 'FRONT_PROCESSING') {
        screen.value = 'live'
        return
      }
    } catch {
      /* retry */
    }
  }
  stepError.value = 'Document processing timed out. Please try again.'
  screen.value = 'front'
}

async function uploadBack() {
  if (!backFile.value || !verificationId.value) return
  processing.value = true
  stepError.value = ''
  try {
    const form = new FormData()
    form.append('document', backFile.value)
    form.append('document_type', documentType.value)
    await api.post(`/v2/verify/${verificationId.value}/back-document`, form, { headers: headers() })
    screen.value = 'checking'
    await api.post(`/v2/verify/${verificationId.value}/cross-validation`, {}, { headers: headers() })
    await pollCrossValidation()
  } catch (err) {
    stepError.value = apiMessage(err)
    screen.value = 'back'
  } finally {
    processing.value = false
  }
}

async function pollCrossValidation() {
  for (let attempt = 0; attempt < MAX_POLL_ATTEMPTS; attempt++) {
    await sleep(2000)
    try {
      const data = await apiGet(`/v2/verify/${verificationId.value}/status`)
      if (data?.final_result) {
        showFinalResult(data)
        return
      }
      if (data?.status === 'AWAITING_LIVE' || data?.status === 'FACE_MATCHING') {
        screen.value = 'live'
        return
      }
    } catch {
      /* retry */
    }
  }
  stepError.value = 'Cross-check is taking too long. Please try again.'
  screen.value = 'back'
}

/* ── Guided document camera ────────────────────────────────────────────── */

async function openDocumentCamera(variant: 'front' | 'back') {
  cameraVariant.value = variant
  if (!navigator.mediaDevices?.getUserMedia) {
    fallbackDocumentUpload()
    return
  }
  showCamera.value = true
  docCameraState.value = 'prompt'
  await nextTick()
  await startDocumentCamera()
}

async function startDocumentCamera() {
  docCameraState.value = 'initializing'
  docCameraBusy.value = true
  docCameraError.value = ''
  try {
    docStream = await navigator.mediaDevices.getUserMedia({
      video: { width: { ideal: 1280 }, height: { ideal: 960 }, facingMode: 'environment' },
      audio: false,
    })
    const video = docVideoEl.value
    if (video) {
      video.srcObject = docStream
      await video.play()
    }
    docCameraState.value = 'ready'
    docRafId = window.requestAnimationFrame(docCameraLoop)
  } catch (err) {
    docCameraState.value = 'error'
    docCameraError.value = (err as Error).message || 'Camera unavailable.'
  } finally {
    docCameraBusy.value = false
  }
}

function docCameraLoop() {
  const video = docVideoEl.value
  const overlay = docOverlayEl.value
  if (docCameraState.value !== 'ready') return
  if (!video || !overlay) {
    docRafId = window.requestAnimationFrame(docCameraLoop)
    return
  }
  if (video.readyState >= 2 && video.videoWidth > 0) {
    const ctx = overlay.getContext('2d')
    if (ctx) {
      ctx.clearRect(0, 0, overlay.width, overlay.height)
      const insetX = overlay.width * 0.08
      const insetY = overlay.height * 0.2
      ctx.strokeStyle = 'rgba(255,255,255,0.9)'
      ctx.lineWidth = 2
      ctx.strokeRect(insetX, insetY, overlay.width - insetX * 2, overlay.height - insetY * 2)
    }
  }
  docRafId = window.requestAnimationFrame(docCameraLoop)
}

async function captureDocument() {
  const video = docVideoEl.value
  if (!video) return
  docCaptureBusy.value = true
  try {
    const canvas = document.createElement('canvas')
    canvas.width = video.videoWidth || 1280
    canvas.height = video.videoHeight || 960
    canvas.getContext('2d')?.drawImage(video, 0, 0, canvas.width, canvas.height)
    const blob = await new Promise<Blob | null>((resolve) => canvas.toBlob(resolve, 'image/jpeg', 0.9))
    if (!blob) throw new Error('Failed to capture image.')
    const file = new File([blob], `${cameraVariant.value}.jpg`, { type: 'image/jpeg' })
    if (cameraVariant.value === 'front') {
      if (frontPreview.value) URL.revokeObjectURL(frontPreview.value)
      frontFile.value = file
      frontPreview.value = URL.createObjectURL(file)
    } else {
      if (backPreview.value) URL.revokeObjectURL(backPreview.value)
      backFile.value = file
      backPreview.value = URL.createObjectURL(file)
    }
    stepError.value = ''
    closeDocumentCamera()
  } catch (err) {
    docCameraError.value = apiMessage(err)
  } finally {
    docCaptureBusy.value = false
  }
}

function closeDocumentCamera() {
  if (docRafId !== null) {
    window.cancelAnimationFrame(docRafId)
    docRafId = null
  }
  docStream?.getTracks().forEach((track) => track.stop())
  docStream = null
  const video = docVideoEl.value
  if (video) video.srcObject = null
  docCameraState.value = 'prompt'
  showCamera.value = false
}

function fallbackDocumentUpload() {
  closeDocumentCamera()
  const input = cameraVariant.value === 'front' ? frontInput.value : backInput.value
  window.setTimeout(() => input?.click(), 100)
}

/* ── Selfie fallback ──────────────────────────────────────────────────── */

function useFallbackSelfie() {
  showLiveness.value = false
  fallbackSelfie.value = true
  stopCamera()
  window.setTimeout(() => selfieInput.value?.click(), 150)
}

function pickSelfie() {
  selfieInput.value?.click()
}

function onSelfiePicked(event: Event) {
  const file = (event.target as HTMLInputElement).files?.[0]
  if (!file) return
  if (selfiePreview.value) URL.revokeObjectURL(selfiePreview.value)
  selfieFile.value = file
  selfiePreview.value = URL.createObjectURL(file)
  stepError.value = ''
}

async function uploadSelfie() {
  if (!selfieFile.value || !verificationId.value) return
  processing.value = true
  stepError.value = ''
  try {
    const form = new FormData()
    form.append('selfie', selfieFile.value)
    const { data } = await api.post(`/v2/verify/${verificationId.value}/live-capture`, form, { headers: headers() })
    if (data?.final_result) {
      showFinalResult(data)
      return
    }
    screen.value = 'done'
    await finalizeThenPoll()
  } catch (err) {
    stepError.value = apiMessage(err)
  } finally {
    processing.value = false
  }
}

/* ── Liveness camera ─────────────────────────────────────────────────── */

async function startCamera() {
  cameraState.value = 'initializing'
  cameraError.value = ''
  cameraBusy.value = true
  try {
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
  } catch (err) {
    cameraState.value = 'error'
    cameraError.value = (err as Error).message || 'Camera unavailable.'
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
  countdown.value = null
  cameraState.value = 'prompt'
}

function retryCamera() {
  stopCamera()
  faceHistory.length = 0
  stabHistory.length = 0
  void startCamera()
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
  const radiusX = Math.min(width, height) * 0.24
  const radiusY = radiusX * 1.35
  ctx.clearRect(0, 0, width, height)
  ctx.strokeStyle = faceDetected.value ? '#10b981' : '#ef4444'
  ctx.lineWidth = 3
  ctx.setLineDash([10, 8])
  ctx.beginPath()
  ctx.ellipse(cx, cy, radiusX, radiusY, 0, 0, 2 * Math.PI)
  ctx.stroke()
  ctx.setLineDash([])
  ctx.fillStyle = ctx.strokeStyle
  ctx.font = '14px monospace'
  ctx.textAlign = 'center'
  ctx.fillText(faceDetected.value ? 'FACE DETECTED' : 'POSITION FACE IN FRAME', cx, cy + radiusY + 26)
}

function startChallenge() {
  if (countdown.value !== null) return
  if (!faceDetected.value) {
    stepError.value = 'No face detected. Keep your face inside the oval.'
    return
  }
  stepError.value = ''
  countdown.value = 3
  const timer = window.setInterval(() => {
    if (countdown.value === null) {
      window.clearInterval(timer)
      return
    }
    if (!faceDetected.value) {
      window.clearInterval(timer)
      countdown.value = null
      stepError.value = 'Face lost during countdown. Please try again.'
      return
    }
    if (countdown.value <= 1) {
      window.clearInterval(timer)
      countdown.value = null
      void captureSelfie()
    } else {
      countdown.value -= 1
    }
  }, 1000)
}

async function captureSelfie() {
  const video = videoEl.value
  if (!video || !verificationId.value) return
  captureBusy.value = true
  stepError.value = ''
  try {
    const canvas = document.createElement('canvas')
    canvas.width = video.videoWidth || 640
    canvas.height = video.videoHeight || 480
    canvas.getContext('2d')?.drawImage(video, 0, 0, canvas.width, canvas.height)
    const blob = await new Promise<Blob | null>((resolve) => canvas.toBlob(resolve, 'image/jpeg', 0.85))
    if (!blob) throw new Error('Failed to capture image.')
    const form = new FormData()
    form.append('selfie', blob, 'selfie.jpg')
    const { data } = await api.post(`/v2/verify/${verificationId.value}/live-capture`, form, { headers: headers() })
    showLiveness.value = false
    stopCamera()
    if (data?.status === 'AWAITING_VOICE') {
      screen.value = 'voice'
      return
    }
    if (data?.final_result) {
      showFinalResult(data)
      return
    }
    screen.value = 'done'
    await finalizeThenPoll()
  } catch (err) {
    stepError.value = apiMessage(err)
  } finally {
    captureBusy.value = false
  }
}

/* ── Final result ─────────────────────────────────────────────────────── */

async function finalizeThenPoll() {
  try {
    const { data } = await api.post(`/v2/verify/${verificationId.value}/finalize`, {}, { headers: headers() })
    if (data?.final_result) {
      showFinalResult(data)
      return
    }
  } catch {
    /* fall back to polling */
  }
  for (let attempt = 0; attempt < MAX_POLL_ATTEMPTS; attempt++) {
    await sleep(3000)
    try {
      const data = await apiGet(`/v2/verify/${verificationId.value}/status`)
      if (data?.status === 'AWAITING_VOICE') {
        screen.value = 'voice'
        return
      }
      if (data?.final_result) {
        showFinalResult(data)
        return
      }
    } catch {
      /* retry */
    }
  }
  stepError.value = 'Verification is taking too long. Please close and try again.'
}

function showFinalResult(data: Record<string, any>) {
  finalResult.value = data
  screen.value = 'done'
  if (data.final_result === 'failed' && data.retry_available === true) return
  void finalizeAndExit(data)
}

async function patchWithRetry(url: string, body: object, maxRetries = 3): Promise<boolean> {
  for (let attempt = 0; attempt < maxRetries; attempt++) {
    try {
      await api.patch(url, body)
      return true
    } catch (err) {
      const status = (err as { response?: { status?: number } }).response?.status
      if (status === 409) return true
      if (status === 410) return false
    }
    if (attempt < maxRetries - 1) await sleep(1000 * 2 ** attempt)
  }
  return false
}

async function finalizeAndExit(data: Record<string, any>) {
  const status = data.final_result ?? data.status
  if (redirectUrl) {
    if (redirectTimer !== null) window.clearTimeout(redirectTimer)
    redirectTimer = window.setTimeout(() => {
      const target = new URL(redirectUrl)
      target.searchParams.set('verification_id', String(data.verification_id ?? verificationId.value))
      target.searchParams.set('status', String(status ?? ''))
      window.location.href = target.toString()
    }, 3000)
  }
  if (!token) return
  const ok = await patchWithRetry(`/verify/handoff/${token}/complete`, {
    status: status === 'failed' ? 'failed' : 'completed',
    result: {
      verification_id: data.verification_id ?? verificationId.value,
      status,
      user_id: data.user_id,
      confidence_score: data.confidence_score,
      face_match_score: data.face_match_results?.similarity_score ?? data.face_match_score,
      liveness_score: data.liveness_results?.liveness_score ?? data.liveness_score,
    },
  })
  if (!ok) patchFailed.value = true
}

async function returnWithoutRetry() {
  if (finalResult.value) await finalizeAndExit(finalResult.value)
}

async function handleRetry() {
  if (!verificationId.value) return
  retryProcessing.value = true
  if (redirectTimer !== null) {
    window.clearTimeout(redirectTimer)
    redirectTimer = null
  }
  try {
    await api.post(`/v2/verify/${verificationId.value}/restart`, {}, { headers: headers() })
    if (frontPreview.value) URL.revokeObjectURL(frontPreview.value)
    if (backPreview.value) URL.revokeObjectURL(backPreview.value)
    if (selfiePreview.value) URL.revokeObjectURL(selfiePreview.value)
    frontFile.value = null
    frontPreview.value = ''
    backFile.value = null
    backPreview.value = ''
    selfieFile.value = null
    selfiePreview.value = ''
    finalResult.value = null
    stepError.value = ''
    skipBack.value = false
    showLiveness.value = false
    fallbackSelfie.value = false
    screen.value = 'front'
  } catch (err) {
    stepError.value = apiMessage(err)
  } finally {
    retryProcessing.value = false
  }
}

/* ── Voice challenge ──────────────────────────────────────────────────── */

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

async function requestVoiceChallenge() {
  if (!verificationId.value) return
  voiceBusy.value = true
  stepError.value = ''
  try {
    const { data } = await api.post(`/v2/verify/${verificationId.value}/voice-challenge`, {}, { headers: headers() })
    voiceDigits.value = data.challenge_digits
    voiceExpiresIn.value = data.expires_in_seconds
    if (voiceExpiryTimer !== null) window.clearInterval(voiceExpiryTimer)
    const startedAt = Date.now()
    const total = Number(data.expires_in_seconds) || 0
    voiceExpiryTimer = window.setInterval(() => {
      const remaining = total - Math.floor((Date.now() - startedAt) / 1000)
      voiceExpiresIn.value = Math.max(0, remaining)
      if (remaining <= 0 && voiceExpiryTimer !== null) {
        window.clearInterval(voiceExpiryTimer)
        voiceExpiryTimer = null
      }
    }, 1000)
  } catch (err) {
    stepError.value = apiMessage(err)
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
      void blobToWav(raw)
        .catch(() => raw)
        .then((audio) => {
          voiceBlob = audio
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
  } catch (err) {
    stepError.value = (err as Error).message || 'Microphone access denied.'
  }
}

function stopVoiceRecording() {
  if (voiceTimer !== null) window.clearTimeout(voiceTimer)
  if (voiceRecorder?.state === 'recording') voiceRecorder.stop()
}

async function submitVoice() {
  if (!verificationId.value || !voiceBlob) return
  voiceBusy.value = true
  stepError.value = ''
  try {
    const form = new FormData()
    form.append('file', voiceBlob, 'voice.wav')
    const { data } = await api.post(`/v2/verify/${verificationId.value}/voice-capture`, form, { headers: headers() })
    if (voiceExpiryTimer !== null) window.clearInterval(voiceExpiryTimer)
    if (data?.final_result) {
      showFinalResult(data)
      return
    }
    screen.value = 'done'
    await finalizeThenPoll()
  } catch (err) {
    stepError.value = apiMessage(err)
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

/* ── Lifecycle ────────────────────────────────────────────────────────── */

function startCheckingCycle() {
  stopCheckingCycle()
  const messages = ['Verifying your document…', 'Cross-checking details…', 'Almost there…']
  let index = 0
  checkingMessage.value = messages[0]
  checkingTimer = window.setInterval(() => {
    index = (index + 1) % messages.length
    checkingMessage.value = messages[index]
  }, 1800)
}

function stopCheckingCycle() {
  if (checkingTimer !== null) window.clearInterval(checkingTimer)
  checkingTimer = null
}

watch(screen, (value) => {
  if (value === 'checking') startCheckingCycle()
  else stopCheckingCycle()
})

onMounted(async () => {
  await loadSession()
  clockTimer = window.setInterval(() => {
    clock.value = formatClock(new Date())
  }, 30000)
})

onBeforeUnmount(() => {
  stopCamera()
  closeDocumentCamera()
  resetVoice()
  stopCheckingCycle()
  if (clockTimer !== null) window.clearInterval(clockTimer)
  if (redirectTimer !== null) window.clearTimeout(redirectTimer)
})
</script>

<style scoped>
.mv-shell { min-height: 100vh; display: flex; flex-direction: column; background: #f7f6f3; color: #0b0b0d; padding-bottom: 28px; }
.mv-state { flex: 1; display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 16px; padding: 32px; text-align: center; }
.mv-state-error h2 { margin: 0; font-size: 20px; }
.mv-error-mark { width: 64px; height: 64px; display: grid; place-items: center; font-size: 28px; color: #b45309; border: 1px solid #f0d9a8; background: #fdf6e3; }
.mv-mono { font-family: ui-monospace, SFMono-Regular, Menlo, monospace; font-size: 12px; color: #6b6b70; }
.mv-accent { color: #2563eb; }
.mv-statusbar { display: flex; align-items: center; justify-content: space-between; padding: 16px 24px 8px; font-family: ui-monospace, Menlo, monospace; font-size: 11px; color: #6b6b70; }
.mv-statusbar-tag { letter-spacing: .1em; font-size: 10px; text-transform: uppercase; }
.mv-signal i { font-style: normal; margin-left: 2px; }
.mv-signal i:nth-child(2) { opacity: .7; }
.mv-signal i:nth-child(3) { opacity: .4; }
.mv-brand { padding: 4px 24px 0; text-align: center; }
.mv-brand img { height: 24px; max-width: 60%; object-fit: contain; }
.mv-steps { padding: 12px 16px 0; }
.mv-steps :deep(.arco-steps-item) { flex: 1 1 0; min-width: 0; padding-right: 6px; }
.mv-steps :deep(.arco-steps-item:last-child) { padding-right: 0; }
.mv-steps :deep(.arco-steps-item-node) { display: none; }
.mv-steps :deep(.arco-steps-item-title) { font-size: 10px; line-height: 1.2; color: #6b6b70; white-space: nowrap; }
.mv-steps :deep(.arco-steps-item-active .arco-steps-item-title) { color: #0b0b0d; font-weight: 600; }
.mv-screen { flex: 1; display: flex; flex-direction: column; gap: 12px; padding: 20px 24px 0; }
.mv-kicker { font-family: ui-monospace, Menlo, monospace; font-size: 10px; letter-spacing: .12em; text-transform: uppercase; color: #2563eb; }
.mv-screen h1 { font-size: 26px; line-height: 1.12; letter-spacing: -.02em; margin: 0; }
.mv-h1-sm { font-size: 20px; }
.mv-hint { color: #6b6b70; font-size: 13px; line-height: 1.55; margin: 0; }
.mv-field { width: 100%; }
.mv-tip { font-size: 12px; }
.mv-gap { margin-top: 12px; }
.mv-viewfinder { position: relative; aspect-ratio: 1.586; border: 1px dashed #c9c6bd; background: #fff; display: grid; place-items: center; overflow: hidden; border-radius: 4px; }
.mv-viewfinder-filled { border-style: solid; }
.mv-viewfinder img { width: 100%; height: 100%; object-fit: cover; }
.mv-viewfinder-empty { display: flex; flex-direction: column; align-items: center; gap: 8px; color: #9a978f; }
.mv-viewfinder-icon { font-size: 30px; }
.mv-viewfinder-selfie { aspect-ratio: 3 / 4; }
.mv-hidden { display: none; }
.mv-step-error { margin: 4px 0 0; text-align: center; font-family: ui-monospace, Menlo, monospace; font-size: 12px; color: #b91c1c; }
.mv-centered { display: flex; flex-direction: column; align-items: center; justify-content: center; text-align: center; gap: 12px; flex: 1; padding-bottom: 24px; }
.mv-checking-msg { font-size: 16px; font-weight: 600; margin: 0; }
.mv-camera { position: relative; border-radius: 4px; overflow: hidden; background: #000; aspect-ratio: 3 / 4; }
.mv-camera-full { aspect-ratio: auto; height: 100vh; border-radius: 0; }
.mv-camera video { width: 100%; height: 100%; object-fit: cover; display: block; }
.mv-camera canvas { position: absolute; inset: 0; width: 100%; height: 100%; }
.mv-camera-state { position: absolute; inset: 0; display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 10px; padding: 24px; background: rgba(0, 0, 0, .72); color: #f8fafc; text-align: center; }
.mv-camera-state .mv-mono { color: #e2e8f0; }
.mv-camera-icon { font-size: 32px; }
.mv-camera-icon-error { color: #ef4444; }
.mv-camera-bar { position: absolute; left: 0; right: 0; bottom: 0; display: flex; align-items: center; justify-content: center; gap: 12px; padding: 18px; background: rgba(0, 0, 0, .55); }
.mv-face { display: flex; align-items: center; gap: 8px; padding: 10px 12px; border: 1px solid; }
.mv-face-ok { border-color: #10b981; background: #ecfdf5; }
.mv-face-bad { border-color: #ef4444; background: #fef2f2; }
.mv-countdown { text-align: center; font-family: ui-monospace, Menlo, monospace; }
.mv-countdown strong { display: block; font-size: 34px; line-height: 1; }
.mv-success-mark { width: 96px; height: 96px; display: grid; place-items: center; font-size: 40px; color: #2563eb; border: 2px solid #2563eb; background: #eff4ff; }
.mv-fail-mark { width: 96px; height: 96px; display: grid; place-items: center; font-size: 40px; color: #b45309; border: 1px solid #f0d9a8; background: #fdf6e3; }
.mv-checklist { display: flex; flex-direction: column; gap: 8px; width: 100%; max-width: 320px; }
.mv-checklist-row { display: flex; align-items: center; gap: 11px; padding: 11px 14px; background: #eff4ff; border: 1px solid #e4e2dc; font-size: 13px; }
.mv-mic { width: 72px; height: 72px; border-radius: 50%; border: 2px solid #e4e2dc; display: grid; place-items: center; font-size: 24px; color: #6b6b70; }
.mv-mic-active { border-color: #2563eb; color: #2563eb; }
.mv-voice-digits { text-align: center; width: 100%; max-width: 280px; }
.mv-voice-digits strong { display: block; font-size: 24px; letter-spacing: .2em; margin: 6px 0; }
.mv-recording { color: #2563eb; }
.mv-recorded { color: #15803d; }
.mv-modal :deep(.arco-modal) { background: #000; }
.mv-modal :deep(.arco-modal-close-icon) { color: #f8fafc; }
</style>
