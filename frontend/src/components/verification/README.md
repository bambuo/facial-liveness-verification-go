# 终端用户验证组件

一个生产可用、可嵌入的验证组件，开发者可以轻松集成到自己的应用中。

## 功能

- ✅ **完整验证流程**——证件上传 → OCR 处理 → 实时采集 → 结果
- ✅ **干净的终端用户体验**——不展示开发者工具、API Key 或原始 JSON
- ✅ **自动跳转**——可配置成功/失败跳转
- ✅ **主题支持**——亮色与暗色主题
- ✅ **移动端响应式**——在所有设备上正常显示
- ✅ **TypeScript 支持**——完整类型安全
- ✅ **可定制**——通过 props 配置与回调

## 基本用法

```tsx
import EndUserVerification from './components/verification/EndUserVerification';

function MyApp() {
  return (
    <EndUserVerification
      apiKey="your-api-key"
      userId="user-123"
      redirectUrl="https://yourapp.com/dashboard"
      onComplete={(result) => {
        console.log('验证完成：', result);
      }}
    />
  );
}
```

## 属性

| Prop | 类型 | 必填 | 描述 |
|------|------|----------|-------------|
| `apiKey` | string | ✅ | 你的 Identity Verification API Key |
| `userId` | string | ✅ | 用户唯一标识 |
| `onComplete` | function | ❌ | 验证完成回调 |
| `onRedirect` | function | ❌ | 自定义跳转处理器 |
| `redirectUrl` | string | ❌ | 完成后跳转的 URL |
| `className` | string | ❌ | 附加 CSS 类 |
| `theme` | 'light' \| 'dark' | ❌ | 主题（默认：'light'） |
| `allowedDocumentTypes` | array | ❌ | 允许的证件类型 |

## 高级用法

### 自定义完成处理器

```tsx
<EndUserVerification
  apiKey="your-api-key"
  userId="user-123"
  onComplete={(result) => {
    if (result.status === 'verified') {
      // 在你的数据库中更新用户状态
      updateUserVerificationStatus(result.user_id, 'verified');
      
      // 显示成功消息
      showSuccessNotification();
    } else {
      // 处理验证失败
      handleVerificationFailure(result);
    }
  }}
  onRedirect={(url) => {
    // 自定义跳转逻辑
    window.location.href = url;
  }}
/>
```

### 使用暗色主题

```tsx
<EndUserVerification
  apiKey="your-api-key"
  userId="user-123"
  theme="dark"
  className="my-custom-class"
/>
```

### 限制证件类型

```tsx
<EndUserVerification
  apiKey="your-api-key"
  userId="user-123"
  allowedDocumentTypes={['passport', 'drivers_license']}
/>
```

## 验证结果

`onComplete` 回调接收 `VerificationResult` 对象：

```typescript
interface VerificationResult {
  verification_id: string;
  status: 'verified' | 'failed' | 'manual_review';
  user_id: string;
  confidence_score?: number;
  face_match_score?: number;
  liveness_score?: number;
}
```

## 集成步骤

1. **安装依赖**（如尚未安装）：
   ```bash
   npm install react react-router-dom react-hot-toast
   ```

2. **复制组件**：
   - 将 `EndUserVerification.tsx` 复制到你的项目
   - 确保包含 API 配置文件

3. **在应用中使用**：
   ```tsx
   import EndUserVerification from './path/to/EndUserVerification';
   
   // 在你的组件中
   <EndUserVerification
     apiKey={import.meta.env.VITE_API_KEY}
     userId={currentUser.id}
     redirectUrl="/dashboard"
   />
   ```

## 样式

组件使用 Tailwind CSS 类。你可以：

1. **直接使用**（项目中已有 Tailwind CSS）
2. **自定义主题**（修改 `themeClasses` 对象）
3. **覆盖样式**（使用 `className` prop）
4. **创建自定义主题**（扩展主题系统）

## 流程控制

组件会自动：

1. 挂载时**启动验证**
2. 内部**处理所有 API 调用**
3. **管理步骤间的状态迁移**
4. 处理期间**轮询完成状态**
5. 完成后**跳转**（如已配置）

## 错误处理

- 展示用户友好的错误消息
- 自动重试失败请求
- 为网络问题提供回退状态
- 记录错误用于调试（开发环境）

## 安全说明

- API Key 应安全存储（环境变量）
- 用户 ID 应正确校验
- 建议在你的侧实现限流
- 始终在你的后端校验验证结果

## 支持

如对验证组件有疑问或问题：

1. 查看 [API 文档](../docs/API.md)
2. 浏览[开发者门户](http://localhost:5173/developer)
3. 先使用[演示页](http://localhost:5173/verify)测试
