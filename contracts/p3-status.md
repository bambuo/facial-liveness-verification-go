# P3 状态报告（2026-11-19）

## 结论

P3 前端代码基线已落地，但**参考页面 1:1 复刻未完成**。Vue3 + Arco 独立前端已落地，22 个页面组件、24 条必需路由（另有 2 条扩展路由），并通过类型检查、生产构建、路由可达、鉴权 UI 冒烟与 24 路由截图回归。

## 交付

| 项 | 结果 | 证据 |
| --- | --- | --- |
| 技术栈 | Vue 3 + Vite + Arco + Pinia + Vue Router | `frontend/package.json` |
| 页面 | 22 个视图组件 | `frontend/src/views` |
| 路由 | 24 必需 + `/admin/thresholds`、`/operator/login` | `frontend/src/router/index.ts` |
| 用户采集 | 初始化、正/反面、活体相机、语音挑战/WAV、finalize | `UserVerificationPage.vue`、`LiveCapturePage.vue` |
| 移动交接 | handoff token 会话读取与继续 | `MobileVerificationPage.vue` |
| 管理端 | 验证列表/详情/审核、开发者、阈值 | `VerificationManagement.vue` 等 |
| 开发者端 | API Key、统计、Page Builder | `DeveloperPage.vue`、`PageBuilderPage.vue` |
| 凭证/文档/初始化/法务 | 已实现 | 对应 views |
| 路由检查 | 26/26 可达且无 console error | `npm run check:routes` |
| 鉴权 UI 冒烟 | 5/5 页面通过 mock API | `npm run check:ui` |
| 截图回归 | 24/24 基线 0% diff | `npm run check:visual` |

## 验证命令

```bash
cd frontend
npm run type-check
npm run build
npm run check:routes
npm run check:ui
npm run check:visual
```

本地联调：`npm run dev`，默认 `127.0.0.1:5173`，`/api` 代理到 `localhost:8080`。

## 边界

0. 当前 22 页为功能基线，不等于参考页面 1:1；逐页对齐时统一使用 Vue3+Arco 标准组件。

1. 页面是 P3 代码基线；视觉基线是本实现自身的回归基线，不等于参考页面逐像素截图对齐。
2. 生产静态托管与 Go API 同源部署尚未接入，当前通过 Vite 代理联调。
3. 相机需要 HTTPS 或 localhost；音频上传当前接受 WAV。
4. 参考页面中的长文案/演示视频/Remotion 内容未逐条复制。
