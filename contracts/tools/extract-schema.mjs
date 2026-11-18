#!/usr/bin/env node
/**
 * 从 idswyft supabase/migrations 与 Go 基座 DDL 抽取数据库契约，生成差异清单。
 *
 * 用法:
 *   node contracts/tools/extract-schema.mjs [--repo <idswyft-repo>] [--out <contracts-dir>]
 *
 * 输出:
 *   contracts/schema.json       idswyft 表/函数/触发器/策略/索引
 *   contracts/go-schema.json    Go 基座表结构
 *   contracts/schema-diff.md    差异清单（契约冻结输入）
 */
import fs from 'node:fs';
import path from 'node:path';
import { execFileSync } from 'node:child_process';
import {
  maskSql,
  splitStatements,
  parseStatement,
} from './sql-parse.mjs';

const args = process.argv.slice(2);
function argValue(name, fallback) {
  const i = args.indexOf(name);
  return i >= 0 && args[i + 1] ? args[i + 1] : fallback;
}

const REPO = path.resolve(argValue('--repo', '/Users/johana/Codes/github/idswyft-community-1.12.39'));
const GO_REPO = path.resolve(argValue('--go-repo', '/Users/johana/Desktop/facial-liveness-verification-go'));
const OUT_DIR = path.resolve(argValue('--out', path.join(process.cwd(), 'contracts')));
const MIGRATIONS_DIR = path.join(REPO, 'supabase/migrations');
const GO_SQL_DIR = path.join(GO_REPO, 'docs/需求/0001/sql');

const EXPECTED_TABLES = 31;

function listSqlFiles(dir) {
  if (!fs.existsSync(dir)) return [];
  return fs
    .readdirSync(dir)
    .filter((f) => f.endsWith('.sql'))
    .sort((a, b) => a.localeCompare(b, 'en', { numeric: true }));
}

function emptyModel() {
  return {
    tables: new Map(),
    functions: new Map(),
    triggers: new Map(),
    policies: [],
    indexes: new Map(),
    enums: new Map(),
    extensions: new Set(),
    doBlocks: [],
    otherStatements: [],
  };
}

function ensureTable(model, name, schema, file) {
  if (!model.tables.has(name)) {
    model.tables.set(name, {
      name,
      schema: schema || 'public',
      columns: [],
      constraints: [],
      rls: false,
      indexes: [],
      policies: [],
      triggers: [],
      otherOps: [],
      sources: [file],
    });
  } else if (!model.tables.get(name).sources.includes(file)) {
    model.tables.get(name).sources.push(file);
  }
  return model.tables.get(name);
}

function apply(model, parsed, file) {
  const { kind, value } = parsed;
  if (!value) {
    if (kind !== 'other') model.otherStatements.push({ file, kind, sql: parsed.sql });
    return;
  }
  switch (kind) {
    case 'table': {
      const table = ensureTable(model, value.name, value.schema, file);
      for (const col of value.columns) {
        const existing = table.columns.findIndex((c) => c.name === col.name);
        const record = { ...col, addedIn: file };
        if (existing >= 0) table.columns[existing] = record;
        else table.columns.push(record);
      }
      for (const c of value.constraints) if (!table.constraints.includes(c)) table.constraints.push(c);
      break;
    }
    case 'alter-table': {
      const table = model.tables.get(value.table);
      if (!table) {
        model.otherStatements.push({ file, kind: 'alter-missing-table', sql: value.table });
        break;
      }
      for (const action of value.actions) {
        if (action.type === 'add-column') {
          if (!table.columns.some((c) => c.name === action.column.name)) {
            table.columns.push({ ...action.column, addedIn: file });
          }
        } else if (action.type === 'drop-column') {
          table.columns = table.columns.filter((c) => c.name !== action.column);
        } else if (action.type === 'rename-column') {
          const col = table.columns.find((c) => c.name === action.from);
          if (col) col.name = action.to;
        } else if (action.type === 'add-constraint') {
          table.constraints.push(action.definition);
        } else if (action.type === 'set-default') {
          const col = table.columns.find((c) => c.name === action.column);
          if (col) col.default = action.value;
        } else if (action.type === 'rls') {
          table.rls = action.enabled;
        } else {
          table.otherOps.push({ file, sql: action.sql });
        }
      }
      break;
    }
    case 'index': {
      model.indexes.set(value.name, { ...value, file });
      const table = model.tables.get(value.table);
      if (table) table.indexes.push({ name: value.name, unique: value.unique, columns: value.columns });
      break;
    }
    case 'function':
      model.functions.set(value.name, { ...value, file });
      break;
    case 'trigger': {
      model.triggers.set(value.name, { ...value, file });
      const table = model.tables.get(value.table);
      if (table) table.triggers.push(value.name);
      break;
    }
    case 'policy': {
      model.policies.push({ ...value, file });
      const table = model.tables.get(value.table.split('.').pop());
      if (table) table.policies.push({ name: value.name, command: value.command, file });
      break;
    }
    case 'type':
      model.enums.set(value.name, { ...value, file });
      break;
    case 'extension':
      model.extensions.add(value.name);
      break;
    case 'drop-table':
      for (const t of value) model.tables.delete(t);
      break;
    case 'do-block':
      model.doBlocks.push({ file });
      break;
    default:
      break;
  }
}

function parseDir(dir) {
  const model = emptyModel();
  const files = listSqlFiles(dir);
  for (const file of files) {
    const raw = fs.readFileSync(path.join(dir, file), 'utf8');
    const masked = maskSql(raw);
    for (const stmt of splitStatements(masked, raw)) {
      const parsed = parseStatement(stmt);
      apply(model, parsed, file);
    }
  }
  return { model, files };
}

function serialize(model, files) {
  const tables = [...model.tables.values()]
    .sort((a, b) => a.name.localeCompare(b.name))
    .map((t) => ({
      ...t,
      columnCount: t.columns.length,
      columns: [...t.columns].sort((a, b) => a.name.localeCompare(b.name)),
      indexCount: t.indexes.length,
      policyCount: t.policies.length,
      triggerCount: t.triggers.length,
    }));
  return {
    appliedFiles: files,
    tables,
    functions: [...model.functions.values()].sort((a, b) => a.name.localeCompare(b.name)),
    triggers: [...model.triggers.values()].sort((a, b) => a.name.localeCompare(b.name)),
    policies: model.policies.sort((a, b) => a.name.localeCompare(b.name)),
    indexes: [...model.indexes.values()].sort((a, b) => a.name.localeCompare(b.name)),
    enums: [...model.enums.values()],
    extensions: [...model.extensions].sort(),
    doBlocks: model.doBlocks,
    otherStatements: model.otherStatements,
    statistics: {
      tables: tables.length,
      columns: tables.reduce((n, t) => n + t.columns.length, 0),
      functions: model.functions.size,
      triggers: model.triggers.size,
      policies: model.policies.length,
      indexes: model.indexes.size,
      enums: model.enums.size,
      extensions: model.extensions.size,
      doBlocks: model.doBlocks.length,
    },
  };
}

function execGit(repo, gitArgs) {
  try {
    return execFileSync('git', ['-C', repo, ...gitArgs], { encoding: 'utf8' }).trim();
  } catch {
    return null;
  }
}

function columnSignature(column) {
  return [
    column.name,
    column.type,
    column.notNull ? 'NOT NULL' : 'NULL',
    column.default ? `DEFAULT ${column.default}` : '',
    column.references ? `REFERENCES ${column.references}` : '',
  ]
    .filter(Boolean)
    .join(' ');
}

function renderDiff(idswyft, go, provenance) {
  const goNames = new Set(go.tables.map((t) => t.name));
  const idNames = new Set(idswyft.tables.map((t) => t.name));
  const missingInGo = idswyft.tables.filter((t) => !goNames.has(t.name));
  const goOnly = go.tables.filter((t) => !idNames.has(t.name));
  const shared = idswyft.tables.filter((t) => goNames.has(t.name));

  const lines = [];
  lines.push('# idswyft vs Go 基座 Schema 差异清单（M0 契约冻结输入）');
  lines.push('');
  lines.push(`- idswyft 仓库：\`${provenance.idswyftRepo}\``);
  lines.push(`- idswyft 提交：\`${provenance.idswyftCommit}\`（${provenance.idswyftDate}）`);
  lines.push(`- Go 基座仓库：\`${provenance.goRepo}\``);
  lines.push(`- Go 基座提交：\`${provenance.goCommit}\``);
  lines.push(`- 生成时间：${provenance.generatedAt}`);
  lines.push('- 解析方式：静态 SQL 解析（注释/字符串/美元引用块掩码后按语句切分）');
  lines.push('');
  lines.push('## 汇总');
  lines.push('');
  lines.push('| 指标 | idswyft | Go 基座 | 差异 |');
  lines.push('| --- | --- | --- | --- |');
  lines.push(`| 表 | ${idswyft.statistics.tables} | ${go.statistics.tables} | 需新增 ${missingInGo.length}，Go 独有 ${goOnly.length} |`);
  lines.push(`| 列 | ${idswyft.statistics.columns} | ${go.statistics.columns} | - |`);
  lines.push(`| SQL 函数 | ${idswyft.statistics.functions} | 0 | 全部需实现 |`);
  lines.push(`| 触发器 | ${idswyft.statistics.triggers} | 0 | 全部需实现 |`);
  lines.push(`| RLS 策略 | ${idswyft.statistics.policies} | 0 | 需决策（保留 SQL 或应用层） |`);
  lines.push(`| 索引 | ${idswyft.statistics.indexes} | - | 随表迁移 |`);
  lines.push(`| 枚举 | ${idswyft.statistics.enums} | 0 | 需实现 |`);
  lines.push(`| 扩展 | ${idswyft.statistics.extensions} | 0 | - |`);
  lines.push(`| DO $$ 动态块 | ${idswyft.statistics.doBlocks} | 0 | 需人工复核 |`);
  lines.push('');
  lines.push('## 表差异');
  lines.push('');
  lines.push('### idswyft 表（31）');
  lines.push('');
  lines.push('| 表 | 列 | 索引 | 策略 | 触发器 | 来源迁移 |');
  lines.push('| --- | --- | --- | --- | --- | --- |');
  for (const t of idswyft.tables) {
    lines.push(`| ${t.name} | ${t.columnCount} | ${t.indexCount} | ${t.policyCount} | ${t.triggerCount} | ${t.sources.join('<br>')} |`);
  }
  lines.push('');
  lines.push('### Go 基座表（9）');
  lines.push('');
  lines.push('| 表 | 列 | 主键/约束 |');
  lines.push('| --- | --- | --- |');
  for (const t of go.tables) {
    const pk = t.columns.filter((c) => c.primaryKey).map((c) => c.name).join(', ') || '-';
    lines.push(`| ${t.name} | ${t.columnCount} | ${pk} |`);
  }
  lines.push('');
  lines.push('### 需新增到 Go 侧的表');
  lines.push('');
  for (const t of missingInGo) lines.push(`- ${t.name}（${t.columnCount} 列）`);
  lines.push('');
  if (goOnly.length > 0) {
    lines.push('### Go 独有表');
    lines.push('');
    for (const t of goOnly) lines.push(`- ${t.name}（${t.columnCount} 列）`);
    lines.push('');
  }
  if (shared.length > 0) {
    lines.push('### 同名表列差异');
    lines.push('');
    for (const t of shared) {
      const gt = go.tables.find((x) => x.name === t.name);
      const gcols = new Set(gt.columns.map((c) => c.name));
      const icols = new Set(t.columns.map((c) => c.name));
      const onlyId = t.columns.filter((c) => !gcols.has(c.name));
      const onlyGo = gt.columns.filter((c) => !icols.has(c.name));
      lines.push(`#### ${t.name}`);
      lines.push('');
      if (onlyId.length) {
        lines.push('idswyft 独有列：');
        for (const c of onlyId) lines.push(`- ${columnSignature(c)}`);
      }
      if (onlyGo.length) {
        lines.push('Go 独有列：');
        for (const c of onlyGo) lines.push(`- ${columnSignature(c)}`);
      }
      lines.push('');
    }
  }
  lines.push('## SQL 函数（需在 Go 侧保留或重写）');
  lines.push('');
  lines.push('| 函数 | 参数 | 返回 | 语言 | 来源 |');
  lines.push('| --- | --- | --- | --- | --- |');
  for (const f of idswyft.functions) {
    lines.push(`| ${f.name} | ${f.args || '-'} | ${f.returns || '-'} | ${f.language || '-'} | ${f.file} |`);
  }
  lines.push('');
  lines.push('## 触发器');
  lines.push('');
  lines.push('| 触发器 | 时机/事件 | 表 | 来源 |');
  lines.push('| --- | --- | --- | --- |');
  for (const t of idswyft.triggers) lines.push(`| ${t.name} | ${t.timing} | ${t.table} | ${t.file} |`);
  lines.push('');
  lines.push('## RLS 策略');
  lines.push('');
  lines.push('| 策略 | 表 | 命令 | USING | WITH CHECK | 来源 |');
  lines.push('| --- | --- | --- | --- | --- | --- |');
  for (const p of idswyft.policies) {
    lines.push(`| ${p.name} | ${p.table} | ${p.command} | ${p.using ? '是' : '否'} | ${p.withCheck ? '是' : '否'} | ${p.file} |`);
  }
  lines.push('');
  lines.push('## 枚举');
  lines.push('');
  if (idswyft.enums.length === 0) lines.push('无 `CREATE TYPE ... AS ENUM`（状态值在应用层常量中定义）。');
  for (const e of idswyft.enums) lines.push(`- ${e.name}: ${e.values.join(', ')}（${e.file}）`);
  lines.push('');
  lines.push('## 待复核与决策');
  lines.push('');
  lines.push(`- 存在 ${idswyft.statistics.doBlocks} 个 DO $$ 动态 SQL 块，静态解析无法展开，M1 需用 pg_dump 重放对账。`);
  lines.push('- 策略/触发器引用 `storage.objects`、`auth.uid()` 等 Supabase 内置对象；Go 侧需决定保留 Supabase Storage 语义还是替换为 S3 + 应用层策略。');
  lines.push('- 迁移存在重复数字前缀（如两个 04），当前按文件名自然序应用；M1 重放时需与 idswyft 迁移器顺序一致。');
  lines.push('- 同名表为 0：Go 侧 9 表与 idswyft 31 表是两套主体模型，需在 M0 决策“同库并存”或“新库独立 + /v1 降级为内部引擎”。');
  lines.push('- 本清单为静态解析结果；M1 门禁为：空库重放 67 个迁移后 `pg_dump --schema-only` 归一化 diff = 0。');
  lines.push('');
  lines.push('## 复算命令');
  lines.push('');
  lines.push('```bash');
  lines.push('cd contracts/tools && node extract-schema.mjs');
  lines.push('```');
  lines.push('');
  return lines.join('\n');
}

function main() {
  const { model: idModel, files: idFiles } = parseDir(MIGRATIONS_DIR);
  const idswyft = serialize(idModel, idFiles);
  const { model: goModel, files: goFiles } = parseDir(GO_SQL_DIR);
  const go = serialize(goModel, goFiles);

  const provenance = {
    idswyftRepo: REPO,
    idswyftCommit: execGit(REPO, ['rev-parse', 'HEAD']),
    idswyftDate: execGit(REPO, ['log', '-1', '--format=%cI']),
    goRepo: GO_REPO,
    goCommit: execGit(GO_REPO, ['rev-parse', 'HEAD']),
    generatedAt: new Date().toISOString(),
  };

  const validation = {
    expectedTables: EXPECTED_TABLES,
    actualTables: idswyft.statistics.tables,
    matches: idswyft.statistics.tables === EXPECTED_TABLES,
  };

  const output = { ...provenance, validation, ...idswyft };
  fs.mkdirSync(OUT_DIR, { recursive: true });
  fs.writeFileSync(path.join(OUT_DIR, 'schema.json'), JSON.stringify(output, null, 2) + '\n');
  fs.writeFileSync(path.join(OUT_DIR, 'go-schema.json'), JSON.stringify({ ...provenance, ...go }, null, 2) + '\n');
  fs.writeFileSync(path.join(OUT_DIR, 'schema-diff.md'), renderDiff(idswyft, go, provenance));

  console.log(JSON.stringify({ idswyft: idswyft.statistics, go: go.statistics, validation }, null, 2));
  if (!validation.matches) process.exitCode = 1;
}

main();
