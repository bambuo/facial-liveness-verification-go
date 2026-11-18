# P0 状态报告（2026-11-18）

## 结论

- **软件可实现项已完成**：鉴权签字、契约 fixtures、DDL/重放、10 个原子函数、嵌套响应、错误码/i18n、字段规则 v1、进程级基准。
- **两项非软件项**：
  1. `.github` 工作流按用户要求禁用（`ci.yml.disabled`），CI 门禁未启用。
  2. P0 第 6 项已补公开 MIDV-500 模拟护照 30 张（中文 mobile rec 字段 75.8%、姓名 53.3%）；真实证件 20–50 张仍**等待脱敏数据**，真实场景精度门禁未完成。

## 逐项状态

| # | 项 | 状态 | 证据 |
| --- | --- | --- | --- |
| 1 | 17 条公开候选鉴权签字 | 完成 | `contracts/auth-decisions.md`；`auth.json` 17 条 |
| 2 | 契约测试骨架 | 完成骨架 | `contracts/fixtures.json` 158 条（含嵌套 44）；2 个引擎端点 shape 测试；后端实现 0/158 |
| 3 | 31 表 DDL + 空库重放 | 完成 | `contracts/ddl/001_baseline.sql`；31/329；GORM parity 通过 |
| 3b | CI 门禁 | 已禁用（用户要求） | `.github/workflows/ci.yml.disabled` |
| 4 | 10 个 SQL 函数 Go 等价 | 完成 | `internal/store` GORM 事务；真实 PG 测试通过 |
| 5 | 嵌套响应 + 错误码/i18n | 完成 | `responses.md` 嵌套字段；`contracts/errors-i18n.md` 142/142 |
| 6 | 字段规则 v1 | 完成 | `internal/document/fields`：MRZ/US/UK/CN/BR + 单测 |
| 6b | 公开模拟证件 20–50 张 | 完成 | MIDV-500 巴西护照 30 张；中文 mobile 校正后字段 75.8%、姓名 53.3%，见 `eval/ocr/midv-bra-zh.md` |
| 6c | 真实证件 20–50 张 + 精度门禁 | **阻塞** | 需要脱敏真实样本目录 |
| 7 | OCR 性能基线 + 进程级 | 完成 | `eval/ocr/bench.md`：1/2/4 进程 QPS 0.239/0.235/0.228 |

## 验证命令与结果

```bash
CGO_ENABLED=0 go build ./...                       # 通过
go test -count=1 ./...                             # 通过
TEST_DATABASE_URL=... go test ./internal/store      # 通过
TEST_GORM_URL=... TEST_BASELINE_URL=... \
  go test ./internal/adapter/pg -run 'TestAutoMigrateSmoke|TestSchemaParity'  # 通过
TEST_GORM_CONCURRENT_URL=... go test ./internal/adapter/pg -run TestAutoMigrateConcurrent  # 通过
go run ./cmd/contractcheck                         # 158 fixtures / 0 后端实现 / 2 引擎
check-docs.mjs --require-flow-diagrams             # 15 文件 0 失败
go test ./internal/document/fields                 # 通过
(cd third_party/onnx-gomlx && go test ./internal/onnxgomlx -run TestConvAutoPad)  # 通过
go run ./cmd/docxeval -rec .../zh_rec.onnx -dict .../zh_dict.txt \
  -images .../rect2/TS/*.png -gt .../06_bra_passport.json \
  -max-side 2000 -width-mode bucket -bucket 64     # 字段 75.8%
```

## 剩余

1. 提供 20–50 张脱敏证件样本后，运行 `cmd/docxeval`，产出字段级准确率与 P95 门禁；若有人工全文标注再补 CER/WER。
2. 若需要 CI，另行指定非 GitHub Actions 的运行方式或重新启用工作流。

## 样本目录核对（2026-11-18）

`/Users/johana/Desktop/facial-liveness-eval` 原始内容为 LFW/CFP/Axon 人脸与 PAD 评测归档，不是证件 OCR 样本：

- README 明确为“LFW 评测归档”，内容为 `lfw.zip`、`lfw-200.json`、`axon-face-anti-spoofing.tar.gz`。
- 目录为 `lfw-deepfunneled/`、`cfp-dataset/`、`axon-pad-eval/`、`auraface/`、`idswyft-face-benchmark/`。
- 原始归档无 passport/license/idcard/document/mrz 图片或字段标注。

该目录可用于人脸 1:1 与 PAD 基线。证件 OCR 评测已在其下补充公开模拟数据：

- `midv500/bra_ts/06_bra_passport/`：MIDV-500 巴西护照 30 张 TS 帧；原始帧字段 51.7%，GT 四边形透视校正 + 中文 mobile rec 后 75.8%。
- 数据来源与许可见 `eval/ocr/README.md`；压缩包与图片不入库。
- MIDV-500 是模拟证件，不能替代真实脱敏样本；P0 第 6c 项仍阻塞。
