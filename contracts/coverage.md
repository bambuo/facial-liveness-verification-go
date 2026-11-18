# 契约覆盖报告

由 `go run ./cmd/contractcheck` 生成。

## 冻结契约

| 项 | 数值 |
| --- | --- |
| 接口注册 | 154 |
| 可达路径 | 158 |
| 数据表 / 列 | 31 / 329 |
| Zod schema | 16 |
| 契约 fixtures | 158（含嵌套响应 57） |
| 公开候选签字 | 17 |

## 实现覆盖

后端接口：**0 / 158** 可达路径（154 注册）

引擎端点：**2 / 6**

### 已实现引擎端点

| 方法 | 路径 |
| --- | --- |
| POST | `/extract/front` |
| POST | `/extract/ocr` |

## 待实现模块

| 模块 | 接口数 |
| --- | --- |
| addressVerification | 2 |
| admin | 15 |
| admin-thresholds | 4 |
| auth | 13 |
| batch | 5 |
| compliance | 9 |
| credentials | 8 |
| developer/analytics | 4 |
| developer/apiKeys | 4 |
| developer/profile | 8 |
| developer/register | 1 |
| developer/settings | 19 |
| developer/webhooks | 7 |
| handoff | 5 |
| health | 5 |
| monitoring | 5 |
| newVerification | 15 |
| pageConfig | 3 |
| setup | 2 |
| status | 1 |
| system | 1 |
| vaas | 8 |
| vault | 7 |
| webhooks | 6 |
| well-known | 1 |

## 校验结果

全部通过。
