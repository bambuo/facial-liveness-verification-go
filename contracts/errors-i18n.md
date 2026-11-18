# 错误码与 i18n 对账表（P0）

- 来源：`/Users/johana/Codes/github/idswyft-community-1.12.39` @ `83dd51bdfb6dec608856b41c110f54f5a8bf3efe`
- 口径：en / zh-CN locale 全量键 + 源码中的 `messageKey` 字面量引用

## 汇总

| 指标 | 数值 |
| --- | --- |
| en 键 | 142 |
| zh-CN 键 | 142 |
| 合并键 | 143 |
| 缺少 en | 1 |
| 缺少 zh-CN | 1 |
| 源码引用但未定义 | 1 |
| 已定义但未引用 | 31 |

## 对账明细

| Key | en | zh-CN | 引用次数 | 引用文件 | 状态 |
| --- | --- | --- | --- | --- | --- |
| email.credential.greeting | Hi {{name}}, | 你好，{{name}}： | 2 | backend/src/services/emailService.ts | ok |
| email.credential.heading | Your Verified Identity Credential | 你的身份验证凭证 | 1 | backend/src/services/emailService.ts | ok |
| email.credential.intro | Your identity has been verified. Scan the QR code below to present your verified credential. | 你的身份已验证。扫描下方二维码即可出示可验证凭证。 | 2 | backend/src/services/emailService.ts | ok |
| email.credential.secured | Secured by Idswyft | 由 Idswyft 提供安全保障 | 2 | backend/src/services/emailService.ts | ok |
| email.credential.subject | Your Verified Identity Credential | 你的身份验证凭证 | 1 | backend/src/services/emailService.ts | ok |
| email.credential.validUntil | This credential is valid until {{date}}. | 该凭证有效期至 {{date}}。 | 2 | backend/src/services/emailService.ts | ok |
| email.credential.verifyButton | Verify Online | 在线校验 | 1 | backend/src/services/emailService.ts | ok |
| email.credential.verifyOnline | Or verify online: | 或在线校验： | 1 | backend/src/services/emailService.ts | ok |
| email.otp.expires | This code expires in 10 minutes. | 验证码将在 10 分钟后失效。 | 2 | backend/src/services/emailService.ts | ok |
| email.otp.heading | Verification Code | 验证码 | 2 | backend/src/services/emailService.ts | ok |
| email.otp.ignore | If you didn't request this code, you can safely ignore this email. | 如果这不是你的操作，可以忽略此邮件。 | 2 | backend/src/services/emailService.ts | ok |
| email.otp.intro | Enter this code to sign in to your developer account: | 输入以下验证码以登录开发者账户： | 1 | backend/src/services/emailService.ts | ok |
| email.otp.secured | Secured by Idswyft | 由 Idswyft 提供安全保障 | 1 | backend/src/services/emailService.ts | ok |
| email.otp.subject | Your Idswyft verification code | 你的 Idswyft 验证码 | 1 | backend/src/services/emailService.ts | ok |
| errors.ACCESS_DENIED_ROLES | Access denied. Required roles: {{roles}} | 访问被拒绝，需要以下角色：{{roles}} | 1 | backend/src/middleware/auth.ts | ok |
| errors.ACCOUNT_EXISTS | An account with this email already exists. | 该邮箱已注册账号。 | 4 | backend/src/routes/auth.ts<br>backend/src/routes/developer/register.ts | ok |
| errors.API_KEY_ALREADY_ROTATED | This key has already been rotated and is in its grace period. | 该 Key 已轮换，当前处于宽限期。 | 1 | backend/src/routes/developer/apiKeys.ts | ok |
| errors.API_KEY_ENV_MISMATCH | This API key type is not allowed in the current environment. | 当前环境不允许该类型的 API Key。 | 0 | - | 未引用 |
| errors.API_KEY_INACTIVE | The associated API key is no longer active. | 关联的 API Key 已停用。 | 2 | backend/src/middleware/auth.ts | ok |
| errors.API_KEY_LIMIT | You have reached the API key limit. Revoke unused keys before creating or rotating. | 已达到 API Key 数量上限，请先撤销未使用的 Key。 | 2 | backend/src/routes/developer/apiKeys.ts | ok |
| errors.API_KEY_NOT_FOUND | API key not found or does not belong to this developer | API 密钥不存在或不属于该开发者。 | 4 | backend/src/routes/developer/analytics.ts<br>backend/src/routes/developer/apiKeys.ts<br>backend/src/routes/developer/webhooks.ts | ok |
| errors.API_KEY_PROD_ONLY | Sandbox API keys are not allowed in production. | 生产环境不允许使用沙箱 API Key。 | 3 | backend/src/middleware/auth.ts<br>backend/src/routes/developer/apiKeys.ts | ok |
| errors.API_KEY_SANDBOX_ONLY | Production API keys are not allowed in development. | 开发环境不允许使用生产 API Key。 | 3 | backend/src/middleware/auth.ts<br>backend/src/routes/developer/apiKeys.ts | ok |
| errors.AUTHENTICATION_ERROR | Authentication failed | 认证失败 | 1 | backend/src/i18n/i18n.test.ts | ok |
| errors.AUTH_ACCESS_TOKEN_REQUIRED | Access token is required | 缺少访问令牌。 | 6 | backend/src/middleware/auth.ts | ok |
| errors.AUTH_API_KEY_EXPIRED | API key has expired | API Key 已过期。 | 3 | backend/src/middleware/auth.ts<br>backend/src/routes/handoff.ts | ok |
| errors.AUTH_CREDENTIAL_REQUIRED | API key, handoff token, or session token is required. | 缺少凭据：API Key、交接令牌或会话令牌至少需要一项。 | 1 | backend/src/middleware/auth.ts | ok |
| errors.AUTH_DEVELOPER_SUSPENDED | Developer account suspended | 开发者账户已被暂停。 | 4 | backend/src/middleware/auth.ts | ok |
| errors.AUTH_FAILED | Authentication failed | 认证失败。 | 9 | backend/src/middleware/auth.ts | ok |
| errors.AUTH_HANDOFF_EXPIRED | Handoff session has expired | 交接会话已过期。 | 1 | backend/src/middleware/auth.ts | ok |
| errors.AUTH_HANDOFF_INACTIVE | Handoff session is no longer active | 交接会话已失效。 | 2 | backend/src/middleware/auth.ts | ok |
| errors.AUTH_HANDOFF_INVALID | Invalid handoff token | 交接令牌无效。 | 2 | backend/src/middleware/auth.ts | ok |
| errors.AUTH_HANDOFF_REQUIRED | Handoff token is required. Include X-Handoff-Token header. | 缺少交接令牌，请携带 X-Handoff-Token 请求头。 | 1 | backend/src/middleware/auth.ts | ok |
| errors.AUTH_INVALID_2FA | Invalid two-factor authentication code. | 双因素验证码无效。 | 1 | backend/src/routes/auth.ts | ok |
| errors.AUTH_INVALID_API_KEY | Invalid API key | API Key 无效。 | 2 | backend/src/middleware/auth.ts<br>backend/src/routes/handoff.ts | ok |
| errors.AUTH_INVALID_CODE | Invalid or expired verification code. | 验证码无效或已过期。 | 2 | backend/src/routes/auth.ts | ok |
| errors.AUTH_INVALID_TOKEN | Invalid token. Please log in again! | 登录凭据无效，请重新登录。 | 10 | backend/src/middleware/auth.ts<br>backend/src/middleware/errorHandler.ts | ok |
| errors.AUTH_NO_REVIEWER_ACCOUNT | No active reviewer account was found for this email. | 该邮箱没有可用的审核员账号。 | 1 | backend/src/routes/auth.ts | ok |
| errors.AUTH_REQUIRED | Authentication required | 需要先登录。 | 2 | backend/src/middleware/auth.ts | ok |
| errors.AUTH_REQUIRED_API_KEY | API key is required. Include X-API-Key header. | 缺少 API Key，请携带 X-API-Key 请求头。 | 2 | backend/src/middleware/auth.ts | ok |
| errors.AUTH_SESSION_EXPIRED | Session token has expired | 会话令牌已过期。 | 2 | backend/src/middleware/auth.ts<br>backend/src/routes/handoff.ts | ok |
| errors.AUTH_SESSION_INVALID | Invalid session token | 会话令牌无效。 | 4 | backend/src/middleware/auth.ts<br>backend/src/routes/handoff.ts | ok |
| errors.AUTH_SESSION_REQUIRED | Session token is required. Include X-Session-Token header. | 缺少会话令牌，请携带 X-Session-Token 请求头。 | 1 | backend/src/middleware/auth.ts | ok |
| errors.AUTH_TOKEN_EXPIRED | Your token has expired. Please log in again. | 登录已过期，请重新登录。 | 5 | backend/src/middleware/auth.ts<br>backend/src/middleware/errorHandler.ts | ok |
| errors.BATCH_NOT_FOUND | Batch job not found. | 批处理任务不存在。 | 3 | backend/src/routes/batch.ts | ok |
| errors.DAILY_LIMIT_EXCEEDED | Daily verification limit exceeded. | 已超出每日验证次数上限。 | 1 | backend/src/middleware/rateLimit.ts | ok |
| errors.DELIVERY_NOT_FOUND | Delivery not found | 投递记录不存在。 | 1 | backend/src/routes/developer/webhooks.ts | ok |
| errors.DEV_AUTH_REQUIRED | Developer authentication required | 需要开发者身份。 | 13 | backend/src/routes/compliance.ts<br>backend/src/routes/developer/apiKeys.ts<br>backend/src/routes/developer/profile.ts<br>…(4) | ok |
| errors.DOCUMENT_REQUIRED | Document file is required | 请上传证件文件。 | 3 | backend/src/routes/newVerification.ts | ok |
| errors.DOES_NOT_EXIST |  |  | 3 | backend/src/i18n/i18n.test.ts | 缺少 en |
| errors.DUPLICATE_FIELD | Duplicate value | 数据已存在 | 2 | backend/src/middleware/errorHandler.ts | ok |
| errors.EXTERNAL_API_ERROR | An external service is temporarily unavailable | 外部服务暂时不可用 | 0 | - | 未引用 |
| errors.FACE_RECOGNITION_ERROR | We could not process the face image | 无法处理人脸图片 | 0 | - | 未引用 |
| errors.FILE_INVALID_IMAGE | File content does not match a valid JPEG or PNG image | 文件内容不是有效的 JPEG 或 PNG 图片。 | 6 | backend/src/routes/addressVerification.ts<br>backend/src/routes/developer/profile.ts<br>backend/src/routes/newVerification.ts | ok |
| errors.FILE_REQUIRED | No file uploaded | 未上传文件。 | 3 | backend/src/routes/addressVerification.ts<br>backend/src/routes/developer/profile.ts | ok |
| errors.FILE_TOO_LARGE | File too large. Maximum size is 10MB. | 文件过大，最大 10MB。 | 1 | backend/src/middleware/errorHandler.ts | ok |
| errors.FILE_UNEXPECTED_FIELD | Unexpected file field. | 文件字段不正确。 | 1 | backend/src/middleware/errorHandler.ts | ok |
| errors.FILE_UPLOAD_ERROR | File upload failed | 文件上传失败 | 0 | - | 未引用 |
| errors.FORBIDDEN | Not authorized to access this resource | 无权访问该资源 | 0 | - | 未引用 |
| errors.HANDOFF_FIELDS_REQUIRED | api_key and user_id are required. | api_key 与 user_id 均为必填。 | 1 | backend/src/routes/handoff.ts | ok |
| errors.HANDOFF_INVALID_RESULT | The result payload is invalid or too large. | 结果载荷无效或过大。 | 2 | backend/src/routes/handoff.ts | ok |
| errors.HANDOFF_INVALID_STATUS | status must be completed or failed. | status 必须为 completed 或 failed。 | 1 | backend/src/routes/handoff.ts | ok |
| errors.HANDOFF_INVALID_VERIFICATION_ID | Invalid verification ID format. | 验证 ID 格式无效。 | 3 | backend/src/routes/handoff.ts | ok |
| errors.HANDOFF_NO_API_KEY | No API key is associated with this session. | 该会话未关联 API Key。 | 1 | backend/src/routes/handoff.ts | ok |
| errors.INTERNAL_ERROR | Something went wrong | 服务暂时不可用 | 1 | backend/src/middleware/errorHandler.ts | ok |
| errors.INVALID_CREDENTIALS | Invalid credentials. | 凭据无效。 | 2 | backend/src/routes/auth.ts | ok |
| errors.LIVENESS_METADATA_INVALID | Invalid liveness_metadata: expected head_turn challenge format with a frames array. | liveness_metadata 无效：应为包含 frames 数组的 head_turn 挑战格式。 | 1 | backend/src/routes/newVerification.ts | ok |
| errors.LLM_API_KEY_REQUIRED | API key is required when configuring a provider | 配置提供方时必须填写 API 密钥。 | 1 | backend/src/routes/developer/settings.ts | ok |
| errors.LLM_ENDPOINT_REQUIRED | Custom provider requires an endpoint URL | 自定义提供方需要填写端点地址。 | 1 | backend/src/routes/developer/settings.ts | ok |
| errors.LLM_MODEL_REQUIRED | Custom provider requires a model name | 自定义提供方需要填写模型名称。 | 1 | backend/src/routes/developer/settings.ts | ok |
| errors.NEW_STATUS_REQUIRED | new_status is required for override decisions | 覆盖决定必须提供 new_status。 | 1 | backend/src/routes/admin.ts | ok |
| errors.NOT_FOUND | Resource not found | 资源不存在 | 0 | - | 未引用 |
| errors.OAUTH_AUTH_FAILED | GitHub authentication failed. | GitHub 登录失败。 | 1 | backend/src/services/githubOAuthService.ts | ok |
| errors.OAUTH_NOT_CONFIGURED | GitHub authentication is not configured. | 未配置 GitHub 登录。 | 1 | backend/src/routes/auth.ts | ok |
| errors.OAUTH_NO_EMAIL | No verified email was found on your GitHub account. | GitHub 账号未提供已验证邮箱。 | 2 | backend/src/routes/auth.ts<br>backend/src/services/githubOAuthService.ts | ok |
| errors.OAUTH_STATE_INVALID | Invalid or expired OAuth state. | OAuth 状态无效或已过期。 | 1 | backend/src/routes/auth.ts | ok |
| errors.OAUTH_TOKEN_EXCHANGE_FAILED | GitHub token exchange failed. | GitHub 令牌交换失败。 | 1 | backend/src/services/githubOAuthService.ts | ok |
| errors.OAUTH_USER_FETCH_FAILED | Failed to fetch the GitHub user profile. | 获取 GitHub 用户资料失败。 | 1 | backend/src/services/githubOAuthService.ts | ok |
| errors.OCR_PROCESSING_ERROR | We could not process the document image | 无法处理证件图片 | 0 | - | 未引用 |
| errors.ONLY_ORG_ADMINS_OVERRIDE | Only organization admins can override verification status | 只有组织管理员可以覆盖验证状态。 | 1 | backend/src/routes/admin.ts | ok |
| errors.OPERATOR_NOT_BOUND | Operator is no longer bound to this service key. | 运营账号已不再绑定该服务 Key。 | 1 | backend/src/middleware/auth.ts | ok |
| errors.ORG_ADMIN_REQUIRED | Organization admin privileges required | 需要组织管理员权限。 | 1 | backend/src/middleware/auth.ts | ok |
| errors.PROFILE_EMAIL_MISMATCH | Email does not match your account email. | 邮箱与账号邮箱不一致。 | 1 | backend/src/routes/developer/profile.ts | ok |
| errors.RATE_LIMIT_EXCEEDED | Too many requests. Please try again later. | 请求过于频繁，请稍后重试。 | 2 | backend/src/middleware/rateLimit.ts | ok |
| errors.REGISTRATION_TOKEN_INVALID | Registration token is invalid or expired. Please start over. | 注册令牌无效或已过期，请重新开始。 | 2 | backend/src/middleware/auth.ts<br>backend/src/routes/auth.ts | ok |
| errors.REVERIFY_NESTED | Cannot re-verify from another re-verification — use the original verification | 不能从另一次再验证发起再验证，请使用原始验证。 | 1 | backend/src/routes/newVerification.ts | ok |
| errors.REVERIFY_NOT_FOUND | Previous verification not found | 上一次验证不存在。 | 1 | backend/src/routes/newVerification.ts | ok |
| errors.REVERIFY_NOT_VERIFIED | Previous verification must be verified before re-verification | 只有已验证的记录才能再次验证。 | 1 | backend/src/routes/newVerification.ts | ok |
| errors.REVERIFY_NO_FRONT | Previous verification has no front extraction data — cannot re-verify | 上一次验证缺少正面提取数据，无法再验证。 | 1 | backend/src/routes/newVerification.ts | ok |
| errors.REVERIFY_OTHER_DEVELOPER | Previous verification belongs to a different developer | 上一次验证属于其他开发者。 | 1 | backend/src/routes/newVerification.ts | ok |
| errors.REVERIFY_OTHER_USER | Previous verification belongs to a different user | 上一次验证属于其他用户。 | 1 | backend/src/routes/newVerification.ts | ok |
| errors.REVIEWER_EXISTS | Reviewer with this email already exists | 该邮箱的审核员已存在。 | 1 | backend/src/routes/developer/profile.ts | ok |
| errors.REVIEWER_NOT_FOUND | Reviewer not found | 审核员不存在。 | 2 | backend/src/middleware/auth.ts<br>backend/src/routes/developer/profile.ts | ok |
| errors.REVIEWER_TOKEN_INVALID | Invalid reviewer token. | 审核员令牌无效。 | 1 | backend/src/middleware/auth.ts | ok |
| errors.SCHEDULE_NOT_FOUND | Schedule not found. | 监控计划不存在。 | 2 | backend/src/routes/monitoring.ts | ok |
| errors.SELECTION_TOKEN_EXPIRED | Selection token has expired. | 选择令牌已过期。 | 1 | backend/src/middleware/auth.ts | ok |
| errors.SELECTION_TOKEN_INVALID | Invalid selection token. | 选择令牌无效。 | 2 | backend/src/middleware/auth.ts | ok |
| errors.SERVICE_AUTH_NOT_CONFIGURED | Service authentication is not configured. | 未配置服务鉴权。 | 1 | backend/src/middleware/auth.ts | ok |
| errors.SERVICE_OPERATOR_TOKEN_INVALID | Invalid service operator token. | 服务运营令牌无效。 | 1 | backend/src/middleware/auth.ts | ok |
| errors.SERVICE_TOKEN_INVALID | Invalid service token. | 服务令牌无效。 | 1 | backend/src/middleware/auth.ts | ok |
| errors.SERVICE_TOKEN_REQUIRED | Service token is required. Include the X-Service-Token header. | 需要服务令牌，请提供 X-Service-Token 请求头。 | 1 | backend/src/middleware/auth.ts | ok |
| errors.SESSION_FLOW_ERROR | This step is not allowed right now | 当前步骤不允许该操作 | 1 | backend/src/middleware/errorHandler.ts | ok |
| errors.SETUP_COMPLETED | Setup already completed. Use the normal sign-in flow. | 安装已完成，请使用正常登录流程。 | 2 | backend/src/routes/setup.ts | ok |
| errors.SMS_FIELDS_REQUIRED | API key, API secret, and phone number are all required | API 密钥、API Secret 和手机号均为必填。 | 1 | backend/src/routes/developer/settings.ts | ok |
| errors.TOTP_INVALID | Invalid two-factor authentication token. | 双因素令牌无效。 | 1 | backend/src/routes/auth.ts | ok |
| errors.TOTP_NOT_STARTED | Two-factor setup has not been started. | 尚未开始双因素设置。 | 1 | backend/src/routes/auth.ts | ok |
| errors.USER_ID_REQUIRED | User ID is required. | 用户 ID 为必填。 | 1 | backend/src/middleware/auth.ts | ok |
| errors.USER_NOT_FOUND | User not found in your verification data | 在该开发者的验证数据中未找到用户。 | 1 | backend/src/routes/admin.ts | ok |
| errors.VALIDATION_ERROR | Validation failed | 请求参数校验失败 | 5 | backend/src/middleware/errorHandler.ts<br>backend/src/routes/developer/analytics.ts | ok |
| errors.VERIFY_NOT_FOUND | Verification request not found | 验证请求不存在。 | 7 | backend/src/routes/admin.ts<br>backend/src/routes/developer/analytics.ts<br>backend/src/routes/newVerification.ts | ok |
| errors.VERIFY_TOKEN_UNAUTHORIZED | Session token is not authorized for this verification | 会话令牌无权访问该验证。 | 1 | backend/src/routes/newVerification.ts | ok |
| errors.VOICE_AUDIO_REQUIRED | Voice audio file is required | 请上传语音文件。 | 1 | backend/src/routes/newVerification.ts | ok |
| errors.WEBHOOK_DUPLICATE | Webhook already exists for this URL and scope | 该 URL 与作用域下的 Webhook 已存在。 | 1 | backend/src/routes/developer/webhooks.ts | ok |
| errors.WEBHOOK_DUPLICATE_URL | Webhook already exists for this URL | 该 URL 的 Webhook 已存在。 | 1 | backend/src/routes/webhooks.ts | ok |
| errors.WEBHOOK_INVALID_EVENTS | Invalid webhook events | Webhook 事件无效。 | 1 | backend/src/routes/developer/webhooks.ts | ok |
| errors.WEBHOOK_NOT_FOUND | Webhook not found | Webhook 不存在。 | 9 | backend/src/routes/developer/webhooks.ts<br>backend/src/routes/webhooks.ts | ok |
| errors.WEBHOOK_URL_INVALID | The webhook URL is invalid. | Webhook URL 无效。 | 2 | backend/src/routes/developer/webhooks.ts<br>backend/src/routes/webhooks.ts | ok |
| verification.failedDefault | Verification failed. Please try again with clearer documents. | 验证失败，请使用更清晰的证件照片重试。 | 1 | backend/src/i18n/index.ts | ok |
| verification.manualReview.compliance_force_review | Your verification needs additional review. We will notify you of the outcome. | 你的验证需要进一步审核，我们会通知你最终结果。 | 0 | - | 未引用 |
| verification.manualReview.cross_validation_review | Your document needs additional review. We will notify you of the outcome. | 你的证件需要进一步审核，我们会通知你最终结果。 | 0 | - | 未引用 |
| verification.manualReview.default | Your verification requires additional review. You will be notified of the outcome. | 你的验证需要进一步审核，我们会通知你最终结果。 | 2 | backend/src/i18n/index.ts | ok |
| verification.manualReview.duplicate_detected | Your verification needs additional review. We will notify you of the outcome. | 你的验证需要进一步审核，我们会通知你最终结果。 | 0 | - | 未引用 |
| verification.manualReview.face_match_skipped | Your verification needs additional review. We will notify you of the outcome. | 你的验证需要进一步审核，我们会通知你最终结果。 | 0 | - | 未引用 |
| verification.manualReview.geo_flags | Your verification needs additional review. We will notify you of the outcome. | 你的验证需要进一步审核，我们会通知你最终结果。 | 0 | - | 未引用 |
| verification.manualReview.velocity_flags | Your verification needs additional review. We will notify you of the outcome. | 你的验证需要进一步审核，我们会通知你最终结果。 | 0 | - | 未引用 |
| verification.rejection.AML_MATCH_FOUND | Verification could not be completed. Please contact support for assistance. | 无法完成验证，请联系客服协助处理。 | 0 | - | 未引用 |
| verification.rejection.AML_POTENTIAL_MATCH | Your verification requires additional review. You will be notified of the outcome. | 你的验证需要进一步审核，我们会通知你最终结果。 | 0 | - | 未引用 |
| verification.rejection.BACK_BARCODE_NOT_FOUND | We could not read the barcode on the back of your document. Please retake the back photo. | 无法读取证件背面的条码，请重新拍摄背面。 | 0 | - | 未引用 |
| verification.rejection.BACK_MRZ_BARCODE_MISMATCH | The information on the front and back of your document does not match. | 证件正反面信息不一致。 | 0 | - | 未引用 |
| verification.rejection.BACK_MRZ_CHECKSUM_FAILED | The machine-readable zone could not be validated. Please retake the back photo. | 证件机读区校验失败，请重新拍摄背面。 | 0 | - | 未引用 |
| verification.rejection.CROSS_VALIDATION_FAILED | The information on the front and back of your document does not match. Please retake both sides. | 证件正反面信息不一致，请重新拍摄两面。 | 0 | - | 未引用 |
| verification.rejection.DEEPFAKE_DETECTED | We could not verify that you are present. Please try again. | 无法确认你在现场，请重试。 | 0 | - | 未引用 |
| verification.rejection.DOB_NOT_FOUND | We could not read the date of birth from your document. Please retake the front photo. | 无法从证件读取出生日期，请重新拍摄正面。 | 0 | - | 未引用 |
| verification.rejection.DOCUMENT_EXPIRED | Your document has expired. Please use a valid, non-expired ID. | 证件已过期，请使用有效期内的证件。 | 0 | - | 未引用 |
| verification.rejection.DOCUMENT_TAMPERED | We could not accept this document. Please contact support. | 无法接受该证件，请联系客服。 | 0 | - | 未引用 |
| verification.rejection.FACE_MATCH_FAILED | Your live capture does not match the photo on your ID. Please try again. | 实时采集与证件照片不匹配，请重试。 | 0 | - | 未引用 |
| verification.rejection.FACE_NOT_DETECTED | We could not detect your face. Please ensure your face is clearly visible and well-lit. | 未检测到人脸，请确保面部清晰、光线充足。 | 0 | - | 未引用 |
| verification.rejection.FRONT_LOW_CONFIDENCE | The document image is too blurry or low quality. Please retake the photo. | 证件图片模糊或质量不足，请重新拍摄。 | 0 | - | 未引用 |
| verification.rejection.FRONT_OCR_FAILED | We could not read your document. Please retake the photo with good lighting. | 无法读取你的证件，请在光线充足时重新拍摄。 | 0 | - | 未引用 |
| verification.rejection.LIVENESS_FAILED | We could not verify that you are present. Please try again with a live photo, not a printed picture or screen. | 无法确认你在现场，请使用实时拍摄，不要使用照片或屏幕翻拍。 | 0 | - | 未引用 |
| verification.rejection.UNDERAGE | You do not meet the minimum age requirement for this verification. | 你不满足本次验证的最低年龄要求。 | 0 | - | 未引用 |
| verification.rejection.VOICE_CHALLENGE_FAILED | The spoken digits did not match the challenge. Please try again. | 朗读数字与挑战不一致，请重试。 | 0 | - | 未引用 |
| verification.rejection.VOICE_MATCH_FAILED | Voice verification failed. Please try again. | 声纹验证失败，请重试。 | 0 | - | 未引用 |

## 源码引用但未定义

- errors.DOES_NOT_EXIST

## 缺失翻译

- errors.DOES_NOT_EXIST（缺少 en）
- errors.DOES_NOT_EXIST（缺少 zh-CN）

## 说明

- 代码中的 `messageKey` 必须同时存在于 en 与 zh-CN；缺失项需在 M1 补齐。
- 未引用键可能是动态拼接或历史遗留，需人工确认后删除或保留。
- 本表由 `contracts/tools/extract-i18n.mjs` 生成，可复算。
