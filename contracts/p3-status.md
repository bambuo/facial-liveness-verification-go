# P3 状态报告（2026-11-19）

## 结论

P3 前端已切换为**直接复用参考 React 实现**：22 个页面、24 条必需路由全部保留，UI 交互组件已统一迁移到 Ant Design 6，源码已去品牌，并通过同源 `/api` 代理接入 Go 服务。前端已完成 React 19 语法现代化、Bun/Vite 8 工具链升级与历史依赖清理；Go 侧补齐开发者/审核员登录 Cookie、CSRF 字段、管理端统计与列表字段别名；React 路由检查 24/24 无控制台错误。

## 交付

| 项 | 结果 | 证据 |
| --- | --- | --- |
| 技术栈 | React 19 + Vite 8 + TypeScript 6 + Ant Design 6 + Tailwind 4 + Bun 1.4 | `frontend/package.json`、`frontend/bun.lock` |
| 页面 | 22 个参考页面组件 | `frontend/src/pages` |
| 路由 | 24 必需（含 `/verify`、`/admin/*` 重定向） | `frontend/src/App.tsx` |
| 源码复用 | 参考 React 实现直接迁入，不再维护 Vue 平行实现 | `frontend/src` |
| 去品牌 | UI、i18n、元信息、Logo 与 API 文档中的参考品牌已替换 | `frontend/src`、`frontend/public`、`frontend/index.html` |
| 会话兼容 | 开发者/审核员 OTP 同时签发 HttpOnly Cookie 与 Bearer token | `internal/adapter/web/auth_idswyft.go`、`admin.go` |
| 管理端适配 | 概览统计折叠为 total/pending/verified/failed/manual_review；列表补 `id` 与状态分类 | `internal/adapter/web/admin.go` |
| 文档 | `/api/docs/markdown` 返回去品牌 Markdown | `internal/adapter/web/docs_markdown.go` |
| OpenCV | 本地 `public/vendor/opencv.js`，不依赖外部 CDN | `frontend/public/vendor/opencv.js` |
| 路由检查 | 24/24 可达且无控制台错误 | `bun run check:routes` |
| antd 规范 | 0 deprecated / 0 usage / 0 a11y 告警 | `bunx antd lint ./src` |
| 语法现代化 | 0 `React.FC`、0 `forwardRef`、0 默认 React 导入、0 浏览器 `process.env` | `grep` 审计 |
| 单元测试 | 8 文件 / 49 用例通过 | `npx vitest run` |

## 验证命令

```bash
cd frontend
bun install
bun run type-check
bun run build
bun run test
bunx vitest run
bun run lint
bunx antd lint ./src
FRONTEND_URL=http://localhost:5173 bun run check:routes
```

本地联调：`VITE_DEV_PORT=5180 VITE_API_TARGET=http://localhost:8090 bun run dev`。默认 `5173`，`/api` 代理到 `localhost:8080`。

## 边界

0. 相机、OCR、语音与真实证件链路需要后端模型与样本，当前只验证页面与鉴权骨架。
1. 参考的演示旁白/视频生成脚本保留，但未重新生成媒体资产。
2. 生产静态托管与 Go API 同源部署尚未接入，当前通过 Vite 代理联调。
3. 管理端部分证据字段依赖存储与后台返回；存储未启用时自拍预签名 URL 不返回。
